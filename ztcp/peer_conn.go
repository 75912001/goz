package ztcp

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/golang/protobuf/proto"
)

/////////////////////////////////////////////////////////////////////////////
//对端连接信息
type PeerConn struct {
	Conn          *net.TCPConn //连接
	RecvBuf       []byte
	RecvProtoHead protoHead
}

//发送消息
//todo 传送指针? pb
func (this *PeerConn) Send(pb proto.Message, messageId MESSAGE_ID,
	sessionId SESSION_ID, userId USER_ID, resultId RESULT_ID) (err error) {
	msgBuf, err := proto.Marshal(pb)
	if nil != err {
		gLog.Error("proto.Marshal:", err)
		return
	}

	var sendBufAllLength PACKET_LENGTH = PACKET_LENGTH(len(msgBuf) + GProtoHeadLength)

	headBuf := new(bytes.Buffer)

	binary.Write(headBuf, binary.LittleEndian, sendBufAllLength)
	binary.Write(headBuf, binary.LittleEndian, sessionId)
	binary.Write(headBuf, binary.LittleEndian, messageId)
	binary.Write(headBuf, binary.LittleEndian, resultId)
	binary.Write(headBuf, binary.LittleEndian, userId)

	//todo [优化]使用一个发送
	_, err = this.Conn.Write(headBuf.Bytes())
	if nil != err {
		gLog.Error("PeerConn.Conn.Write:", err)
		return
	}
	_, err = this.Conn.Write(msgBuf)
	if nil != err {
		gLog.Error("PeerConn.Conn.Write:", err)
		return
	}
	return
}

////////////////////////////////////////////////////////////////////////////////
//解析协议包头长度
func (this *PeerConn) parseProtoHeadPacketLength() {
	buf1 := bytes.NewBuffer(this.RecvBuf[0:protoHeadPacketLengthSize])
	binary.Read(buf1, binary.LittleEndian, &this.RecvProtoHead.PacketLength)
}

//解析协议包头
func (this *PeerConn) parseProtoHead() {
	buf1 := bytes.NewBuffer(this.RecvBuf[0:sumSize1])
	buf2 := bytes.NewBuffer(this.RecvBuf[sumSize1:sumSize2])
	buf3 := bytes.NewBuffer(this.RecvBuf[sumSize2:sumSize3])
	buf4 := bytes.NewBuffer(this.RecvBuf[sumSize3:sumSize4])
	buf5 := bytes.NewBuffer(this.RecvBuf[sumSize4:sumSize5])

	binary.Read(buf1, binary.LittleEndian, &this.RecvProtoHead.PacketLength)
	binary.Read(buf2, binary.LittleEndian, &this.RecvProtoHead.SessionId)
	binary.Read(buf3, binary.LittleEndian, &this.RecvProtoHead.MessageId)
	binary.Read(buf4, binary.LittleEndian, &this.RecvProtoHead.ResultId)
	binary.Read(buf5, binary.LittleEndian, &this.RecvProtoHead.UserId)
}
