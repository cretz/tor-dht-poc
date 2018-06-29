package ipfs

import "github.com/libp2p/go-libp2p-kad-dht"

type discoverer struct {
	ipfsDHT *dht.IpfsDHT
}
