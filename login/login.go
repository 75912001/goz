package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"loginserv_msg"
	"net/http"
	"strings"

	//	"net/url"
	"time"

	"github.com/75912001/goz/zhttp"
	"github.com/75912001/goz/zutility"
	"github.com/golang/protobuf/proto"
)

//gLoginHttpServer
var gLoginHTTPServer zhttp.Server

type loginJSONPart struct {
	Platform uint32 `json:"platform"`
	Account  string `json:"account"` //设备号,和支付(payGetIdJsonPart)的设备号一致才可索引到对应的gateway
	Verify   string `json:"verify"`
	WxCode   string `json:"wxcode"`

	AccessToken string `json:"access_token"`
	Openid      string `json:"openid"`
}

type gatewayJSON struct {
	IP         string `json:"ip"`
	Port       uint32 `json:"port"`
	Session    string `json:"session"`
	ErrorCode  uint32 `json:"errorcode"` //0:正常,其他失败(2:access_token过期)
	NewAccount string `json:"newaccount"`
}

/*
2017/07/06 04:04:36 [trace][/home/meng/work/project_1/trunk/login/login.go][33][main.loginHttpHandler]
&{POST /login HTTP/1.1 1 1 map[
Accept-Encoding:[identity]
Content-Type:[application/json]
User-Agent:[Dalvik/1.6.0 (Linux; U; Android 4.4.4; EC6108V9A_pub_hnylt Build/KTU84Q)]
Connection:[Keep-Alive]
Content-Length:[85]] 0xc820532980 85 [] false 139.196.55.173:22501 map[] map[] <nil> map[] 202.99.114.62:9417 /login <nil> <nil>}

2017/07/06 04:04:36 [trace][/home/meng/work/project_1/trunk/login/login.go][72][main.loginHttpHandler]
loginHttpHandler loginJson: {1 39100000117694 6d29911253ca8f22b6f9e06e222a738b}
*/
func loginHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)
	var gj gatewayJSON

	defer func() {
		js, _ := json.Marshal(gj)
		w.Write(js)

		gLog.Trace(gj)
	}()

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("loginHttpHandler err")
			gj.ErrorCode = 1
			return
		}
	}

	if "POST" != req.Method {
		gLog.Error("loginHttpHandler err req.Method:", req.Method)
		gj.ErrorCode = 1
		return
	}

	result, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	/////////////////////////////////////////
	//	gLog.Trace(req.Body)
	//	gLog.Trace(string(result))
	//	gLog.Trace([]byte(result))

	////////////////////////////////////////////////////////////////////////
	//json str 转struct

	var loginJSON loginJSONPart

	err := json.Unmarshal([]byte(result), &loginJSON)
	if nil != err {
		gLog.Error("loginHttpHandler loginJSON err:", loginJSON, err)
		//返回失败
		gj.ErrorCode = 1
		return
	}

	if 0 == len(loginJSON.Account) {
		gLog.Error("loginHttpHandler Account:", loginJSON)
		//返回失败
		gj.ErrorCode = 1
		return
	}

	gLog.Trace("loginHttpHandler loginJSON:", loginJSON)

	//////////////////////////////////////////////////////////////////////
	//获取wx access_token
	type WxUserJSONData struct {
		Openid     string `json:"openid"`
		Nickname   string `json:"nickname"`
		Sex        uint32 `json:"sex"`
		Headimgurl string `json:"headimgurl"`
		Unionid    string `json:"unionid"`

		Errcode uint32 `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	var wxUserJSONData WxUserJSONData

	var headClient zhttp.Client
	var newAccount string
	newAccount = loginJSON.Account
	gj.NewAccount = newAccount

	//是否微信登录
	var isWxLogin = false
	if 0 != len(loginJSON.AccessToken) {
		isWxLogin = true
	}
	if isWxLogin {
		//https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1419317853&token=bbab5347c533de384dc72126fc98afbe94523309&lang=zh_CN
		type WxLoginJSONData struct {
			AccessToken  string `json:"access_token"`
			ExpiresIn    uint32 `json:"expires_in"`
			RefreshToken string `json:"refresh_token"`
			Openid       string `json:"openid"`
			Scope        string `json:"scope"`

			Errcode uint32 `json:"errcode"`
			Errmsg  string `json:"errmsg"`
		}
		var wxLoginJSONData WxLoginJSONData
		wxLoginJSONData.AccessToken = loginJSON.AccessToken
		wxLoginJSONData.Openid = loginJSON.Openid
		/*
			{ //通过code获取access_token的接口
				//https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
				var client zhttp.Client
				var urlStr string
				urlStr = "https://api.weixin.qq.com/sns/oauth2/access_token?appid="
				urlStr += gWxAppid
				urlStr += "&secret="
				urlStr += gWxSecret
				urlStr += "&code="
				urlStr += loginJson.WxCode
				urlStr += "&grant_type=authorization_code"

				gLog.Trace(urlStr)
				fmt.Println(urlStr)
				if nil == client.Get(urlStr) {
					var resultStr string = string(client.Result)
					gLog.Trace(resultStr)
					fmt.Println(resultStr)

					err := json.Unmarshal([]byte(resultStr), &wxLoginJSONData)
					if nil != err {
						gLog.Error("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						//					fmt.Println("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						gj.ErrorCode = 1
						return
					}
					fmt.Println(wxLoginJSONData)
					if 0 != wxLoginJSONData.Errcode {
						gLog.Error("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						//					fmt.Println("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						gj.ErrorCode = 2
						return
					}
				} else {
					gj.ErrorCode = 1
					return
				}
			}
		*/
		/*
			{ //刷新或续期access_token
				//https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=APPID&grant_type=refresh_token&refresh_token=REFRESH_TOKEN
				var client zhttp.Client
				var urlStr string
				urlStr = "https://api.weixin.qq.com/sns/oauth2/refresh_token?appid="
				urlStr += gWxAppid
				urlStr += "&grant_type=refresh_token&refresh_token="
				urlStr += wxLoginJSONData.Refresh_token

				gLog.Trace(urlStr)
				fmt.Println(urlStr)
				if nil == client.Get(urlStr) {
					var resultStr string = string(client.Result)
					gLog.Trace(resultStr)
					fmt.Println(resultStr)

					err := json.Unmarshal([]byte(resultStr), &wxLoginJSONData)
					if nil != err {
						gLog.Error("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						fmt.Println("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						gj.ErrorCode = 1
						return
					}
					fmt.Println(wxLoginJSONData)
					if 0 != wxLoginJSONData.Errcode {
						gLog.Error("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						fmt.Println("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
						gj.ErrorCode = 2
						return
					}
				} else {
					gj.ErrorCode = 1
					return
				}
			}
		*/

		{ //检验授权凭证（access_token）是否有效
			//https://api.weixin.qq.com/sns/auth?access_token=ACCESS_TOKEN&openid=OPENID
			var client zhttp.Client
			var urlStr string
			urlStr = "https://api.weixin.qq.com/sns/auth?access_token="
			urlStr += wxLoginJSONData.AccessToken
			urlStr += "&openid="
			urlStr += wxLoginJSONData.Openid

			gLog.Trace(urlStr)
			//			fmt.Println(urlStr)
			if nil == client.Get(urlStr) {
				var resultStr = string(client.Result)
				gLog.Trace(resultStr)
				//				fmt.Println(resultStr)

				err := json.Unmarshal([]byte(resultStr), &wxLoginJSONData)
				if nil != err {
					gLog.Error("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
					//					fmt.Println("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
					gj.ErrorCode = 1
					return
				}
				//				fmt.Println(wxLoginJSONData)
				if 0 != wxLoginJSONData.Errcode {
					gLog.Error("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
					//					fmt.Println("###### wxHttpHandler wxLoginJSONData err:", wxLoginJSONData, err)
					gj.ErrorCode = 2
					return
				}
			} else {
				gj.ErrorCode = 1
				return
			}
		}

		{ //获取用户个人信息（UnionID机制）
			//https://api.weixin.qq.com/sns/userinfo?access_token=ACCESS_TOKEN&openid=OPENID
			var client zhttp.Client
			var urlStr string
			urlStr = "https://api.weixin.qq.com/sns/userinfo?access_token="
			urlStr += wxLoginJSONData.AccessToken
			urlStr += "&openid="
			urlStr += wxLoginJSONData.Openid

			gLog.Trace(urlStr)
			//			fmt.Println(urlStr)
			if nil == client.Get(urlStr) {
				var resultStr = string(client.Result)
				gLog.Trace(resultStr)
				//				fmt.Println(resultStr)

				err := json.Unmarshal([]byte(resultStr), &wxUserJSONData)
				if nil != err {
					gLog.Error("###### wxHttpHandler wxUserJSONData err:", wxUserJSONData, err)
					//					fmt.Println("###### wxHttpHandler wxUserJSONData err:", wxUserJSONData, err)
					gj.ErrorCode = 1
					return
				}

				newAccount = wxUserJSONData.Unionid
				gj.NewAccount = newAccount

				//				fmt.Println(wxUserJSONData)
				if 0 != wxUserJSONData.Errcode {
					gLog.Error("###### wxHttpHandler wxUserJSONData err:", wxUserJSONData, err)
					//					fmt.Println("###### wxHttpHandler wxUserJSONData err:", wxUserJSONData, err)
					gj.ErrorCode = 2
					return
				}
			} else {
				gj.ErrorCode = 1
				return
			}
		}
		{ //获取头像数据
			if 0 != len(wxUserJSONData.Headimgurl) {
				//http://wx.qlogo.cn/mmopen/g3MonUZtNHkdmzicIlibx6iaFqAc56vxLSUfpb6n5WKSYVY0ChQKkiaJSgQ1dZuTOgvLLrhJbERQQ4eMsv84eavHiaiceqxibJxCfHe/96
				//var client zhttp.Client
				var urlStr = wxUserJSONData.Headimgurl

				var idx = strings.LastIndex(urlStr, "/")
				urlStr = zutility.StringSubstrRune(&urlStr, idx+1)
				urlStr += gWxHeadSize

				wxUserJSONData.Headimgurl = urlStr
				//			gLog.Trace(urlStr)
				//			fmt.Println(urlStr)
				if nil == headClient.Get(urlStr) {
					//				fmt.Println(headClient.Result)
					//				fmt.Println("111111")
					//				headStr = string(headClient.Result)
					//				gLog.Trace(headStr)
					//				fmt.Println(headStr)
					//				fmt.Println(len(headStr))

					//				var d1 = []byte(headStr)
					//				ioutil.WriteFile("./xx.jpg", d1, 0666) //写入文件(字节数组)

				}
			}
		}
		////////////////////////////////////////////////////////////////////////////
	}
	//////////////////////////////////////////////////////////////////////
	var uid UserID

	var session string
	{ //检查签名
		if 0 == GuserMgr.userCnt {
			gLog.Error("loginHttpHandler err")
			gj.ErrorCode = 1
			return
		}
		if gPlatform != loginJSON.Platform {
			gLog.Error("loginHttpHandler err")
			gj.ErrorCode = 1
			return
		}

		uid = genGatewayID(loginJSON.Platform, newAccount)

		var verifyString string
		verifyString = genVerify(loginJSON.Platform, loginJSON.Account)
		if loginJSON.Verify != verifyString {
			gLog.Error("loginHttpHandler err")
			gj.ErrorCode = 1
			return
		}

		verifyString += fmt.Sprint(time.Now().UnixNano())

		session = zutility.GenMd5(&verifyString)
	}

	{ //通知对应的服务器
		res := new(loginserv_msg.LoginMsgRes)
		res.Platform = proto.Uint32(loginJSON.Platform)
		res.Account = proto.String(newAccount)
		res.Session = proto.String(session)
		res.WxUnionid = proto.String(wxUserJSONData.Unionid)
		res.WxNick = proto.String(wxUserJSONData.Nickname)
		res.WxHeadurl = proto.String(wxUserJSONData.Headimgurl)
		//		res.WxHead = headClient.Result //proto.String(headStr)
		res.WxSex = proto.Uint32(wxUserJSONData.Sex)

		zutility.Lock()
		defer func() {
			zutility.UnLock()
		}()
		user := GuserMgr.FindID(uid)
		if nil == user {
			gLog.Error("loginHttpHandler find id:", uid)
			//返回失败
			gj.ErrorCode = 1
			return
		}
		peerConn := user.PeerConn

		send(peerConn, res, MessageID(loginserv_msg.CMD_LOGIN_MSG), 0, 0, 0)

		//////////////////////////////////////////

		gj.IP = user.IP
		gj.Port = uint32(user.Port)
		gj.Session = session
	}
}

func genVerify(platForm uint32, account string) (verifyString string) {
	var s string
	s += "platform=" + fmt.Sprint(platForm)
	s += "account=" + account
	s += "xiaokanggogogo"

	verifyString = zutility.GenMd5(&s)
	return
}

func genGatewayID(platForm uint32, account string) (uid UserID) {
	var strGatewayKey string
	strGatewayKey += "platform=" + fmt.Sprint(platForm)
	strGatewayKey += "account=" + account
	uid = UserID(zutility.HASHEL(&strGatewayKey))
	uid = uid%UserID(GuserMgr.userCnt) + 1
	return
}
