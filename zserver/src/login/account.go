package login

//账号信息
type account struct {
	session    string //登录需要验证的session
	timeOutSec int64  //超时时间(超时后session数据无效)
}

type accountMap map[string]*account //key:account string
