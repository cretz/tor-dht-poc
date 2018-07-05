package ipfs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peerstore "github.com/libp2p/go-libp2p-peerstore"

	ma "github.com/multiformats/go-multiaddr"

	addr "github.com/ipfs/go-ipfs-addr"
)

type torDHT struct {
	debug    bool
	tor      *tor.Tor
	ipfsHost host.Host
	ipfsDHT  *dht.IpfsDHT
	peerInfo *tordht.PeerInfo
}

func (t *torDHT) Close() (err error) {
	if t.ipfsDHT != nil {
		err = t.ipfsDHT.Close()
	}
	if t.ipfsHost != nil {
		if hostCloseErr := t.ipfsHost.Close(); hostCloseErr != nil {
			// Just overwrite
			err = hostCloseErr
		}
	}
	return
}

func (t *torDHT) PeerInfo() *tordht.PeerInfo { return t.peerInfo }

func (t *torDHT) Provide(ctx context.Context, id []byte) error {
	panic("TODO")
}

func (t *torDHT) FindProviders(ctx context.Context, id []byte, maxCount int) ([]*tordht.PeerInfo, error) {
	panic("TODO")
}

func (t *torDHT) debugf(format string, args ...interface{}) {
	if t.debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func (t *torDHT) applyPeerInfo() error {
	t.peerInfo = &tordht.PeerInfo{ID: t.ipfsHost.ID().Pretty()}
	if listenAddrs := t.ipfsHost.Network().ListenAddresses(); len(listenAddrs) > 1 {
		return fmt.Errorf("Expected at most 1 listen onion address, got %v", listenAddrs)
	} else if len(listenAddrs) == 0 {
		// no addr
		return nil
	} else if onionAddrStr, err := listenAddrs[0].ValueForProtocol(ma.P_ONION); err != nil {
		return fmt.Errorf("Failed getting onion info from %v: %v", listenAddrs[0], err)
	} else if id, portStr, ok := torutil.PartitionString(onionAddrStr, ':'); !ok {
		return fmt.Errorf("Missing port on %v", onionAddrStr)
	} else if port, portErr := strconv.Atoi(portStr); portErr != nil {
		return fmt.Errorf("Invalid port '%v': %v", portStr, portErr)
	} else {
		t.peerInfo.OnionServiceID = id
		t.peerInfo.OnionPort = port
		return nil
	}
}

func (t *torDHT) connectPeers(ctx context.Context, peers []*tordht.PeerInfo, minRequired int) error {
	if len(peers) < minRequired {
		minRequired = len(peers)
	}
	t.debugf("Starting %v peer connections, waiting for at least %v", len(peers), minRequired)
	// Connect to a bunch asynchronously
	peerConnCh := make(chan error, len(peers))
	for _, peer := range peers {
		go func(peer *tordht.PeerInfo) {
			if connErr := t.connectPeer(ctx, peer); connErr != nil {
				peerConnCh <- fmt.Errorf("Peer connection to %v failed: %v", peer, connErr)
			} else {
				peerConnCh <- nil
			}
		}(peer)
	}
	peerErrs := []error{}
	peersConnected := 0
	// Until there is an error or we have enough
	for {
		select {
		case peerErr := <-peerConnCh:
			if peerErr == nil {
				peersConnected++
				if peersConnected >= minRequired {
					return nil
				}
			} else {
				peerErrs = append(peerErrs, peerErr)
				if len(peerErrs) > len(peers)-minRequired {
					return fmt.Errorf("Many failures, unable to get enough peers: %v", peerErrs)
				}
			}
		case <-ctx.Done():
			return fmt.Errorf("Context errored with '%v', peer errors: %v", ctx.Err(), peerErrs)
		}
	}
}

func (t *torDHT) connectPeer(ctx context.Context, peerInfo *tordht.PeerInfo) error {
	ipfsAddrStr := fmt.Sprintf("/onion/%v:%v/ipfs/%v", peerInfo.OnionServiceID, peerInfo.OnionPort, peerInfo.ID)
	if ipfsAddr, err := addr.ParseString(ipfsAddrStr); err != nil {
		return err
	} else if peer, err := peerstore.InfoFromP2pAddr(ipfsAddr.Multiaddr()); err != nil {
		return err
	} else {
		t.ipfsHost.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
		return t.ipfsHost.Connect(ctx, *peer)
	}
}
