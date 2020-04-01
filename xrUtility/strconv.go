package xrUtility

import (
	"strconv"
	"strings"
)

//字符串转->数值类型

//StringToInt
func StringToInt(s *string) (value int, err error) {
	vaule, err := strconv.ParseInt(*s, 10, 0)
	if nil != err {
		return 0, err
	}
	return int(vaule), err
}

//StringToUint16
func StringToUint16(s *string) (value uint16, err error) {
	vaule, err := strconv.ParseUint(*s, 10, 16)
	if nil != err {
		return 0, err
	}
	return uint16(vaule), err
}

//StringToUint32
func StringToUint32(s *string) (value uint32, err error) {
	vaule, err := strconv.ParseUint(*s, 10, 32)
	if nil != err {
		return 0, err
	}
	return uint32(vaule), err
}

//StringToUint64
func StringToUint64(s *string) (value uint64, err error) {
	vaule, err := strconv.ParseUint(*s, 10, 64)
	if nil != err {
		return 0, err
	}
	return vaule, err
}

//StringToInt32
func StringToInt32(s *string) (value int32, err error) {
	vaule, err := strconv.ParseInt(*s, 10, 32)
	if nil != err {
		return 0, err
	}
	return int32(vaule), err
}

//StringToInt64
func StringToInt64(s *string) (value int64, err error) {
	return strconv.ParseInt(*s, 10, 64)
}

//IntToString int->string
func IntToString(v int) string {
	return strconv.Itoa(v)
}

//StringSplit [s:"1,2,3,4"] sep:"," => return:[1 2 3 4]
func StringSplit(s *string, sep string) []string {
	return strings.Split(*s, sep)
}

//StringSubstrRune 获取string前length字符(unicode)的string
//StringSubstrRune("你好,我是rune,3,4,5,6,7,8,9", 7)
//return:你好,我是ru
func StringSubstrRune(s *string, length int) (value string) {
	r := []rune(*s)
	return string(r[0:length])
}

//Byte2String byte->string
func Byte2String(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[:i])
		}
	}
	return string(p)
}
