package main

import (
	"github.com/75912001/goz/ztcp"
)

//UserMap 用户map
type UserMap map[*ztcp.PeerConn]*User

//UserIDMap 用户map
type UserIDMap map[UserID]*User

type userMgr struct {
	userMap   UserMap
	userIDMap UserIDMap
	userCnt   uint32
}

//GuserMgr  用户管理器
var GuserMgr userMgr

func init() {
	GuserMgr.Init()
}

func (userMgr *userMgr) Init() {
	userMgr.userMap = make(UserMap)
	userMgr.userIDMap = make(UserIDMap)
}

func (userMgr *userMgr) AddUser(peerConn *ztcp.PeerConn) (user *User) {
	user = new(User)

	user.PeerConn = peerConn
	userMgr.userMap[peerConn] = user
	return
}

func (userMgr *userMgr) DelUser(peerConn *ztcp.PeerConn) {
	delete(userMgr.userMap, peerConn)
}

func (userMgr *userMgr) Find(peerConn *ztcp.PeerConn) (user *User) {
	user, _ = userMgr.userMap[peerConn]
	return
}

func (userMgr *userMgr) AddUserID(uid UserID, user *User) {
	userMgr.userIDMap[uid] = user
}

func (userMgr *userMgr) DelUserID(uid UserID) {
	delete(userMgr.userIDMap, uid)
}

func (userMgr *userMgr) FindID(uid UserID) (user *User) {
	user, _ = userMgr.userIDMap[uid]
	return
}
