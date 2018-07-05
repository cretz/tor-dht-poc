package ipfs

import (
	"encoding/base32"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

// Can be

var serviceIDEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)
var onionListenAddr ma.Multiaddr

const ONION_LISTEN_PROTO_CODE = 0x55

var onionListenProto = ma.Protocol{
	ONION_LISTEN_PROTO_CODE, 0, "onionListen", ma.CodeToVarint(ONION_LISTEN_PROTO_CODE), false, nil}

// var onionProto = ma.Protocol{ma.P_ONION, 96, "onion", ma.CodeToVarint(ma.P_ONION), false,
// 	ma.NewTranscoderFromFunctions(onionStringToBytes, onionBytesToString, nil)}
var onionProto = ma.Protocol{ma.P_ONION, ma.LengthPrefixedVarSize, "onion", ma.CodeToVarint(ma.P_ONION), false,
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
			ma.ProtocolsByName[onionProto.Name] = onionProto
			break
		}
	}
}

func onionStringToBytes(str string) ([]byte, error) {
	// Just convert the whole thing for now
	// log.Printf("Asked to convert onion string to bytes: %v", str)
	return []byte(str), nil
}

func onionBytesToString(byts []byte) (string, error) {
	// log.Printf("Asked to convert onion bytes back to string: %v", string(byts))
	return string(byts), nil
}
