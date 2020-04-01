package xrUtility

import (
	"fmt"
	"testing"
)

func TestStringToInt(t *testing.T) {
	{
		var s1 string = "01aa"
		v, e := StringToInt(&s1)
		fmt.Println(v, e)
	}
	{
		var s1 string = "01"
		v, e := StringToInt(&s1)
		fmt.Println(v, e)
	}
	{
		var s1 string = "123"
		v, e := StringToInt(&s1)
		fmt.Println(v, e)
	}
}
