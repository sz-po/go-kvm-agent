package machine

import (
	"context"
	"fmt"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/api/client"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/peripheral"
	machineClient "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/service/machine"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

type RemoteMachineSelectorConfig struct {
	Id   *machineSDK.MachineId   `json:"id"`
	Name *machineSDK.MachineName `json:"name"`
}

type RemoteMachineConfig struct {
	Connection client.Config               `json:"connection"`
	Selector   RemoteMachineSelectorConfig `json:"selector"`
}

type RemoteMachine struct {
	apiMachineService *machineClient.Service

	localName machineSDK.MachineName

	remoteId   machineSDK.MachineId
	remoteName machineSDK.MachineName

	peripheralsRepository *peripheral.RemoteRepository
}

var _ machineSDK.Machine = (*RemoteMachine)(nil)

func NewRemoteMachine(ctx context.Context, name machineSDK.MachineName, config RemoteMachineConfig) (*RemoteMachine, error) {
	machineIdentifier := machineAPI.MachineIdentifier{
		Id:   config.Selector.Id,
		Name: config.Selector.Name,
	}
	if err := machineIdentifier.Validate(); err != nil {
		return nil, fmt.Errorf("machine identifier: %w", err)
	}

	apiClient, err := client.CreateClientFromConfig(config.Connection)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	machineService, err := apiClient.Machine(machineIdentifier)
	if err != nil {
		return nil, fmt.Errorf("machine service: %w", err)
	}

	apiMachine, err := machineService.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get machine: %w", err)
	}

	peripheralRepository, err := peripheral.NewRemoteRepository(ctx, machineService)

	return &RemoteMachine{
		apiMachineService: machineService,

		localName: name,

		remoteName: apiMachine.Name,
		remoteId:   apiMachine.Id,

		peripheralsRepository: peripheralRepository,
	}, nil
}

func (machine *RemoteMachine) GetName() machineSDK.MachineName {
	return machine.localName
}

func (machine *RemoteMachine) GetId() machineSDK.MachineId {
	return machine.remoteId
}

func (machine *RemoteMachine) Peripherals() peripheralSDK.Repository {
	return machine.peripheralsRepository
}

func (machine *RemoteMachine) Terminate(ctx context.Context) error {
	return nil
}
