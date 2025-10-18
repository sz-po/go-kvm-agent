package machine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	clientTransport "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func TestNewService_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	service, err := NewService(roundTripper, machineIdentifier)

	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, roundTripper, service.roundTripper)
	assert.Equal(t, machineIdentifier, service.machineIdentifier)
}

func TestNewService_ValidationError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineIdentifier := machineAPI.MachineIdentifier{}

	service, err := NewService(roundTripper, machineIdentifier)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "machine identifier:")
	assert.Nil(t, service)
}

func TestService_Get_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	responseBodyJSON := []byte(`{"machine":{"id":"test-machine","name":"test-machine"}}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine",
		PathParam: transport.PathParams{
			"machineIdentifier": "id:test-machine",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/json",
		},
		Body: responseBodyJSON,
	}, nil).Once()

	service, err := NewService(roundTripper, machineIdentifier)
	assert.NoError(t, err)

	machine, err := service.Get(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, machine)
	assert.Equal(t, machineId, machine.Id)
}

func TestService_Get_TransportError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	expectedError := errors.New("transport error")

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine",
		PathParam: transport.PathParams{
			"machineIdentifier": "id:test-machine",
		},
	}).Return(nil, expectedError).Once()

	service, err := NewService(roundTripper, machineIdentifier)
	assert.NoError(t, err)

	machine, err := service.Get(context.Background())

	assert.Error(t, err)
	assert.ErrorContains(t, err, "call:")
	assert.Nil(t, machine)
}

func TestService_Peripherals_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	service, err := NewService(roundTripper, machineIdentifier)
	assert.NoError(t, err)

	peripheralIterator, err := service.Peripherals()

	assert.NoError(t, err)
	assert.NotNil(t, peripheralIterator)
}
