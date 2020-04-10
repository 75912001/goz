package xrServer

import (
	"net"
	"sync"
	"time"

	"github.com/75912001/goz/xrLog"
	"github.com/75912001/goz/xrTcpHandle"
	"github.com/75912001/goz/xrTimer"
)

type TcpServer struct {
	OnPeerConn             func(peerConn *xrTcpHandle.Peer) int                 //对端连上
	OnPeerConnClosedServer func(peerConn *xrTcpHandle.Peer) int                 //对端连接关闭 基于自身是server
	OnPeerConnClosedClient func(peerConn *xrTcpHandle.Peer) int                 //对端链接关闭 基于自身是client
	OnPeerPacketServer     func(peerConn *xrTcpHandle.Peer, recvBuf []byte) int //对端包 基于自身是server
	OnPeerPacketClient     func(peerConn *xrTcpHandle.Peer, recvBuf []byte) int //对端包 基于自身是client
	OnParseProtoHead       func(buf []byte, length int) int                     //解析协议包头 返回长度:完整包总长度  返回0:不是完整包 返回-1:包错误
	OnAddrMulticast 	   func(name string, svr_id uint32, ip string, port uint16, data string)
	log                    *xrLog.Log
	eventChan              chan interface{} //服务处理的事件
	rwBuffLen 				int
	packetLengthMax uint32
}

//初始化
//log:外部创建好的日志
//chanCnt:事件chan大小
//eventChan:外部传递的事件处理
func (p *TcpServer) Init(log *xrLog.Log, eventChan chan interface{}) {
	p.log = log
	p.eventChan = eventChan
}

//运行服务
//address:127.0.0.1:8787
//rwBuffLen:tcp recv/send 缓冲大小
//return:err
//packetLengthMax:最大包长(包头+包体)
func (p *TcpServer) Run(address string, rwBuffLen int, packetLengthMax uint32) (err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if nil != err {
		p.log.Emerg("net.ResolveTCPAddr:", err)
		return err
	}

	p.rwBuffLen = rwBuffLen
	p.packetLengthMax = packetLengthMax
	//优化[设置地址复用]
	//优化[设置监听的缓冲数量]

	listen, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		p.log.Emerg("net.Listen:", err)
		return err
	}
	defer func() {
		listen.Close()
	}()

	go func() {
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
					p.log.Crit("listen.Accept:", err, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				p.log.Emerg("listen.Accept:", err)
				return
			}
			tempDelay = 0

			go p.handleConnection(conn)
		}
	}()

	go p.HandleEventChan()

	for {
		time.Sleep(1 * time.Second)
	}
}

func (p *TcpServer) HandleEventChan() (err error) {
		//处理数据
		for v := range p.eventChan {
			switch v.(type) {
			case *xrTcpHandle.ConnectEventChan:
				vv, ok := v.(*xrTcpHandle.ConnectEventChan)
				if ok {
					//fmt.Println("xrTcpHandle.ConnectEventChan:",vv)
					p.OnPeerConn(vv.Peer)
				}
			case *xrTcpHandle.CloseConnectEventChanServer:
				vv, ok := v.(*xrTcpHandle.CloseConnectEventChanServer)
				vv.Peer.Lock.Lock()
				if ok {
					if vv.Peer.IsValid() {
						p.OnPeerConnClosedServer(vv.Peer)
					}
				}
				vv.Peer.Close()
				vv.Peer.Lock.Unlock()
			case *xrTcpHandle.CloseConnectEventChanClient:
				vv, ok := v.(*xrTcpHandle.CloseConnectEventChanClient)
				vv.Peer.Lock.Lock()
				if ok {
					if vv.Peer.IsValid() {
						p.OnPeerConnClosedClient(vv.Peer)
					}
				}
				vv.Peer.Close()
				vv.Peer.Lock.Unlock()
			case *xrTcpHandle.RecvEventChanServer:
				vv, ok := v.(*xrTcpHandle.RecvEventChanServer)
				if ok {
					if !vv.Peer.IsValid() {
						continue
					}
					p.OnPeerPacketServer(vv.Peer, vv.Buf)
				}
			case *xrTcpHandle.RecvEventChanClient:
				vv, ok := v.(*xrTcpHandle.RecvEventChanClient)
				if ok {
					if !vv.Peer.IsValid() {
						continue
					}
					p.OnPeerPacketClient(vv.Peer, vv.Buf)
				}
			case *xrTcpHandle.AddrMulticastEvent:
				vv, ok := v.(*xrTcpHandle.AddrMulticastEvent)
				if ok {
					p.OnAddrMulticast(vv.Name, vv.ServerID, vv.IP,vv.Port,vv.Data)
				}
			case *xrTimer.TimerSecond:
				vv, ok := v.(*xrTimer.TimerSecond)
				if ok {
					vv.Function(vv.Owner, vv.Data)
				}
			case *xrTimer.TimerMillisecond:
				vv, ok := v.(*xrTimer.TimerMillisecond)
				if ok {
					vv.Function(vv.Owner, vv.Data)
				}

			default:
				p.log.Crit("no find event:", v)
			}
		}
		return err
}
//发送数据(必须在处理EventChan事件中调用)
//func (p *TcpServer) Send(tcpPeer *xrTcpHandle.Peer, buf []byte) (err error) {
//	if !tcpPeer.IsValid() {
//		return
//	}
//	var c xrTcpHandle.SendEventChan
//	c.Buf = buf
//	c.Peer = tcpPeer
//	tcpPeer.SendChan <- &c
//	return err
//}

