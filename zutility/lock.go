package zutility

import "sync"

//Lock 锁
func Lock() {
	gLock.Lock()
}

//UnLock 解锁
func UnLock() {
	gLock.Unlock()
}

func RLock() {
	gLock.RLock()
}

func RUnLock() {
	gLock.RUnlock()
}

////////////////////////////////////////////////////////////////////////////////

//锁定顺序
var gLock sync.RWMutex
