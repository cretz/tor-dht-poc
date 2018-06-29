package ipfs

import (
	"context"
	"fmt"
	"net"

	"github.com/cretz/bine/torutil"

	"github.com/whyrusleeping/mafmt"

	"github.com/cretz/bine/tor"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-transport"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"

	upgrader "github.com/libp2p/go-libp2p-transport-upgrader"
)

// impls libp2p's transport.Transport
type TorTransport struct {
	bineTor  *tor.Tor
	conf     *TorTransportConf
	upgrader *upgrader.Upgrader
}

type TorTransportConf struct {
	DialConf  *tor.DialConf
	OnlyOnion bool
}

var OnionMultiaddrFormat = mafmt.Base(ma.P_ONION)
var TorMultiaddrFormat = mafmt.Or(OnionMultiaddrFormat, mafmt.TCP)

var _ transport.Transport = &TorTransport{}

func NewTorTransport(bineTor *tor.Tor, conf *TorTransportConf) func(*upgrader.Upgrader) *TorTransport {
	return func(upgrader *upgrader.Upgrader) *TorTransport {
		bineTor.Debugf("Creating transport with upgrader: %v", upgrader)
		if conf == nil {
			conf = &TorTransportConf{}
		}
		return &TorTransport{bineTor, conf, upgrader}
	}
}

func (t *TorTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.Conn, error) {
	t.bineTor.Debugf("For peer ID %v, dialing %v", p, raddr)
	var network, addr string
	// Try net addr first
	if !t.conf.OnlyOnion {
		if netAddr, err := manet.ToNetAddr(raddr); err == nil {
			network, addr = netAddr.Network(), netAddr.String()
		}
	}
	// Now onion addr
	if network == "" {
		if onionAddress, err := raddr.ValueForProtocol(ma.P_ONION); err != nil {
			return nil, fmt.Errorf("Invalid onion or net address")
		} else {
			host, port, _ := torutil.PartitionString(onionAddress, ':')
			network = "tcp4"
			addr = host + ".onion"
			if port != "" {
				addr += ":" + port
			}
		}
	}
	// Now dial
	if dialer, err := t.bineTor.Dialer(ctx, t.conf.DialConf); err != nil {
		return nil, err
	} else if netConn, err := dialer.DialContext(ctx, network, addr); err != nil {
		return nil, err
	} else if manetConn, err := manet.WrapNetConn(netConn); err != nil {
		return nil, err
	} else {
		return t.upgrader.UpgradeOutbound(ctx, t, manetConn, p)
	}
}

func (t *TorTransport) CanDial(addr ma.Multiaddr) bool {
	t.bineTor.Debugf("Checking if can dial %v", addr)
	if t.conf.OnlyOnion {
		return OnionMultiaddrFormat.Matches(addr)
	}
	return TorMultiaddrFormat.Matches(addr)
}

func (t *TorTransport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	t.bineTor.Debugf("Called listen for %v", laddr)
	panic("TODO")
}

func (t *TorTransport) Protocols() []int { return []int{ma.P_TCP, ma.P_ONION} }
func (t *TorTransport) Proxy() bool      { return true }

type torConn struct {
	conn net.Conn
}
