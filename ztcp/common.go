package ztcp

import (
	"github.com/goz/zutility"
)

//SetLog 设置日志
func SetLog(v *zutility.Log) {
	gLog = v
}

////////////////////////////////////////////////////////////////////////////////
var gLog *zutility.Log
