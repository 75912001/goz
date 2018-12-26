/*
////////////////////////////////////////////////////////////////////////////////
//使用方法
import (
	"zredis"
)

var GRedis zredis.Server
err := GRedis.Connect("127.0.0.1", 6379, 0)
if nil != err {
	fmt.Println("######GRedis.Connect(ip, port, redisDatabases) err:", err)
	return
}
*/

package zredis

import (
	"github.com/75912001/goz/zutility"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type Server struct {
	Conn      redis.Conn
	ip        string
	port      uint16
	dataBases int
}

func SetLog(v *zutility.Log) {
	gLog = v
}

//连接
func (p *Server) Connect(ip string, port uint16, dataBases int) (err error) {
	p.ip = ip
	p.port = port
	p.dataBases = dataBases

	var addr = ip + ":" + strconv.Itoa(int(port))
	dialOption := redis.DialDatabase(dataBases)

	p.Conn, err = redis.Dial("tcp", addr, dialOption)
	if nil != err {
		gLog.Crit("redis.Dial err:", err, ip, port, dataBases)
	}
	return err
}

var gLog *zutility.Log
