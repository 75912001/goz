//读取ini文件
//#开始的是注释行,不读取
//以下是ini文件的例子
/*
[server]
ip=192.168.8.101
port=9988

[common]
#child fd max value, def:20000
max_fd_num=1000
#tcp listen number, def:1024
#listen_num=1024
*/

//使用方法
/*
var ini Ini
ini.Load("xxx.ini")
ip := ini.GetString("server", "ip", "")
*/

package xrUtility

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type keyMap map[string]string
type sectionMap map[string]keyMap

//Ini ini文件
type Ini struct {
	sectionMap sectionMap //存取配置文件
}

//Load 加载文件
func (p *Ini) Load(path string) (err error) {
	p.init()

	var file *os.File
	{
		file, err = os.Open(path)

		if nil != err {
			return err
		}
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var section string

	for {
		line, err := reader.ReadString('\n')
		if nil != err {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		case 0 == len(line):
		case '#' == line[0]:
		//匹配[xxx]然后存储
		case '[' == line[0] && ']' == line[len(line)-1]:
			section = line[1 : len(line)-1]
		default:
			symbolIndex := strings.IndexAny(line, "=")
			if -1 == symbolIndex {
				fmt.Println("err: Ini Load no '=' symbol:", line)
				break
			}
			key := line[0:symbolIndex]
			value := line[symbolIndex+1:]
			p.load(section, key, value)
		}
	}

	return err
}

//GetString 获取对应的值
func (p *Ini) GetString(section string, key string, defaultValue string) (value string) {
	sectionValue, valid := p.sectionMap[section]
	if valid {
		keyValue, valid := sectionValue[key]
		if valid {
			return keyValue
		}
	}
	return defaultValue
}

//GetUint32 获取uint32
func (p *Ini) GetUint32(section string, key string, defaultValue uint32) (value uint32) {
	def := strconv.FormatUint(uint64(defaultValue), 10)
	str := p.GetString(section, key, def)
	value, _ = StringToUint32(&str)
	return value
}

//GetInt 获取int
func (p *Ini) GetInt(section string, key string, defaultValue int) (value int) {
	def := strconv.Itoa(defaultValue)
	str := p.GetString(section, key, def)
	value, _ = StringToInt(&str)
	return value
}

//GetUint16 获取uint16
func (p *Ini) GetUint16(section string, key string, defaultValue uint16) (value uint16) {
	def := strconv.FormatUint(uint64(defaultValue), 10)
	str := p.GetString(section, key, def)
	value, _ = StringToUint16(&str)
	return value
}

//GetInt64 获取int64
func (p *Ini) GetInt64(section string, key string, defaultValue int64) (value int64) {
	def := strconv.FormatInt(int64(defaultValue), 10)
	str := p.GetString(section, key, def)
	value, _ = StringToInt64(&str)
	return value
}

////////////////////////////////////////////////////////////////////////////////
//初始化
func (p *Ini) init() {
	p.sectionMap = make(sectionMap)
}

//加载文件到内存中
func (p *Ini) load(section string, key string, value string) {
	_, valid := p.sectionMap[section]
	if valid {
		p.sectionMap[section][key] = value
	} else {
		keyMap := make(keyMap)
		keyMap[key] = value
		p.sectionMap[section] = keyMap
	}
}
