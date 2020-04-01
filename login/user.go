package main

import (
	"github.com/75912001/goz/ztcp"
)

//User 用户
type User struct {
	PeerConn *ztcp.PeerConn
	UID      UserID
	IP       string
	Port     uint16
}
