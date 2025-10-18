package remote

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	displaySourceAPIService "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine/peripheral/display_source"
	displaySourceAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral/display_source"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const DisplaySourceFramebufferSize = 1024 * 1024 * 8

type DisplaySource struct {
	*Peripheral

	displaySourceService *displaySourceAPIService.Service
}

func (displaySource *DisplaySource) GetDisplayFrameBuffer(ctx context.Context) (*peripheralSDK.DisplayFrameBuffer, error) {
	memoryPool, err := memory.GetDefaultMemoryPool()
	if err != nil {
		return nil, fmt.Errorf("get default memory pool: %w", err)
	}

	return displaySource.displaySourceService.GetFramebuffer(ctx, memoryPool, DisplaySourceFramebufferSize, displaySourceAPI.FramebufferMediaTypeRGB24)
}

func (displaySource *DisplaySource) GetDisplayMode(ctx context.Context) (*peripheralSDK.DisplayMode, error) {
	return displaySource.displaySourceService.GetDisplayMode(ctx)
}

func (displaySource *DisplaySource) GetDisplayPixelFormat(ctx context.Context) (*peripheralSDK.DisplayPixelFormat, error) {
	return displaySource.displaySourceService.GetPixelFormat(ctx)
}

func (displaySource *DisplaySource) GetDisplaySourceMetrics() peripheralSDK.DisplaySourceMetrics {
	//TODO implement me
	panic("implement me")
}
