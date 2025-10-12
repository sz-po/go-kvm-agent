//go:build unix || darwin || linux

package process

import (
	"fmt"
	"os"
	"syscall"
)

// gracefulTerminate sends a SIGTERM signal to the process, requesting graceful termination.
// Returns an error if the signal cannot be sent.
func gracefulTerminate(process *os.Process) error {
	if process == nil {
		return fmt.Errorf("process is nil")
	}
	return process.Signal(syscall.SIGTERM)
}

// forceTerminate sends a SIGKILL signal to the process, forcing immediate termination.
// Returns an error if the signal cannot be sent.
func forceTerminate(process *os.Process) error {
	if process == nil {
		return fmt.Errorf("process is nil")
	}
	return process.Signal(syscall.SIGKILL)
}

// extractSignalName extracts the signal name from a process's exit status if it was terminated by a signal.
// Returns nil if the process exited normally (not via signal).
func extractSignalName(processState *os.ProcessState) *string {
	if processState == nil {
		return nil
	}

	waitStatus, ok := processState.Sys().(syscall.WaitStatus)
	if !ok {
		return nil
	}

	if waitStatus.Signaled() {
		signalName := waitStatus.Signal().String()
		return &signalName
	}

	return nil
}
