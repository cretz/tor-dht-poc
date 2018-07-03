package ipfs

import (
	"context"
	"crypto"
	"fmt"

	"github.com/libp2p/go-libp2p-peerstore"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"

	"github.com/cretz/bine/tor"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	addr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	routed "github.com/libp2p/go-libp2p/p2p/host/routed"
)

type provider struct {
	bineTor *tor.Tor // not closed on Close
	host    host.Host
	ds      datastore.Batching
	ipfsDHT *dht.IpfsDHT
}

var bootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/104.236.176.52/tcp/4001/ipfs/QmSoLnSGccFuZQJzRadHn95W2CrSFmZuTdDWP8HXaHca9z",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/162.243.248.213/tcp/4001/ipfs/QmSoLueR4xBeUbY9WZ9xGUUxunbKWcrNFTDAadQJmocnWm",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
	"/ip4/178.62.61.185/tcp/4001/ipfs/QmSoLMeWqB7YGVLJN3pNLQpmmEk35v6wYtsMGLzSr5QBU3",
	"/ip4/104.236.151.122/tcp/4001/ipfs/QmSoLju6m7xTh3DuokvT3886QRYqxAzb1kShaanJgW36yx",
}

func newProvider(ctx context.Context, bineTor *tor.Tor) (tordht.Provider, error) {
	prov := &provider{bineTor: bineTor, ds: sync.MutexWrap(datastore.NewMapDatastore())}
	// Close the prov on any error when creating
	var err error
	defer func() {
		if err != nil {
			prov.Close()
		}
	}()
	// Create host
	prov.host, err = libp2p.New(ctx,
		// libp2p.RandomIdentity,
		libp2p.Transport(NewTorTransport(bineTor, nil)),
		// libp2p.DefaultPeerstore,
		// libp2p.NoSecurity,
	)
	if err != nil {
		return nil, err
	}

	bineTor.Debugf("Creating DHT with host: %v", prov.host)
	// Create dht
	if prov.ipfsDHT, err = dht.New(ctx, prov.host, opts.Datastore(prov.ds)); err != nil {
		return nil, err
	}

	// Make the host a routed one
	prov.host = routed.Wrap(prov.host, prov.ipfsDHT)

	// Start listening
	if err = prov.host.Network().Listen(onionListenAddr); err != nil {
		bineTor.Debugf("Failed listening: %v", err)
		return nil, err
	}

	// Boostrap it by waiting for the first 3 peers
	bineTor.Debugf("Bootstrapping host connections: %v", prov.host)
	peerConnCh := make(chan struct{}, len(bootstrapPeers))
	for _, bootstrapPeer := range bootstrapPeers {
		go func(bootstrapPeer string) {
			if connErr := prov.connectPeer(ctx, bootstrapPeer); connErr != nil {
				bineTor.Debugf("Peer connection to %v failed: %v", bootstrapPeer, connErr)
			} else {
				peerConnCh <- struct{}{}
			}
		}(bootstrapPeer)
	}
	for i := 0; i < 3; i++ {
		select {
		case <-peerConnCh:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Attempt to bootstrap DHT
	bineTor.Debugf("Bootstrapping DHT: %v", prov.ipfsDHT)
	if err = prov.ipfsDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}
	bineTor.Debugf("Bootstrap complete", prov.ipfsDHT)
	return prov, nil
}

func (p *provider) connectPeer(ctx context.Context, ipfsAddrStr string) error {
	if ipfsAddr, err := addr.ParseString(ipfsAddrStr); err != nil {
		return err
	} else if peer, err := peerstore.InfoFromP2pAddr(ipfsAddr.Multiaddr()); err != nil {
		return err
	} else {
		p.host.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
		return p.host.Connect(ctx, *peer)
	}
}

func (p *provider) Close() (err error) {
	if p.ipfsDHT != nil {
		err = p.ipfsDHT.Close()
	}
	if p.host != nil {
		if hostCloseErr := p.host.Close(); hostCloseErr != nil {
			// Just overwrite
			err = hostCloseErr
		}
	}
	return
}

func (p *provider) Provide(ctx context.Context, id []byte, pubKey crypto.PublicKey) error {
	hash, err := multihash.Sum(id, multihash.SHA3_256, -1)
	if err != nil {
		return fmt.Errorf("Failed hashing ID: %v", err)
	}
	c := cid.NewCidV1(0, hash)
	p.bineTor.Debugf("Providing CID: %v", c)
	return p.ipfsDHT.Provide(ctx, c, true)
}
