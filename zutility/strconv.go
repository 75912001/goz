package zutility

import (
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////
//字符串转->数值类型

//StringToInt 失败返回0
func StringToInt(s *string) (value int) {
	vaule, err := strconv.ParseInt(*s, 10, 0)
	if nil != err {
		return 0
	}
	return int(vaule)
}

//StringToUint16 失败返回0
func StringToUint16(s *string) (value uint16) {
	vaule, err := strconv.ParseUint(*s, 10, 16)
	if nil != err {
		return 0
	}
	return uint16(vaule)
}

//StringToUint32 失败返回0
func StringToUint32(s *string) (value uint32) {
	vaule, err := strconv.ParseUint(*s, 10, 32)
	if nil != err {
		return 0
	}
	return uint32(vaule)
}

//StringToUint64 失败返回0
func StringToUint64(s *string) (value uint64) {
	vaule, err := strconv.ParseUint(*s, 10, 64)
	if nil != err {
		return 0
	}
	return vaule
}

//StringToInt32 失败返回0
func StringToInt32(s *string) (value int32) {
	vaule, err := strconv.ParseInt(*s, 10, 32)
	if nil != err {
		return 0
	}
	return int32(vaule)
}

//StringToInt64 失败返回0
func StringToInt64(s *string) (value int64) {
	vaule, err := strconv.ParseInt(*s, 10, 64)
	if nil != err {
		return 0
	}
	return vaule
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
