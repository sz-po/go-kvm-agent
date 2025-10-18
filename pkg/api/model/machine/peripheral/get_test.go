package peripheral

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestGetRequestPath_String(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	machineName, err := machineSDK.NewMachineName("machine-name")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	peripheralName, err := peripheralSDK.NewPeripheralName("peripheral-name")
	assert.NoError(t, err)

	testCases := []struct {
		name        string
		requestPath GetRequestPath
		want        string
	}{
		{
			name: "machine and peripheral by id",
			requestPath: GetRequestPath{
				MachineIdentifier:    machineAPI.MachineIdentifier{Id: &machineId},
				PeripheralIdentifier: PeripheralIdentifier{Id: &peripheralId},
			},
			want: "/machine/id:machine-1/peripheral/id:peripheral-1",
		},
		{
			name: "machine and peripheral by name",
			requestPath: GetRequestPath{
				MachineIdentifier:    machineAPI.MachineIdentifier{Name: &machineName},
				PeripheralIdentifier: PeripheralIdentifier{Name: &peripheralName},
			},
			want: "/machine/name:machine-name/peripheral/name:peripheral-name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path, err := testCase.requestPath.String()
			assert.NoError(t, err)

			if assert.NotNil(t, path) {
				assert.Equal(t, testCase.want, *path)
			}
		})
	}
}

func TestGetRequestPath_StringInvalid(t *testing.T) {
	requestPath := GetRequestPath{}

	path, err := requestPath.String()

	assert.Error(t, err)
	assert.Nil(t, path)
}

func TestGetRequestPath_Params(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	requestPath := GetRequestPath{
		MachineIdentifier:    machineAPI.MachineIdentifier{Id: &machineId},
		PeripheralIdentifier: PeripheralIdentifier{Id: &peripheralId},
	}

	params, err := requestPath.Params()

	assert.NoError(t, err)
	if assert.NotNil(t, params) {
		assert.Equal(t, transport.PathParams{
			machineAPI.MachineIdentifierPathFieldName: "id:machine-1",
			PeripheralIdentifierPathFieldName:         "id:peripheral-1",
		}, *params)
	}
}

func TestGetRequestPath_ParamsInvalid(t *testing.T) {
	requestPath := GetRequestPath{}

	params, err := requestPath.Params()

	assert.Error(t, err)
	assert.Nil(t, params)
}

func TestGetRequest_Request(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	peripheralId, err := peripheralSDK.NewPeripheralId("peripheral-1")
	assert.NoError(t, err)

	request := &GetRequest{
		Path: GetRequestPath{
			MachineIdentifier:    machineAPI.MachineIdentifier{Id: &machineId},
			PeripheralIdentifier: PeripheralIdentifier{Id: &peripheralId},
		},
	}

	transportRequest, err := request.Request()

	assert.NoError(t, err)
	if assert.NotNil(t, transportRequest) {
		assert.Equal(t, http.MethodGet, transportRequest.Method)
		assert.Equal(t, "/machine/id:machine-1/peripheral/id:peripheral-1", transportRequest.Path)
		assert.Equal(t, transport.PathParams{
			machineAPI.MachineIdentifierPathFieldName: "id:machine-1",
			PeripheralIdentifierPathFieldName:         "id:peripheral-1",
		}, transportRequest.PathParam)
	}
}

func TestGetRequest_RequestInvalid(t *testing.T) {
	request := &GetRequest{}

	transportRequest, err := request.Request()

	assert.Error(t, err)
	assert.Nil(t, transportRequest)
	assert.ErrorContains(t, err, "path")
}
