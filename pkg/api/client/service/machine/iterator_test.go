package machine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	clientTransport "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func TestNewIterator(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	iterator, err := NewIterator(roundTripper)

	assert.NoError(t, err)
	assert.NotNil(t, iterator)
	assert.Equal(t, roundTripper, iterator.roundTripper)
}

func TestIterator_List_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId1, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	machineId2, err := machineSDK.NewMachineId("machine-2")
	assert.NoError(t, err)

	machineName1, err := machineSDK.NewMachineName("machine-1")
	assert.NoError(t, err)

	machineName2, err := machineSDK.NewMachineName("machine-2")
	assert.NoError(t, err)

	responseBodyJSON := []byte(`{"machines":[{"id":"machine-1","name":"machine-1"},{"id":"machine-2","name":"machine-2"}]}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine",
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/json",
		},
		Body: responseBodyJSON,
	}, nil).Once()

	iterator, err := NewIterator(roundTripper)
	assert.NoError(t, err)

	machines, err := iterator.List(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, machines)
	assert.Len(t, machines, 2)
	assert.Equal(t, machineId1, machines[0].Id)
	assert.Equal(t, machineName1, machines[0].Name)
	assert.Equal(t, machineId2, machines[1].Id)
	assert.Equal(t, machineName2, machines[1].Name)
}

func TestIterator_List_EmptyList(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	responseBodyJSON := []byte(`{"machines":[]}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine",
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/json",
		},
		Body: responseBodyJSON,
	}, nil).Once()

	iterator, err := NewIterator(roundTripper)
	assert.NoError(t, err)

	machines, err := iterator.List(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, machines)
	assert.Len(t, machines, 0)
}

func TestIterator_List_TransportError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	expectedError := errors.New("transport error")

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine",
	}).Return(nil, expectedError).Once()

	iterator, err := NewIterator(roundTripper)
	assert.NoError(t, err)

	machines, err := iterator.List(context.Background())

	assert.Error(t, err)
	assert.ErrorContains(t, err, "call:")
	assert.Nil(t, machines)
}
