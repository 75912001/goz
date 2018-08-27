package ztcp

import (
	"net"
	"strconv"

	"github.com/goz/zutility"
)

//己方作为客户端
type Client struct {
	OnConnClosed func(peerConn *PeerConn) int //对端连接关闭
	//对端消息
	OnPacket func(peerConn *PeerConn) int
	PeerConn PeerConn
}

//连接
func (this *Client) Connect(ip string, port uint16, recvBufMax int) (err error) {
	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err, addr)
		return
	}
	this.PeerConn.Conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		gLog.Crit("net.Dial:", err, addr)
		return
	}

	go this.recv(recvBufMax)
	return
}

func (this *Client) recv(recvBufMax int) {
	//优化[消耗内存过大]
	this.PeerConn.Buf = make([]byte, recvBufMax)

	defer func() {
		zutility.Lock()
		this.OnConnClosed(&this.PeerConn)
		zutility.UnLock()

		this.PeerConn.Conn.Close()
	}()

	var readIndex int
	for {
		readNum, err := this.PeerConn.Conn.Read(this.PeerConn.Buf[readIndex:])
		if nil != err {
			gLog.Error("Conn.Read:", readNum, err)
			break
		}

		readIndex += readNum

		if readIndex < GProtoHeadLength { //长度不足一个包头中的长度大小
			continue
		}

		this.PeerConn.parseProtoHeadPacketLength()

		if int(this.PeerConn.ProtoHead.PacketLength) < GProtoHeadLength {
			gLog.Error("this.PeerConn.ProtoHead.PacketLength:", this.PeerConn.ProtoHead.PacketLength)
			break
		}

		if readIndex < int(this.PeerConn.ProtoHead.PacketLength) {
			continue
		}

		//有完整的包
		this.PeerConn.parseProtoHead()

		zutility.Lock()
		this.OnPacket(&this.PeerConn)
		zutility.UnLock()

		copy(this.PeerConn.Buf, this.PeerConn.Buf[this.PeerConn.ProtoHead.PacketLength:readIndex])
		readIndex -= int(this.PeerConn.ProtoHead.PacketLength)
	}
}
