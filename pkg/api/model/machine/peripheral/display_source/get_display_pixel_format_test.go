package display_source

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine/peripheral"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestParseGetDisplayPixelFormatRequestPath(t *testing.T) {
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
		want    *GetDisplayPixelFormatRequestPath
		wantErr string
	}{
		{
			name: "machine and peripheral by id",
			path: transport.PathParams{
				"machineIdentifier":    "id:machine-1",
				"peripheralIdentifier": "id:peripheral-1",
			},
			want: &GetDisplayPixelFormatRequestPath{
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
			want: &GetDisplayPixelFormatRequestPath{
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
			requestPath, err := ParseGetDisplayPixelFormatRequestPath(testCase.path)

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

func TestParseGetDisplayPixelFormatRequest(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)
	peripheralId, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		request transport.Request
		want    *GetDisplayPixelFormatRequest
		wantErr string
	}{
		{
			name: "valid request with machine and peripheral by id",
			request: transport.Request{
				PathParam: transport.PathParams{
					"machineIdentifier":    "id:machine-1",
					"peripheralIdentifier": "id:peripheral-1",
				},
			},
			want: &GetDisplayPixelFormatRequest{
				Path: GetDisplayPixelFormatRequestPath{
					MachineIdentifier:    machine.MachineIdentifier{Id: &machineId},
					PeripheralIdentifier: peripheral.PeripheralIdentifier{Id: &peripheralId},
				},
			},
		},
		{
			name: "invalid request with missing machine identifier",
			request: transport.Request{
				PathParam: transport.PathParams{
					"peripheralIdentifier": "id:peripheral-1",
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
			},
			wantErr: "parse path: parse peripheral identifier: missing identifier type",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, err := ParseGetDisplayPixelFormatRequest(testCase.request)

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
			}
		})
	}
}
