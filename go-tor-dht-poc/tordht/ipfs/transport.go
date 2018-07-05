package ipfs

import (
	"context"
	"fmt"
	"net"
	"time"

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
		} else {
			t.bineTor.Debugf("Invalid net address trying onion address: %v", err)
		}
	}
	// Now onion addr
	if network == "" {
		if onionAddress, err := raddr.ValueForProtocol(ma.P_ONION); err != nil {
			return nil, fmt.Errorf("Invalid onion or net address: %v", err)
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
		t.bineTor.Debugf("Failed creating Tor dialer: %v", err)
		return nil, err
	} else if netConn, err := dialer.DialContext(ctx, network, addr); err != nil {
		t.bineTor.Debugf("Failed dialing network '%v' addr '%': %v", network, addr, err)
		return nil, err
	} else if manetConn, err := manet.WrapNetConn(netConn); err != nil {
		t.bineTor.Debugf("Failed wrapping the net connection: %v", err)
		return nil, err
	} else if conn, err := t.upgrader.UpgradeOutbound(ctx, t, manetConn, p); err != nil {
		t.bineTor.Debugf("Failed upgrading connection: %v", err)
		return nil, err
	} else {
		return conn, nil
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
	// TODO: support a bunch of config options on this if we want
	t.bineTor.Debugf("Called listen for %v", laddr)
	if laddr.String() != "/onionListen" {
		return nil, fmt.Errorf("Must be '/onionListen' for now")
	}
	// Listen with version 3, wait 1 min for bootstrap
	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFn()
	onion, err := t.bineTor.Listen(ctx, &tor.ListenConf{Version3: true})
	if err != nil {
		t.bineTor.Debugf("Failed creating onion service: %v", err)
		return nil, err
	}

	t.bineTor.Debugf("Listening on onion: %v", onion.String())
	// Close it if there is another error in here
	defer func() {
		if err != nil {
			t.bineTor.Debugf("Failed listen after onion creation: %v", err)
			onion.Close()
		}
	}()

	// Return a listener
	manetListen := &manetListener{transport: t, onion: onion}
	manetListen.multiaddr, err = ma.NewMultiaddr(fmt.Sprintf("/onion/%v:%v", onion.ID, onion.RemotePorts[0]))
	if err != nil {
		return nil, fmt.Errorf("Failed converting onion address: %v", err)
	}

	// Encapsulate the underlying tcp
	t.bineTor.Debugf("Completed creating IPFS listener from onion, addr: %v", manetListen.multiaddr)
	return manetListen.Upgrade(t.upgrader), nil
}

func (t *TorTransport) Protocols() []int { return []int{ma.P_TCP, ma.P_ONION, ONION_LISTEN_PROTO_CODE} }
func (t *TorTransport) Proxy() bool      { return true }

type manetListener struct {
	transport *TorTransport
	onion     *tor.OnionService
	multiaddr ma.Multiaddr
}

func (m *manetListener) Accept() (manet.Conn, error) {
	if c, err := m.onion.Accept(); err != nil {
		return nil, err
	} else {
		ret := &manetConn{Conn: c, localMultiaddr: m.multiaddr}
		if ret.remoteMultiaddr, err = manet.FromNetAddr(c.RemoteAddr()); err != nil {
			return nil, err
		}
		return ret, nil
	}
}
func (m *manetListener) Close() error            { return m.onion.Close() }
func (m *manetListener) Addr() net.Addr          { return m.onion }
func (m *manetListener) Multiaddr() ma.Multiaddr { return m.multiaddr }
func (m *manetListener) Upgrade(u *upgrader.Upgrader) transport.Listener {
	return u.UpgradeListener(m.transport, m)
}

type manetConn struct {
	net.Conn
	localMultiaddr  ma.Multiaddr
	remoteMultiaddr ma.Multiaddr
}

func (m *manetConn) LocalMultiaddr() ma.Multiaddr  { return m.localMultiaddr }
func (m *manetConn) RemoteMultiaddr() ma.Multiaddr { return m.remoteMultiaddr }
