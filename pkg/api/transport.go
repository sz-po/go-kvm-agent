package api

import (
	"context"
	"io"

	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type AppId string

type Transport interface {
	GetLocalNodeId() nodeSDK.NodeId
	OpenServiceStream(ctx context.Context, serviceId nodeSDK.ServiceId, remoteNodeId nodeSDK.NodeId) (io.ReadWriteCloser, error)
	Terminate(ctx context.Context) error
}
