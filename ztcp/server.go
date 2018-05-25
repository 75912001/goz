package ztcp

import (
	"net"
	"strconv"
	"time"

	"github.com/goz/zutility"
)

//己方作为服务
type Server struct {
	IsRun            bool                         //是否运行
	OnInit           func() int                   //初始化服务器
	OnFini           func() int                   //服务器结束
	OnPeerConn       func(peerConn *PeerConn) int //对端连上
	OnPeerConnClosed func(peerConn *PeerConn) int //对端连接关闭
	//客户端消息
	//返回:EC_DISCONNECT_PEER断开客户端
	OnPacket        func(peerConn *PeerConn) int
	PacketLengthMax uint32 //每个socket fd 最大包长
}

//运行
func (this *Server) Run(ip string, port uint16, noDelay bool) (err error) {
	this.IsRun = true
	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err)
		return
	}
	//todo 优化[设置地址复用]
	//todo 优化[设置监听的缓冲数量]
	listen, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		gLog.Crit("net.Listen:", err)
		return
	}

	gLock.Lock()
	this.OnInit()
	gLock.Unlock()

	defer func() {
		gLock.Lock()
		this.OnFini()
		gLock.Unlock()

		listen.Close()
	}()

	go this.handleAccept(listen, noDelay)

	//todo 优化[使用信号通知的方式结束循环]
	for this.IsRun {
		time.Sleep(60 * time.Second)
		//gLog.Debug("server run...")
		//		gLog.Debug("server run...")
	}

	gLog.Crit("server done...")

	return
}

func (this *Server) handleAccept(listen *net.TCPListener, noDelay bool) {
	for {
		conn, err := listen.AcceptTCP()
		if nil != err {
			gLog.Crit("listen.Accept:", err)
			this.IsRun = false
			return
		}

		conn.SetNoDelay(noDelay)
		conn.SetReadBuffer((int)(this.PacketLengthMax))
		conn.SetWriteBuffer((int)(this.PacketLengthMax))
		go this.handleConnection(conn)
	}
}

func (this *Server) handleConnection(conn *net.TCPConn) {
	var peerConn PeerConn
	peerConn.Conn = conn

	//优化[消耗内存过大]
	peerConn.RecvBuf = make([]byte, this.PacketLengthMax)

	var peerIp = peerConn.Conn.RemoteAddr().String()
	gLog.Trace("connection from:", peerIp)

	gLock.Lock()
	this.OnPeerConn(&peerConn)
	gLock.Unlock()

	defer func() {
		gLock.Lock()
		this.OnPeerConnClosed(&peerConn)
		gLock.Unlock()

		peerConn.Conn.Close()
	}()

	var readIndex int
	for {
		readNum, err := peerConn.Conn.Read(peerConn.RecvBuf[readIndex:])
		if nil != err {
			gLog.Error("peerConn.Conn.Read:", readNum, err)
			break
		}

		readIndex += readNum
		if readIndex < GProtoHeadLength { //长度不足一个包头中的长度大小
			continue
		}

		peerConn.parseProtoHeadPacketLength()

		if int(peerConn.RecvProtoHead.PacketLength) < GProtoHeadLength {
			gLog.Error("peerConn.RecvProtoHead.PacketLength:", peerConn.RecvProtoHead.PacketLength)
			break
		}

		if this.PacketLengthMax <= uint32(peerConn.RecvProtoHead.PacketLength) {
			gLog.Error("this.PacketLengthMax:", this.PacketLengthMax, peerConn.RecvProtoHead.PacketLength)
			break
		}

		if readIndex < int(peerConn.RecvProtoHead.PacketLength) {
			continue
		}

		//有完整的包
		peerConn.parseProtoHead()

		gLock.Lock()
		ret := this.OnPacket(&peerConn)
		gLock.Unlock()

		if zutility.EC_DISCONNECT_PEER == ret {
			gLog.Error("OnPacket:", zutility.EC_DISCONNECT_PEER)
			break
		}
		copy(peerConn.RecvBuf, peerConn.RecvBuf[peerConn.RecvProtoHead.PacketLength:readIndex])
		readIndex -= int(peerConn.RecvProtoHead.PacketLength)
	}
}
