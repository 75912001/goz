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
	ApproximateTimeSecond      int64 //近似时间（秒），上一次调用Update更新的时间
	ApproximateTimeMillisecond int64
}

func (this *TimeMgr) Update() {
	t := time.Now()
	this.ApproximateTimeSecond = t.Unix()
	this.ApproximateTimeMillisecond = t.UnixNano() / 1000000
}
