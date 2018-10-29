package ztcp

import (
	"net"
	"strconv"

	"github.com/75912001/goz/zutility"
)

//Client 己方作为客户端
type Client struct {
	OnPeerConnClosed func(peerConn *PeerConn) int             //对端连接关闭
	OnParseProtoHead func(peerConn *PeerConn, length int) int //解析协议包头 返回长度:完整包总长度  返回0:不是完整包 返回-1:包错误
	OnPeerPacket     func(peerConn *PeerConn) int             //对端包
	PeerConn         PeerConn
}

//Connect 连接
func (p *Client) Connect(ip string, port uint16, recvBufMax int) (err error) {
	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err, addr)
		return err
	}
	p.PeerConn.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		gLog.Crit("net.Dial:", err, addr)
		return err
	}

	go p.recv(recvBufMax)
	return nil
}

func (p *Client) recv(recvBufMax int) {
	p.PeerConn.Buf = make([]byte, recvBufMax)

	defer func() {
		zutility.Lock()
		p.OnPeerConnClosed(&p.PeerConn)
		p.PeerConn.conn.Close()
		p.PeerConn.conn = nil
		zutility.UnLock()
	}()

	var readIndex int
	for {
		readNum, err := p.PeerConn.conn.Read(p.PeerConn.Buf[readIndex:])
		if nil != err {
			gLog.Error("Conn.Read:", readNum, err)
			break
		}

		readIndex += readNum

		var packetLength int
		{
			zutility.Lock()

			packetLength = p.OnParseProtoHead(&p.PeerConn, readIndex)
			if 0 == packetLength {
				zutility.UnLock()
				continue
			}
			if -1 == packetLength {
				zutility.UnLock()
				gLog.Error("packetLength")
				break
			}
			p.OnPeerPacket(&p.PeerConn)

			zutility.UnLock()
		}
		copy(p.PeerConn.Buf, p.PeerConn.Buf[packetLength:readIndex])
		readIndex -= packetLength
	}
}
