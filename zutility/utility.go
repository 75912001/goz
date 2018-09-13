package zutility

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
)

//-2147483648
const INT32_MIN = ^INT32_MAX

//2147483647
const INT32_MAX = int32(^uint32(0) >> 1)

//-9223372036854775808
const INT64_MIN = ^INT64_MAX

//9223372036854775807
const INT64_MAX = int64(^uint64(0) >> 1)

//0
const UINT32_MIN uint32 = 0

//4294967295
const UINT32_MAX = ^uint32(0)

//0
const UNT64_MIN = ^UNT64_MAX

//18446744073709551615
const UNT64_MAX = ^uint64(0)

//-9223372036854775808
const INT_MIN = ^INT_MAX

//9223372036854775807
const INT_MAX = int(^uint(0) >> 1)

//0
const UINT_MIN uint = 0

//18446744073709551615
const UINT_MAX = ^uint(0)

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

//配合libel库
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
//var strJson string = "{\"tradeNo\":\"5c84ad403373ec0803dbddddc77246b1\",\"productId\":\"tjlhxkgddj0o1\"}"
//var jsonMap map[string]interface{}
//jsonMap = make(map[string]interface{}, 0)
//if jsonMap, err = Json2map(&strJson); err == nil {
//成功
//} else {
//失败
//}
//tradeNo, ok := jsonMap["tradeNo"]
//if ok {
//	var TradeNo string = tradeNo.(string)
//} else {
//失败
//}
func Json2map(strJson *string) (s map[string]interface{}, err error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(*strJson), &result); err != nil {
		return nil, err
	}
	return result, nil
}
