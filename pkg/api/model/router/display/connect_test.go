package display

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func TestParseConnectRequestSuccess(t *testing.T) {
	requestBody := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "name": "sink-machine"
            },
            "peripheralIdentifier": {
                "name": "sink-peripheral"
            }
        }
    }`

	request := transport.Request{Body: strings.NewReader(requestBody)}

	connectRequest, err := ParseConnectRequest(request)
	assert.NoError(t, err)
	if !assert.NotNil(t, connectRequest) {
		return
	}

	displaySource := connectRequest.Body.DisplaySource
	displaySink := connectRequest.Body.DisplaySink

	if assert.NotNil(t, displaySource.MachineIdentifier.Id) {
		assert.Equal(t, "source-machine", displaySource.MachineIdentifier.Id.String())
	}

	if assert.NotNil(t, displaySource.PeripheralIdentifier.Id) {
		assert.Equal(t, "source-peripheral", displaySource.PeripheralIdentifier.Id.String())
	}

	if assert.NotNil(t, displaySink.MachineIdentifier.Name) {
		assert.Equal(t, "sink-machine", displaySink.MachineIdentifier.Name.String())
	}

	if assert.NotNil(t, displaySink.PeripheralIdentifier.Name) {
		assert.Equal(t, "sink-peripheral", displaySink.PeripheralIdentifier.Name.String())
	}
}

func TestParseConnectRequestBodyInvalidJSON(t *testing.T) {
	_, err := ParseConnectRequestBody(strings.NewReader("not-json"))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "decode json")
}

func TestParseConnectRequestBodyValidationFailure(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
            },
            "peripheralIdentifier": {
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "validate display source machine identifier")
	assert.ErrorContains(t, err, "either id or name must be provided")
}

func TestParseConnectRequestInvalidDomainValue(t *testing.T) {
	requestBody := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "INVALID_UPPER"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	request := transport.Request{Body: strings.NewReader(requestBody)}

	connectRequest, err := ParseConnectRequest(request)
	assert.Nil(t, connectRequest)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "parse body")
	assert.ErrorContains(t, err, "decode json")
}

func TestParseConnectRequestBodyDisplaySourcePeripheralValidationFailure(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "validate display source peripheral identifier")
	assert.ErrorContains(t, err, "either id or name must be provided")
}

func TestParseConnectRequestBodyDisplaySinkMachineValidationFailure(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "validate display sink machine identifier")
	assert.ErrorContains(t, err, "either id or name must be provided")
}

func TestParseConnectRequestBodyDisplaySinkPeripheralValidationFailure(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "validate display sink peripheral identifier")
	assert.ErrorContains(t, err, "either id or name must be provided")
}

func TestParseConnectRequestBodyAmbiguousMachineIdentifier(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine",
                "name": "source-machine-name"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "validate display source machine identifier")
	assert.ErrorContains(t, err, "id and name are mutually exclusive")
}

func TestParseConnectRequestBodyAmbiguousPeripheralIdentifier(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral",
                "name": "source-peripheral-name"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "validate display source peripheral identifier")
	assert.ErrorContains(t, err, "id and name are mutually exclusive")
}

func TestParseConnectRequestBodyInvalidPeripheralId(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "INVALID_UPPER"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "decode json")
}

func TestParseConnectRequestBodyInvalidDisplaySinkMachineId(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "INVALID_UPPER"
            },
            "peripheralIdentifier": {
                "id": "sink-peripheral"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "decode json")
}

func TestParseConnectRequestBodyInvalidDisplaySinkPeripheralId(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "id": "source-machine"
            },
            "peripheralIdentifier": {
                "id": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "id": "sink-machine"
            },
            "peripheralIdentifier": {
                "id": "INVALID_UPPER"
            }
        }
    }`

	_, err := ParseConnectRequestBody(strings.NewReader(body))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "decode json")
}

func TestParseConnectRequestWithNames(t *testing.T) {
	body := `{
        "displaySource": {
            "machineIdentifier": {
                "name": "source-machine"
            },
            "peripheralIdentifier": {
                "name": "source-peripheral"
            }
        },
        "displaySink": {
            "machineIdentifier": {
                "name": "sink-machine"
            },
            "peripheralIdentifier": {
                "name": "sink-peripheral"
            }
        }
    }`

	request := transport.Request{Body: strings.NewReader(body)}

	connectRequest, err := ParseConnectRequest(request)
	assert.NoError(t, err)
	if !assert.NotNil(t, connectRequest) {
		return
	}

	displaySource := connectRequest.Body.DisplaySource
	displaySink := connectRequest.Body.DisplaySink

	if assert.NotNil(t, displaySource.MachineIdentifier.Name) {
		assert.Equal(t, "source-machine", displaySource.MachineIdentifier.Name.String())
	}

	if assert.NotNil(t, displaySource.PeripheralIdentifier.Name) {
		assert.Equal(t, "source-peripheral", displaySource.PeripheralIdentifier.Name.String())
	}

	if assert.NotNil(t, displaySink.MachineIdentifier.Name) {
		assert.Equal(t, "sink-machine", displaySink.MachineIdentifier.Name.String())
	}

	if assert.NotNil(t, displaySink.PeripheralIdentifier.Name) {
		assert.Equal(t, "sink-peripheral", displaySink.PeripheralIdentifier.Name.String())
	}
}
