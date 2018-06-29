package ipfs

import (
	"context"
	"crypto"

	"github.com/cretz/bine/tor"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
)

type provider struct {
	host    host.Host
	ds      datastore.Batching
	ipfsDHT *dht.IpfsDHT
}

func newProvider(ctx context.Context, bineTor *tor.Tor) (tordht.Provider, error) {
	prov := &provider{ds: sync.MutexWrap(datastore.NewMapDatastore())}
	// Close the prov on any error when creating
	var err error
	defer func() {
		if err != nil {
			prov.Close()
		}
	}()
	// Create host
	prov.host, err = libp2p.NewWithoutDefaults(ctx,
		libp2p.RandomIdentity,
		libp2p.Transport(NewTorTransport(bineTor, nil)),
		libp2p.DefaultPeerstore,
	)
	if err != nil {
		return nil, err
	}
	// Create dht
	if prov.ipfsDHT, err = dht.New(ctx, prov.host, opts.Datastore(prov.ds)); err != nil {
		return nil, err
	}
	return prov, nil
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
	panic("TODO")
}
