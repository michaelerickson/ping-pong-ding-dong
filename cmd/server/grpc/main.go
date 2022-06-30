package main

import (
	"context"
	"github.com/michaelerickson/ping-pong-ding-dong/pkg/core"
	"github.com/michaelerickson/ping-pong-ding-dong/pkg/transport/grpc"
	"log"
	"os"
	"os/signal"
)

//type foo struct{}
//
//func (*foo) ListenAndServ(ctx context.Context, rx chan core.Message, tx chan core.Request) error {
//	// Prime the pump with an unsolicited response from the pong service
//	go func() { rx <- core.NewMessage(apiVersion, core.Pong) }()
//
//	// Basically a loopback interface for a Ping service.
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				log.Print("context canceled, exiting send loop")
//				return
//			case req := <-tx:
//				log.Printf("sending: %s -> %s", req.Msg.Msg.String(), req.Endpoint.String())
//				switch req.Endpoint {
//				case core.Pong:
//					go func() { rx <- core.NewMessage(apiVersion, core.Pong) }()
//				case core.Ding:
//					go func() { rx <- core.NewMessage(apiVersion, core.Ding) }()
//					go func() { rx <- core.NewMessage(apiVersion, core.Dong) }()
//				}
//			}
//		}
//	}()
//
//	return nil
//}

func main() {
	const apiVersion = "v1"

	// Log in UTC time with microsecond resolution
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds)

	var cfg core.Config
	cfg.Init()
	if cfg.ApiVersion != apiVersion {
		log.Fatalf("api version mismatch: %s != %s", apiVersion, cfg.ApiVersion)
	}

	// Establish an outer context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on control-c or outer context being canceled.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			select {
			case <-c:
				log.Println("received control-c, shutting down")
				cancel()
				return
			}
		}
	}()

	// Configure a gRPC transport provider.
	var transport grpc.PpddTransport

	err := core.RunService(ctx, cfg, &transport)
	if err != nil {
		log.Fatalf("Error running service: %s", err)
	}
}
