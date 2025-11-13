package p2p

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	p2p "github.com/libp2p/go-libp2p"
	p2pevent "github.com/libp2p/go-libp2p/core/event"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	p2pnetwork "github.com/libp2p/go-libp2p/core/network"
	p2ppeer "github.com/libp2p/go-libp2p/core/peer"
	p2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
	p2pyamux "github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	p2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	p2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	p2ptcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	apiSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
	nodeAPI "github.com/szymonpodeszwa/go-kvm-agent/pkg/api/service/node"
	nodeSDK "github.com/szymonpodeszwa/go-kvm-agent/pkg/node"
)

type noopNodeRegistrar struct{}

func (r *noopNodeRegistrar) AttachNode(ctx context.Context, node nodeSDK.Node) error {
	return nil
}

func (r *noopNodeRegistrar) DetachNode(nodeId nodeSDK.NodeId) error {
	return nil
}

func (r *noopNodeRegistrar) IsAttached(nodeId nodeSDK.NodeId) bool {
	return false
}

func (r *noopNodeRegistrar) WatchEvents(ctx context.Context) <-chan nodeSDK.NodeRegistrarEvents {
	return nil
}

type transportOptions struct {
	logger        *slog.Logger
	bindAddress   string
	hostOptions   []p2p.Option
	nodeRegistrar nodeSDK.NodeRegistrar
	services      []nodeSDK.Service
	appId         apiSDK.AppId
}

func defaultTransportOptions() *transportOptions {
	return &transportOptions{
		logger:        slog.New(slog.DiscardHandler),
		bindAddress:   "0.0.0.0",
		hostOptions:   []p2p.Option{},
		nodeRegistrar: &noopNodeRegistrar{},
		services:      []nodeSDK.Service{},
		appId:         "orbiqd",
	}
}

type TransportOpt func(options *transportOptions)

func WithTransportLogger(logger *slog.Logger) TransportOpt {
	return func(options *transportOptions) {
		options.logger = logger
	}
}
func WithTransportIdentity(identity *Identity) TransportOpt {
	return func(options *transportOptions) {
		options.hostOptions = append(options.hostOptions, p2p.Identity(identity.GetPrivateKey()))
	}
}

func WithTransportNodeRegistrar(nodeRegistrar nodeSDK.NodeRegistrar) TransportOpt {
	return func(options *transportOptions) {
		options.nodeRegistrar = nodeRegistrar
	}
}

func WithTransportServices(services ...nodeSDK.Service) TransportOpt {
	return func(options *transportOptions) {
		for _, service := range services {
			options.services = append(options.services, service)
		}

	}
}

func WithTransportBindAddress(bindAddress string) TransportOpt {
	return func(options *transportOptions) {
		options.bindAddress = bindAddress
	}
}

type Transport struct {
	appId apiSDK.AppId
	host  p2phost.Host

	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	nodeRegistrar nodeSDK.NodeRegistrar

	logger *slog.Logger
}

func NewTransport(opts ...TransportOpt) (*Transport, error) {
	options := defaultTransportOptions()
	for _, opt := range opts {
		opt(options)
	}

	hostOptions := append(options.hostOptions,
		p2p.DisableRelay(),
		p2p.WithDialTimeout(10*time.Second),
		p2p.NoTransports,
		p2p.Transport(p2ptcp.NewTCPTransport),
		p2p.Transport(p2pquic.NewTransport),
		p2p.Muxer(p2pyamux.ID, p2pyamux.DefaultTransport),
		p2p.Security(p2ptls.ID, p2ptls.New),
		p2p.ListenAddrStrings([]string{
			fmt.Sprintf("/ip4/%s/tcp/0", options.bindAddress),
			fmt.Sprintf("/ip4/%s/udp/0/quic-v1", options.bindAddress),
		}...),
	)

	host, err := p2p.New(hostOptions...)
	if err != nil {
		return nil, err
	}

	logger := options.logger.With(slog.String("localNodeId", host.ID().String()))

	for _, address := range host.Addrs() {
		logger.Debug("Transport listening.", slog.String("localAddress", address.String()))
	}

	lifecycleCtx, lifecycleCancel := context.WithCancel(context.Background())

	transport := &Transport{
		appId: options.appId,
		host:  host,

		lifecycleCtx:    lifecycleCtx,
		lifecycleCancel: lifecycleCancel,

		nodeRegistrar: options.nodeRegistrar,

		logger: logger,
	}

	for _, service := range options.services {
		err := transport.registerService(service)
		if err != nil {
			lifecycleCancel()
			return nil, fmt.Errorf("register service: %w", err)
		}
	}

	host.Network().Notify(&p2pnetwork.NotifyBundle{
		ConnectedF:    transport.handlePeerConnected,
		DisconnectedF: transport.handlePeerDisconnected,
	})

	go transport.watchEvents(lifecycleCtx)

	return transport, nil
}

func (transport *Transport) GetLocalNodeId() nodeSDK.NodeId {
	return nodeSDK.NodeId(transport.host.ID().String())
}

