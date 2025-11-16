package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/deb2000-sudo/trackshift/internal/relay"
)

func main() {
	listenPort := flag.Int("listen-port", 9001, "UDP port to listen on")
	forwardAddr := flag.String("forward-address", "127.0.0.1:9090", "destination UDP address")
	relayID := flag.String("relay-id", "relay-1", "unique relay identifier")
	orchestratorURL := flag.String("orchestrator-url", "", "orchestrator URL (optional)")
	flag.Parse()

	listen := ":" + strconv.Itoa(*listenPort)

	fwd, err := relay.NewForwarder(listen, *forwardAddr, *relayID, *orchestratorURL)
	if err != nil {
		log.Fatalf("create forwarder: %v", err)
	}

	log.Printf("Relay %s listening on %s, forwarding to %s", *relayID, listen, *forwardAddr)
	fwd.Start()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	log.Println("Shutting down relay...")
	if err := fwd.Close(); err != nil {
		log.Printf("error closing forwarder: %v", err)
	}
}

