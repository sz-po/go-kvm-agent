package orbiqd_peripheral

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mitchellh/go-homedir"
	"github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/cli"
	driverSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/driver"
	peripheralSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/peripheral"
	"sigs.k8s.io/yaml"
)

type PeripheralConfig struct {
	DriverKind driverSDK.Kind     `json:"driverKind" required:"true"`
	Name       peripheralSDK.Name `json:"name" required:"true"`
	Config     any                `json:"config"`
}

func (peripheralConfig *PeripheralConfig) Decode(ctx *kong.DecodeContext) error {
	var rawConfigLocation string
	if err := ctx.Scan.PopValueInto("string", &rawConfigLocation); err != nil {
		return fmt.Errorf("read peripheral config flag: %w", err)
	}

	parsedURL, err := url.Parse(rawConfigLocation)
	if err != nil {
		return fmt.Errorf("parse peripheral config url %q: %w", rawConfigLocation, err)
	}

	if parsedURL.Scheme != "file" {
		return fmt.Errorf("peripheral config %q: unsupported scheme %q", rawConfigLocation, parsedURL.Scheme)
	}

	var filePath string
	switch {
	case parsedURL.Host != "":
		filePath = path.Join(parsedURL.Host, strings.TrimPrefix(parsedURL.Path, "/"))
	case parsedURL.Path != "":
		filePath = parsedURL.Path
	default:
		filePath = parsedURL.Opaque
	}
	if filePath == "" {
		return fmt.Errorf("peripheral config %q: missing file path", rawConfigLocation)
	}

	decodedPath, err := url.PathUnescape(filePath)
	if err != nil {
		return fmt.Errorf("decode peripheral config path %q: %w", filePath, err)
	}

	expandedPath, err := homedir.Expand(decodedPath)
	if err != nil {
		return fmt.Errorf("expand home directory in %q: %w", decodedPath, err)
	}

	normalizedPath := filepath.Clean(filepath.FromSlash(expandedPath))

	configBytes, err := os.ReadFile(normalizedPath)
	if err != nil {
		return fmt.Errorf("read peripheral config %s: %w", normalizedPath, err)
	}

	var loadedConfig PeripheralConfig
	if err := yaml.Unmarshal(configBytes, &loadedConfig); err != nil {
		return fmt.Errorf("unmarshal peripheral config %s: %w", normalizedPath, err)
	}

	*peripheralConfig = loadedConfig
	return nil
}

type Config struct {
	cli.LogConfigHelper
	cli.TransportConfigHelper

	Peripheral []PeripheralConfig `help:"Path to the peripheral config as url. Currently only file:// is supported."`
}
