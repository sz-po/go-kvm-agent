package process

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	rxgo "github.com/reactivex/rxgo/v2"
)

// Specification defines the configuration used to launch and manage a process.
// It contains the executable path, arguments, environment variables, and working directory.
// This structure provides the necessary data to construct an os/exec.Cmd or equivalent.
type Specification struct {
	// ExecutablePath specifies the absolute or relative path to the executable binary
	// that should be started and supervised.
	//
	// The path may include a leading '~', which will be expanded to the user's home directory
	// before process startup.
	ExecutablePath string `json:"executablePath"`

	// Arguments defines the list of command-line arguments passed to the executable.
	// The executable itself should not be included as the first argument; it will be
	// automatically handled by the supervisor when starting the process.
	Arguments []string `json:"arguments,omitempty"`

	// EnvironmentVariables defines a set of environment variables that should be available
	// to the process. The map key represents the variable name, and the value represents
	// its assigned value. If nil or empty, the process inherits the parent environment.
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`

	// WorkingDirectory specifies the directory in which the process will start.
	// If nil, the process inherits the current working directory of the supervisor.
	//
	// The path may also include a leading '~', which will be expanded to the user's home directory.
	WorkingDirectory *string `json:"workingDirectory,omitempty"`
}

// EventKind enumerates event types emitted by the Supervisor.
type EventKind string

const (
	EventStarted   EventKind = "started"
	EventStopped   EventKind = "stopped"
	EventRestarted EventKind = "restarted"
	EventReloaded  EventKind = "reloaded"
	EventKilled    EventKind = "killed"
	EventExited    EventKind = "exited"
)

// SupervisorState represents the current lifecycle state of the supervisor.
type SupervisorState string

const (
	// StateUnknown represents an uninitialized or invalid state.
	StateUnknown SupervisorState = "unknown"
	// StateIdle means the supervisor exists but no process is running.
	StateIdle SupervisorState = "idle"
	// StateStarting means a process is being launched and stabilized.
	StateStarting SupervisorState = "starting"
	// StateRunning means a process is running and stable.
	StateRunning SupervisorState = "running"
	// StateStopping means a process is being gracefully stopped.
	StateStopping SupervisorState = "stopping"
	// StateReloading means a new process is being started to replace the current one.
	StateReloading SupervisorState = "reloading"
	// StateRestarting means the supervisor is automatically restarting a crashed process.
	StateRestarting SupervisorState = "restarting"
)

// BaseEvent carries common metadata for all events.
type BaseEvent struct {
	Kind EventKind `json:"kind"`
	Time time.Time `json:"time"`
	Seq  uint64    `json:"seq"`
}

// Event is the common interface for all Supervisor events.
type Event interface {
	// Kind returns the event kind.
	Kind() EventKind
}

// StartedEvent is emitted after the process has started successfully
// and remained stable for the requested stable period.
type StartedEvent struct {
	BaseEvent
	PID int `json:"pid"`
}

func (event StartedEvent) Kind() EventKind { return EventStarted }

// StoppedEvent is emitted when the process has been stopped gracefully,
// even if a hard kill was ultimately required.
type StoppedEvent struct {
	BaseEvent
	PID int `json:"pid"`
}

func (event StoppedEvent) Kind() EventKind { return EventStopped }

// RestartCause describes why a restart happened.
type RestartCause string

const (
	CauseManual RestartCause = "manual" // explicit Reload/Restart request
	CauseCrash  RestartCause = "crash"  // process died unexpectedly
	CausePolicy RestartCause = "policy" // restart policy/backoff triggered
)

// RestartStrategy defines how backoff delays are calculated between restart attempts.
type RestartStrategy string

const (
	// StrategyUnknown represents an uninitialized or invalid strategy.
	StrategyUnknown RestartStrategy = "unknown"
	// StrategyExponential increases delay exponentially (1s, 2s, 4s, 8s, ...).
	StrategyExponential RestartStrategy = "exponential"
	// StrategyLinear increases delay linearly (1s, 2s, 3s, 4s, ...).
	StrategyLinear RestartStrategy = "linear"
	// StrategyConstant uses a fixed delay between all restart attempts.
	StrategyConstant RestartStrategy = "constant"
)

// RestartPolicy configures automatic restart behavior when a supervised process crashes.
// The supervisor will automatically restart a crashed process according to this policy.
type RestartPolicy struct {
	// Enabled controls whether automatic restart is active.
	// When false, processes that crash will not be automatically restarted.
	Enabled bool `json:"enabled"`

	// MaxAttempts limits the number of restart attempts.
	// Set to 0 for unlimited restart attempts.
	// The counter resets after the process runs stably for ResetWindow duration.
	MaxAttempts int `json:"maxAttempts"`

	// Strategy determines how backoff delays are calculated.
	// Valid values: "exponential", "linear", "constant".
	Strategy RestartStrategy `json:"strategy"`

	// InitialDelay is the delay before the first restart attempt.
	// This value is also used as the base delay for exponential and linear strategies.
	InitialDelay time.Duration `json:"initialDelay"`

	// MaxDelay caps the maximum backoff delay between restart attempts.
	// This prevents delays from growing unbounded in exponential/linear strategies.
	MaxDelay time.Duration `json:"maxDelay"`

	// ResetWindow is the duration a process must run stably before the restart
	// attempt counter is reset to zero. This allows recovery from transient failures.
	ResetWindow time.Duration `json:"resetWindow"`
}

// RestartedEvent is emitted when the process has been restarted.
// RestartCount is the total number of successful restarts observed so far
// (including the one that produced this event).
type RestartedEvent struct {
	BaseEvent
	OldPID       int          `json:"oldPid"`
	NewPID       int          `json:"newPid"`
	RestartCount int          `json:"restartCount"`
	Cause        RestartCause `json:"cause"`
}

func (event RestartedEvent) Kind() EventKind { return EventRestarted }

// ReloadedEvent is emitted after a successful reload cycle completes.
// It is always emitted after the corresponding RestartedEvent for the same cycle.
type ReloadedEvent struct {
	BaseEvent
	OldPID     int    `json:"oldPid"`
	NewPID     int    `json:"newPid"`
	RelatedSeq uint64 `json:"relatedSeq,omitempty"` // sequence of the RestartedEvent in this cycle
}

func (event ReloadedEvent) Kind() EventKind { return EventReloaded }

// KilledEvent is emitted when the process was forcibly terminated.
type KilledEvent struct {
	BaseEvent
	PID int `json:"pid"`
}

func (event KilledEvent) Kind() EventKind { return EventKilled }

// ExitedEvent is emitted whenever the process exits, including exits that occur during
// Start, Reload, or Terminate. Either ExitCode or SignalName will be set.
type ExitedEvent struct {
	BaseEvent
	PID        int     `json:"pid"`
	ExitCode   *int    `json:"exitCode,omitempty"`
	SignalName *string `json:"signal,omitempty"` // platform-appropriate signal name; nil on platforms without signals
}

func (event ExitedEvent) Kind() EventKind { return EventExited }

// Status is a snapshot of the current process state.
type Status struct {
	Running      bool
	PID          int
	RestartCount int
	StartedAt    time.Time
	Uptime       time.Duration
}

// Supervisor manages a single process: starting, stopping, reloading, and emitting events.
// It maintains internal process state and may restart the process if it exits.
//
// Event delivery semantics:
//   - Implementations may buffer or drop events on overflow. The drop policy MUST be documented
//     (e.g., drop-oldest). Implementations SHOULD set Seq monotonically increasing.
//   - Reload MUST publish RestartedEvent followed by ReloadedEvent for the same cycle.
//
// IO semantics:
//   - Stdout/Stderr remain continuous across reloads (concatenated byte streams without implicit markers).
//     Implementations MAY insert textual markers if configured.
//   - Stdin writes always target the current process; writes during switchover may block or return a temporary error.
type Supervisor interface {
	// Start launches the process and waits until it stays stable for at least stableFor.
	// If the process exits before stability is reached, Start returns ErrFailedToStart wrapping a typed cause
	// (e.g., *ExitError, *StableDeadlineExceededError).
	Start(ctx context.Context, stableFor time.Duration) error

	// Stop requests a graceful termination (platform-appropriate), and after the context deadline,
	// performs a hard termination. If a hard kill was required, Stop returns a *KilledError.
	Stop(ctx context.Context) error

	// Reload performs a graceful restart of the same process specification.
	// The old process is stopped first, then the new process is started.
	// On failure to start the new process, returns ErrFailedToStart wrapping the underlying cause.
	// Must publish RestartedEvent then ReloadedEvent.
	Reload(ctx context.Context, stableFor time.Duration) error

	// ReloadWithSpecification performs a graceful restart with an updated process specification.
	// Only Arguments, EnvironmentVariables, and WorkingDirectory can be changed.
	// The ExecutablePath must remain the same, otherwise returns ErrCannotChangeExecutable.
	// The old process is stopped first, then the new process is started with the updated specification.
	// Must publish RestartedEvent then ReloadedEvent.
	ReloadWithSpecification(ctx context.Context, newSpecification Specification, stableFor time.Duration) error

	// Stdout returns a reader for standard output. The stream remains open across Reload and closes after Stop.
	Stdout() io.ReadCloser
	// Stderr returns a reader for standard error. The stream remains open across Reload and closes after Stop.
	Stderr() io.ReadCloser
	// Stdin returns a writer for standard input of the current process.
	Stdin() io.WriteCloser

	// Events returns an observable with process events.
	Events() rxgo.Observable

	// Specification returns initial specification of the process.
	Specification() Specification

	// Status returns a race-free snapshot of the current state.
	Status() Status

	// State returns the current lifecycle state of the supervisor.
	State() SupervisorState

	// Wait blocks until the current process instance exits, returning *ExitError or nil if stopped cleanly.
	Wait(ctx context.Context) error
}

// ExitError describes a process exit (code or signal) and when it occurred.
type ExitError struct {
	PID    int
	Code   int    // -1 if not exited by code
	Signal string // empty if not terminated by signal
	When   time.Time
}

func (e *ExitError) Error() string {
	if e.Signal != "" {
		return "process exited by signal: " + e.Signal
	}
	return "process exited with code: " + strconv.Itoa(e.Code)
}

// StableDeadlineExceededError indicates the process failed to remain stable for the requested period.
type StableDeadlineExceededError struct {
	PID       int
	StableFor time.Duration
	Deadline  time.Duration
}

func (e *StableDeadlineExceededError) Error() string {
	return fmt.Sprintf("stable deadline (%d ms) exceeded for process %d", e.PID, e.Deadline.Milliseconds())
}

// KilledError indicates a hard termination was required.
type KilledError struct {
	PID   int
	After time.Duration
}

func (e KilledError) Error() string {
	return fmt.Sprintf("process %d killed after %d ms", e.PID, e.After.Milliseconds())
}

var (
	// ErrFailedToStart is a high-level error signaling a failed start; it wraps a typed cause.
	ErrFailedToStart = errors.New("failed to start process")

	// ErrCannotChangeExecutable indicates an attempt to change the executable path during reload.
	ErrCannotChangeExecutable = errors.New("cannot change executable path during reload")

	// ErrInvalidState indicates an operation was attempted while the supervisor is in an inappropriate state.
	ErrInvalidState = errors.New("invalid state")
)
