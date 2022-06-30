// Package grpc implements a ping-pong-ding-dong service transport layer
// that uses gRPC.
package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"

	pb "github.com/michaelerickson/ping-pong-ding-dong/internal/api/proto/v1"
	"github.com/michaelerickson/ping-pong-ding-dong/pkg/core"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const apiVersion = "v1"

// PpddTransport fulfills the PpddServiceServer interface defined by the
// protocol buffer.
type PpddTransport struct {
	pb.UnimplementedPpddServiceServer
	tx chan core.Request
	rx chan core.Message
}

// Init configures the transport provider
func (s *PpddTransport) Init(rx chan core.Message, tx chan core.Request, version string) error {
	if version != apiVersion {
		return fmt.Errorf("API version error: transport support %s, wanted %s",
			apiVersion, version)
	}
	s.rx = rx
	s.tx = tx
	return nil
}

// ListenAndServe is part of the core.PpddTransport interface and is called
// by the server.
func (s *PpddTransport) ListenAndServe(ctx context.Context) error {
	// Launch a go routine to handle outbound requests from our service to
	// other services.
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Print("context canceled, exiting send loop")
				return
			case req := <-s.tx:
				log.Printf("sending %#v", req)
			}
		}
	}()

	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	pb.RegisterPpddServiceServer(server, s)

	// Monitor for a global shutdown to gracefully stop the server.
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Print("context canceled, shutting down transport server")
				server.GracefulStop()
				return
			}
		}
	}()

	log.Printf("grpc transport listening at %v", listen.Addr())
	return server.Serve(listen)
}

// Trigger is the gRPC endpoint where the service receives messages.
func (s *PpddTransport) Trigger(_ context.Context, msg *pb.Message) (res *pb.Response, err error) {
	log.Printf("rx: meta %s : msg %s", msg.GetMeta(), msg.GetMessage())

	// Unpack and receive the message
	go s.rxMessage(msg)

	// Craft the response
	sentAt := timestamppb.New(time.Now().UTC())
	var meta = pb.Meta{
		ApiVersion: apiVersion,
		SentAt:     sentAt,
	}
	res.Meta = &meta
	return res, nil
}

// rxMessage converts the protocol buffer into the core data structure and
// puts it on the rx channel.
func (s *PpddTransport) rxMessage(msg *pb.Message) {
	var coreMsg core.Message
	coreMsg.Meta.ApiVersion = msg.Meta.GetApiVersion()
	coreMsg.Meta.SentAt = msg.Meta.GetSentAt().AsTime()
	coreMsg.Msg = core.MessageType(msg.GetMessage())
	s.rx <- coreMsg
}
