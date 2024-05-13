package network

import (
	"net"
	"sync"
)

type AgentSet map[connTrackKey] Agent


type UDPConn struct {
	sync.Mutex

	conn    	*net.UDPConn
	remote     	*net.UDPAddr

	closeFlag 	bool
	key       	*connTrackKey
	agent 		Agent

	msgParser 	*UdpMsgParser

	timeEvent	chan Conn
}

func newUDPConn(msgParser *UdpMsgParser) *UDPConn {
	udpConn := new(UDPConn)
	udpConn.closeFlag = false
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
	return conn.closeFlag
}

func (conn *UDPConn) Close() {
	conn.Lock()
	defer conn.Unlock()
	if conn.timeEvent != nil {
		conn.timeEvent <- conn
	}
	conn.closeFlag = true
}

func (conn *UDPConn) Destroy() {

}