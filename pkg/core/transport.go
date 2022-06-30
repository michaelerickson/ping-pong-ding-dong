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
	// ListenAndServ receives Request types on the tx channel and places
	// received Message types on the rx channel. If an error occurs, the
	// function will stop immediately and return an error.
	ListenAndServ(ctx context.Context, rx chan Message, tx chan Request) error
}
