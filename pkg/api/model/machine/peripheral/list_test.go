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

func TestParseListRequestPath(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	machineName, err := machineSDK.NewMachineName("machine-name")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		path    transport.Path
		want    *ListRequestPath
		wantErr string
	}{
		{
			name: "machine by id",
			path: transport.Path{
				"machineIdentifier": "id:machine-1",
			},
			want: &ListRequestPath{
				MachineIdentifier: machineAPI.MachineIdentifier{Id: &machineId},
			},
		},
		{
			name: "machine by name",
			path: transport.Path{
				"machineIdentifier": "name:machine-name",
			},
			want: &ListRequestPath{
				MachineIdentifier: machineAPI.MachineIdentifier{Name: &machineName},
			},
		},
		{
			name:    "missing machine identifier",
			path:    transport.Path{},
			wantErr: "missing path key: machineIdentifier",
		},
		{
			name: "invalid machine identifier prefix",
			path: transport.Path{
				"machineIdentifier": "uuid:machine-1",
			},
			wantErr: "parse machine identifier: unknown machine type: uuid",
		},
		{
			name: "invalid machine identifier format",
			path: transport.Path{
				"machineIdentifier": "invalid",
			},
			wantErr: "parse machine identifier: missing identifier type",
		},
		{
			name: "invalid machine id value",
			path: transport.Path{
				"machineIdentifier": "id:INVALID_UPPER",
			},
			wantErr: "parse machine identifier: invalid machine id",
		},
		{
			name: "invalid machine name value",
			path: transport.Path{
				"machineIdentifier": "name:INVALID_UPPER",
			},
			wantErr: "parse machine identifier: invalid machine name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			requestPath, err := ParseListRequestPath(testCase.path)

			if testCase.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, testCase.wantErr)
				assert.Nil(t, requestPath)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, requestPath) {
				assert.Equal(t, testCase.want.MachineIdentifier, requestPath.MachineIdentifier)
			}
		})
	}
}

func TestParseListRequest(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		request transport.Request
		want    *ListRequest
		wantErr string
	}{
		{
			name: "valid request with machine by id",
			request: transport.Request{
				Path: transport.Path{
					"machineIdentifier": "id:machine-1",
				},
			},
			want: &ListRequest{
				Path: ListRequestPath{
					MachineIdentifier: machineAPI.MachineIdentifier{Id: &machineId},
				},
			},
		},
		{
			name: "invalid request with missing machine identifier",
			request: transport.Request{
				Path: transport.Path{},
			},
			wantErr: "parse path: missing path key: machineIdentifier",
		},
		{
			name: "invalid request with invalid machine identifier",
			request: transport.Request{
				Path: transport.Path{
					"machineIdentifier": "invalid",
				},
			},
			wantErr: "parse path: parse machine identifier: missing identifier type",
		},
		{
			name: "invalid request with unknown prefix",
			request: transport.Request{
				Path: transport.Path{
					"machineIdentifier": "uuid:machine-1",
				},
			},
			wantErr: "parse path: parse machine identifier: unknown machine type: uuid",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, err := ParseListRequest(testCase.request)

			if testCase.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, testCase.wantErr)
				assert.Nil(t, request)
				return
			}

			assert.NoError(t, err)
			if assert.NotNil(t, request) {
				assert.Equal(t, testCase.want.Path.MachineIdentifier, request.Path.MachineIdentifier)
			}
		})
	}
}

func TestListResponse_Response(t *testing.T) {
	peripheralId, err := peripheralSDK.NewPeripheralId("test-peripheral")
	assert.NoError(t, err)

	peripheralName, err := peripheralSDK.NewPeripheralName("test-name")
	assert.NoError(t, err)

	listResponse := ListResponse{
		Body: ListResponseBody{
			Result: []Peripheral{
				{
					Id:   peripheralId,
					Name: peripheralName,
				},
			},
			TotalCount: 1,
		},
	}

	response := listResponse.Response()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Header["Content-Type"])
	assert.Equal(t, listResponse.Body, response.Body)
}
