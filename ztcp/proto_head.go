package ztcp

import (
	"encoding/binary"
)

//////////////////////////////////////////////////////////////////////////////
//协议包头
type PACKET_LENGTH uint32
type SESSION_ID uint32
type MESSAGE_ID uint32
type RESULT_ID uint32
type USER_ID uint64

//包头长度
var GProtoHeadLength int = sumSize5

////////////////////////////////////////////////////////////////////////////////
//协议包头
type protoHead struct {
	PacketLength PACKET_LENGTH //总包长度,包含包头＋包体长度
	SessionId    SESSION_ID    //会话id
	MessageId    MESSAGE_ID    //消息号
	ResultId     RESULT_ID     //结果id
	UserId       USER_ID       //用户id
}

/////////////////////////////////////////////////////////////////////////////
//计算字段长度
var packetLength PACKET_LENGTH
var messageId MESSAGE_ID
var sessionId SESSION_ID
var userId USER_ID
var resultId RESULT_ID

var protoHeadPacketLengthSize int = binary.Size(packetLength)
var protoHeadSessionIdSize int = binary.Size(sessionId)
var protoHeadMessageIdSize int = binary.Size(messageId)
var protoHeadResultIdSize int = binary.Size(resultId)
var protoHeadUserIdSize int = binary.Size(userId)

//前n个字段的总长度
var sumSize1 int = protoHeadPacketLengthSize
var sumSize2 int = sumSize1 + protoHeadSessionIdSize
var sumSize3 int = sumSize2 + protoHeadMessageIdSize
var sumSize4 int = sumSize3 + protoHeadResultIdSize
var sumSize5 int = sumSize4 + protoHeadUserIdSize
