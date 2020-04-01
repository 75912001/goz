package login

import (
	"context"
	"fmt"
	"net"
	"net/rpc"

	"github.com/75912001/goz/zlogin"
	"github.com/smallnest/rpcx/server"
)

/////////////////////////////////////////////
type IFRPC interface {
	create(addr *string)
}
type RPC struct {
}

func (p *RPC) create(addr *string) {
	err := rpc.Register(new(RPC))
	if nil != err {
		fmt.Println("failed to register publisher: ", err)
	}
	listen, err := net.Listen("tcp", *addr)
	if nil != err {
		fmt.Println("failed to listen tcp:", err)
	}
	defer listen.Close()
	rpc.Accept(listen)
}

func (p *RPC) VerifySession(req zlogin.REQVerifySession, res *zlogin.RESVerifySession) (err error) {
	//检查account的session是否合法
	{
		gServer.accMapLock.RLock()
		defer gServer.accMapLock.RUnlock()

		acc := gServer.Find(&req.Account)
		if nil == acc {
			res.ErrorCode = 1
			return nil
		}
		if acc.session != req.Session {
			res.ErrorCode = 2
			return nil
		}
		//验证通过后,标记超时,自动删除
		acc.timeOutSec = 0
	}
	res.ErrorCode = 0

	return nil
}

///////////////
type RPCX struct {
}

func (p *RPCX) create(addr *string) {
	s := server.NewServer()
	s.Register(new(RPCX), "")
	s.Serve("tcp", *addr)
}

func (p *RPCX) VerifySession(ctx context.Context, req *zlogin.REQVerifySession, res *zlogin.RESVerifySession) (err error) {
	res.ErrorCode = 0
	//	time.Sleep(10 * time.Second)
	return nil
}
