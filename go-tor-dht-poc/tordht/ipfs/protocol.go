package ipfs

import (
	"encoding/base32"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

var serviceIDEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)
var onionListenAddr ma.Multiaddr

const ONION_LISTEN_PROTO_CODE = 0x55

var onionListenProto = ma.Protocol{
	"onionListen", ONION_LISTEN_PROTO_CODE, ma.CodeToVarint(ONION_LISTEN_PROTO_CODE), 0, false, nil}

var onionProto = ma.Protocol{"onion", ma.P_ONION, ma.CodeToVarint(ma.P_ONION), ma.LengthPrefixedVarSize, false,
	ma.NewTranscoderFromFunctions(onionStringToBytes, onionBytesToString, nil)}

func init() {
	var err error
	if err = ma.AddProtocol(onionListenProto); err != nil {
		panic(fmt.Errorf("Failed adding onionListen protocol: %v", err))
	} else if onionListenAddr, err = ma.NewMultiaddr("/onionListen"); err != nil {
		panic(fmt.Errorf("Failed creating onionListen addr: %v", err))
	}
	// Replace the existing onion protocol with one that is more lenient
	ma.TranscoderOnion = onionProto.Transcoder
	for i, p := range ma.Protocols {
		if p.Code == ma.P_ONION {
			ma.Protocols[i] = onionProto
			//ma.ProtocolsByName[onionProto.Name] = onionProto
			break
		}
	}
}

func onionStringToBytes(str string) ([]byte, error) {
	// Just convert the whole thing for now
	return []byte(str), nil
}

func onionBytesToString(byts []byte) (string, error) {
	// Just convert the whole thing for now
	return string(byts), nil
}
