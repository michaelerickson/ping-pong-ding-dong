// Package core implements the core logic of a ping-pong-ding-dong service.
// This includes defining the interfaces that things like transport and config
// providers need to implement.
package core

import (
	"context"
	"fmt"
	"log"
	"time"
)

// The version of the api this server supports.
const apiVersion = "v1"

// dingInterval defines the interval between when a ping server sends to ding.
const dingInterval = 4

// pongCount allows us to send a ping to ding every dingInterval.
var pongCount int

// RunService takes as PpddTransport and runs continuously until it is stopped.
// The behavior of the service is dependent on the mode it is started with.
func RunService(ctx context.Context, cfg Config, transport PpddTransport) error {
	if cfg.Mode == Undefined {
		return fmt.Errorf("refusing to run in %s mode", cfg.Mode)
	}

	if cfg.ApiVersion != apiVersion {
		return fmt.Errorf("unsupported API version, this server supports %s", apiVersion)
	}
	// Create the channels our transport plugin will send/receive on and
	// initialize the transport
	rx := make(chan Message)
	tx := make(chan Request)
	if err := transport.Init(rx, tx, cfg); err != nil {
		return err
	}

	// Monitor for received messages or global shutdown
	go func() {
		log.Printf("PPDD service started in mode: %s", cfg.Mode)
		for {
			select {
			case msg := <-rx:
				log.Printf("received %s", msg.Msg.String())
				handleMessage(cfg.Mode, msg, tx)
			case <-ctx.Done():
				log.Println("context canceled, ending receive loop")
				return
			}
		}
	}()

	// If we are in ping mode, bootstrap some calls to Pong
	go bootstrap(tx)

	if err := transport.ListenAndServe(ctx, cfg); err != nil {
		return fmt.Errorf("transport ListenAndServe() failed: %s", err)
	}
	log.Println("Service ending.")
	return nil
}

// The handleMessage function dispatches a response based on the mode the
// service is running in and the message it received.
func handleMessage(mode MessageType, msg Message, tx chan Request) {
	time.Sleep(3 * time.Second)
	switch mode {
	case Ping:
		switch msg.Msg {
		case Pong:
			// After dingInterval number of Pongs, we message the ding service
			pongCount++
			if 0 == (pongCount % dingInterval) {
				go func() { tx <- NewRequest(apiVersion, mode, Ding) }()
			}
			// Respond back to pong, ignore all others
			go func() { tx <- NewRequest(apiVersion, mode, Pong) }()
		}
	case Pong:
		// The Pong service always just responds back to the Ping service
		go func() { tx <- NewRequest(apiVersion, mode, Ping) }()
	case Ding:
		// The Ding service responds to Ping, and also messages Dong
		go func() { tx <- NewRequest(apiVersion, mode, Ping) }()
		go func() { tx <- NewRequest(apiVersion, mode, Dong) }()
	case Dong:
		// The Dong service just messages Ping
		go func() { tx <- NewRequest(apiVersion, mode, Ping) }()
	}
}

// bootstrap runs if the service starts in Mode ping. It runs until it sees
// a response from a pong service.
func bootstrap(tx chan Request) {
	count := 0
	for pongCount == 0 {
		time.Sleep(3 * time.Second)
		log.Printf("bootstrapping: %d", count)
		go func() { tx <- NewRequest(apiVersion, Ping, Pong) }()
		count++
	}
}
