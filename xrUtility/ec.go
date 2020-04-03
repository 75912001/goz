//错误码
package xrUtility

/*
型号名称：	MacBook Pro
  型号标识符：	MacBookPro11,4
  处理器名称：	Intel Core i7
  处理器速度：	2.2 GHz
  处理器数目：	1
  核总数：	4
  L2 缓存（每个核）：	256 KB
  L3 缓存：	6 MB
  内存：	16 GB
*/

/////////////////////////////////////////////////////////////////////////////
const (
	ECSucc           int = 0x0000 //成功
	ECDisconnectPeer int = 0x0001 //断开对方的连接
	ECSYS            int = 0x0002 //系统错误
	ECParam          int = 0x0003 //参数错误
	ECPacket         int = 0x0004 //包错误
	ECSMSSending     int = 0x0501 //sms已发出
	ECSMSBind        int = 0x0502 //sms号码已绑定
	ECRedisSYS       int = 0x0503 //redis系统错误
)
