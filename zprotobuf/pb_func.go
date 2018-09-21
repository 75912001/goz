package zprotobuf

import (
	"github.com/golang/protobuf/proto"
	"github.com/goz/zutility"
)

//MessageID 消息ID
type MessageID uint32

//PacketLength 包总长度
type PacketLength uint32

//PbFunMgr 管理器
type PbFunMgr struct {
	pbFunMap ProtoBufFunMap
	log      *zutility.Log
}

//Init 初始化管理器
func (p *PbFunMgr) Init(v *zutility.Log) {
	p.log = v
	p.pbFunMap = make(ProtoBufFunMap)
}

//Register 注册消息
func (p *PbFunMgr) Register(messageID MessageID, pbFun ProtoBufFun,
	protoMessage proto.Message) (ret int) {
	{
		pbFunHandle := p.find(messageID)
		p.log.Trace(messageID)
		if nil != pbFunHandle {
			p.log.Error("MessageId exist:", messageID)
			return zutility.ECSYS
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
func (p *PbFunMgr) OnRecv(messageID MessageID, recvProtoHeadBuf []byte, RecvBuf []byte, obj interface{}) (ret int) {
	pbFunHandle, ok := p.pbFunMap[messageID]
	if !ok {
		p.log.Error("MessageId inexist:", messageID)
		return zutility.ECDisconnectPeer
	}

	err := proto.Unmarshal(RecvBuf, *pbFunHandle.protoMessage)
	if nil != err {
		p.log.Error("proto.Unmarshal:", messageID, err)
		return zutility.ECDisconnectPeer
	}
	return pbFunHandle.pbFun(recvProtoHeadBuf, pbFunHandle.protoMessage, obj)
}

type pbFunHandle struct {
	pbFun        ProtoBufFun
	protoMessage *proto.Message
}

//ProtoBufFunMap 协议function map
type ProtoBufFunMap map[MessageID]*pbFunHandle

//ProtoBufFun 协议function
type ProtoBufFun func(recvProtoHeadBuf []byte, protoMessage *proto.Message, obj interface{}) (ret int)

func (p *PbFunMgr) find(messageID MessageID) (pbFunHandle *pbFunHandle) {
	pbFunHandle, _ = p.pbFunMap[messageID]
	return pbFunHandle
}
