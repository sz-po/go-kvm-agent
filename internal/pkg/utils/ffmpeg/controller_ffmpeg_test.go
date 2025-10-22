package ffmpeg

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	rxgo "github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/process"
)

// mockSupervisor is a mock implementation of process.Supervisor for testing.
type mockSupervisor struct {
	mock.Mock
}

func (mockSupervisor *mockSupervisor) Start(ctx context.Context, stableFor time.Duration) error {
	args := mockSupervisor.Called(ctx, stableFor)
	return args.Error(0)
}

func (mockSupervisor *mockSupervisor) Stop(ctx context.Context) error {
	args := mockSupervisor.Called(ctx)
	return args.Error(0)
}

func (mockSupervisor *mockSupervisor) Reload(ctx context.Context, stableFor time.Duration) error {
	args := mockSupervisor.Called(ctx, stableFor)
	return args.Error(0)
}

func (mockSupervisor *mockSupervisor) ReloadWithSpecification(ctx context.Context, specification process.Specification, stableFor time.Duration) error {
	args := mockSupervisor.Called(ctx, specification, stableFor)
	return args.Error(0)
}

func (mockSupervisor *mockSupervisor) Stdout() io.ReadCloser {
	args := mockSupervisor.Called()
	return args.Get(0).(io.ReadCloser)
}

func (mockSupervisor *mockSupervisor) Stderr() io.ReadCloser {
	args := mockSupervisor.Called()
	return args.Get(0).(io.ReadCloser)
}

func (mockSupervisor *mockSupervisor) Stdin() io.WriteCloser {
	args := mockSupervisor.Called()
	return args.Get(0).(io.WriteCloser)
}

func (mockSupervisor *mockSupervisor) Events() rxgo.Observable {
	args := mockSupervisor.Called()
	return args.Get(0).(rxgo.Observable)
}

func (mockSupervisor *mockSupervisor) Specification() process.Specification {
	args := mockSupervisor.Called()
	return args.Get(0).(process.Specification)
}

func (mockSupervisor *mockSupervisor) Status() process.Status {
	args := mockSupervisor.Called()
	return args.Get(0).(process.Status)
}

func (mockSupervisor *mockSupervisor) State() process.SupervisorState {
	args := mockSupervisor.Called()
	return args.Get(0).(process.SupervisorState)
}

func (mockSupervisor *mockSupervisor) Wait(ctx context.Context) error {
	args := mockSupervisor.Called(ctx)
	return args.Error(0)
}

// TestNewFFmpegController_DefaultSupervisorProvider tests that the default supervisor provider is used when none is specified.
func TestNewFFmpegController_DefaultSupervisorProvider(t *testing.T) {
	input := NewInputStdin()
	output := NewOutputStdout()
	configuration := RawConfiguration{}

	controller, err := NewFFmpegController(input, output, configuration)
	require.NoError(t, err)
	require.NotNil(t, controller)
	require.NotNil(t, controller.process)
}

// TestWithSupervisorProvider tests that a custom supervisor provider can be set.
func TestWithSupervisorProvider(t *testing.T) {
	customSupervisor := &mockSupervisor{}
	providerCalled := false
	var capturedSpecification process.Specification
	var capturedRestartPolicy process.RestartPolicy

	customProvider := func(specification process.Specification, restartPolicy process.RestartPolicy) process.Supervisor {
		providerCalled = true
		capturedSpecification = specification
		capturedRestartPolicy = restartPolicy
		return customSupervisor
	}

	input := NewInputStdin()
	output := NewOutputStdout()
	configuration := RawConfiguration{}

	controller, err := NewFFmpegController(
		input,
		output,
		configuration,
		WithSupervisorProvider(customProvider),
	)
	require.NoError(t, err)
	require.NotNil(t, controller)

	// Verify provider was called
	assert.True(t, providerCalled, "Custom provider should have been called")

	// Verify the returned supervisor is our mock
	assert.Equal(t, customSupervisor, controller.process)

	// Verify the specification was passed correctly
	assert.Equal(t, "/usr/local/bin/ffmpeg", capturedSpecification.ExecutablePath)
	assert.Contains(t, capturedSpecification.Arguments, "-progress")
	assert.Contains(t, capturedSpecification.Arguments, "pipe:2")

	// Verify restart policy was passed correctly
	assert.True(t, capturedRestartPolicy.Enabled)
	assert.Equal(t, 10, capturedRestartPolicy.MaxAttempts)
	assert.Equal(t, process.StrategyExponential, capturedRestartPolicy.Strategy)
	assert.Equal(t, time.Second, capturedRestartPolicy.InitialDelay)
	assert.Equal(t, time.Second*5, capturedRestartPolicy.MaxDelay)
	assert.Equal(t, time.Second*2, capturedRestartPolicy.ResetWindow)
}

// TestWithFFmpegLogger tests that a custom logger can be set.
func TestWithFFmpegLogger(t *testing.T) {
	customLogger := slog.New(slog.DiscardHandler)

	input := NewInputStdin()
	output := NewOutputStdout()
	configuration := RawConfiguration{}

	controller, err := NewFFmpegController(
		input,
		output,
		configuration,
		WithFFmpegLogger(customLogger),
	)
	require.NoError(t, err)
	require.NotNil(t, controller)
	assert.Equal(t, customLogger, controller.logger)
}

