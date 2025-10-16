package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unixSocketTestServer creates and runs a real unix socket listener for testing
type unixSocketTestServer struct {
	socketPath string
	listener   net.Listener
	connected  chan net.Conn
	received   chan []byte
	errors     chan error
	done       chan struct{}
}

func newUnixSocketTestServer(t *testing.T, socketPath string) *unixSocketTestServer {
	t.Helper()

	// Remove socket file if it exists
	_ = os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	require.NoError(t, err, "Failed to create unix socket listener")

	server := &unixSocketTestServer{
		socketPath: socketPath,
		listener:   listener,
		connected:  make(chan net.Conn, 1),
		received:   make(chan []byte, 10),
		errors:     make(chan error, 10),
		done:       make(chan struct{}),
	}

	go server.acceptLoop()

	return server
}

func (server *unixSocketTestServer) acceptLoop() {
	for {
		select {
		case <-server.done:
			return
		default:
			connection, err := server.listener.Accept()
			if err != nil {
				select {
				case server.errors <- err:
				case <-server.done:
					return
				}
				continue
			}

			server.connected <- connection
			go server.readLoop(connection)
		}
	}
}

func (server *unixSocketTestServer) readLoop(connection net.Conn) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-server.done:
			return
		default:
			bytesRead, err := connection.Read(buffer)
			if err != nil {
				if err != io.EOF {
					select {
					case server.errors <- err:
					case <-server.done:
					}
				}
				return
			}

			data := make([]byte, bytesRead)
			copy(data, buffer[:bytesRead])

			select {
			case server.received <- data:
			case <-server.done:
				return
			}
		}
	}
}

func (server *unixSocketTestServer) close() {
	close(server.done)
	_ = server.listener.Close()
	_ = os.Remove(server.socketPath)
}

func (server *unixSocketTestServer) waitForConnection(timeout time.Duration) (net.Conn, error) {
	select {
	case connection := <-server.connected:
		return connection, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for connection")
	}
}

func (server *unixSocketTestServer) waitForData(timeout time.Duration) ([]byte, error) {
	select {
	case data := <-server.received:
		return data, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for data")
	}
}

func TestInputWriterUnixSocket_Connect(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	// First connect should succeed
	err = mode.Connect()
	assert.NoError(t, err)
	assert.True(t, mode.IsConnected())

	_, err = server.waitForConnection(1 * time.Second)
	assert.NoError(t, err, "Server should have received connection")

	// Second connect should fail
	err = mode.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already connected")

	// Cleanup
	err = mode.Close()
	assert.NoError(t, err)

	_ = ctx
}

func TestInputWriterUnixSocket_ConnectWithoutListener(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "nonexistent.sock")

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	err = mode.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to unix socket")
	assert.False(t, mode.IsConnected())
}

func TestInputWriterUnixSocket_Write(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	err = mode.Connect()
	require.NoError(t, err)

	_, err = server.waitForConnection(1 * time.Second)
	require.NoError(t, err)

	// Write data
	testData := []byte("test data payload")
	written, err := mode.Write(testData)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), written)

	// Verify server received the data
	received, err := server.waitForData(1 * time.Second)
	assert.NoError(t, err)
	assert.Equal(t, testData, received)

	// Cleanup
	err = mode.Close()
	assert.NoError(t, err)

	_ = ctx
}

func TestInputWriterUnixSocket_WriteWithoutConnect(t *testing.T) {
	mode := &inputWriterUnixSocket{
		socketPath: "/tmp/doesnt-matter.sock",
	}

	testData := []byte("test data")
	written, err := mode.Write(testData)
	assert.Error(t, err)
	assert.Equal(t, ErrInputWriterNotConnected, err)
	assert.Equal(t, 0, written)
}

func TestInputWriterUnixSocket_WriteAfterConnectionClosed(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	err = mode.Connect()
	require.NoError(t, err)

	connection, err := server.waitForConnection(1 * time.Second)
	require.NoError(t, err)

	// Close connection from server side
	err = connection.Close()
	require.NoError(t, err)

	// Give it a moment to propagate
	time.Sleep(100 * time.Millisecond)

	// Write should fail and clear the connection
	testData := []byte("test data")
	written, err := mode.Write(testData)
	assert.Error(t, err)
	assert.Equal(t, 0, written)

	// After failed write, connection should be marked as disconnected
	assert.False(t, mode.IsConnected())

	_ = ctx
}

func TestInputWriterUnixSocket_Close(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	err = mode.Connect()
	require.NoError(t, err)

	_, err = server.waitForConnection(1 * time.Second)
	require.NoError(t, err)

	// Close should succeed
	err = mode.Close()
	assert.NoError(t, err)
	assert.False(t, mode.IsConnected())

	// Second close should fail
	err = mode.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already disconnected")

	_ = ctx
}

