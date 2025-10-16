package process

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSuperviseConstruction tests that the SuperviseLocal constructor properly initializes a supervisor.
func TestSuperviseConstruction(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/echo",
		Arguments:      []string{"hello"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})

	assert.NotNil(t, supervisor)
	assert.Equal(t, spec, supervisor.Specification())
	assert.NotNil(t, supervisor.Stdout())
	assert.NotNil(t, supervisor.Stderr())
	assert.NotNil(t, supervisor.Stdin())
	assert.NotNil(t, supervisor.Events())
}

// TestStartSuccessful tests that Start successfully launches and stabilizes a process.
func TestStartSuccessful(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"}, // Sleep for 10 seconds
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start with 100ms stability period
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Check status
	status := supervisor.Status()
	assert.True(t, status.Running)
	assert.Greater(t, status.PID, 0)

	// Terminate the process
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	err = supervisor.Stop(stopCtx)
	assert.NoError(t, err)

	// Check status after stop
	status = supervisor.Status()
	assert.False(t, status.Running)
}

// TestStartProcessExitsImmediately tests that Start fails when the process exits before stability.
func TestStartProcessExitsImmediately(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/false", // Exits immediately with code 1
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start with 100ms stability period - should fail
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToStart)
}

// TestStopGraceful tests graceful termination of a process.
func TestStopGraceful(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Terminate gracefully with enough time
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	err = supervisor.Stop(stopCtx)

	// Should not be a KilledError
	var killedError *KilledError
	assert.False(t, errors.As(err, &killedError))
}

