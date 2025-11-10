package node

import (
	"context"
	"os"
	"time"

	host_info "github.com/shirou/gopsutil/v4/host"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type Node struct {
	id    nodeSDK.NodeId
	roles []nodeSDK.NodeRole
}

var _ nodeSDK.Node = (*Node)(nil)

type NodeOpt func(*Node)

func WithNodeRole(role nodeSDK.NodeRole) NodeOpt {
	return func(node *Node) {
		node.roles = append(node.roles, role)
	}
}

func NewNode(id nodeSDK.NodeId, opts ...NodeOpt) *Node {
	node := &Node{
		id:    id,
		roles: []nodeSDK.NodeRole{},
	}

	for _, opt := range opts {
		opt(node)
	}

	return node
}

func (node *Node) GetId(ctx context.Context) (*nodeSDK.NodeId, error) {
	return &node.id, nil
}

func (node *Node) GetRoles(ctx context.Context) ([]nodeSDK.NodeRole, error) {
	return node.roles, nil
}

func (node *Node) GetHostName(ctx context.Context) (*string, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &hostName, nil
}

func (node *Node) GetUptime(ctx context.Context) (*time.Duration, error) {
	uptimeRaw, err := host_info.UptimeWithContext(ctx)
	if err != nil {
		return nil, err
	}

	uptime := time.Second * time.Duration(uptimeRaw)
	return &uptime, nil
}

func (node *Node) GetPlatform(ctx context.Context) (*nodeSDK.NodePlatform, error) {
	hostInfo, err := host_info.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return &nodeSDK.NodePlatform{
		OperatingSystem: nodeSDK.NodeOperatingSystem(hostInfo.OS),
		Architecture:    nodeSDK.NodeArchitecture(hostInfo.KernelArch),
	}, nil
}
