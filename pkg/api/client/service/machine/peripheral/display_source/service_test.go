package display_source

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	clientTransport "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/client/transport"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	peripheralAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	memorySDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/memory"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

// mockWriterTo is a test helper that implements io.WriterTo interface.
type mockWriterTo struct {
	data     []byte
	writeErr error
}

func (mock *mockWriterTo) WriteTo(writer io.Writer) (int64, error) {
	if mock.writeErr != nil {
		return 0, mock.writeErr
	}
	bytesWritten, err := writer.Write(mock.data)
	return int64(bytesWritten), err
}

func TestNewService(t *testing.T) {
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

func TestGetDisplayMode_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	expectedDisplayMode := peripheralSDK.DisplayMode{
		Width:  1920,
		Height: 1080,
	}

	responseBodyJSON := []byte(`{"displayMode":{"width":1920,"height":1080}}`)

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/display-mode",
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

	displayMode, err := service.GetDisplayMode(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, displayMode)
	assert.Equal(t, expectedDisplayMode.Width, displayMode.Width)
	assert.Equal(t, expectedDisplayMode.Height, displayMode.Height)
}

func TestGetDisplayMode_TransportError(t *testing.T) {
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
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/display-mode",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
	}).Return(nil, expectedError).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	displayMode, err := service.GetDisplayMode(context.Background())

	assert.Error(t, err)
	assert.ErrorContains(t, err, "call:")
	assert.Nil(t, displayMode)
}

func TestGetFramebuffer_Success(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)
	memoryBuffer := memorySDK.NewBufferMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	testData := []byte("test framebuffer data")
	bufferSize := 1024

	mockWriter := &mockWriterTo{
		data:     testData,
		writeErr: nil,
	}

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/x-rgb24",
		},
		Body: mockWriter,
	}, nil).Once()

	memoryPool.EXPECT().Borrow(bufferSize).Return(memoryBuffer, nil).Once()
	memoryBuffer.EXPECT().Write(testData).Return(len(testData), nil).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.NoError(t, err)
	assert.NotNil(t, frameBuffer)
}

func TestGetFramebuffer_TransportError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	bufferSize := 1024
	expectedError := errors.New("transport error")

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(nil, expectedError).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "call:")
	assert.Nil(t, frameBuffer)
}

func TestGetFramebuffer_BorrowError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	bufferSize := 1024
	mockWriter := &mockWriterTo{
		data:     []byte("test data"),
		writeErr: nil,
	}

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/x-rgb24",
		},
		Body: mockWriter,
	}, nil).Once()

	borrowError := errors.New("borrow error")
	memoryPool.EXPECT().Borrow(bufferSize).Return(nil, borrowError).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "borrow memory buffer:")
	assert.Nil(t, frameBuffer)
}

func TestGetFramebuffer_WriteToError_ReleaseSuccess(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)
	memoryBuffer := memorySDK.NewBufferMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	bufferSize := 1024
	writeError := errors.New("write error")

	mockWriter := &mockWriterTo{
		data:     []byte("test data"),
		writeErr: writeError,
	}

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/x-rgb24",
		},
		Body: mockWriter,
	}, nil).Once()

	memoryPool.EXPECT().Borrow(bufferSize).Return(memoryBuffer, nil).Once()
	memoryBuffer.EXPECT().Release().Return(nil).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "write data to memory buffer:")
	assert.Nil(t, frameBuffer)
}

func TestGetFramebuffer_WriteToError_ReleaseError(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)
	memoryBuffer := memorySDK.NewBufferMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	bufferSize := 1024
	writeError := errors.New("write error")
	releaseError := errors.New("release error")

	mockWriter := &mockWriterTo{
		data:     []byte("test data"),
		writeErr: writeError,
	}

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/x-rgb24",
		},
		Body: mockWriter,
	}, nil).Once()

	memoryPool.EXPECT().Borrow(bufferSize).Return(memoryBuffer, nil).Once()
	memoryBuffer.EXPECT().Release().Return(releaseError).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "write data to memory buffer:")
	assert.Nil(t, frameBuffer)
}

func TestGetFramebuffer_BufferWrapping(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)
	memoryBuffer := memorySDK.NewBufferMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	testData := []byte("test framebuffer data")
	bufferSize := 1024
	expectedCapacity := 2048
	expectedSize := len(testData)

	mockWriter := &mockWriterTo{
		data:     testData,
		writeErr: nil,
	}

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/x-rgb24",
		},
		Body: mockWriter,
	}, nil).Once()

	memoryPool.EXPECT().Borrow(bufferSize).Return(memoryBuffer, nil).Once()
	memoryBuffer.EXPECT().Write(testData).Return(len(testData), nil).Once()
	memoryBuffer.EXPECT().GetCapacity().Return(expectedCapacity).Once()
	memoryBuffer.EXPECT().GetSize().Return(expectedSize).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.NoError(t, err)
	assert.NotNil(t, frameBuffer)
	assert.Equal(t, expectedCapacity, frameBuffer.GetCapacity())
	assert.Equal(t, expectedSize, frameBuffer.GetSize())
}

func TestGetFramebuffer_WithReadCloser(t *testing.T) {
	t.Parallel()

	roundTripper := clientTransport.NewRoundTripperMock(t)
	memoryPool := memorySDK.NewPoolMock(t)
	memoryBuffer := memorySDK.NewBufferMock(t)

	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	machineIdentifier := machineAPI.MachineIdentifier{Id: &machineId}
	peripheralIdentifier := peripheralAPI.PeripheralIdentifier{Id: &peripheralId}

	testData := []byte("framebuffer data from reader")
	bufferSize := 2048

	readCloser := io.NopCloser(bytes.NewBuffer(testData))

	roundTripper.EXPECT().Call(context.Background(), transport.Request{
		Method: "GET",
		Path:   "/machine/id:test-machine/peripheral/id:test-peripheral/display-source/framebuffer",
		PathParam: transport.PathParams{
			"machineIdentifier":    "id:test-machine",
			"peripheralIdentifier": "id:test-peripheral",
		},
		Header: transport.Header{
			"accept": "",
		},
	}).Return(&transport.Response{
		StatusCode: 200,
		Header: transport.Header{
			"content-type": "application/x-rgb24",
		},
		Body: readCloser,
	}, nil).Once()

	memoryPool.EXPECT().Borrow(bufferSize).Return(memoryBuffer, nil).Once()
	memoryBuffer.EXPECT().Write(testData).Return(len(testData), nil).Once()

	service, err := NewService(roundTripper, machineIdentifier, peripheralIdentifier)
	assert.NoError(t, err)

	frameBuffer, err := service.GetFramebuffer(context.Background(), memoryPool, bufferSize)

	assert.NoError(t, err)
	assert.NotNil(t, frameBuffer)
}
