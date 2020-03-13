package xrLog

import (
	"testing"
)

/*
型号名称：	MacBook Pro
型号标识符：	MacBookPro11,4
处理器名称：	Intel Core i7
处理器速度：	2.2 GHz
处理器数目：	1
核总数：	4
L2 缓存（每个核）：	256 KB
L3 缓存：	6 MB
内存：	16 GB

每行126字节
共125878284byte=>120M
////////////////////////////////////////////////////////////////////////////////
100W 7.678s=>130242/s=>130/ms
125878284 字节=>16394671byte/s=>16010k/s=>15M/s
*/
func TestLog(t *testing.T) {
	var log *Log = new(Log)
	log.Init("test_log")

	for i := 1; i < 1000000; i++ {
		log.Emerg("debug")
	}
}
