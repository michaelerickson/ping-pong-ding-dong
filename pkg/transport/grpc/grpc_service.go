// Package grpc implements a ping-pong-ding-dong service transport layer
// that uses gRPC.
package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
func (s *PpddTransport) Init(rx chan core.Message, tx chan core.Request, cfg core.Config) error {
	if cfg.ApiVersion != apiVersion {
		return fmt.Errorf("API version error: transport support %s, wanted %s",
			apiVersion, cfg.ApiVersion)
	}
	s.rx = rx
	s.tx = tx
	return nil
}

// ListenAndServe is part of the core.PpddTransport interface and is called
// by the server.
func (s *PpddTransport) ListenAndServe(ctx context.Context, cfg core.Config) error {
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
				go send(req, cfg)
			}
		}
	}()

	listen, err := net.Listen("tcp", "localhost:"+cfg.ServicePort)
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
func (s *PpddTransport) Trigger(_ context.Context, msg *pb.Message) (*pb.Response, error) {
	log.Printf("rx: meta %s : msg %s", msg.GetMeta(), msg.GetMessage())

	// Unpack and receive the message
	go s.rxMessage(msg)

	// Craft the response
	sentAt := timestamppb.New(time.Now().UTC())
	var meta = pb.Meta{
		ApiVersion: apiVersion,
		SentAt:     sentAt,
	}
	var res = pb.Response{
		Meta: &meta,
	}
	return &res, nil
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

// send is responsible for sending a gRPC request to another service
func send(req core.Request, cfg core.Config) {
	// Convert internal representation to gRPC
	meta := pb.Meta{
		ApiVersion: apiVersion,
		SentAt:     timestamppb.New(time.Now().UTC()),
	}
	msg := pb.Message{
		Meta:    &meta,
		Message: pb.MessageType(req.Msg.Msg),
	}

	var endpoint string
	switch req.Endpoint {
	case core.Ping:
		endpoint = cfg.PingSvc
	case core.Pong:
		endpoint = cfg.PongSvc
	case core.Ding:
		endpoint = cfg.DingSvc
	case core.Dong:
		endpoint = cfg.DongSvc
	default:
		log.Printf("error attempting to send to endpoint %s", req.Endpoint)
		return
	}
	// Create a connection to the server
	credentials := insecure.NewCredentials()
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(credentials))
	if err != nil {
		log.Printf("error dialing: %s", err)
		return
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("error closing connection: %s", err)
		}
	}(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := pb.NewPpddServiceClient(conn)
	res, err := client.Trigger(ctx, &msg)
	if err != nil {
		log.Printf("error calling service: %s", err)
		return
	}
	log.Printf("got response %#v", res)
	return
}
