package xrMulticast

import (
	"bytes"
	"encoding/binary"
	"github.com/75912001/goz/xrLog"
	"github.com/75912001/goz/xrTcpHandle"
	"github.com/75912001/goz/xrUtility"
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
	"strconv"
	"time"
)

//使用与libel匹配的组播协议

var gLog *xrLog.Log

type serverAddr struct {
	name string //32个byte
	id   uint32
	ip   string //16个byte
	port uint16
	data string //32个byte
}
type serverIDMap map[uint32]*serverAddr
type serverNameMap map[string]serverIDMap

//AddrMulticast 地址组播
type AddrMulticast struct {
	conn           *net.UDPConn
	mcaddr         *net.UDPAddr
	serverMap      serverNameMap    //服务器地址信息
	addrBuffer     *bytes.Buffer    //同步的服务器地址信息(发送数据)
	selfServerAddr serverAddr       //自己服务器地址信息
	eventChan      chan interface{} //服务处理的事件
}

//Run 运行
func (p *AddrMulticast) Run(ip string, port uint16, netName string,
	addrName string, addrID uint32, addrIP string, addrPort uint16, addrData string,
	log *xrLog.Log, eventChan chan interface{}) (err error) {

	p.selfServerAddr.name = addrName
	p.selfServerAddr.id = addrID
	p.selfServerAddr.ip = addrIP
	p.selfServerAddr.port = addrPort
	p.selfServerAddr.data = addrData

	gLog = log

	p.eventChan = eventChan
	p.init()
	var strAddr = ip + ":" + strconv.Itoa(int(port))
	p.mcaddr, err = net.ResolveUDPAddr("udp4", strAddr)
	if nil != err {
		gLog.Crit("net.ResolveUDPAddr err:", err)
		return err
	}

	p.conn, err = net.ListenUDP("udp4", p.mcaddr)
	if err != nil {
		gLog.Crit("ListenUDP err:", err)
		return err
	}

	pc := ipv4.NewPacketConn(p.conn)

	iface, err := net.InterfaceByName(netName)
	if nil != err {
		gLog.Crit("can't find specified interface err:", err)
		return err
	}

	strAddrIpv4, _ := net.ResolveIPAddr("ip4", ip)
	err = pc.JoinGroup(iface, strAddrIpv4)
	if nil != err {
		gLog.Crit("err:", err, strAddrIpv4)
		return err
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
	var arrName [32]byte
	var arrIP [16]byte
	var arrData [32]byte
	copy(arrName[:], addrName)
	copy(arrIP[:], addrIP)
	copy(arrData[:], addrData)

	p.addrBuffer = new(bytes.Buffer)
	binary.Write(p.addrBuffer, binary.LittleEndian, cmd)
	binary.Write(p.addrBuffer, binary.LittleEndian, addrID)
	binary.Write(p.addrBuffer, binary.LittleEndian, arrName)
	binary.Write(p.addrBuffer, binary.LittleEndian, arrIP)
	binary.Write(p.addrBuffer, binary.LittleEndian, addrPort)
	binary.Write(p.addrBuffer, binary.LittleEndian, arrData)

	go p.handleRecv()
	go func() {
		for {
			p.doAddrSYN()
			//20-40sec 发送一次
			time.Sleep(time.Duration(rand.Intn(20)+20) * time.Second)
		}
	}()
	return err
}

func (p *AddrMulticast) doAddrSYN() {
	_, err := p.conn.WriteToUDP(p.addrBuffer.Bytes(), p.mcaddr)
	if nil != err {
		gLog.Error("PeerConn.Conn.Write:", err)
		return
	}
}

func (p *AddrMulticast) handleRecv() {
	defer func() {
		p.conn.Close()
	}()

	var recvBuf []byte
	recvBuf = make([]byte, 1024)

	for {
		_, _, err := p.conn.ReadFromUDP(recvBuf)
		if nil != err {
			gLog.Crit("err:", err)
			break
		}

		var cmd uint32

		var ser serverAddr

		bufCmd := bytes.NewBuffer(recvBuf[0:4])
		binary.Read(bufCmd, binary.LittleEndian, &cmd)

		bufSvrID := bytes.NewBuffer(recvBuf[4:8])
		binary.Read(bufSvrID, binary.LittleEndian, &ser.id)

		ser.name = xrUtility.Byte2String(recvBuf[8:40])
		ser.ip = xrUtility.Byte2String(recvBuf[40:56])

		bufPort := bytes.NewBuffer(recvBuf[56:58])
		binary.Read(bufPort, binary.LittleEndian, &ser.port)

		ser.data = xrUtility.Byte2String(recvBuf[58:90])

		if p.selfServerAddr.name != ser.name || p.selfServerAddr.id != ser.id {
			if nil == p.find(ser.name, ser.id) {
				p.doAddrSYN()

				p.add(ser.name, ser.id, &ser)
			}

			{
				var c xrTcpHandle.AddrMulticastEvent
				c.Name = ser.name
				c.ServerID = ser.id

				c.IP = ser.ip
				c.Port = ser.port
				c.Data = ser.data

				p.eventChan <- &c
			}

		}
	}
}

//初始化
func (p *AddrMulticast) init() {
	p.serverMap = make(serverNameMap)
}

func (p *AddrMulticast) find(name string, id uint32) (s *serverAddr) {
	value, valid := p.serverMap[name]
	if valid {
		value2, valid2 := value[id]
		if valid2 {
			return value2
		}
	}
	return nil
}

//添加到内存中
func (p *AddrMulticast) add(name string, id uint32, s *serverAddr) {
	_, valid := p.serverMap[name]
	if valid {
		p.serverMap[name][id] = s
	} else {
		serverIDMap := make(serverIDMap)
		serverIDMap[id] = s
		p.serverMap[name] = serverIDMap
	}
}
