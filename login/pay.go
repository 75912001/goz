package main

import (
	//	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"mime/multipart"
	"sort"

	//	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"loginserv_msg"
	"net/url"
	"strconv"
	"strings"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/75912001/goz/zutility"

	"encoding/xml"

	"github.com/75912001/goz/zhttp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
)

var gPayURLCallback string
var gZfbPayURLCallback string
var gWxPayURLCallback string

var gPayHTTPServer zhttp.Server

const gPaySaveOrderID int = 1

//PayCfgResult 支付配置
type PayCfgResult struct {
	PayCfg []PayCfg `xml:"pay"`
}

//PayCfg 支付配置
type PayCfg struct {
	//	Id           uint32 `xml:"id,attr"`
	//	Name         string `xml:"name,attr"`
	Rmb          string `xml:"rmb,attr"`
	Product      string `xml:"product,attr"`
	ProductPuTao string `xml:"productPuTao,attr"`
}

//PayCfgMAP 支付配置
type PayCfgMAP map[string]*PayCfg

func init() {
	GpayCfgMgr.payCfgMAP = make(PayCfgMAP)
}

type payCfgMgr struct {
	payCfgMAP PayCfgMAP
}

//GpayCfgMgr 支付配置管理器
var GpayCfgMgr payCfgMgr

func (pay_cfg_mgr *payCfgMgr) LoadPayCfg() (err error) {
	content, err := ioutil.ReadFile("./pay.xml")

	if nil != err {
		fmt.Println(err)
		return err
	}

	var result PayCfgResult

	err = xml.Unmarshal(content, &result)

	if nil != err {
		fmt.Println(err)
		return err
	}
	for _, v := range result.PayCfg {
		//		fmt.Println(v.Id, v.Name, v.Rmb, v.Product, v.ProductPuTao)
		newPayCfg := new(PayCfg)
		*newPayCfg = v
		if 11 == gPlatform || 12 == gPlatform {
			if 0 == len(v.ProductPuTao) {
				continue
			}
			p := pay_cfg_mgr.Find(v.ProductPuTao)
			if nil != p {
				fmt.Println("error!!!!!!!!pay.xml!!!!!!!!", v.ProductPuTao)
			}
			pay_cfg_mgr.payCfgMAP[v.ProductPuTao] = newPayCfg
		} else {
			if 0 == len(v.Product) {
				continue
			}
			p := pay_cfg_mgr.Find(v.Product)
			if nil != p {
				fmt.Println("error!!!!!!!!pay.xml!!!!!!!!")
			}
			pay_cfg_mgr.payCfgMAP[v.Product] = newPayCfg
		}
	}
	for k, v := range pay_cfg_mgr.payCfgMAP {
		fmt.Println(k)
		fmt.Println(v)
	}

	return nil
}
func (pay_cfg_mgr *payCfgMgr) Find(product string) (payCfg *PayCfg) {
	payCfg, _ = pay_cfg_mgr.payCfgMAP[product]
	return
}

//OnPayGetIDMsg 获取支付ID
func OnPayGetIDMsg(recvProtoHeadBuf []byte, protoMessage *proto.Message, obj interface{}) (ret int) {
	var userInterface UserInterface
	{
		var ok bool
		userInterface, ok = obj.(UserInterface)
		if !ok {
			return -1
		}
	}

	msgIn := (*protoMessage).(*loginserv_msg.PayGetIdMsg)
	platform := msgIn.GetPlatform()
	account := msgIn.GetAccount()
	product := msgIn.GetProduct()
	gLog.Trace(msgIn.String())

	peerConn := userInterface.User.PeerConn
	_, sessionID, _, _, userID := parseProtoHead(recvProtoHeadBuf)
	var payID = genPayID(platform, account, product, userID)

	res := new(loginserv_msg.PayGetIdMsgRes)
	res.PayId = proto.String(payID)
	res.UrlCallback = proto.String(gPayURLCallback)
	res.ZfbUrlCallback = proto.String(gZfbPayURLCallback)
	res.WxUrlCallback = proto.String(gWxPayURLCallback)
	res.Product = proto.String(product)

	//	peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_GET_ID_MSG),
	//		peerConn.ProtoHead.SessionId, peerConn.ProtoHead.UserId, 0)

	send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_GET_ID_MSG), sessionID, userID, 0)

	return
}

func genPayID(platform uint32, account string, product string, uid UserID) (payID string) {
	//生成订单ID
	ts := time.Now().Unix()
	var timeSec = fmt.Sprint(ts)
	var s string
	s += "platform=" + fmt.Sprint(platform)
	s += "account=" + account
	s += "time_sec=" + timeSec
	s += "xiaokanggogogo"
	s += "product=" + product
	s += "uid=" + strconv.FormatUint(uint64(uid), 10)
	s += "nano=" + strconv.FormatInt(time.Now().UnixNano(), 10)

	var md5 = zutility.GenMd5(&s)

	payID = md5

	//存到mysql中
	if 0 != payMysqlInsert(payID, platform, account, uid, 0, uint32(ts), "", "", product) {
		payID = ""
	}

	return
}

////////////////////////////////////////////////////////////////////////////////
//葡萄
////////////////////////////////////////////////////////////////////////////////
//putaoPayRequest 葡萄请求
type putaoPayRequest struct {
	//  参数名称		    类型			含义													允许空值
	Product   string //字符			商品编号													否
	Extra     string //字符			透传参数													是
	TransNo   string //字符			交易流水号(订单号)										否
	Result    string //字符			支付结果，T表示成功，F表示支付失败 C表示交易关闭			否
	NotifyID  string //字符			支付平台的通知流水号										否
	TradeTime int    //时间戳		交易时间戳，单位为秒										否
	Amount    int    //数字			支付金额，单位为分										是
	Currency  string //字符			支付币种													是
	Sign      string //字符			签名														否
	SignType  string //字符			签名方式													否
}

/*
包名：com.topdraw.qjqp.putao

DeveloperID: 5AA38443DDDBD0C44FA36DA476C9C618

PT_PAY_KEY：PT6654BDB6D4CE81A190CA2210599L32

请把每个key的换行符去掉连接成一个字符串使用

Client验证公钥 (clientVerifyKey)

-----BEGIN PUBLIC KEY-----
MIIBIDANBgkqhkiG9w0BAQEFAAOCAQ0AMIIBCAKCAQEAyJQsybLUEylt8MD+c8El
uZEysH8ocZgR3aLPwkL+O7Ce9+f55aQtcRJ8Xujoq4vNUfrgEVLDsp8vwNnkswhZ
PXaW5pLXLupQH+0X8BQiUBQF5q9HGxwi8l//lnqBri1yWhWkM0nWWSj1RPS32tVC
A/xh5lgM+rwAVqmSSqBWDFf8pGAGl24GG5xuJN+PtX1hiYx3mPalH6hI3lK2nVPm
eO5lLiUtagJqkN/XRXZO+FyJLmeZS/SmqObYO1MStEP8u68TGHZYRbpDvQfZn6hG
oSAhSHUitTP4NMYLXY8tR+vATLNzRWhi5UAiIrwVVATywPc8PEeEYSLsMDc/9Zz1
9QIBEQ==
-----END PUBLIC KEY-----



Client签名私钥 (clientSignKey)

-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAox1bAAIq5/N2tuSVb9RnGWNtH4svBRsLpH/cyiUW7hLKDYB3
yIFtS4pTO8W5fvUnunEemRDtmGARspd4KoqML2Gs+P69s7H64I8/RQjF23sspbih
oGS3ZyVJNWmwDQ2N/nktPqlznUhlRRYkBHnRCTWCKi4eBUDyAUOVhBaFbnMOla/U
DVoGCB2b03ppphE45HUCeg2ktRaNvQIIypL5FwaKhIUisGt0d0mjyanRYxA0f7U0
REVMapNSWWtEysLmhYQn0BtlRgYM5hIkAMh9LywPKASQvZm+R4/mqQct+L6uEwEY
Fx8oRvz9fbPXDDr6VXrtbLSDn4R3OILgD+1pCwIBEQKCAQB8vBhpaxHAbuJPn72R
wIsTaiZFTFEiBZ99yzBeWJkQaLig6cUCvVOUHnviagZSBsQHCzWEHAD8DTq1vx+o
D5hgd96gSlTUtUdgbYq8UgC25bi69pm245tO4EcKujtVRpnCmOZdGCs8Ci9S8tA/
qHKssHKYudq4uTGIfvndeqJFZdJca9jgCju2dVfMOXLYp/WZAzN0AA3o8l5ahXe5
HXxBMvhe3msKgtapJQl4n3zwH4df48kzdS0qbKpR8hVDK5NN0fiibpvLZ6jQKfBT
Kt0QMjY/RXn/9GMT8cE8i3F2/4MaSMzF31Qd8Fffixk0vxR50scBF3P8AZdp3fKV
BJqxAoGBANTlXdWDFrf7gV69tqiN93AL9pM3X6zwhmXcpUJO1cxlbsyJx2JveFHm
1N2cc5toEzCPZHejSqpmu08RVL3vrvFVIs+1k30tASz7WKL3mMvE1Y8ff8LrGmrK
khe+Kieo2WZorTQ5iH/oy/h385TWoogw99IHH+bNqUPqtQF96ANrAoGBAMQjxPag
XU67jKLjUT6aUdZy3LXSXyCnoftCmoxiR9j53u9XTCCVqoACCUOjKpKZmLVfOo4i
nHqk+ySOKLnSAlvx//9GHqY2aaCzIDMSlS3fASVnApC+3UgUJCD+Jq0DDU10Q3Wf
wWHA4whD4vnc7Vdd0e/GnVPqxW1NNXn2iTjhAoGBAKLNg/2gesjtgQw2uNs/Yt02
UyVIdlcSSKg/UTK0wZxNkPbDtpaReiCDV3xKdoXmLMrIAYiqC+u3+KXgE6BN4CH1
z1OK2jKL07j8cPUXsRRLOeXq6TqzyOhAjdXrxeIXtU5QC/rClY79UK7yUNs6uIZD
kFVQvgrZgXAr89P23pkzAoGBALiaIsoAV9GhdU4DH0n6p2BsGwV6s+J/p4MRgmYC
JYDrLCyOZcRQoHh6YxJ7c1zMy/X/Ritr3pGMN6nvNWOYerDjw8MU0Y1gY2oSADAR
fVhZan2OIIg7Kp4xEvHgJGaZde6LiskO1B/EmXEw1Z/e/X9nXCz3KqlVboT9X4HY
+Z7xAoGAIMsHrvZdDt/PtfgRCU0Ftev9mxssGihSgom3ubTeBeWUfnaC25fZnv+f
nEW2dZphgVY7H5vkw+RX2gA7N4QAaWaHxXyw02+TkPCSTSLun6vN4SuPgJXhyxvi
LVmk8eJ/wCtTJ+Og1FFHWFsAAaFQxc6q+8XEICcvCogAuWT5S9o=
-----END RSA PRIVATE KEY-----

*/

