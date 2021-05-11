package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/golang/protobuf/proto"

	"math/rand"
	"runtime"
	"strconv"

	"time"

	"github.com/75912001/goz/zhttp"
	"github.com/75912001/goz/ztcp"
	"github.com/75912001/goz/zudp"
	"github.com/75912001/goz/zutility"
)

var gLog *zutility.Log
var gServer *ztcp.Server
var gIni *zutility.Ini
var gAddrMulticast *zudp.AddrMulticast
var gPlatform uint32

var gWxAppid string
var gWxSecret string
var gWxHeadSize string

func send(peerConn *ztcp.PeerConn, pb proto.Message,
	messageID MessageID,
	sessionID SessionID,
	userID UserID,
	resultID ResultID) (err error) {
	msgBuf, err := proto.Marshal(pb)
	if nil != err {
		gLog.Error("proto.Marshal:", err)
		return
	}

	var sendBufAllLength = PacketLength(GProtoHeadLength + len(msgBuf))

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, sendBufAllLength)
	binary.Write(buf, binary.LittleEndian, sessionID)
	binary.Write(buf, binary.LittleEndian, messageID)
	binary.Write(buf, binary.LittleEndian, resultID)
	binary.Write(buf, binary.LittleEndian, userID)

	binary.Write(buf, binary.LittleEndian, msgBuf)
	if nil == err {

		return peerConn.Send(buf.Bytes())
	}
	return err
}

/*
var strxml2 string = `<xml><appid><![CDATA[wx37999154014daf66]]></appid>
<attach><![CDATA[全家斗牛 50点券]]></attach>
<bank_type><![CDATA[CFT]]></bank_type>
<cash_fee><![CDATA[1]]></cash_fee>
<fee_type><![CDATA[CNY]]></fee_type>
<is_subscribe><![CDATA[N]]></is_subscribe>
<mch_id><![CDATA[1291085501]]></mch_id>
<nonce_str><![CDATA[qtic58brqx8q62d6n1s8thps9svfz1i5]]></nonce_str>
<openid><![CDATA[oW8uswjSq1wwpOO0y3GJB-tcIb7s]]></openid>
<out_trade_no><![CDATA[a9df3de9b266930548033661ffd47d26]]></out_trade_no>
<result_code><![CDATA[SUCCESS]]></result_code>
<return_code><![CDATA[SUCCESS]]></return_code>
<sign><![CDATA[7069DA8E00EAD9C88EECF6DF854DB0B0]]></sign>
<time_end><![CDATA[20170912152630]]></time_end>
<total_fee>1</total_fee>
<trade_type><![CDATA[NATIVE]]></trade_type>
<transaction_id><![CDATA[4006722001201709121681531020]]></transaction_id>
</xml>`
*/
////////////////////////////////////////////////////////////////////////////////
const isTest = false

