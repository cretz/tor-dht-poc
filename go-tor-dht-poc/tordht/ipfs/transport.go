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
			onion.Close()
		}
	}()

	// Return a listener
	ret := &torListener{addr: onion}
	var manetListen manet.Listener
	if ret.multiaddr, err = ma.NewMultiaddr(fmt.Sprintf("/onion/%v:%v", onion.ID, onion.RemotePorts[0])); err != nil {
		t.bineTor.Debugf("Failed creating onion service: %v", err)
		return nil, err
	} else if manetListen, err = manet.WrapNetListener(onion); err != nil {
		t.bineTor.Debugf("Failed wrapping onion listener: %v", err)
		return nil, err
	}
	ret.underlying = t.upgrader.UpgradeListener(t, manetListen)
	t.bineTor.Debugf("Completed creating IPFS listener from onion")
	return ret, nil
}

func (t *TorTransport) Protocols() []int { return []int{ma.P_TCP, ma.P_ONION, ONION_LISTEN_PROTO_CODE} }
func (t *TorTransport) Proxy() bool      { return true }

type torConn struct {
	conn net.Conn
}

type torListener struct {
	addr       net.Addr
	multiaddr  ma.Multiaddr
	underlying transport.Listener
}

func (t *torListener) Accept() (transport.Conn, error) { return t.underlying.Accept() }
func (t *torListener) Close() error                    { return t.underlying.Close() }
func (t *torListener) Addr() net.Addr                  { return t.addr }
func (t *torListener) Multiaddr() ma.Multiaddr         { return t.multiaddr }
