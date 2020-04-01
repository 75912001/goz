package xrUtility

import (
	"fmt"
	"testing"
)

func TestMD5(t *testing.T) {
	var str string
	str = "19426792@qq.com"

	md5str := GenMd5(&str)
	fmt.Println("md5字符串:", md5str)
}

func TestHASH(t *testing.T) {
	var str string
	str = "19426792@qq.com"

	{
		u := HASH32(&str)
		fmt.Println("hash32:", u)
	}
	{
		u := HASH64(&str)
		fmt.Println("hash64", u)
	}
}

func TestJson2map(t *testing.T) {
	var strJson string = "{\"tradeNo\":\"5c84ad403373ec0803dbddddc77246b1\",\"productId\":\"tjlhxkgddj0o1\"}"
	//var jsonMap map[string]interface{}
	//jsonMap = make(map[string]interface{}, 0)

	jsonMap, err := JSON2map(&strJson)
	if nil == err {
		//成功
		fmt.Println("解析json成功:", jsonMap)
	} else {
		//失败
		fmt.Println("解析json失败")
	}
	tradeNo, ok := jsonMap["tradeNo"]
	if ok {
		var TradeNo string = tradeNo.(string)
		fmt.Println("tradeNo:", TradeNo)
	} else {
		//失败
		fmt.Println("失败")
	}
}