func TestInputWriterUnixSocket_IsConnected(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	// Initially not connected
	assert.False(t, mode.IsConnected())

	// After connect
	err = mode.Connect()
	require.NoError(t, err)
	assert.True(t, mode.IsConnected())

	_, err = server.waitForConnection(1 * time.Second)
	require.NoError(t, err)

	// After close
	err = mode.Close()
	require.NoError(t, err)
	assert.False(t, mode.IsConnected())

	_ = ctx
}

func TestInputWriterUnixSocket_Parameters(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")

	mode := &inputWriterUnixSocket{
		socketPath: socketPath,
	}

	parameters := mode.Parameters()
	assert.Equal(t, []string{
		"-i",
		"-listen",
		"1",
		fmt.Sprintf("unix://%s", socketPath),
	}, parameters)
}

func TestInputWriterUnixSocketMode_Provider(t *testing.T) {
	ctx := context.Background()

	provider := InputWriterUnixSocketMode()
	mode, err := provider(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, mode)

	// Verify the mode is usable
	parameters := mode.Parameters()
	assert.Contains(t, parameters, "-i")
	assert.Contains(t, parameters, "-listen")
	assert.Contains(t, parameters, "1")

	// Find the unix socket path
	var socketPath string
	for _, param := range parameters {
		if len(param) > 7 && param[:7] == "unix://" {
			socketPath = param[7:]
			break
		}
	}
	assert.NotEmpty(t, socketPath, "Socket path should be in parameters")

	// Cleanup temp directory
	tempDir := path.Dir(socketPath)
	defer os.RemoveAll(tempDir)
}

func TestInputWriter_ControlLoop_AutoReconnect(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")

	// Create writer with custom mode provider
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	modeProvider := func(ctx context.Context) (InputWriterMode, error) {
		return &inputWriterUnixSocket{
			socketPath: socketPath,
		}, nil
	}

	writer, err := NewInputWriter(ctx, modeProvider)
	require.NoError(t, err)
	require.NotNil(t, writer)

	// Initially no server, writer should not be connected
	time.Sleep(200 * time.Millisecond)

	// Start server
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	// Control loop should auto-connect within ~100ms ticker interval
	_, err = server.waitForConnection(2 * time.Second)
	assert.NoError(t, err, "Control loop should have auto-connected")

	// Write data
	testData := []byte("auto reconnect test")
	written, err := writer.Write(testData)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), written)

	// Verify server received data
	received, err := server.waitForData(1 * time.Second)
	assert.NoError(t, err)
	assert.Equal(t, testData, received)
}

func TestInputWriter_ControlLoop_Cleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	modeProvider := func(ctx context.Context) (InputWriterMode, error) {
		return &inputWriterUnixSocket{
			socketPath: socketPath,
		}, nil
	}

	writer, err := NewInputWriter(ctx, modeProvider)
	require.NoError(t, err)

	_ = writer // Writer is used implicitly for connection establishment and cleanup

	// Wait for connection
	connection, err := server.waitForConnection(2 * time.Second)
	require.NoError(t, err)

	// Cancel context - should trigger cleanup in control loop
	cancel()

	// Give control loop time to cleanup
	time.Sleep(200 * time.Millisecond)

	// Try to read from connection - should fail because it was closed
	buffer := make([]byte, 1)
	connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, err = connection.Read(buffer)
	assert.Error(t, err, "Connection should be closed after context cancellation")
}

func TestInputWriter_Parameters(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	modeProvider := InputWriterUnixSocketMode()
	writer, err := NewInputWriter(ctx, modeProvider)
	require.NoError(t, err)
	require.NotNil(t, writer)

	parameters := writer.Parameters()
	assert.Contains(t, parameters, "-i")
	assert.Contains(t, parameters, "-listen")
	assert.Contains(t, parameters, "1")

	// Should have a unix socket path
	hasUnixPath := false
	for _, param := range parameters {
		if len(param) > 7 && param[:7] == "unix://" {
			hasUnixPath = true
			break
		}
	}
	assert.True(t, hasUnixPath, "Parameters should contain unix socket path")
}

func TestInputWriter_MultipleWrites(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	socketPath := path.Join(tempDir, "test.sock")
	server := newUnixSocketTestServer(t, socketPath)
	defer server.close()

	modeProvider := func(ctx context.Context) (InputWriterMode, error) {
		return &inputWriterUnixSocket{
			socketPath: socketPath,
		}, nil
	}

	writer, err := NewInputWriter(ctx, modeProvider)
	require.NoError(t, err)

	// Wait for auto-connect
	_, err = server.waitForConnection(2 * time.Second)
	require.NoError(t, err)

	// Write multiple chunks
	testChunks := [][]byte{
		[]byte("chunk 1"),
		[]byte("chunk 2"),
		[]byte("chunk 3"),
	}

	for _, chunk := range testChunks {
		written, err := writer.Write(chunk)
		assert.NoError(t, err)
		assert.Equal(t, len(chunk), written)

		received, err := server.waitForData(1 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, chunk, received)
	}
}
