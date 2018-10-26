package zmysql

import (
	"database/sql"

	//_ "github.com/go-sql-driver/mysql"
	"github.com/75912001/goz/zutility"
)

var gLog *zutility.Log

//MySQL mysql
type MySQL struct {
	dbName string
	db     *sql.DB
}

//Init 初始化
func (p *MySQL) Init(log *zutility.Log, ip string, port string, user string, pwd string, dbName string) int {
	gLog = log
	p.dbName = dbName

	dataSourceName := user + ":" + pwd + "@tcp(" + ip + ":" + port + ")/" + p.dbName + "?charset=utf8"

	//连接 mysql
	var err error
	p.db, err = sql.Open("mysql", dataSourceName)

	if err != nil {
		gLog.Crit("mysql open:", err)
		return -1
	}
	err = p.db.Ping()
	if err != nil {
		gLog.Crit("ping database: %s", err.Error())
		return -1
	}

	return 0
}

//Close 关闭
func (p *MySQL) Close() {
	p.db.Close()
}
