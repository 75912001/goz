package xrTimer

import (
	"container/list"
	"math"
	"time"
)

//TimerMgr 定时器管理器
type TimerMgr struct {
	secondVec       [eTimerVecSize]*tvecRoot //秒,数据
	millisecondList *list.List               //毫秒,数据
	//!!!
	timerOutChan chan<- interface{} //超时的*TimerSecond/*TimerMillisecond都会放入其中,添加/删除定时器必须在处理该chan中
}

//OnTimerFun 回调定时器函数(使用协程回调)
type OnTimerFun func(owner interface{}, data interface{}) int

//TimerSecond 秒级定时器
type TimerSecond struct {
	tvecRootIdx int //轮转序号
	TimerMillisecond
}

//TimerMillisecond 毫秒级定时器
type TimerMillisecond struct {
	expire   int64 //过期时间戳
	Owner    interface{}
	Data     interface{}
	Function OnTimerFun //超时调用的函数
	valid    bool       //有效(false:不执行,扫描时自动删除)
}

//Run millisecond:毫秒间隔(如50,则每50毫秒扫描一次毫秒定时器)
//!!!
//ifChan 是超时事件放置的channel,由外部传入
func (p *TimerMgr) Run(millisecond int64, ifChan chan<- interface{}) {
	p.timerOutChan = ifChan

	for idx := range p.secondVec {
		p.secondVec[idx] = new(tvecRoot)
		p.secondVec[idx].init()
		p.secondVec[idx].expire = genExpire(idx)
		gTvecRootExpire[idx] = p.secondVec[idx].expire
		p.secondVec[idx].minExpire = math.MaxInt64
	}

	p.millisecondList = list.New()

	//每秒更新
	//todo 操作共享数据？
	go func() {
		for {
			time.Sleep(time.Second)
			p.scanSecond()
		}
	}()
	//每millisecond个毫秒更新
	//todo 操作共享数据?
	go func() {
		for {
			time.Sleep(time.Duration(millisecond) * time.Millisecond)
			p.scanMillisecond()
		}
	}()
}

//AddSecond 添加秒级定时器
func (p *TimerMgr) AddSecond(cb OnTimerFun, owner interface{}, data interface{}, expire int64) (t *TimerSecond) {
	return p.addSecond(cb, owner, data, expire, nil)
}

//DelSecond 删除秒级定时器
func (p *TimerMgr) DelSecond(t *TimerSecond) {
	t.valid = false
}

//AddMillisecond 添加毫秒级定时器
func (p *TimerMgr) AddMillisecond(cb OnTimerFun, owner interface{}, data interface{}, expireMillisecond int64) (t *TimerMillisecond) {
	t = new(TimerMillisecond)
	t.valid = true
	t.Data = data
	t.expire = expireMillisecond
	t.Function = cb
	t.Owner = owner
	p.millisecondList.PushBack(t)
	return t
}

//DelMillisecond 删除毫秒级定时器
func (p *TimerMgr) DelMillisecond(t *TimerMillisecond) {
	t.valid = false
}

////////////////////////////////////////////////////////////////////////////////
//时间轮数量
const eTimerVecSize int = 9

//每个时间轮到期时间
var gTvecRootExpire [eTimerVecSize]int64

func (p *TimerMgr) addSecond(cb OnTimerFun, owner interface{}, data interface{}, expire int64, oldTimerSecond *TimerSecond) (t *TimerSecond) {
	if nil == oldTimerSecond {
		oldTimerSecond = new(TimerSecond)
		oldTimerSecond.valid = true
	}
	oldTimerSecond.Data = data
	oldTimerSecond.expire = expire
	oldTimerSecond.Function = cb
	oldTimerSecond.Owner = owner
	oldTimerSecond.tvecRootIdx = p.findTvecRootIdx(expire)

	p.secondVec[oldTimerSecond.tvecRootIdx].data.PushBack(oldTimerSecond)

	if expire < p.secondVec[oldTimerSecond.tvecRootIdx].minExpire {
		p.secondVec[oldTimerSecond.tvecRootIdx].minExpire = expire
	}

	return oldTimerSecond
}

//根据到期时间找到时间轮的序号
func (p *TimerMgr) findTvecRootIdx(expire int64) (idx int) {
	var diff = expire - time.Now().Unix()
	for _, v := range p.secondVec {
		if diff <= v.expire {
			return idx
		}
		idx++
	}
	if eTimerVecSize <= idx {
		idx = eTimerVecSize - 1
	}
	return idx
}

//扫描秒级定时器
func (p *TimerMgr) scanSecond() {
	second := time.Now().Unix()

	var next *list.Element
	if p.secondVec[0].minExpire <= second {
		//更新最小过期时间戳
		p.secondVec[0].minExpire = math.MaxInt64
		for e := p.secondVec[0].data.Front(); nil != e; e = next {
			t := e.Value.(*TimerSecond)
			if !t.valid {
				next = e.Next()
				p.secondVec[0].data.Remove(e)
				continue
			}
			if t.expire <= second {
				p.timerOutChan <- t
				next = e.Next()
				p.secondVec[0].data.Remove(e)
			} else {
				if t.expire < p.secondVec[0].minExpire {
					p.secondVec[0].minExpire = t.expire
				}
				next = e.Next()
			}
		}
	}

	//更新时间轮,从序号为1的数组开始
	for idx := 1; idx < eTimerVecSize; idx++ {
		if (p.secondVec[idx].minExpire - second) <= gTvecRootExpire[idx-1] {
			p.secondVec[idx].minExpire = math.MaxInt64
			for e := p.secondVec[idx].data.Front(); e != nil; e = next {
				t := e.Value.(*TimerSecond)
				if !t.valid {
					next = e.Next()
					p.secondVec[idx].data.Remove(e)
					continue
				}
				newIdx := p.findPrevTvecRootIdx(t.expire-second, idx)
				if idx != newIdx {
					next = e.Next()
					p.secondVec[idx].data.Remove(e)
					p.addSecond(t.Function, t.Owner, t.Data, t.expire, t)
				} else {
					if t.expire < p.secondVec[idx].minExpire {
						p.secondVec[idx].minExpire = t.expire
					}
					next = e.Next()
				}
			}
		}
	}
}

//向前查找符合时间差的时间轮序号
func (p *TimerMgr) findPrevTvecRootIdx(diff int64, srcIdx int) (idx int) {
	for {
		if 0 != srcIdx && diff <= p.secondVec[srcIdx-1].expire {
			srcIdx--
		} else {
			break
		}
	}
	return srcIdx
}

//扫描毫秒级定时器
func (p *TimerMgr) scanMillisecond() {
	t := time.Now()
	millisecond := t.UnixNano() / 1000000

	var next *list.Element
	for e := p.millisecondList.Front(); e != nil; e = next {
		t := e.Value.(*TimerMillisecond)
		if !t.valid {
			next = e.Next()
			p.millisecondList.Remove(e)
			continue
		}
		if t.expire <= millisecond {
			p.timerOutChan <- t
			next = e.Next()
			p.millisecondList.Remove(e)
		} else {
			next = e.Next()
		}
	}
}

type tvecRoot struct {
	data      *list.List
	expire    int64 //轮子的到期时间
	minExpire int64 //最小到期时间
}

func (p *tvecRoot) init() {
	p.data = list.New()
}

//4,8,16,32,64,128,256,512,1024...
func genExpire(idx int) (expire int64) {
	expire = 1 << (uint)(idx+2)
	return expire
}
