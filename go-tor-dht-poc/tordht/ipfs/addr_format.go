package ipfs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cretz/bine/torutil"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
)

type addrFormat interface {
	onionInfo(addr ma.Multiaddr) (id string, port int, err error)
	onionAddr(id string, port int) string
}

// var defaultAddrFormat addrFormat = addrFormatProtocol{}
var defaultAddrFormat addrFormat = addrFormatDns{}

// In the form /onion/<onion-id>:<port>
type addrFormatProtocol struct{}

func (addrFormatProtocol) onionInfo(addr ma.Multiaddr) (string, int, error) {
	if onionAddrStr, err := addr.ValueForProtocol(ma.P_ONION); err != nil {
		return "", -1, fmt.Errorf("Failed getting onion info from %v: %v", addr, err)
	} else if id, portStr, ok := torutil.PartitionString(onionAddrStr, ':'); !ok {
		return "", -1, fmt.Errorf("Missing port on %v", onionAddrStr)
	} else if port, portErr := strconv.Atoi(portStr); portErr != nil {
		return "", -1, fmt.Errorf("Invalid port '%v': %v", portStr, portErr)
	} else {
		return id, port, nil
	}
}

func (addrFormatProtocol) onionAddr(id string, port int) string {
	return fmt.Sprintf("/onion/%v:%v", id, port)
}

// In the form /dns4/<onion-id>.onion/tcp/<port>
type addrFormatDns struct{}

func (addrFormatDns) onionInfo(addr ma.Multiaddr) (string, int, error) {
	if addrPieces := ma.Split(addr); len(addrPieces) < 2 {
		return "", -1, fmt.Errorf("Invalid pieces: %v", addrPieces)
	} else if onionAddrStr, err := addrPieces[0].ValueForProtocol(madns.Dns4Protocol.Code); err != nil {
		return "", -1, fmt.Errorf("Can't get onion part of %v: %v", addr, err)
	} else if !strings.HasSuffix(onionAddrStr, ".onion") {
		return "", -1, fmt.Errorf("Invalid onion addr: %v", onionAddrStr)
	} else if portStr, err := addrPieces[1].ValueForProtocol(ma.P_TCP); err != nil {
		return "", -1, fmt.Errorf("Can't get port part of %v: %v", addr, err)
	} else if port, portErr := strconv.Atoi(portStr); portErr != nil {
		return "", -1, fmt.Errorf("Invalid port '%v': %v", portStr, portErr)
	} else {
		return onionAddrStr[:len(onionAddrStr)-6], port, nil
	}
}

func (addrFormatDns) onionAddr(id string, port int) string {
	return fmt.Sprintf("/dns4/%v.onion/tcp/%v", id, port)
}
