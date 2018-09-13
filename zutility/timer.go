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
	zutility.GTimerMgr.Run(100)

	zutility.Lock()
	for i := int64(1); i < 100; i++ {
		zutility.GTimerMgr.AddSecond(cb, nil, i, t.ApproximateTimeSecond+i)
	}
	zutility.UnLock()
}
*/
var GTimerMgr timerMgr

//回调定时器函数
type OnTimerFun func(owner interface{}, data interface{}) int

type TimerSecond struct {
	tvecRootIdx int
	element     *list.Element
	expire      int64
	owner       interface{}
	data        interface{}
	function    OnTimerFun
	invalid     bool //无效(true:不执行,扫描时自动删除)
}

////////////////////////////////////////////////////////////////////////////
type TimerMillisecond struct {
	element  *list.Element
	expire   int64
	owner    interface{}
	data     interface{}
	function OnTimerFun
	invalid  bool //无效(true:不执行,扫描时自动删除)
}

//millisecond:毫秒间隔(如50,则每50毫秒扫描一次毫秒定时器)
func (this *timerMgr) Run(millisecond int64) {
	for idx, v := range this.secondVec {
		v = new(tvecRoot)
		v.init()
		v.expire = genExpire(idx)
		v.min_expire = INT64_MAX
		this.secondVec[idx] = v
	}
	this.millisecondList = list.New()

	//每秒更新
	gTimeMgr.Update()
	go func() {
		for {
			time.Sleep(1 * time.Second)

			Lock()
			gTimeMgr.Update()
			this.scanSecond()
			UnLock()
		}
	}()
	//每millisecond个毫秒更新
	go func() {
		for {
			time.Sleep(time.Duration(millisecond) * time.Millisecond)

			Lock()
			gTimeMgr.Update()
			this.scanMillisecond()
			UnLock()
		}
	}()
}

func (this *timerMgr) AddSecond(cb OnTimerFun, owner interface{}, data interface{}, expire int64) (t *TimerSecond) {
	return this.addSecond(cb, owner, data, expire, nil)
}

func (this *timerMgr) DelSecond(t *TimerSecond) {
	t.invalid = true
	//this.secondVec[t.tvecRootIdx].data.Remove(t.element)
}

func (this *timerMgr) AddMillisecond(cb OnTimerFun, owner interface{}, data interface{}, expireMillisecond int64) (t *TimerMillisecond) {
	t = new(TimerMillisecond)
	t.data = data
	t.expire = expireMillisecond
	t.function = cb
	t.owner = owner
	t.element = this.millisecondList.PushBack(t)
	return t
}

func (this *timerMgr) DelMillisecond(t *TimerMillisecond) {
	t.invalid = true
	//this.millisecondList.Remove(t.element)
}

func (this *timerMgr) addSecond(cb OnTimerFun, owner interface{}, data interface{}, expire int64, oldTimerSecond *TimerSecond) (t *TimerSecond) {
	if nil == oldTimerSecond {
		oldTimerSecond = new(TimerSecond)
	}
	oldTimerSecond.data = data
	oldTimerSecond.expire = expire
	oldTimerSecond.function = cb
	oldTimerSecond.owner = owner
	oldTimerSecond.tvecRootIdx = this.findTvecRootIdx(expire)
	oldTimerSecond.element = this.secondVec[oldTimerSecond.tvecRootIdx].data.PushBack(oldTimerSecond)

	if expire < this.secondVec[oldTimerSecond.tvecRootIdx].min_expire {
		this.secondVec[oldTimerSecond.tvecRootIdx].min_expire = expire
	}
	return oldTimerSecond
}

////////////////////////////////////////////////////////////////////////////
//定时器管理器
type timerMgr struct {
	//秒,数据
	secondVec [eTimerVecSize]*tvecRoot
	//毫秒,数据
	millisecondList *list.List
}

//时间轮数量
const eTimerVecSize int = 5

var gTimeMgr TimeMgr

////////////////////////////////////////////////////////////////////////////
//根据到期时间找到时间轮的序号
func (this *timerMgr) findTvecRootIdx(expire int64) (idx int) {
	var diff = expire - gTimeMgr.ApproximateTimeSecond
	for _, v := range this.secondVec {
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
func (this *timerMgr) scanSecond() {
	var next *list.Element
	if this.secondVec[0].min_expire <= gTimeMgr.ApproximateTimeSecond {
		//更新最小过期时间戳
		this.secondVec[0].min_expire = INT64_MAX
		for e := this.secondVec[0].data.Front(); e != nil; e = next {
			t := e.Value.(*TimerSecond)
			if t.invalid {
				next = e.Next()
				this.secondVec[0].data.Remove(e)
				continue
			}
			if t.expire <= gTimeMgr.ApproximateTimeSecond {
				t.function(t.owner, t.data)
				next = e.Next()
				this.secondVec[0].data.Remove(e)
			} else {
				if t.expire < this.secondVec[0].min_expire {
					this.secondVec[0].min_expire = t.expire
				}
				next = e.Next()
			}
		}
	}

	//更新时间轮,从序号为1的数组开始
	for idx := 1; idx < eTimerVecSize; idx++ {
		if (this.secondVec[idx].min_expire - gTimeMgr.ApproximateTimeSecond) < genExpire(idx) {
			this.secondVec[idx].min_expire = INT64_MAX
			for e := this.secondVec[idx].data.Front(); e != nil; e = next {
				t := e.Value.(*TimerSecond)
				if t.invalid {
					next = e.Next()
					this.secondVec[idx].data.Remove(e)
					continue
				}
				new_idx := this.findPrevTvecRootIdx(t.expire-gTimeMgr.ApproximateTimeSecond, idx)
				if idx != new_idx {
					next = e.Next()
					this.secondVec[idx].data.Remove(e)
					this.addSecond(t.function, t.owner, t.data, t.expire, t)
				} else {
					if t.expire < this.secondVec[idx].min_expire {
						this.secondVec[idx].min_expire = t.expire
					}
					next = e.Next()
				}
			}
		}
	}
}

//向前查找符合时间差的时间轮序号
func (this *timerMgr) findPrevTvecRootIdx(diff int64, srcIdx int) (idx int) {
	for {
		if 0 != srcIdx && diff <= this.secondVec[srcIdx-1].expire {
			srcIdx--
		} else {
			break
		}
	}
	return srcIdx
}

//扫描毫秒级定时器
func (this *timerMgr) scanMillisecond() {
	var next *list.Element
	for e := this.millisecondList.Front(); e != nil; e = next {
		t := e.Value.(*TimerMillisecond)
		if t.invalid {
			next = e.Next()
			this.millisecondList.Remove(e)
			continue
		}
		if t.expire <= gTimeMgr.ApproximateTimeMillisecond {
			t.function(t.owner, t.data)
			next = e.Next()
			this.millisecondList.Remove(e)
		} else {
			next = e.Next()
		}
	}
}

////////////////////////////////////////////////////////////////////////////
type tvecRoot struct {
	data *list.List
	//轮子的到期时间
	expire int64
	//最小到期时间
	min_expire int64
}

func (this *tvecRoot) init() {
	this.data = list.New()
}

//4,8,16,32,64,128,256...
func genExpire(idx int) (expire int64) {
	expire = 1 << (uint)(idx+2)
	return expire
}
