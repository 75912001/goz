package main

import (
	"bytes"
	"encoding/binary"
)

//PacketLength 包总长度
type PacketLength uint32

//SessionID 会话ID
type SessionID uint32

//MessageID 消息ID
type MessageID uint32

//ResultID 结果ID
type ResultID uint32

//UserID 玩家ID
type UserID uint64

//GProtoHeadLength 包头长度
var GProtoHeadLength = 24

//ProtoHead 协议包头
type ProtoHead struct {
	PacketLength PacketLength //总包长度,包含包头＋包体长度
	SessionID    SessionID    //会话id
	MessageID    MessageID    //消息号
	ResultID     ResultID     //结果id
	UserID       UserID       //用户id
}

////////////////////////////////////////////////////////////////////////////////
//解析协议包头长度
func parseProtoHeadPacketLength(buf []byte) (packetLength PacketLength) {
	buf1 := bytes.NewBuffer(buf[0:4])
	binary.Read(buf1, binary.LittleEndian, &packetLength)
	return packetLength
}

//解析协议包头
func parseProtoHead(buf []byte) (packetLength PacketLength, sessionID SessionID, messageID MessageID, resultID ResultID, userID UserID) {
	buf1 := bytes.NewBuffer(buf[0:4])
	buf2 := bytes.NewBuffer(buf[4:8])
	buf3 := bytes.NewBuffer(buf[8:12])
	buf4 := bytes.NewBuffer(buf[12:16])
	buf5 := bytes.NewBuffer(buf[16:24])

	binary.Read(buf1, binary.LittleEndian, &packetLength)
	binary.Read(buf2, binary.LittleEndian, &sessionID)
	binary.Read(buf3, binary.LittleEndian, &messageID)
	binary.Read(buf4, binary.LittleEndian, &resultID)
	binary.Read(buf5, binary.LittleEndian, &userID)
	return packetLength, sessionID, messageID, resultID, userID
}
