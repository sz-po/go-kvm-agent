package go_kvm_agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
	"gopkg.in/yaml.v3"
)

func loadMachineConfigFromPath(configPath string) ([]machine.MachineConfig, error) {
	matches, err := filepath.Glob(configPath)
	if err != nil {
		return nil, fmt.Errorf("glob pattern error: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no configuration files found matching pattern: %s", configPath)
	}

	var configs []machine.MachineConfig
	for _, filePath := range matches {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", filePath, err)
		}

		var config machine.MachineConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("unmarshal YAML from %s: %w", filePath, err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}
