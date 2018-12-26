package ztcp

import (
	"net"
)

//PeerConn 对端连接信息
type PeerConn struct {
	Conn *net.TCPConn //连接
	Buf  []byte
	IP   string
	Port uint16
}

//Send 发送消息
func (p *PeerConn) Send(msgBuf []byte) (err error) {
	_, err = p.Conn.Write(msgBuf)
	if nil != err {
		gLog.Error("peerConn.Conn.Write:", err)
		return err
	}
	return nil
}

//连接是否有效
func (p *PeerConn) IsValid() bool {
	return nil != p.Conn
}

////////////////////////////////////////////////////////////////////////////////
//链接管理
type PeerConnData struct {
	PeerConn *PeerConn
}

//UserMap 用户map
type PeerConnMap map[*PeerConn]*PeerConnData

type PeerConnMgr struct {
	PeerConnMap PeerConnMap
}

//Init 初始化
func (p *PeerConnMgr) Init() {
	p.PeerConnMap = make(PeerConnMap)
}

//Add 加
func (p *PeerConnMgr) Add(peerConn *PeerConn) (peerConnData *PeerConnData) {
	peerConnData = new(PeerConnData)
	peerConnData.PeerConn = peerConn

	p.PeerConnMap[peerConn] = peerConnData
	return peerConnData
}

//Del 删
func (p *PeerConnMgr) Del(peerConn *PeerConn) {
	delete(p.PeerConnMap, peerConn)
}

//Find 查
func (p *PeerConnMgr) Find(peerConn *PeerConn) (peerConnData *PeerConnData) {
	peerConnData, _ = p.PeerConnMap[peerConn]
	return peerConnData
}
