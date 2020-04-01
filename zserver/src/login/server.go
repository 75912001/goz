package login

import (
	"fmt"
	"sync"
	"time"

	"github.com/75912001/goz/zhttp"
	"github.com/75912001/goz/zutility"
)

//锁定顺序
var gServer Server

type Server struct {
	accountMap accountMap
	accMapLock sync.RWMutex
}

var gTimeOutSec int64

func (p *Server) Add(strAccount *string) (session string) {
	acc := new(account)
	session = genSession(*strAccount)
	acc.session = session
	acc.timeOutSec = time.Now().Unix() + gTimeOutSec

	p.accMapLock.Lock()
	defer p.accMapLock.Unlock()
	p.accountMap[*strAccount] = acc
	return
}

func (p *Server) Del(strAccount *string) {
	delete(p.accountMap, *strAccount)
}

func (p *Server) Find(strAccount *string) (acc *account) {
	acc, _ = p.accountMap[*strAccount]
	return
}

func (p *Server) ClearTimeOutAccountMap() {
	nowTimeSec := time.Now().Unix()
	//清理登录的过期session
	p.accMapLock.Lock()
	defer p.accMapLock.Unlock()
	for key, value := range p.accountMap {
		if value.timeOutSec <= nowTimeSec {
			p.Del(&key)
		}
	}
}

func (p *Server) Run(log *zutility.Log, ini *zutility.Ini) {
	//初始化
	p.accountMap = make(map[string]*account)

	go func() {
		for {
			time.Sleep(6 * time.Second)
			p.ClearTimeOutAccountMap()
		}
	}()
	////////////////////////////////////////////////////////
	//httpServer
	httpServerIP := ini.GetString("loginServer", "httpIP", "")
	httpServerPort := ini.GetUint16("loginServer", "httpPORT", 0)
	fmt.Println("httpServerIP, httpServerPort:", httpServerIP, httpServerPort)
	////////////////////////////////////////////////////////
	//启动http
	{
		zhttp.SetLog(log)
		//HTTP 登录服务
		{
			gLoginHTTPServer.AddHandler("/login", loginHTTPHandler)
			go gLoginHTTPServer.Run(httpServerIP, httpServerPort)
		}
	}
	////////////////////////////////////////////////////////
	//rcpTcpServer
	rcpTcpAddr := ini.GetString("loginServer", "rcpTcpAddr", "")

	fmt.Println("rcp tcp addr:", rcpTcpAddr)
	gTimeOutSec = ini.GetInt64("loginServer", "timeOutSec", 30)

	//启动rpc
	var rpc IFRPC
	rpc = new(RPCX)
	rpc.create(&rcpTcpAddr)
}

func genSession(account string) (session string) {
	var s string
	s += "account=" + account
	s += "xiaokanggogogo"
	s += fmt.Sprint(time.Now().UnixNano())

	session = zutility.GenMd5(&s)
	return
}
