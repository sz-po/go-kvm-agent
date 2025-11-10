//go:build linux

package tc358743

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/memory"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/io"
	v4l2io "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/utils/linux/v4l2/io"
	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
	"golang.org/x/sys/unix"
)

type FrameHandler func(memoryBuffer memorySDK.Buffer) error

func DiscardFrameHandler(memoryBuffer memorySDK.Buffer) error {
	return memoryBuffer.Release()
}

type DeviceOptions struct {
	memoryPoolProvider memorySDK.PoolProvider
	frameHandler       FrameHandler
	pixelFormat        peripheralSDK.DisplayPixelFormat
	logger             *slog.Logger
}

type DeviceOpt func(*DeviceOptions)

func WithMemoryPoolProvider(provider memorySDK.PoolProvider) DeviceOpt {
	return func(options *DeviceOptions) {
		options.memoryPoolProvider = provider
	}
}

func WithFrameHandler(handler FrameHandler) DeviceOpt {
	return func(options *DeviceOptions) {
		options.frameHandler = handler
	}
}

func WithLogger(logger *slog.Logger) DeviceOpt {
	return func(options *DeviceOptions) {
		options.logger = logger
	}
}

func defaultDeviceOptions() DeviceOptions {
	return DeviceOptions{
		memoryPoolProvider: memory.DefaultMemoryPoolProvider,
		frameHandler:       DiscardFrameHandler,
		pixelFormat:        peripheralSDK.DisplayPixelFormatRGB24,
		logger:             slog.New(slog.DiscardHandler),
	}
}

type Device struct {
	devicePath string

	pixelFormat peripheralSDK.DisplayPixelFormat

	currentDisplayMode     *peripheralSDK.DisplayMode
	currentDisplayModeLock *sync.RWMutex

	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	streamLock   *sync.Mutex
	streamCtx    context.Context
	streamCancel context.CancelFunc

	frameHandler FrameHandler

	memoryPool memorySDK.Pool
	logger     *slog.Logger
}

func Open(ctx context.Context, devicePath string, opts ...DeviceOpt) (*Device, error) {
	options := defaultDeviceOptions()

	for _, opt := range opts {
		opt(&options)
	}

	logger := options.logger.With(slog.String("devicePath", devicePath))

	memoryPool, err := options.memoryPoolProvider()
	if err != nil {
		return nil, fmt.Errorf("memory pool: %w", err)
	}

	lifecycleCtx, lifecycleCancel := context.WithCancel(ctx)

	device := &Device{
		devicePath:  devicePath,
		pixelFormat: options.pixelFormat,

		currentDisplayMode:     nil,
		currentDisplayModeLock: &sync.RWMutex{},

		lifecycleCtx:    lifecycleCtx,
		lifecycleCancel: lifecycleCancel,

		streamLock: &sync.Mutex{},

		frameHandler: options.frameHandler,

		memoryPool: memoryPool,
		logger:     logger,
	}

	descriptor, err := device.openDevice(ctx)
	if err != nil {
		lifecycleCancel()
		return nil, fmt.Errorf("open device: %w", err)
	}

	err = device.closeDevice(ctx, descriptor)
	if err != nil {
		lifecycleCancel()
		return nil, fmt.Errorf("close device: %w", err)
	}

	logger.Info("Device capabilities verified and it is ready for use.")

	go device.controlLoop(lifecycleCtx)

	return device, nil
}

func (device *Device) GetDisplayMode() (*peripheralSDK.DisplayMode, error) {
	device.currentDisplayModeLock.RLock()
	defer device.currentDisplayModeLock.RUnlock()

	return device.currentDisplayMode, nil
}

func (device *Device) GetPixelFormat() (*peripheralSDK.DisplayPixelFormat, error) {
	return &device.pixelFormat, nil
}

