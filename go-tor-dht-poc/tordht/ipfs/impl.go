package ipfs

import (
	"context"
	"fmt"

	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	cid "github.com/ipfs/go-cid"
	datastore "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	log "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	routed "github.com/libp2p/go-libp2p/p2p/host/routed"
	multihash "github.com/multiformats/go-multihash"
	mplex "github.com/whyrusleeping/go-smux-multiplex"
)

type impl struct{}

var ipfsImpl = impl{}
var Impl tordht.Impl = ipfsImpl

const minPeersRequired = 2

func (impl) ApplyDebugLogging() {
	log.SetDebugLogging()
	// log.SetAllLoggers(logging.INFO)
}

func (impl) RawStringDataID(id []byte) (string, error) {
	if raw, err := ipfsImpl.hashedCID(id); err != nil {
		return "", err
	} else {
		return raw.String(), nil
	}
}

func (impl) NewDHT(ctx context.Context, conf *tordht.DHTConf) (tordht.DHT, error) {
	t := &torDHT{debug: conf.Verbose, tor: conf.Tor}
	// Close the dht on any error when creating, so make sure err is populated before returning
	var err error
	defer func() {
		if err != nil {
			t.Close()
		}
	}()

	// Create the host with only the tor transport
	t.debugf("Creating host")
	transportConf := &TorTransportConf{
		WebSocket: true,
	}
	hostOpts := []libp2p.Option{
		// libp2p.NoSecurity,
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		libp2p.Transport(NewTorTransport(conf.Tor, transportConf)),
	}
	if !conf.ClientOnly {
		// Add an address to listen to
		hostOpts = append(hostOpts, libp2p.ListenAddrs(onionListenAddr))
	}
	if t.ipfsHost, err = libp2p.New(ctx, hostOpts...); err != nil {
		return nil, fmt.Errorf("Failed creating host: %v", err)
	}
	// Get the peer info out since we need it
	if !conf.ClientOnly {
		if err = t.applyPeerInfo(); err != nil {
			return nil, fmt.Errorf("Failed obtaining listen addr: %v", err)
		}
		t.debugf("Listening on %v", t.peerInfo)
	}

	// Create the DHT with a normal datastore
	t.debugf("Creating DHT on host")
	ds := sync.MutexWrap(datastore.NewMapDatastore())
	if t.ipfsDHT, err = dht.New(ctx, t.ipfsHost, opts.Datastore(ds)); err != nil {
		return nil, fmt.Errorf("Failed creating DHT: %v", err)
	}

	// Create a host that is routed with the DHT
	t.debugf("Creating routed host")
	t.ipfsHost = routed.Wrap(t.ipfsHost, t.ipfsDHT)

	// Connect to at least X (or total count if fewer than X)
	if len(conf.BootstrapPeers) > 0 {
		if err = t.connectPeers(ctx, conf.BootstrapPeers, minPeersRequired); err != nil {
			return nil, fmt.Errorf("Failed connecting to peers: %v", err)
		}
	}

	// Bootstrap the DHT
	t.debugf("Bootstrapping DHT")
	if err = t.ipfsDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("Failed boostrapping DHT: %v", err)
	}
	return t, nil
}

func (impl) hashedCID(v []byte) (*cid.Cid, error) {
	if hash, err := multihash.Sum(v, multihash.SHA3_256, -1); err != nil {
		return nil, fmt.Errorf("Failed hashing ID: %v", err)
	} else {
		return cid.NewCidV1(0, hash), nil
	}
}
