package xrServer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/75912001/goz/xrLog"
	"github.com/75912001/goz/xrServer"
	"github.com/75912001/goz/xrTcpHandle"
	"github.com/75912001/goz/xrTimer"
)

/*
型号名称：	MacBook Pro
型号标识符：	MacBookPro11,4
处理器名称：	Intel Core i7
处理器速度：	2.2 GHz
处理器数目：	1
核总数：	4
L2 缓存（每个核）：	256 KB
L3 缓存：	6 MB
内存：	16 GB


*/
//1.测试客户端链接
//1秒内多少次操作,持续10分钟后系统基本情况
//2.测试客户端关闭
//1秒内多少次操作,持续10分钟后系统基本情况
//3.测试客户端完整包
//1秒内多少次操作,持续10分钟后系统基本情况

func OnPeerConn(tcpPeer *xrTcpHandle.Peer) int {
	return 0
}
func OnPeerConnClosedServer(tcpPeer *xrTcpHandle.Peer) int {
	return 0
}
func OnPeerPacketServer(tcpPeer *xrTcpHandle.Peer, recvBuf []byte) int {
	return 0
}
func OnParseProtoHead(buf []byte, length int) int {
	return 0
}

func addCB(owner interface{}, data interface{}) int {
	second := time.Now().Unix()
	fmt.Println("begin:", second)

	switch owner.(type) {
	case *xrTimer.TimerMgr:
		vv, ok := owner.(*xrTimer.TimerMgr)
		if ok {
			vv.AddSecond(addCB, owner, 1, second)
		}
	}

	return 0
}
func TestFun(t *testing.T) {

	var log *xrLog.Log = new(xrLog.Log)
	log.Init("test_log")

	var ts xrServer.TcpServer
	ts.OnParseProtoHead = OnParseProtoHead
	ts.OnPeerConn = OnPeerConn
	ts.OnPeerConnClosedServer = OnPeerConnClosedServer
	ts.OnPeerPacketServer = OnPeerPacketServer

	var eventChan chan interface{}
	eventChan = make(chan interface{}, 10000)
	ts.Init(log, eventChan)

	//////////
	second := time.Now().Unix()

	var tm xrTimer.TimerMgr
	tm.Run(100, eventChan)

	tm.AddSecond(addCB, &tm, 1, second)

	//////////

	ts.Run("127.0.0.1:6677", 1024, 1024)

}
