package peripheral

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	clientTransport "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestNewIterator_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	iterator, err := NewIterator(roundTripper, machineIdentifier)

	assert.NoError(t, err)
	assert.NotNil(t, iterator)
	assert.Equal(t, roundTripper, iterator.roundTripper)
	assert.Equal(t, machineIdentifier, iterator.machineIdentifier)
}

func TestNewIterator_ValidationError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineIdentifier := machineAPI.MachineIdentifier{}

	iterator, err := NewIterator(roundTripper, machineIdentifier)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "machine identifier:")
	assert.Nil(t, iterator)
}

func TestIterator_List_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId1, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	peripheralId2, err := peripheralSDK.NewPeripheralId("peripheral-2")
	assert.NoError(t, err)

	peripheralName1, err := peripheralSDK.NewPeripheralName("peripheral-1")
	assert.NoError(t, err)

	peripheralName2, err := peripheralSDK.NewPeripheralName("peripheral-2")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	responseBodyJSON := []byte(`{"peripherals":[{"id":"peripheral-1","name":"peripheral-1","capabilities":[]},{"id":"peripheral-2","name":"peripheral-2","capabilities":[]}]}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral",
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

	iterator, err := NewIterator(roundTripper, machineIdentifier)
	assert.NoError(t, err)

	peripherals, err := iterator.List(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, peripherals)
	assert.Len(t, peripherals, 2)
	assert.Equal(t, peripheralId1, peripherals[0].Id)
	assert.Equal(t, peripheralName1, peripherals[0].Name)
	assert.Equal(t, peripheralId2, peripherals[1].Id)
	assert.Equal(t, peripheralName2, peripherals[1].Name)
}

func TestIterator_List_EmptyList(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	responseBodyJSON := []byte(`{"peripherals":[]}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral",
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

	iterator, err := NewIterator(roundTripper, machineIdentifier)
	assert.NoError(t, err)

	peripherals, err := iterator.List(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, peripherals)
	assert.Len(t, peripherals, 0)
}

func TestIterator_List_TransportError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}

	expectedError := errors.New("transport error")

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral",
		PathParam: transport.PathParams{
			"machineIdentifier": "id:test-machine",
		},
	}).Return(nil, expectedError).Once()

	iterator, err := NewIterator(roundTripper, machineIdentifier)
	assert.NoError(t, err)

	peripherals, err := iterator.List(context.Background())

	assert.Error(t, err)
	assert.ErrorContains(t, err, "call:")
	assert.Nil(t, peripherals)
}
