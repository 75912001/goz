package zutility

/////////////////////////////////////////////////////////////////////////////
const (
	EC_SUCC            int = 0 //成功
	EC_DISCONNECT_PEER int = 1 //断开对方的连接 (优化,去掉这个错误码)
	EC_SYS             int = 2 //系统错误
	EC_PARAM           int = 3 //参数错误
	EC_PACKET          int = 4 //包错误
)
