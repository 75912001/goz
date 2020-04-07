package xrProtobuf

import (
	"github.com/75912001/goz/xrLog"

	"github.com/75912001/goz/xrUtility"
	"github.com/golang/protobuf/proto"
)

//MessageID 消息ID
type MessageID uint32

//PacketLength 包总长度
type PacketLength uint32

//PbFunMgr 管理器
type PbFunMgr struct {
	pbFunMap ProtoBufFunMap
	log      *xrLog.Log
}

//Init 初始化管理器
func (p *PbFunMgr) Init(v *xrLog.Log) {
	p.log = v
	p.pbFunMap = make(ProtoBufFunMap)
}

//Register 注册消息
func (p *PbFunMgr) Register(messageID MessageID, pbFun ProtoBufFun,
	protoMessage proto.Message) (ret int) {
	{
		pbFunHandle := p.find(messageID)
		if nil != pbFunHandle {
			p.log.Emerg("MessageId exist:", messageID)
			return xrUtility.ECSYS
		}
	}
	{
		var pbFunHandle = new(pbFunHandle)
		pbFunHandle.pbFun = pbFun
		pbFunHandle.protoMessage = &protoMessage
		p.pbFunMap[messageID] = pbFunHandle
	}

	return 0
}

//OnRecv 收到消息
func (p *PbFunMgr) OnRecv(messageID MessageID, protoHead interface{}, bodyBuf []byte, obj interface{}) (ret int) {
	pbFunHandle, ok := p.pbFunMap[messageID]
	if !ok {
		p.log.Error("MessageId inexist:", messageID)
		return xrUtility.ECDisconnectPeer
	}

	err := proto.Unmarshal(bodyBuf, *pbFunHandle.protoMessage)
	if nil != err {
		p.log.Error("proto.Unmarshal:", messageID, err)
		return xrUtility.ECDisconnectPeer
	}
	return pbFunHandle.pbFun(protoHead, pbFunHandle.protoMessage, obj)
}

type pbFunHandle struct {
	pbFun        ProtoBufFun
	protoMessage *proto.Message
}

//ProtoBufFunMap 协议function map
type ProtoBufFunMap map[MessageID]*pbFunHandle

//ProtoBufFun 协议function
type ProtoBufFun func(protoHead interface{}, protoMessage *proto.Message, obj interface{}) (ret int)

func (p *PbFunMgr) find(messageID MessageID) (pbFunHandle *pbFunHandle) {
	pbFunHandle, _ = p.pbFunMap[messageID]
	return pbFunHandle
}
