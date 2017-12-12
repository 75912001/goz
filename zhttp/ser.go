package zhttp

import (
	"net/http"
	"strconv"

	"github.com/goz/zutility"
)

/*
////////////////////////////////////////////////////////////////////////////////
//使用方法
import (
	"zhttp"
)

func main() {
	var gHttpServer zhttp.Server_t
	gHttpServer.AddHandler("/PhoneRegister", PhoneRegisterHttpHandler)
	go gHttpServer.Run(ip, port)
}

func PhoneRegisterHttpHandler(w http.ResponseWriter, req *http.Request) {
}
*/
func SetLog(v *zutility.Log) {
	gLog = v
}

type Server struct {
}

func (this *Server) AddHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (this *Server) Run(ip string, port uint16) {
	httpAddr := ip + ":" + strconv.Itoa(int(port))
	err := http.ListenAndServe(httpAddr, nil)
	if nil != err {
		gLog.Crit("ListenAndServe err: ", err, httpAddr)
	}
}

/*
func (this *Server) RunHttps(ip string, port uint16, certFile string, keyFile string) {
	httpAddr := ip + ":" + strconv.Itoa(int(port))
	err := http.ListenAndServeTLS(httpAddr, certFile, keyFile, nil)
	if nil != err {
		gLog.Crit("ListenAndServe https err: ", err, httpAddr)
	}
}
*/
var gLog *zutility.Log