func (device *Device) controlLoop(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for {
		descriptor, err := device.openDevice(ctx)
		if err != nil {
			device.logger.Warn("Open device error. Retrying initialization.", slog.String("error", err.Error()))
			device.closeDevice(ctx, descriptor)
			continue
		}

		device.logger.Info("Waiting for signal.")

		timings, err := device.setupTimings(ctx, descriptor)
		if err != nil {
			device.logger.Warn("Wait for signal error. Retrying initialization.", slog.String("error", err.Error()))
			device.closeDevice(ctx, descriptor)
			continue
		}

		device.logger.Info("Signal acquired.",
			slog.Int("inputWidth", int(timings.Width)),
			slog.Int("inputHeight", int(timings.Height)),
			slog.Float64("inputFrameRate", timings.GetFrameRate()),
		)

		videoFormat, err := device.setupFormat(ctx, descriptor, int(timings.Width), int(timings.Height), timings.GetFrameRate())
		if err != nil {
			device.logger.Warn("Setup video format error. Retrying initialization.", slog.String("error", err.Error()))
			device.closeDevice(ctx, descriptor)
			continue
		}

		device.logger.Debug("Video format set.",
			slog.Int("inputWidth", int(videoFormat.Width)),
			slog.Int("inputHeight", int(videoFormat.Height)),
			slog.String("inputPixelFormat", videoFormat.PixelFormat.String()),
			slog.Int("inputImageSize", int(videoFormat.SizeImage)),
		)

		buffers, err := device.initMemory(ctx, descriptor)
		if err != nil {
			device.logger.Warn("Init memory error. Retrying initialization.", slog.String("error", err.Error()))
			device.closeDevice(ctx, descriptor)
			continue
		}

		wg.Add(1)

		err = device.stream(wg, descriptor, buffers)
		if err != nil {
			device.logger.Warn("Stream error. Retrying initialization.", slog.String("error", err.Error()))
			device.closeDevice(ctx, descriptor)
			continue
		}

		device.logger.Info("Capturing input from device.")

		wg.Wait()

		err = device.releaseMemory(ctx, descriptor, buffers)
		if err != nil {
			device.logger.Warn("Release memory error. Retrying initialization.", slog.String("error", err.Error()))
			continue
		}

		err = device.closeDevice(ctx, descriptor)
		if err != nil {
			device.logger.Warn("Close device error. Retrying initialization.", slog.String("error", err.Error()))
			continue
		}

		if ctx.Err() != nil {
			device.logger.Debug("Control loop terminated.")
			return
		}
	}
}

func (device *Device) handleEvent(ctx context.Context, descriptor io.DeviceDescriptor) {
	device.logger.Debug("Handling event.")

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()

	event, err := io.RetryOnErrorWithValue(ctx, func() (v4l2io.Event, error) {
		return v4l2io.DequeueEvent(descriptor)
	}, io.ErrTemporary, io.ErrNoEntity)
	if err != nil {
		device.logger.Warn("Dequeue event error.", slog.String("error", err.Error()))
		return
	}

	switch event.Type {
	case v4l2io.EventTypeSourceChange:
		device.logger.Debug("Source change event. Requesting stop stream watch.")
		if err := device.stopStream(); err != nil {
			device.logger.Warn("Stop stream watch error.", slog.String("error", err.Error()))
		}
	default:
		device.logger.Warn("Unknown event type.")
	}
}

func (device *Device) stream(wg *sync.WaitGroup, descriptor io.DeviceDescriptor, buffers v4l2io.BoundMmapBuffers) error {
	device.streamLock.Lock()
	defer device.streamLock.Unlock()

	if device.streamCtx != nil {
		return fmt.Errorf("already streaming")
	}

	streamCtx, streamCancel := context.WithCancel(device.lifecycleCtx)

	go device.streamLoop(streamCtx, wg, descriptor, buffers)

	device.streamCtx = streamCtx
	device.streamCancel = streamCancel

	return nil
}

func (device *Device) stopStream() error {
	device.streamLock.Lock()
	defer device.streamLock.Unlock()

	if device.streamCtx == nil {
		return fmt.Errorf("not streaming")
	}

	device.streamCancel()

	return nil
}

