package peripheral

import (
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

const DisplaySinkServiceId = nodeSDK.ServiceId("node/peripheral/display-sink")

const (
	DisplaySinkSetFrameBufferProviderMethod   nodeSDK.MethodName = "set-frame-buffer-provider"
	DisplaySinkClearFrameBufferProviderMethod nodeSDK.MethodName = "clear-frame-buffer-provider"
)

type DisplaySinkSetFrameBufferProviderRequest struct {
	NodeId     nodeSDK.NodeId       `json:"nodeId"`
	Peripheral peripheralDescriptor `json:"peripheral"`
}

type DisplaySinkSetFrameBufferProviderResponse struct{}

type DisplaySinkClearFrameBufferProviderRequest struct{}

type DisplaySinkClearFrameBufferProviderResponse struct{}
