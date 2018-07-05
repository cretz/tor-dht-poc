package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cretz/bine/tor"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht/ipfs"
)

// Change to true to see lots of logs
const debug = false
const participatingPeerCount = 3
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
	if len(args) > 0 {
		return fmt.Errorf("No args accepted for 'provide' currently")
	}
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Fire up tor
	bineTor, err := startTor(ctx, "data-dir-temp-provide")
	if err != nil {
		return fmt.Errorf("Failed starting tor: %v", err)
	}
	defer bineTor.Close()

	// Make multiple DHTs, passing the known set to the other ones for connecting
	log.Printf("Creating %v peers", participatingPeerCount)
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
		log.Printf("Created peer #%v: %v\n", i+1, dht.PeerInfo())
	}

	// Have a couple provide our key
	log.Printf("Providing key on the first one (%v)\n", dhts[0].PeerInfo())
	if err = dhts[0].Provide(ctx, []byte(dataID)); err != nil {
		return fmt.Errorf("Failed providing on first: %v", err)
	}
	log.Printf("Providing key on the last one (%v)\n", dhts[len(dhts)-1].PeerInfo())
	if err = dhts[len(dhts)-1].Provide(ctx, []byte(dataID)); err != nil {
		return fmt.Errorf("Failed providing on last: %v", err)
	}

	// Wait for key press...
	log.Printf("Press enter to quit...\n")
	_, err = fmt.Scanln()
	return err
}

func find(args []string) error {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Get all the peers from the args
	var err error
	dhtConf := &tordht.DHTConf{
		ClientOnly:     true,
		Verbose:        debug,
		BootstrapPeers: make([]*tordht.PeerInfo, len(args)),
	}
	for i := 0; i < len(args); i++ {
		if dhtConf.BootstrapPeers[i], err = tordht.NewPeerInfo(args[i]); err != nil {
			return fmt.Errorf("Failed parsing arg #%v: %v", i+1, err)
		}
	}

	// Fire up tor
	if dhtConf.Tor, err = startTor(ctx, "data-dir-temp-find"); err != nil {
		return fmt.Errorf("Failed starting tor: %v", err)
	}
	defer dhtConf.Tor.Close()

	// Make a client-only DHT
	log.Printf("Creating DHT and connecting to peers\n")
	dht, err := impl.NewDHT(ctx, dhtConf)
	if err != nil {
		return fmt.Errorf("Failed creating DHT: %v", err)
	}

	// Now find who is providing the id
	providers, err := dht.FindProviders(ctx, []byte(dataID), 2)
	if err != nil {
		return fmt.Errorf("Failed finding providers: %v", err)
	}
	for _, provider := range providers {
		log.Printf("Found data ID on %v\n", provider)
	}
	return nil
}

func startTor(ctx context.Context, dataDir string) (*tor.Tor, error) {
	startConf := &tor.StartConf{DataDir: dataDir}
	if debug {
		impl.ApplyDebugLogging()
		startConf.NoHush = true
		startConf.DebugWriter = os.Stderr
	}
	return tor.Start(ctx, startConf)
}