func (device *Device) streamLoop(ctx context.Context, wg *sync.WaitGroup, descriptor io.DeviceDescriptor, buffers v4l2io.BoundMmapBuffers) {
	defer wg.Done()
	done := ctx.Done()

	err := v4l2io.StartStream(descriptor, v4l2io.BufferTypeVideoCapture)
	if err != nil {
		device.logger.Warn("Start stream error.", slog.String("error", err.Error()))
		return
	}

	pollEvents := io.Poll(ctx, descriptor, io.PollEventInput, io.PollEventPriority)

	device.logger.Debug("Watching for frames.")

LOOP:
	for {
		select {
		case <-done:
			break LOOP
		case pollEvent := <-pollEvents:
			switch pollEvent {
			case io.PollEventInput:
				device.handleFrame(ctx, descriptor, buffers)
			case io.PollEventPriority:
				device.handleEvent(ctx, descriptor)
			default:
				device.logger.Warn("Unknown poll event.")
			}
		}
	}

	device.logger.Debug("Stream watch terminated.")

	device.streamLock.Lock()
	defer device.streamLock.Unlock()

	device.streamCtx = nil
	device.streamCancel = nil

	err = v4l2io.StopStream(descriptor, v4l2io.BufferTypeVideoCapture)
	if err != nil {
		device.logger.Warn("Stop stream error.", slog.String("error", err.Error()))
		return
	}
}

func (device *Device) handleFrame(ctx context.Context, descriptor io.DeviceDescriptor, buffers v4l2io.BoundMmapBuffers) {
	dequeueCtx, dequeueCancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer dequeueCancel()

	bufferDescriptor, err := io.RetryOnErrorWithValue(dequeueCtx, func() (v4l2io.BufferDescriptor, error) {
		return v4l2io.DequeueMmapBuffer(descriptor, v4l2io.BufferTypeVideoCapture)
	}, io.ErrTemporary, io.ErrTimeout)
	if err != nil {
		device.logger.Warn("Dequeue buffer error.", slog.String("error", err.Error()))
		return
	}

	defer func() {
		err = v4l2io.QueueBuffer(descriptor, bufferDescriptor)
		if err != nil {
			device.logger.Warn("Queue buffer error.", slog.String("error", err.Error()))
		}
	}()

	bytesUsed := int(bufferDescriptor.BytesUsed)

	videoBuffer := buffers[bufferDescriptor.Index]

	memoryBuffer, err := device.memoryPool.Borrow(bytesUsed)
	if err != nil {
		device.logger.Warn("Borrow memory error.", slog.String("error", err.Error()))
		return
	}

	bytesWritten, err := memoryBuffer.Write(videoBuffer.Data[:bytesUsed])
	if err != nil {
		memoryBuffer.Release()
		device.logger.Warn("Write memory error.", slog.String("error", err.Error()))
		return
	}

	if bytesWritten != bytesUsed {
		memoryBuffer.Release()
		device.logger.Warn("Write memory error.", slog.String("error", "bytes written not equal to bytes used"))
		return
	}

	err = device.frameHandler(memoryBuffer)
	if err != nil {
		device.logger.Warn("Frame handler error.", slog.String("error", err.Error()))
		return
	}
}

func (device *Device) setupTimings(ctx context.Context, descriptor io.DeviceDescriptor) (v4l2io.DigitalVideoBTTimings, error) {
	timingsSetupCtx, timingsSetupCancel := context.WithTimeout(ctx, time.Second*10)
	defer timingsSetupCancel()

	timings, err := io.RetryOnErrorWithValue(timingsSetupCtx, func() (v4l2io.DigitalVideoBTTimings, error) {
		return v4l2io.QueryDigitalVideoBTTimings(descriptor)
	}, io.ErrDeviceOrResourceBusy, io.ErrNoLink)
	if err != nil {
		return v4l2io.EmptyDigitalVideoBTTimings, fmt.Errorf("query digital video timings: %w", err)
	}

	err = io.RetryOnError(timingsSetupCtx, func() error {
		return v4l2io.SetDigitalVideoBTTimings(descriptor, timings)
	}, io.ErrDeviceOrResourceBusy, io.ErrNoLink)
	if err != nil {
		return v4l2io.EmptyDigitalVideoBTTimings, fmt.Errorf("set digital video timings: %w", err)
	}

	return timings, nil
}

