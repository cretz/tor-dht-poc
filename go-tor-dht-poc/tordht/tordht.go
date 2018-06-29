package tordht

import (
	"context"
	"crypto"
	"io"

	"github.com/cretz/bine/tor"
)

type Impl interface {
	ApplyDebugLogging()
	NewDiscoverer(context.Context, *tor.Tor) (Discoverer, error)
	NewProvider(context.Context, *tor.Tor) (Provider, error)
}

type Discoverer interface {
	io.Closer
	Discover(context.Context, []byte) <-chan crypto.PublicKey
}

type Provider interface {
	io.Closer
	Provide(context.Context, []byte, crypto.PublicKey) error
}
