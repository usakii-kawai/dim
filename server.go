package dim

import (
	"context"
	"net"
	"time"
)

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHeartbeat = time.Second * 55
)

type OpCode byte

// OpCode Token
const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

// server define a tcp/websocket interface
type Server interface {
	SetAcceptor(Acceptor)
	SetMessageListener(MessageListener)
	SetStateListener(StateListener)
	SetReadWait(time.Duration)
	SetChannelMap(ChannelMap)
	Start() error
	Push(string, []byte) error
	Shutdown(context.Context) error
}

// Acceptor
type Acceptor interface {
	Accept(Conn, time.Duration) (string, error)
}

// MessageListener
type MessageListener interface {
	Receive(Agent, []byte)
}

// StateListener
type StateListener interface {
	Disconnect(string) error
}

type Agent interface {
	ID() string
	Push([]byte) error
}

// Conn Connection
type Conn interface {
	net.Conn
	ReadFrame() (Frame, error)
	WriteFrame(OpCode, []byte) error
	Flush() error
}

// Channel is interface of client side
type Channel interface {
	Conn
	Agent
	Close() error
	Readloop(lst MessageListener) error
	SetWriteWait(time.Duration)
	SetReadWait(time.Duration)
}

// Client interface
type Client interface {
	ID() string
	Name() string
	Connect(string) error
	SetDialer(Dialer)
	Send([]byte) error
	Read() (Frame, error)
	Close()
}

// Diaer Diaer
type Dialer interface {
	DialAndHandshake(DialerContext) (net.Conn, error)
}

type DialerContext struct {
	Id      string
	Name    string
	Address string
	Timeout time.Duration
}

// Frame Frame
type Frame interface {
	SetOpCode(OpCode)
	GetOpCode() OpCode
	SetPayload([]byte)
	GetPayload() []byte
}
