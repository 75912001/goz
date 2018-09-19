package ztcp

import (
	"net"
	"strconv"
	"time"

	"github.com/goz/zutility"
)

const (
	recvChanMaxCnt = 1000 //收数据channel的最大数量

)

//PeerConnRecvChan 对端链接接收的Chan
type PeerConnRecvChan struct {
	eventType int //0:普通消息,1:断开链接
	Buf       []byte
	PeerConn  *PeerConn
}

//Server 己方作为服务
type Server struct {
	IsRun           bool   //是否运行
	PacketLengthMax uint32 //每个socket fd 最大包长

	OnInit           func() int                                   //初始化服务器
	OnFini           func() int                                   //服务器结束
	OnPeerConn       func(peerConn *PeerConn) int                 //对端连上
	OnPeerConnClosed func(peerConn *PeerConn) int                 //对端连接关闭
	OnPacket         func(peerConn *PeerConn, recvBuf []byte) int //客户端消息
	//解析协议包头 返回长度:完整包总长度  返回0:不是完整包 返回-1:包错误
	OnParseProtoHead func(peerConn *PeerConn, length int) int
	peerConnRecvChan chan PeerConnRecvChan
}

//Run 运行
func (p *Server) Run(ip string, port uint16, noDelay bool) (err error) {
	p.IsRun = true

	p.peerConnRecvChan = make(chan PeerConnRecvChan, recvChanMaxCnt)

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

	zutility.Lock()

	p.OnInit()
	zutility.UnLock()

	defer func() {

		zutility.Lock()

		p.OnFini()
		zutility.UnLock()

		listen.Close()
	}()

	go p.handleAccept(listen, noDelay)

	go func() {
		//处理收到的数据
		for v := range p.peerConnRecvChan {

			zutility.Lock()

			if v.PeerConn.Invalid {
				zutility.UnLock()
				continue
			}
			if 1 == v.eventType {
				p.ClosePeer(v.PeerConn)
			} else {
				ret := p.OnPacket(v.PeerConn, v.PeerConn.Buf)
				if zutility.ErrorCodeDisconnectPeer == ret {
					p.ClosePeer(v.PeerConn)
				}
			}
			zutility.UnLock()
		}
	}()

	//优化[使用信号通知的方式结束循环]
	for p.IsRun {
		time.Sleep(1 * time.Second)
		//gLog.Debug("server run...")
	}

	gLog.Crit("server done...")

	return
}

//ClosePeer 关闭对方
func (p *Server) ClosePeer(peerConn *PeerConn) {
	peerConn.Invalid = true
	p.OnPeerConnClosed(peerConn)
	peerConn.Conn.Close()
}

func (p *Server) handleAccept(listen *net.TCPListener, noDelay bool) {
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
			p.IsRun = false
			return
		}
		tempDelay = 0

		conn.SetNoDelay(noDelay)
		conn.SetReadBuffer((int)(p.PacketLengthMax))
		conn.SetWriteBuffer((int)(p.PacketLengthMax))
		go p.handleConnection(conn)
	}
}

func (p *Server) handleConnection(conn *net.TCPConn) {
	var peerConn PeerConn
	peerConn.Conn = conn

	//优化[消耗内存过大]
	peerConn.Buf = make([]byte, p.PacketLengthMax)

	var peerIP = peerConn.Conn.RemoteAddr().String()
	gLog.Trace("connection from:", peerIP)

	zutility.Lock()

	p.OnPeerConn(&peerConn)
	zutility.UnLock()

	defer func() {
		//断开链接.
		var peerConnRecvChan PeerConnRecvChan
		peerConnRecvChan.PeerConn = &peerConn
		peerConnRecvChan.eventType = 1
		p.peerConnRecvChan <- peerConnRecvChan
	}()

	var readIndex int
	for {
	LoopRead:
		readNum, err := peerConn.Conn.Read(peerConn.Buf[readIndex:])
		if nil != err {
			gLog.Error("peerConn.Conn.Read:", readNum, err)
			return
		}

		readIndex += readNum

		for {

			zutility.Lock()

			packetLength := p.OnParseProtoHead(&peerConn, readIndex)
			if 0 == packetLength {
				zutility.UnLock()
				goto LoopRead
			}

			if -1 == packetLength {
				zutility.UnLock()
				gLog.Error("packetLength")
				return
			}

			var peerConnRecvChan PeerConnRecvChan
			peerConnRecvChan.eventType = 0
			peerConnRecvChan.Buf = make([]byte, packetLength)
			peerConnRecvChan.PeerConn = &peerConn
			copy(peerConnRecvChan.Buf, peerConn.Buf[:packetLength])
			p.peerConnRecvChan <- peerConnRecvChan

			copy(peerConn.Buf, peerConn.Buf[packetLength:readIndex])
			readIndex -= packetLength
			zutility.UnLock()
		}
	}
}
