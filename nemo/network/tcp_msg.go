package network

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// --------------
// | len | data |
// --------------
type TcpMsgParser struct {
	lenMsgLen    int
	minMsgLen    int
	maxMsgLen    int
	littleEndian bool
}

func newTcpMsgParser() *TcpMsgParser {
	p := new(TcpMsgParser)
	p.lenMsgLen = 2
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.littleEndian = false

	return p
}

// SetMsgLen It's dangerous to call the method on reading or writing
func (p *TcpMsgParser) SetMsgLen(lenMsgLen int, minMsgLen int, maxMsgLen int) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		p.lenMsgLen = lenMsgLen
	}
	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max int
	switch p.lenMsgLen {
	case 1:
		max = math.MaxInt8
	case 2:
		max = math.MaxInt16
	case 4:
		max = math.MaxInt32
	}
	if p.minMsgLen > max {
		p.minMsgLen = max
	}
	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// SetByteOrder It's dangerous to call the method on reading or writing
func (p *TcpMsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// goroutine safe
func (p *TcpMsgParser) Read(conn *TCPConn) ([]byte, error) {
	var b [4]byte
	var bufMsgLen = b[:p.lenMsgLen] // SetReadDeadLine will execute anytime ... ?
	//if err := conn.conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
	//	return nil, err
	//}
	// read len
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err
	}
	//conn.conn.SetReadDeadline(time.Time{})

	// parse len
	var msgLen int
	switch p.lenMsgLen {
	case 1:
		msgLen = int(bufMsgLen[0])
	case 2:
		if p.littleEndian {
			msgLen = int(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen = int(binary.BigEndian.Uint16(bufMsgLen))
		}
	case 4:
		if p.littleEndian {
			msgLen = int(binary.LittleEndian.Uint32(bufMsgLen))
		} else {
			msgLen = int(binary.BigEndian.Uint32(bufMsgLen))
		}
	}

	// check len
	if msgLen > p.maxMsgLen {
		return nil, errors.New("message too long")
	} else if msgLen < p.minMsgLen {
		return nil, errors.New("message too short")
	}
	// data
	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}

	return msgData, nil
}

// goroutine safe
func (p *TcpMsgParser) Write(conn *TCPConn, args ...[]byte) error {
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
	msg := make([]byte, p.lenMsgLen+msgLen)

	// write len
	switch p.lenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		if p.littleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if p.littleEndian {
			binary.LittleEndian.PutUint32(msg, uint32(msgLen))
		} else {
			binary.BigEndian.PutUint32(msg, uint32(msgLen))
		}
	}

	// write data
	l := p.lenMsgLen
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	conn.Write(msg)

	return nil
}
