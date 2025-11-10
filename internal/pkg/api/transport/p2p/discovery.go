package p2p

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type DiscovererOpt func(discoverer *Discoverer)
type Discoverer struct {
	transport *Transport

	mdns mdns.Service

	logger *slog.Logger
}

func WithDiscovererLogger(logger *slog.Logger) DiscovererOpt {
	return func(discoverer *Discoverer) {
		discoverer.logger = logger
	}
}

func WithMulticastDNSDiscovery() DiscovererOpt {
	return func(discoverer *Discoverer) {
		discoverer.mdns = mdns.NewMdnsService(discoverer.transport.host, string(discoverer.transport.appId), discoverer)
	}
}

func NewDiscoverer(transport *Transport, opts ...DiscovererOpt) (*Discoverer, error) {
	discoverer := &Discoverer{
		transport: transport,
	}

	for _, opt := range opts {
		opt(discoverer)
	}

	discoverer.logger = discoverer.logger.With(slog.String("localNodeId", discoverer.transport.host.ID().String()))

	if discoverer.mdns != nil {
		if err := discoverer.mdns.Start(); err != nil {
			return nil, fmt.Errorf("start mdns service: %w", err)
		}
	}

	return discoverer, nil
}

func (discoverer *Discoverer) Terminate(ctx context.Context) error {
	if discoverer.mdns != nil {
		if err := discoverer.mdns.Close(); err != nil {
			return fmt.Errorf("close mdns service: %w", err)
		}
	}

	return nil
}

func (discoverer *Discoverer) HandlePeerFound(info peer.AddrInfo) {
	localNodeId := discoverer.transport.host.ID()
	peerNodeId := info.ID

	if localNodeId == peerNodeId {
		return
	}

	logger := discoverer.logger.With(
		slog.String("peerNodeId", peerNodeId.String()),
	)

	if localNodeId < peerNodeId {
		logger.Info("Discovered new node. Waiting for connection.")
		return
	}

	if err := discoverer.transport.host.Connect(context.Background(), info); err != nil {
		logger.Warn("Discovered new node, but connection fails.", slog.String("error", err.Error()))
	}

	logger.Info("Discovered new node and connected to it.")
}
