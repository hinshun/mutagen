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

	var (
		host string
		path string
	)
	if kind == Kind_Synchronization {
		raddr, last := multiaddr.SplitLast(maddr)
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
		maddrs := multiaddr.Split(maddr)

		i := len(maddrs) - 1
		for ; i >= 0; i-- {
			protos := maddrs[i].Protocols()
			if len(protos) > 0 && protos[0].Code == multiaddr.P_P2P {
				break
			}
		}

		if len(maddrs[:i+1]) > 0 {
			host = multiaddr.Join(maddrs[:i+1]...).String()
		}
		path = multiaddr.Join(maddrs[i+1:]...).String()
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