func putaoRSACheckContent(signContent string, sign string) (ok bool) {
	gLog.Trace(signContent, sign)
	var sPublicKeyPEM string
	sPublicKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIBIDANBgkqhkiG9w0BAQEFAAOCAQ0AMIIBCAKCAQEAyJQsybLUEylt8MD+c8EluZEysH8ocZgR3aLPwkL+O7Ce9+f55aQtcRJ8Xujoq4vNUfrgEVLDsp8vwNnkswhZPXaW5pLXLupQH+0X8BQiUBQF5q9HGxwi8l//lnqBri1yWhWkM0nWWSj1RPS32tVCA/xh5lgM+rwAVqmSSqBWDFf8pGAGl24GG5xuJN+PtX1hiYx3mPalH6hI3lK2nVPmeO5lLiUtagJqkN/XRXZO+FyJLmeZS/SmqObYO1MStEP8u68TGHZYRbpDvQfZn6hGoSAhSHUitTP4NMYLXY8tR+vATLNzRWhi5UAiIrwVVATywPc8PEeEYSLsMDc/9Zz19QIBEQ==
-----END RSA PRIVATE KEY-----
`

	{
		//加载RSA的公钥
		block, _ := pem.Decode([]byte(sPublicKeyPEM))
		//gLog.Trace()
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if nil != err {
			gLog.Error("###### putaoRSACheckContent Failed to parse RSA public key: ", err)
			return false
		}
		//gLog.Trace()
		rsaPub, _ := pub.(*rsa.PublicKey)
		//gLog.Trace()

		////////////////////////////////////////////////////////////////////////////
		//hash := sha256.New()
		//io.WriteString(hash, string(signContent))
		//digest := hash.Sum(nil)

		hash := crypto.Hash.New(crypto.SHA1)
		hash.Write([]byte(signContent))
		hashed := hash.Sum(nil)

		//gLog.Trace()
		////////////////////////////////////////////////////////////////////////////
		// base64解码
		data, err := base64.StdEncoding.DecodeString(sign)
		if nil != err {
			gLog.Error("###### putaoRSACheckContent:", err)
			return
		}
		//data := []byte(sign)

		//gLog.Trace()

		////////////////////////////////////////////////////////////////////////////
		err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA1, hashed, data)
		if nil != err {
			gLog.Error("###### putaoRSACheckContent Verify sig error, reason: ", err)
			return false
		}

		return true
	}
	/*
		{
				//加载RSA的公钥
				block, _ := pem.Decode([]byte(sPublicKeyPEM))
				gLog.Trace()
				pub, err := x509.ParsePKIXPublicKey(block.Bytes)
				if nil != err {
					gLog.Error("###### putaoRSACheckContent Failed to parse RSA public key: ", err)
					return false
				}
				gLog.Trace()
				rsaPub, _ := pub.(*rsa.PublicKey)
				gLog.Trace()
				////////////////////////////////////////////////////////////////////////////
				hash := sha256.New()
				io.WriteString(hash, string(signContent))
				digest := hash.Sum(nil)
				//digest := []byte(signContent)
				gLog.Trace()
				////////////////////////////////////////////////////////////////////////////
				// base64解码
				data, err := base64.StdEncoding.DecodeString(sign)
				if nil != err {
					gLog.Error("###### putaoRSACheckContent:", err)
					return
				}
				//data := []byte(sign)

				gLog.Trace()
				////////////////////////////////////////////////////////////////////////////
				err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, digest, data)
				if nil != err {
					gLog.Error("###### putaoRSACheckContent Verify sig error, reason: ", err)
					return false
				}

				return true

		}
	*/

}

func putaoPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### putaoPayHttpHandler")
			return
		}
	}

	if "GET" != req.Method {
		gLog.Error("###### putaoPayHttpHandler req.Method:", req.Method)
		w.Write([]byte(`err`))
		return
	}
	var payReq putaoPayRequest
	payReq.Product = strings.Join(req.Form["product"], "")
	payReq.Extra = strings.Join(req.Form["extra"], "")
	payReq.TransNo = strings.Join(req.Form["trans_no"], "")
	payReq.Result = strings.Join(req.Form["result"], "")
	payReq.NotifyID = strings.Join(req.Form["notify_id"], "")

	strTradeTime := strings.Join(req.Form["trade_time"], "")
	payReq.TradeTime, _ = strconv.Atoi(strTradeTime)

	strAmount := strings.Join(req.Form["amount"], "")
	payReq.Amount, _ = strconv.Atoi(strAmount)

	payReq.Currency = strings.Join(req.Form["currency"], "")
	payReq.Sign = strings.Join(req.Form["sign"], "")
	payReq.SignType = strings.Join(req.Form["sign_type"], "")

	gLog.Trace(payReq)

	{
		var reqMap map[string]interface{}
		reqMap = make(map[string]interface{}, 0)
		for k, v := range req.Form {
			if "sign" == k {
				continue
			}

			if "sign_type" == k {
				continue
			}

			reqMap[k] = v[0]
		}
		//获取要进行计算哈希的sign string
		signContent := genAlipaySignString(reqMap)
		gLog.Trace(signContent)

		{ //检查签名
			//使用RSA的验签方法
			if !putaoRSACheckContent(signContent, payReq.Sign) {
				gLog.Error("###### putaoPayHttpHandler check")
				w.Write([]byte(`err`))
				return
			}
		}
	}

	if "C" == payReq.Result {
		//删除该订单
		payMysqlDel(payReq.Extra)
		w.Write([]byte(`success`))
		return
	} else if "T" != payReq.Result {
		gLog.Error("###### putaoPayHttpHandler trade_status:", payReq.Result)
		w.Write([]byte(`err`))
		return
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(payReq.Extra)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### putaoPayHttpHandler pay_req.Extra:", payReq.Extra)
			w.Write([]byte(`err`))
			return
		}
		if 2 == pay.OrderStatus || 3 == pay.OrderStatus {
			gLog.Error("###### putaoPayHttpHandler pay.OrderStatus:", pay.OrderStatus)
			w.Write([]byte(`err`))
			return
		}
	}
	pay.OrderStatus = 2
	pay.ConsumeStreamID = payReq.TransNo
	pay.Product = payReq.Product

	var floatAmount = float64(payReq.Amount) / 100.00
	pay.Amount = strconv.FormatFloat(floatAmount, 'f', -1, 64)

	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pay.Amount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("putaoPayHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### putaoPayHttpHandler FIND GATEWAY FAILED")
			w.Write([]byte(`err`))
			return
		}
		peerConn := user.PeerConn

		//		peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`success`))

	gLog.Info("pay:", uid, payReq.Amount)
}

////////////////////////////////////////////////////////////////////////////////
//天津
////////////////////////////////////////////////////////////////////////////////
type tjPayRequest struct {
	Act                 string  //	3	是	表示支付结果通知，默认为100。
	AppID               string  //	64	是	SDK分配给开发者的应用ID。
	ThirdAppID          string  //	64	否	开发者给第三方应用分配的唯一应用ID。
	Uin                 string  //	21	否	发起支付的用户ID。如果用户未登录，此ID为空。
	ConsumeStreamID     string  //	50	是	平台生成的消费流水号。当开发者需要与平台进行详单对账时，需要提供平台的流水号信息用于比对。
	TradeNo             string  //	40	是	应用在调用支付接口时传入的订单号。
	Subject             string  //	100	是	应用在调用支付接口时传入的商品名称。
	Amount              float64 //	N/A	是	应用在调用支付接口时传入的商品价格。
	ChargeAmount        float64 //	N/A	是	保留字段，暂未使用，取值为0。用户需要支付的费用。存在折扣支付场景时（如对于VIP用户），用户需要支付的费用可能与商品定价不同；此字段为折扣后实际需要支付的费用。
	ChargeAmountIncVAT  float64 //	N/A	是	保留字段，暂未使用，取值为0。用户实际支付的费用，包含VAT。在有消费税（VAT）的国家，用户在购买商品时需要额外支付消费税。有的国家显示的商品价格为含税价格，有的国家显示为不含税价格，由支付渠道在从用户账户中扣费时同时扣除消费税。chargeAmountIncVAT/chargeAmountExclVAT分别为用户实际需要支付的含税价和不含税价。对于商品价格即为含税价格的国家，chargeAmountIncVAT=chargeAmount；对于商品价格为不含税价格的国家，chargeAmountExclVAT=chargeAmount。
	ChargeAmountExclVAT float64 //	N/A	是	保留字段，暂未使用，取值为0。用户实际支付的费用，不包含VAT。chargeAmountExclVAT的含义见上。最终对账结算时的收入，应当以chargeAmountExclVAT为准。
	Country             string  //	2	是	用户所属国家，使用ISO 3166-1规范定义的2位字母编码。
	Currency            string  //	3	是	用户支付使用的实际货币，使用ISO 4217规范中定义的3位字母形式编码。用户支付时实际使用的货币可能与购买请求中传入的货币不同（比如：商品以本地货币定价，而用户选择了一些只支持美金的支付渠道时），最终分成结算时将以实际支付货币为准。
	Share               float64 //	N/A	是	保留字段，暂未使用，取值为0。商户按照分成比例得到的收入。实际收入可能因为汇率、坏账等因素变化，应以对账单中的数值为准。
	Note                string  //	32	否	支付时应用传入SDK的透传信息。
	TradeStatus         string  //	3	是	支付结果：completed:支付成功failed:失败canceled:取消支付expired:处理中
	CreateTime          string  //	19	是	创建时间(yyyy-MM-dd HH:mm:ss)
	IsTest              string  //	1	是	是否是测试支付，测试支付产生的消费流水，将不参与最终的分成结算：false：正常支付true：测试支付
	PayChannel          string  //	32	是	支付渠道名称。与华为的结算流程中，收入将按支付渠道汇总，并在结算单中体现。如果需要对账，开发者应当根据此字段汇总各个支付渠道的收入，并按支付渠道与华为对账。
	Sign                string  //	32	是	以上参数的MD5值，其中AppKey为游戏平台分配的应用密钥。String.format("{%s}{%s}{%s}{%s}{%s}{%s}{%s}{%.2f}{%.2f}{%.2f}{%.2f}{%s}{%s}{%s}{%s}{%s}{%.2f}{%s}{%s}",Act,AppId,ThirdAppId,Uin,ConsumeStreamId,TradeNo,Subject,Amount,ChargeAmount,ChargeAmountIncVAT,ChargeAmountExclVAT,Country,Currency,Note,TradeStatus,CreateTime,Share,IsTest, AppKey).HashToMD5Hex()1）注意double型参数的精度；2）HashToMD5Hex()由CP自己定义计算MD5的算法。为。MD5加密算法的参考样例请参考6.1
}

func tjPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### tjPayHttpHandler")
			return
		}
	}

	if "GET" != req.Method {
		gLog.Error("###### tjPayHttpHandler req.Method:", req.Method)
		w.Write([]byte(`{"ErrorCode":"0","ErrorDesc":"GET FAILED"}`))
		return
	}
	var payReq tjPayRequest
	payReq.TradeStatus = strings.Join(req.Form["TradeStatus"], "")
	payReq.Act = strings.Join(req.Form["Act"], "")
	payReq.IsTest = strings.Join(req.Form["IsTest"], "")
	payReq.ConsumeStreamID = strings.Join(req.Form["ConsumeStreamId"], "")
	payReq.PayChannel = strings.Join(req.Form["PayChannel"], "")
	payReq.TradeNo = strings.Join(req.Form["TradeNo"], "")
	payReq.AppID = strings.Join(req.Form["AppId"], "")
	payReq.ThirdAppID = strings.Join(req.Form["ThirdAppId"], "")
	payReq.Uin = strings.Join(req.Form["Uin"], "")
	payReq.Subject = strings.Join(req.Form["Subject"], "")
	strAmount := strings.Join(req.Form["Amount"], "")
	payReq.Amount, _ = strconv.ParseFloat(strAmount, 64)
	payReq.ChargeAmount, _ = strconv.ParseFloat(strings.Join(req.Form["ChargeAmount"], ""), 64)
	payReq.ChargeAmountIncVAT, _ = strconv.ParseFloat(strings.Join(req.Form["ChargeAmountIncVAT"], ""), 64)
	payReq.ChargeAmountExclVAT, _ = strconv.ParseFloat(strings.Join(req.Form["ChargeAmountExclVAT"], ""), 64)
	payReq.Country = strings.Join(req.Form["Country"], "")
	payReq.Currency = strings.Join(req.Form["Currency"], "")
	payReq.Share, _ = strconv.ParseFloat(strings.Join(req.Form["Share"], ""), 64)
	payReq.Note = strings.Join(req.Form["Note"], "")
	payReq.CreateTime = strings.Join(req.Form["CreateTime"], "")
	payReq.Sign = strings.Join(req.Form["Sign"], "")
	gLog.Trace(payReq)

	{ //检查sign
		//AppId = tjlhxkgd
		var AppKey = "hjlhjf"
		//appid：   tjlhqjmj

		var md5String string
		md5String = fmt.Sprintf("{%s}{%s}{%s}{%s}{%s}{%s}{%s}{%.2f}{%.2f}{%.2f}{%.2f}{%s}{%s}{%s}{%s}{%s}{%.2f}{%s}{%s}",
			payReq.Act, payReq.AppID, payReq.ThirdAppID, payReq.Uin,
			payReq.ConsumeStreamID, payReq.TradeNo, payReq.Subject,
			payReq.Amount, payReq.ChargeAmount, payReq.ChargeAmountIncVAT, payReq.ChargeAmountExclVAT,
			payReq.Country, payReq.Currency, payReq.Note, payReq.TradeStatus, payReq.CreateTime,
			payReq.Share,
			payReq.IsTest, AppKey)

		var sign string
		sign = zutility.GenMd5(&md5String)
		{
			if sign != payReq.Sign {
				gLog.Error("###### tjPayHttpHandler sign err[sign:, pay_request.Sign:]", sign, payReq.Sign)
				w.Write([]byte(`{"ErrorCode":"5","ErrorDesc":"SIGN FAILED"}`))
				return
			}
		}
	}
	if "100" != payReq.Act {
		gLog.Error("###### tjPayHttpHandler pay_request.Act:", payReq.Act)
		w.Write([]byte(`{"ErrorCode":"3","ErrorDesc":"ACT FAILED"}`))
		return
	}
	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(payReq.TradeNo)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### tjPayHttpHandler pay_request.TradeNo:", payReq.TradeNo)
			w.Write([]byte(`{"ErrorCode":"4","ErrorDesc":"TradeNo FAILED"}`))
			return
		}
		if 2 == pay.OrderStatus || 3 == pay.OrderStatus {
			gLog.Error("###### tjPayHttpHandler pay.OrderStatus:", pay.OrderStatus)
			w.Write([]byte(`{"ErrorCode":"4","ErrorDesc":"OrderStatus FAILED"}`))
			return
		}

	}

	pay.Amount = strAmount
	{
		//订单状态
		if "failed" == payReq.TradeStatus || "canceled" == payReq.TradeStatus {
			//删除该订单
			payMysqlDel(pay.OrderID)
			w.Write([]byte(`{"ErrorCode":"1","ErrorDesc":"Success"}`))
			return
		} else if "expired" == payReq.TradeStatus {
			w.Write([]byte(`{"ErrorCode":"1","ErrorDesc":"Success"}`))
			return
		} else if "completed" != payReq.TradeStatus {
			w.Write([]byte(`{"ErrorCode":"4","ErrorDesc":"TradeStatus FAILED"}`))
			return
		}
	}
	{
		pay.ConsumeStreamID = payReq.ConsumeStreamID
		pay.PayChannel = payReq.PayChannel
		if "true" == payReq.IsTest {
			pay.OrderStatus = 3
		} else {
			pay.OrderStatus = 2
		}
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(strAmount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("tjPayHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### tjPayHttpHandler FIND GATEWAY FAILED")
			w.Write([]byte(`{"ErrorCode":"0","ErrorDesc":"FIND GATEWAY FAILED"}`))
			return
		}
		peerConn := user.PeerConn

		//		peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`{"ErrorCode":"1","ErrorDesc":"Success"}`))

	gLog.Info("pay:", uid, pay.Amount)
}

//PbsPayJSONData 鹏博士
type PbsPayJSONData struct {
	OrderID          string `json:"orderId"`
	OrderTime        string `json:"orderTime"`
	OrderStatus      string `json:"orderStatus"`
	PayType          string `json:"payType"`
	CashAmt          string `json:"cashAmt"`
	MlAmt            string `json:"mlAmt"`
	Sn               string `json:"sn"`
	Mac              string `json:"mac"`
	Devicecode       string `json:"devicecode"`
	ProductID        string `json:"productId"`
	ProductType      string `json:"productType"`
	ProductName      string `json:"productName"`
	ChargingPrice    string `json:"chargingPrice"`
	ChargingDuration string `json:"chargingDuration"`
	ChargingDegree   string `json:"chargingDegree"`
	UserID           string `json:"userId"`
	UserName         string `json:"userName"`
	PayTime          string `json:"payTime"`
	PayRealml        string `json:"payRealml"`
	PayFreeml        string `json:"payFreeml"`
	OrderAppend      string `json:"orderAppend"`
	Sign             string `json:"sign"`
}

//pbsOrderAppendJson 鹏博士
type pbsOrderAppendJSON struct {
	Callback   string `json:"callback"`
	OutTradeNo string `json:"out_trade_no"`
}

/*
2018/03/12 09:40:03 [trace][/home/meng/work/project_1/trunk/login/pay.go][378][main.pbsPayHttpHandler]&{POST /pbs_pay HTTP/1.1 1 1 map[Accept-Enc
oding:[gzip,deflate] Content-Length:[783] Content-Type:[text/plain; charset=UTF-8] Content-Encoding:[UTF-8] Connection:[Keep-Alive] User-Agent:[A
pache-HttpClient/4.3.6 (java 1.5)] Expect:[100-continue]] 0xc820256620 783 [] false 139.196.55.173:22504 map[] map[] <nil> map[] 124.192.140.229:
6309 /pbs_pay <nil> <nil>}
2018/03/12 09:40:03 [trace][/home/meng/work/project_1/trunk/login/pay.go][427][main.pbsPayHttpHandler]cashAmt=2.00&chargingDegree=1&chargingDurat
ion=-1&chargingId=-1&chargingName=-&chargingPrice=2.00&devicecode=fGIH1NAQP&mac=14:3D:F2:90:94:C8&mlAmt=2.00&orderAppend={"callback":"http://139.
196.55.173:22504/pbs_pay","out_trade_no":"0126edab21ffea6d673bd30a562979eb"}&orderId=100420180312093949005181603&orderStatus=2&orderTime=2018-03-
12 09:39:49&payFreeml=0&payRealml=2.00&payTime=2018-03-12 09:40:02&payType=4&productId=1&productName=全家棋牌：每日首充礼包&productType=1003&sn=D
BD3320J161100947&userId=5181603&userName=-&sign=53a144d64e17e5c82a8dc9d128ff26a0
2018/03/12 09:40:03 [info][/home/meng/work/project_1/trunk/login/pay.go][564][main.pbsPayHttpHandler]pay: 1 2.00
*/
func pbsPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)
	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### pbsPayHttpHandler:", err)
			return
		}
	}
	if "POST" != req.Method {
		gLog.Error("###### pbsPayHttpHandler req.Method:", req.Method)
		return
	}

	result, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	var urlStr = string(result)

	urlStr, err := url.QueryUnescape(urlStr)
	if nil != err {
		gLog.Error("###### pbsPayHttpHandler:", err)
		return
	}
	/*
		cashAmt=0.01
		&chargingDegree=1
		&chargingDuration=-1
		&chargingId=-1
		&chargingName=-
		&chargingPrice=0.01
		&devicecode=1501UA748
		&mac=14:3D:F2:90:8B:6A
		&mlAmt=0.01
		&orderAppend={"callback":"http://139.196.55.173:22504/pbs_pay","out_trade_no":"6befb5e6b497e31090a40fb96ba01bc3"}
		&orderId=100320170707110400005129433
		&orderStatus=2
		&orderTime=2017-07-07 11:04:00
		&payFreeml=0
		&payRealml=0.01
		&payTime=2017-07-07 11:04:14
		&payType=11
		&productId=1
		&productName=全家掼蛋：10点券
		&productType=1003
		&sn=DBD3310G161100749
		&userId=5129433
		&userName=-
		&sign=2ec09168f1000e1cb3c95b701d82fa69
	*/
	gLog.Trace(urlStr)
	str, err := url.ParseQuery(urlStr)
	if nil != err {
		gLog.Error("###### pbsPayHttpHandler:", err)
		return
	}
	/*
		map[chargingName:[-] mlAmt:[0.01] payTime:[2017-07-07 11:04:14] chargingPrice:[0.01] orderId:[100320170707110400005129433] orderStatus:[2] orderTime:[2017-07-07 11:04:00] payRealml:[0.01] chargingDegree:[1] chargingDuration:[-1] chargingId:[-1] payType:[11] sign:[2ec09168f1000e1cb3c95b701d82fa69] productId:[1] productType:[1003] sn:[DBD3310G161100749] userId:[5129433] userName:[-] cashAmt:[0.01] mac:[14:3D:F2:90:8B:6A] orderAppend:[{"callback":"http://139.196.55.173:22504/pbs_pay","out_trade_no":"6befb5e6b497e31090a40fb96ba01bc3"}] devicecode:[1501UA748] payFreeml:[0] productName:[全家掼蛋：10点券]]
	*/
	//		gLog.Trace(str)

	var pbsPayJSONData PbsPayJSONData
	//gLog.Trace(str.Get("orderId"))
	pbsPayJSONData.OrderID = strings.Join(str["orderId"], "")
	pbsPayJSONData.OrderTime = strings.Join(str["orderTime"], "")
	pbsPayJSONData.OrderStatus = strings.Join(str["orderStatus"], "")
	pbsPayJSONData.PayType = strings.Join(str["payType"], "")
	pbsPayJSONData.CashAmt = strings.Join(str["cashAmt"], "")
	pbsPayJSONData.MlAmt = strings.Join(str["mlAmt"], "")
	pbsPayJSONData.Sn = strings.Join(str["sn"], "")
	pbsPayJSONData.Mac = strings.Join(str["mac"], "")
	pbsPayJSONData.Devicecode = strings.Join(str["devicecode"], "")
	pbsPayJSONData.ProductID = strings.Join(str["productId"], "")
	pbsPayJSONData.ProductType = strings.Join(str["productType"], "")
	pbsPayJSONData.ProductName = strings.Join(str["productName"], "")
	pbsPayJSONData.ChargingPrice = strings.Join(str["chargingPrice"], "")
	pbsPayJSONData.ChargingDuration = strings.Join(str["chargingDuration"], "")
	pbsPayJSONData.ChargingDegree = strings.Join(str["chargingDegree"], "")
	pbsPayJSONData.UserID = strings.Join(str["userId"], "")
	pbsPayJSONData.UserName = strings.Join(str["userName"], "")
	pbsPayJSONData.PayTime = strings.Join(str["payTime"], "")
	pbsPayJSONData.PayRealml = strings.Join(str["payRealml"], "")
	pbsPayJSONData.PayFreeml = strings.Join(str["payFreeml"], "")
	pbsPayJSONData.OrderAppend = strings.Join(str["orderAppend"], "")
	pbsPayJSONData.Sign = strings.Join(str["sign"], "")

	{ //检查签名

		/*
			cashAmt=2.00&chargingDegree=1&chargingDurat
			ion=-1&chargingId=-1&chargingName=-&chargingPrice=2.00&devicecode=fGIH1NAQP&mac=14:3D:F2:90:94:C8&mlAmt=2.00&orderAppend={"callback":"http://139.
			196.55.173:22504/pbs_pay","out_trade_no":"0126edab21ffea6d673bd30a562979eb"}&orderId=100420180312093949005181603&orderStatus=2&orderTime=2018-03-
			12 09:39:49&payFreeml=0&payRealml=2.00&payTime=2018-03-12 09:40:02&payType=4&productId=1&productName=全家棋牌：每日首充礼包&productType=1003&sn=D
			BD3320J161100947&userId=5181603&userName=-&sign=53a144d64e17e5c82a8dc9d128ff26a0
		*/
		var md5String string
		md5String = urlStr

		//&sign=2ec09168f1000e1cb3c95b701d82fa69
		idx := strings.Index(md5String, "&sign=")
		if -1 == idx {
			gLog.Error("###### pbsPayHttpHandler")
			return
		}

		//cashAmt=0.01&sign=123456&chargingId=-1
		//str1:cashAmt=0.01
		str1 := md5String[0:idx]
		//str2:sign=123456&chargingId=-1
		str2 := md5String[idx+1:]
		idx2 := strings.Index(str2, "&")
		if -1 == idx2 {
			str2 = ""
		} else {
			str2 = str2[idx2:]
		}

		md5String = str1 + str2

		//Partener ID：p170601142728076
		//Partener KEY：f96b0227825feb0abf0cc3c932a37d6e
		md5String += "&partnerKey=f96b0227825feb0abf0cc3c932a37d6e"
		pbsSign := zutility.GenMd5(&md5String)
		if pbsSign != pbsPayJSONData.Sign {
			gLog.Error("###### pbsPayHttpHandler sign err[pbsSign:, pbsPayJsonData.Sign:]", pbsSign, pbsPayJSONData.Sign)
			return
		}
	}

	var pbsOrderAppendJSON pbsOrderAppendJSON

	err = json.Unmarshal([]byte(pbsPayJSONData.OrderAppend), &pbsOrderAppendJSON)
	if nil != err {
		gLog.Error("###### pbsPayHttpHandler pbsOrderAppendJson err:", pbsOrderAppendJSON, err)
		return
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(pbsOrderAppendJSON.OutTradeNo)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### payHttpHandler pbsOrderAppendJson.Out_trade_no:", pbsOrderAppendJSON.OutTradeNo)
			return
		}
		if 2 == pay.OrderStatus || 3 == pay.OrderStatus {
			gLog.Error("###### payHttpHandler pay.OrderStatus:", pay.OrderStatus)
			return
		}
	}
	pay.Amount = pbsPayJSONData.ChargingPrice
	{
		//订单状态
		if "3" == pbsPayJSONData.OrderStatus || "4" == pbsPayJSONData.OrderStatus {
			//删除该订单
			payMysqlDel(pay.OrderID)
			w.Write([]byte(`success`))
			return
		} else if "2" != pbsPayJSONData.OrderStatus {
			return
		}
		pay.ConsumeStreamID = pbsPayJSONData.OrderID
		pay.PayChannel = pbsPayJSONData.PayType
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pbsPayJSONData.ChargingPrice)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("payHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### payHttpHandler FIND GATEWAY FAILED")
			return
		}
		peerConn := user.PeerConn

		//		peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`success`))
	gLog.Info("pay:", uid, pay.Amount)
}

//http://dev.hismarttv.com/doc/13

//HaiXinPayJSONData 海信
type HaiXinPayJSONData struct {
	//													字段名称 		类型 	必填 	说明
	NotifyTime      string `json:"notify_time"`       //通知时间  		String 	是 		通知时间yyyy-MM-dd HH:mm:ss
	NotifyID        string `json:"notify_id"`         //通知 ID 			string 	是 		通知 ID
	SignType        string `json:"sign_type"`         //签名类型			string 	是 		签名类型，目前暂时只支持 MD5
	Sign            string `json:"sign"`              //签名值			string  是 		签名
	PayPlatform     string `json:"pay_platform"`      //支付平台			string 	是 		支付平台：1：支付宝2：微信 其他保留
	TotalFee        string `json:"total_fee"`         //交易金额			string 	是 		交易金额,单位:元
	OutTradeNo      string `json:"out_trade_no"`      //商户唯一订单号	string 	是 		商户唯一订单号
	PlatformTradeNo string `json:"platform_trade_no"` //支付平台交易号	string 	是 		支付平台的交易号
	TradeStatus     string `json:"trade_status"`      //支付结果			string 	是 		支付结果 TRADE_SUCCESS：成功 其他保留
	PayTime         string `json:"pay_time"`          //买家付款时间		string 	是 		买家付款时间。格式 yyyy-MM-dd HH:mm:ss
	AttachData      string `json:"attach_data"`       //附加数据			string 	否 		json 格式的字符串。根据情况不同内容也会有差别。详细请参考下表
}

/***************************************************************
*函数目的：获得从参数列表拼接而成的待签名字符串
*mapBody：是我们从HTTP request body parse出来的参数的一个map
*返回值：sign是拼接好排序后的待签名字串。
***************************************************************/
func genHaiXinSignString(mapBody map[string]interface{}) (sign string) {
	sortedKeys := make([]string, 0)
	for k := range mapBody {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var signStrings string

	index := 0
	for _, k := range sortedKeys {
		//		gLog.Trace("k=", k, "v =", mapBody[k])
		value := fmt.Sprintf("%v", mapBody[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value
		}
		//最后一项后面不要&
		if index < len(sortedKeys)-1 {
			signStrings = signStrings + "&"
		}
		index++
	}

	return signStrings
}

//md5Key:1BF0D1C348B94919260C91912961C1A2
//appKey:1176172227
//appSecret:wa9aswbnmqtths1963xmv6ijg7383lti
func haiXinPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)
	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### haiXinPayHttpHandler:", err)
			w.Write([]byte(`false`))
			return
		}
	}
	if "POST" != req.Method {
		gLog.Error("###### haiXinPayHttpHandler req.Method:", req.Method)
		w.Write([]byte(`false`))
		return
	}

	gLog.Trace(req.Form)

	var payJSONData HaiXinPayJSONData
	var paymap map[string]interface{}
	paymap = make(map[string]interface{}, 0)

	for k, v := range req.Form {
		if "notify_time" == k {
			payJSONData.NotifyTime = v[0]
		}
		if "notify_id" == k {
			payJSONData.NotifyID = v[0]
		}
		if "pay_platform" == k {
			payJSONData.PayPlatform = v[0]
		}
		if "total_fee" == k {
			payJSONData.TotalFee = v[0]
		}
		if "out_trade_no" == k {
			payJSONData.OutTradeNo = v[0]
		}
		if "platform_trade_no" == k {
			payJSONData.PlatformTradeNo = v[0]
		}
		if "trade_status" == k {
			payJSONData.TradeStatus = v[0]
		}
		if "pay_time" == k {
			payJSONData.PayTime = v[0]
		}
		if "attach_data" == k {
			payJSONData.AttachData = v[0]
		}
		if "sign_type" == k {
			payJSONData.SignType = v[0]
			continue
		}
		if "sign" == k {
			payJSONData.Sign = v[0]
			continue
		}
		if 0 == len(v[0]) {
			continue
		}
		paymap[k] = v[0]
	}

	gLog.Trace(payJSONData)

	//获取要进行计算哈希的sign string
	signContent := genHaiXinSignString(paymap)
	signContent += "1BF0D1C348B94919260C91912961C1A2"
	gLog.Trace(signContent)

	//签名结果统一小写
	sign := zutility.GenMd5(&signContent)
	if sign != payJSONData.Sign {
		gLog.Error("###### haiXinPayHttpHandler sign err[sign:, payJsonData.Sign:]", sign, payJSONData.Sign)
		w.Write([]byte(`false`))
		return
	}

	if "TRADE_SUCCESS" != payJSONData.TradeStatus {
		gLog.Error("###### haiXinPayHttpHandler TradeStatus err")
		w.Write([]byte(`false`))
		return
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(payJSONData.OutTradeNo)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### haiXinPayHttpHandler payJsonData.OutTradeNo:", payJSONData.OutTradeNo)
			w.Write([]byte(`false`))
			return
		}
	}

	pay.Amount = payJSONData.TotalFee
	{
		pay.ConsumeStreamID = payJSONData.PlatformTradeNo
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pay.Amount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("haiXinPayHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### haiXinPayHttpHandler FIND GATEWAY FAILED")
			w.Write([]byte(`false`))
			return
		}
		peerConn := user.PeerConn

		//peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`success`))
	gLog.Info("pay:", uid, pay.Amount)
}

//https://dev.ott.chinanetcenter.com/doc/list?docId=501700#_%E5%8F%91%E8%B4%A7%E9%80%9A%E7%9F%A5%E6%8E%A5%E5%8F%A3

//WangSuPayJSONData 网宿
type WangSuPayJSONData struct {
	//参数名 		参数规格 							必填 	说明
	AppKey          string `json:"appKey"`          //	Y
	SellerOrderCode string `json:"sellerOrderCode"` //	Y 		CP订单号
	PackageName     string `json:"packageName"`     //	Y 		应用包名
	OrderCode       string `json:"orderCode"`       //	Y 		订单号
	Price           string `json:"price"`           //	Y 		支付金额(保留两位有效小数)
	ProdName        string `json:"prodName"`        //	Y 		商品名称
	ProdNum         string `json:"prodNum"`         //	Y 		产品数量
	PayTime         string `json:"payTime"`         //	Y 		支付时间，时间戳(单位ms)
	UID             string `json:"uid"`             //	Y 		订购的用户id
	Status          string `json:"status"`          //	Y 		PAID:已支付
	PayType         string `json:"payType"`         //	N 		付费类型 0表示免费订单 空值或者非0表示正常
	Note            string `json:"note"`            //	N 		订单备注
	Sign            string `json:"sign"`            //	Y 		签名 (对通知参数的签名，可用于对通知参数的有效性校验)
}

/***************************************************************
*函数目的：获得从参数列表拼接而成的待签名字符串
*mapBody：是我们从HTTP request body parse出来的参数的一个map
*返回值：sign是拼接好排序后的待签名字串。
***************************************************************/
func genWangSuSignString(mapBody map[string]interface{}) (sign string) {
	sortedKeys := make([]string, 0)
	for k := range mapBody {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var signStrings string

	index := 0
	for _, k := range sortedKeys {
		//		gLog.Trace("k=", k, "v =", mapBody[k])
		value := fmt.Sprintf("%v", mapBody[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value
		}
		//最后一项后面不要&
		if index < len(sortedKeys)-1 {
			signStrings = signStrings + "&"
		}
		index++
	}

	return signStrings
}

//Appkey: 7642042843
//AppSecret: cFvAZsDZmX
func wangSuPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)
	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### wangSuPayHttpHandler:", err)
			w.Write([]byte(`{"returnCode":0,"returnMsg":"false"}`))
			return
		}
	}
	if "POST" != req.Method {
		gLog.Error("###### wangSuPayHttpHandler req.Method:", req.Method)
		w.Write([]byte(`{"returnCode":0,"returnMsg":"false"}`))
		return
	}

	gLog.Trace(req.Form)

	var wangSuPayJSONData WangSuPayJSONData
	wangSuPayJSONData.PayType = "1"
	var payMap map[string]interface{}
	payMap = make(map[string]interface{}, 0)

	for k, v := range req.Form {
		if "appKey" == k {
			wangSuPayJSONData.AppKey = v[0]
		}
		if "sellerOrderCode" == k {
			wangSuPayJSONData.SellerOrderCode = v[0]
		}
		if "packageName" == k {
			wangSuPayJSONData.PackageName = v[0]
		}
		if "orderCode" == k {
			wangSuPayJSONData.OrderCode = v[0]
		}
		if "price" == k {
			wangSuPayJSONData.Price = v[0]
		}
		if "prodName" == k {
			wangSuPayJSONData.ProdName = v[0]
		}
		if "prodNum" == k {
			wangSuPayJSONData.ProdNum = v[0]
		}
		if "payTime" == k {
			wangSuPayJSONData.PayTime = v[0]
		}
		if "uid" == k {
			wangSuPayJSONData.UID = v[0]
		}
		if "status" == k {
			wangSuPayJSONData.Status = v[0]
		}
		if "payType" == k {
			wangSuPayJSONData.PayType = v[0]
		}
		if "note" == k {
			wangSuPayJSONData.Note = v[0]
		}
		if "sign" == k {
			wangSuPayJSONData.Sign = v[0]
			continue
		}
		payMap[k] = v[0]
	}

	gLog.Trace(wangSuPayJSONData)

	//获取要进行计算哈希的sign string
	signContent := genWangSuSignString(payMap)
	signContent += "cFvAZsDZmX"
	gLog.Trace(signContent)

	//签名结果统一小写
	sign := zutility.GenMd5(&signContent)
	if sign != wangSuPayJSONData.Sign {
		gLog.Error("###### wangSuPayHttpHandler sign err[sign:, wangSuPayJsonData.Sign:]", sign, wangSuPayJSONData.Sign)
		w.Write([]byte(`{"returnCode":-100,"returnMsg":"false"}`))
		return
	}

	if "0" == wangSuPayJSONData.PayType {
		//		gLog.Error("###### wangSuPayHttpHandler free pay err")
		//		w.Write([]byte(`{"returnCode":0,"returnMsg":"false"}`))
		//		return
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(wangSuPayJSONData.SellerOrderCode)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### wangSuPayHttpHandler wangSuPayJsonData.SellerOrderCode:", wangSuPayJSONData.SellerOrderCode)
			w.Write([]byte(`{"returnCode":-103,"returnMsg":"false"}`))
			return
		}
	}
	pay.Amount = wangSuPayJSONData.Price
	{
		pay.ConsumeStreamID = wangSuPayJSONData.OrderCode
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pay.Amount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("wangSuPayHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### wangSuPayHttpHandler FIND GATEWAY FAILED")
			w.Write([]byte(`{"returnCode":0,"returnMsg":"false"}`))
			return
		}
		peerConn := user.PeerConn

		//peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`{"returnCode":1,"returnMsg":"SUCCESS"}`))
	gLog.Info("pay:", uid, pay.Amount)
}

//DangbeiPayJSONData 当贝
type DangbeiPayJSONData struct {
	Mtime      string `json:"mtime"`        //支付时间
	Start      string `json:"start"`        //支付状态
	TotalFee   string `json:"Total_fee"`    //支付金额
	OutTradeNo string `json:"Out_trade_no"` //订单号
	UserNo     string `json:"User_no"`      //商户订单号
	PayUser    string `json:"Pay_user"`     //支付标识
	PayType    string `json:"Pay_type"`     //支付方式
	Extra      string `json:"extra"`        //备用字段
	Pid        string `json:"pid"`          //商品ID
}

func dangbeiPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	//DBkey:07e497a24d10e1f89fe41f84b8a92e5e
	//APPkey:789a8917f314ba6827b10ceb

	gLog.Trace(req)
	contentType := req.Header.Get("Content-Type")

	var boundary string
	{
		str, err := url.ParseQuery(contentType)
		if nil != err {
			gLog.Error("###### dangbeiPayHttpHandler:", err)
			return
		}

		boundary = str.Get(" boundary")
	}
	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### dangbeiPayHttpHandler:", err)
			return
		}
	}
	if "POST" != req.Method {
		gLog.Error("###### dangbeiPayHttpHandler req.Method:", req.Method)
		return
	}

	result, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	gLog.Trace(result)

	var urlStr = string(result)

	urlStr, err := url.QueryUnescape(urlStr)
	if nil != err {
		gLog.Error("###### dangbeiPayHttpHandler:", err)
		return
	}
	/*
		--------------------------1232e1cf6de5d5b5
		Content-Disposition: form-data; name="datastr"

		{
		"mtime":"2017-11-14 17:07:48","start":"success","Total_fee":"0.01",
		"Out_trade_no":"2017111421001104750555179523","User_no":"932ca5b7e6ee71e62e47462d618ca943",
		"Pay_user":"75912001@qq.com","Pay_type":"2","extra":"","notify_url":"http:\/\/139.196.55.173:22514\/dangbei_pay",
		"pid":"test001"
		}
		--------------------------1232e1cf6de5d5b5
		Content-Disposition: form-data; name="sign"

		a7ce7a015c75479ede582dda41eccce4
		--------------------------1232e1cf6de5d5b5--
	*/
	gLog.Trace(urlStr)

	dataMap := make(map[string]string)
	{
		var r io.Reader = strings.NewReader(urlStr)
		multipartReader := multipart.NewReader(r, boundary)
		for {
			part, err := multipartReader.NextPart()
			if io.EOF == err {
				break
			}
			if nil != err {
				gLog.Error("###### dangbeiPayHttpHandler:", err)
				break
			}
			content, err := ioutil.ReadAll(part)
			if nil != err {
				gLog.Error("###### dangbeiPayHttpHandler:", err)
				return
			}
			dataMap[part.FormName()] = string(content)
		}
	}
	gLog.Trace(dataMap)

	datastr, exist := dataMap["datastr"]
	if !exist {
		gLog.Error("###### dangbeiPayHttpHandle:")
		return
	}
	/*
		{"mtime":"2017-11-15 13:59:31","start":"success","Total_fee":"0.01","Out_trade_no":"2017111521001104750555337016","User_no":"24cba632bb0304fbf01d41819b2da0e1","Pay_user":"75912001@qq.com","Pay_type":"2","extra":"","notify_url":"http:\/\/139.196.55.173:22514\/dangbei_pay","pid":"test001"}
	*/

	var payJSONData DangbeiPayJSONData
	{
		var r io.Reader = strings.NewReader(datastr)
		err := json.NewDecoder(r).Decode(&payJSONData)
		if nil != err {
			gLog.Error("###### dangbeiPayHttpHandle:")
			return
		}
	}

	sign, exist := dataMap["sign"]
	if !exist {
		gLog.Error("###### dangbeiPayHttpHandle:")
		return
	}

	gLog.Trace(payJSONData)
	{ //检查签名

		//DBkey:07e497a24d10e1f89fe41f84b8a92e5e
		//APPkey:789a8917f314ba6827b10ceb
		var md5String string

		md5String += payJSONData.OutTradeNo
		md5String += "789a8917f314ba6827b10ceb"
		md5String += payJSONData.PayUser
		md5String += "sign_85445221145"
		dangbeiSign := zutility.GenMd5(&md5String)
		if dangbeiSign != sign {
			gLog.Error("###### dangbeiPayHttpHandler sign err[dangbeiSign:, sign:]", dangbeiSign, sign)
			return
		}
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(payJSONData.UserNo)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### dangbeiPayHttpHandler payJsonData.User_no:", payJSONData.UserNo)
			return
		}
	}
	pay.Amount = payJSONData.TotalFee
	{
		//订单状态
		if "success" != payJSONData.Start {
			gLog.Error("###### dangbeiPayHttpHandler payJsonData.Start:", payJSONData.Start)
			return
		}

		pay.ConsumeStreamID = payJSONData.OutTradeNo
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pay.Amount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("payHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### payHttpHandler FIND GATEWAY FAILED")
			return
		}
		peerConn := user.PeerConn

		//peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`success`))
	gLog.Info("pay:", uid, pay.Amount)
}

/*
参数 				参数名称 		类型 		必填 	描述 													范例
notify_time 		通知时间 		Date 		是 		通知的发送时间。格式为yyyy-MM-dd HH:mm:ss 				2015-14-27 15:45:58
notify_type 		通知类型 		String(64) 	是 		通知的类型 												trade_status_sync
notify_id 			通知校验ID 		String(128) 是 		通知校验ID 												ac05099524730693a8b330c5ecf72da9786
sign_type 			签名类型 		String(10) 	是 		商户生成签名字符串所使用的签名算法类型，目前支持RSA2和RSA，推荐使用RSA2 	RSA2
sign 				签名 			String(256) 是 		请参考异步返回结果的验签 									601510b7970e52cc63db0f44997cf70e
trade_no 			支付宝交易号 	String(64) 	是 		支付宝交易凭证号 										2013112011001004330000121536
app_id 				开发者的app_id 	String(32) 	是 		支付宝分配给开发者的应用Id 								2014072300007148
out_trade_no 		商户订单号 		String(64) 	是 		原支付请求的商户订单号 									6823789339978248
out_biz_no 			商户业务号 		String(64) 	否 		商户业务ID，主要是退款通知中返回退款申请的流水号 	HZRF001
buyer_id 			买家支付宝用户号 String(16) 	否 		买家支付宝账号对应的支付宝唯一用户号。以2088开头的纯16位数字 	2088102122524333
buyer_logon_id 		买家支付宝账号 	String(100) 否 		买家支付宝账号 											15901825620
seller_id 			卖家支付宝用户号 String(30) 	否 		卖家支付宝用户号 										2088101106499364
seller_email 		卖家支付宝账号 	String(100) 否 		卖家支付宝账号 											zhuzhanghu@alitest.com


trade_status 		交易状态 		String(32) 	否 		交易目前所处的状态 										TRADE_CLOSED
					交易状态说明
					枚举名称 	枚举说明
					WAIT_BUYER_PAY 	交易创建，等待买家付款
					TRADE_CLOSED 	未付款交易超时关闭，或支付完成后全额退款
					TRADE_SUCCESS 	交易支付成功
					TRADE_FINISHED 	交易结束，不可退款


total_amount 		订单金额 		Number(9,2) 否 		本次交易支付的订单金额，单位为人民币（元） 				20
receipt_amount 		实收金额 		Number(9,2) 否 		商家在交易中实际收到的款项，单位为元 						15
invoice_amount 		开票金额 		Number(9,2) 否 		用户在交易中支付的可开发票的金额 							10.00
buyer_pay_amount 	付款金额 		Number(9,2) 否 		用户在交易中支付的金额 									13.88
point_amount 		集分宝金额 		Number(9,2) 否 		使用集分宝支付的金额 										12.00
refund_fee 			总退款金额 		Number(9,2) 否 		退款通知中，返回总退款金额，单位为元，支持两位小数 		2.58
send_back_fee 		实际退款金额 	Number(9,2) 否 		商户实际退款给用户的金额，单位为元，支持两位小数 			2.08
subject 			订单标题 		String(256) 否 		商品的标题/交易标题/订单标题/订单关键字等，是请求时对应的参数，原样通知回来 	当面付交易
body 				商品描述 		String(400) 否 		该订单的备注、描述、明细等。对应请求时的body参数，原样通知回来 	当面付交易内容
gmt_create 			交易创建时间 	Date 		否 		该笔交易创建的时间。格式为yyyy-MM-dd HH:mm:ss 			2015-04-27 15:45:57
gmt_payment 		交易付款时间 	Date 		否 		该笔交易的买家付款时间。格式为yyyy-MM-dd HH:mm:ss 		2015-04-27 15:45:57
gmt_refund 			交易退款时间 	Date 		否 		该笔交易的退款时间。格式为yyyy-MM-dd HH:mm:ss.S 			2015-04-28 15:45:57.320
gmt_close 			交易结束时间 	Date 		否 		该笔交易结束时间。格式为yyyy-MM-dd HH:mm:ss 				2015-04-29 15:45:57
fund_bill_list 		支付金额信息 	String(512) 否 		支付成功的各个渠道金额信息，详见资金明细信息说明 			[{"amount":"15.00","fundChannel":"ALIPAYACCOUNT"}]
*/

/***************************************************************
*函数目的：获得从参数列表拼接而成的待签名字符串
*mapBody：是我们从HTTP request body parse出来的参数的一个map
*返回值：sign是拼接好排序后的待签名字串。
***************************************************************/
func genAlipaySignString(mapBody map[string]interface{}) (sign string) {
	sortedKeys := make([]string, 0)
	for k := range mapBody {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var signStrings string

	index := 0
	for _, k := range sortedKeys {
		//		gLog.Trace("k=", k, "v =", mapBody[k])
		value := fmt.Sprintf("%v", mapBody[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value
		}
		//最后一项后面不要&
		if index < len(sortedKeys)-1 {
			signStrings = signStrings + "&"
		}
		index++
	}

	return signStrings
}

/*
//签名方式,默认为RSA2(RSA2048)
'sign_type' => "RSA2",
//支付宝公钥
'alipay_public_key' => "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAg47z9VKnRT20TM8fgyXYaGP3vL301JcQAIk/f5CpRGVxzXX2f75VaBwb8r+sNZYbSg0G/by9EeUZ1Uty/CDOywgZyA0H9WinARtKus7D+EKtv1C8Rynf8JMgBb6vXUB/5GI4fo1YqoSb27QzPL46Lnn3vibnrjctWYIOYjVxIUD/aO6jbOcEW77MfGfa5OKCHBYfDAccn3X4kk8vaahBBn6jqxov5XxDA/+z4z2NHri+Uak9S0o6bW2uZ9xtTTNg7QJdiqG09BAC6cr4qNYxSxqZfsa17x8xI4arxbJbk4QR7GmfPzTq8KnHTvmuxJXIk3fL4mKkTHZoG0pgvWAbHwIDAQAB",
//商户私钥
'merchant_private_key' => "MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC+GU8ffMEZ+5aVWJjcqgktGQxm+2IYdp1SLLEGJFFE+H48yImACIXqGFEJ1Uvu9QgQP9ovA8dOU3kt5Ctj++OHofc8KXcaY3QnBTbhiu/0wBsLg0KOc/1YZlWBC3RQ9zPdDe+NDy4x7m6nmZ8jiWsJT1mGYmXsG6Qdjk78HT7X4AwVWLYrDEICNhuCg1qnrKRMVG8nZr47jcG4FZPAH2j0lVfEiQkpBBrw2WWnfHQlKGVZWocPHYWF0gRnXpnD8NRywYhnKKxX7p52cAYSXPLdV0tQP97ydwLb+fIrIoAZTv5VLE7xQM8cPM0ZgeJkCG7SthZ6fP32YOXGsd2qpV03AgMBAAECggEAUoIvqm3+biWZpTawGk6e7vkJPgVr/Uw2Wj1VlGHc+D+WoxEzROPuI73sJoVykMO/fTYJoBBWyDNIzFdVUe85QVxWL8GblVOHTYxg1qH0JlnfIy8UizniwySfhgQPtzikRRTQXXwyQ6/GTW5K+SSi1YagR8ibjlAs+jsTIzAaX515ZqBU5ki1YIuVuVvr5svIbLppnXmJcdhVSMcq7FRZLzhvvbA/ORj4oZUP8/HmgWTFftn4vvx9sM/Y5tG4cWgyGuYgFSbSdXTuJvY5pHDVKZzh48Z2GqOdu/lP6wtn9mQQteSQFvlbGPm2G4vC/29ihSEjxuTdgBWE3epvNt36sQKBgQD1m3NKOmtdMDc9TcEoOf10JVJ4Klw1bR21xCIwV6AnuNHvohAUZsqsGhmXi485mhi076i+UHyzu7c9Tl72Ofe3SAEoOPHTqlXTRec/SCdIzLmzWH9l9Zq6Q3KCfOxVyEAd8YUd+tB6XcdKcaRipVO5ciGwrhE75zhd5nBKkjn5/wKBgQDGJI/vw4r7eHc2EmaDcDvvMi/aEGOv+F6vU2PKMNJdEuO3bAc1RnvoebrVF8dxgKMSPEAJ/BhJO39fEAwTACl7vsvnluIQA/lB4VdoNGmK+bqIVERSNOMMT0dwgVqBq7aNJCp2fY7igZhNjfAbEjY5snH0x0a4YDg1eovijVfsyQKBgQDDOmO0NyeslWzzX+EQFrhvIFOjjRhqp2ecWmFKx/xYVsMZllrtvJ+RmdWJ7rdUdDb7bB1X2ialv6ryIl+9nWpY1/WDgXBIbfd2zvP4C2Seq41ZEBmEdGwfbwmQy7gYn+rHYnoL0JjzC6QkepzOhNg+aoh5JoQwd6UIjunnfMB1BQKBgQCdMC474F3mhzfTXp+S0DvL032gufXLiPbckgQNR9Pq4Gxke7/wJL1xvPhZyqZ/RbSYZ9HJ2gMOPbQbHyjk/fDq6X7rd4hZej2IZRMpaML97IVtV6RnruscPdyHxSaezjFhIPrKy2rKCFNh2yNK5pS8CvNaY6iX5kVRL6m/ja/d+QKBgQDRolfYwoR7eYQaGKfIX23nun+Dn1u9r2Snv6CGvpiaVBHNJeQ121flotXr4syIguavB0tesXfphDvsj+iuFLVM4ugvbT5JXnF2SBox0Lsp+YFwfK5QWhQp/RrzkWr8/QaVw8Z6DFaiTK4IvQRN0B3EAbs3bQk5xFsxyGs59WfS5Q==",
//编码格式
'charset' => "UTF-8",
//支付宝网关
'gatewayUrl' => "https://openapi.alipay.com/gateway.do",
//应用ID
'app_id' => "2017090608583664",
//异步通知地址,只有扫码支付预下单可用
'notify_url' => "http://www.baidu.com",
//最大查询重试次数
'MaxQueryRetry' => "10",
//查询间隔
'QueryDuration' => "3"
*/

//ZfbRSACheckContent 支付宝
func ZfbRSACheckContent(signContent string, sign string) (ok bool) {
	var sPublicKeyPEM string
	sPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAg47z9VKnRT20TM8fgyXYaGP3vL301JcQAIk/f5CpRGVxzXX2f75VaBwb8r+sNZYbSg0G/by9EeUZ1Uty/CDOywgZyA0H9WinARtKus7D+EKtv1C8Rynf8JMgBb6vXUB/5GI4fo1YqoSb27QzPL46Lnn3vibnrjctWYIOYjVxIUD/aO6jbOcEW77MfGfa5OKCHBYfDAccn3X4kk8vaahBBn6jqxov5XxDA/+z4z2NHri+Uak9S0o6bW2uZ9xtTTNg7QJdiqG09BAC6cr4qNYxSxqZfsa17x8xI4arxbJbk4QR7GmfPzTq8KnHTvmuxJXIk3fL4mKkTHZoG0pgvWAbHwIDAQAB
-----END PUBLIC KEY-----
`

	//加载RSA的公钥
	block, _ := pem.Decode([]byte(sPublicKeyPEM))

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if nil != err {
		gLog.Error("###### zfbPayHttpHandler Failed to parse RSA public key: ", err)
		return false
	}

	rsaPub, _ := pub.(*rsa.PublicKey)

	////////////////////////////////////////////////////////////////////////////
	hash := sha256.New()
	io.WriteString(hash, string(signContent))
	digest := hash.Sum(nil)

	////////////////////////////////////////////////////////////////////////////
	// base64解码
	data, err := base64.StdEncoding.DecodeString(sign)
	if nil != err {
		gLog.Error("###### zfbPayHttpHandler:", err)
		return
	}

	////////////////////////////////////////////////////////////////////////////
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, digest, data)
	if nil != err {
		gLog.Error("###### zfbPayHttpHandler Verify sig error, reason: ", err)
		return false
	}

	return true
}

