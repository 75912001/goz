package zutility

import (
	"time"
)

//GenYYYYMMDD 获取yyyymmdd
func GenYYYYMMDD(sec int64) int {
	strYYYYMMDD := time.Unix(sec, 0).Format("20060102")
	return StringToInt(&strYYYYMMDD)
}

//TimeMgr 时间管理器
type TimeMgr struct {
	ApproximateTimeSecond      int64 //近似时间（秒），上一次调用Update更新的时间
	ApproximateTimeMillisecond int64
}

//Update 更新时间管理器中的,当前时间
func (p *TimeMgr) Update() {
	t := time.Now()
	p.ApproximateTimeSecond = t.Unix()
	p.ApproximateTimeMillisecond = t.UnixNano() / 1000000
}
