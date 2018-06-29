package tordht

import (
	"context"
	"crypto"
)

type Discoverer interface {
	Discover(context.Context, []byte) <-chan crypto.PublicKey
}

type Provider interface {
	Provide(context.Context, []byte, crypto.PublicKey) error
}