func (device *Device) setupFormat(ctx context.Context, descriptor io.DeviceDescriptor, width int, height int, refreshRate float64) (v4l2io.VideoFormat, error) {
	formatSetupCtx, formatSetupCancel := context.WithTimeout(ctx, time.Second*10)
	defer formatSetupCancel()

	pixelFormat, err := device.negotiatePixelFormat(descriptor, device.pixelFormat)
	if err != nil {
		return v4l2io.EmptyVideoFormat, fmt.Errorf("negotiate pixel format: %s: %w", device.pixelFormat.String(), err)
	}

	videoFormat := v4l2io.VideoFormat{
		Width:        uint32(width),
		Height:       uint32(height),
		PixelFormat:  pixelFormat.Code,
		Colorspace:   v4l2io.VideoFormatColorspaceSRGB,
		TransferFunc: v4l2io.VideoFormatTransferFunctionSRGB,
		Quantization: v4l2io.VideoFormatQuantizationFullRange,
	}

	videoFormat, err = v4l2io.TryVideoFormat(descriptor, v4l2io.BufferTypeVideoCapture, videoFormat)
	if err != nil {
		return v4l2io.EmptyVideoFormat, fmt.Errorf("try video format: %w", err)
	}

	err = io.RetryOnError(formatSetupCtx, func() error {
		return v4l2io.SetVideoFormat(descriptor, v4l2io.BufferTypeVideoCapture, videoFormat)
	}, io.ErrDeviceOrResourceBusy)
	if err != nil {
		return v4l2io.EmptyVideoFormat, fmt.Errorf("set video format: %w", err)
	}

	videoFormat, err = v4l2io.GetVideoFormat(descriptor, v4l2io.BufferTypeVideoCapture)
	if err != nil {
		return v4l2io.EmptyVideoFormat, fmt.Errorf("get video format: %w", err)
	}

	device.currentDisplayModeLock.Lock()
	defer device.currentDisplayModeLock.Unlock()

	device.currentDisplayMode = &peripheralSDK.DisplayMode{
		Width:       videoFormat.Width,
		Height:      videoFormat.Height,
		RefreshRate: uint32(refreshRate),
	}

	return videoFormat, nil
}

func (device *Device) negotiatePixelFormat(descriptor io.DeviceDescriptor, pixelFormat peripheralSDK.DisplayPixelFormat) (v4l2io.PixelFormat, error) {
	devicePixelFormats, err := v4l2io.ListPixelFormats(descriptor, v4l2io.BufferTypeVideoCapture)
	if err != nil {
		return v4l2io.EmptyPixelFormat, fmt.Errorf("list pixel formats: %w", err)
	}

	for _, devicePixelFormat := range devicePixelFormats {
		if devicePixelFormat.Code.String() == pixelFormat.FourCC() {
			return devicePixelFormat, nil
		}
	}

	return v4l2io.EmptyPixelFormat, peripheralSDK.ErrUnsupportedPixelFormat
}

