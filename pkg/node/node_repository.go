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
	Nodes []NodeId `json:"nodes"`
}

type NodeRegistrar interface {
	AttachNode(ctx context.Context, node Node) error
	DetachNode(nodeId NodeId) error
	IsAttached(nodeId NodeId) bool

	WatchEvents(ctx context.Context) <-chan NodeRegistrarEvents
}

type NodeRepository interface {
	GetNodeById(ctx context.Context, nodeId NodeId) (Node, error)
	GetAllNodeIds(ctx context.Context) ([]NodeId, error)
}

var ErrNodeIdAlreadyExists = errors.New("node id already exists")
var ErrNodeIdNotFound = errors.New("node id not found")
