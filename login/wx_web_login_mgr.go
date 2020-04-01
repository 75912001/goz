package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

//晓康 微信登录
//wxf00b7194cb71998e
//f882ac5b09bc7f89134e769e7fa7cc44

//GwxWebLoginLock 锁定顺序
var GwxWebLoginLock sync.Mutex

//WxWebLoginMap 微信登录
type WxWebLoginMap map[string]*WxWebLogin

type wxWebLoginMgr struct {
	wxWebLoginMap WxWebLoginMap
}

//GwxWebLoginMgr 微信登录管理器
var GwxWebLoginMgr wxWebLoginMgr

func init() {
	GwxWebLoginMgr.Init()
}

func (wxWebLoginMgr *wxWebLoginMgr) Init() {
	wxWebLoginMgr.wxWebLoginMap = make(WxWebLoginMap)
}

func (wxWebLoginMgr *wxWebLoginMgr) Add(code string, state string) (wxWebLogin *WxWebLogin) {
	wxWebLogin = new(WxWebLogin)

	wxWebLogin.Code = code
	wxWebLogin.TimeSec = time.Now().Unix()

	wxWebLoginMgr.wxWebLoginMap[state] = wxWebLogin
	return
}

func (wxWebLoginMgr *wxWebLoginMgr) Del(state string) {
	delete(wxWebLoginMgr.wxWebLoginMap, state)
}

func (wxWebLoginMgr *wxWebLoginMgr) Find(state string) (wxWebLogin *WxWebLogin) {
	wxWebLogin, _ = wxWebLoginMgr.wxWebLoginMap[state]
	return
}

//微信回调
func wxWebLoginHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("###### wxWebLoginHttpHandler")
			return
		}
	}

	if "GET" != req.Method {
		gLog.Error("###### wxWebLoginHttpHandler req.Method:", req.Method)
		//		w.Write([]byte(`{"ErrorCode":"0","ErrorDesc":"GET FAILED"}`))
		return
	}
	var code = strings.Join(req.Form["code"], "")
	var state = strings.Join(req.Form["state"], "")

	gLog.Trace(code, state)

	if 0 == len(code) {
		gLog.Error("###### wxWebLoginHttpHandler")
		return
	}
	if 0 == len(state) {
		gLog.Error("###### wxWebLoginHttpHandler")
		return
	}
	//	GwxWebLoginLock.Lock()
	GwxWebLoginMgr.Add(code, state)
	//	GwxWebLoginLock.Unlock()
}

type wxWebLoginJSON struct {
	State string `json:"state"`
}

type wxWebLoginResJSON struct {
	Code      string `json:"code"`
	ErrorCode uint32 `json:"errorcode"` //0:正常,其他失败(2:access_token过期)
}

func getWxWebLoginHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gLog.Trace(req)
	var wx wxWebLoginResJSON

	defer func() {
		js, _ := json.Marshal(wx)
		w.Write(js)

		gLog.Trace(wx)
	}()

	{ //解析参数
		err := req.ParseForm()
		if nil != err {
			gLog.Error("getWxWebLoginHttpHandler err")
			wx.ErrorCode = 1
			return
		}
	}

	if "POST" != req.Method {
		gLog.Error("getWxWebLoginHttpHandler err req.Method:", req.Method)
		wx.ErrorCode = 1
		return
	}

	result, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	////////////////////////////////////////////////////////////////////////
	//json str 转struct

	var loginJSON wxWebLoginJSON

	err := json.Unmarshal([]byte(result), &loginJSON)
	if nil != err {
		gLog.Error("getWxWebLoginHttpHandler loginJson err:", loginJSON, err)
		//返回失败
		wx.ErrorCode = 1
		return
	}

	if 0 == len(loginJSON.State) {
		gLog.Error("getWxWebLoginHttpHandler State:", loginJSON)
		//返回失败
		wx.ErrorCode = 1
		return
	}

	gLog.Trace("getWxWebLoginHttpHandler loginJson:", loginJSON)

	//	GwxWebLoginLock.Lock()

	var wxWebLogin = GwxWebLoginMgr.Find(loginJSON.State)
	if nil == wxWebLogin {
		gLog.Error("getWxWebLoginHttpHandler State:", loginJSON)
		//返回失败
		wx.ErrorCode = 1
		//GwxWebLoginLock.Unlock()
		return
	}

	gLog.Trace("getWxWebLoginHttpHandler wxWebLogin:", wxWebLogin)

	wx.Code = wxWebLogin.Code

	GwxWebLoginMgr.Del(loginJSON.State)

	gLog.Trace("getWxWebLoginHttpHandler GwxWebLoginMgr:", GwxWebLoginMgr)

	//	GwxWebLoginLock.Unlock()
}
