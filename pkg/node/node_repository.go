package node

import (
	"context"
	"errors"
)

type NodeRegistrarEvents any

type NodeAttachedEvent struct {
	Id       NodeId `json:"id"`
	HostName string `json:"hostName"`
}

type NodeDetachedEvent struct {
	Id NodeId `json:"id"`
}

type RepositorySnapshotEvent struct {
	Nodes map[NodeId]string `json:"nodes"`
}

type NodeRegistrar interface {
	AttachNode(ctx context.Context, node Node) error
	DetachNode(nodeId NodeId) error
	IsAttached(nodeId NodeId) bool

	WatchEvents(ctx context.Context) <-chan NodeRegistrarEvents
}

type NodeRepository interface {
	GetNodeByHostName(ctx context.Context, hostName string) (Node, error)
	GetNodeById(ctx context.Context, nodeId NodeId) (Node, error)
	GetAllNodeIds(ctx context.Context) ([]NodeId, error)
}

var ErrNodeHostNameAlreadyExists = errors.New("node host name already exists")
var ErrNodeHostNameNotFound = errors.New("node host name not found")
var ErrNodeIdAlreadyExists = errors.New("node id already exists")
var ErrNodeIdNotFound = errors.New("node id not found")
