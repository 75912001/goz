package main

import (
	"context"
	"fmt"
	"net/rpc"
	"sync"
	"time"

	"github.com/75912001/goz/zlogin"
	"github.com/smallnest/rpcx/client"
)

//同步调用
//net/rpc
//100w 空req,空res, connect:500, 每个connect2000请求,耗时15秒=>6.666w/秒=>66.66次/毫秒
//100w 空req,空res, connect:5000, 每个connect200请求,耗时17秒=>5.882w/秒=>58.82次/毫秒

//同步调用
//rpcx
//100w 空req,空res, connect:500, 每个connect2000请求,耗时18秒=>5.555w/秒=>55.55次/毫秒
//100w 空req,空res, connect:5000, 每个connect200请求,耗时20秒=>5w/秒=>50次/毫秒

const connectNum int = 500
const rpcNum int = 2000

var addr string = "127.0.0.1:22003"

var req zlogin.REQVerifySession
var res zlogin.RESVerifySession

var gCnt int
var l sync.RWMutex

func RPC() {
	var client [connectNum]*rpc.Client
	var err error
	for i := 0; i < connectNum; i++ {
		client[i], err = rpc.Dial("tcp", addr)
		if err != nil {
			fmt.Println("failed to dial", err)
		}
	}
	for i := 0; i < connectNum; i++ {
		go func(client *rpc.Client) {
			for i := 0; i < rpcNum; i++ {
				err := client.Call("RPC.VerifySession", &req, &res)
				if err != nil {
					fmt.Println("failed to call", err)
				}
				fmt.Println(time.Now().Unix())
			}
		}(client[i])
	}
	time.Sleep(600 * time.Second)
	for i := 0; i < connectNum; i++ {
		client[i].Close()
	}
}

////////////////////////////////////////////

func RPCX() {
	var xclient [connectNum]client.XClient
	//d := client.NewMultipleServersDiscovery()
	d := client.NewPeer2PeerDiscovery("tcp@"+addr, "")
	for i := 0; i < connectNum; i++ {
		xclient[i] = client.NewXClient("RPCX", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	}
	for i := 0; i < connectNum; i++ {
		go func(c client.XClient) {
			for j := 0; j < rpcNum; j++ {
				err := c.Call(context.Background(), "VerifySession", &req, &res)
				if err != nil {
					fmt.Println("failed to call:", err)
				} else {
					l.Lock()
					gCnt++
					var c int = gCnt
					l.Unlock()
					if 1 == c {
						fmt.Println("begin:", time.Now().Unix())
					} else if c == connectNum*rpcNum {
						fmt.Println("end:", time.Now().Unix())
					}
				}
			}
		}(xclient[i])
	}
	time.Sleep(60 * time.Second)
	for i := 0; i < connectNum; i++ {
		xclient[i].Close()
	}
}