func (transport *Transport) OpenServiceStream(ctx context.Context, serviceId nodeSDK.ServiceId, nodeId nodeSDK.NodeId) (io.ReadWriteCloser, error) {
	peerId, err := transport.getPeerId(nodeId)
	if err != nil {
		return nil, fmt.Errorf("peer id: %w", err)
	}

	protocolId := transport.getProtocolId(serviceId)

	return transport.host.NewStream(ctx, *peerId, protocolId)
}

func (transport *Transport) Terminate(ctx context.Context) error {
	transport.lifecycleCancel()

	return transport.host.Close()
}

func (transport *Transport) registerService(service nodeSDK.Service) error {
	serviceId := service.GetServiceId()

	protocolId := transport.getProtocolId(serviceId)

	transport.host.SetStreamHandler(protocolId, func(stream p2pnetwork.Stream) {
		ctx := context.WithValue(context.Background(), "transport", transport)
		ctx, ctxCancel := context.WithTimeout(ctx, time.Second*10)
		defer ctxCancel()

		service.Handle(ctx, stream)
	})

	transport.logger.Info("Registered service.", slog.String("protocolId", string(protocolId)))

	return nil
}

func (transport *Transport) getProtocolId(serviceId nodeSDK.ServiceId) p2pprotocol.ID {
	return p2pprotocol.ID(fmt.Sprintf("/%s/%s", transport.appId, serviceId))
}

func (transport *Transport) getPeerId(nodeId nodeSDK.NodeId) (*p2ppeer.ID, error) {
	peerId, err := p2ppeer.Decode(string(nodeId))
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return &peerId, nil
}

func (transport *Transport) handlePeerConnected(network p2pnetwork.Network, connection p2pnetwork.Conn) {
	peerNodeId := nodeSDK.NodeId(connection.RemotePeer().String())

	logger := transport.logger.With(
		slog.String("peerNodeId", string(peerNodeId)),
		slog.String("localAddress", connection.LocalMultiaddr().String()),
		slog.String("peerAddress", connection.RemoteMultiaddr().String()),
	)

	logger.Debug("Peer connected.")
}

func (transport *Transport) handlePeerDisconnected(network p2pnetwork.Network, connection p2pnetwork.Conn) {
	peerNodeId := nodeSDK.NodeId(connection.RemotePeer().String())

	logger := transport.logger.With(slog.String("peerNodeId", string(peerNodeId)))

	logger.Info("Peer node disconnected.")

	if !transport.nodeRegistrar.IsAttached(peerNodeId) {
		return
	}

	err := transport.nodeRegistrar.DetachNode(peerNodeId)
	if err != nil {
		logger.Warn("Detach node failed.", slog.String("error", err.Error()))
		return
	}

	logger.Info("Peer node detached.")

}

func (transport *Transport) handlePeerIdentificationCompleted(event p2pevent.EvtPeerIdentificationCompleted) {
	attachCtx, attachCtxCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer attachCtxCancel()

	peerNodeId := nodeSDK.NodeId(event.Peer.String())
	peerNode := nodeAPI.NewNodeClient(peerNodeId, transport)

	logger := transport.logger.With(
		slog.String("peerNodeId", string(peerNodeId)),
		slog.String("localAddress", event.Conn.LocalMultiaddr().String()),
		slog.String("peerAddress", event.Conn.RemoteMultiaddr().String()),
	)

	if transport.nodeRegistrar.IsAttached(peerNodeId) {
		return
	}

	err := transport.nodeRegistrar.AttachNode(attachCtx, peerNode)
	if err != nil {
		logger.Warn("Peer node attach failed.", slog.String("error", err.Error()))
		return
	}

	transport.host.ConnManager().Protect(event.Peer, "orbiqd")

	logger.Info("Peer node attached.")
}

func (transport *Transport) handlePeerIdentificationFailed(event p2pevent.EvtPeerIdentificationFailed) {
	peerNodeId := nodeSDK.NodeId(event.Peer.String())

	logger := transport.logger.With(
		slog.String("peerNodeId", string(peerNodeId)),
	)

	logger.Warn("Peer node identification failed.")
}

func (transport *Transport) watchEvents(ctx context.Context) {
	peerIdentificationCompletedSubscription, err := transport.host.EventBus().Subscribe(&p2pevent.EvtPeerIdentificationCompleted{})
	if err != nil {
		transport.logger.Warn("Error while subscribing on identification completed.", slog.String("error", err.Error()))
		return
	}
	defer peerIdentificationCompletedSubscription.Close()

	peerIdentificationFailedSubscription, err := transport.host.EventBus().Subscribe(&p2pevent.EvtPeerIdentificationFailed{})
	if err != nil {
		transport.logger.Warn("Error while subscribing on identification failed.", slog.String("error", err.Error()))
		return
	}
	defer peerIdentificationFailedSubscription.Close()

	done := ctx.Done()

	for {
		select {
		case <-done:
			return
		case event := <-peerIdentificationCompletedSubscription.Out():
			transport.handlePeerIdentificationCompleted(event.(p2pevent.EvtPeerIdentificationCompleted))
		case event := <-peerIdentificationFailedSubscription.Out():
			transport.handlePeerIdentificationFailed(event.(p2pevent.EvtPeerIdentificationFailed))
		}
	}
}
