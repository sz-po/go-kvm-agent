package peripheral

import (
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const DisplaySourceServiceId = nodeSDK.ServiceId("node/peripheral/display-source")

const (
	DisplaySourceGetFrameBufferMethod nodeSDK.MethodName = "get-frame-buffer"
	DisplaySourceGetDisplayModeMethod nodeSDK.MethodName = "get-display-mode"
	DisplaySourceGetPixelFormatMethod nodeSDK.MethodName = "get-pixel-format"
	DisplaySourceGetMetricsMethod     nodeSDK.MethodName = "get-metrics"
)

type DisplaySourceGetFrameBufferRequest struct{}

type DisplaySourceGetFrameBufferResponse struct {
	Size int `json:"size"`
}

type DisplaySourceGetDisplayModeRequest struct{}

type DisplaySourceGetDisplayModeResponse struct {
	DisplayMode *peripheralSDK.DisplayMode `json:"displayMode"`
}

type DisplaySourceGetPixelFormatRequest struct{}

type DisplaySourceGetPixelFormatResponse struct {
	PixelFormat *peripheralSDK.DisplayPixelFormat `json:"pixelFormat"`
}

type DisplaySourceGetMetricsRequest struct{}

type DisplaySourceGetMetricsResponse struct {
	Metrics peripheralSDK.DisplaySourceMetrics `json:"metrics"`
}
