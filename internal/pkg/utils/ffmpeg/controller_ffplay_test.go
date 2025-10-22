package ffmpeg

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/process"
)

// TestNewFFplayController_DefaultSupervisorProvider tests that the default supervisor provider is used when none is specified.
func TestNewFFplayController_DefaultSupervisorProvider(t *testing.T) {
	input := NewInputStdin()
	configuration := RawConfiguration{}

	controller, err := NewFFplayController(input, configuration)
	require.NoError(t, err)
	require.NotNil(t, controller)
	require.NotNil(t, controller.process)
}

// TestWithFFplaySupervisorProvider tests that a custom supervisor provider can be set.
func TestWithFFplaySupervisorProvider(t *testing.T) {
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
	configuration := RawConfiguration{}

	controller, err := NewFFplayController(
		input,
		configuration,
		WithFFplaySupervisorProvider(customProvider),
	)
	require.NoError(t, err)
	require.NotNil(t, controller)

	// Verify provider was called
	assert.True(t, providerCalled, "Custom provider should have been called")

	// Verify the returned supervisor is our mock
	assert.Equal(t, customSupervisor, controller.process)

	// Verify the specification was passed correctly
	assert.Equal(t, "/usr/local/bin/ffplay", capturedSpecification.ExecutablePath)
	assert.Contains(t, capturedSpecification.Arguments, "-hide_banner")
	assert.Contains(t, capturedSpecification.Arguments, "-stats")

	// Verify restart policy was passed correctly
	assert.True(t, capturedRestartPolicy.Enabled)
	assert.Equal(t, 10, capturedRestartPolicy.MaxAttempts)
	assert.Equal(t, process.StrategyExponential, capturedRestartPolicy.Strategy)
	assert.Equal(t, time.Second, capturedRestartPolicy.InitialDelay)
	assert.Equal(t, time.Second*5, capturedRestartPolicy.MaxDelay)
	assert.Equal(t, time.Second*2, capturedRestartPolicy.ResetWindow)
}

// TestWithFFplayLogger tests that a custom logger can be set.
func TestWithFFplayLogger(t *testing.T) {
	customLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	input := NewInputStdin()
	configuration := RawConfiguration{}

	controller, err := NewFFplayController(
		input,
		configuration,
		WithFFplayLogger(customLogger),
	)
	require.NoError(t, err)
	require.NotNil(t, controller)
	assert.Equal(t, customLogger, controller.logger)
}

// TestWithFFplayExecutablePath tests that a custom executable path can be set.
func TestWithFFplayExecutablePath(t *testing.T) {
	customPath := "/custom/path/to/ffplay"
	providerCalled := false
	var capturedSpecification process.Specification

	customProvider := func(specification process.Specification, restartPolicy process.RestartPolicy) process.Supervisor {
		providerCalled = true
		capturedSpecification = specification
		return &mockSupervisor{}
	}

	input := NewInputStdin()
	configuration := RawConfiguration{}

	controller, err := NewFFplayController(
		input,
		configuration,
		WithFFplayExecutablePath(customPath),
		WithFFplaySupervisorProvider(customProvider),
	)
	require.NoError(t, err)
	require.NotNil(t, controller)

	assert.True(t, providerCalled)
	assert.Equal(t, customPath, capturedSpecification.ExecutablePath)
}

// TestNewFFplayController_NilInput tests that nil input returns an error.
func TestNewFFplayController_NilInput(t *testing.T) {
	configuration := RawConfiguration{}

	controller, err := NewFFplayController(nil, configuration)
	assert.Error(t, err)
	assert.Nil(t, controller)
	assert.ErrorIs(t, err, ErrFFplayNilInput)
}

// TestNewFFplayController_NilConfiguration tests that nil configuration returns an error.
func TestNewFFplayController_NilConfiguration(t *testing.T) {
	input := NewInputStdin()

	controller, err := NewFFplayController(input, nil)
	assert.Error(t, err)
	assert.Nil(t, controller)
	assert.ErrorIs(t, err, ErrFFplayNilConfiguration)
}

// TestParseStatusLine_FrameDrops tests parsing of frame drops (fd=) status line.
func TestParseStatusLine_FrameDrops(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("fd=123 aq=0KB vq=0KB sq=0B")

	status := controller.GetStatus()
	assert.Equal(t, int64(123), status.FrameDrops)
}

// TestParseStatusLine_AudioQueue tests parsing of audio queue (aq=) status line.
func TestParseStatusLine_AudioQueue(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("fd=0 aq=456KB vq=0KB sq=0B")

	status := controller.GetStatus()
	assert.Equal(t, int64(456), status.AudioQueue)
}

// TestParseStatusLine_VideoQueue tests parsing of video queue (vq=) status line.
func TestParseStatusLine_VideoQueue(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("fd=0 aq=0KB vq=789KB sq=0B")

	status := controller.GetStatus()
	assert.Equal(t, int64(789), status.VideoQueue)
}

// TestParseStatusLine_SubtitleQueue tests parsing of subtitle queue (sq=) status line.
func TestParseStatusLine_SubtitleQueue(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("fd=0 aq=0KB vq=0KB sq=12B")

	status := controller.GetStatus()
	assert.Equal(t, int64(12), status.SubtitleQueue)
}

