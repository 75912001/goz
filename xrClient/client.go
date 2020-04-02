package xrClient

import (
	"net"
	"strconv"

	"github.com/75912001/goz/xrLog"
	"github.com/75912001/goz/xrTcpHandle"
)

//Client 己方作为客户端
type TcpClient struct {
	OnParseProtoHead func(buf []byte, length int) int //解析协议包头 返回长度:完整包总长度  返回0:不是完整包 返回-1:包错误
	PeerConn         xrTcpHandle.Peer                 //对端链接
	recvBufMax       uint32                           //接受内存的最大尺寸
	log              *xrLog.Log
	eventChan        chan interface{} //服务处理的事件
}

//初始化
//log:外部创建好的日志
//recvBufMax:接受数据的最大长度
//eventChan:外部传递的事件处理
func (p *TcpClient) Init(log *xrLog.Log, recvBufMax uint32, eventChan chan interface{}) (err error) {
	p.log = log
	p.recvBufMax = recvBufMax
	p.eventChan = eventChan

	return err
}

//发送数据(必须在处理EventChan事件中调用)
//func (p *TcpClient) Send(buf []byte) (err error) {
//	if !p.PeerConn.IsValid() {
//		return
//	}
//	var c xrTcpHandle.SendEventChan
//	c.Buf = buf
//	c.Peer = &p.PeerConn
//	p.PeerConn.SendChan <- &c
//	return err
//}

//Connect 连接
func (p *TcpClient) Connect(ip string, port uint16) (err error) {
	var addr = ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if nil != err {
		p.log.Crit("net.ResolveTCPAddr:", err, addr)
		return err
	}
	p.PeerConn.Conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		p.log.Crit("net.Dial:", err, addr)
		return err
	}

	p.PeerConn.SendChan = make(chan interface{}, 1000)

	go p.recv()
	return nil
}

func (p *TcpClient) recv() {
	defer func() { //断开链接
		var c xrTcpHandle.CloseConnectEventChanClient
		c.Peer = &p.PeerConn
		p.eventChan <- &c
	}()

	go xrTcpHandle.HandleEventSend(p.PeerConn.SendChan, p.log)

	//优化为内存池
	var buf []byte
	buf = make([]byte, p.recvBufMax)
	var readIndex int
	for {
	LoopRead:
		readNum, err := p.PeerConn.Conn.Read(buf[readIndex:])
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
				var c xrTcpHandle.RecvEventChanClient
				c.Buf = make([]byte, packetLength)
				c.Peer = &p.PeerConn
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
