package ztcp

import (
	"github.com/goz/zutility"
)

func SetLog(v *zutility.Log) {
	gLog = v
}

////////////////////////////////////////////////////////////////////////////////
var gLog *zutility.Log

//锁定顺序
//var gLock sync.Mutex

//只用在非ztcp&zudp的回调函数中，否则死锁！
//func Lock() {
//	gLock.Lock()
//}

//func UnLock() {
//	gLock.Unlock()
//}
