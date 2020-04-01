package xrUtility

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"runtime"
)

//通道,接口
//type IFChan interface {
//}

//GenMd5 生成md5
func GenMd5(s *string) (value string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(*s))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

//HASH32
func HASH32(s *string) uint32 {
	h := fnv.New32()
	h.Write([]byte(*s))
	return h.Sum32()
}

//HASH64
func HASH64(s *string) uint64 {
	h := fnv.New64()
	h.Write([]byte(*s))
	return h.Sum64()
}

//JSON2map JSON转换成为Map
func JSON2map(strJSON *string) (s map[string]interface{}, err error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(*strJSON), &result); err != nil {
		return nil, err
	}
	return result, nil
}

/*
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func GB2312ToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.HZGB2312.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
*/

//IsWindows win
func IsWindows() bool {
	return `windows` == runtime.GOOS
}

//IsLinux linux
func IsLinux() bool {
	return `linux` == runtime.GOOS
}

//IsDarwin darwin
func IsDarwin() bool {
	return `darwin` == runtime.GOOS
}
