package ztcp

import (
	"net"
	"strconv"

	"github.com/goz/zutility"
)

//Client 己方作为客户端
type Client struct {
	OnConnClosed func(peerConn *PeerConn) int //对端连接关闭
	//解析协议包头 返回长度:完整包总长度  返回0:不是完整包 返回-1:包错误
	OnParseProtoHead func(peerConn *PeerConn, length int) int
	//对端消息
	OnPacket func(peerConn *PeerConn) int
	PeerConn PeerConn
}

//Connect 连接
func (p *Client) Connect(ip string, port uint16, recvBufMax int) (err error) {
	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err, addr)
		return
	}
	p.PeerConn.Conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		gLog.Crit("net.Dial:", err, addr)
		return
	}

	go p.recv(recvBufMax)
	return
}

func (p *Client) recv(recvBufMax int) {
	//优化[消耗内存过大]
	p.PeerConn.Buf = make([]byte, recvBufMax)

	defer func() {

		zutility.Lock()

		p.OnConnClosed(&p.PeerConn)
		zutility.UnLock()

		p.PeerConn.Conn.Close()
	}()

	var readIndex int
	for {
		readNum, err := p.PeerConn.Conn.Read(p.PeerConn.Buf[readIndex:])
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
			p.OnPacket(&p.PeerConn)
			zutility.UnLock()
		}
		copy(p.PeerConn.Buf, p.PeerConn.Buf[packetLength:readIndex])
		readIndex -= packetLength

		//以下移到应用层OnParseProtoHead中
		/*
			if readIndex < GProtoHeadLength { //长度不足一个包头中的长度大小
				continue
			}

			p.PeerConn.parseProtoHeadPacketLength()

			if int(p.PeerConn.ProtoHead.PacketLength) < GProtoHeadLength {
				gLog.Error("client.PeerConn.ProtoHead.PacketLength:", p.PeerConn.ProtoHead.PacketLength)
				break
			}

			if readIndex < int(p.PeerConn.ProtoHead.PacketLength) {
				continue
			}

			//有完整的包
			p.PeerConn.parseProtoHead()
		*/
	}
}
