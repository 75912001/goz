package zuser

import (
	"github.com/goz/ztcp"
)

type User struct {
	PeerConn *ztcp.PeerConn
	Uid      ztcp.USER_ID
	Ip       string
	Port     uint16
}
