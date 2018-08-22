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

func init() {
	// Add the listen protocol
	if err := ma.AddProtocol(onionListenProto); err != nil {
		panic(fmt.Errorf("Failed adding onionListen protocol: %v", err))
	} else if onionListenAddr, err = ma.NewMultiaddr("/onionListen"); err != nil {
		panic(fmt.Errorf("Failed creating onionListen addr: %v", err))
	}
	// Change existing onion protocol to support v3 and be more lenient when transcoding
	ma.TranscoderOnion = ma.NewTranscoderFromFunctions(onionStringToBytes, onionBytesToString, nil)
	for _, p := range ma.Protocols {
		if p.Code == ma.P_ONION {
			p.Size = ma.LengthPrefixedVarSize
			p.Transcoder = ma.TranscoderOnion
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
