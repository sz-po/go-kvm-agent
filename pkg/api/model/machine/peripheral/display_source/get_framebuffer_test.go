package display_source

import (
	"bytes"
	"io"
	"testing"

	"github.com/elnormous/contenttype"
	"github.com/stretchr/testify/assert"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestParseGetFramebufferRequestPath(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)
	peripheralId, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	machineName, err := machineSDK.NewMachineName("machine-name")
	assert.NoError(t, err)
	peripheralName, err := peripheralSDK.NewPeripheralName("peripheral-name")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		path    transport.PathParams
		want    *GetFramebufferRequestPath
		wantErr string
	}{
		{
			name: "machine and peripheral by id",
			path: transport.PathParams{
				"machineIdentifier":    "id:machine-1",
				"peripheralIdentifier": "id:peripheral-1",
			},
			want: &GetFramebufferRequestPath{
				MachineIdentifier:    machine.MachineIdentifier{Id: &machineId},
				PeripheralIdentifier: peripheral.PeripheralIdentifier{Id: &peripheralId},
			},
		},
		{
			name: "machine and peripheral by name",
			path: transport.PathParams{
				"machineIdentifier":    "name:machine-name",
				"peripheralIdentifier": "name:peripheral-name",
			},
			want: &GetFramebufferRequestPath{
				MachineIdentifier:    machine.MachineIdentifier{Name: &machineName},
				PeripheralIdentifier: peripheral.PeripheralIdentifier{Name: &peripheralName},
			},
		},
		{
			name: "missing machine identifier",
			path: transport.PathParams{
				"peripheralIdentifier": "id:peripheral-1",
			},
			wantErr: "missing path param key: machineIdentifier",
		},
		{
			name: "missing peripheral identifier",
			path: transport.PathParams{
				"machineIdentifier": "id:machine-1",
			},
			wantErr: "missing path param key: peripheralIdentifier",
		},
		{
			name: "invalid machine identifier prefix",
			path: transport.PathParams{
				"machineIdentifier":    "uuid:machine-1",
				"peripheralIdentifier": "id:peripheral-1",
			},
			wantErr: "parse machine identifier: unknown machine type: uuid",
		},
		{
			name: "invalid peripheral identifier prefix",
			path: transport.PathParams{
				"machineIdentifier":    "id:machine-1",
				"peripheralIdentifier": "uuid:peripheral-1",
			},
			wantErr: "parse peripheral identifier: unknown peripheral type: uuid",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			requestPath, err := ParseGetFramebufferRequestPath(testCase.path)

			if testCase.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.wantErr)
				assert.Nil(t, requestPath)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, requestPath) {
				assert.Equal(t, testCase.want.MachineIdentifier, requestPath.MachineIdentifier)
				assert.Equal(t, testCase.want.PeripheralIdentifier, requestPath.PeripheralIdentifier)
			}
		})
	}
}

func TestParseGetFramebufferRequest(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)
	peripheralId, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		request transport.Request
		want    *GetFramebufferRequest
		wantErr string
	}{
		{
			name: "valid request with machine and peripheral by id",
			request: transport.Request{
				PathParam: transport.PathParams{
					"machineIdentifier":    "id:machine-1",
					"peripheralIdentifier": "id:peripheral-1",
				},
				Header: transport.Header{
					transport.HeaderAccept: "application/x-rgb24",
				},
			},
			want: &GetFramebufferRequest{
				Path: GetFramebufferRequestPath{
					MachineIdentifier:    machine.MachineIdentifier{Id: &machineId},
					PeripheralIdentifier: peripheral.PeripheralIdentifier{Id: &peripheralId},
				},
				Headers: GetFramebufferRequestHeaders{
					Accept: "application/x-rgb24",
				},
				MediaType: FramebufferMediaTypeRGB24,
			},
		},
		{
			name: "invalid request with missing machine identifier",
			request: transport.Request{
				PathParam: transport.PathParams{
					"peripheralIdentifier": "id:peripheral-1",
				},
				Header: transport.Header{
					transport.HeaderAccept: "application/x-rgb24",
				},
			},
			wantErr: "parse path: missing path param key: machineIdentifier",
		},
		{
			name: "invalid request with missing peripheral identifier",
			request: transport.Request{
				PathParam: transport.PathParams{
					"machineIdentifier": "id:machine-1",
				},
				Header: transport.Header{
					transport.HeaderAccept: "application/x-rgb24",
				},
			},
			wantErr: "parse path: missing path param key: peripheralIdentifier",
		},
		{
			name: "invalid request with invalid machine identifier",
			request: transport.Request{
				PathParam: transport.PathParams{
					"machineIdentifier":    "invalid",
					"peripheralIdentifier": "id:peripheral-1",
				},
				Header: transport.Header{
					transport.HeaderAccept: "application/x-rgb24",
				},
			},
			wantErr: "parse path: parse machine identifier: missing identifier type",
		},
		{
			name: "invalid request with invalid peripheral identifier",
			request: transport.Request{
				PathParam: transport.PathParams{
					"machineIdentifier":    "id:machine-1",
					"peripheralIdentifier": "invalid",
				},
				Header: transport.Header{
					transport.HeaderAccept: "application/x-rgb24",
				},
			},
			wantErr: "parse path: parse peripheral identifier: missing identifier type",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, err := ParseGetFramebufferRequest(testCase.request, []contenttype.MediaType{FramebufferMediaTypeRGB24})

			if testCase.wantErr != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.wantErr)
				assert.Nil(t, request)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, request) {
				assert.Equal(t, testCase.want.Path.MachineIdentifier, request.Path.MachineIdentifier)
				assert.Equal(t, testCase.want.Path.PeripheralIdentifier, request.Path.PeripheralIdentifier)
				assert.Equal(t, testCase.want.Headers.Accept, request.Headers.Accept)
				assert.Equal(t, testCase.want.MediaType, request.MediaType)
			}
		})
	}
}