// TestWithFFmpegExecutablePath tests that a custom executable path can be set.
func TestWithFFmpegExecutablePath(t *testing.T) {
	customPath := "/custom/path/to/ffmpeg"
	providerCalled := false
	var capturedSpecification process.Specification

	customProvider := func(specification process.Specification, restartPolicy process.RestartPolicy) process.Supervisor {
		providerCalled = true
		capturedSpecification = specification
		return &mockSupervisor{}
	}

	input := NewInputStdin()
	output := NewOutputStdout()
	configuration := RawConfiguration{}

	controller, err := NewFFmpegController(
		input,
		output,
		configuration,
		WithFFmpegExecutablePath(customPath),
		WithSupervisorProvider(customProvider),
	)
	require.NoError(t, err)
	require.NotNil(t, controller)

	assert.True(t, providerCalled)
	assert.Equal(t, customPath, capturedSpecification.ExecutablePath)
}

// TestParseStatusLine_DropFrames tests parsing of drop_frames status line.
func TestParseStatusLine_DropFrames(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("drop_frames=123")

	status := controller.GetStatus()
	assert.Equal(t, int64(123), status.DroppedFrames)
}

// TestParseStatusLine_Speed tests parsing of speed status line.
func TestParseStatusLine_Speed(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("speed=1.5x")

	status := controller.GetStatus()
	assert.Equal(t, 1.5, status.Speed)
}

// TestParseStatusLine_FPS tests parsing of fps status line.
func TestParseStatusLine_FPS(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("fps=60.0")

	status := controller.GetStatus()
	assert.Equal(t, 60.0, status.FrameRate)
}

// TestParseStatusLine_TotalSize tests parsing of total_size status line.
func TestParseStatusLine_TotalSize(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("total_size=1024")

	status := controller.GetStatus()
	assert.Equal(t, int64(1024), status.TotalSize)
}

// TestParseStatusLine_InvalidFormat tests parsing of invalid status line formats.
func TestParseStatusLine_InvalidFormat(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	// No equals sign
	controller.parseStatusLine("invalid_line")
	status := controller.GetStatus()
	assert.Equal(t, FFmpegStatus{}, status)

	// Multiple equals signs - should only split on first
	controller.parseStatusLine("drop_frames=123=456")
	status = controller.GetStatus()
	// Should fail to parse "123=456" as integer and leave status unchanged
	assert.Equal(t, FFmpegStatus{}, status)
}

// TestParseStatusLine_UnknownKey tests parsing of unknown status keys.
func TestParseStatusLine_UnknownKey(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("unknown_key=123")

	status := controller.GetStatus()
	// Unknown keys should be ignored, status should remain empty
	assert.Equal(t, FFmpegStatus{}, status)
}

// TestParseStatusLine_InvalidValues tests parsing of invalid values.
func TestParseStatusLine_InvalidValues(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	// Invalid integer
	controller.parseStatusLine("drop_frames=not_a_number")
	status := controller.GetStatus()
	assert.Equal(t, int64(0), status.DroppedFrames)

	// Invalid float
	controller.parseStatusLine("speed=not_a_float")
	status = controller.GetStatus()
	assert.Equal(t, 0.0, status.Speed)
}

// TestGetStatus_ThreadSafety tests concurrent access to status.
func TestGetStatus_ThreadSafety(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	var waitGroup sync.WaitGroup
	iterations := 100

	// Spawn multiple goroutines that write status
	for i := 0; i < 10; i++ {
		waitGroup.Add(1)
		go func(iteration int) {
			defer waitGroup.Done()
			for j := 0; j < iterations; j++ {
				controller.parseStatusLine("drop_frames=" + string(rune(j)))
				controller.parseStatusLine("speed=1.5x")
				controller.parseStatusLine("fps=60.0")
				controller.parseStatusLine("total_size=1024")
			}
		}(i)
	}

	// Spawn multiple goroutines that read status
	for i := 0; i < 10; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for j := 0; j < iterations; j++ {
				_ = controller.GetStatus()
			}
		}()
	}

	waitGroup.Wait()

	// Test should complete without race conditions or deadlocks
}

// TestParseStatusLine_MultipleUpdates tests that status is updated correctly with multiple parse calls.
func TestParseStatusLine_MultipleUpdates(t *testing.T) {
	controller := &FFmpegController{
		currentStatus:     &FFmpegStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	// Parse multiple status lines
	controller.parseStatusLine("drop_frames=10")
	controller.parseStatusLine("speed=1.0x")
	controller.parseStatusLine("fps=30.0")
	controller.parseStatusLine("total_size=512")

	status := controller.GetStatus()
	assert.Equal(t, int64(10), status.DroppedFrames)
	assert.Equal(t, 1.0, status.Speed)
	assert.Equal(t, 30.0, status.FrameRate)
	assert.Equal(t, int64(512), status.TotalSize)

	// Update some values
	controller.parseStatusLine("drop_frames=20")
	controller.parseStatusLine("fps=60.0")

	status = controller.GetStatus()
	assert.Equal(t, int64(20), status.DroppedFrames)
	assert.Equal(t, 1.0, status.Speed) // Should remain unchanged
	assert.Equal(t, 60.0, status.FrameRate)
	assert.Equal(t, int64(512), status.TotalSize) // Should remain unchanged
}
