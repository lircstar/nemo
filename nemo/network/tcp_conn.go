package network

import (
	"nemo/sys/log"
	"net"
	"sync/atomic"
)

type TCPConn struct {
	//ConnOption
	conn      net.Conn
	writeChan chan []byte
	closeFlag atomic.Bool
	msgParser *TcpMsgParser
}

func newTCPConn(pendingWriteNum int, msgParser *TcpMsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser
	//tcpConn.closeFlag.Store(false)
	return tcpConn
}

func (tcpConn *TCPConn) start() {
	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				break
			}

			_, err := tcpConn.conn.Write(b)
			if err != nil {
				break
			}
		}

		tcpConn.closeFlag.Store(true)
		_ = tcpConn.conn.Close()
	}()
}

func (tcpConn *TCPConn) bindConn(conn net.Conn) {
	tcpConn.conn = conn
}

func (tcpConn *TCPConn) doDestroy() {

	if !tcpConn.closeFlag.Load() {

		tcpConn.closeFlag.Store(true)

		_ = tcpConn.conn.(*net.TCPConn).SetLinger(0)
		_ = tcpConn.conn.Close()

		close(tcpConn.writeChan)
	}
}

func (tcpConn *TCPConn) Destroy() {
	tcpConn.doDestroy()
}

func (tcpConn *TCPConn) Close() {
	if tcpConn.closeFlag.Load() {
		return
	}

	tcpConn.doWrite(nil)
	tcpConn.closeFlag.Store(true)
}

func (tcpConn *TCPConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doDestroy()
		return
	}

	tcpConn.writeChan <- b
}

// b must not be modified by the others goroutines
func (tcpConn *TCPConn) Write(b []byte) {

	if tcpConn.closeFlag.Load() || b == nil {
		return
	}

	tcpConn.doWrite(b)
}

func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

func (tcpConn *TCPConn) IsClosed() bool {
	return tcpConn.closeFlag.Load()
}

func (tcpConn *TCPConn) ReadMsg() ([]byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

func (tcpConn *TCPConn) WriteMsg(args ...[]byte) error {
	return tcpConn.msgParser.Write(tcpConn, args...)
}
