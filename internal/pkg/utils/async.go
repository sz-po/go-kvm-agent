package utils

import (
	"time"
)

func GoFuncWithStability(stableDuration time.Duration, fn func() error) error {
	errChan := make(chan error, 1)

	go func() {
		if err := fn(); err != nil {
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	timer := time.NewTimer(stableDuration)
	defer timer.Stop()

	select {
	case err := <-errChan:
		return err
	case <-timer.C:
		return nil
	}
}
