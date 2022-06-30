package core

import (
	"context"
	"time"
)

// A MessageType enumerates the set of messages passed between services.
type MessageType int

// Messages to a ping-pong-ding-dong service will be one of these tokens.
const (
	Undefined MessageType = iota
	Ping
	Pong
	Ding
	Dong
)

// The String method converts the MessageType integer to its string equivalent.
func (m MessageType) String() string {
	switch m {
	case Undefined:
		return "Undefined"
	case Ping:
		return "Ping"
	case Pong:
		return "Pong"
	case Ding:
		return "Ding"
	case Dong:
		return "Dong"
	}
	return "unknown"
}

// Meta information about a message sent to or from the service.
type Meta struct {
	ApiVersion string
	SentAt     time.Time
}

// Message is the format of data sent to our service.
type Message struct {
	Meta Meta
	Msg  MessageType
}

// Response is the format of a response from our service.
type Response struct {
	Meta Meta
}

// A Request is passed to the transport implementation to be sent to another
// service. The idea is that the services are discoverable by the same name
// as their MessageType. So we can use that as the value of the Endpoint we
// want to send the Msg too.
type Request struct {
	Msg      Message
	Endpoint MessageType
}

// PpddTransport defines the interface a ping-pong-ding-dong service expects
// of a transport plugin.
type PpddTransport interface {
	// Init stores the channels the transport provider uses to send and receive
	// messages on, as well as the context it should monitor for graceful exit.
	// The transport should receive Message types and place them on the rx
	// channel for the service to process. The service will place Request types
	// on the tx channel for the transport to send.
	Init(rx chan Message, tx chan Request, version string) error

	// ListenAndServe processes received Message and Request types. It launches
	// its own go routines and will monitor the context passed to Init to shut
	// down.
	ListenAndServe(ctx context.Context) error
}

// NewMessage is useful for creating mocks and testing
func NewMessage(apiVersion string, msg MessageType) Message {
	req := Message{
		Meta: struct {
			ApiVersion string
			SentAt     time.Time
		}{ApiVersion: apiVersion, SentAt: time.Now().UTC()},
		Msg: msg,
	}
	return req
}

// NewRequest creates a new request to another service.
func NewRequest(apiVersion string, msg MessageType, endpoint MessageType) Request {
	req := Request{
		Msg: Message{
			Meta: struct {
				ApiVersion string
				SentAt     time.Time
			}{ApiVersion: apiVersion, SentAt: time.Now().UTC()},
			Msg: msg,
		},
		Endpoint: endpoint,
	}
	return req
}
