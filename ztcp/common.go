package ztcp

import (
	"github.com/goz/zutility"
)

func SetLog(v *zutility.Log) {
	gLog = v
}

////////////////////////////////////////////////////////////////////////////////
var gLog *zutility.Log
