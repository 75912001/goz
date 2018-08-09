package zutility

//使用系统log,自带锁
//使用协程操作io输出日志
//目前性能:20W行/s,20M/s
//每天自动创建新的日志文件

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	LOG_CHAN_MAX_CNT = 1000 //日志channel的最大数量
)

//日志等级
const (
	LEVEL_OFF int = iota //关闭
	LEVEL_EMERG
	LEVEL_CRIT
	LEVEL_ERROR
	LEVEL_WARNING
	LEVEL_NOTICE
	LEVEL_INFO
	LEVEL_DEBUG
	LEVEL_TRACE
	LEVEL_ON //全部打开
)

type Log struct {
	level       int      //日志等级
	file        *os.File //日志文件
	logger      *log.Logger
	logChan     chan string
	yyyymmdd    int    //日志年月日
	name_prefix string //日志文件名称前缀
}

//初始化
func (this *Log) Init(name string) (err error) {
	this.level = LEVEL_ON
	this.name_prefix = name
	this.yyyymmdd = GenYYYYMMDD(time.Now().Unix())

	log_name := this.name_prefix + IntToString(this.yyyymmdd)
	this.file, err = os.OpenFile(log_name, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if nil != err {
		return
	}
	this.logger = log.New(this.file, "", log.Ldate|log.Ltime) //|log.Llongfile)

	this.logChan = make(chan string, LOG_CHAN_MAX_CNT)
	go this.onOutPut()
	return
}

//设置日志等级
func (this *Log) SetLevel(level int) {
	this.level = level
}

//反初始化
func (this *Log) DeInit() {
	this.file.Close()
}

/////////////////////////////////////////////////////////////////////////////
//日志方法
//踪迹日志
func (this *Log) Trace(v ...interface{}) {
	if this.level < LEVEL_TRACE {
		return
	}
	this.outPut(2, "trace", fmt.Sprintln(v...))
}

//调试日志
func (this *Log) Debug(v ...interface{}) {
	if this.level < LEVEL_DEBUG {
		return
	}
	this.outPut(2, "debug", fmt.Sprintln(v...))
}

//报告日志
func (this *Log) Info(v ...interface{}) {
	if this.level < LEVEL_INFO {
		return
	}
	this.outPut(2, "info", fmt.Sprintln(v...))
}

//公告日志
func (this *Log) Notice(v ...interface{}) {
	if this.level < LEVEL_NOTICE {
		return
	}
	this.outPut(2, "notice", fmt.Sprintln(v...))
}

//警告日志
func (this *Log) Warning(v ...interface{}) {
	if this.level < LEVEL_WARNING {
		return
	}
	this.outPut(2, "warning", fmt.Sprintln(v...))
}

//错误日志
func (this *Log) Error(v ...interface{}) {
	if this.level < LEVEL_ERROR {
		return
	}
	this.outPut(2, "error", fmt.Sprintln(v...))
}

//临界日志
func (this *Log) Crit(v ...interface{}) {
	if this.level < LEVEL_CRIT {
		return
	}
	this.outPut(2, "crit", fmt.Sprintln(v...))
}

//不可用日志
func (this *Log) Emerg(v ...interface{}) {
	if this.level < LEVEL_EMERG {
		return
	}
	this.outPut(2, "emerg", fmt.Sprintln(v...))
}

////////////////////////////////////////////////////////////////////////////////
//写日志
func (this *Log) onOutPut() {
	for {
		now_yyyymmdd := GenYYYYMMDD(time.Now().Unix())
		if this.yyyymmdd != now_yyyymmdd {
			this.file.Close()

			this.yyyymmdd = now_yyyymmdd
			log_name := this.name_prefix + IntToString(this.yyyymmdd)
			this.file, _ = os.OpenFile(log_name, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
			this.logger = log.New(this.file, "", log.Ldate|log.Ltime) //|log.Llongfile)
		}

		this.logger.Print(<-this.logChan)
	}
}

//路径,文件名,行数,函数名称
func (this *Log) outPut(calldepth int, prefix string, str string) {
	pc, file, line, ok := runtime.Caller(calldepth)
	if true != ok {
		return
	}
	funName := runtime.FuncForPC(pc).Name()

	var strLine string = strconv.Itoa(line)
	this.logChan <- "[" + prefix + "][" + file + "][" + strLine + "][" + funName + "]" + str
}
