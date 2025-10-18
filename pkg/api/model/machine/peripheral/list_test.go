package peripheral

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	machineAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/model/machine"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
	machineSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/machine"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
)

func TestListRequestPath_String(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	requestPath := ListRequestPath{
		MachineIdentifier: machineAPI.MachineIdentifier{Id: &machineId},
	}

	path, err := requestPath.String()

	assert.NoError(t, err)
	if assert.NotNil(t, path) {
		assert.Equal(t, "/machine/id:machine-1/peripheral", *path)
	}
}

func TestListRequestPath_StringInvalid(t *testing.T) {
	requestPath := ListRequestPath{}

	path, err := requestPath.String()

	assert.Error(t, err)
	assert.Nil(t, path)
}

func TestListRequestPath_Params(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	requestPath := ListRequestPath{
		MachineIdentifier: machineAPI.MachineIdentifier{Id: &machineId},
	}

	params, err := requestPath.Params()

	assert.NoError(t, err)
	if assert.NotNil(t, params) {
		assert.Equal(t, transport.PathParams{
			machineAPI.MachineIdentifierPathFieldName: "id:machine-1",
		}, *params)
	}
}

func TestListRequestPath_ParamsInvalid(t *testing.T) {
	requestPath := ListRequestPath{}

	params, err := requestPath.Params()

	assert.Error(t, err)
	assert.Nil(t, params)
}

func TestListRequest_Request(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	request := &ListRequest{
		Path: ListRequestPath{
			MachineIdentifier: machineAPI.MachineIdentifier{Id: &machineId},
		},
	}

	transportRequest, err := request.Request()

	assert.NoError(t, err)
	if assert.NotNil(t, transportRequest) {
		assert.Equal(t, http.MethodGet, transportRequest.Method)
		assert.Equal(t, "/machine/id:machine-1/peripheral", transportRequest.Path)
		assert.Equal(t, transport.PathParams{
			machineAPI.MachineIdentifierPathFieldName: "id:machine-1",
		}, transportRequest.PathParam)
	}
}

func TestListRequest_RequestInvalid(t *testing.T) {
	request := &ListRequest{}

	transportRequest, err := request.Request()

	assert.Error(t, err)
	assert.Nil(t, transportRequest)
	assert.ErrorContains(t, err, "path")
}

func TestParseListResponseBody(t *testing.T) {
	body, err := ParseListResponseBody(strings.NewReader("{\"result\":[],\"totalCount\":0}"))

	assert.NoError(t, err)
	if assert.NotNil(t, body) {
		assert.Empty(t, body.Peripherals)
		assert.Equal(t, 0, body.TotalCount)
	}
}

func TestParseListResponse(t *testing.T) {
	responseBody := strings.NewReader("{\"peripherals\":[{\"id\":\"peripheral-1\",\"name\":\"peripheral-name\",\"capability\":[]}],\"totalCount\":1}")

	response, err := ParseListResponse(transport.Response{Body: responseBody})

	assert.NoError(t, err)
	if assert.NotNil(t, response) {
		assert.Len(t, response.Body.Peripherals, 1)
		assert.Equal(t, 1, response.Body.TotalCount)
	}
}

func TestParseListRequestPath(t *testing.T) {
	machineId, err := machineSDK.NewMachineId("machine-1")
	assert.NoError(t, err)

	machineName, err := machineSDK.NewMachineName("machine-name")
	assert.NoError(t, err)

	testCases := []struct {
		name    string
		path    transport.PathParams
		want    *ListRequestPath
		wantErr string
	}{
		{
			name: "machine by id",
			path: transport.PathParams{
				"machineIdentifier": "id:machine-1",
			},
			want: &ListRequestPath{
				MachineIdentifier: machineAPI.MachineIdentifier{Id: &machineId},
			},
		},
		{
			name: "machine by name",
			path: transport.PathParams{
				"machineIdentifier": "name:machine-name",
			},
			want: &ListRequestPath{
				MachineIdentifier: machineAPI.MachineIdentifier{Name: &machineName},
			},
		},
		{
			name:    "missing machine identifier",
			path:    transport.PathParams{},
			wantErr: "missing path param key: machineIdentifier",
		},
		{
			name: "invalid machine identifier prefix",
			path: transport.PathParams{
				"machineIdentifier": "uuid:machine-1",
			},
			wantErr: "parse machine identifier: unknown machine type: uuid",
		},
		{
			name: "invalid machine identifier format",
			path: transport.PathParams{
				"machineIdentifier": "invalid",
			},
			wantErr: "parse machine identifier: missing identifier type",
		},
		{
			name: "invalid machine id value",
			path: transport.PathParams{
				"machineIdentifier": "id:INVALID_UPPER",
			},
			wantErr: "parse machine identifier: invalid machine id",
		},
		{
			name: "invalid machine name value",
			path: transport.PathParams{
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
				PathParam: transport.PathParams{
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
				PathParam: transport.PathParams{},
			},
			wantErr: "parse path: missing path param key: machineIdentifier",
		},
		{
			name: "invalid request with invalid machine identifier",
			request: transport.Request{
				PathParam: transport.PathParams{
					"machineIdentifier": "invalid",
				},
			},
			wantErr: "parse path: parse machine identifier: missing identifier type",
		},
		{
			name: "invalid request with unknown prefix",
			request: transport.Request{
				PathParam: transport.PathParams{
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
			Peripherals: []Peripheral{
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
	assert.Equal(t, "application/json", response.Header[transport.HeaderContentType])
	assert.Equal(t, listResponse.Body, response.Body)
}
