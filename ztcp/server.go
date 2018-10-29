package ztcp

import (
	"net"
	"strconv"
	"time"

	"github.com/75912001/goz/zutility"
)

const (
	eventTypeMsg        int = 0 //消息
	eventTypeDisConnect int = 1 //断开连接
	eventTypeConnect    int = 2 //链接上
)

//PeerConnEventChan 对端链接事件的Chan
type PeerConnEventChan struct {
	eventType int //0:普通消息,1:断开链接,2:链接上
	buf       []byte
	peerConn  *PeerConn
}

//Server 己方作为服务
type Server struct {
	IsRun           bool   //是否运行
	PacketLengthMax uint32 //每个socket fd 最大包长

	OnInit           func() int                                   //初始化服务器
	OnFini           func() int                                   //服务器结束
	OnPeerConn       func(peerConn *PeerConn) int                 //对端连上
	OnPeerConnClosed func(peerConn *PeerConn) int                 //对端连接关闭
	OnPeerPacket     func(peerConn *PeerConn, recvBuf []byte) int //对端包
	OnParseProtoHead func(peerConn *PeerConn, length int) int     //解析协议包头 返回长度:完整包总长度  返回0:不是完整包 返回-1:包错误
	peerConnRecvChan chan PeerConnEventChan
}

//Run 运行 recvChanMaxCnt:收数据channel的最大数量
func (p *Server) Run(ip string, port uint16, noDelay bool, recvChanMaxCnt uint32) (err error) {
	p.IsRun = true

	p.peerConnRecvChan = make(chan PeerConnEventChan, recvChanMaxCnt)

	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if nil != err {
		gLog.Crit("net.ResolveTCPAddr:", err)
		return err
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

			if !v.peerConn.IsValid() {
				zutility.UnLock()
				continue
			}

			if eventTypeDisConnect == v.eventType {
				p.ClosePeer(v.peerConn)
			} else if eventTypeMsg == v.eventType {
				ret := p.OnPeerPacket(v.peerConn, v.buf)
				if zutility.ECDisconnectPeer == ret {
					p.ClosePeer(v.peerConn)
				}
			} else if eventTypeConnect == v.eventType {
				//todo 链接上来的处理
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
	p.OnPeerConnClosed(peerConn)
	peerConn.conn.Close()
	peerConn.conn = nil
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
	peerConn.conn = conn

	peerConn.Buf = make([]byte, p.PacketLengthMax)

	var peerIP = peerConn.conn.RemoteAddr().String()
	gLog.Trace("connection from:", peerIP)

	zutility.Lock()
	p.OnPeerConn(&peerConn)
	zutility.UnLock()

	defer func() {
		//断开链接.
		var peerConnRecvChan PeerConnEventChan
		peerConnRecvChan.peerConn = &peerConn
		peerConnRecvChan.eventType = eventTypeDisConnect
		p.peerConnRecvChan <- peerConnRecvChan
	}()

	var readIndex int
	for {
	LoopRead:
		readNum, err := peerConn.conn.Read(peerConn.Buf[readIndex:])
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

			var peerConnRecvChan PeerConnEventChan
			peerConnRecvChan.eventType = eventTypeMsg
			peerConnRecvChan.buf = make([]byte, packetLength)
			peerConnRecvChan.peerConn = &peerConn
			copy(peerConnRecvChan.buf, peerConn.Buf[:packetLength])
			p.peerConnRecvChan <- peerConnRecvChan

			copy(peerConn.Buf, peerConn.Buf[packetLength:readIndex])
			readIndex -= packetLength
			zutility.UnLock()

			if 0 == readIndex {
				goto LoopRead
			}
		}
	}
}
