package p2p

import (
	"context"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

const (
	RelayAddr = "/ip4/128.199.6.92/tcp/4001/p2p/12D3KooWGDynerXsf3KeXAQAp4RUpQKkusYYEjMq918czPxUXCRX"
)

func New(ctx context.Context) (host.Host, error) {
	relay, err := peer.AddrInfoFromString(RelayAddr)
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(
		libp2p.Defaults,
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT.
		libp2p.EnableAutoRelay(),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		libp2p.EnableNATService(),
		// EnableHolePunching enables NAT traversal by enabling NATT'd peers to both
		// initiate and respond to hole punching attempts to create direct /
		// NAT-traversed connections with other peers.
		libp2p.EnableHolePunching(),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return dht.New(ctx, h, dht.BootstrapPeers(*relay))
		}),
	)
	if err != nil {
		return host, err
	}

	return host, host.Connect(ctx, *relay)
}
