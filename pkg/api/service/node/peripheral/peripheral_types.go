package peripheral

import (
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const PeripheralServiceId = nodeSDK.ServiceId("node/peripheral")

const (
	PeripheralTerminateMethod nodeSDK.MethodName = "terminate"
)

type PeripheralTerminateRequest struct {
}

type PeripheralTerminateResponse struct {
}

type peripheralDescriptor struct {
	Id           peripheralSDK.Id                     `json:"id"`
	Name         peripheralSDK.Name                   `json:"name"`
	Capabilities []peripheralSDK.PeripheralCapability `json:"capabilities"`
}

func createPeripheralDescriptor(peripheral peripheralSDK.Peripheral) peripheralDescriptor {
	return peripheralDescriptor{
		Id:           peripheral.GetId(),
		Name:         peripheral.GetName(),
		Capabilities: peripheral.GetCapabilities(),
	}
}

func (descriptor peripheralDescriptor) hasCapability(capability peripheralSDK.PeripheralCapability) bool {
	for _, descriptorCapability := range descriptor.Capabilities {
		if descriptorCapability.Equals(capability) {
			return true
		}
	}
	return false
}

func (descriptor peripheralDescriptor) isDisplaySource() bool {
	return descriptor.hasCapability(peripheralSDK.DisplaySourceCapability)
}

func (descriptor peripheralDescriptor) isDisplaySink() bool {
	return descriptor.hasCapability(peripheralSDK.DisplaySinkCapability)
}
