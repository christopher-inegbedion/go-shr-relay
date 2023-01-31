package main

import (
	"fmt"
	"log"
	"time"

	"github.com/TwiN/go-color"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

var km KeyManager

func main() {
	startRelay()
}

const (
	// This is the maximum amount of data that can be transferred at any one time through the relay.
	// Any more than this and the relay will reset the connection.
	MAX_DATA_TRANSFER_ALLOWED = 1 * 1024 * 1024 * 1024 // 1GB

	// This is how long a relayed connection can last before it is reset.
	NODE_CONNECTION_DURATION  = time.Hour * 24         // 1 day
)

func startRelay() {
	km.InitKeyManager()

	relayResource := relay.Resources{
		Limit: &relay.RelayLimit{
			Duration: NODE_CONNECTION_DURATION,
			Data:     MAX_DATA_TRANSFER_ALLOWED,
		},

		MaxReservations: 128,
		MaxCircuits:     16,
		BufferSize:      4096,

		MaxReservationsPerPeer: 100,
		MaxReservationsPerIP:   100,
		MaxReservationsPerASN:  32,

		ReservationTTL: time.Hour * 24 * 7 * 4 * 12,
	}

	// Create a host to act as a middleman to relay messages on our behalf
	relay1, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2468"),
		libp2p.Identity(km.PrivKey),
	)
	if err != nil {
		log.Printf("Failed to create relay1: %v", err)
		return
	}

	// Configure the host to offer the ciruit relay service.
	// Any host that is directly dialable in the network (or on the internet)
	// can offer a circuit relay service, this isn't just the job of
	// "dedicated" relay services.
	// In circuit relay v2 (which we're using here!) it is rate limited so that
	// any node can offer this service safely
	_, err = relay.New(relay1, relay.WithResources(relayResource))

	if err != nil {
		log.Printf("Failed to instantiate the relay: %v", err)
		return
	}

	relay1info := peer.AddrInfo{
		ID:    relay1.ID(),
		Addrs: relay1.Addrs(),
	}

	addrs, _ := peer.AddrInfoToP2pAddrs(&relay1info)
	fmt.Printf("Relay address: %v", color.InGreen(addrs[0].String()))

	select {}
}
