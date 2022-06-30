package main

import (
	"context"
	"log"
	"time"

	"github.com/michaelerickson/ping-pong-ding-dong/pkg/core"
)

type foo struct{}

func (*foo) ListenAndServ(ctx context.Context, rx chan core.Message, tx chan core.Request) error {

	go func() {
		// Send some messages
		for i := 0; true; i++ {
			var msg core.Message
			msg.Meta.ApiVersion = "v1"
			msg.Meta.SentAt = time.Now().UTC()
			switch {
			case 0 == (i % 4):
				msg.Msg = core.Dong
			case 0 == (i % 2):
				msg.Msg = core.Pong
			case 0 == (i % 3):
				msg.Msg = core.Ding
			default:
				msg.Msg = core.Ping
			}
			// As long as the context hasn't been canceled, we should be able
			// to send fake messages on the rx channel.
			if done := ctx.Err(); done != nil {
				log.Println("context canceled, exiting receive loop")
				return
			}
			log.Printf("receiving %#v", msg)
			rx <- msg // Send a message to the service
			time.Sleep(3 * time.Second)
		}
		return
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Print("context canceled, exiting send loop")
				return
			case msg := <-tx:
				log.Printf("sending: %#v", msg)
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
