package main

// 简单的代理，缺点是没有支持https协议，或者说没有tls支持
// 这导致A站和B站的一些验证页面无效，进而导致不能观看视频

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"crypto/tls"
)

var  client *http.Client

func init() {
	proxy, err := url.Parse("http://127.0.0.1:8087")
	if err != nil {
		return
	}
	client = &http.Client{
		Transport: &http.Transport {
			TLSClientConfig   : &tls.Config{InsecureSkipVerify: true},
			Proxy			  : http.ProxyURL(proxy),
			DisableCompression: false,
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return fmt.Errorf("Redirect Disabled") // 禁止自动重定向
		},
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	req.RequestURI = "" // 发送的Request这俩字段必须是空字符串
	req.RemoteAddr = ""
	req.Header.Del("Proxy-Connection")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 410) // 简单粗暴的把错误全部改成404
		return
	}
	defer resp.Body.Close()

	h := w.Header()
	for k, v := range resp.Header {
		for _, m := range v {
			h.Add(k, m)
		}
	}
	h.Del("Connection")
	w.WriteHeader(resp.StatusCode) // 传递响应头

	data := make([]byte, 4096)
	s, l := int64(0), resp.ContentLength
	if l <= 0 {
		l = (1 << 63) - 1
	}
	for {
		n, er1 := resp.Body.Read(data)
		s += int64(n)
		_, er2 := w.Write(data[:n]) // 成功的Read读取字节数也有可能达不到长度
		if er1 != nil || er2 != nil || s >= l {
			break // 放在最后判断，在读、写错误或者完成时跳出
		}
	}
}

func main() {
	http.HandleFunc("/", proxy)
	log.Println("Start serving on port 8080")
	http.ListenAndServe(":8080", nil)
}
