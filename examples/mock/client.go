package mock

import (
	"context"
	"dim"
	"dim/logger"
	"dim/tcp"
	"dim/websocket"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// ClientDemo
type ClientDemo struct{}

type WebsocketDialer struct{}

type TCPDialer struct{}

func (c *ClientDemo) Start(userID, protocol, addr string) {
	var cli dim.Client

	// 1. init clinet
	if protocol == "ws" {
		cli = websocket.NewClient(userID, "client", websocket.ClientOptions{})
		cli.SetDialer(&WebsocketDialer{})
	} else if protocol == "tcp" {
		cli = tcp.NewClient("test1", "client", tcp.ClientOptions{})
		cli.SetDialer(&TCPDialer{})
	}

	// 2. connect
	err := cli.Connect(addr)
	if err != nil {
		logger.Error(err)
	}
	count := 5
	go func() {
		// 3. send
		for i := 0; i < count; i++ {
			err := cli.Send([]byte("hello"))
			if err != nil {
				logger.Error(err)
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	// 4. recv
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info(err)
			break
		}
		if frame.GetOpCode() != dim.OpBinary {
			continue
		}
		recv++
		logger.Warnf("%s receive message [%s]", cli.ID(), frame.GetPayload())
		if recv == count {
			break
		}
	}
	cli.Close()
}

func (d *WebsocketDialer) DialAndHandshake(ctx dim.DialerContext) (net.Conn, error) {
	// 1 ws.Dial
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	// 2 user check
	err = wsutil.WriteClientBinary(conn, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}
	// 3 return conn
	return conn, nil
}

func (d *TCPDialer) DialAndHandshake(ctx dim.DialerContext) (net.Conn, error) {
	logger.Info("start dial: ", ctx.Address)
	// 1 net.Dial
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}

	// 2. send check
	err = tcp.WriteFrame(conn, dim.OpBinary, []byte(ctx.Id))
	if err != nil {
		return nil, err
	}

	// 3. return conn
	return conn, nil
}
