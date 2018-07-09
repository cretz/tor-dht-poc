package main

import (
	"context"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/cretz/bine/tor"

	"github.com/cretz/bine/torutil/ed25519"
)

const debug = false

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Create/load key
	key, err := createOrLoadKey()
	if err != nil {
		return fmt.Errorf("Failed creating or loading key: %v", err)
	}

	// Start Tor
	startConf := &tor.StartConf{DataDir: "data-dir-temp-web"}
	if debug {
		startConf.NoHush = true
		startConf.DebugWriter = os.Stderr
	}
	bineTor, err := tor.Start(ctx, startConf)
	if err != nil {
		return fmt.Errorf("Failed starting tor: %v", err)
	}
	defer bineTor.Close()

	// Create onion service
	onion, err := bineTor.Listen(ctx, &tor.ListenConf{
		Key:         key,
		RemotePorts: []int{80},
	})
	if err != nil {
		return fmt.Errorf("Unable to create onion service: %v", err)
	}
	defer onion.Close()

	fmt.Printf("Open Tor browser and navigate to http://%v.onion\n", onion.ID)
	fmt.Printf("Press enter to exit")
	// Serve the current folder from HTTP
	errCh := make(chan error, 1)
	go func() { errCh <- http.Serve(onion, http.FileServer(http.Dir("public"))) }()
	// End when enter is pressed
	go func() {
		fmt.Scanln()
		errCh <- nil
	}()
	if err = <-errCh; err != nil {
		return fmt.Errorf("Failed serving: %v", err)
	}
	return nil
}

func createOrLoadKey() (ed25519.KeyPair, error) {
	if byts, err := ioutil.ReadFile("onion.pem"); err == nil {
		if block, _ := pem.Decode(byts); block == nil || block.Type != "PRIVATE KEY" {
			return nil, fmt.Errorf("Invalid block")
		} else {
			return ed25519.PrivateKey(block.Bytes).KeyPair(), nil
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("Error loading file: %v", err)
	} else if key, err := ed25519.GenerateKey(nil); err != nil {
		return nil, fmt.Errorf("Failed generating key: %v", err)
	} else {
		block := &pem.Block{Type: "PRIVATE KEY", Bytes: key.PrivateKey()}
		return key, ioutil.WriteFile("onion.pem", pem.EncodeToMemory(block), 0600)
	}
}
