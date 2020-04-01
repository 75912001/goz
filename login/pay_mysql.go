package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var gMysqlDbName string
var gPayDataSourceName string
var gMysqldb *sql.DB

//Pay 支付
type Pay struct {
	OrderID         string
	Platform        uint32
	Account         string
	UID             uint64
	OrderStatus     uint32
	ConsumeStreamID string
	PayChannel      string
	Product         string
	Amount          string
}

func initMysql() int {
	////////////////////////////////////////////////////////////////////////////
	//连接 mysql
	var err error
	gMysqldb, err = sql.Open("mysql", gPayDataSourceName)

	if err != nil {
		gLog.Crit("###### mysql open:", err)
		return -1
	}
	err = gMysqldb.Ping()
	if err != nil {
		gLog.Crit("###### ping database: %s", err.Error())
		return -1
	}
	//	defer mysqldb.Close()
	////////////////////////////////////////////////////////////////////////////

	/*
		{ // 获取t_pay表中的记录
			rows, err := g_mysqldb.Query("SELECT order_id,platform,account,uid FROM t_pay where order_status=0")
			if err != nil {
				gLog.Crit("###### fetech data failed:", err.Error())
				return -1
			}
			defer rows.Close()

			g_PayMap = make(Pay_MAP)

			for rows.Next() {
				var pay Pay_t
				rows.Scan(&pay.OrderId, &pay.Platform, &pay.Account, &pay.Uid)
				g_PayMap[pay.OrderId] = pay
			}
			gLog.Trace(g_PayMap)
		}
	*/
	return 0
}

func payMysqlInsert(orderID string, platform uint32, account string,
	uid UserID, orderStatus uint32, timeSec uint32, consumeStreamID string,
	payChannel string, product string) int {

	////////////////////////////////////////////////////////////////////////////
	// 插入一条新数据
	_, err := gMysqldb.Exec("INSERT INTO t_pay(order_id,platform,account,uid,order_status,time_sec,consume_stream_id,pay_channel, product) VALUES(?,?,?,?,?,?,?,?,?)",
		orderID, platform, account, uid, orderStatus, timeSec, consumeStreamID, payChannel, product)
	if err != nil {
		gLog.Crit("###### insert data failed:", err.Error())
		return -1
	}

	return 0
}

func payMysqlDel(orderID string) {
	_, err := gMysqldb.Exec("DELETE FROM t_pay WHERE order_id=?", orderID)
	if err != nil {
		gLog.Crit("###### DELETE failed:", err.Error())
	}
}

func payMysqlTimeOutDel() {
	nowTimeSec := uint32(time.Now().Unix())
	_, err := gMysqldb.Exec("DELETE FROM t_pay WHERE time_sec<?", nowTimeSec-24*60*60*7)
	if err != nil {
		gLog.Crit("###### DELETE failed:", err.Error())
	}
}

func payMysqlSelectPay(orderID string) (ret int, pay Pay) {
	{ // 获取t_pay表中的记录
		rows, err := gMysqldb.Query("SELECT order_id,platform,account,uid,order_status,consume_stream_id,pay_channel,product FROM t_pay where order_id=?", orderID)
		if err != nil {
			gLog.Crit("###### fetech data failed:", err.Error())
			ret = -1
			return
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&pay.OrderID, &pay.Platform, &pay.Account, &pay.UID, &pay.OrderStatus, &pay.ConsumeStreamID, &pay.PayChannel, &pay.Product)
		}
	}
	return
}

