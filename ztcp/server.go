package ztcp

import (
	"net"
	"strconv"
	"time"
)

type PeerConnRecvChan struct {
	RealPeerConn *PeerConn
	PeerConn     PeerConn
}

//己方作为服务
type Server struct {
	IsRun           bool   //是否运行
	PacketLengthMax uint32 //每个socket fd 最大包长

	OnInit           func() int                                                                 //初始化服务器
	OnFini           func() int                                                                 //服务器结束
	OnPeerConn       func(realPeerConn *PeerConn) int                                           //对端连上
	OnPeerConnClosed func(realPeerConn *PeerConn) int                                           //对端连接关闭
	OnPacket         func(RecvProtoHead *ProtoHead, RecvBuf []byte, realPeerConn *PeerConn) int //客户端消息

	peerConnRecvChan chan PeerConnRecvChan
}

//运行
func (this *Server) Run(ip string, port uint16, noDelay bool) (err error) {
	this.IsRun = true

	this.peerConnRecvChan = make(chan PeerConnRecvChan, 2000)

	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err)
		return
	}
	//优化[设置地址复用]
	//优化[设置监听的缓冲数量]
	listen, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		gLog.Crit("net.Listen:", err)
		return
	}

	Lock()
	this.OnInit()
	UnLock()

	defer func() {
		Lock()
		this.OnFini()
		UnLock()

		listen.Close()
	}()

	go this.handleAccept(listen, noDelay)

	go func() {
		//处理收到的数据
		for v := range this.peerConnRecvChan {
			Lock()
			if 0 == v.PeerConn.RecvProtoHead.MessageId {
				this.OnPeerConnClosed(v.RealPeerConn)
				v.RealPeerConn.Conn.Close()
			} else {
				this.OnPacket(&v.PeerConn.RecvProtoHead, v.PeerConn.RecvBuf, v.RealPeerConn)
			}
			UnLock()
		}
	}()
	//优化  这里添加一个chan 来处理此socket 发送出去的数据

	//优化[使用信号通知的方式结束循环]
	for this.IsRun {
		time.Sleep(60 * time.Second)
		//gLog.Debug("server run...")
	}

	gLog.Crit("server done...")

	return
}

func (this *Server) handleAccept(listen *net.TCPListener, noDelay bool) {
	var tempDelay time.Duration
	for {
		conn, err := listen.AcceptTCP()
		if nil != err {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				gLog.Crit("listen.Accept:", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}

			gLog.Crit("listen.Accept:", err)
			this.IsRun = false
			return
		}
		tempDelay = 0

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

	Lock()
	this.OnPeerConn(&peerConn)
	UnLock()

	defer func() {
		//使用MessageId ==0 的方式,表示断开链接.
		var peerConnRecvChan PeerConnRecvChan
		peerConnRecvChan.RealPeerConn = &peerConn
		peerConnRecvChan.PeerConn.RecvProtoHead.MessageId = 0
		this.peerConnRecvChan <- peerConnRecvChan
	}()

	var readIndex int
	for {
LoopRead:
		readNum, err := peerConn.Conn.Read(peerConn.RecvBuf[readIndex:])
		if nil != err {
			gLog.Error("peerConn.Conn.Read:", readNum, err)
			return
		}

		readIndex += readNum
		
		for{
			if readIndex < GProtoHeadLength { //长度不足一个包头中的长度大小
				goto LoopRead
			}
	
			/////////////////////////////////
			peerConn.parseProtoHeadPacketLength()
	
			if int(peerConn.RecvProtoHead.PacketLength) < GProtoHeadLength {
				gLog.Error("peerConn.RecvProtoHead.PacketLength:", peerConn.RecvProtoHead.PacketLength)
				return
			}
	
			if this.PacketLengthMax <= uint32(peerConn.RecvProtoHead.PacketLength) {
				gLog.Error("this.PacketLengthMax:", this.PacketLengthMax, peerConn.RecvProtoHead.PacketLength)
				return
			}
	
			if readIndex < int(peerConn.RecvProtoHead.PacketLength) {
				goto LoopRead
			}
	
			//有完整的包
			peerConn.parseProtoHead()
	
			var peerConnRecvChan PeerConnRecvChan
			peerConnRecvChan.PeerConn.RecvBuf = make([]byte, peerConn.RecvProtoHead.PacketLength)
			peerConnRecvChan.RealPeerConn = &peerConn
			copy(peerConnRecvChan.PeerConn.RecvBuf, peerConn.RecvBuf[:peerConn.RecvProtoHead.PacketLength])
			peerConnRecvChan.PeerConn.RecvProtoHead = peerConn.RecvProtoHead
			this.peerConnRecvChan <- peerConnRecvChan
	
			copy(peerConn.RecvBuf, peerConn.RecvBuf[peerConn.RecvProtoHead.PacketLength:readIndex])
			readIndex -= int(peerConn.RecvProtoHead.PacketLength)
		}
	}
}
