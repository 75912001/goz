package zutility

import (
	"time"
)

func GenYYYYMMDD(sec int64) int {
	str_yyyymmdd := time.Unix(sec, 0).Format("20060102")
	return StringToInt(&str_yyyymmdd)
}

////////////////////////////////////////////////////////////////////////////
type TimeMgr struct {
	ApproximateTimeSecond int64 //近似时间（秒），上一次调用Update更新的时间
}

func (this *TimeMgr) Update() {
	this.ApproximateTimeSecond = time.Now().Unix()
}

//定时器,秒
func Second(value uint32, f func()) *time.Timer {
	v := time.Duration(value)
	return time.AfterFunc(v*time.Second, f)
}
