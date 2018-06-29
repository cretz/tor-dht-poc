package ipfs

import (
	"context"

	"github.com/cretz/bine/tor"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	logging "github.com/ipfs/go-log"
)

type impl struct{}

func (impl) ApplyDebugLogging() { logging.SetDebugLogging() }
func (impl) NewDiscoverer(ctx context.Context, bineTor *tor.Tor) (tordht.Discoverer, error) {
	panic("TODO")
}
func (impl) NewProvider(ctx context.Context, bineTor *tor.Tor) (tordht.Provider, error) {
	return newProvider(ctx, bineTor)
}

var Impl tordht.Impl = impl{}
