package peripheral

import (
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

const PeripheralRepositoryServiceId = nodeSDK.ServiceId("node/peripheral/repository")

const (
	RepositoryGetPeripheralByIdMethod    nodeSDK.MethodName = "get-peripheral-by-id"
	RepositoryGetPeripheralByNameMethod  nodeSDK.MethodName = "get-peripheral-by-name"
	RepositoryGetAllPeripheralsMethod    nodeSDK.MethodName = "get-all-peripherals"
	RepositoryGetAllDisplaySourcesMethod nodeSDK.MethodName = "get-all-display-sources"
	RepositoryGetAllDisplaySinksMethod   nodeSDK.MethodName = "get-all-display-sinks"
)

type RepositoryGetPeripheralByIdRequest struct {
	Id peripheralSDK.Id `json:"id"`
}

type RepositoryGetPeripheralByIdResponse struct {
	Peripheral peripheralDescriptor `json:"peripheral"`
}

type RepositoryGetPeripheralByNameRequest struct {
	Name peripheralSDK.Name `json:"name"`
}

type RepositoryGetPeripheralByNameResponse struct {
	Peripheral peripheralDescriptor `json:"peripheral"`
}

type RepositoryGetAllPeripheralsRequest struct {
}

type RepositoryGetAllPeripheralsResponse struct {
	Peripherals []peripheralDescriptor `json:"peripherals"`
}

type RepositoryGetAllDisplaySourcesRequest struct {
}

type RepositoryGetAllDisplaySourcesResponse struct {
	DisplaySources []peripheralDescriptor `json:"displaySources"`
}

type RepositoryGetAllDisplaySinksRequest struct {
}

type RepositoryGetAllDisplaySinksResponse struct {
	DisplaySinks []peripheralDescriptor `json:"displaySinks"`
}
