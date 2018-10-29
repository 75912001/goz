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

const (
	eventTypeMsg        int = 0 //??
	eventTypeDisConnect int = 1 //????
	eventTypeConnect    int = 2 //???
	eventTypeSendMsg    int = 3 //??????
)