func payMysqlUpdatePay(pay *Pay) int {
	nowTime := time.Now()
	yyyymm := nowTime.Format("200601")
	nowTimeSec := nowTime.Unix()

	strSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.t_pay_%s("+
		"order_id char(64) NOT NULL COMMENT '充值订单号',"+
		"platform int(11) UNSIGNED NOT NULL COMMENT '账号登陆平台',"+
		"account char(64) NOT NULL COMMENT '账号',"+
		"uid BIGINT(18) UNSIGNED NOT NULL COMMENT'用户id',"+
		"order_status int(11) UNSIGNED NOT NULL COMMENT '订单状态[0:新单,1:F失败,2:完成,3:完成测试]',"+
		"time_sec int(11) UNSIGNED NOT NULL COMMENT '时间',"+
		"consume_stream_id char(64) NOT NULL COMMENT '消费流水号',"+
		"pay_channel char(32) NOT NULL COMMENT '支付渠道',"+
		"product char(32) not null comment 'product',"+
		"amount char(32) not null comment 'amount',"+
		"PRIMARY KEY (`order_id`)"+
		")ENGINE=INNODB DEFAULT CHARSET=utf8", gMysqlDbName, yyyymm)
	{
		_, err := gMysqldb.Exec(strSQL)
		if err != nil {
			gLog.Crit("###### CREATE TABLE failed:", err.Error())
			return -1
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// 插入一条新数据
	strSQLInsert := fmt.Sprintf("INSERT INTO t_pay_%s(order_id,platform,account,uid,order_status,time_sec,consume_stream_id,pay_channel,product,amount)", yyyymm)
	_, err := gMysqldb.Exec(strSQLInsert+"VALUES(?,?,?,?,?,?,?,?,?,?)",
		pay.OrderID, pay.Platform, pay.Account, pay.UID, pay.OrderStatus,
		nowTimeSec, pay.ConsumeStreamID, pay.PayChannel, pay.Product, pay.Amount)
	if err != nil {
		_, errUpdate := gMysqldb.Exec("UPDATE t_pay SET order_status=?,consume_stream_id=?,pay_channel=? WHERE order_id=?",
			pay.OrderStatus, pay.ConsumeStreamID, pay.PayChannel, pay.OrderID)
		if errUpdate != nil {
			gLog.Crit("###### UPDATE data failed:", errUpdate.Error())
			return -1
		}
		gLog.Crit("###### insert data failed:", err.Error())
		return -1
	}
	payMysqlDel(pay.OrderID)
	return 0
}

//ZfbPay 支付宝支付
type ZfbPay struct {
	OrderID         string
	Platform        uint32
	Account         string
	UID             uint64
	OrderStatus     uint32
	ConsumeStreamID string
	PayChannel      string
	Product         string
	Amount          string
	ReceiptAmount   string
}

func zfbPayMysqlSelectPay(orderID string) (ret int, pay ZfbPay) {
	{ // 获取t_pay表中的记录
		rows, err := gMysqldb.Query("SELECT order_id,platform,account,uid,order_status,consume_stream_id,pay_channel,product FROM t_pay where order_id=?", orderID)
		if err != nil {
			gLog.Crit("###### fetech data failed:", err.Error())
			ret = -1
			return
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&pay.OrderID, &pay.Platform, &pay.Account, &pay.UID, &pay.OrderStatus, &pay.ConsumeStreamID, &pay.PayChannel, &pay.Product)
		}
	}
	return
}

func zfbPayMysqlUpdatePay(pay *ZfbPay) int {
	nowTime := time.Now()
	yyyymm := nowTime.Format("200601")
	nowTimeSec := nowTime.Unix()

	strSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.t_zfb_pay_%s("+
		"order_id char(64) NOT NULL COMMENT '充值订单号',"+
		"platform int(11) UNSIGNED NOT NULL COMMENT '账号登陆平台',"+
		"account char(64) NOT NULL COMMENT '账号',"+
		"uid BIGINT(18) UNSIGNED NOT NULL COMMENT'用户id',"+
		"order_status int(11) UNSIGNED NOT NULL COMMENT '订单状态[0:新单,1:F失败,2:完成,3:完成测试]',"+
		"time_sec int(11) UNSIGNED NOT NULL COMMENT '时间',"+
		"consume_stream_id char(64) NOT NULL COMMENT '消费流水号',"+
		"pay_channel char(32) NOT NULL COMMENT '支付渠道',"+
		"product char(32) not null comment 'product',"+
		"amount char(32) not null comment '订单金额',"+
		"receipt_amount char(32) not null comment '实收金额',"+
		"PRIMARY KEY (`order_id`)"+
		")ENGINE=INNODB DEFAULT CHARSET=utf8", gMysqlDbName, yyyymm)
	{
		_, err := gMysqldb.Exec(strSQL)
		if err != nil {
			gLog.Crit("###### CREATE TABLE failed:", err.Error())
			return -1
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// 插入一条新数据
	strSQLInsert := fmt.Sprintf("INSERT INTO t_zfb_pay_%s(order_id,platform,account,uid,order_status,time_sec,consume_stream_id,pay_channel,product,amount,receipt_amount)", yyyymm)
	_, err := gMysqldb.Exec(strSQLInsert+"VALUES(?,?,?,?,?,?,?,?,?,?,?)",
		pay.OrderID, pay.Platform, pay.Account, pay.UID, pay.OrderStatus,
		nowTimeSec, pay.ConsumeStreamID, pay.PayChannel, pay.Product, pay.Amount, pay.ReceiptAmount)
	if err != nil {
		_, errUpdate := gMysqldb.Exec("UPDATE t_pay SET order_status=?,consume_stream_id=?,pay_channel=? WHERE order_id=?",
			pay.OrderStatus, pay.ConsumeStreamID, pay.PayChannel, pay.OrderID)
		if errUpdate != nil {
			gLog.Crit("###### UPDATE data failed:", errUpdate.Error())
			return -1
		}
		gLog.Crit("###### insert data failed:", err.Error())
		return -1
	}
	payMysqlDel(pay.OrderID)

	return 0
}

func wxPayMysqlUpdatePay(pay *Pay) int {
	nowTime := time.Now()
	yyyymm := nowTime.Format("200601")
	nowTimeSec := nowTime.Unix()

	strSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.t_wx_pay_%s("+
		"order_id char(64) NOT NULL COMMENT '充值订单号',"+
		"platform int(11) UNSIGNED NOT NULL COMMENT '账号登陆平台',"+
		"account char(64) NOT NULL COMMENT '账号',"+
		"uid BIGINT(18) UNSIGNED NOT NULL COMMENT'用户id',"+
		"order_status int(11) UNSIGNED NOT NULL COMMENT '订单状态[0:新单,1:F失败,2:完成,3:完成测试]',"+
		"time_sec int(11) UNSIGNED NOT NULL COMMENT '时间',"+
		"consume_stream_id char(64) NOT NULL COMMENT '消费流水号',"+
		"pay_channel char(32) NOT NULL COMMENT '支付渠道',"+
		"product char(32) not null comment 'product',"+
		"amount char(32) not null comment '订单金额',"+
		"PRIMARY KEY (`order_id`)"+
		")ENGINE=INNODB DEFAULT CHARSET=utf8", gMysqlDbName, yyyymm)
	{
		_, err := gMysqldb.Exec(strSQL)
		if err != nil {
			gLog.Crit("###### CREATE TABLE failed:", err.Error())
			return -1
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// 插入一条新数据
	strSQLInsert := fmt.Sprintf("INSERT INTO t_wx_pay_%s(order_id,platform,account,uid,order_status,time_sec,consume_stream_id,pay_channel,product,amount)", yyyymm)
	_, err := gMysqldb.Exec(strSQLInsert+"VALUES(?,?,?,?,?,?,?,?,?,?)",
		pay.OrderID, pay.Platform, pay.Account, pay.UID, pay.OrderStatus,
		nowTimeSec, pay.ConsumeStreamID, pay.PayChannel, pay.Product, pay.Amount)
	if err != nil {
		_, errUpdate := gMysqldb.Exec("UPDATE t_pay SET order_status=?,consume_stream_id=?,pay_channel=? WHERE order_id=?",
			pay.OrderStatus, pay.ConsumeStreamID, pay.PayChannel, pay.OrderID)
		if errUpdate != nil {
			gLog.Crit("###### UPDATE data failed:", errUpdate.Error())
			return -1
		}
		gLog.Crit("###### insert data failed:", err.Error())
		return -1
	}
	payMysqlDel(pay.OrderID)

	return 0
}
