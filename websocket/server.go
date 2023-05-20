package websocket

import (
	"context"
	"dim"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"dim/naming"

	"dim/logger"

	"github.com/gobwas/ws"
	"github.com/segmentio/ksuid"
)

// ServerOptions ServerOptions
type ServerOptions struct {
	loginwait time.Duration
	readwait  time.Duration
	writewait time.Duration
}

// websocket implent of the server interface
type Server struct {
	listen string
	naming.ServiceRegistration
	dim.Acceptor
	dim.ChannelMap
	dim.MessageListener
	dim.StateListener
	once    sync.Once
	options ServerOptions
}

// NewServer NewServer
func NewServer(listen string, service naming.ServiceRegistration) dim.Server {
	return &Server{
		listen:              listen,
		ServiceRegistration: service,
		options: ServerOptions{
			loginwait: dim.DefaultLoginWait,
			readwait:  dim.DefaultReadWait,
			writewait: time.Second * 10,
		},
	}
	// return nil
}

type defaultAcceptor struct{}

// Start server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"listen": s.listen,
		"id":     s.ServiceID(),
	})

	if s.Acceptor == nil {
		s.Acceptor = new(defaultAcceptor)
	}
	if s.StateListener == nil {
		return fmt.Errorf("StateListener is nil")
	}
	if s.ChannelMap == nil {
		s.ChannelMap = dim.NewChannels(100)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// step1 update
		rawconn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			resp(w, http.StatusBadRequest, err.Error())
			return
		}

		// step2 conn
		conn := NewConn(rawconn)

		// step3
		id, err := s.Accept(conn, s.options.loginwait)
		if err != nil {
			_ = conn.WriteFrame(dim.OpClose, []byte(err.Error()))
			conn.Close()
			return
		}
		if _, ok := s.Get(id); ok {
			log.Warnf("channel %s existed", id)
			_ = conn.WriteFrame(dim.OpClose, []byte("channelId is repeated"))
			conn.Close()
			return
		}

		// step4
		channel := dim.NewChannel(id, conn)
		channel.SetWriteWait(s.options.writewait)
		channel.SetReadWait(s.options.readwait)
		s.Add(channel)

		go func(ch dim.Channel) {
			// step5
			err := ch.Readloop(s.MessageListener)
			if err != nil {
				log.Info(err)
			}

			// step6
			s.Remove(ch.ID())
			err = s.Disconnect(ch.ID())
			if err != nil {
				log.Warn(err)
			}
			ch.Close()

		}(channel)
	})

	log.Infoln("started")
	return http.ListenAndServe(s.listen, mux)
}

// shutdown shutdown
func (s *Server) Shutdown(ctx context.Context) error {
	log := logger.WithFields(logger.Fields{
		"module": "ws.server",
		"id":     s.ServiceID(),
	})

	s.once.Do(func() {
		defer func() {
			log.Infoln("shutdown")
		}()

		// close channels
		chs := s.ChannelMap.All()
		for _, ch := range chs {
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
	if !ok {
		return errors.New("channel doesn't exist")
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

// SetChannels
func (s *Server) SetChannelMap(channels dim.ChannelMap) {
	s.ChannelMap = channels
}

// SetReadWait set read wait time.duration
func (s *Server) SetReadWait(readwait time.Duration) {
	s.options.readwait = readwait
}

func resp(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	if body != "" {
		_, _ = w.Write([]byte(body))
	}
	logger.Warn("response with code:%d %s", code, body)
}

// Accept defaultAcceptor
func (a *defaultAcceptor) Accept(conn dim.Conn, timeout time.Duration) (string, error) {
	return ksuid.Nil.String(), nil
}
