package machine

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
)

func TestParseGetRequestPath(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	machineName, err := machineSDK.NewMachineName("machine-name")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		path    transport.Path
		want    *GetRequestPath
		wantErr string
	}{
		{
			name: "machine by id",
			path: transport.Path{
				"machineIdentifier": "id:machine-1",
			},
			want: &GetRequestPath{
				MachineIdentifier: MachineIdentifier{Id: &machineId},
			},
		},
		{
			name: "machine by name",
			path: transport.Path{
				"machineIdentifier": "name:machine-name",
			},
			want: &GetRequestPath{
				MachineIdentifier: MachineIdentifier{Name: &machineName},
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
			requestPath, err := ParseGetRequestPath(testCase.path)

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

func TestParseGetRequest(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		request transport.Request
		want    *GetRequest
		wantErr string
	}{
		{
			name: "valid request with machine by id",
			request: transport.Request{
				Path: transport.Path{
					"machineIdentifier": "id:machine-1",
				},
			},
			want: &GetRequest{
				Path: GetRequestPath{
					MachineIdentifier: MachineIdentifier{Id: &machineId},
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
			request, err := ParseGetRequest(testCase.request)

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

func TestGetResponse_Response(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("test-machine")
	assert.NoError(t, err)

	machineName, err := machineSDK.NewMachineName("test-name")
	assert.NoError(t, err)

	getResponse := &GetResponse{
		Body: GetResponseBody{
			Machine: Machine{
				Id:   machineId,
				Name: machineName,
			},
		},
	}

	response := getResponse.Response()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Header["Content-Type"])
	assert.Equal(t, getResponse.Body, response.Body)
}
