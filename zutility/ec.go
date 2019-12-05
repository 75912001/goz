//错误码
package zutility

/////////////////////////////////////////////////////////////////////////////
const (
	ECSucc           int = 0 //成功
	ECDisconnectPeer int = 1 //断开对方的连接
	ECSYS            int = 2 //系统错误
	ECParam          int = 3 //参数错误
	ECPacket         int = 4 //包错误
	//	ECSMSSending     int = 5 //sms已发出
	//	ECSMSBind        int = 6 //sms号码已绑定
	ECRedisSYS int = 7 //redis系统错误
)
