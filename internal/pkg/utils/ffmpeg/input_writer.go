package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils"
)

type InputWriterMode interface {
	Connect() error
	Write(byte []byte) (int, error)
	Close() error
	IsConnected() bool
	Parameters() []string
}

type InputWriterModeProvider func(ctx context.Context) (InputWriterMode, error)

type inputWriterUnixSocket struct {
	socketPath string

	connection     net.Conn
	connectionLock sync.Mutex
}

func InputWriterUnixSocketMode() InputWriterModeProvider {
	return func(ctx context.Context) (InputWriterMode, error) {
		tempDir, err := os.MkdirTemp("/tmp", "ffmpeg-input")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir: %w", err)
		}

		socketPath := path.Join(tempDir, "ffmpeg-input.sock")

		mode := &inputWriterUnixSocket{
			socketPath:     socketPath,
			connectionLock: sync.Mutex{},
		}

		return mode, nil
	}
}

func (input *inputWriterUnixSocket) Connect() error {
	input.connectionLock.Lock()
	defer input.connectionLock.Unlock()

	if input.connection != nil {
		return fmt.Errorf("already connected")
	}

	connection, err := net.Dial("unix", input.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to unix socket: %w", err)
	}

	input.connection = connection

	return nil
}

func (input *inputWriterUnixSocket) Write(byte []byte) (int, error) {
	input.connectionLock.Lock()
	defer input.connectionLock.Unlock()

	if input.connection == nil {
		return 0, ErrInputWriterNotConnected
	}

	writtenBytes, err := input.connection.Write(byte)
	if err != nil {
		if utils.IsConnectionClosedError(err) {
			input.connection = nil
		}
	}

	return writtenBytes, err
}

func (input *inputWriterUnixSocket) Close() error {
	input.connectionLock.Lock()
	defer input.connectionLock.Unlock()

	if input.connection == nil {
		return fmt.Errorf("already disconnected")
	}

	err := input.connection.Close()

	input.connection = nil

	return err
}

func (input *inputWriterUnixSocket) IsConnected() bool {
	input.connectionLock.Lock()
	defer input.connectionLock.Unlock()

	return input.connection != nil
}

func (input *inputWriterUnixSocket) Parameters() []string {
	return []string{
		"-i",
		"-listen",
		"1",
		fmt.Sprintf("unix://%s", input.socketPath),
	}
}

// InputWriter exposes standard io.Writer interface for providing data to ffmpeg/ffplay.
type InputWriter struct {
	mode InputWriterMode
}

func NewInputWriter(ctx context.Context, modeProvider InputWriterModeProvider) (*InputWriter, error) {
	mode, err := modeProvider(ctx)
	if err != nil {
		return nil, err
	}

	writer := &InputWriter{
		mode: mode,
	}

	go writer.controlLoop(ctx)

	return writer, nil
}

func (input *InputWriter) Write(payload []byte) (n int, err error) {
	return input.mode.Write(payload)
}

func (input *InputWriter) Parameters() []string {
	return input.mode.Parameters()
}

func (input *InputWriter) controlLoop(ctx context.Context) {
	done := ctx.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			_ = input.mode.Close()
			return
		case <-ticker.C:
			if input.mode.IsConnected() {
				break
			}

			_ = input.mode.Connect()
		}
	}
}

var ErrInputWriterNotConnected = errors.New("input writer not connected")
