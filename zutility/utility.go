package zutility

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
)

//Int32Min -2147483648
const Int32Min = ^Int32Max

//Int32Max 2147483647
const Int32Max = int32(^uint32(0) >> 1)

//Int64Min -9223372036854775808
const Int64Min = ^Int64Max

//Int64Max 9223372036854775807
const Int64Max = int64(^uint64(0) >> 1)

//Uint32Min 0
const Uint32Min uint32 = 0

//Uint32Max 4294967295
const Uint32Max = ^uint32(0)

//Uint64Min 0
const Uint64Min = ^Uint64Max

//Uint64Max 18446744073709551615
const Uint64Max = ^uint64(0)

//IntMin -9223372036854775808
const IntMin = ^IntMax

//IntMax 9223372036854775807
const IntMax = int(^uint(0) >> 1)

//UintMin 0
const UintMin uint = 0

//UintMax 18446744073709551615
const UintMax = ^uint(0)

//GenMd5 md5
func GenMd5(s *string) (value string) {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(*s))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

//HASH 哈希
func HASH(s *string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(*s))
	return h.Sum32()
}

//HASHEL 配合libel库
func HASHEL(s *string) uint32 {
	var h uint32
	rs := []rune(*s)
	n := len(rs)
	for i := 0; i < n; i++ {
		h = 31*h + uint32(rs[i])
	}
	return h
}

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

//JSON2map JSON => MAP
func JSON2map(strJSON *string) (s map[string]interface{}, err error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(*strJSON), &result); err != nil {
		return nil, err
	}
	return result, nil
}
