package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type NodeRegistrar struct {
	repository *NodeRepository

	eventListeners     map[uuid.UUID]chan nodeSDK.NodeRegistrarEvents
	eventListenersLock *sync.RWMutex
}

var _ nodeSDK.NodeRegistrar = (*NodeRegistrar)(nil)

func NewNodeRegistrar(repository *NodeRepository) *NodeRegistrar {
	return &NodeRegistrar{
		repository: repository,

		eventListeners:     make(map[uuid.UUID]chan nodeSDK.NodeRegistrarEvents),
		eventListenersLock: &sync.RWMutex{},
	}
}

func (registrar *NodeRegistrar) IsAttached(nodeId nodeSDK.NodeId) bool {
	registrar.repository.nodeLock.RLock()
	defer registrar.repository.nodeLock.RUnlock()

	_, found := registrar.repository.nodeIdIndex[nodeId]
	return found
}

func (registrar *NodeRegistrar) AttachNode(ctx context.Context, node nodeSDK.Node) error {
	nodeId, err := node.GetId(ctx)
	if err != nil {
		return fmt.Errorf("get id: %w", err)
	}

	nodeHostName, err := node.GetHostName(ctx)
	if err != nil {
		return fmt.Errorf("get host name: %w", err)
	}

	registrar.repository.nodeLock.Lock()
	defer registrar.repository.nodeLock.Unlock()

	if _, found := registrar.repository.nodeIdIndex[*nodeId]; found {
		return nodeSDK.ErrNodeIdAlreadyExists
	}

	registrar.repository.nodeIdIndex[*nodeId] = node

	defer registrar.emitEvent(nodeSDK.NodeAttachedEvent{
		Id:       *nodeId,
		HostName: *nodeHostName,
	})

	defer func() {
		var nodeIdList []nodeSDK.NodeId
		for nodeId := range registrar.repository.nodeIdIndex {
			nodeIdList = append(nodeIdList, nodeId)
		}
		registrar.emitEvent(nodeSDK.RepositorySnapshotEvent{
			Nodes: nodeIdList,
		})
	}()

	return nil
}

func (registrar *NodeRegistrar) DetachNode(nodeId nodeSDK.NodeId) error {
	registrar.repository.nodeLock.Lock()
	defer registrar.repository.nodeLock.Unlock()

	if _, idFound := registrar.repository.nodeIdIndex[nodeId]; !idFound {
		return nodeSDK.ErrNodeIdNotFound
	}

	delete(registrar.repository.nodeIdIndex, nodeId)

	defer registrar.emitEvent(nodeSDK.NodeDetachedEvent{
		Id: nodeId,
	})

	defer func() {
		var nodeIdList []nodeSDK.NodeId
		for nodeId := range registrar.repository.nodeIdIndex {
			nodeIdList = append(nodeIdList, nodeId)
		}
		registrar.emitEvent(nodeSDK.RepositorySnapshotEvent{
			Nodes: nodeIdList,
		})
	}()

	return nil
}

func (registrar *NodeRegistrar) WatchEvents(ctx context.Context) <-chan nodeSDK.NodeRegistrarEvents {
	registrar.eventListenersLock.Lock()
	defer registrar.eventListenersLock.Unlock()

	listenerId := uuid.New()
	listener := make(chan nodeSDK.NodeRegistrarEvents, 16)
	registrar.eventListeners[listenerId] = listener

	go func() {
		<-ctx.Done()

		registrar.eventListenersLock.Lock()
		delete(registrar.eventListeners, listenerId)
		registrar.eventListenersLock.Unlock()

		close(listener)
	}()

	return listener
}

func (registrar *NodeRegistrar) emitEvent(event nodeSDK.NodeRegistrarEvents) {
	registrar.eventListenersLock.RLock()
	defer registrar.eventListenersLock.RUnlock()

	for _, listener := range registrar.eventListeners {
		select {
		case listener <- event:
		default:
		}
	}
}
