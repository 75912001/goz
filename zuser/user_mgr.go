package zuser

import (
	"github.com/goz/ztcp"
)

type USER_MAP map[*ztcp.PeerConn]*User

type UserMgr struct {
	userMap USER_MAP
}

func (this *UserMgr) Init() {
	this.userMap = make(USER_MAP)
}

func (this *UserMgr) AddUser(peerConn *ztcp.PeerConn) (user *User) {
	user = new(User)

	user.PeerConn = peerConn
	this.userMap[peerConn] = user
	return
}

func (this *UserMgr) DelUser(peerConn *ztcp.PeerConn) {
	delete(this.userMap, peerConn)
}

func (this *UserMgr) Find(peerConn *ztcp.PeerConn) (user *User) {
	user, _ = this.userMap[peerConn]
	return
}
