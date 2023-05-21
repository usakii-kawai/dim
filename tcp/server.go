package tcp

import (
	"context"
	"dim"
	"dim/logger"
	"dim/naming"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
)

// ServetOptions
type ServerOptions struct {
	loginwait time.Duration
	readwait  time.Duration
	writewait time.Duration
}

// Serve is a tcp implement of the server
type Server struct {
	listen string
	naming.ServiceRegistration
	dim.Acceptor
	dim.ChannelMap
	dim.MessageListener
	dim.StateListener
	once    sync.Once
	options ServerOptions
	quit    *dim.Event
}

// NewServer
func NewServer(listen string, service naming.ServiceRegistration) dim.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		ChannelMap:          dim.NewChannels(100),
		quit:                dim.NewEvent(),
		options: ServerOptions{
			loginwait: dim.DefaultLoginWait,
			readwait:  dim.DefaultReadWait,
			writewait: time.Second * 10,
		},
	}
}

// Start Server
func (s *Server) Start() error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}

	lst, err := net.Listen("tcp", s.listen)
	if err != nil {
		return err
	}
	log.Info("Started")

	for {
		rawconn, err := lst.Accept()
		if err != nil {
			rawconn.Close()
			log.Warn(err)
			continue
		}
		go func(rawconn net.Conn) {
			conn := NewConn(rawconn)

			id, err := s.Accept(conn, s.options.loginwait)
			if err != nil {
				_ = conn.WriteFrame(dim.OpClose, []byte(err.Error()))
				conn.Close()
				return
			}

			if _, ok := s.Get(id); ok {
				log.Warnf("channel %s existed", id)
				_ = conn.WriteFrame(dim.OpClose, []byte("channelID is repated"))
				conn.Close()
				return
			}

			channel := dim.NewChannel(id, conn)
			channel.SetReadWait(s.options.readwait)
			channel.SetWriteWait(s.options.writewait)

			s.Add(channel)
			log.Info("accept: ", channel)

			err = channel.Readloop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}
			s.Remove(channel.ID())
			_ = s.Disconnect(channel.ID())
			channel.Close()

		}(rawconn)

		select {
		case <-s.quit.Done():
			return fmt.Errorf("listen exited")
		default:
		}
	}
}

// Shutdown
func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "tcp.server",
		"id":     s.ServiceID(),
	})

	s.once.Do(func() {
		defer func() {
			log.Info("shutdown")
		}()

		// close channel
		channels := s.ChannelMap.All()
		for _, ch := range channels {
			ch.Close()

			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
	})

	return nil
}

// string channelID
// []byte data
func (s *Server) Push(id string, data []byte) error {
	ch, ok := s.ChannelMap.Get(id)
	if ok {
		return errors.New("channel no found")
	}
	return ch.Push(data)
}

// SetAcceptor
func (s *Server) SetAcceptor(acceptor dim.Acceptor) {
	s.Acceptor = acceptor
}

// SetMessageListener
func (s *Server) SetMessageListener(listener dim.MessageListener) {
	s.MessageListener = listener
}

// SetStateListener
func (s *Server) SetStateListener(listener dim.StateListener) {
	s.StateListener = listener
}

// SetReadWait set read wait duration
func (s *Server) SetReadWait(readwait time.Duration) {
	s.options.readwait = readwait
}

// SetChannels
func (s *Server) SetChannelMap(channels dim.ChannelMap) {
	s.ChannelMap = channels
}

type defaultAcceptor struct{}

// Accept defaultAcceptor
func (a *defaultAcceptor) Accept(conn dim.Conn, timeout time.Duration) (string, error) {
	return ksuid.New().String(), nil
}
