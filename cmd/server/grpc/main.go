package main

import (
	"context"
	"github.com/michaelerickson/ping-pong-ding-dong/pkg/core"
	"log"
)

const apiVersion = "v1"

type foo struct{}

func (*foo) ListenAndServ(ctx context.Context, rx chan core.Message, tx chan core.Request) error {
	// Prime the pump with an unsolicited response from the pong service
	go func() { rx <- core.NewMessage(apiVersion, core.Pong) }()

	// Basically a loopback interface for a Ping service.
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Print("context canceled, exiting send loop")
				return
			case req := <-tx:
				log.Printf("sending: %s -> %s", req.Msg.Msg.String(), req.Endpoint.String())
				switch req.Endpoint {
				case core.Pong:
					go func() { rx <- core.NewMessage(apiVersion, core.Pong) }()
				case core.Ding:
					go func() { rx <- core.NewMessage(apiVersion, core.Ding) }()
					go func() { rx <- core.NewMessage(apiVersion, core.Dong) }()
				}
			}
		}
	}()

	return nil
}

func main() {
	// Log in UTC time with microsecond resolution
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds)
	var myFoo foo
	mode := core.Ping

	err := core.RunService(mode, &myFoo)
	if err != nil {
		log.Fatalf("Error running service: %s", err)
	}
}
