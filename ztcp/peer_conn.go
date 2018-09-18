package ztcp

import (
	"net"
)

//PeerConn 对端连接信息
type PeerConn struct {
	Conn    *net.TCPConn //连接
	Buf     []byte
	Invalid bool //无效
}

//Send 发送消息
func (p *PeerConn) Send(msgBuf []byte) (err error) {
	_, err = p.Conn.Write(msgBuf)
	if nil != err {
		gLog.Error("peerConn.Conn.Write:", err)
		return
	}
	return
}
