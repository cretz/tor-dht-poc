package tordht

import (
	"context"
	"fmt"
	"io"

	"github.com/cretz/bine/tor"
)

type Impl interface {
	ApplyDebugLogging()
	NewDHT(ctx context.Context, conf *DHTConf) (DHT, error)
}

type DHTConf struct {
	Tor            *tor.Tor
	BootstrapPeers []*PeerInfo
	ClientOnly     bool
	Verbose        bool
}

type PeerInfo struct {
	ID string
	// May be empty string if not listening
	OnionServiceID string
	// May be 0 if not listening
	OnionPort int
}

func (p *PeerInfo) String() string {
	return fmt.Sprintf("ID %v, addr %v:%v", p.ID, p.OnionServiceID, p.OnionPort)
}

type DHT interface {
	io.Closer

	PeerInfo() *PeerInfo
	Provide(ctx context.Context, id []byte) error
	FindProviders(ctx context.Context, id []byte, maxCount int) ([]*PeerInfo, error)
}
