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

// Every 4 pongs, the ping hits the ding :)
const pongInterval = 4

// pongCount allows us to send a ping to ding every fourth pong
var pongCount int

// RunService takes as PpddTransport and runs continuously until it is stopped.
// The behavior of the service is dependent on the mode it is started with.
func RunService(mode MessageType, transport PpddTransport) error {
	if mode == Undefined {
		return fmt.Errorf("refusing to run in %s mode", mode.String())
	}
	// Establish a context so we can shut things down cleanly
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the channels our transport plugin will send/receive on
	rx := make(chan Message)
	tx := make(chan Request)

	if err := transport.ListenAndServ(ctx, rx, tx); err != nil {
		return fmt.Errorf("transport ListenAndServe() failed: %s", err)
	}

	log.Printf("PPDD service started in mode: %s", mode.String())
	count := 0
	for loop := true; loop; {
		select {
		case msg := <-rx:
			count++
			log.Printf("Received %s", msg.Msg.String())
			var req Request
			req.Msg.Meta.ApiVersion = "v1"
			req.Msg.Meta.SentAt = time.Now().UTC()
			req.Msg.Msg = msg.Msg
			req.Endpoint = msg.Msg
			tx <- req
			if count == 3 {
				cancel()
				loop = false
			}
		}
	}

	log.Println("Service ending.")
	time.Sleep(10 * time.Second)
	return nil
}

// The handleMessage function dispatches a response based on the mode the
// service is running in and the message it received.
//func handleMessage(mode MessageType, msg MessageType, transport PpddTransport) error {
//	switch mode {
//	case Ping:
//		switch msg {
//		case Pong:
//			pongCount++
//			if 0 == (pongCount % pongInterval) {
//				go transport.Send(Ding, Ding)
//			}
//		}
//	}
//}
