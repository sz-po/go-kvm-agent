package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
)

// GolangChannelOutputMode defines how data should be transmitted from FFmpeg to the Go process.
type GolangChannelOutputMode string

const (
	// GolangChannelOutputModeUnixSocket transmits data through a Unix socket.
	GolangChannelOutputModeUnixSocket GolangChannelOutputMode = "unix"
)

// GolangChannelOutput allows reading output from FFmpeg as a standard Go channel. It survives FFmpeg restarts
// by maintaining the channel open and continuing to emit data from new FFmpeg instances. Multiple consumers can
// read from the same output simultaneously using the broadcast pattern.
type GolangChannelOutput struct {
	outputMode   GolangChannelOutputMode
	eventEmitter *utils.EventEmitter[[]byte]

	parameters []string
}

// NewGolangChannelOutput creates a new instance and validates the outputMode.
func NewGolangChannelOutput(ctx context.Context, outputMode GolangChannelOutputMode) (*GolangChannelOutput, error) {
	eventEmitter := utils.NewEventEmitter[[]byte](
		utils.WithEventEmitterQueueSize[[]byte](16),
	)
	parameters := []string{}

	switch outputMode {
	case GolangChannelOutputModeUnixSocket:
		outputParameters, err := createGolangChannelOutputUnixSocket(ctx, eventEmitter)
		if err != nil {
			return nil, fmt.Errorf("create unix socket output: %w", err)
		}
		parameters = append(parameters, outputParameters...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrGolangChannelOutputModeInvalid, outputMode)
	}

	return &GolangChannelOutput{
		outputMode:   outputMode,
		eventEmitter: eventEmitter,
		parameters:   parameters,
	}, nil
}

func (output *GolangChannelOutput) Parameters() []string {
	return output.parameters
}

// Channel returns a channel of byte chunks. Each call to Channel returns a new channel, allowing multiple consumers
// to receive the same data simultaneously (broadcast pattern). Channels are automatically closed when the context is cancelled.
func (output *GolangChannelOutput) Channel(ctx context.Context) <-chan []byte {
	return output.eventEmitter.Listen(ctx)
}

// socketState manages the current active FFmpeg connection with thread-safe access.
// Only one FFmpeg instance can write at a time; new connections replace old ones.
type socketState struct {
	currentConnection net.Conn
	mutex             sync.Mutex
}

// replaceConnection sets a new connection as current and closes the old one if it exists.
// Returns the old connection that was replaced.
func (state *socketState) replaceConnection(newConnection net.Conn) net.Conn {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	oldConnection := state.currentConnection
	state.currentConnection = newConnection

	if oldConnection != nil {
		slog.Info("Replacing old FFmpeg connection with new one.")
		oldConnection.Close()
	}

	return oldConnection
}

// clearIfCurrent clears the current connection only if it matches the provided connection.
// This prevents clearing a newer connection that has already replaced this one.
func (state *socketState) clearIfCurrent(connection net.Conn) {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	if state.currentConnection == connection {
		state.currentConnection = nil
	}
}

// closeAll closes the current connection if one exists and clears it.
func (state *socketState) closeAll() {
	state.mutex.Lock()
	defer state.mutex.Unlock()

	if state.currentConnection != nil {
		state.currentConnection.Close()
		state.currentConnection = nil
	}
}

// createGolangChannelOutputUnixSocket creates a Unix socket implementation that reads data from FFmpeg
// and emits it as events through the eventEmitter. Returns FFmpeg command-line parameters for the socket output.
func createGolangChannelOutputUnixSocket(ctx context.Context, eventEmitter *utils.EventEmitter[[]byte]) ([]string, error) {
	// Use shorter filename to avoid exceeding Unix socket path length limit (typically 104 chars on macOS)
	socketPath := filepath.Join(os.TempDir(), "ffmpeg-"+uuid.New().String()[:8]+".sock")

	// Remove existing socket file if present
	_ = os.Remove(socketPath)

	listener, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: socketPath,
		Net:  "unix",
	})
	if err != nil {
		return nil, fmt.Errorf("create unix socket listener at %s: %w", socketPath, err)
	}

	state := &socketState{}
	go acceptConnections(ctx, listener, socketPath, eventEmitter, state)

	return []string{"unix://" + socketPath}, nil
}

// acceptConnections runs a loop accepting new FFmpeg connections. When a new connection arrives,
// it replaces the old one (if any) and starts reading data in a separate goroutine.
func acceptConnections(ctx context.Context, listener net.Listener, socketPath string, eventEmitter *utils.EventEmitter[[]byte], state *socketState) {
	defer cleanup(listener, socketPath, state)

	acceptChan := make(chan net.Conn)
	errorChan := make(chan error)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				errorChan <- err
				return
			}

			acceptChan <- conn
		}
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping Unix socket listener.")
			return
		case conn := <-acceptChan:
			slog.Info("New FFmpeg connection accepted.", slog.String("socketPath", socketPath))
			state.replaceConnection(conn)
			go handleConnection(ctx, conn, eventEmitter, state)
		case err := <-errorChan:
			slog.Error("Error accepting connection.", slog.Any("error", err))
			return
		}
	}
}

// handleConnection reads data from the FFmpeg connection and emits it through the EventEmitter.
// It runs until the connection is closed (EOF), an error occurs, or the context is cancelled.
func handleConnection(ctx context.Context, conn net.Conn, eventEmitter *utils.EventEmitter[[]byte], state *socketState) {
	defer func() {
		conn.Close()
		state.clearIfCurrent(conn)
	}()

	buffer := make([]byte, 1024*1024)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			n, err := conn.Read(buffer)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buffer[:n])
				eventEmitter.Emit(chunk)
			}
			if err != nil {
				if err != io.EOF {
					slog.Error("Error reading from FFmpeg connection.", slog.Any("error", err))
				}
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		conn.Close()
	case <-done:
	}
}

// cleanup closes the listener, all active connections, and removes the socket file.
func cleanup(listener net.Listener, socketPath string, state *socketState) {
	listener.Close()
	state.closeAll()
	_ = os.Remove(socketPath)
	slog.Info("Unix socket cleaned up.", slog.String("socketPath", socketPath))
}

var ErrGolangChannelOutputModeInvalid = errors.New("invalid golang channel output mode")
