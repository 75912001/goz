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
func (client *Client) Connect(ip string, port uint16, recvBufMax int) (err error) {
	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err, addr)
		return
	}
	client.PeerConn.Conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		gLog.Crit("net.Dial:", err, addr)
		return
	}

	go client.recv(recvBufMax)
	return
}

func (client *Client) recv(recvBufMax int) {
	//优化[消耗内存过大]
	client.PeerConn.Buf = make([]byte, recvBufMax)

	defer func() {
		zutility.Lock()
		client.OnConnClosed(&client.PeerConn)
		zutility.UnLock()

		client.PeerConn.Conn.Close()
	}()

	var readIndex int
	for {
		readNum, err := client.PeerConn.Conn.Read(client.PeerConn.Buf[readIndex:])
		if nil != err {
			gLog.Error("Conn.Read:", readNum, err)
			break
		}

		readIndex += readNum

		var packetLength int
		{
			defer func() {
				zutility.UnLock()
			}()
			zutility.Lock()
			packetLength = client.OnParseProtoHead(&client.PeerConn, readIndex)
			if 0 == packetLength {
				continue
			}
			if -1 == packetLength {
				gLog.Error("packetLength")
				break
			}
			client.OnPacket(&client.PeerConn)
		}
		copy(client.PeerConn.Buf, client.PeerConn.Buf[packetLength:readIndex])
		readIndex -= packetLength

		//以下移到应用层OnParseProtoHead中
		/*
			if readIndex < GProtoHeadLength { //长度不足一个包头中的长度大小
				continue
			}

			client.PeerConn.parseProtoHeadPacketLength()

			if int(client.PeerConn.ProtoHead.PacketLength) < GProtoHeadLength {
				gLog.Error("client.PeerConn.ProtoHead.PacketLength:", client.PeerConn.ProtoHead.PacketLength)
				break
			}

			if readIndex < int(client.PeerConn.ProtoHead.PacketLength) {
				continue
			}

			//有完整的包
			client.PeerConn.parseProtoHead()
		*/
	}
}
