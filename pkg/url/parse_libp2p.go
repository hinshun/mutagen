package url

import (
	"errors"

	multiaddr "github.com/multiformats/go-multiaddr"
)

// isLibp2pURL checks whether or not a URL is a libp2p URL.
func isLibp2pURL(raw string) bool {
	_, err := multiaddr.NewMultiaddr(raw)
	return err == nil
}

// parseLibp2p parses a Libp2p URL.
func parseLibp2p(raw string, kind Kind) (*URL, error) {
	maddr, err := multiaddr.NewMultiaddr(raw)
	if err != nil {
		return nil, err
	}
	path := maddr.String()

	var host string
	raddr, last := multiaddr.SplitLast(maddr)
	if kind == Kind_Synchronization {
		if !last.Protocol().Path {
			return nil, errors.New("last protocol must be a path protocol")
		}

		path, err = last.ValueForProtocol(multiaddr.P_UNIX)
		if err != nil {
			return nil, err
		}

		_, last = multiaddr.SplitLast(raddr)
		if last.Protocol().Code != multiaddr.P_P2P {
			return nil, errors.New("protocol must include a libp2p peer")
		}
		host = raddr.String()
	} else if kind == Kind_Forwarding {
		if last.Protocol().Path {
			return nil, errors.New("last protocol must not be a path protocol")
		}
	} else {
		panic("unhandled URL kind")
	}

	return &URL{
		Kind:     kind,
		Protocol: Protocol_Libp2p,
		Host:     host,
		Path:     path,
	}, nil
}