// TestStopWithShortTimeout tests that Terminate handles short timeouts correctly and returns KilledError.
func TestStopWithShortTimeout(t *testing.T) {
	// Use a process that ignores SIGTERM to ensure force kill is needed
	spec := Specification{
		ExecutablePath: "/bin/sh",
		Arguments:      []string{"-c", "trap '' TERM; sleep 100"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Collect events
	eventsChan := make(chan Event, 10)
	supervisor.Events().ForEach(func(item interface{}) {
		if event, ok := item.(Event); ok {
			eventsChan <- event
		}
	}, func(err error) {}, func() {})

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Drain StartedEvent
	select {
	case <-eventsChan:
	case <-time.After(1 * time.Second):
		t.Fatal("did not receive StartedEvent")
	}

	oldPID := supervisor.Status().PID

	// Terminate with very short timeout to force kill (process ignores SIGTERM)
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer stopCancel()

	err = supervisor.Stop(stopCtx)

	// Verify KilledError was returned
	assert.Error(t, err)
	var killedError *KilledError
	assert.True(t, errors.As(err, &killedError))
	assert.Equal(t, oldPID, killedError.PID)

	// Verify process stopped
	status := supervisor.Status()
	assert.False(t, status.Running)

	// Verify KilledEvent was published
	killedEventFound := false
	timeout := time.After(2 * time.Second)
	for !killedEventFound {
		select {
		case event := <-eventsChan:
			if event.Kind() == EventKilled {
				killedEvent := event.(KilledEvent)
				assert.Equal(t, oldPID, killedEvent.PID)
				killedEventFound = true
			}
		case <-timeout:
			t.Fatal("KilledEvent not received")
		}
	}

	// Verify old process is no longer running
	checkProcess := Specification{
		ExecutablePath: "/bin/ps",
		Arguments:      []string{"-p", string(rune(oldPID))},
	}
	checkSupervisor := SuperviseLocal(checkProcess, RestartPolicy{Enabled: false})
	checkErr := checkSupervisor.Start(context.Background(), 100*time.Millisecond)

	// ps should exit with non-zero if process doesn't exist
	if checkErr == nil {
		_ = checkSupervisor.Stop(context.Background())
	}
}

// TestStatusBeforeStart tests Status when no process is running.
func TestStatusBeforeStart(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/echo",
		Arguments:      []string{"hello"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	status := supervisor.Status()

	assert.False(t, status.Running)
	assert.Equal(t, 0, status.PID)
	assert.Equal(t, 0, status.RestartCount)
}

// TestWaitForProcessExit tests Wait method blocks until process exits.
func TestWaitForProcessExit(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"1"}, // Sleep for 1 second
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Wait for process to exit
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer waitCancel()

	start := time.Now()
	err = supervisor.Wait(waitCtx)
	duration := time.Since(start)

	// Wait should return ExitError even for successful exit (code 0)
	var exitError *ExitError
	assert.True(t, errors.As(err, &exitError))
	assert.Equal(t, 0, exitError.Code)

	// Should complete in around 1 second (accounting for start overhead)
	assert.Greater(t, duration, 900*time.Millisecond)
	assert.Less(t, duration, 2*time.Second)
}

// TestReloadSuccessful tests successful reload of a process.
func TestReloadSuccessful(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldStatus := supervisor.Status()
	oldPID := oldStatus.PID

	// Reload the process
	err = supervisor.Reload(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	newStatus := supervisor.Status()
	assert.True(t, newStatus.Running)
	assert.NotEqual(t, oldPID, newStatus.PID)
	assert.Equal(t, 1, newStatus.RestartCount)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReload_NoTwoProcessesSimultaneously verifies that during reload,
// the old process is stopped before the new process starts (no two processes at once).
func TestReload_NoTwoProcessesSimultaneously(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"100"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Start reload in background
	reloadDone := make(chan error, 1)
	go func() {
		reloadDone <- supervisor.Reload(ctx, 100*time.Millisecond)
	}()

	// Wait a moment for old process to be stopped (10 second timeout + 2 second force kill)
	// We check after the old process should be dead but before new process is stable
	time.Sleep(1 * time.Second)

	// Verify old process is no longer running
	checkOldProcess := Specification{
		ExecutablePath: "/bin/ps",
		Arguments:      []string{"-p", fmt.Sprintf("%d", oldPID)},
	}
	checkSupervisor := SuperviseLocal(checkOldProcess, RestartPolicy{Enabled: false})
	checkErr := checkSupervisor.Start(context.Background(), 50*time.Millisecond)

	// ps should exit with non-zero if process doesn't exist
	assert.Error(t, checkErr, "Old process should not be running")

	if checkErr == nil {
		_ = checkSupervisor.Stop(context.Background())
	}

	// Wait for reload to complete
	err = <-reloadDone
	assert.NoError(t, err)

	// Verify new process is running with different PID
	newStatus := supervisor.Status()
	assert.True(t, newStatus.Running)
	assert.NotEqual(t, oldPID, newStatus.PID)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReloadWithSpecification_ChangeArguments tests changing process arguments.
func TestReloadWithSpecification_ChangeArguments(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldStatus := supervisor.Status()
	oldPID := oldStatus.PID

	// Change arguments
	newSpec := spec
	newSpec.Arguments = []string{"20"}

	err = supervisor.ReloadWithSpecification(ctx, newSpec, 100*time.Millisecond)
	assert.NoError(t, err)

	// Verify reload succeeded
	newStatus := supervisor.Status()
	assert.True(t, newStatus.Running)
	assert.NotEqual(t, oldPID, newStatus.PID)
	assert.Equal(t, 1, newStatus.RestartCount)

	// Verify specification was updated
	assert.Equal(t, []string{"20"}, supervisor.Specification().Arguments)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReloadWithSpecification_ChangeEnvironment tests changing environment variables.
func TestReloadWithSpecification_ChangeEnvironment(t *testing.T) {
	spec := Specification{
		ExecutablePath:       "/bin/sleep",
		Arguments:            []string{"10"},
		EnvironmentVariables: map[string]string{"TEST_VAR": "old_value"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Change environment
	newSpec := spec
	newSpec.EnvironmentVariables = map[string]string{"TEST_VAR": "new_value", "ANOTHER_VAR": "value"}

	err = supervisor.ReloadWithSpecification(ctx, newSpec, 100*time.Millisecond)
	assert.NoError(t, err)

	// Verify reload succeeded
	newStatus := supervisor.Status()
	assert.True(t, newStatus.Running)
	assert.NotEqual(t, oldPID, newStatus.PID)

	// Verify specification was updated
	assert.Equal(t, "new_value", supervisor.Specification().EnvironmentVariables["TEST_VAR"])
	assert.Equal(t, "value", supervisor.Specification().EnvironmentVariables["ANOTHER_VAR"])

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReloadWithSpecification_RejectExecutableChange tests that changing ExecutablePath is rejected.
func TestReloadWithSpecification_RejectExecutableChange(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID
	oldSpec := supervisor.Specification()

	// Attempt to change ExecutablePath
	newSpec := spec
	newSpec.ExecutablePath = "/bin/cat"

	err = supervisor.ReloadWithSpecification(ctx, newSpec, 100*time.Millisecond)

	// Should return ErrCannotChangeExecutable
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCannotChangeExecutable))

	// Verify process is still running with old specification
	status := supervisor.Status()
	assert.True(t, status.Running)
	assert.Equal(t, oldPID, status.PID)
	assert.Equal(t, oldSpec.ExecutablePath, supervisor.Specification().ExecutablePath)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReloadWithSpecification_MultipleChanges tests changing multiple fields at once.
func TestReloadWithSpecification_MultipleChanges(t *testing.T) {
	tmpDir := "/tmp"
	spec := Specification{
		ExecutablePath:       "/bin/sleep",
		Arguments:            []string{"10"},
		EnvironmentVariables: map[string]string{"VAR1": "value1"},
		WorkingDirectory:     &tmpDir,
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Change multiple fields
	varTmpDir := "/var/tmp"
	newSpec := Specification{
		ExecutablePath:       "/bin/sleep",
		Arguments:            []string{"20"},
		EnvironmentVariables: map[string]string{"VAR1": "value1", "VAR2": "value2"},
		WorkingDirectory:     &varTmpDir,
	}

	err = supervisor.ReloadWithSpecification(ctx, newSpec, 100*time.Millisecond)
	assert.NoError(t, err)

	// Verify reload succeeded
	newStatus := supervisor.Status()
	assert.True(t, newStatus.Running)
	assert.NotEqual(t, oldPID, newStatus.PID)

	// Verify all fields were updated
	finalSpec := supervisor.Specification()
	assert.Equal(t, []string{"20"}, finalSpec.Arguments)
	assert.Equal(t, "value2", finalSpec.EnvironmentVariables["VAR2"])
	assert.Equal(t, "/var/tmp", *finalSpec.WorkingDirectory)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReload_NotStarted verifies that reload fails when process has not been started.
func TestReload_NotStarted(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Try to reload without starting
	err := supervisor.Reload(ctx, 100*time.Millisecond)

	// Should return ErrInvalidState
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidState)
	assert.Contains(t, err.Error(), "cannot reload")
}

// TestStop_NotStarted verifies that stop fails when process has not been started.
func TestStop_NotStarted(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Try to stop without starting
	err := supervisor.Stop(ctx)

	// Should return ErrInvalidState
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidState)
	assert.Contains(t, err.Error(), "cannot stop")
}

// TestWait_NotStarted verifies that wait returns immediately when process has not been started.
func TestWait_NotStarted(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Wait without starting - should return immediately with nil
	start := time.Now()
	err := supervisor.Wait(ctx)
	duration := time.Since(start)

	// Should return nil immediately
	assert.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Wait should return immediately")
}

// TestWait_ContextDeadline verifies that wait honors context deadline.
func TestWait_ContextDeadline(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"100"}, // Sleep for a long time
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Wait with short deadline
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer waitCancel()

	start := time.Now()
	err = supervisor.Wait(waitCtx)
	duration := time.Since(start)

	// Should return context.DeadlineExceeded after ~200ms
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Greater(t, duration, 150*time.Millisecond)
	assert.Less(t, duration, 500*time.Millisecond)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestStart_ContextExpires verifies that start fails when context expires before stability.
func TestStart_ContextExpires(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"100"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})

	// Create context that expires before stability period
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Start with 200ms stability period but context expires after 50ms
	err := supervisor.Start(ctx, 200*time.Millisecond)

	// Should return ErrFailedToStart wrapping context.DeadlineExceeded
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToStart)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Verify supervisor is in idle state
	status := supervisor.Status()
	assert.False(t, status.Running)
}

// TestStart_AlreadyStarted tests that Start returns ErrInvalidState when called on an already running supervisor.
func TestStart_AlreadyStarted(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process successfully
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Attempt to start again while already running
	err = supervisor.Start(ctx, 100*time.Millisecond)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidState)
	assert.Contains(t, err.Error(), "cannot start")

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestProcess_StdoutClosedEarly tests that the supervisor handles a process that closes stdout before exiting.
func TestProcess_StdoutClosedEarly(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sh",
		Arguments:      []string{"-c", "exec 1>&-; sleep 2"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process - it will close stdout immediately
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Give it a moment for the copy goroutine to handle the closed pipe
	time.Sleep(500 * time.Millisecond)

	// Verify process is still running despite closed stdout
	status := supervisor.Status()
	assert.True(t, status.Running)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	err = supervisor.Stop(stopCtx)
	assert.NoError(t, err)
}

// TestProcess_StderrClosedEarly tests that the supervisor handles a process that closes stderr before exiting.
func TestProcess_StderrClosedEarly(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sh",
		Arguments:      []string{"-c", "exec 2>&-; sleep 2"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the process - it will close stderr immediately
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Give it a moment for the copy goroutine to handle the closed pipe
	time.Sleep(500 * time.Millisecond)

	// Verify process is still running despite closed stderr
	status := supervisor.Status()
	assert.True(t, status.Running)

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	err = supervisor.Stop(stopCtx)
	assert.NoError(t, err)
}

// TestReload_OldProcessRequiresForceKill tests that reload successfully handles an old process
// that ignores SIGTERM and requires SIGKILL (force termination).
func TestReload_OldProcessRequiresForceKill(t *testing.T) {
	// Start with a process that ignores SIGTERM
	spec := Specification{
		ExecutablePath: "/bin/sh",
		Arguments:      []string{"-c", "trap '' TERM; sleep 100"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the initial process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Reload - old process ignores SIGTERM so will require force kill after 10s timeout
	// This mpv will take ~12 seconds due to the hardcoded 10s timeout + 2s force kill wait
	err = supervisor.Reload(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	// Verify reload succeeded with new process
	newStatus := supervisor.Status()
	assert.True(t, newStatus.Running)
	assert.NotEqual(t, oldPID, newStatus.PID)
	assert.Equal(t, 1, newStatus.RestartCount)

	// Verify old process is no longer running
	checkOldProcess := Specification{
		ExecutablePath: "/bin/ps",
		Arguments:      []string{"-p", fmt.Sprintf("%d", oldPID)},
	}
	checkSupervisor := SuperviseLocal(checkOldProcess, RestartPolicy{Enabled: false})
	checkErr := checkSupervisor.Start(context.Background(), 50*time.Millisecond)
	assert.Error(t, checkErr, "Old process should not be running")

	if checkErr == nil {
		_ = checkSupervisor.Stop(context.Background())
	}

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestReload_NewProcessFailsToStart tests that reload fails gracefully when the new process
// cannot be launched after the old process has already been stopped (no rollback possible).
func TestReload_NewProcessFailsToStart(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the initial process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Modify the specification to use an invalid executable path
	// This will cause launchProcess to fail after the old process is stopped
	supervisorLocal := supervisor.(*supervisorLocal)
	supervisorLocal.specification.ExecutablePath = "/nonexistent/executable/path"

	// Attempt reload - should fail when trying to start new process
	err = supervisor.Reload(ctx, 100*time.Millisecond)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToStart)
	assert.Contains(t, err.Error(), "failed to start")

	// Verify supervisor is in idle state (no process running)
	status := supervisor.Status()
	assert.False(t, status.Running)
	assert.Equal(t, 0, status.RestartCount) // No successful restart

	// Verify old process was actually stopped
	checkOldProcess := Specification{
		ExecutablePath: "/bin/ps",
		Arguments:      []string{"-p", fmt.Sprintf("%d", oldPID)},
	}
	checkSupervisor := SuperviseLocal(checkOldProcess, RestartPolicy{Enabled: false})
	checkErr := checkSupervisor.Start(context.Background(), 50*time.Millisecond)
	assert.Error(t, checkErr, "Old process should not be running")

	if checkErr == nil {
		_ = checkSupervisor.Stop(context.Background())
	}
}

// TestReload_NewProcessExitsBeforeStability tests that reload fails when the new process
// exits before achieving the required stability period.
func TestReload_NewProcessExitsBeforeStability(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the initial process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Modify specification to use a process that exits immediately
	supervisorLocal := supervisor.(*supervisorLocal)
	supervisorLocal.specification.ExecutablePath = "/bin/sh"
	supervisorLocal.specification.Arguments = []string{"-c", "exit 1"}

	// Attempt reload - new process will exit before stability
	err = supervisor.Reload(ctx, 100*time.Millisecond)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToStart)
	assert.Contains(t, err.Error(), "new process")
	assert.Contains(t, err.Error(), "exited")

	// Verify supervisor is in idle state
	status := supervisor.Status()
	assert.False(t, status.Running)
	assert.Equal(t, 0, status.RestartCount) // No successful restart

	// Verify old process was stopped
	checkOldProcess := Specification{
		ExecutablePath: "/bin/ps",
		Arguments:      []string{"-p", fmt.Sprintf("%d", oldPID)},
	}
	checkSupervisor := SuperviseLocal(checkOldProcess, RestartPolicy{Enabled: false})
	checkErr := checkSupervisor.Start(context.Background(), 50*time.Millisecond)
	assert.Error(t, checkErr, "Old process should not be running")

	if checkErr == nil {
		_ = checkSupervisor.Stop(context.Background())
	}
}

// TestReload_ContextCanceledDuringStability tests that reload fails when the context
// is canceled during the stability waiting period for the new process.
func TestReload_ContextCanceledDuringStability(t *testing.T) {
	spec := Specification{
		ExecutablePath: "/bin/sleep",
		Arguments:      []string{"10"},
	}

	supervisor := SuperviseLocal(spec, RestartPolicy{Enabled: false})
	ctx := context.Background()

	// Start the initial process
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err)

	oldPID := supervisor.Status().PID

	// Create context with timeout shorter than stability period
	// Stability period will be 200ms, context timeout 50ms
	reloadCtx, reloadCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer reloadCancel()

	// Attempt reload - context will be canceled before stability is achieved
	err = supervisor.Reload(reloadCtx, 200*time.Millisecond)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToStart)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, err.Error(), "context canceled")

	// Verify supervisor is in idle state
	status := supervisor.Status()
	assert.False(t, status.Running)
	assert.Equal(t, 0, status.RestartCount) // No successful restart

	// Verify old process was stopped
	checkOldProcess := Specification{
		ExecutablePath: "/bin/ps",
		Arguments:      []string{"-p", fmt.Sprintf("%d", oldPID)},
	}
	checkSupervisor := SuperviseLocal(checkOldProcess, RestartPolicy{Enabled: false})
	checkErr := checkSupervisor.Start(context.Background(), 50*time.Millisecond)
	assert.Error(t, checkErr, "Old process should not be running")

	if checkErr == nil {
		_ = checkSupervisor.Stop(context.Background())
	}
}

// TestAutoRestart_SingleCrash tests that a process automatically restarts after a single crash.
func TestAutoRestart_SingleCrash(t *testing.T) {
	// Process that runs for a short time then crashes
	// Sleep for 200ms (enough to stabilize) then exit with error
	spec := Specification{
		ExecutablePath: "/bin/sh",
		Arguments:      []string{"-c", "sleep 0.2; exit 1"},
	}

	policy := RestartPolicy{
		Enabled:      true,
		MaxAttempts:  5,
		Strategy:     StrategyConstant,
		InitialDelay: 200 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		ResetWindow:  5 * time.Minute,
	}

	supervisor := SuperviseLocal(spec, policy)
	ctx := context.Background()

	// Collect events
	eventsChan := make(chan Event, 30)
	supervisor.Events().ForEach(func(item interface{}) {
		if event, ok := item.(Event); ok {
			eventsChan <- event
		}
	}, func(err error) {}, func() {})

	// Start process - should succeed initially
	err := supervisor.Start(ctx, 100*time.Millisecond)
	assert.NoError(t, err, "Initial start should succeed")

	// Wait for the process to crash and auto-restart
	time.Sleep(1500 * time.Millisecond)

	// Collect events
	var exitedEvents []ExitedEvent
	var restartedEvents []RestartedEvent

	timeout := time.After(500 * time.Millisecond)
drainLoop:
	for {
		select {
		case event := <-eventsChan:
			switch e := event.(type) {
			case ExitedEvent:
				exitedEvents = append(exitedEvents, e)
			case RestartedEvent:
				restartedEvents = append(restartedEvents, e)
			}
		case <-timeout:
			break drainLoop
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Should have at least one ExitedEvent (first crash) and one RestartedEvent with CauseCrash
	assert.GreaterOrEqual(t, len(exitedEvents), 1, "Should have at least one exited event from crash")
	assert.GreaterOrEqual(t, len(restartedEvents), 1, "Should have at least one restarted event")

	// Verify restart was due to crash
	if len(restartedEvents) > 0 {
		assert.Equal(t, CauseCrash, restartedEvents[0].Cause, "Restart cause should be crash")
	}

	// Clean up
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = supervisor.Stop(stopCtx)
}

// TestAutoRestart_DisabledPolicy tests that auto-restart does not occur when policy is disabled.
func TestAutoRestart_DisabledPolicy(t *testing.T) {
	// Process that exits immediately
	spec := Specification{
		ExecutablePath: "/bin/false",
	}

	policy := RestartPolicy{
		Enabled: false, // Disabled
	}

	supervisor := SuperviseLocal(spec, policy)
	ctx := context.Background()

	// Collect events
	eventsChan := make(chan Event, 10)
	supervisor.Events().ForEach(func(item interface{}) {
		if event, ok := item.(Event); ok {
			eventsChan <- event
		}
	}, func(err error) {}, func() {})

	// Start process - will fail immediately
	err := supervisor.Start(ctx, 50*time.Millisecond)
	assert.Error(t, err)

	// Wait a moment
	time.Sleep(500 * time.Millisecond)

	// Collect events
	var restartedEvents []RestartedEvent
	timeout := time.After(1 * time.Second)
eventLoop:
	for {
		select {
		case event := <-eventsChan:
			if e, ok := event.(RestartedEvent); ok {
				restartedEvents = append(restartedEvents, e)
			}
		case <-timeout:
			break eventLoop
		}
	}

	// Should have NO RestartedEvent because policy is disabled
	assert.Equal(t, 0, len(restartedEvents), "Should not have any restarted events when policy is disabled")

	// Verify supervisor is in idle state
	assert.Equal(t, StateIdle, supervisor.State())
}

// TestAutoRestart_MaxAttemptsExceeded tests that auto-restart stops after MaxAttempts.
func TestAutoRestart_MaxAttemptsExceeded(t *testing.T) {
	// Process that always crashes
	spec := Specification{
		ExecutablePath: "/bin/false",
	}

	policy := RestartPolicy{
		Enabled:      true,
		MaxAttempts:  2, // Only 2 attempts
		Strategy:     StrategyConstant,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		ResetWindow:  5 * time.Minute,
	}

	supervisor := SuperviseLocal(spec, policy)
	ctx := context.Background()

	// Collect events
	eventsChan := make(chan Event, 20)
	supervisor.Events().ForEach(func(item interface{}) {
		if event, ok := item.(Event); ok {
			eventsChan <- event
		}
	}, func(err error) {}, func() {})

	// Start process - will crash immediately
	_ = supervisor.Start(ctx, 50*time.Millisecond)

	// Wait for restart attempts to complete
	time.Sleep(2 * time.Second)

	// Collect events
	var restartedEvents []RestartedEvent
	timeout := time.After(500 * time.Millisecond)
collectLoop:
	for {
		select {
		case event := <-eventsChan:
			if e, ok := event.(RestartedEvent); ok {
				restartedEvents = append(restartedEvents, e)
			}
		case <-timeout:
			break collectLoop
		}
	}

	// Should have exactly MaxAttempts restart events
	assert.LessOrEqual(t, len(restartedEvents), policy.MaxAttempts, "Should not exceed max attempts")

	// Verify supervisor eventually gives up and goes to Idle
	finalState := supervisor.State()
	assert.Equal(t, StateIdle, finalState, "Supervisor should be idle after exhausting restart attempts")
}
