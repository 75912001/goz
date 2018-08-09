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
