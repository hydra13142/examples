package main

// 简单的代理，缺点是没有支持https协议，或者说没有tls支持
// 这导致A站和B站的一些验证页面无效，进而导致不能观看视频

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var client = http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return fmt.Errorf("Redirect Disabled") // 禁止自动重定向
		},
	}

func proxy(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	r.RequestURI = "" // 发送的Request这个字段必须是空字符串

	m, err := client.Do(r)
	if err != nil {
		http.NotFound(w, r) // 简单粗暴的把错误全部改成404
		return
	}
	defer m.Body.Close()

	for i, j := range m.Header {
		for _, k := range j {
			w.Header().Add(i, k)
		}
	}
	w.WriteHeader(m.StatusCode) // 传递响应头

	data := make([]byte, 4096)
	for {
		n, err := m.Body.Read(data)
		if err != nil {
			if err == io.EOF {
				w.Write(data[:n]) // 发送结束
			}
			break
		}
		w.Write(data[:n]) // 成功的Read读取字节数也有可能达不到长度
	}
}

func main() {
	http.HandleFunc("/", proxy)
	log.Println("Start serving on port 8080")
	http.ListenAndServe(":8080", nil)
}
