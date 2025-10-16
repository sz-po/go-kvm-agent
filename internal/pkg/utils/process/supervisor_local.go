package process

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	rxgo "github.com/reactivex/rxgo/v2"
)

// LocalSupervisorOpt is a functional option for configuring a local Supervisor implementation.
type LocalSupervisorOpt func(*supervisorLocal)

// WithLogger configures the supervisor to use the provided logger.
// By default, the supervisor uses a discard logger (no logging).
func WithLogger(logger *slog.Logger) LocalSupervisorOpt {
	return func(supervisor *supervisorLocal) {
		supervisor.logger = logger
	}
}

// processInstance represents a single running process instance managed by the supervisor.
type processInstance struct {
	// command is the exec.Cmd managing this process.
	command *exec.Cmd

	// pid is the process ID of the running process.
	pid int

	// startedAt records when this process instance was started.
	startedAt time.Time

	// exitChannel receives exactly one ExitError when the process exits, then closes.
	exitChannel chan *ExitError

	// context is used to signal cancellation to goroutines associated with this instance.
	context context.Context

	// cancel cancels the context for this instance.
	cancel context.CancelFunc

	// stdoutPipe is the pipe reading from the process's stdout.
	stdoutPipe io.ReadCloser

	// stderrPipe is the pipe reading from the process's stderr.
	stderrPipe io.ReadCloser

	// stdinPipe is the pipe writing to the process's stdin.
	stdinPipe io.WriteCloser
}

// supervisorLocal is the concrete implementation of the Supervisor interface.
type supervisorLocal struct {
	// specification contains the initial process configuration.
	specification Specification

	// restartPolicy configures automatic restart behavior for crashed processes.
	restartPolicy RestartPolicy

	// stateMutex protects state transitions and current instance access.
	stateMutex sync.RWMutex

	// state is the current lifecycle state of the supervisor.
	state SupervisorState

	// currentInstance points to the currently running (or most recent) process instance.
	currentInstance *processInstance

	// restartCount tracks the total number of successful restarts.
	restartCount int

	// crashRestartAttempts tracks the number of consecutive crash-triggered restart attempts.
	// This counter is reset to zero when the process runs stably for restartPolicy.ResetWindow.
	crashRestartAttempts int

	// lastCrashTime records the time of the most recent process crash.
	// Used to determine when to reset the crashRestartAttempts counter.
	lastCrashTime time.Time

	// eventChannel is the channel for publishing events to the observable.
	eventChannel chan rxgo.Item

	// eventObservable is the RxGo observable for event streaming.
	eventObservable rxgo.Observable

	// eventSequence is an atomic counter for event sequence numbers.
	eventSequence atomic.Uint64

	// supervisorStdoutReader is the user-facing reader for continuous stdout.
	supervisorStdoutReader *io.PipeReader

	// supervisorStdoutWriter is the internal writer for continuous stdout.
	supervisorStdoutWriter *io.PipeWriter

	// supervisorStderrReader is the user-facing reader for continuous stderr.
	supervisorStderrReader *io.PipeReader

	// supervisorStderrWriter is the internal writer for continuous stderr.
	supervisorStderrWriter *io.PipeWriter

	// supervisorStdinReader is the internal reader for continuous stdin.
	supervisorStdinReader *io.PipeReader

	// supervisorStdinWriter is the user-facing writer for continuous stdin.
	supervisorStdinWriter *io.PipeWriter

	// stdinMutex protects stdin routing operations during reload.
	stdinMutex sync.Mutex

	// stdinCancelFunc cancels the current stdin routing goroutine.
	stdinCancelFunc context.CancelFunc

	// supervisorContext is the overall context for the supervisor lifecycle.
	supervisorContext context.Context

	// supervisorCancelFunc cancels the supervisor context.
	supervisorCancelFunc context.CancelFunc

	// waitGroup tracks active goroutines for clean shutdown.
	waitGroup sync.WaitGroup

	// logger for printing notifications.
	logger *slog.Logger
}

// shouldResetRestartCounter determines whether the crash restart attempt counter should be reset
// based on the time since the last crash and the configured reset window.
//
// If lastCrashTime is zero (no previous crash) or enough time has passed since the last crash
// (greater than or equal to resetWindow), the counter should be reset.
func shouldResetRestartCounter(lastCrashTime time.Time, resetWindow time.Duration) bool {
	if lastCrashTime.IsZero() {
		return true
	}

	if resetWindow <= 0 {
		return false
	}

	timeSinceLastCrash := time.Since(lastCrashTime)
	return timeSinceLastCrash >= resetWindow
}

// calculateBackoffDelay computes the delay before attempting another restart based on
// the restart attempt number and the configured backoff strategy.
//
// The attempt parameter is the number of consecutive restart attempts (0-indexed).
// For attempt 0, the delay is policy.InitialDelay.
// For subsequent attempts, the delay grows according to the strategy, capped at policy.MaxDelay.
func calculateBackoffDelay(attempt int, policy RestartPolicy) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	// First attempt uses initial delay
	if attempt == 0 {
		return policy.InitialDelay
	}

	var delay time.Duration

	switch policy.Strategy {
	case StrategyExponential:
		// Exponential: initialDelay * 2^attempt
		// attempt 0: 1x, attempt 1: 2x, attempt 2: 4x, attempt 3: 8x
		// Protect against overflow by capping attempt at 30 (2^30 = ~1 billion seconds = ~31 years)
		if attempt > 30 {
			attempt = 30
		}
		multiplier := 1 << uint(attempt) // 2^attempt
		delay = policy.InitialDelay * time.Duration(multiplier)

	case StrategyLinear:
		// Linear: initialDelay * (attempt + 1)
		// attempt 0: 1x, attempt 1: 2x, attempt 2: 3x, attempt 3: 4x
		delay = policy.InitialDelay * time.Duration(attempt+1)

	case StrategyConstant:
		// Constant: always initialDelay
		delay = policy.InitialDelay

	default:
		// Unknown strategy defaults to constant
		delay = policy.InitialDelay
	}

	// Cap at max delay
	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		return policy.MaxDelay
	}

	return delay
}

