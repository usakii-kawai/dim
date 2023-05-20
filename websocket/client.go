package websocket

import (
	"dim"
	"dim/logger"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientOptions ClientOptions
type ClientOptions struct {
	Heartbeat time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

// Client is a websocket implement of the terminal
type Client struct {
	sync.Mutex
	dim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    net.Conn
	state   int32
	options ClientOptions
	dc      *dim.DialerContext
}

// NewClient
func NewClient(id, name string, opts ClientOptions) dim.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = dim.DefaultWriteWait
	}
	if opts.ReadWait == 0 {
		opts.ReadWait = dim.DefaultReadWait
	}

	cli := &Client{
		id:      id,
		name:    name,
		options: opts,
	}
	return cli
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) Connect(addr string) error {
	_, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if !atomic.CompareAndSwapInt32(&c.state, 0, 1) {
		return fmt.Errorf("client has connected")
	}

	// step 1 dial & handshake
	conn, err := c.Dialer.DialAndHandshake(dim.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: dim.DefaultLoginWait,
	})

	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return err
	}
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = conn

	// step 2 set heartbeat
	if c.options.Heartbeat > 0 {
		go func() {
			err := c.heartbealoop(conn)
			if err != nil {
				logger.Error("heartbealoop stopped ", err)
			}
		}()
	}

	return nil
}

// SetDialer
func (c *Client) SetDialer(dialer dim.Dialer) {
	c.Dialer = dialer
}

// Send data to connection
func (c *Client) Send(payload []byte) error {
	if atomic.LoadInt32(&c.state) == 0 {
		return fmt.Errorf("connection is nil")
	}
	c.Lock()
	defer c.Unlock()

	err := c.conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return nil
	}
	return wsutil.WriteClientMessage(c.conn, ws.OpBinary, payload)
}

// Close
func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}

		_ = wsutil.WriteClientMessage(c.conn, ws.OpClose, nil)

		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

// Read
func (c *Client) Read() (dim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := ws.ReadFrame(c.conn)
	if err != nil {
		return nil, err
	}

	if frame.Header.OpCode == ws.OpClose {
		return nil, errors.New("remove side close the channel")
	}

	return &Frame{
		raw: frame,
	}, nil
}

func (c *Client) heartbealoop(conn net.Conn) error {
	tick := time.NewTicker(c.options.Heartbeat)
	for range tick.C {
		if err := c.ping(conn); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ping(conn net.Conn) error {
	c.Lock()
	defer c.Unlock()
	err := conn.SetWriteDeadline(time.Now().Add(c.options.WriteWait))
	if err != nil {
		return err
	}
	logger.Trace("%s send ping to server", c.id)
	return wsutil.WriteClientMessage(conn, ws.OpPing, nil)
}
