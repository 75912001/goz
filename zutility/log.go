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

//日志等级
const (
	levelOff     int = 0 //关闭
	levelEmerg   int = 1
	levelCrit    int = 2
	levelError   int = 3
	levelWarning int = 4
	levelNotice  int = 5
	levelInfo    int = 6
	levelDebug   int = 7
	levelTrace   int = 8
	levelOn      int = 9 //9 全部打开
)

//Log 日志
type Log struct {
	level      int      //日志等级
	file       *os.File //日志文件
	logger     *log.Logger
	logChan    chan string
	yyyymmdd   int    //日志年月日
	namePrefix string //日志文件名称前缀
	perm       os.FileMode
	fileFlag   int
	logFlag    int
}

//Init 初始化 logChanMaxCnt:日志channel的最大数量
func (p *Log) Init(name string, logChanMaxCnt uint32) (err error) {
	p.perm = os.ModePerm
	p.fileFlag = os.O_CREATE | os.O_APPEND | os.O_RDWR
	p.logFlag = log.Ltime //log.Ldate|log.Llongfile
	p.level = levelOn
	p.namePrefix = name
	p.yyyymmdd = GenYYYYMMDD(time.Now().Unix())

	logName := p.namePrefix + IntToString(p.yyyymmdd)
	p.file, err = os.OpenFile(logName, p.fileFlag, p.perm)
	if nil != err {
		return err
	}
	p.logger = log.New(p.file, "", p.logFlag)

	p.logChan = make(chan string, logChanMaxCnt)
	go p.onOutPut()
	return err
}

//SetLevel 设置日志等级
func (p *Log) SetLevel(level int) {
	p.level = level
}

//DeInit 反初始化
func (p *Log) DeInit() {
	p.file.Close()
}

//Trace 踪迹日志
func (p *Log) Trace(v ...interface{}) {
	if p.level < levelTrace {
		return
	}
	p.outPut(2, "trace", fmt.Sprintln(v...))
}

//Debug 调试日志
func (p *Log) Debug(v ...interface{}) {
	if p.level < levelDebug {
		return
	}
	p.outPut(2, "debug", fmt.Sprintln(v...))
}

//Info 报告日志
func (p *Log) Info(v ...interface{}) {
	if p.level < levelInfo {
		return
	}
	p.outPut(2, "info", fmt.Sprintln(v...))
}

//Notice 公告日志
func (p *Log) Notice(v ...interface{}) {
	if p.level < levelNotice {
		return
	}
	p.outPut(2, "notice", fmt.Sprintln(v...))
}

//Warning 警告日志
func (p *Log) Warning(v ...interface{}) {
	if p.level < levelWarning {
		return
	}
	p.outPut(2, "warning", fmt.Sprintln(v...))
}

//Error 错误日志
func (p *Log) Error(v ...interface{}) {
	if p.level < levelError {
		return
	}
	p.outPut(2, "error", fmt.Sprintln(v...))
}

//Crit 临界日志
func (p *Log) Crit(v ...interface{}) {
	if p.level < levelCrit {
		return
	}
	p.outPut(2, "crit", fmt.Sprintln(v...))
}

//Emerg 不可用日志
func (p *Log) Emerg(v ...interface{}) {
	if p.level < levelEmerg {
		return
	}
	p.outPut(2, "emerg", fmt.Sprintln(v...))
}

////////////////////////////////////////////////////////////////////////////////
//写日志
func (p *Log) onOutPut() {
	for {
		nowYYYYMMDD := GenYYYYMMDD(time.Now().Unix())
		if p.yyyymmdd != nowYYYYMMDD {
			p.file.Close()

			p.yyyymmdd = nowYYYYMMDD
			logName := p.namePrefix + IntToString(p.yyyymmdd)
			p.file, _ = os.OpenFile(logName, p.fileFlag, p.perm)
			p.logger = log.New(p.file, "", p.logFlag)
		}

		p.logger.Print(<-p.logChan)
	}
}

//路径,文件名,行数,函数名称
func (p *Log) outPut(calldepth int, prefix string, str string) {
	pc, file, line, ok := runtime.Caller(calldepth)
	if true != ok {
		return
	}
	funName := runtime.FuncForPC(pc).Name()

	var strLine = strconv.Itoa(line)
	p.logChan <- "[" + prefix + "][" + file + "][" + strLine + "][" + funName + "]" + str
}
