package ztcp

import (
	"bytes"
	"encoding/binary"
	"net"
)

/////////////////////////////////////////////////////////////////////////////
//对端连接信息
type PeerConn struct {
	Conn      *net.TCPConn //连接
	Buf       []byte
	ProtoHead ProtoHead //优化 将这个字段去除,与应用层解耦
	//	ph        interface{}
}

//发送消息
func (this *PeerConn) Send(msgBuf *bytes.Buffer) (err error) {
	_, err = this.Conn.Write(msgBuf.Bytes())
	if nil != err {
		gLog.Error("PeerConn.Conn.Write:", err)
		return
	}
	return
}

////////////////////////////////////////////////////////////////////////////////
//解析协议包头长度
func (this *PeerConn) parseProtoHeadPacketLength() {
	buf1 := bytes.NewBuffer(this.Buf[0:protoHeadPacketLengthSize])
	binary.Read(buf1, binary.LittleEndian, &this.ProtoHead.PacketLength)
}

//解析协议包头
func (this *PeerConn) parseProtoHead() {
	buf1 := bytes.NewBuffer(this.Buf[0:sumSize1])
	buf2 := bytes.NewBuffer(this.Buf[sumSize1:sumSize2])
	buf3 := bytes.NewBuffer(this.Buf[sumSize2:sumSize3])
	buf4 := bytes.NewBuffer(this.Buf[sumSize3:sumSize4])
	buf5 := bytes.NewBuffer(this.Buf[sumSize4:sumSize5])

	binary.Read(buf1, binary.LittleEndian, &this.ProtoHead.PacketLength)
	binary.Read(buf2, binary.LittleEndian, &this.ProtoHead.SessionId)
	binary.Read(buf3, binary.LittleEndian, &this.ProtoHead.MessageId)
	binary.Read(buf4, binary.LittleEndian, &this.ProtoHead.ResultId)
	binary.Read(buf5, binary.LittleEndian, &this.ProtoHead.UserId)
}
