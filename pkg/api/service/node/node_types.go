package node

import (
	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

const NodeServiceId = nodeSDK.ServiceId("node")

const (
	NodeGetIdMethod       nodeSDK.MethodName = "get-id"
	NodeGetHostNameMethod nodeSDK.MethodName = "get-host-name"
	NodeGetUptimeMethod   nodeSDK.MethodName = "get-uptime"
	NodeGetPlatformMethod nodeSDK.MethodName = "get-platform"
	NodeGetRolesMethod    nodeSDK.MethodName = "get-role"
)

type NodeGetHostNameRequest struct {
}

type NodeGetHostNameResponse struct {
	HostName string `json:"hostName"`
}

type NodeGetIdRequest struct {
}

type NodeGetIdResponse struct {
	Id nodeSDK.NodeId `json:"id"`
}

type NodeGetUptimeRequest struct {
}

type NodeGetUptimeResponse struct {
	Uptime api.Duration `json:"uptime"`
}

type NodeGetPlatformRequest struct {
}

type NodeGetPlatformResponse struct {
	Platform nodeSDK.NodePlatform `json:"platform"`
}

type NodeGetRoleRequest struct {
}

type NodeGetRoleResponse struct {
	Roles []nodeSDK.NodeRole `json:"roles"`
}