type mockWriterTo struct {
	data []byte
}

func (mock *mockWriterTo) WriteTo(writer io.Writer) (int64, error) {
	bytesWritten, err := writer.Write(mock.data)
	return int64(bytesWritten), err
}

type mockReadCloser struct {
	reader *bytes.Reader
}

func (mock *mockReadCloser) Read(p []byte) (n int, err error) {
	return mock.reader.Read(p)
}

func (mock *mockReadCloser) Close() error {
	return nil
}

func TestParseGetFramebufferResponse(t *testing.T) {
	t.Run("with io.WriterTo body", func(t *testing.T) {
		testData := []byte("test framebuffer data")
		mockWriter := &mockWriterTo{data: testData}

		response := transport.Response{
			StatusCode: 200,
			Header: transport.Header{
				"content-type": "application/x-rgb24",
			},
			Body: mockWriter,
		}

		parsed, err := ParseGetFramebufferResponse(response)

		assert.NoError(t, err)
		assert.NotNil(t, parsed)
		assert.Equal(t, mockWriter, parsed.Body)
		assert.Equal(t, "application", parsed.Headers.ContentType.Type)
		assert.Equal(t, "x-rgb24", parsed.Headers.ContentType.Subtype)

		buffer := &bytes.Buffer{}
		bytesWritten, err := parsed.Body.WriteTo(buffer)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(testData)), bytesWritten)
		assert.Equal(t, testData, buffer.Bytes())
	})

	t.Run("with io.ReadCloser body", func(t *testing.T) {
		testData := []byte("framebuffer data from reader")
		readCloser := &mockReadCloser{reader: bytes.NewReader(testData)}

		response := transport.Response{
			StatusCode: 200,
			Header: transport.Header{
				"content-type": "application/x-rgb24",
			},
			Body: readCloser,
		}

		parsed, err := ParseGetFramebufferResponse(response)

		assert.NoError(t, err)
		assert.NotNil(t, parsed)
		assert.IsType(t, transport.ResponseWriterTo{}, parsed.Body)
		assert.Equal(t, "application", parsed.Headers.ContentType.Type)
		assert.Equal(t, "x-rgb24", parsed.Headers.ContentType.Subtype)

		buffer := &bytes.Buffer{}
		bytesWritten, err := parsed.Body.WriteTo(buffer)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(testData)), bytesWritten)
		assert.Equal(t, testData, buffer.Bytes())
	})

	t.Run("with unsupported body type", func(t *testing.T) {
		response := transport.Response{
			StatusCode: 200,
			Header: transport.Header{
				"content-type": "application/x-rgb24",
			},
			Body: "invalid body type",
		}

		parsed, err := ParseGetFramebufferResponse(response)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "unsupported response body type:")
		assert.Nil(t, parsed)
	})

	t.Run("with missing content-type header", func(t *testing.T) {
		testData := []byte("test data")
		mockWriter := &mockWriterTo{data: testData}

		response := transport.Response{
			StatusCode: 200,
			Header:     transport.Header{},
			Body:       mockWriter,
		}

		parsed, err := ParseGetFramebufferResponse(response)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "parse headers:")
		assert.Nil(t, parsed)
	})
}