//关闭链接
func (p *TcpServer) CloseConnServer(tcpPeer *xrTcpHandle.Peer) (err error) {
	var c xrTcpHandle.CloseConnectEventChanServer
	c.Peer = tcpPeer
	p.eventChan <- &c
	return err
}

//关闭链接
func (p *TcpServer) CloseConnClient(tcpPeer *xrTcpHandle.Peer) (err error) {
	var c xrTcpHandle.CloseConnectEventChanClient
	c.Peer = tcpPeer
	p.eventChan <- &c
	return err
}
var handleConnectioncnt int
var handleConnectioncnt_lock     sync.RWMutex


func (p *TcpServer) handleConnection(conn *net.TCPConn) {
	conn.SetNoDelay(true)
	conn.SetReadBuffer(p.rwBuffLen)
	conn.SetWriteBuffer(p.rwBuffLen)

	var tcpPeer =new(xrTcpHandle.Peer)
	tcpPeer.Conn = conn
	tcpPeer.SendChan = make(chan interface{}, 10)

	tcpPeer.IP = tcpPeer.Conn.RemoteAddr().String()
	p.log.Trace("connection from:", tcpPeer.IP)

	{ //链接上
		var c xrTcpHandle.ConnectEventChan
		c.Peer = tcpPeer
		p.eventChan <- &c
	}

	var buf []byte
	defer func() { //断开链接
		var c xrTcpHandle.CloseConnectEventChanServer
		c.Peer = tcpPeer
		p.eventChan <- &c

		buf = nil
	}()

	go xrTcpHandle.HandleEventSend(tcpPeer.SendChan, p.log)

	//优化为内存池
	buf = make([]byte, p.packetLengthMax)
	var readIndex int
	for {
	LoopRead:
		readNum, err := tcpPeer.Conn.Read(buf[readIndex:])
		if nil != err {
			p.log.Error("tcpPeer.Conn.Read:", readNum, err)
			return
		}

		readIndex += readNum

		for {
			packetLength := p.OnParseProtoHead(buf, readIndex)
			if 0 == packetLength {
				goto LoopRead
			}

			if -1 == packetLength {
				p.log.Error("packetLength")
				return
			}

			{ //接受数据
				var c xrTcpHandle.RecvEventChanServer
				c.Buf = make([]byte, packetLength)
				c.Peer = tcpPeer
				copy(c.Buf, buf[:packetLength])
				p.eventChan <- &c
			}
			copy(buf, buf[packetLength:readIndex])
			readIndex -= packetLength

			if 0 == readIndex {
				goto LoopRead
			}
		}
	}
}
/*
func HandleRecv(tcpPeer *xrTcpHandle.Peer, packetLengthMax uint32, log *xrLog.Log, eventChan chan interface{}, onParseProtoHead func(buf []byte, length int) int) {
	//优化为内存池
	var buf []byte
	buf = make([]byte, packetLengthMax)
	var readIndex int
	for {
	LoopRead:
		readNum, err := tcpPeer.Conn.Read(buf[readIndex:])
		if nil != err {
			log.Error("tcpPeer.Conn.Read:", readNum, err)
			return
		}

		readIndex += readNum

		for {
			packetLength := onParseProtoHead(buf, readIndex)
			if 0 == packetLength {
				goto LoopRead
			}

			if -1 == packetLength {
				log.Error("packetLength")
				return
			}

			{ //接受数据
				var c xrTcpHandle.RecvEventChanServer
				c.Buf = make([]byte, packetLength)
				c.Peer = tcpPeer
				copy(c.Buf, buf[:packetLength])
				eventChan <- &c
			}
			copy(buf, buf[packetLength:readIndex])
			readIndex -= packetLength

			if 0 == readIndex {
				goto LoopRead
			}
		}
	}
}
*/