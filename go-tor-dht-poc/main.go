package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht/ipfs"
)

// Change to true to see lots of logs
const debug = true
const participatingPeerCount = 5
const dataID = "tor-dht-poc-test"

var impl tordht.Impl = ipfs.Impl

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("Expected 'provide' or 'find' command")
	} else if cmd, subArgs := os.Args[1], os.Args[2:]; cmd == "provide" {
		return provide(subArgs)
	} else if cmd == "find" {
		return find(subArgs)
	} else {
		return fmt.Errorf("Invalid command '%v'", cmd)
	}
}

func provide(args []string) error {
	// We'll give it 2 minutes to startup everything
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelFn()
	// Fire up tor
	startConf := &tor.StartConf{DataDir: "data-dir-temp"}
	if debug {
		impl.ApplyDebugLogging()
		startConf.NoHush = true
		startConf.DebugWriter = os.Stderr
	}
	bineTor, err := tor.Start(ctx, startConf)
	if err != nil {
		return fmt.Errorf("Failed starting tor: %v", err)
	}
	defer bineTor.Close()

	// Make multiple DHTs, passing the known set to the other ones for connecting
	fmt.Printf("Creating %v peers\n", participatingPeerCount)
	dhts := make([]tordht.DHT, participatingPeerCount)
	prevPeers := []*tordht.PeerInfo{}
	for i := 0; i < len(dhts); i++ {
		// Start DHT
		conf := &tordht.DHTConf{
			Tor:            bineTor,
			Verbose:        debug,
			BootstrapPeers: make([]*tordht.PeerInfo, len(prevPeers)),
		}
		copy(conf.BootstrapPeers, prevPeers)
		dht, err := impl.NewDHT(ctx, conf)
		if err != nil {
			return fmt.Errorf("Failed starting DHT: %v", err)
		}
		defer dht.Close()
		dhts[i] = dht
		prevPeers = append(prevPeers, dht.PeerInfo())
		fmt.Printf("Created peer #%v: %v\n", i+1, dht.PeerInfo())
	}

	// Have a couple provide our key
	fmt.Printf("Providing key on the first one (%v)\n", dhts[0].PeerInfo())
	if err = dhts[0].Provide(ctx, []byte(dataID)); err != nil {
		return fmt.Errorf("Failed providing on first: %v", err)
	}
	fmt.Printf("Providing key on the last one (%v)\n", dhts[len(dhts)-1].PeerInfo())
	if err = dhts[len(dhts)-1].Provide(ctx, []byte(dataID)); err != nil {
		return fmt.Errorf("Failed providing on last: %v", err)
	}

	// Wait for key press...
	fmt.Printf("Press enter to quit...\n")
	_, err = fmt.Scanln()
	return err
}

func find(args []string) error {
	panic("TODO")
}
