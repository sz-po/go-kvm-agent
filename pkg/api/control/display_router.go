package control

import "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"

type DisplayRouterConnectRequest struct {
	DisplaySourceId peripheral.PeripheralId `json:"displaySourceId"`
	DisplaySinkId   peripheral.PeripheralId `json:"displaySinkId"`
}
