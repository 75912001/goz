package xrRingBuf_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/75912001/goz/xrRingBuf"
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

buf 大小:1024字节 1k
每次1字节
w/r:1124w/s	11242/ms

每次10字节
w/r:1080w/s	10805/ms

每次100字节
w/r:951w/s	9519/ms

////////////////////////////////////////////////////////////////////////////////
buf 大小:1048576字节 1024k
每次1字节
w/r:1231w/s	12314/ms

每次100字节
w/r:987w/s	9877/ms

每次1000字节
w/r:802w/s	8025/ms
*/

const bufLen uint32 = 1048576
const oneCnt uint32 = 1

var gStr string

var wcnt uint64
var rcnt uint64

func TestFun(t *testing.T) {
	var bf xrRingBuf.RingBuf
	bf.Create(bufLen)

	gStr += "0123456789"

	go func(p *xrRingBuf.RingBuf) {
		str := gStr
		w := []byte(str)

		for {
			var lLen uint32 = uint32(len(gStr))
			for 0 != lLen {
				wLen := p.Write(w[uint32(len(gStr))-lLen:], oneCnt)
				lLen -= wLen
			}
			wcnt++
		}
	}(&bf)

	go func(p *xrRingBuf.RingBuf) {
		var r []byte
		r = make([]byte, uint32(len(gStr)))
		for {
			var lLen uint32 = uint32(len(gStr))
			for 0 != lLen {
				rLen := p.Read(r[uint32(len(gStr))-lLen:], oneCnt)
				lLen -= rLen
			}
			rcnt++
			//time.Sleep(time.Second)
			//fmt.Println(r)
			var str string
			str = string(r[:])
			if gStr != str {
				fmt.Println("error!!!!!!")
			}
		}
	}(&bf)
	go func() {
		var scnt uint64
		for {
			time.Sleep(time.Second)
			scnt++
			fmt.Println(wcnt, rcnt, wcnt/scnt, rcnt/scnt)
		}

	}()
	time.Sleep(100 * time.Second)
}
