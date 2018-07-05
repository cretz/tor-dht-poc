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
	impl.ApplyDebugLogging()
	// We'll give it 2 minutes to startup everything
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelFn()
	// Fire up tor
	bineTor, err := tor.Start(ctx, &tor.StartConf{
		NoHush:      true,
		DebugWriter: os.Stdout,
		DataDir:     "data-dir-temp",
	})
	if err != nil {
		return fmt.Errorf("Failed starting tor: %v", err)
	}
	defer bineTor.Close()
	// Start DHT
	dht, err := impl.NewDHT(ctx, &tordht.DHTConf{
		Tor:     bineTor,
		Verbose: true,
	})
	if err != nil {
		return fmt.Errorf("Failed starting DHT: %v", err)
	}
	defer dht.Close()
	// Provide our key
	if err = dht.Provide(ctx, []byte(dataID)); err != nil {
		return fmt.Errorf("Failed providing: %v", err)
	}
	// Wait for key press...
	fmt.Printf("Press enter to quit...")
	_, err = fmt.Scanln()
	return err
}

func find(args []string) error {
	panic("TODO")
}