func zfbPayHTTPSHandler(w http.ResponseWriter, req *http.Request) {
	//https://docs.open.alipay.com/194/103296

	var sign string
	var outTradeNo string
	var tradeStatus string
	var totalAmount string
	var tradeNo string
	var receiptAmount string
	var payMap map[string]interface{}
	payMap = make(map[string]interface{}, 0)

	//	gLog.Trace(req.URL)
	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### zfbPayHttpsHandler:", err)
			return
		}
	}

	if "POST" != req.Method {
		gLog.Error("###### zfbPayHttpsHandler req.Method:", req.Method)
		return
	}

	gLog.Trace(req.Form)
	{
		for k, v := range req.Form {
			if "sign" == k {
				sign = v[0]
				continue
			}

			if "sign_type" == k {
				continue
			}

			if "out_trade_no" == k {
				outTradeNo = v[0]
			}

			if "trade_status" == k {
				tradeStatus = v[0]
			}
			if "total_amount" == k {
				totalAmount = v[0]
			}
			if "trade_no" == k {
				tradeNo = v[0]
			}

			if "receipt_amount" == k {
				receiptAmount = v[0]
			}

			payMap[k] = v[0]
		}
	}

	//获取要进行计算哈希的sign string
	signContent := genAlipaySignString(payMap)
	gLog.Trace(signContent)

	{ //检查签名
		//使用RSA的验签方法，通过签名字符串、签名参数（经过base64解码）及支付宝公钥验证签名。
		if !ZfbRSACheckContent(signContent, sign) {
			gLog.Error("###### zfbPayHttpsHandler check")
			return
		}
	}

	if "TRADE_SUCCESS" != tradeStatus && "TRADE_FINISHED" != tradeStatus {
		gLog.Error("###### zfbPayHttpsHandler trade_status:", tradeStatus)
		return
	}

	var pay ZfbPay
	{
		//该对订单号
		var ret int
		ret, pay = zfbPayMysqlSelectPay(outTradeNo)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### zfbPayHttpHandler out_trade_no:", outTradeNo)
			return
		}
		if 2 == pay.OrderStatus || 3 == pay.OrderStatus {
			gLog.Error("###### zfbPayHttpHandler pay.OrderStatus:", pay.OrderStatus)
			return
		}
	}

	pay.Amount = totalAmount
	pay.ReceiptAmount = receiptAmount

	{
		pay.ConsumeStreamID = tradeNo
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(totalAmount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("zfbPayHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### zfbPayHttpHandler FIND GATEWAY FAILED")
			return
		}
		peerConn := user.PeerConn

		//peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	zfbPayMysqlUpdatePay(&pay)

	gLog.Info("pay:", uid, totalAmount)

	w.Write([]byte(`success`))
}

