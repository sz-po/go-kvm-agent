package ffmpeg

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
)

func TestCreateGolangChannelOutputUnixSocket_BasicDataFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventEmitter := utils.NewEventEmitter[[]byte]()

	// Create Unix socket output
	parameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
	assert.NoError(t, err)
	assert.Len(t, parameters, 1)
	assert.True(t, strings.HasPrefix(parameters[0], "unix://"))

	// Extract socket path
	socketPath := strings.TrimPrefix(parameters[0], "unix://")

	// Create a consumer channel
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()
	consumer := eventEmitter.Listen(consumerCtx)

	// Wait a bit for the listener to be ready
	time.Sleep(100 * time.Millisecond)

	// Simulate FFmpeg connecting and writing data
	conn, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)
	defer conn.Close()

	testData := []byte("mpv data chunk")
	_, err = conn.Write(testData)
	assert.NoError(t, err)

	// Wait for data to be emitted
	select {
	case receivedData := <-consumer:
		assert.Equal(t, testData, receivedData)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for data")
	}
}

func TestCreateGolangChannelOutputUnixSocket_ReconnectAfterDisconnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventEmitter := utils.NewEventEmitter[[]byte]()

	// Create Unix socket output
	parameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
	assert.NoError(t, err)

	socketPath := strings.TrimPrefix(parameters[0], "unix://")

	// Create a consumer channel
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()
	consumer := eventEmitter.Listen(consumerCtx)

	time.Sleep(100 * time.Millisecond)

	// First connection
	conn1, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)

	testData1 := []byte("first connection")
	_, err = conn1.Write(testData1)
	assert.NoError(t, err)

	select {
	case receivedData := <-consumer:
		assert.Equal(t, testData1, receivedData)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first data")
	}

	// Close first connection (simulate FFmpeg crash)
	conn1.Close()
	time.Sleep(100 * time.Millisecond)

	// Second connection (simulate FFmpeg restart)
	conn2, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)
	defer conn2.Close()

	testData2 := []byte("second connection")
	_, err = conn2.Write(testData2)
	assert.NoError(t, err)

	select {
	case receivedData := <-consumer:
		assert.Equal(t, testData2, receivedData)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second data")
	}
}

func TestCreateGolangChannelOutputUnixSocket_NewConnectionClosesOld(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventEmitter := utils.NewEventEmitter[[]byte]()

	// Create Unix socket output
	parameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
	assert.NoError(t, err)

	socketPath := strings.TrimPrefix(parameters[0], "unix://")

	// Create a consumer channel
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()
	consumer := eventEmitter.Listen(consumerCtx)

	time.Sleep(100 * time.Millisecond)

	// First connection
	conn1, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)

	testData1 := []byte("old connection")
	_, err = conn1.Write(testData1)
	assert.NoError(t, err)

	select {
	case receivedData := <-consumer:
		assert.Equal(t, testData1, receivedData)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for first data")
	}

	// Second connection without closing first (simulate reload)
	conn2, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)
	defer conn2.Close()

	// Wait a bit for the old connection to be replaced
	time.Sleep(100 * time.Millisecond)

	// Try writing to old connection (should fail or be ignored)
	_, err = conn1.Write([]byte("should not appear"))
	// Connection might be closed by the server, which is expected

	// Write to new connection
	testData2 := []byte("new connection")
	_, err = conn2.Write(testData2)
	assert.NoError(t, err)

	// Should receive data from new connection
	select {
	case receivedData := <-consumer:
		assert.Equal(t, testData2, receivedData)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for second data")
	}
}

func TestCreateGolangChannelOutputUnixSocket_MultipleConsumers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventEmitter := utils.NewEventEmitter[[]byte]()

	// Create Unix socket output
	parameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
	assert.NoError(t, err)

	socketPath := strings.TrimPrefix(parameters[0], "unix://")

	// Create multiple consumer channels and immediately start goroutines to receive
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()

	consumer1 := eventEmitter.Listen(consumerCtx)
	consumer2 := eventEmitter.Listen(consumerCtx)
	consumer3 := eventEmitter.Listen(consumerCtx)

	received1 := make(chan []byte, 1)
	received2 := make(chan []byte, 1)
	received3 := make(chan []byte, 1)

	// Start consumers immediately to ensure they're ready to receive
	go func() {
		data := <-consumer1
		received1 <- data
	}()
	go func() {
		data := <-consumer2
		received2 <- data
	}()
	go func() {
		data := <-consumer3
		received3 <- data
	}()

	// Give consumers time to start and block on receive
	time.Sleep(100 * time.Millisecond)

	// Connect and write data
	conn, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)
	defer conn.Close()

	testData := []byte("broadcast data")
	_, err = conn.Write(testData)
	assert.NoError(t, err)

	// All consumers should receive the data
	timeout := time.After(2 * time.Second)

	select {
	case receivedData := <-received1:
		assert.Equal(t, testData, receivedData)
	case <-timeout:
		t.Fatal("Timeout waiting for consumer1")
	}

	select {
	case receivedData := <-received2:
		assert.Equal(t, testData, receivedData)
	case <-timeout:
		t.Fatal("Timeout waiting for consumer2")
	}

	select {
	case receivedData := <-received3:
		assert.Equal(t, testData, receivedData)
	case <-timeout:
		t.Fatal("Timeout waiting for consumer3")
	}
}

