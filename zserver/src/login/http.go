package login

import (
	"fmt"
	"net/http"

	"github.com/75912001/goz/zhttp"
)

var gLoginHTTPServer zhttp.Server

func loginHTTPHandler(w http.ResponseWriter, req *http.Request) {
	//todo 找到合适的gateway.
	var strAccount string
	strSession := gServer.Add(&strAccount)
	fmt.Println(strSession)
	//todo 返回给客户端 1.合适的gateway 2.session.

}
