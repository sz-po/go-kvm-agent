package node

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lensesio/tableprinter"
	nodeInternal "github.com/szymonpodeszwa/go-kvm-agent/internal/pkg/node"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"

	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type ListOutput struct {
	Id       nodeSDK.NodeId     `json:"id" header:"ID"`
	HostName string             `json:"hostName" header:"Hostname"`
	Roles    []nodeSDK.NodeRole `json:"roles" header:"Roles"`
	Platform string             `json:"platform" header:"Platform"`
}

type List struct {
}

func (cmd *List) Run(ctx context.Context, repository *nodeInternal.NodeRepository, transport apiSDK.Transport, logger *slog.Logger) error {
	var output []ListOutput

	nodes := map[nodeSDK.NodeId]nodeSDK.Node{}

	logger.Debug("Fetching all nodes id list.")

	nodeIds, err := repository.GetAllNodeIds(ctx)
	if err != nil {
		return fmt.Errorf("get all node ids: %w", err)
	}
	logger.Info("All nodes id list feched.", slog.Int("nodesCount", len(nodeIds)))

	for _, nodeId := range nodeIds {
		logger.Debug("Fetching node by id.", slog.String("nodeId", string(nodeId)))
		node, err := repository.GetNodeById(ctx, nodeId)
		if err != nil {
			return fmt.Errorf("get node: %w", err)
		}

		nodes[nodeId] = node
	}

	for nodeId, node := range nodes {
		var outputItem ListOutput
		outputItem.Id = nodeId

		nodeHostName, err := node.GetHostName(ctx)
		if err != nil {
			outputItem.HostName = fmt.Sprintf("error: %s", err)
		} else {
			outputItem.HostName = *nodeHostName
		}

		nodeRoles, err := node.GetRoles(ctx)
		if err != nil {
			outputItem.Roles = []nodeSDK.NodeRole{nodeSDK.NodeRole(fmt.Sprintf("error: %s", err))}
		} else {
			outputItem.Roles = nodeRoles
		}

		nodePlatform, err := node.GetPlatform(ctx)
		if err != nil {
			outputItem.Platform = fmt.Sprintf("error: %s", err)
		} else {
			outputItem.Platform = nodePlatform.String()
		}

		output = append(output, outputItem)
	}

	tableprinter.Print(os.Stdout, output)

	return nil
}
