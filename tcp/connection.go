package tcp

import (
	"dim"
	"io"
	"net"

	"dim/wire/endian"
)

// Frame Frame
type Frame struct {
	OpCode  dim.OpCode
	Payload []byte
}

// SetOpCode
func (f *Frame) SetOpCode(code dim.OpCode) {
	f.OpCode = code
}

// GetOpCode
func (f *Frame) GetOpCode() dim.OpCode {
	return f.OpCode
}

// SetPayload
func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

// GetPayload
func (f *Frame) GetPayload() []byte {
	return f.Payload
}

// Conn Conn
type TcpConn struct {
	net.Conn
}

// NewConn NewConn
func NewConn(conn net.Conn) *TcpConn {
	return &TcpConn{
		Conn: conn,
	}
}

// ReadFrame ReadFrame
func (c *TcpConn) ReadFrame() (dim.Frame, error) {
	opcode, err := endian.ReadUint8(c.Conn)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.Conn)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  dim.OpCode(opcode),
		Payload: payload,
	}, nil
}

// writeFrame writeFrame
func (c *TcpConn) WriteFrame(code dim.OpCode, payload []byte) error {
	return WriteFrame(c.Conn, code, payload)
}

// Flush Flush
func (c *TcpConn) Flush() error {
	return nil
}

// WriteFrame write a frame to w
func WriteFrame(w io.Writer, code dim.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