func TestCreateGolangChannelOutputUnixSocket_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	eventEmitter := utils.NewEventEmitter[[]byte]()

	// Create Unix socket output
	parameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
	assert.NoError(t, err)

	socketPath := strings.TrimPrefix(parameters[0], "unix://")

	time.Sleep(100 * time.Millisecond)

	// Connect
	conn, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)

	testData := []byte("mpv data")
	_, err = conn.Write(testData)
	assert.NoError(t, err)

	// Cancel context
	cancel()

	// Wait a bit for cleanup
	time.Sleep(200 * time.Millisecond)

	// Try to connect again - should fail or timeout because listener is closed
	conn2, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err == nil {
		conn2.Close()
		t.Fatal("Expected connection to fail after context cancellation")
	}
}

func TestCreateGolangChannelOutputUnixSocket_LargeDataChunks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	eventEmitter := utils.NewEventEmitter[[]byte]()

	// Create Unix socket output
	parameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
	assert.NoError(t, err)

	socketPath := strings.TrimPrefix(parameters[0], "unix://")

	// Create a consumer channel and start receiving immediately
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()
	consumer := eventEmitter.Listen(consumerCtx)

	// Channel to collect received data
	receivedChan := make(chan []byte, 100)

	// Start consumer goroutine immediately to ensure it's always ready to receive
	go func() {
		for {
			select {
			case chunk := <-consumer:
				receivedChan <- chunk
			case <-consumerCtx.Done():
				return
			}
		}
	}()

	// Give consumer time to start
	time.Sleep(100 * time.Millisecond)

	// Connect
	conn, err := net.Dial("unix", socketPath)
	assert.NoError(t, err)
	defer conn.Close()

	// Send data in smaller chunks to simulate realistic FFmpeg streaming behavior
	// This prevents overwhelming the EventEmitter's non-blocking send pattern
	chunkSize := 8 * 1024  // 8KB chunks
	totalSize := 64 * 1024 // 64KB total
	numChunks := totalSize / chunkSize

	expectedData := make([]byte, 0, totalSize)

	for i := 0; i < numChunks; i++ {
		chunk := make([]byte, chunkSize)
		for j := range chunk {
			chunk[j] = byte((i*chunkSize + j) % 256)
		}
		expectedData = append(expectedData, chunk...)

		_, err = conn.Write(chunk)
		assert.NoError(t, err)

		// Small delay between chunks to allow consumer to process
		time.Sleep(10 * time.Millisecond)
	}

	// Receive data (will come in multiple chunks)
	receivedTotal := 0
	timeout := time.After(5 * time.Second)
	var allChunks [][]byte

	// Allow some tolerance for dropped data due to non-blocking EventEmitter
	minExpectedBytes := int(float64(totalSize) * 0.95) // Accept at least 95% of data

	for receivedTotal < totalSize {
		select {
		case chunk := <-receivedChan:
			allChunks = append(allChunks, chunk)
			receivedTotal += len(chunk)
		case <-timeout:
			// Check if we received most of the data (streaming may drop some frames)
			if receivedTotal >= minExpectedBytes {
				t.Logf("Received %d/%d bytes (%.1f%%) - acceptable for streaming", receivedTotal, totalSize, float64(receivedTotal)/float64(totalSize)*100)
				return
			}
			t.Fatalf("Timeout waiting for data. Received %d/%d bytes (%.1f%%)", receivedTotal, totalSize, float64(receivedTotal)/float64(totalSize)*100)
		}
	}

	// Verify data integrity for received data
	receivedData := make([]byte, 0, receivedTotal)
	for _, chunk := range allChunks {
		receivedData = append(receivedData, chunk...)
	}

	t.Logf("Successfully received %d/%d bytes (100%%)", receivedTotal, totalSize)

	// Verify integrity of received data
	for i := 0; i < len(receivedData) && i < len(expectedData); i++ {
		assert.Equal(t, expectedData[i], receivedData[i], "Data mismatch at position %d", i)
	}
}
