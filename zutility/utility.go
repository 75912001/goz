package zutility

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
)

////////////////////////////////////////////////////////////////////////////////
//md5
func GenMd5(s *string) (value string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(*s))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func HASH(s *string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(*s))
	return h.Sum32()
}

func HASH_EL(s *string) uint32 {
	var h uint32
	rs := []rune(*s)
	n := len(rs)
	for i := 0; i < n; i++ {
		h = 31*h + uint32(rs[i])
	}

	return h
}

////////////////////////////////////////////////////////////////////////////////
//
//把请求包定义成一个结构体
type JsonRequestBody struct {
	Req string
}

//以指针的方式传入，但在使用时却可以不用关心
// result 是函数内的临时变量，作为返回值可以直接返回调用层
func (r *JsonRequestBody) Json2map() (s map[string]interface{}, err error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(r.Req), &result); err != nil {
		return nil, err
	}
	return result, nil
}
