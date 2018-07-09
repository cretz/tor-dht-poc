package tordht

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/cretz/bine/torutil"

	"github.com/cretz/bine/tor"
)

type Impl interface {
	ApplyDebugLogging()
	RawStringDataID(id []byte) (string, error)
	NewDHT(ctx context.Context, conf *DHTConf) (DHT, error)
}

type DHTConf struct {
	Tor            *tor.Tor
	BootstrapPeers []*PeerInfo
	ClientOnly     bool
	Verbose        bool
}

type DHT interface {
	io.Closer

	PeerInfo() *PeerInfo
	Provide(ctx context.Context, id []byte) error
	FindProviders(ctx context.Context, id []byte, maxCount int) ([]*PeerInfo, error)
}

type PeerInfo struct {
	ID string
	// May be empty string if not listening
	OnionServiceID string
	// Invalid value if OnionServiceID is empty
	OnionPort int
}

func (p *PeerInfo) String() string {
	return fmt.Sprintf("%v:%v/%v", p.OnionServiceID, p.OnionPort, p.ID)
}

func NewPeerInfo(str string) (*PeerInfo, error) {
	if onion, id, ok := torutil.PartitionString(str, '/'); !ok {
		return nil, fmt.Errorf("Missing ID portion")
	} else if onionID, portStr, ok := torutil.PartitionString(onion, ':'); !ok {
		return nil, fmt.Errorf("Missing onion port")
	} else {
		ret := &PeerInfo{ID: id, OnionServiceID: onionID}
		ret.OnionPort, _ = strconv.Atoi(portStr)
		return ret, nil
	}
}
