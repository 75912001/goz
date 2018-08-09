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

package zutility

import (
	"bufio"
	"os"
	"strings"
)

type keyMap map[string]string
type sectionMap map[string]keyMap

//ini文件
type Ini struct {
	sectionMap sectionMap //存取配置文件
}

//加载文件
func (this *Ini) Load(path string) (err error) {
	this.init()

	var file *os.File
	file, err = os.Open(path)

	if nil != err {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	var section string

	for {
		line, readErr := reader.ReadString('\n')
		if nil != readErr {
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
				break
			}
			key := line[0:symbolIndex]
			value := line[symbolIndex+1:]
			this.load(section, key, value)
		}
	}

	return
}

//获取对应的值
func (this *Ini) GetString(section string, key string, defaultValue string) (value string) {
	sectionValue, valid := this.sectionMap[section]
	if valid {
		keyValue, valid := sectionValue[key]
		if valid {
			return keyValue
		}
	}
	return defaultValue
}

func (this *Ini) GetUint32(section string, key string, defaultValue string) (value uint32) {
	var str string
	str = this.GetString(section, key, defaultValue)
	return StringToUint32(&str)
}

func (this *Ini) GetInt(section string, key string, defaultValue string) (value int) {
	var str string
	str = this.GetString(section, key, defaultValue)
	return StringToInt(&str)
}

func (this *Ini) GetUint16(section string, key string, defaultValue string) (value uint16) {
	var str string
	str = this.GetString(section, key, defaultValue)
	return StringToUint16(&str)
}

////////////////////////////////////////////////////////////////////////////////
//初始化
func (p *Ini) init() {
	p.sectionMap = make(sectionMap)
}

//加载文件到内存中
func (this *Ini) load(section string, key string, value string) {
	_, valid := this.sectionMap[section]
	if valid {
		this.sectionMap[section][key] = value
	} else {
		keyMap := make(keyMap)
		keyMap[key] = value
		this.sectionMap[section] = keyMap
	}
}
