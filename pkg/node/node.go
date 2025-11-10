package node

import (
	"context"
	"time"
)

type NodeRole string

func (role NodeRole) String() string {
	return string(role)
}

const (
	Peripheral NodeRole = "peripheral"
	CLI                 = "cli"
)

type NodeId string

type NodeOperatingSystem string

const (
	Windows NodeOperatingSystem = "windows"
	Linux                       = "linux"
	MacOS                       = "macos"
)

type NodeArchitecture string

const (
	AMD64 NodeArchitecture = "amd64"
	ARM64                  = "arm64"
)

type NodePlatform struct {
	Architecture    NodeArchitecture    `json:"architecture"`
	OperatingSystem NodeOperatingSystem `json:"operatingSystem"`
}

func (platform NodePlatform) String() string {
	return string(platform.OperatingSystem) + "-" + string(platform.Architecture)
}

type Node interface {
	GetId(ctx context.Context) (*NodeId, error)
	GetRoles(ctx context.Context) ([]NodeRole, error)
	GetHostName(ctx context.Context) (hostName *string, err error)
	GetUptime(ctx context.Context) (uptime *time.Duration, err error)
	GetPlatform(ctx context.Context) (platform *NodePlatform, err error)
}
