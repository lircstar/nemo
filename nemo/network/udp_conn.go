package network

import (
	"net"
	"sync/atomic"
)

type UDPConn struct {
	conn   *net.UDPConn
	remote *net.UDPAddr

	closeFlag atomic.Bool
	key       *connTrackKey
	agent     Agent

	msgParser *UdpMsgParser

	timeEvent chan Conn
}

func newUDPConn(msgParser *UdpMsgParser) *UDPConn {
	udpConn := new(UDPConn)
	udpConn.closeFlag.Store(false)
	udpConn.timeEvent = nil
	udpConn.msgParser = msgParser
	return udpConn
}

func (conn *UDPConn) ReadMsg() ([]byte, error) {
	return conn.msgParser.Read(conn)
}

func (conn *UDPConn) WriteMsg(args ...[]byte) error {
	return conn.msgParser.Write(conn, args...)
}
func (conn *UDPConn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}
func (conn *UDPConn) RemoteAddr() net.Addr {
	return conn.remote
}

func (conn *UDPConn) IsClosed() bool {
	return conn.closeFlag.Load()
}

func (conn *UDPConn) Close() {
	if conn.timeEvent != nil {
		conn.timeEvent <- conn
	}
	conn.closeFlag.Store(true)
}

func (conn *UDPConn) Destroy() {

}