//WxXML 微信
type WxXML struct {
	Appid         string `xml:"appid"`
	Attach        string `xml:"attach"`
	BankType      string `xml:"bank_type"`
	CashFee       int    `xml:"cash_fee"`
	FeeType       string `xml:"fee_type"`
	IsSubscribe   string `xml:"is_subscribe"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Openid        string `xml:"openid"`
	OutTradeNo    string `xml:"out_trade_no"`
	ResultCode    string `xml:"result_code"`
	ReturnCode    string `xml:"return_code"`
	Sign          string `xml:"sign"`
	TimeEnd       string `xml:"time_end"`
	TotalFee      int    `xml:"total_fee"`
	TradeType     string `xml:"trade_type"`
	TransactionID string `xml:"transaction_id"`
}

func genWxPaySignString(mapBody map[string]interface{}) (sign string) {
	sortedKeys := make([]string, 0)
	for k := range mapBody {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var signStrings string

	index := 0
	for _, k := range sortedKeys {
		//		gLog.Trace("k=", k, "v =", mapBody[k])
		value := fmt.Sprintf("%v", mapBody[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value
		}
		//最后一项后面不要&
		if index < len(sortedKeys)-1 {
			signStrings = signStrings + "&"
		}
		index++
	}

	return signStrings
}

func wxPayHTTPSHandler(w http.ResponseWriter, req *http.Request) {
	//https://pay.weixin.qq.com/wiki/doc/api/native.php?chapter=9_7
	gLog.Trace("====================================")

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### wxPayHttpsHandler:", err)
			return
		}
	}

	if "POST" != req.Method {
		gLog.Error("###### wxPayHttpsHandler req.Method:", req.Method)
		return
	}

	////////////////////////////////////////////////////////////////////////////

	result, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	var urlStr = string(result)
	gLog.Trace(urlStr)

	////////////////////////////////////////////////////////////////////////////
	var payMap map[string]interface{}
	payMap = make(map[string]interface{}, 0)
	var t xml.Token
	urlStr = strings.Replace(urlStr, "\n", "", -1)
	inputReader := strings.NewReader(urlStr)
	decoder := xml.NewDecoder(inputReader)
	var strName string
	var strVal string
	var err error
	for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		case xml.StartElement:
			strName = token.Name.Local
		case xml.CharData:
			strVal = string([]byte(token))
			payMap[strName] = strVal
		default:
		}
	}
	gLog.Trace(payMap)
	////////////////////////////////////////////////////////////////////////////
	if "SUCCESS" != payMap["return_code"] {
		gLog.Error("###### wxPayHttpsHandler return_code:", payMap["return_code"])
		return
	}
	strSign, _ := payMap["sign"]
	outTradeNo, _ := payMap["out_trade_no"]
	totalFee, _ := payMap["total_fee"]
	transactionID, _ := payMap["transaction_id"]
	delete(payMap, "sign")

	var stringA = genWxPaySignString(payMap)

	var KEY = `DGFGeert124445DFGtghgjh678988888`
	var stringSignTemp = stringA + `&`
	stringSignTemp += `key=` + KEY

	{ //检查签名
		wxSign := zutility.GenMd5(&stringSignTemp)
		wxSign = strings.ToUpper(wxSign)
		if wxSign != strSign {
			gLog.Error("###### wxPayHttpHandler sign err[wxSign:, Sign:]", wxSign, strSign)
			w.Write([]byte(`<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[签名失败]]></return_msg></xml>`))
			return
		}
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(outTradeNo.(string))
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### wxPayHttpHandler out_trade_no:", outTradeNo.(string))
			w.Write([]byte(`<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[商户订单号失效]]></return_msg></xml>`))
			return
		}
		if 2 == pay.OrderStatus || 3 == pay.OrderStatus {
			gLog.Error("###### wxPayHttpHandler pay.OrderStatus:", pay.OrderStatus)
			return
		}
	}
	floatTotalFee, _ := strconv.ParseFloat(totalFee.(string), 64)
	pay.Amount = strconv.FormatFloat(floatTotalFee/100.0, 'f', -1, 32)

	{
		pay.ConsumeStreamID = transactionID.(string)
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pay.Amount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("wxPayHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### wxPayHttpHandler FIND GATEWAY FAILED")

			w.Write([]byte(`<xml><return_code><![CDATA[FAIL]]></return_code><return_msg><![CDATA[内部处理错误]]></return_msg></xml>`))
			return
		}
		peerConn := user.PeerConn

		//peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	wxPayMysqlUpdatePay(&pay)

	gLog.Info("pay:", uid, pay.Amount)

	w.Write([]byte(`<xml><return_code><![CDATA[SUCCESS]]></return_code><return_msg><![CDATA[OK]]></return_msg></xml>`))
}

//ShafaPayJSONData 沙发
type ShafaPayJSONData struct {
	OrderID        string `json:"order_id"`        //沙发支付订单号
	PaymentType    int    `json:"payment_type"`    //支付方式//1 支付宝//2 微信
	PaymentAccount string `json:"payment_account"` //支付宝/微信支付帐号
	PaymentID      string `json:"payment_id"`      //支付宝/微信支付订单号
	IsSuccess      bool   `json:"is_success"`      //是否支付成功
	Name           string `json:"name"`            //商品名称
	Price          string `json:"price"`           //单价
	Quantity       int    `json:"quantity"`        //数量
	CustomData     string `json:"custom_data"`     //自定义数据，json格式
	Time           int    `json:"time"`            //时间戳
	Key            string `json:"key"`             //应用app key
	IP             string `json:"ip"`              //订单发起时用户的IP
	Sign           string `json:"sign"`            //签名
}

//ShafaPayJSONDataCustomData 沙发
type ShafaPayJSONDataCustomData struct {
	TradeNo   string `json:"tradeNo"`
	ProductID string `json:"productId"`
}

func genShafaPaySignString(mapBody map[string]interface{}) (sign string) {
	sortedKeys := make([]string, 0)
	for k := range mapBody {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var signStrings string

	index := 0
	for _, k := range sortedKeys {
		//		gLog.Trace("k=", k, "v =", mapBody[k])

		if "is_success" == k {
			signStrings = signStrings + k + "=1"
		} else if "time" == k {
			signStrings = signStrings + k + "=" + strconv.Itoa(int(mapBody[k].(float64)))
		} else {
			value := fmt.Sprintf("%v", mapBody[k])
			signStrings = signStrings + k + "=" + value
		}

		//最后一项后面不要&
		if index < len(sortedKeys)-1 {
			signStrings = signStrings + "&"
		}
		index++
	}

	return signStrings
}

func shafaPayHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### shafaPayHttpHandler:", err)
			return
		}
	}
	if "POST" != req.Method {
		gLog.Error("###### shafaPayHttpHandler req.Method:", req.Method)
		return
	}

	/*
		map[
			data:[
				{
					"order_id":"5a2f45da33875",
					"payment_type":1,
					"payment_id":"2017121221001004750596167561",
					"payment_account":"75912001@qq.com",
					"is_success":true,
					"name":"10\u70b9\u5238",
					"price":"0.01",
					"quantity":1,
					"custom_data":"{\"tradeNo\":\"5c84ad403373ec0803dbddddc77246b1\",\"productId\":\"tjlhxkgddj0o1\"}",
					"time":1513047530,
					"key":"5a20ea76",
					"sign":"c7b83273d1cdd619b96a3107c8c77bbf"
				}
			]
		]
	*/
	//	gLog.Trace(req.Form)
	strData := strings.Join(req.Form["data"], "")
	gLog.Trace(strData)

	//json转map
	var payJSONMap map[string]interface{}
	payJSONMap = make(map[string]interface{}, 0)
	{
		var err error
		if payJSONMap, err = zutility.JSON2map(&strData); err == nil {
			//gLog.Trace(payJsonMap)
		} else {
			gLog.Error("###### shafaPayHttpHandler:", err)
			return
		}
	}

	var payJSONCustomMap map[string]interface{}
	payJSONCustomMap = make(map[string]interface{}, 0)
	{
		//json转map
		req, _ := payJSONMap["custom_data"]
		strJSON := req.(string)
		var err error
		if payJSONCustomMap, err = zutility.JSON2map(&strJSON); err == nil {
			//gLog.Trace(payJsonCustomMap)
		} else {
			gLog.Error("###### shafaPayHttpHandler:", err)
			return
		}
	}

	var payJSONData ShafaPayJSONData
	orderID, ok := payJSONMap["order_id"]
	if ok {
		payJSONData.OrderID = orderID.(string)
	} else {
		gLog.Error("###### shafaPayHttpHandler:")
		return
	}
	paymentType, ok := payJSONMap["payment_type"]
	if ok {
		payJSONData.PaymentType = int(paymentType.(float64))
	}
	paymentAccount, ok := payJSONMap["payment_account"]
	if ok {
		payJSONData.PaymentAccount = paymentAccount.(string)
	}
	paymentID, ok := payJSONMap["payment_id"]
	if ok {
		payJSONData.PaymentID = paymentID.(string)
	}
	isSuccess, ok := payJSONMap["is_success"]
	if ok {
		payJSONData.IsSuccess = isSuccess.(bool)
	} else {
		gLog.Error("###### shafaPayHttpHandler:")
		return
	}
	name, ok := payJSONMap["name"]
	if ok {
		payJSONData.Name = name.(string)
	}
	price, ok := payJSONMap["price"]
	if ok {
		payJSONData.Price = price.(string)
	} else {
		gLog.Error("###### shafaPayHttpHandler:")
		return
	}
	quantity, ok := payJSONMap["quantity"]
	if ok {
		payJSONData.Quantity = int(quantity.(float64))
	}
	customData, ok := payJSONMap["custom_data"]
	if ok {
		payJSONData.CustomData = customData.(string)
	}
	time, ok := payJSONMap["time"]
	if ok {
		payJSONData.Time = int(time.(float64))
	}
	key, ok := payJSONMap["key"]
	if ok {
		payJSONData.Key = key.(string)
	}
	ip, ok := payJSONMap["ip"]
	if ok {
		payJSONData.IP = ip.(string)
	}
	sign, ok := payJSONMap["sign"]
	if ok {
		payJSONData.Sign = sign.(string)
	} else {
		gLog.Error("###### shafaPayHttpHandler:")
		return
	}
	delete(payJSONMap, "sign")

	//gLog.Trace(payJsonData)

	//订单状态
	if !payJSONData.IsSuccess {
		gLog.Error("###### shafaPayHttpHandler payJsonData.Is_success:", payJSONData.IsSuccess)
		return
	}

	var payJSONCustom ShafaPayJSONDataCustomData
	tradeNo, ok := payJSONCustomMap["tradeNo"]
	if ok {
		payJSONCustom.TradeNo = tradeNo.(string)
	} else {
		gLog.Error("###### shafaPayHttpHandler:")
		return
	}

	productID, ok := payJSONCustomMap["productId"]
	if ok {
		payJSONCustom.ProductID = productID.(string)
	} else {
		gLog.Error("###### shafaPayHttpHandler:")
		return
	}

	//gLog.Trace(payJsonCustom)

	{ //检查签名
		var stringA = genShafaPaySignString(payJSONMap)
		var Secret = `6c4f6a45f6ceb7ea250293e505efb87c`
		var stringSignTemp = stringA + Secret

		//gLog.Trace(stringSignTemp)

		shafaSign := zutility.GenMd5(&stringSignTemp)

		if shafaSign != payJSONData.Sign {
			gLog.Error("###### shafaPayHttpHandler sign err[shafaSign:, Sign:]", shafaSign, payJSONData.Sign)
			return
		}
	}

	var pay Pay
	{
		//该对订单号
		var ret int
		ret, pay = payMysqlSelectPay(payJSONCustom.TradeNo)
		if 0 != ret || "" == pay.OrderID {
			gLog.Error("###### shafaPayHttpHandler payJsonCustom.TradeNo:", payJSONCustom.TradeNo)
			return
		}
	}
	pay.Amount = payJSONData.Price
	pay.PayChannel = strconv.Itoa(payJSONData.PaymentType)
	{
		pay.ConsumeStreamID = payJSONData.OrderID
		pay.OrderStatus = 2
	}
	var uid UserID
	{ //通知对应的服务器
		uid = genGatewayID(pay.Platform, pay.Account)

		res := new(loginserv_msg.PaySuccMsgRes)

		res.Product = proto.String(pay.Product)
		res.Amount = proto.String(pay.Amount)

		zutility.Lock()

		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("payHttpHandler find id:", uid)
			//返回失败
			zutility.UnLock()

			gLog.Error("###### payHttpHandler FIND GATEWAY FAILED")
			return
		}
		peerConn := user.PeerConn

		//peerConn.Send(res, ztcp.MESSAGE_ID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, ztcp.USER_ID(pay.UID), 0)

		send(peerConn, res, MessageID(loginserv_msg.CMD_PAY_SUCC_MSG), 0, UserID(pay.UID), 0)

		zutility.UnLock()
	}
	//更新数据库
	payMysqlUpdatePay(&pay)

	w.Write([]byte(`success`))
	gLog.Info("pay:", uid, pay.Amount)
}

////////////////////////////////////////////////////////////////////////////////
type payGetIDJSONPart struct {
	UID     string `json:"uid"`
	Product string `json:"product"`
	Account string `json:"account"` //设备号,和登录(loginJsonPart)的设备号一致才可索引到对应的gateway
}

type payGetIDJSONRes struct {
	PayID          string `json:"pay_id"`            //订单号
	URLCallback    string `json:"url_callback"`      //回调URL
	PengBoShiToken string `json:"peng_bo_shi_token"` //鹏博士token
	ZfbURLCallback string `json:"zfb_url_callback"`  //支付宝回调URL
	WxURLCallback  string `json:"wx_url_callback"`   //微信回调URL
}

func payGetIDHTTPHandler(w http.ResponseWriter, req *http.Request) {
	//	gLog.Trace(req)
	var res payGetIDJSONRes

	res.URLCallback = gPayURLCallback
	res.ZfbURLCallback = gZfbPayURLCallback
	res.WxURLCallback = gWxPayURLCallback
	defer func() {
		js, _ := json.Marshal(res)
		w.Write(js)
		gLog.Trace(res)
	}()

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("payGetIdHttpHandler err")
			return
		}
	}

	if "POST" != req.Method {
		gLog.Error("payGetIdHttpHandler err req.Method:", req.Method)
		return
	}

	result, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	////////////////////////////////////////////////////////////////////////
	//json str 转struct
	var payGetIDJSON payGetIDJSONPart

	err := json.Unmarshal([]byte(result), &payGetIDJSON)
	if nil != err {
		gLog.Error("payGetIdHttpHandler payGetIdJson err:", payGetIDJSON, err)
		return
	}

	if 0 == len(payGetIDJSON.UID) {
		gLog.Error("payGetIdHttpHandler:", payGetIDJSON)
		return
	}
	if 0 == len(payGetIDJSON.Product) {
		gLog.Error("payGetIdHttpHandler:", payGetIDJSON)
		return
	}
	if 0 == len(payGetIDJSON.Account) {
		gLog.Error("payGetIdHttpHandler:", payGetIDJSON)
		return
	}

	//检查product 在配置表中是否合法
	payCfg := GpayCfgMgr.Find(payGetIDJSON.Product)
	if nil == payCfg {
		gLog.Error("payGetIdHttpHandler:", payGetIDJSON)
		return
	}

	gLog.Trace("payGetIdHttpHandler:", payGetIDJSON)

	res.PayID = genPayID(gPlatform, payGetIDJSON.Account, payGetIDJSON.Product, UserID(zutility.StringToUint64(&payGetIDJSON.UID)))

	{
		// gen token
		//Partener ID：p170601142728076
		//Partener KEY：f96b0227825feb0abf0cc3c932a37d6e
		var str = "appendAttr={\"callback\":\""
		str += res.URLCallback
		str += "\",\"out_trade_no\":\""
		str += res.PayID
		str += "\"}&cashAmt="
		str += payCfg.Rmb
		str += "&chargingDuration=-1&partnerId=p170601142728076f96b0227825feb0abf0cc3c932a37d6e"
		res.PengBoShiToken = zutility.GenMd5(&str)
	}

	return
}
