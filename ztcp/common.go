package ztcp

import (
	"github.com/75912001/goz/zutility"
)

//SetLog
func SetLog(v *zutility.Log) {
	gLog = v
}

////////////////////////////////////////////////////////////////////////////////
var gLog *zutility.Log
