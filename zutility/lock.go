package zutility

import (
	"sync"
)

func Lock() {
	gLock.Lock()
}

func UnLock() {
	gLock.Unlock()
}

////////////////////////////////////////////////////////////////////////////////

//锁定顺序
var gLock sync.Mutex
