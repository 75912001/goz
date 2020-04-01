package xrTimer_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/75912001/goz/xrTimer"
)

var tm xrTimer.TimerMgr
var c chan interface{}

var allCnt int64 = 10000000 //1000w
var iChan chan int64
var eachCnt int64 = 100000 //每次处理多少个打印一次日志 10w

func cb(owner interface{}, data interface{}) int {
	cb_cnt := data.(int64)
	if 0 == cb_cnt%eachCnt {
		fmt.Println(cb_cnt)
	}
	iChan <- cb_cnt
	return 0
}
func addCB(owner interface{}, data interface{}) int {
	n := time.Now()
	second := n.Unix()
	fmt.Println("begin:", second)
	for i := int64(1); i <= allCnt; i++ {
		tm.AddSecond(cb, nil, i, second+i/eachCnt)
	}
	n = time.Now()
	second = n.Unix()
	fmt.Println("end:", second)
	return 0
}
func TestTimerSecond(t *testing.T) {
	second := time.Now().Unix()

	c = make(chan interface{}, 10000)
	iChan = make(chan int64, 10000)

	tm.Run(100, c)

	var outChan <-chan interface{}
	outChan = c
	go func() {
		for v := range outChan {
			switch v.(type) {
			case *xrTimer.TimerSecond:
				tv, ok := v.(*xrTimer.TimerSecond)
				if ok {
					tv.Function(tv.Owner, tv.Data)
				}
			}
		}
	}()

	tm.AddSecond(addCB, nil, 1, second)

	for i := int64(1); i <= allCnt; i++ {
		<-iChan
	}
}
func cb2(owner interface{}, data interface{}) int {
	cb_cnt := data.(int64)

	if 0 == cb_cnt%eachCnt {
		fmt.Println(cb_cnt)
	}
	iChan <- cb_cnt
	return 0
}
func addCB2(owner interface{}, data interface{}) int {
	n := time.Now()
	second := n.Unix()
	millisecond := n.UnixNano() / 1000000
	fmt.Println("begin:", second)
	for i := int64(1); i <= allCnt; i++ {
		tm.AddMillisecond(cb2, nil, i, millisecond+i/eachCnt)
	}
	n = time.Now()
	second = n.Unix()
	fmt.Println("end:", second)
	return 0
}
func TestTimerMillisecond(t *testing.T) {
	n := time.Now()
	millisecond := n.UnixNano() / 1000000

	c = make(chan interface{}, 10000)
	iChan = make(chan int64, 10000)

	tm.Run(100, c)

	tm.AddMillisecond(addCB2, nil, int64(1), millisecond)
	go func() {
		for v := range c {
			switch v.(type) {
			case *xrTimer.TimerMillisecond:
				tv, ok := v.(*xrTimer.TimerMillisecond)
				if ok {
					tv.Function(tv.Owner, tv.Data)

				}
			}
		}
	}()
	for i := int64(1); i <= allCnt; i++ {
		<-iChan
	}
}
