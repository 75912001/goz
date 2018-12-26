package ztcp

import (
	"net"
	"strconv"
	"time"

	"github.com/75912001/goz/zutility"
)

//PeerConnEventChan 对端链接事件的Chan
type PeerConnEventChan struct {
	eventType int //0:收到消息,1:断开链接,2:链接上,3:发送
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
	peerConnSendChan chan PeerConnEventChan
}

//Run 运行 recvChanMaxCnt:收数据channel的最大数量
func (p *Server) Run(ip string, port uint16, noDelay bool, recvChanMaxCnt uint32) (err error) {
	p.IsRun = true

	p.peerConnRecvChan = make(chan PeerConnEventChan, recvChanMaxCnt)
	p.peerConnSendChan = make(chan PeerConnEventChan, recvChanMaxCnt)

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
			} else if eventTypeRecvMsg == v.eventType {
				ret := p.OnPeerPacket(v.peerConn, v.buf)
				if zutility.ECDisconnectPeer == ret {
					p.ClosePeer(v.peerConn)
				}
			} else if eventTypeConnect == v.eventType {
				p.OnPeerConn(v.peerConn)
			}
			zutility.UnLock()
		}
	}()

	go func() {
		//处理待发的数据(这里与处理接收到的数据 使用同一个互斥锁，性能无任何提升。)
		for v := range p.peerConnSendChan {
			zutility.Lock()
			if !v.peerConn.IsValid() {
				zutility.UnLock()
				continue
			}

			//Send 发送消息
			_, err = v.peerConn.Conn.Write(v.buf)
			if nil != err {
				zutility.UnLock()
				continue
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

//发送到channel中
func (p *Server) AsyncSend(peer *PeerConn, msgBuf []byte) {
	//已完成(会将本应该立即发送的消息，放入队列中，不能保证发送消息的顺序在 队列中已接收消息之前处理)
	return
	{
		var peerConnSendChan PeerConnEventChan
		peerConnSendChan.peerConn = peer
		peerConnSendChan.eventType = eventTypeSendMsg
		peerConnSendChan.buf = make([]byte, len(msgBuf))
		copy(peerConnSendChan.buf, msgBuf[:])
		p.peerConnSendChan <- peerConnSendChan
	}
}

//ClosePeer 关闭对方
func (p *Server) ClosePeer(peerConn *PeerConn) {
	p.OnPeerConnClosed(peerConn)
	peerConn.Conn.Close()
	peerConn.Conn = nil
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

	peerConn.Buf = make([]byte, p.PacketLengthMax)

	var peerIP = peerConn.Conn.RemoteAddr().String()
	gLog.Trace("connection from:", peerIP)

	{
		var peerConnRecvChan PeerConnEventChan
		peerConnRecvChan.peerConn = &peerConn
		peerConnRecvChan.eventType = eventTypeConnect
		p.peerConnRecvChan <- peerConnRecvChan
	}

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
		readNum, err := peerConn.Conn.Read(peerConn.Buf[readIndex:])
		if nil != err {
			gLog.Error("peerConn.Conn.Read:", readNum, err)
			return
		}

		readIndex += readNum

		for {
			packetLength := p.OnParseProtoHead(&peerConn, readIndex)
			if 0 == packetLength {
				goto LoopRead
			}

			if -1 == packetLength {
				gLog.Error("packetLength")
				return
			}

			var peerConnRecvChan PeerConnEventChan
			peerConnRecvChan.eventType = eventTypeRecvMsg
			peerConnRecvChan.buf = make([]byte, packetLength)
			peerConnRecvChan.peerConn = &peerConn
			copy(peerConnRecvChan.buf, peerConn.Buf[:packetLength])
			p.peerConnRecvChan <- peerConnRecvChan

			copy(peerConn.Buf, peerConn.Buf[packetLength:readIndex])
			readIndex -= packetLength

			if 0 == readIndex {
				goto LoopRead
			}
		}
	}
}
