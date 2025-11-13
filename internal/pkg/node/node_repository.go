package node

import (
	"context"
	"sync"

	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
	"golang.org/x/exp/maps"
)

type NodeRepository struct {
	nodeIdIndex map[nodeSDK.NodeId]nodeSDK.Node
	nodeLock    *sync.RWMutex
}

var _ nodeSDK.NodeRepository = (*NodeRepository)(nil)

func NewNodeRepository() *NodeRepository {
	return &NodeRepository{
		nodeIdIndex: make(map[nodeSDK.NodeId]nodeSDK.Node),
		nodeLock:    &sync.RWMutex{},
	}
}

func (repository *NodeRepository) GetNodeById(ctx context.Context, nodeId nodeSDK.NodeId) (nodeSDK.Node, error) {
	repository.nodeLock.RLock()
	defer repository.nodeLock.RUnlock()

	node, found := repository.nodeIdIndex[nodeId]
	if !found {
		return nil, nodeSDK.ErrNodeIdNotFound
	}

	return node, nil
}

func (repository *NodeRepository) GetAllNodeIds(ctx context.Context) ([]nodeSDK.NodeId, error) {
	repository.nodeLock.RLock()
	defer repository.nodeLock.RUnlock()

	return maps.Keys(repository.nodeIdIndex), nil
}
