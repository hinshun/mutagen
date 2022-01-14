package libp2p

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/mutagen-io/mutagen/pkg/forwarding"
	"github.com/mutagen-io/mutagen/pkg/forwarding/endpoint/local"
	"github.com/mutagen-io/mutagen/pkg/forwarding/endpoint/remote"
	"github.com/mutagen-io/mutagen/pkg/logging"
	"github.com/mutagen-io/mutagen/pkg/p2p"
	urlpkg "github.com/mutagen-io/mutagen/pkg/url"
)

var (
	ForwardingProtocol = protocol.ID("/mutagen/forwarding")
)

// protocolHandler implements the forwarding.ProtocolHandler interface for
// connecting to remote endpoints over libp2p. It uses the agent infrastructure
// over an libp2p transport.
type protocolHandler struct{}

// dialResult provides asynchronous agent dialing results.
type dialResult struct {
	// stream is the stream returned by agent dialing.
	stream io.ReadWriteCloser
	// error is the error returned by agent dialing.
	error error
}

// Connect connects to an Libp2p endpoint.
func (p *protocolHandler) Connect(
	ctx context.Context,
	logger *logging.Logger,
	url *urlpkg.URL,
	prompter string,
	session string,
	version forwarding.Version,
	configuration *forwarding.Configuration,
	source bool,
) (forwarding.Endpoint, error) {
	// Verify that the URL is of the correct kind and protocol.
	if url.Kind != urlpkg.Kind_Forwarding {
		panic("non-forwarding URL dispatched to forwarding protocol handler")
	} else if url.Protocol != urlpkg.Protocol_Libp2p {
		panic("non-libp2p URL dispatched to libp2p protocol handler")
	}

	// Ensure that no environment variables or parameters are specified. These
	// are neither expected nor supported for Libp2p URLs.
	if len(url.Environment) > 0 {
		return nil, errors.New("Libp2p URL contains environment variables")
	} else if len(url.Parameters) > 0 {
		return nil, errors.New("Libp2p URL contains internal parameters")
	}

	maddr, err := multiaddr.NewMultiaddr(url.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url path as multiaddr: %w", err)
	}

	proto, address, err := manet.DialArgs(maddr)
	if err != nil {
		return nil, fmt.Errorf("unable to extract network and address from url path: %w", err)
	}

	if len(url.Host) == 0 {
		if source {
			return local.NewListenerEndpoint(version, configuration, proto, address, true)
		}
		return local.NewDialerEndpoint(version, configuration, proto, address)
	}

	addrInfo, err := peer.AddrInfoFromString(url.Host)
	if err != nil {
		return nil, fmt.Errorf("unable to get addr info from url host: %w", err)
	}

	host, err := p2p.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create p2p peer: %w", err)
	}

	err = host.Connect(ctx, *addrInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to url: %w", err)
	}

	ctx = network.WithUseTransient(ctx, "hole-punch")
	stream, err := host.NewStream(ctx, addrInfo.ID, ForwardingProtocol)
	if err != nil {
		return nil, fmt.Errorf("unable to create libp2p %s stream: %w", ForwardingProtocol, err)
	}

	// Create the endpoint.
	return remote.NewEndpoint(stream, version, configuration, proto, address, source)
}

func init() {
	// Register the Libp2p protocol handler with the forwarding package.
	forwarding.ProtocolHandlers[urlpkg.Protocol_Libp2p] = &protocolHandler{}
}
