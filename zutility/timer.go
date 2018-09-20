package zutility

import (
	"container/list"
	"time"
)

/*
//使用方法
func cb(owner interface{}, data interface{}) int {
	fmt.Println(data.(int64))
	return 0
}

func main() {
	var t zutility.TimeMgr
	t.Update()
	var GTimerMgr zutility.TimerMgr
	GTimerMgr.Run(100)

	zutility.Lock()
	for i := int64(1); i < 100; i++ {
		GTimerMgr.AddSecond(cb, nil, i, t.ApproximateTimeSecond+i)
	}
	zutility.UnLock()
}
*/

//TimerMgr 定时器管理器
type TimerMgr struct {
	secondVec       [eTimerVecSize]*tvecRoot //秒,数据
	millisecondList *list.List               //毫秒,数据
}

//OnTimerFun 回调定时器函数
type OnTimerFun func(owner interface{}, data interface{}) int

//TimerSecond 秒级定时器
type TimerSecond struct {
	tvecRootIdx int   //轮转序号
	expire      int64 //过期时间戳
	owner       interface{}
	data        interface{}
	function    OnTimerFun //超时调用的函数
	invalid     bool       //无效(true:不执行,扫描时自动删除)
}

//TimerMillisecond 毫秒级定时器
type TimerMillisecond struct {
	expire   int64 //过期时间戳
	owner    interface{}
	data     interface{}
	function OnTimerFun //超时调用的函数
	invalid  bool       //无效(true:不执行,扫描时自动删除)
}

//Run millisecond:毫秒间隔(如50,则每50毫秒扫描一次毫秒定时器)
func (p *TimerMgr) Run(millisecond int64) {
	for idx := range p.secondVec {
		p.secondVec[idx] = new(tvecRoot)
		p.secondVec[idx].init()
		p.secondVec[idx].expire = genExpire(idx)
		p.secondVec[idx].minExpire = Int64Max
	}

	p.millisecondList = list.New()

	gTimeMgr.Update()
	//每秒更新
	go func() {
		for {
			time.Sleep(1 * time.Second)

			Lock()
			gTimeMgr.Update()
			p.scanSecond()
			UnLock()
		}
	}()
	//每millisecond个毫秒更新
	go func() {
		for {
			time.Sleep(time.Duration(millisecond) * time.Millisecond)

			Lock()
			gTimeMgr.Update()
			p.scanMillisecond()
			UnLock()
		}
	}()
}

//AddSecond 添加秒级定时器
func (p *TimerMgr) AddSecond(cb OnTimerFun, owner interface{}, data interface{}, expire int64) (t *TimerSecond) {
	return p.addSecond(cb, owner, data, expire, nil)
}

//DelSecond 删除秒级定时器
func (p *TimerMgr) DelSecond(t *TimerSecond) {
	t.invalid = true
}

//AddMillisecond 添加毫秒级定时器
func (p *TimerMgr) AddMillisecond(cb OnTimerFun, owner interface{}, data interface{}, expireMillisecond int64) (t *TimerMillisecond) {
	t = new(TimerMillisecond)
	t.data = data
	t.expire = expireMillisecond
	t.function = cb
	t.owner = owner
	p.millisecondList.PushBack(t)
	return t
}

//DelMillisecond 删除毫秒级定时器
func (p *TimerMgr) DelMillisecond(t *TimerMillisecond) {
	t.invalid = true
}

func (p *TimerMgr) addSecond(cb OnTimerFun, owner interface{}, data interface{}, expire int64, oldTimerSecond *TimerSecond) (t *TimerSecond) {
	if nil == oldTimerSecond {
		oldTimerSecond = new(TimerSecond)
	}
	oldTimerSecond.data = data
	oldTimerSecond.expire = expire
	oldTimerSecond.function = cb
	oldTimerSecond.owner = owner
	oldTimerSecond.tvecRootIdx = p.findTvecRootIdx(expire)
	p.secondVec[oldTimerSecond.tvecRootIdx].data.PushBack(oldTimerSecond)

	if expire < p.secondVec[oldTimerSecond.tvecRootIdx].minExpire {
		p.secondVec[oldTimerSecond.tvecRootIdx].minExpire = expire
	}
	return oldTimerSecond
}

//时间轮数量
const eTimerVecSize int = 9

var gTimeMgr TimeMgr

//根据到期时间找到时间轮的序号
func (p *TimerMgr) findTvecRootIdx(expire int64) (idx int) {
	var diff = expire - gTimeMgr.Second
	for _, v := range p.secondVec {
		if diff <= v.expire {
			break
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
	var next *list.Element
	if p.secondVec[0].minExpire <= gTimeMgr.Second {
		//更新最小过期时间戳
		p.secondVec[0].minExpire = Int64Max
		for e := p.secondVec[0].data.Front(); nil != e; e = next {
			t := e.Value.(*TimerSecond)
			if t.invalid {
				next = e.Next()
				p.secondVec[0].data.Remove(e)
				continue
			}
			if t.expire <= gTimeMgr.Second {
				t.function(t.owner, t.data)
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
		if (p.secondVec[idx].minExpire - gTimeMgr.Second) < genExpire(idx) {
			p.secondVec[idx].minExpire = Int64Max
			for e := p.secondVec[idx].data.Front(); e != nil; e = next {
				t := e.Value.(*TimerSecond)
				if t.invalid {
					next = e.Next()
					p.secondVec[idx].data.Remove(e)
					continue
				}
				newIdx := p.findPrevTvecRootIdx(t.expire-gTimeMgr.Second, idx)
				if idx != newIdx {
					next = e.Next()
					p.secondVec[idx].data.Remove(e)
					p.addSecond(t.function, t.owner, t.data, t.expire, t)
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
	var next *list.Element
	for e := p.millisecondList.Front(); e != nil; e = next {
		t := e.Value.(*TimerMillisecond)
		if t.invalid {
			next = e.Next()
			p.millisecondList.Remove(e)
			continue
		}
		if t.expire <= gTimeMgr.Millisecond {
			t.function(t.owner, t.data)
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

//4,8,16,32,64,128,256,512,1024,2048...
func genExpire(idx int) (expire int64) {
	expire = 1 << (uint)(idx+2)
	return expire
}
