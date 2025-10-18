package peripheral

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	clientTransport "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestNewService_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, roundTripper, service.roundTripper)
	assert.Equal(t, machineIdentifier, service.machineIdentifier)
	assert.Equal(t, peripheralIdentifier, service.peripheralIdentifier)
}

func TestNewService_MachineIdentifierValidationError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "machine identifier:")
	assert.Nil(t, service)
}

func TestNewService_PeripheralIdentifierValidationError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{}

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "peripheral identifier:")
	assert.Nil(t, service)
}

func TestService_Get_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	responseBodyJSON := []byte(`{"peripheral":{"id":"test-peripheral","name":"test-peripheral","capabilities":[]}}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/json",
		},
		Body: responseBodyJSON,
	}, nil).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	peripheral, err := service.Get(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, peripheral)
	assert.Equal(t, peripheralId, peripheral.Id)
}

func TestService_Get_TransportError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	expectedError := errors.New("transport error")

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
	}).Return(nil, expectedError).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	peripheral, err := service.Get(context.Background())

	assert.Error(t, err)
	assert.ErrorContains(t, err, "call:")
	assert.Nil(t, peripheral)
}

func TestService_DisplaySource_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	displaySourceService, err := service.DisplaySource()

	assert.NoError(t, err)
	assert.NotNil(t, displaySourceService)
}