func main() {


	////////////////////////////////////////////////////////////////////////////
	//log
	logPath := gIni.GetString("log", "path", "default.log.")
	gLog = new(zutility.Log)
	gLog.Init(logPath, 1000)
	gLog.SetLevel(int(logLevel))
	defer gLog.DeInit()

	gLog.Trace("server runing...", time.Now())
	gLog.Trace("OS:", zutility.ShowOS())

	//加载pay.xml
	{
		ret := GpayCfgMgr.LoadPayCfg()
		if nil != ret {
			fmt.Println("load pay.xml err!")
			return
		}
	}
	////////////////////////////////////////////////////////////////////////////
	//mysql
	var mysqlPort uint16
	{
		ip := gIni.GetString("pay", "mysql_ip", "")
		mysqlPort = gIni.GetUint16("pay", "mysql_port", 0)
		user := gIni.GetString("pay", "mysql_user", "")
		pwd := gIni.GetString("pay", "mysql_pwd", "")
		gMysqlDbName = gIni.GetString("pay", "mysql_db_name", "")
		if 0 != mysqlPort {
			gPayDataSourceName = user + ":" + pwd + "@tcp(" + ip + ":" + strconv.Itoa(int(mysqlPort)) + ")/" + gMysqlDbName + "?charset=utf8"
			if 0 != initMysql() {
				gLog.Crit("mysql err")
				return
			}
		}
	}

	////////////////////////////////////////////////////////////////////////////
	//组播包
	//	addr_mcast_ip := gIni.GetString("addr_multicast", "mcast_ip", "")
	//	addr_mcast_port := gIni.GetUint16("addr_multicast", "mcast_port", "")
	//	addr_mcast_incoming_if := gIni.GetString("addr_multicast", "mcast_incoming_if", "")
	//	addr_mcast_data := gIni.GetString("addr_multicast", "data", "")
	//	gAddrMulticast = new(zudp.AddrMulticast)
	//	gAddrMulticast.OnAddrMulticast = onAddrMulticast
	//	gAddrMulticast.Run(addr_mcast_ip, addr_mcast_port, addr_mcast_incoming_if,
	//	server_name, server_id, server_ip, server_port, addr_mcast_data, gLog)
	////////////////////////////////////////////////////////////////////////////
	{
		ip := gIni.GetString("http_server", "ip", "")
		port := gIni.GetUint16("http_server", "port", 0)
		zhttp.SetLog(gLog)
		//HTTP 登录服务
		{
			gLoginHTTPServer.AddHandler("/login", loginHTTPHandler)
			go gLoginHTTPServer.Run(ip, port)
		}

		//HTTP 微信web登录服务 cb
		{
			gLoginHTTPServer.AddHandler("/wx_web_login_cb", wxWebLoginHTTPHandler)
			go gLoginHTTPServer.Run(ip, port)
		}
		//HTTP 微信web登录服务 客户端获取code
		{
			gLoginHTTPServer.AddHandler("/get_wx_web_login", getWxWebLoginHTTPHandler)
			go gLoginHTTPServer.Run(ip, port)
		}
		//HTTP 客户端获取pay_get_id
		{
			gLoginHTTPServer.AddHandler("/pay_get_id", payGetIDHTTPHandler)
			go gLoginHTTPServer.Run(ip, port)
		}
	}
	////////////////////////////////////////////////////////////////////////////
	//HTTP pay服务
	{
		//	天津:/tj_pay
		//	鹏博士:/pbs_pay
		//	支付宝:/zfb_pay
		//	微信:/wx_pay
		//	当贝:/dangbei_pay

		ip := gIni.GetString("pay", "http_ip", "")
		port := gIni.GetUint16("pay", "http_port", 0)
		urlCallbackPattern := gIni.GetString("pay", "url_callback_pattern", "")
		if 0 != port {
			fmt.Println(urlCallbackPattern)
			ucp := zutility.StringSplit(&urlCallbackPattern, ";")
			fmt.Println(ucp)
			for index, value := range ucp {
				fmt.Println(index, value)
				if "/tj_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, tjPayHTTPHandler)
				}
				if "/pbs_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, pbsPayHTTPHandler)
				}
				if "/zfb_pay" == value {
					gZfbPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, zfbPayHTTPSHandler)
				}
				if "/wx_pay" == value {
					gWxPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, wxPayHTTPSHandler)
				}
				if "/dangbei_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, dangbeiPayHTTPHandler)
				}
				if "/shafa_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, shafaPayHTTPHandler)
				}
				if "/putao_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, putaoPayHTTPHandler)
				}
				if "/wangsu_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, wangSuPayHTTPHandler)
				}
				if "/haixin_pay" == value {
					gPayURLCallback = "http://" + ip + ":" + zutility.IntToString(int(port)) + value
					gPayHTTPServer.AddHandler(value, haiXinPayHTTPHandler)
				}
			}
			go gPayHTTPServer.Run(ip, port)
		}
	}

	//////////////////////////////////////////////////////////////////
	//做为服务端
	{ //设置回调函数
		gServer = new(ztcp.Server)
		ztcp.SetLog(gLog)
		gServer.OnInit = onInit
		gServer.OnFini = onFini
		gServer.OnPeerConnClosed = onCliConnClosed
		gServer.OnPeerConn = onCliConn
		gServer.OnPeerPacket = onCliPacket
		gServer.OnParseProtoHead = onParseProtoHead

		//运行
		delay := true



		if 0 != serverPort {
			gLog.Trace(serverIP, serverPort, delay)
			go gServer.Run(serverIP, serverPort, delay, 1000)
		}
	}
	{ //测试

	}

	for {
		time.Sleep(60 * time.Second)
		//		gLog.Trace("server run...")
		{ //清理wx web 登录的过期数据
			nowTimeSec := time.Now().Unix()
			//			GwxWebLoginLock.Lock()
			for key, value := range GwxWebLoginMgr.wxWebLoginMap {
				if value.TimeSec+60 < nowTimeSec {
					delete(GwxWebLoginMgr.wxWebLoginMap, key)
				}
			}
			//			GwxWebLoginLock.Unlock()
		}
		{
			if 0 != mysqlPort {
				payMysqlTimeOutDel()
			}
		}
	}

	gLog.Trace("server done!")

	gMysqldb.Close()
	return
}