// TestParseStatusLine_MultipleValues tests parsing of status line with multiple values.
func TestParseStatusLine_MultipleValues(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("fd=10 aq=20KB vq=30KB sq=5B")

	status := controller.GetStatus()
	assert.Equal(t, int64(10), status.FrameDrops)
	assert.Equal(t, int64(20), status.AudioQueue)
	assert.Equal(t, int64(30), status.VideoQueue)
	assert.Equal(t, int64(5), status.SubtitleQueue)
}

// TestParseStatusLine_NoMatches tests parsing of status line with no matching patterns.
func TestParseStatusLine_NoMatches(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	controller.parseStatusLine("some random output without matches")

	status := controller.GetStatus()
	assert.Equal(t, FFplayStatus{}, status)
}

// TestFFplayParseStatusLine_InvalidValues tests parsing of status line with invalid numeric values.
func TestFFplayParseStatusLine_InvalidValues(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
		currentStatusLock: sync.RWMutex{},
	}

	// Invalid frame drops value
	controller.parseStatusLine("fd=not_a_number aq=0KB vq=0KB sq=0B")
	status := controller.GetStatus()
	assert.Equal(t, int64(0), status.FrameDrops)

	// Invalid audio queue value
	controller.parseStatusLine("fd=0 aq=invalidKB vq=0KB sq=0B")
	status = controller.GetStatus()
	assert.Equal(t, int64(0), status.AudioQueue)
}

// TestBuildSpecification_NormalInputAndConfiguration tests buildSpecification with normal values.
func TestBuildSpecification_NormalInputAndConfiguration(t *testing.T) {
	controller := &FFplayController{
		options: &ffplayControllerOptions{
			executablePath: "/usr/local/bin/ffplay",
		},
	}

	input := NewInputStdin()
	configuration := RawConfiguration{"-vf", "scale=640:480"}

	specification := controller.buildSpecification(input, configuration)

	assert.Equal(t, "/usr/local/bin/ffplay", specification.ExecutablePath)
	assert.Contains(t, specification.Arguments, "-hide_banner")
	assert.Contains(t, specification.Arguments, "-stats")
	assert.Contains(t, specification.Arguments, "-i")
	assert.Contains(t, specification.Arguments, "pipe:0")
	assert.Contains(t, specification.Arguments, "-vf")
	assert.Contains(t, specification.Arguments, "scale=640:480")
}

// TestBuildSpecification_NilInput tests buildSpecification with nil input.
func TestBuildSpecification_NilInput(t *testing.T) {
	controller := &FFplayController{
		options: &ffplayControllerOptions{
			executablePath: "/usr/local/bin/ffplay",
		},
	}

	configuration := RawConfiguration{"-vf", "scale=640:480"}

	specification := controller.buildSpecification(nil, configuration)

	assert.Equal(t, "/usr/local/bin/ffplay", specification.ExecutablePath)
	assert.Contains(t, specification.Arguments, "-hide_banner")
	assert.Contains(t, specification.Arguments, "-stats")
	assert.Contains(t, specification.Arguments, "-vf")
	assert.Contains(t, specification.Arguments, "scale=640:480")
	assert.NotContains(t, specification.Arguments, "-i")
	assert.NotContains(t, specification.Arguments, "pipe:0")
}

// TestBuildSpecification_NilConfiguration tests buildSpecification with nil configuration.
func TestBuildSpecification_NilConfiguration(t *testing.T) {
	controller := &FFplayController{
		options: &ffplayControllerOptions{
			executablePath: "/usr/local/bin/ffplay",
		},
	}

	input := NewInputStdin()

	specification := controller.buildSpecification(input, nil)

	assert.Equal(t, "/usr/local/bin/ffplay", specification.ExecutablePath)
	assert.Contains(t, specification.Arguments, "-hide_banner")
	assert.Contains(t, specification.Arguments, "-stats")
	assert.Contains(t, specification.Arguments, "-i")
	assert.Contains(t, specification.Arguments, "pipe:0")
}

// TestBuildSpecification_BothNil tests buildSpecification with both nil input and configuration.
func TestBuildSpecification_BothNil(t *testing.T) {
	controller := &FFplayController{
		options: &ffplayControllerOptions{
			executablePath: "/usr/local/bin/ffplay",
		},
	}

	specification := controller.buildSpecification(nil, nil)

	assert.Equal(t, "/usr/local/bin/ffplay", specification.ExecutablePath)
	assert.Contains(t, specification.Arguments, "-hide_banner")
	assert.Contains(t, specification.Arguments, "-stats")
	assert.Len(t, specification.Arguments, 2) // Only -hide_banner and -stats
}

// TestFFplayGetStatus_ThreadSafety tests concurrent access to status.
func TestFFplayGetStatus_ThreadSafety(t *testing.T) {
	controller := &FFplayController{
		currentStatus:     &FFplayStatus{},
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
				controller.parseStatusLine("fd=10 aq=20KB vq=30KB sq=5B")
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
