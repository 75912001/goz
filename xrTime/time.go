package xrTime

import (
	"time"

	"github.com/75912001/goz/xrUtility"
)

//GenYYYYMMDD 获取yyyymmdd
func GenYYYYMMDD(sec int64) int {
	strYYYYMMDD := time.Unix(sec, 0).Format("20060102")
	v, _ := xrUtility.StringToInt(&strYYYYMMDD)
	return v
}

//TimeMgr 时间管理器
type TimeMgr struct {
	Second      int64 //近似时间（秒），上一次调用Update更新的时间
	Millisecond int64 //近似时间（毫秒），上一次调用Update更新的时间
}

//Update 更新时间管理器中的,当前时间
func (p *TimeMgr) Update() {
	t := time.Now()
	p.Second = t.Unix()
	p.Millisecond = t.UnixNano() / 1000000
}
