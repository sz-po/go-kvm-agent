package go_kvm_agent

import (
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/http"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
)

type ControlApiConfig struct {
	Server http.ServerConfig `json:"server" validate:"omitempty,dive"`
}

type ApiConfig struct {
	ControlApi ControlApiConfig `json:"control" validate:"omitempty,dive"`
}

// Config contains global configuration for the agent runtime.
type Config struct {
	Machines []machine.MachineConfig `json:"machines" validate:"omitempty,dive"`
	Api      ApiConfig               `json:"api" validate:"omitempty,dive"`
}
