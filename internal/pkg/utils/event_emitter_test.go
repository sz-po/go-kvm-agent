package utils

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestEventEmitter_ConcurrentListenAndEmit reproduces race condition between Listen() and Emit().
// This mpv should fail with "concurrent map iteration and map write" with the buggy code,
// and pass after fixing the race condition.
func TestEventEmitter_ConcurrentListenAndEmit(t *testing.T) {
	emitter := NewEventEmitter[int]()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start emitting data continuously in the background
	go func() {
		for i := 0; i < 1000; i++ {
			emitter.Emit(i)
			time.Sleep(time.Microsecond)
		}
	}()

	// Concurrently add many listeners while emitting is happening
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listenerCtx, listenerCancel := context.WithCancel(ctx)
			defer listenerCancel()

			ch := emitter.Listen(listenerCtx)

			// Try to receive some data
			select {
			case <-ch:
				// Received data successfully
			case <-time.After(100 * time.Millisecond):
				// Timeout is fine
			}
		}()
		time.Sleep(time.Microsecond * 100)
	}

	wg.Wait()
}

// TestEventEmitter_ConcurrentEmitAndListenerCleanup reproduces race condition between Emit() and listener cleanup.
func TestEventEmitter_ConcurrentEmitAndListenerCleanup(t *testing.T) {
	emitter := NewEventEmitter[int]()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create many short-lived listeners that will be cleaned up rapidly
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listenerCtx, listenerCancel := context.WithTimeout(ctx, time.Millisecond*10)
			defer listenerCancel()

			ch := emitter.Listen(listenerCtx)

			// Try to receive
			select {
			case <-ch:
			case <-listenerCtx.Done():
			}
		}()
	}

	// Emit data continuously while listeners are being created and cleaned up
	go func() {
		for i := 0; i < 1000; i++ {
			emitter.Emit(i)
			time.Sleep(time.Microsecond * 50)
		}
	}()

	wg.Wait()
}

// TestEventEmitter_ManyListenersAndEmits tests high concurrency scenario.
func TestEventEmitter_ManyListenersAndEmits(t *testing.T) {
	emitter := NewEventEmitter[int]()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const numListeners = 100
	const numEmitters = 10

	var wg sync.WaitGroup

	// Start many emitters
	for e := 0; e < numEmitters; e++ {
		wg.Add(1)
		go func(emitterID int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				emitter.Emit(emitterID*1000 + i)
				time.Sleep(time.Microsecond * 100)
			}
		}(e)
	}

	// Start many listeners
	for l := 0; l < numListeners; l++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listenerCtx, listenerCancel := context.WithCancel(ctx)
			defer listenerCancel()

			ch := emitter.Listen(listenerCtx)

			count := 0
			for {
				select {
				case <-ch:
					count++
					if count >= 10 {
						return
					}
				case <-listenerCtx.Done():
					return
				case <-time.After(500 * time.Millisecond):
					return
				}
			}
		}()
		time.Sleep(time.Microsecond * 50)
	}

	wg.Wait()
}
