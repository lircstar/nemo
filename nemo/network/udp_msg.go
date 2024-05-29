package network

import (
	"errors"
)

// --------------
// | len | data |
// --------------
type UdpMsgParser struct {
	minMsgLen    int
	maxMsgLen    int
	littleEndian bool
}

func newUdpMsgParser() *UdpMsgParser {
	p := new(UdpMsgParser)
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.littleEndian = false

	return p
}

// SetMsgLen It's dangerous to call the method on reading or writing
func (p *UdpMsgParser) SetMsgLen(minMsgLen int, maxMsgLen int) {

	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}
}

// SetByteOrder It's dangerous to call the method on reading or writing
func (p *UdpMsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// goroutine safe
func (p *UdpMsgParser) Read(conn *UDPConn) ([]byte, error) {
	if conn.closeFlag.Load() || conn.conn == nil {
		return nil, errors.New("connection is closed")
	}
	msgData := make([]byte, p.maxMsgLen)
	n, err := conn.conn.Read(msgData)
	if err != nil {
		return nil, err
	}

	return msgData[:n], nil
}

// goroutine safe
func (p *UdpMsgParser) Write(conn *UDPConn, args ...[]byte) error {
	// get len
	var msgLen int
	for i := 0; i < len(args); i++ {
		msgLen += len(args[i])
	}

	// check len
	if msgLen > p.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return errors.New("message too short")
	}

	msg := make([]byte, msgLen)

	// write data
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	if conn.closeFlag.Load() || conn.conn == nil {
		return errors.New("connection is closed")
	}
	var err error
	if conn.remote == nil {
		_, err = conn.conn.Write(msg)

	} else {
		_, err = conn.conn.WriteToUDP(msg, conn.remote)
	}

	if err != nil {
		return err
	}

	return nil
}
