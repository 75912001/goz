package xrTcpHandle

import (
	"net"
	"sync"

	"fmt"
	"github.com/75912001/goz/xrLog"
)

//Peer 对端连接信息
type Peer struct {
	Conn     *net.TCPConn //连接
	IP       string
	Lock     sync.RWMutex
	SendChan chan interface{} //需要发送
}

//连接是否有效
func (p *Peer) IsValid() bool {
	return nil != p.Conn
}

//关闭链接
func (p *Peer) Close() {
	if nil != p.Conn {
		p.Conn.Close()
		p.Conn = nil
		close(p.SendChan)
	}
}

//发送数据(必须在处理EventChan事件中调用)
func (p *Peer) Send(buf []byte) (err error) {
	if !p.IsValid() {
		return
	}
	var c SendEventChan
	c.Buf = buf
	c.Peer = p
	p.SendChan <- &c
	return err
}

//链接成功事件channel
type ConnectEventChan struct {
	Peer *Peer
}

//断开链接事件channel 基于自身是server
type CloseConnectEventChanServer struct {
	Peer *Peer
}

//断开链接事件channel 基于自身是client
type CloseConnectEventChanClient struct {
	Peer *Peer
}

//收到数据事件channel 基于自身是server
type RecvEventChanServer struct {
	Buf  []byte
	Peer *Peer
}

//收到数据事件channel 基于自身是client
type RecvEventChanClient struct {
	Buf  []byte
	Peer *Peer
}

//添加组播事件
type AddrMulticastEvent struct{
	Name string
	ServerID uint32
	IP string
	Port uint16
	Data string
}

//发送数据事件channel
type SendEventChan struct {
	Buf  []byte
	Peer *Peer
}

//处理待发的数据
func HandleEventSend(sendChan chan interface{}, log *xrLog.Log) {
	for v := range sendChan {
		switch v.(type) {
		case *SendEventChan:
			vv, ok := v.(*SendEventChan)
			if ok {
				vv.Peer.Lock.Lock()
				if !vv.Peer.IsValid() {
					vv.Peer.Lock.Unlock()
					continue
				}
				_, err := vv.Peer.Conn.Write(vv.Buf)
				if nil != err {
					log.Error("send chan err:", err)
				}
				vv.Peer.Lock.Unlock()
			}
		default:
			log.Crit("no find event:", v)
		}
	}
}