// SuperviseLocal creates a new local Supervisor instance for managing a process according to the given specification
// and restart policy. The supervisor starts in the Idle state and must be explicitly started using the Start method.
//
// The restartPolicy parameter configures automatic restart behavior when the process crashes unexpectedly.
// If policy.Enabled is false, the process will not be automatically restarted after crashes.
//
// The returned Supervisor owns IO pipes (Stdout, Stderr, Stdin) that remain open across the supervisor's
// lifecycle, including process reloads. These pipes should be closed by calling Terminate when supervision is
// no longer needed.
//
// Optional configuration can be provided via LocalSupervisorOpt functions (e.g., WithLogger).
func SuperviseLocal(specification Specification, restartPolicy RestartPolicy, options ...LocalSupervisorOpt) Supervisor {
	supervisorContext, supervisorCancelFunc := context.WithCancel(context.Background())

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()
	stdinReader, stdinWriter := io.Pipe()

	// Create event channel with buffer to prevent blocking
	eventChannel := make(chan rxgo.Item, 100)
	eventObservable := rxgo.FromEventSource(eventChannel)

	supervisor := &supervisorLocal{
		specification:          specification,
		restartPolicy:          restartPolicy,
		state:                  StateIdle,
		eventChannel:           eventChannel,
		eventObservable:        eventObservable,
		supervisorStdoutReader: stdoutReader,
		supervisorStdoutWriter: stdoutWriter,
		supervisorStderrReader: stderrReader,
		supervisorStderrWriter: stderrWriter,
		supervisorStdinReader:  stdinReader,
		supervisorStdinWriter:  stdinWriter,
		supervisorContext:      supervisorContext,
		supervisorCancelFunc:   supervisorCancelFunc,
		logger:                 slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Apply options
	for _, option := range options {
		option(supervisor)
	}

	return supervisor
}

// Specification returns the initial specification provided to the supervisor.
func (supervisor *supervisorLocal) Specification() Specification {
	return supervisor.specification
}

// Stdout returns a reader for the process's standard output stream.
// This stream remains continuous across process reloads and only closes when Stop is called.
func (supervisor *supervisorLocal) Stdout() io.ReadCloser {
	return supervisor.supervisorStdoutReader
}

// Stderr returns a reader for the process's standard error stream.
// This stream remains continuous across process reloads and only closes when Stop is called.
func (supervisor *supervisorLocal) Stderr() io.ReadCloser {
	return supervisor.supervisorStderrReader
}

// Stdin returns a writer for the process's standard input stream.
// Writes are routed to the currently running process instance.
// During a reload, writes may block temporarily until the new process is ready.
func (supervisor *supervisorLocal) Stdin() io.WriteCloser {
	return supervisor.supervisorStdinWriter
}

// Events returns an observable stream of supervisor events.
// Events are emitted with monotonically increasing sequence numbers.
// If the event buffer overflows, the oldest events are dropped.
func (supervisor *supervisorLocal) Events() rxgo.Observable {
	return supervisor.eventObservable
}

// publishEvent sends an event to the event channel with proper sequencing.
// Events are published non-blocking; if the buffer is full, the oldest event is dropped.
func (supervisor *supervisorLocal) publishEvent(event Event) {
	// Non-blocking send to the channel using select with default
	select {
	case supervisor.eventChannel <- rxgo.Of(event):
		// Event sent successfully
	default:
		// Channel full - drop event (this implements drop-oldest behavior implicitly
		// since the channel is buffered and will naturally drop old events)
	}
}

// publishStartedEvent publishes a StartedEvent indicating the process has started and stabilized.
func (supervisor *supervisorLocal) publishStartedEvent(pid int) {
	sequence := supervisor.eventSequence.Add(1)
	event := StartedEvent{
		BaseEvent: BaseEvent{
			Kind: EventStarted,
			Time: time.Now(),
			Seq:  sequence,
		},
		PID: pid,
	}
	supervisor.publishEvent(event)
}

// publishStoppedEvent publishes a StoppedEvent indicating the process has stopped gracefully.
func (supervisor *supervisorLocal) publishStoppedEvent(pid int) {
	sequence := supervisor.eventSequence.Add(1)
	event := StoppedEvent{
		BaseEvent: BaseEvent{
			Kind: EventStopped,
			Time: time.Now(),
			Seq:  sequence,
		},
		PID: pid,
	}
	supervisor.publishEvent(event)
}

// publishRestartedEvent publishes a RestartedEvent indicating the process has been restarted.
func (supervisor *supervisorLocal) publishRestartedEvent(oldPID, newPID int, cause RestartCause) uint64 {
	sequence := supervisor.eventSequence.Add(1)
	event := RestartedEvent{
		BaseEvent: BaseEvent{
			Kind: EventRestarted,
			Time: time.Now(),
			Seq:  sequence,
		},
		OldPID:       oldPID,
		NewPID:       newPID,
		RestartCount: supervisor.restartCount,
		Cause:        cause,
	}
	supervisor.publishEvent(event)
	return sequence
}

// publishReloadedEvent publishes a ReloadedEvent indicating a reload cycle has completed successfully.
func (supervisor *supervisorLocal) publishReloadedEvent(oldPID, newPID int, relatedSeq uint64) {
	sequence := supervisor.eventSequence.Add(1)
	event := ReloadedEvent{
		BaseEvent: BaseEvent{
			Kind: EventReloaded,
			Time: time.Now(),
			Seq:  sequence,
		},
		OldPID:     oldPID,
		NewPID:     newPID,
		RelatedSeq: relatedSeq,
	}
	supervisor.publishEvent(event)
}

// publishKilledEvent publishes a KilledEvent indicating the process was forcibly terminated.
func (supervisor *supervisorLocal) publishKilledEvent(pid int) {
	sequence := supervisor.eventSequence.Add(1)
	event := KilledEvent{
		BaseEvent: BaseEvent{
			Kind: EventKilled,
			Time: time.Now(),
			Seq:  sequence,
		},
		PID: pid,
	}
	supervisor.publishEvent(event)
}

// publishExitedEvent publishes an ExitedEvent indicating the process has exited.
func (supervisor *supervisorLocal) publishExitedEvent(pid int, exitCode *int, signalName *string) {
	sequence := supervisor.eventSequence.Add(1)
	event := ExitedEvent{
		BaseEvent: BaseEvent{
			Kind: EventExited,
			Time: time.Now(),
			Seq:  sequence,
		},
		PID:        pid,
		ExitCode:   exitCode,
		SignalName: signalName,
	}
	supervisor.publishEvent(event)
}

// createCommand constructs an exec.Cmd from the given specification.
// It expands paths with ~ and sets up environment variables and working directory.
// Returns an error if path expansion fails.
func createCommand(specification Specification) (*exec.Cmd, error) {
	expandedExecutablePath, err := homedir.Expand(specification.ExecutablePath)
	if err != nil {
		return nil, fmt.Errorf("expand executable path: %w", err)
	}

	command := exec.Command(expandedExecutablePath, specification.Arguments...)

	// Set up environment variables if provided
	if len(specification.EnvironmentVariables) > 0 {
		environment := os.Environ()
		for key, value := range specification.EnvironmentVariables {
			environment = append(environment, fmt.Sprintf("%s=%s", key, value))
		}
		command.Env = environment
	}

	// Set working directory if provided
	if specification.WorkingDirectory != nil {
		expandedWorkingDirectory, err := homedir.Expand(*specification.WorkingDirectory)
		if err != nil {
			return nil, fmt.Errorf("expand working directory: %w", err)
		}
		command.Dir = expandedWorkingDirectory
	}

	return command, nil
}

// launchProcess creates and starts a new process instance from the specification.
// It sets up pipes for stdin, stdout, and stderr, and starts monitoring the process.
// Returns the processInstance or an error if launch fails.
func (supervisor *supervisorLocal) launchProcess(ctx context.Context) (*processInstance, error) {
	cmdInfo := fmt.Sprintf("[%s %s]",
		supervisor.specification.ExecutablePath,
		strings.Join(supervisor.specification.Arguments, " "))

	command, err := createCommand(supervisor.specification)
	if err != nil {
		supervisor.logger.Error("Failed to create command.",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("create command %s: %w", cmdInfo, err)
	}

	// Create pipes for process IO
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		supervisor.logger.Error("Failed to create stdout pipe.",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("create stdout pipe %s: %w", cmdInfo, err)
	}

	stderrPipe, err := command.StderrPipe()
	if err != nil {
		supervisor.logger.Error("Failed to create stderr pipe.",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("create stderr pipe %s: %w", cmdInfo, err)
	}

	stdinPipe, err := command.StdinPipe()
	if err != nil {
		supervisor.logger.Error("Failed to create stdin pipe.",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("create stdin pipe %s: %w", cmdInfo, err)
	}

	// Start the process
	if err := command.Start(); err != nil {
		supervisor.logger.Error("Failed to start process.",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("start process %s: %w", cmdInfo, err)
	}

	instanceContext, instanceCancelFunc := context.WithCancel(ctx)

	instance := &processInstance{
		command:     command,
		pid:         command.Process.Pid,
		startedAt:   time.Now(),
		exitChannel: make(chan *ExitError, 1),
		context:     instanceContext,
		cancel:      instanceCancelFunc,
		stdoutPipe:  stdoutPipe,
		stderrPipe:  stderrPipe,
		stdinPipe:   stdinPipe,
	}

	supervisor.logger.Info("Process launched.",
		slog.Int("pid", instance.pid))

	return instance, nil
}

// monitorProcessExit waits for the process to exit and publishes an ExitedEvent.
// It extracts the exit code or signal and sends an ExitError to the instance's exitChannel.
// If auto-restart is enabled and the process crashed, it attempts to restart it automatically.
// This function should be run as a goroutine.
func (supervisor *supervisorLocal) monitorProcessExit(instance *processInstance) {
	defer supervisor.waitGroup.Done()
	defer close(instance.exitChannel)

	// Wait for the process to exit
	waitErr := instance.command.Wait()

	exitTime := time.Now()
	var exitCode *int
	var signalName *string

	if waitErr != nil {
		// Process exited with error or signal
		if exitError, ok := waitErr.(*exec.ExitError); ok {
			code := exitError.ExitCode()
			exitCode = &code

			// Check if terminated by signal (Unix-specific)
			signalName = extractSignalName(exitError.ProcessState)
		}
	} else {
		// Process exited successfully (code 0)
		code := 0
		exitCode = &code
	}

	// Log process exit
	logAttrs := []any{slog.Int("pid", instance.pid)}
	if exitCode != nil {
		logAttrs = append(logAttrs, slog.Int("exitCode", *exitCode))
	}
	if signalName != nil {
		logAttrs = append(logAttrs, slog.String("signal", *signalName))
	}
	supervisor.logger.Info("Process exited.", logAttrs...)

	// Publish ExitedEvent
	supervisor.publishExitedEvent(instance.pid, exitCode, signalName)

	// Create ExitError and send to exit channel
	exitErrorObj := &ExitError{
		PID:  instance.pid,
		When: exitTime,
	}

	if exitCode != nil {
		exitErrorObj.Code = *exitCode
	} else {
		exitErrorObj.Code = -1
	}

	if signalName != nil {
		exitErrorObj.Signal = *signalName
	}

	// Send exit error to channel (non-blocking since channel has buffer of 1)
	instance.exitChannel <- exitErrorObj

	// Check if this was a crash and auto-restart is enabled
	supervisor.stateMutex.Lock()
	currentState := supervisor.state
	isCrash := currentState == StateRunning
	restartEnabled := supervisor.restartPolicy.Enabled
	supervisor.stateMutex.Unlock()

	if !isCrash || !restartEnabled {
		// Not a crash or auto-restart disabled - nothing more to do
		if isCrash {
			supervisor.logger.Info("Process crashed but auto-restart is disabled.",
				slog.Int("pid", instance.pid))
		}
		return
	}

	// This was a crash and auto-restart is enabled
	// Transition to StateRestarting
	supervisor.stateMutex.Lock()
	supervisor.state = StateRestarting
	oldPID := instance.pid
	supervisor.stateMutex.Unlock()

	supervisor.logger.Warn("Process crashed, initiating auto-restart.",
		slog.Int("pid", oldPID),
		slog.Bool("restartEnabled", restartEnabled))

	// Auto-restart loop with backoff
	stableFor := 100 * time.Millisecond // Use a default stability period for auto-restart

	for {
		// Check if we should reset the attempt counter
		supervisor.stateMutex.Lock()
		if shouldResetRestartCounter(supervisor.lastCrashTime, supervisor.restartPolicy.ResetWindow) {
			supervisor.crashRestartAttempts = 0
			supervisor.logger.Info("Restart attempt counter reset.",
				slog.Duration("resetWindow", supervisor.restartPolicy.ResetWindow))
		}
		supervisor.lastCrashTime = time.Now()

		currentAttempt := supervisor.crashRestartAttempts
		maxAttempts := supervisor.restartPolicy.MaxAttempts
		supervisor.stateMutex.Unlock()

		// Check if we've exceeded max attempts
		if maxAttempts > 0 && currentAttempt >= maxAttempts {
			// Max attempts exceeded - give up and transition to Idle
			supervisor.stateMutex.Lock()
			supervisor.state = StateIdle
			supervisor.currentInstance = nil
			supervisor.stateMutex.Unlock()

			supervisor.logger.Error("Max restart attempts exceeded, giving up.",
				slog.Int("maxAttempts", maxAttempts),
				slog.Int("currentAttempt", currentAttempt))
			return
		}

		// Calculate backoff delay
		delay := calculateBackoffDelay(currentAttempt, supervisor.restartPolicy)

		supervisor.logger.Info("Waiting before restart attempt.",
			slog.Int("attempt", currentAttempt+1),
			slog.Duration("delay", delay),
			slog.String("strategy", string(supervisor.restartPolicy.Strategy)))

		// Sleep with backoff (interruptible by supervisor context)
		select {
		case <-time.After(delay):
			// Backoff complete, proceed with restart
		case <-supervisor.supervisorContext.Done():
			// Supervisor is shutting down
			supervisor.stateMutex.Lock()
			supervisor.state = StateIdle
			supervisor.currentInstance = nil
			supervisor.stateMutex.Unlock()
			supervisor.logger.Info("Auto-restart canceled due to supervisor shutdown.")
			return
		}

		// Attempt restart
		supervisor.logger.Info("Attempting auto-restart.",
			slog.Int("attempt", currentAttempt+1),
			slog.Int("oldPid", oldPID))

		restartCtx, restartCancel := context.WithTimeout(supervisor.supervisorContext, 30*time.Second)
		restartErr := supervisor.attemptAutoRestart(restartCtx, oldPID, stableFor)
		restartCancel()

		if restartErr == nil {
			// Restart succeeded - reset attempt counter for next crash
			supervisor.stateMutex.Lock()
			supervisor.crashRestartAttempts = 0
			supervisor.stateMutex.Unlock()
			supervisor.logger.Info("Auto-restart succeeded.",
				slog.Int("attempt", currentAttempt+1))
			return
		}

		// Restart failed - increment attempt counter and try again
		supervisor.logger.Warn("Auto-restart attempt failed.",
			slog.Int("attempt", currentAttempt+1),
			slog.String("error", restartErr.Error()))

		supervisor.stateMutex.Lock()
		supervisor.crashRestartAttempts++
		supervisor.stateMutex.Unlock()
	}
}

// copyProcessStdout copies data from the process's stdout to the supervisor's stdout writer.
// This enables continuous stdout stream across process reloads.
// This function should be run as a goroutine.
func (supervisor *supervisorLocal) copyProcessStdout(instance *processInstance) {
	defer supervisor.waitGroup.Done()

	_, copyErr := io.Copy(supervisor.supervisorStdoutWriter, instance.stdoutPipe)
	if copyErr != nil && copyErr != io.EOF {
		// Log error but don't terminate - the process may still be running
		// In a production system, you might want to log this with slog
		_ = copyErr
	}
}

// copyProcessStderr copies data from the process's stderr to the supervisor's stderr writer.
// This enables continuous stderr stream across process reloads.
// This function should be run as a goroutine.
func (supervisor *supervisorLocal) copyProcessStderr(instance *processInstance) {
	defer supervisor.waitGroup.Done()

	_, copyErr := io.Copy(supervisor.supervisorStderrWriter, instance.stderrPipe)
	if copyErr != nil && copyErr != io.EOF {
		// Log error but don't terminate - the process may still be running
		// In a production system, you might want to log this with slog
		_ = copyErr
	}
}

// routeStdinToProcess copies data from the supervisor's stdin reader to the process's stdin pipe.
// This routing can be canceled and switched to a different process during reload.
// This function should be run as a goroutine.
func (supervisor *supervisorLocal) routeStdinToProcess(instance *processInstance) {
	defer supervisor.waitGroup.Done()

	// Create a context-aware reader that will stop on cancellation
	doneChan := make(chan struct{})
	defer close(doneChan)

	go func() {
		select {
		case <-instance.context.Done():
			// Close the process stdin pipe to unblock any pending writes
			_ = instance.stdinPipe.Close()
		case <-doneChan:
			// Normal completion
		}
	}()

	_, copyErr := io.Copy(instance.stdinPipe, supervisor.supervisorStdinReader)
	if copyErr != nil && copyErr != io.EOF {
		// Stdin routing error - this can happen during process reload
		_ = copyErr
	}
}

// switchStdinTarget cancels the current stdin routing and starts routing to a new process.
// This is called during reload to redirect stdin to the new process instance.
func (supervisor *supervisorLocal) switchStdinTarget(newInstance *processInstance) {
	supervisor.stdinMutex.Lock()
	defer supervisor.stdinMutex.Unlock()

	supervisor.logger.Info("Switching stdin routing to new process.",
		slog.Int("newPid", newInstance.pid))

	// Cancel the old stdin routing if it exists
	if supervisor.stdinCancelFunc != nil {
		supervisor.stdinCancelFunc()
	}

	// Create new stdin pipes for the supervisor side
	newStdinReader, newStdinWriter := io.Pipe()

	// Close old writer to unblock any pending writes
	_ = supervisor.supervisorStdinWriter.Close()

	// Update supervisor's stdin pipes
	supervisor.supervisorStdinReader = newStdinReader
	supervisor.supervisorStdinWriter = newStdinWriter

	// Start routing to the new instance
	supervisor.waitGroup.Add(1)
	go supervisor.routeStdinToProcess(newInstance)

	supervisor.logger.Info("Stdin routing switched successfully.",
		slog.Int("newPid", newInstance.pid))
}

// Start launches the process and waits until it stays stable for at least stableFor duration.
// If the process exits before stability is reached, Start returns ErrFailedToStart wrapping the cause.
// Start can only be called when the supervisor is in the Idle state.
func (supervisor *supervisorLocal) Start(ctx context.Context, stableFor time.Duration) error {
	cmdInfo := fmt.Sprintf("[%s %s]",
		supervisor.specification.ExecutablePath,
		strings.Join(supervisor.specification.Arguments, " "))

	supervisor.logger.Info("Starting process.",
		slog.String("executablePath", supervisor.specification.ExecutablePath),
		slog.Duration("stableFor", stableFor))

	// State transition guard: Idle → Starting
	supervisor.stateMutex.Lock()
	if supervisor.state != StateIdle {
		currentState := supervisor.state
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to start process: invalid state.",
			slog.String("currentState", string(currentState)))
		return fmt.Errorf("%w: cannot start from %s state", ErrInvalidState, currentState)
	}
	supervisor.state = StateStarting
	supervisor.stateMutex.Unlock()

	// Launch the process
	instance, err := supervisor.launchProcess(supervisor.supervisorContext)
	if err != nil {
		// Rollback state to Idle
		supervisor.stateMutex.Lock()
		supervisor.state = StateIdle
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to launch process.",
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %w", ErrFailedToStart, err)
	}

	// Store the instance
	supervisor.stateMutex.Lock()
	supervisor.currentInstance = instance
	supervisor.stateMutex.Unlock()

	// Start monitoring goroutines
	supervisor.waitGroup.Add(4)
	go supervisor.monitorProcessExit(instance)
	go supervisor.copyProcessStdout(instance)
	go supervisor.copyProcessStderr(instance)
	go supervisor.routeStdinToProcess(instance)

	// Wait for stability
	stabilityTimer := time.NewTimer(stableFor)
	defer stabilityTimer.Stop()

	select {
	case exitError := <-instance.exitChannel:
		// Process exited before reaching stability
		supervisor.stateMutex.Lock()
		supervisor.state = StateIdle
		supervisor.currentInstance = nil
		supervisor.stateMutex.Unlock()

		supervisor.logger.Error("Process exited before reaching stability.",
			slog.Int("pid", instance.pid),
			slog.String("error", exitError.Error()))
		return fmt.Errorf("%w %s: %w", ErrFailedToStart, cmdInfo, exitError)

	case <-ctx.Done():
		// Context canceled before stability
		supervisor.stateMutex.Lock()
		supervisor.state = StateIdle
		supervisor.stateMutex.Unlock()

		// Terminate the process
		_ = forceTerminate(instance.command.Process)
		instance.cancel()

		supervisor.logger.Warn("Process start canceled before stability.",
			slog.Int("pid", instance.pid),
			slog.String("reason", ctx.Err().Error()))
		return fmt.Errorf("%w: %w", ErrFailedToStart, ctx.Err())

	case <-stabilityTimer.C:
		// Process remained stable for the required duration
		supervisor.stateMutex.Lock()
		supervisor.state = StateRunning
		supervisor.stateMutex.Unlock()

		supervisor.logger.Info("Process started successfully.",
			slog.Int("pid", instance.pid),
			slog.Duration("stableFor", stableFor))
		supervisor.publishStartedEvent(instance.pid)
		return nil
	}
}

// Status returns a race-free snapshot of the current supervisor state.
func (supervisor *supervisorLocal) Status() Status {
	supervisor.stateMutex.RLock()
	defer supervisor.stateMutex.RUnlock()

	status := Status{
		Running:      supervisor.state == StateRunning,
		RestartCount: supervisor.restartCount,
	}

	if supervisor.currentInstance != nil {
		status.PID = supervisor.currentInstance.pid
		status.StartedAt = supervisor.currentInstance.startedAt
		status.Uptime = time.Since(supervisor.currentInstance.startedAt)
	}

	return status
}

// State returns the current lifecycle state of the supervisor.
func (supervisor *supervisorLocal) State() SupervisorState {
	supervisor.stateMutex.RLock()
	defer supervisor.stateMutex.RUnlock()
	return supervisor.state
}

// Wait blocks until the current process instance exits, returning an ExitError or nil if stopped cleanly.
// If no process is currently running, Wait returns immediately with nil.
// The context can be used to cancel the wait operation.
func (supervisor *supervisorLocal) Wait(ctx context.Context) error {
	// Get the current instance's exit channel
	supervisor.stateMutex.RLock()
	currentInstance := supervisor.currentInstance
	supervisor.stateMutex.RUnlock()

	if currentInstance == nil {
		return nil
	}

	select {
	case exitError := <-currentInstance.exitChannel:
		return exitError
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop requests a graceful termination of the process and waits for it to exit.
// After the context deadline expires, Stop performs a hard termination (SIGKILL).
// If a hard kill was required, Stop returns a KilledError.
// Stop transitions the supervisor from any state to Stopping, then to Idle.
func (supervisor *supervisorLocal) Stop(ctx context.Context) error {
	// State transition guard: any state except Idle → Stopping
	supervisor.stateMutex.Lock()
	if supervisor.state == StateIdle {
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to stop process: invalid state.",
			slog.String("currentState", string(StateIdle)))
		return fmt.Errorf("%w: cannot stop from idle state", ErrInvalidState)
	}

	currentInstance := supervisor.currentInstance
	if currentInstance == nil {
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to stop process: no process instance.")
		return fmt.Errorf("cannot stop: no process instance")
	}

	instancePID := currentInstance.pid
	supervisor.state = StateStopping
	supervisor.stateMutex.Unlock()

	supervisor.logger.Info("Stopping process.",
		slog.Int("pid", instancePID))

	wasKilled := false

	// Attempt graceful termination
	if err := gracefulTerminate(currentInstance.command.Process); err != nil {
		// Process may have already exited, continue
		_ = err
	}

	// Wait for exit or context deadline
	select {
	case <-currentInstance.exitChannel:
		// Process exited gracefully
		// exitChannel will be closed after ExitedEvent is published
		supervisor.logger.Info("Process exited gracefully.",
			slog.Int("pid", instancePID))

	case <-ctx.Done():
		// Deadline exceeded, force terminate
		supervisor.logger.Warn("Graceful shutdown timeout, forcing termination.",
			slog.Int("pid", instancePID))

		if err := forceTerminate(currentInstance.command.Process); err != nil {
			// Process may have already exited
			_ = err
		}

		// Wait briefly for forced termination
		forceWaitTimer := time.NewTimer(5 * time.Second)
		select {
		case <-currentInstance.exitChannel:
			// Process exited after force kill
			supervisor.logger.Info("Process terminated forcefully.",
				slog.Int("pid", instancePID))
		case <-forceWaitTimer.C:
			// Even force kill didn't work (very rare)
			supervisor.logger.Error("Force termination failed.",
				slog.Int("pid", instancePID))
		}
		forceWaitTimer.Stop()

		wasKilled = true
		supervisor.publishKilledEvent(instancePID)
	}

	// Cancel instance context to stop all goroutines
	currentInstance.cancel()

	// Cleanup: close all pipes and cancel supervisor context
	_ = supervisor.supervisorStdoutWriter.Close()
	_ = supervisor.supervisorStderrWriter.Close()
	_ = supervisor.supervisorStdinWriter.Close()
	supervisor.supervisorCancelFunc()

	// Wait for all goroutines to finish
	supervisor.waitGroup.Wait()

	// Transition to Idle
	supervisor.stateMutex.Lock()
	supervisor.state = StateIdle
	supervisor.currentInstance = nil
	supervisor.stateMutex.Unlock()

	// Publish StoppedEvent if not killed
	if !wasKilled {
		supervisor.publishStoppedEvent(instancePID)
	}

	supervisor.logger.Info("Process stopped successfully.",
		slog.Int("pid", instancePID),
		slog.Bool("wasKilled", wasKilled))

	// Return KilledError if force termination was required
	if wasKilled {
		return &KilledError{
			PID:   instancePID,
			After: 0, // We could track actual time if needed
		}
	}

	return nil
}

// attemptAutoRestart attempts to restart a crashed process automatically.
// This method assumes the old process has already exited and only needs to launch a new instance.
// It returns nil on success or an error if the restart failed.
//
// Unlike Reload, this method:
// - Does not stop the old process (it's already dead)
// - Publishes RestartedEvent with CauseCrash upon success
// - Does not publish ReloadedEvent (only for manual reload operations)
//
// The caller must hold appropriate locks and manage state transitions.
func (supervisor *supervisorLocal) attemptAutoRestart(ctx context.Context, oldPID int, stableFor time.Duration) error {
	cmdInfo := fmt.Sprintf("[%s %s]",
		supervisor.specification.ExecutablePath,
		strings.Join(supervisor.specification.Arguments, " "))

	supervisor.logger.Info("Launching new process for auto-restart.",
		slog.Int("oldPid", oldPID))

	// Launch new process
	newInstance, err := supervisor.launchProcess(supervisor.supervisorContext)
	if err != nil {
		supervisor.logger.Error("Failed to launch new process for auto-restart.",
			slog.String("error", err.Error()))
		return fmt.Errorf("launch process: %w", err)
	}

	supervisor.logger.Info("New process launched successfully.",
		slog.Int("newPid", newInstance.pid))

	// Start monitoring goroutines for new instance
	supervisor.waitGroup.Add(3) // exit monitor, stdout copy, stderr copy (stdin will be handled by switchStdinTarget)
	go supervisor.monitorProcessExit(newInstance)
	go supervisor.copyProcessStdout(newInstance)
	go supervisor.copyProcessStderr(newInstance)

	// Switch stdin routing to new process
	supervisor.switchStdinTarget(newInstance)

	// Wait for new process stability
	stabilityTimer := time.NewTimer(stableFor)
	defer stabilityTimer.Stop()

	select {
	case exitError := <-newInstance.exitChannel:
		// New process exited before reaching stability
		newInstance.cancel()
		supervisor.logger.Error("New process for auto-restart exited before stability.",
			slog.Int("newPid", newInstance.pid),
			slog.String("error", exitError.Error()))
		return fmt.Errorf("new process %s exited before stability: %w", cmdInfo, exitError)

	case <-ctx.Done():
		// Context canceled before stability
		_ = forceTerminate(newInstance.command.Process)
		newInstance.cancel()
		supervisor.logger.Warn("Auto-restart context canceled before stability.",
			slog.Int("newPid", newInstance.pid))
		return fmt.Errorf("context canceled: %w", ctx.Err())

	case <-stabilityTimer.C:
		// New process is stable, proceed with restart
	}

	// Update supervisor state
	supervisor.stateMutex.Lock()
	supervisor.restartCount++
	supervisor.currentInstance = newInstance
	supervisor.state = StateRunning
	supervisor.stateMutex.Unlock()

	// Publish RestartedEvent (no ReloadedEvent for auto-restart)
	supervisor.publishRestartedEvent(oldPID, newInstance.pid, CauseCrash)

	supervisor.logger.Info("Auto-restart completed successfully.",
		slog.Int("oldPid", oldPID),
		slog.Int("newPid", newInstance.pid),
		slog.Int("restartCount", supervisor.restartCount))

	return nil
}

// Reload performs a graceful restart of the process with the same specification.
// It stops the old process first, then starts a new process and waits for stability.
// Reload publishes RestartedEvent followed by ReloadedEvent upon success.
func (supervisor *supervisorLocal) Reload(ctx context.Context, stableFor time.Duration) error {
	cmdInfo := fmt.Sprintf("[%s %s]",
		supervisor.specification.ExecutablePath,
		strings.Join(supervisor.specification.Arguments, " "))

	// State transition guard: Running → Reloading
	supervisor.stateMutex.Lock()
	if supervisor.state != StateRunning {
		currentState := supervisor.state
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to reload process: invalid state.",
			slog.String("currentState", string(currentState)))
		return fmt.Errorf("%w: cannot reload from %s state", ErrInvalidState, currentState)
	}

	oldInstance := supervisor.currentInstance
	if oldInstance == nil {
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to reload process: no process instance.")
		return fmt.Errorf("cannot reload: no current process instance")
	}

	oldInstancePID := oldInstance.pid
	supervisor.state = StateReloading
	supervisor.stateMutex.Unlock()

	supervisor.logger.Info("Reloading process.",
		slog.Int("oldPid", oldInstancePID),
		slog.Duration("stableFor", stableFor))

	// Terminate old process first
	supervisor.logger.Info("Stopping old process.",
		slog.Int("pid", oldInstancePID))

	if err := gracefulTerminate(oldInstance.command.Process); err != nil {
		// Old process may have already exited
		_ = err
	}

	// Wait for old process to exit (with timeout)
	oldProcessWaitTimer := time.NewTimer(10 * time.Second)
	select {
	case <-oldInstance.exitChannel:
		// Old process exited gracefully
		supervisor.logger.Info("Old process exited gracefully.",
			slog.Int("pid", oldInstancePID))
	case <-oldProcessWaitTimer.C:
		// Timeout - force terminate old process
		supervisor.logger.Warn("Old process did not exit gracefully, forcing termination.",
			slog.Int("pid", oldInstancePID))
		_ = forceTerminate(oldInstance.command.Process)
		// Wait a bit more for forced termination
		time.Sleep(2 * time.Second)
	}
	oldProcessWaitTimer.Stop()

	// Cancel old instance context
	oldInstance.cancel()

	// Clear current instance - old process is gone
	supervisor.stateMutex.Lock()
	supervisor.currentInstance = nil
	supervisor.stateMutex.Unlock()

	// Launch new process
	supervisor.logger.Info("Launching new process.")
	newInstance, err := supervisor.launchProcess(supervisor.supervisorContext)
	if err != nil {
		// Old process is already stopped, cannot rollback
		supervisor.stateMutex.Lock()
		supervisor.state = StateIdle
		supervisor.stateMutex.Unlock()
		supervisor.logger.Error("Failed to launch new process.",
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %w", ErrFailedToStart, err)
	}

	// Start monitoring goroutines for new instance
	supervisor.waitGroup.Add(3) // exit monitor, stdout copy, stderr copy (stdin will be handled by switchStdinTarget)
	go supervisor.monitorProcessExit(newInstance)
	go supervisor.copyProcessStdout(newInstance)
	go supervisor.copyProcessStderr(newInstance)

	// Switch stdin routing to new process
	supervisor.switchStdinTarget(newInstance)

	// Wait for new process stability
	stabilityTimer := time.NewTimer(stableFor)
	defer stabilityTimer.Stop()

	select {
	case exitError := <-newInstance.exitChannel:
		// New process exited before reaching stability
		newInstance.cancel()

		supervisor.stateMutex.Lock()
		supervisor.state = StateIdle
		supervisor.stateMutex.Unlock()

		supervisor.logger.Error("New process exited before reaching stability.",
			slog.Int("pid", newInstance.pid),
			slog.String("error", exitError.Error()))
		return fmt.Errorf("%w: new process %s exited: %w", ErrFailedToStart, cmdInfo, exitError)

	case <-ctx.Done():
		// Context canceled before stability
		_ = forceTerminate(newInstance.command.Process)
		newInstance.cancel()

		supervisor.stateMutex.Lock()
		supervisor.state = StateIdle
		supervisor.stateMutex.Unlock()

		supervisor.logger.Warn("Reload canceled before new process reached stability.",
			slog.Int("pid", newInstance.pid),
			slog.String("reason", ctx.Err().Error()))
		return fmt.Errorf("%w: context canceled: %w", ErrFailedToStart, ctx.Err())

	case <-stabilityTimer.C:
		// New process is stable, proceed with reload
	}

	// Increment restart count
	supervisor.stateMutex.Lock()
	supervisor.restartCount++
	supervisor.currentInstance = newInstance
	supervisor.state = StateRunning
	supervisor.stateMutex.Unlock()

	// Publish RestartedEvent
	restartedEventSeq := supervisor.publishRestartedEvent(oldInstancePID, newInstance.pid, CauseManual)

	// Publish ReloadedEvent
	supervisor.publishReloadedEvent(oldInstancePID, newInstance.pid, restartedEventSeq)

	supervisor.logger.Info("Process reloaded successfully.",
		slog.Int("oldPid", oldInstancePID),
		slog.Int("newPid", newInstance.pid),
		slog.Int("restartCount", supervisor.restartCount))

	return nil
}

// ReloadWithSpecification performs a graceful restart with an updated process specification.
// Only Arguments, EnvironmentVariables, and WorkingDirectory can be changed.
// The ExecutablePath must remain the same, otherwise returns ErrCannotChangeExecutable.
// The old process is stopped first, then the new process is started with the updated specification.
func (supervisor *supervisorLocal) ReloadWithSpecification(
	ctx context.Context,
	newSpecification Specification,
	stableFor time.Duration,
) error {
	supervisor.logger.Info("Reloading process with updated specification.")

	// Validate ExecutablePath hasn't changed
	if newSpecification.ExecutablePath != supervisor.specification.ExecutablePath {
		supervisor.logger.Error("Cannot change executable path during reload.",
			slog.String("oldPath", supervisor.specification.ExecutablePath),
			slog.String("newPath", newSpecification.ExecutablePath))
		return fmt.Errorf("%w: cannot change from %s to %s",
			ErrCannotChangeExecutable,
			supervisor.specification.ExecutablePath,
			newSpecification.ExecutablePath)
	}

	// Update specification
	supervisor.stateMutex.Lock()
	supervisor.specification = newSpecification
	supervisor.stateMutex.Unlock()

	supervisor.logger.Info("Specification updated, proceeding with reload.")

	// Call existing Reload() which handles the control
	return supervisor.Reload(ctx, stableFor)
}
