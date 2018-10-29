package ztcp

import (
	"net"
)

//PeerConn 对端连接信息
type PeerConn struct {
	conn *net.TCPConn //连接
	Buf  []byte
	//	valid bool //有效
}

//Send 发送消息
func (p *PeerConn) Send(msgBuf []byte) (err error) {
	_, err = p.conn.Write(msgBuf)
	if nil != err {
		gLog.Error("peerConn.Conn.Write:", err)
		return err
	}
	return nil
}

//连接是否有效
func (p *PeerConn) IsValid() bool {
	return nil != p.conn
}
