package main

import (
	"fmt"
	"os"
	"path/filepath"

	go_kvm_agent "github.com/szymonpodeszwa/go-kvm-agent/internal/app/go-kvm-agent"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/machine"
	"sigs.k8s.io/yaml"
)

func loadConfigFromPath(configPath string) (*go_kvm_agent.Config, error) {
	var config go_kvm_agent.Config

	configBuffer, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	err = yaml.Unmarshal(configBuffer, &config)
	if err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	return &config, nil
}

func loadMachineConfigFromPath(configPath string) ([]machine.MachineConfig, error) {
	matches, err := filepath.Glob(configPath)
	if err != nil {
		return nil, fmt.Errorf("glob pattern error: %w", err)
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
