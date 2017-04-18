package zhttp

import (
	"io"
	"io/ioutil"
	"net/http"
)

var defContentLength int64 = 102400

type Client struct {
	Result []byte
}

func (this *Client) Get(url string) (err error) {
	resp, err := http.Get(url)
	if nil != err {
		gLog.Error("######HttpClient.Get err:", err, url)
		return err
	}
	defer resp.Body.Close()

	var contentLength int64
	if resp.ContentLength < 0 {
		contentLength = defContentLength
	} else {
		contentLength = resp.ContentLength
	}
	this.Result = make([]byte, contentLength)

	this.Result, err = ioutil.ReadAll(resp.Body)
	if nil != err {
		gLog.Error("######HttpClient.Get err:", err, resp.Body)
		return err
	}

	return err
}

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
