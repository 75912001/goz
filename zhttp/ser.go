package zhttp

import (
	"fmt"
	"net/http"
	"strconv"
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

type Server struct {
}

func (this *Server) AddHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}
func (this *Server) Run(ip string, port uint16) {
	httpAddr := ip + ":" + strconv.Itoa(int(port))
	err := http.ListenAndServe(httpAddr, nil)
	if nil != err {
		fmt.Println("ListenAndServe err: ", err, httpAddr)
	}
}
