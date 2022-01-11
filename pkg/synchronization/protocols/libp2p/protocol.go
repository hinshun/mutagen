package libp2p

import (
	"context"
	"errors"
	"io"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/mutagen-io/mutagen/pkg/logging"
	"github.com/mutagen-io/mutagen/pkg/synchronization"
	"github.com/mutagen-io/mutagen/pkg/synchronization/endpoint/remote"
	urlpkg "github.com/mutagen-io/mutagen/pkg/url"
)

var (
	SynchronizationProtocol = protocol.ID("/mutagen/synchronization")
)

// protocolHandler implements the synchronization.ProtocolHandler interface for
// connecting to remote endpoints over libp2p.
type protocolHandler struct{}

// dialResult provides asynchronous agent dialing results.
type dialResult struct {
	// stream is the stream returned by agent dialing.
	stream io.ReadWriteCloser
	// error is the error returned by agent dialing.
	error error
}

// Connect connects to an libp2p endpoint.
func (h *protocolHandler) Connect(
	ctx context.Context,
	logger *logging.Logger,
	url *urlpkg.URL,
	prompter string,
	session string,
	version synchronization.Version,
	configuration *synchronization.Configuration,
	alpha bool,
) (synchronization.Endpoint, error) {
	// Verify that the URL is of the correct kind and protocol.
	if url.Kind != urlpkg.Kind_Synchronization {
		panic("non-synchronization URL dispatched to synchronization protocol handler")
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

	p2pAddr, err := multiaddr.NewMultiaddr(url.Host)
	if err != nil {
		return nil, err
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(p2pAddr)
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(libp2p.Defaults)
	if err != nil {
		return nil, err
	}
	host.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)

	stream, err := host.NewStream(ctx, addrInfo.ID, SynchronizationProtocol)
	if err != nil {
		return nil, err
	}

	// Create the endpoint client.
	return remote.NewEndpoint(stream, url.Path, session, version, configuration, alpha)
}

func init() {
	// Register the libp2p protocol handler with the synchronization package.
	synchronization.ProtocolHandlers[urlpkg.Protocol_Libp2p] = &protocolHandler{}
}
