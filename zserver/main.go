package main

import (
	"fmt"

	"math/rand"
	"runtime"

	"time"

	"github.com/75912001/goz/zserver/src/login"
	"github.com/75912001/goz/zutility"
)

var gLog *zutility.Log
var gIni *zutility.Ini
var gPlatform uint32
var gServerName string

//1:login, 2:gateway, 3:dbproxy, 4:db
var gServerType int

func main() {
	///////////////////////////////////////////////////////////////////
	//时间戳
	rand.Seed(time.Now().Unix())
	fmt.Println(time.Now().Unix(), time.Now().UnixNano())

	yyyymm := time.Now().Format("200601")
	fmt.Println(yyyymm)

	fmt.Println("OS:", zutility.ShowOS())

	///////////////////////////////////////////////////////////////////
	//加载配置文件bench.ini
	{
		gIni = new(zutility.Ini)
		err := gIni.Load("./bench.ini")
		if nil != err {
			fmt.Println("load bench.ini err!")
			return
		}
	}

	//common
	gPlatform = gIni.GetUint32("common", "platform", 0)
	goProcessMax := gIni.GetInt("common", "goProcessMax", runtime.NumCPU())
	runtime.GOMAXPROCS(goProcessMax)

	//server
	gServerType = gIni.GetInt("server", "type", 0)
	gServerName = gIni.GetString("server", "name", "serverName.default")

	//log
	logLevel := gIni.GetInt("log", "logLevel", 0)
	logPath := gIni.GetString("log", "path", "log.default.") + gServerName + "."

	//启动日志
	gLog = new(zutility.Log)
	err := gLog.Init(logPath, 1000)
	if nil != err {
		fmt.Println("log err:", err)
		return
	}
	gLog.SetLevel(logLevel)
	defer gLog.DeInit()

	fmt.Println("serverType:", gServerType)
	switch gServerType {
	case 1: //login
		var loginServer login.Server
		go loginServer.Run(gLog, gIni)
	case 2: //gateway
	case 3: //dbproxy
	case 4: //db
	}

	////////////////////////////////////////////////////////////////////////////

	gLog.Trace("server runing...", time.Now())

	for {
		time.Sleep(60 * time.Second)
		gLog.Trace("server run...")
	}
}
