package zprotobuf

import (
	"github.com/golang/protobuf/proto"
	"github.com/goz/ztcp"
	"github.com/goz/zutility"
)

//protobuf 管理器
type PbFunMgr struct {
	pbFunMap PB_FUN_MAP
	log      *zutility.Log
}

//初始化管理器
func (this *PbFunMgr) Init(v *zutility.Log) {
	this.log = v
	this.pbFunMap = make(PB_FUN_MAP)
}

//注册消息
func (this *PbFunMgr) Register(messageId ztcp.MESSAGE_ID, pbFun PB_FUN,
	protoMessage proto.Message) (ret int) {
	{
		pb_fun_handle := this.find(messageId)
		this.log.Trace(messageId)
		if nil != pb_fun_handle {
			this.log.Error("MessageId exist:", messageId)
			return zutility.EC_SYS
		}
	}
	{
		var pb_fun_handle *pbFunHandle = new(pbFunHandle)
		pb_fun_handle.pbFun = pbFun
		pb_fun_handle.protoMessage = &protoMessage
		this.pbFunMap[messageId] = pb_fun_handle
	}

	return 0
}

//收到消息
func (this *PbFunMgr) OnRecv(RecvProtoHead *ztcp.ProtoHead, RecvBuf []byte, obj interface{}) (ret int) {
	packetLength := RecvProtoHead.PacketLength
	messageId := RecvProtoHead.MessageId

	pbFunHandle, ok := this.pbFunMap[messageId]
	if !ok {
		this.log.Error("MessageId inexist:", messageId)
		return zutility.EC_DISCONNECT_PEER
	}

	err := proto.Unmarshal(RecvBuf[ztcp.GProtoHeadLength:packetLength],
		*pbFunHandle.protoMessage)
	if nil != err {
		this.log.Error("proto.Unmarshal:", messageId, err)
		return zutility.EC_DISCONNECT_PEER
	}
	return pbFunHandle.pbFun(obj, pbFunHandle.protoMessage)
}

////////////////////////////////////////////////////////////////////////////////
type pbFunHandle struct {
	pbFun        PB_FUN
	protoMessage *proto.Message
}
type PB_FUN_MAP map[ztcp.MESSAGE_ID]*pbFunHandle

////////////////////////////////////////////////////////////////////////////

type PB_FUN func(obj interface{}, protoMessage *proto.Message) (ret int)

func (this *PbFunMgr) find(messageId ztcp.MESSAGE_ID) (pb_fun_handle *pbFunHandle) {
	pb_fun_handle, _ = this.pbFunMap[messageId]
	return
}
