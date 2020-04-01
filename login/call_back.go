package main

import (
	"fmt"
	"loginserv_msg"

	"github.com/75912001/goz/zprotobuf"
	"github.com/75912001/goz/ztcp"
	"github.com/golang/protobuf/proto"
	//	"github.com/goz/zutility"
)

//GcliPbFunMgr 客户端protobuf 回调函数管理器
var GcliPbFunMgr zprotobuf.PbFunMgr

func onInit() (ret int) {
	gLog.Trace("onInit")
	GcliPbFunMgr.Init(gLog)
	registerHandleFun()
	return 0
}

func onFini() (ret int) {
	gLog.Trace("onFini")
	return 0
}

////////////////////////////////////////////////////////////////////////////////
//客户端相关的回调函数
func onCliConn(realPeerConn *ztcp.PeerConn) (ret int) {
	gLog.Trace("onCliConn")
	user := GuserMgr.AddUser(realPeerConn)
	user.PeerConn = realPeerConn
	return 0
}

func onCliConnClosed(realPeerConn *ztcp.PeerConn) (ret int) {
	gLog.Trace("onCliConnClosed")

	user := GuserMgr.Find(realPeerConn)
	GuserMgr.DelUserID(user.UID)
	GuserMgr.DelUser(realPeerConn)

	return 0
}

var sumRecv int

func onParseProtoHead(peerConn *ztcp.PeerConn, length int) (ret int) {
	if isTest {
		sumRecv += length
		fmt.Println("总数:", sumRecv)
		return length
	}
	if length < GProtoHeadLength { //长度不足一个包头的长度大小
		return 0
	}

	packetLength := int(parseProtoHeadPacketLength(peerConn.Buf))

	if int(packetLength) < GProtoHeadLength {
		gLog.Error("PacketLength:", packetLength)
		return -1
	}
	if gServer.PacketLengthMax <= uint32(length) {
		gLog.Error("PacketLengthMax:", gServer.PacketLengthMax, length)
		return -1
	}

	if length < int(packetLength) {
		return 0
	}

	return packetLength
}

func onCliPacket(peerConn *ztcp.PeerConn, recvBuf []byte) (ret int) {
	if isTest {
		return 0
	}
	packetLength, sessionID, messageID, resultID, userID := parseProtoHead(recvBuf)
	gLog.Trace(packetLength, sessionID, messageID, resultID, userID)

	user := GuserMgr.Find(peerConn)
	if nil == user {
		gLog.Crit("")
		return 0
	}
	var userInterface UserInterface
	userInterface.User = user

	return GcliPbFunMgr.OnRecv(zprotobuf.MessageID(messageID), recvBuf[:GProtoHeadLength], recvBuf[GProtoHeadLength:packetLength], userInterface)
}

func onAddrMulticast(name string, id uint32, ip string, port uint16, data string) {
	gLog.Trace("", name, id, ip, port, data)
	//	if "gateway" != name {
	//		return
	//	}
	//	GuserMgr.userCnt = zutility.StringToUint32(&data)
	//	gLog.Trace(name, id, ip, port, data)
	return
}

//UserInterface 逻辑处理
type UserInterface struct {
	User *User
}

//注册函数
func registerHandleFun() (ret int) {
	GcliPbFunMgr.Register(zprotobuf.MessageID(loginserv_msg.CMD_LOGIN_MSG), OnLoginMsg, new(loginserv_msg.LoginMsg))
	GcliPbFunMgr.Register(zprotobuf.MessageID(loginserv_msg.CMD_PAY_GET_ID_MSG), OnPayGetIDMsg, new(loginserv_msg.PayGetIdMsg))
	return
}

//OnLoginMsg 登录消息
func OnLoginMsg(recvProtoHeadBuf []byte, protoMessage *proto.Message, obj interface{}) (ret int) {
	var userInterface UserInterface
	{
		var ok bool
		userInterface, ok = obj.(UserInterface)
		if !ok {
			return -1
		}
	}

	msgIn := (*protoMessage).(*loginserv_msg.LoginMsg)
	userInterface.User.IP = msgIn.GetIp()
	userInterface.User.Port = uint16(msgIn.GetPort())
	GuserMgr.userCnt = msgIn.GetGatewayCnt()

	gLog.Trace(msgIn.String())

	//peerConn := userInterface.User.PeerConn

	_, _, _, _, userID := parseProtoHead(recvProtoHeadBuf)
	GuserMgr.AddUserID(userID, userInterface.User)
	return
}
