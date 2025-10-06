package go_kvm_agent

import (
	"context"
	"sync"
)

func Start(config Config, wg *sync.WaitGroup, ctx context.Context) error {
	defer wg.Done()

	<-ctx.Done()
	return nil
}
