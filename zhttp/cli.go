package zhttp

import (
	//	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var defContentLength int64 = 102400

//Client 客户端
type Client struct {
	Result []byte
}

//Get 获取
func (p *Client) Get(url string) (err error) {
	resp, err := http.Get(url)
	if nil != err {
		gLog.Error("######HttpClient.Get err:", err, url)
		return err
	}
	//	fmt.Println(resp)
	defer resp.Body.Close()

	var contentLength int64
	if resp.ContentLength < 0 {
		contentLength = defContentLength
	} else {
		contentLength = resp.ContentLength
	}

	p.Result = make([]byte, contentLength)

	//	fmt.Println(resp.Body)
	p.Result, err = ioutil.ReadAll(resp.Body)

	if nil != err {
		gLog.Error("######HttpClient.Get err:", err, resp.Body)
		return err
	}
	return err
}

//Post 发送
func (p *Client) Post(urlData string, bodyType string, body io.Reader) (err error) {
	resp, err := http.Post(urlData, bodyType, body)
	if nil != err {
		gLog.Error("######HttpClient.Post err:", err)
		return err
	}

	defer resp.Body.Close()

	var contentLength int64
	if resp.ContentLength < 0 {
		contentLength = defContentLength
	} else {
		contentLength = resp.ContentLength
	}

	p.Result = make([]byte, contentLength)

	p.Result, err = ioutil.ReadAll(resp.Body)
	return err
}
