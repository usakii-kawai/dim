package mock

import (
	"dim"
	"dim/logger"
	"dim/naming"
	"errors"
	"time"

	"dim/websocket"

	"dim/tcp"
)

type ServerDemo struct{}
type ServerHandler struct{}

func (s *ServerDemo) Start(id, protocol, addr string) {
	var srv dim.Server
	service := &naming.DefaultService{
		Id:       id,
		Protocol: protocol,
	}
	if protocol == "ws" {
		srv = websocket.NewServer(addr, service)
	} else if protocol == "tcp" {
		srv = tcp.NewServer(addr, service)
	}

	handler := &ServerHandler{}

	srv.SetReadWait(dim.DefaultReadWait)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	err := srv.Start()
	if err != nil {
		panic(err)
	}
}

func (h *ServerHandler) Accept(conn dim.Conn, timeout time.Duration) (string, error) {
	// 1. read:
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	logger.Info("recv", frame.GetOpCode())

	// 2. parse
	userID := string(frame.GetOpCode())

	// 3. check
	if userID == "" {
		return "", errors.New("user id is invaild")
	}

	return userID, nil
}

// receive
func (h *ServerHandler) Receive(ag dim.Agent, payload []byte) {
	ack := string(payload) + " from server "
	_ = ag.Push([]byte(ack))
}

// disconnect
func (h *ServerHandler) Disconnect(id string) error {
	logger.Warnf("disconnect %s", id)
	return nil
}