func (device *Device) initMemory(ctx context.Context, descriptor io.DeviceDescriptor) (v4l2io.BoundMmapBuffers, error) {
	bufferCount, err := v4l2io.RequestBuffers(descriptor, v4l2io.BufferTypeVideoCapture, v4l2io.MemoryTypeMmap, 4)
	if err != nil {
		return v4l2io.EmptyBoundMmapBuffers, fmt.Errorf("request buffers: %w", err)
	}

	buffers := make(v4l2io.BoundMmapBuffers, bufferCount)

	for bufferIndex := v4l2io.BufferIndex(0); bufferIndex < bufferCount; bufferIndex++ {
		buffer, err := v4l2io.QueryMmapBuffer(descriptor, v4l2io.BufferTypeVideoCapture, bufferIndex)
		if err != nil {
			return v4l2io.EmptyBoundMmapBuffers, fmt.Errorf("query buffer: %d: %w", bufferIndex, err)
		}

		boundBuffer, err := v4l2io.BindMmapBuffer(descriptor, buffer)
		if err != nil {
			for i := v4l2io.BufferIndex(0); i < bufferIndex; i++ {
				_ = v4l2io.UnbindMmapBuffer(buffers[bufferIndex])
			}
			return v4l2io.EmptyBoundMmapBuffers, fmt.Errorf("bind buffer: %d: %w", bufferIndex, err)
		}

		err = v4l2io.QueueBuffer(descriptor, buffer.BufferDescriptor)
		if err != nil {
			for i := v4l2io.BufferIndex(0); i < bufferIndex; i++ {
				_ = v4l2io.UnbindMmapBuffer(buffers[bufferIndex])
			}
			return v4l2io.EmptyBoundMmapBuffers, fmt.Errorf("queue buffer: %d: %w", bufferIndex, err)
		}

		buffers[bufferIndex] = boundBuffer
	}

	device.logger.Debug("Memory initialized.", slog.Int("bufferCount", int(bufferCount)))

	return buffers, nil
}

func (device *Device) releaseMemory(ctx context.Context, descriptor io.DeviceDescriptor, buffers v4l2io.BoundMmapBuffers) error {
	releaseMemoryCtx, releaseMemoryCancel := context.WithTimeout(ctx, time.Second*10)
	defer releaseMemoryCancel()

	for bufferIndex := v4l2io.BufferIndex(0); bufferIndex < v4l2io.BufferIndex(len(buffers)); bufferIndex++ {
		err := v4l2io.UnbindMmapBuffer(buffers[bufferIndex])
		if err != nil {
			return fmt.Errorf("unbind buffer: %w", err)
		}
	}

	err := io.RetryOnError(releaseMemoryCtx, func() error {
		return v4l2io.ReleaseBuffers(descriptor, v4l2io.BufferTypeVideoCapture, v4l2io.MemoryTypeMmap)
	}, io.ErrDeviceOrResourceBusy)
	if err != nil {
		return fmt.Errorf("release buffers: %w", err)
	}

	device.logger.Debug("Memory released.")

	return nil
}

func (device *Device) openDevice(ctx context.Context) (io.DeviceDescriptor, error) {
	descriptor, err := io.OpenDevice(device.devicePath, unix.O_RDWR|unix.O_NONBLOCK, 0)
	if err != nil {
		return io.EmptyDeviceDescriptor, fmt.Errorf("open device: %w", err)
	}

	capabilities, err := v4l2io.QueryCapabilities(descriptor)
	if err != nil {
		_ = io.Close(descriptor)
		return io.EmptyDeviceDescriptor, err
	}

	if !capabilities.Features.VideoCapture {
		_ = io.Close(descriptor)
		return io.EmptyDeviceDescriptor, ErrVideoCaptureNotSupported
	}

	if !capabilities.Features.Streaming {
		_ = io.Close(descriptor)
		return io.EmptyDeviceDescriptor, ErrStreamingNotSupported
	}

	err = v4l2io.SubscribeEvent(descriptor, v4l2io.EventTypeSourceChange)
	if err != nil {
		_ = io.Close(descriptor)
		return io.EmptyDeviceDescriptor, fmt.Errorf("subscribe event: %w", err)
	}

	device.logger.Debug("Device open.",
		slog.String("deviceDriver", capabilities.Driver),
		slog.String("deviceBus", capabilities.BusInfo),
	)

	return descriptor, nil
}

func (device *Device) closeDevice(ctx context.Context, descriptor io.DeviceDescriptor) error {
	err := io.Close(descriptor)

	device.logger.Debug("Device closed.")

	return err
}
