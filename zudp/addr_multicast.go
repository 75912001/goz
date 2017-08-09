package zudp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/goz/ztcp"
	"github.com/goz/zutility"
	"golang.org/x/net/ipv4"
)

var gLog *zutility.Log

type serverAddr struct {
	name string
	id   uint32
	ip   string
	port uint16
	data string
}
type serverIdMap map[uint32]*serverAddr
type serverNameMap map[string]serverIdMap

type AddrMulticast struct {
	OnAddrMulticast func(name string, svr_id uint32, ip string, port uint16, data string)
	conn            *net.UDPConn
	mcaddr          *net.UDPAddr
	serverMap       serverNameMap //服务器地址信息
	addrBuffer      *bytes.Buffer //同步的服务器地址信息(发送数据)
	selfServerAddr  serverAddr    //自己服务器地址信息
}

/*
bug:该组播会收到其他组的消息,比如本组为#mcast_ip=239.0.0.1#mcast_port=5001,
也会收到#mcast_ip=239.0.0.2#mcast_port=5001的消息
*/
func (this *AddrMulticast) Run(ip string, port uint16, netName string,
	addrName string, addrId uint32, addrIp string, addrPort uint16, addrData string,
	log *zutility.Log) (err error) {
	this.selfServerAddr.name = addrName
	this.selfServerAddr.id = addrId
	this.selfServerAddr.ip = addrIp
	this.selfServerAddr.port = addrPort
	this.selfServerAddr.data = addrData

	gLog = log
	this.init()
	var str_addr = ip + ":" + strconv.Itoa(int(port))
	this.mcaddr, err = net.ResolveUDPAddr("udp4", str_addr)
	if nil != err {
		gLog.Crit("net.ResolveUDPAddr err:", err)
		return
	}

	this.conn, err = net.ListenUDP("udp4", this.mcaddr)
	if err != nil {
		gLog.Crit("ListenUDP err:", err)
		return
	}

	pc := ipv4.NewPacketConn(this.conn)

	iface, err := net.InterfaceByName(netName)
	if nil != err {
		gLog.Crit("can't find specified interface err:", err)
		return
	}

	str_addr_ipv4, _ := net.ResolveIPAddr("ip4", ip)
	err = pc.JoinGroup(iface, str_addr_ipv4)
	if nil != err {
		return
	}

	if loop, err := pc.MulticastLoopback(); err == nil {
		gLog.Trace("MulticastLoopback status:", loop)
		if !loop {
			if err := pc.SetMulticastLoopback(true); err != nil {
				gLog.Crit("SetMulticastLoopback err:", err)
			}
		}
	}

	var cmd uint32 = 2
	var arr_name [32]byte
	var arr_ip [16]byte
	var arr_data [32]byte
	copy(arr_name[:], addrName)
	copy(arr_ip[:], addrIp)
	copy(arr_data[:], addrData)

	this.addrBuffer = new(bytes.Buffer)
	binary.Write(this.addrBuffer, binary.LittleEndian, cmd)
	binary.Write(this.addrBuffer, binary.LittleEndian, addrId)
	binary.Write(this.addrBuffer, binary.LittleEndian, arr_name)
	binary.Write(this.addrBuffer, binary.LittleEndian, arr_ip)
	binary.Write(this.addrBuffer, binary.LittleEndian, addrPort)
	binary.Write(this.addrBuffer, binary.LittleEndian, arr_data)

	go this.handleRecv()
	go func() {
		for {
			this.doAddrSYN()
			//20-40sec 发送一次
			time.Sleep(time.Duration(rand.Intn(20)+20) * time.Second)
		}
	}()
	return
}

func (this *AddrMulticast) doAddrSYN() {
	_, err := this.conn.WriteToUDP(this.addrBuffer.Bytes(), this.mcaddr)
	if nil != err {
		gLog.Error("PeerConn.Conn.Write:", err)
		return
	}
}

func byteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[:i])
		}
	}
	return string(p)
}

func (this *AddrMulticast) handleRecv() {
	defer func() {
		this.conn.Close()
	}()

	var recvBuf []byte
	recvBuf = make([]byte, 1024)

	for {
		_, _, err := this.conn.ReadFromUDP(recvBuf)
		if nil != err {
			gLog.Crit("err:", err)
			break
		}

		var cmd uint32

		var ser serverAddr

		buf_cmd := bytes.NewBuffer(recvBuf[0:4])
		binary.Read(buf_cmd, binary.LittleEndian, &cmd)

		buf_svr_id := bytes.NewBuffer(recvBuf[4:8])
		binary.Read(buf_svr_id, binary.LittleEndian, &ser.id)

		ser.name = byteString(recvBuf[8:40])
		ser.ip = byteString(recvBuf[40:56])

		buf_port := bytes.NewBuffer(recvBuf[56:58])
		binary.Read(buf_port, binary.LittleEndian, &ser.port)

		ser.data = byteString(recvBuf[58:90])

		if this.selfServerAddr.name != ser.name || this.selfServerAddr.id != ser.id {
			if nil == this.find(ser.name, ser.id) {
				this.doAddrSYN()

				this.add(ser.name, ser.id, &ser)
			}
			ztcp.Lock()
			this.OnAddrMulticast(ser.name, ser.id, ser.ip, ser.port, ser.data)
			ztcp.UnLock()
		}
	}
}

//初始化
func (this *AddrMulticast) init() {
	this.serverMap = make(serverNameMap)
}

func (this *AddrMulticast) find(name string, id uint32) (s *serverAddr) {
	value, valid := this.serverMap[name]
	if valid {
		value2, valid2 := value[id]
		if valid2 {
			return value2
		}
	}
	return nil
}

//添加到内存中
func (this *AddrMulticast) add(name string, id uint32, s *serverAddr) {
	_, valid := this.serverMap[name]
	if valid {
		this.serverMap[name][id] = s
	} else {
		serverIdMap := make(serverIdMap)
		serverIdMap[id] = s
		this.serverMap[name] = serverIdMap
	}
}
