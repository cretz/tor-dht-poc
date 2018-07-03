package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht"
	"github.com/cretz/tor-dht-poc/go-tor-dht-poc/tordht/ipfs"
)

func main() {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancelFn()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

const dataID = "tor-dht-poc-test"

func run(ctx context.Context) error {
	impl.ApplyDebugLogging()
	log.Printf("Starting Tor")
	bineTor, err := tor.Start(ctx, &tor.StartConf{
		NoHush:      true,
		DebugWriter: os.Stdout,
		DataDir:     "data-dir-temp",
	})
	if err != nil {
		return fmt.Errorf("Failed starting Tor: %v", err)
	}
	defer bineTor.Close()

	log.Printf("Creating provider")
	prov, err := impl.NewProvider(ctx, bineTor)
	if err != nil {
		return err
	}
	defer prov.Close()

	log.Printf("Creating an ed25519 key pair for provider")
	keyPair, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("Failed generating key: %v", err)
	}

	log.Printf("Marking us as a provider for '%v'", dataID)
	if err = prov.Provide(ctx, []byte(dataID), keyPair.PublicKey()); err != nil {
		return fmt.Errorf("Failed providing: %v", err)
	}

	log.Printf("Waiting 5 seconds")
	time.Sleep(5 * time.Second)
	return fmt.Errorf("TODO: the rest")
}

var impl tordht.Impl = ipfs.Impl
