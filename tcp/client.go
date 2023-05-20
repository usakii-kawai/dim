package tcp

import (
	"dim"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// ClientOptions ClientOptions
type ClientOptions struct {
	Heartbeat time.Duration
	ReadWait  time.Duration
	WriteWait time.Duration
}

// Client is a websocket implement of terminal
type Client struct {
	sync.Mutex
	dim.Dialer
	once    sync.Once
	id      string
	name    string
	conn    dim.Conn
	state   int32
	options ClientOptions
}

// NewClient NewClient
func NewClient(id, name string, opts ClientOptions) dim.Client {
	if opts.WriteWait == 0 {
		opts.WriteWait = dim.DefaultWriteWait
	}
	if opts.WriteWait == 0 {
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
		return fmt.Errorf("Client has connected")
	}

	rawconn, err := c.DialAndHandshake(dim.DialerContext{
		Id:      c.id,
		Name:    c.name,
		Address: addr,
		Timeout: dim.DefaultLoginWait,
	})

	if err != nil {
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
		return fmt.Errorf("client has connected")
	}
	if rawconn == nil {
		return fmt.Errorf("conn is nil")
	}
	c.conn = NewConn(rawconn)

	if c.options.Heartbeat > 0 {
		go func() {
			////
		}()
	}
	return nil
}

func (c *Client) SetDialer(dialer dim.Dialer) {
	c.Dialer = dialer
}

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
	return c.conn.WriteFrame(dim.OpBinary, payload)
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.conn == nil {
			return
		}
		_ = WriteFrame(c.conn, dim.OpClose, nil)

		c.conn.Close()
		atomic.CompareAndSwapInt32(&c.state, 1, 0)
	})
}

func (c *Client) Read() (dim.Frame, error) {
	if c.conn == nil {
		return nil, errors.New("connection is nil")
	}
	if c.options.Heartbeat > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.options.ReadWait))
	}
	frame, err := c.conn.ReadFrame()
	if err != nil {
		return nil, err
	}
	if frame.GetOpCode() == dim.OpClose {
		return nil, errors.New("remote side close the channel")
	}

	return frame, nil
}
