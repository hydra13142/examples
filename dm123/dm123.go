package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

var (
	digi, _ = regexp.Compile(`<b>(\d+)</b>/<b>(\d+)</b>`)
	addr, _ = regexp.Compile(`href="([^"]+)"(?:\starget="_blank")?><img\sborder="0"`)
	sepr, _ = regexp.Compile(`(.+?)(\d+)(?:_\d+)?\.(\w+)$`)
	client  *http.Client
)

func init() {

	// 创建一个使用GoAgent/Wallproxy的代理
	proxy, err := url.Parse("http://127.0.0.1:8087")
	if err != nil {
		return
	}

	// 一个使用代理的客户端，默认客户端失败时进行切换
	client = &http.Client{
		Transport: &http.Transport{
			Proxy:              http.ProxyURL(proxy),
			DisableCompression: false,
		},
	}
}

func main() {

	// 获取相册的网络地址
	if len(os.Args) <= 1 {
		println("usage : url")
		return
	}
	u := os.Args[1]

	// 创建存放图片的目录
	s := sepr.FindStringSubmatch(u)
	if s == nil {
		println("can't get dir name")
		return
	}
	x, y, z := s[1], s[2], s[3]
	os.Mkdir(y, os.ModeDir)

	// 创建一个读取页面的函数
	read := func(u string) []byte {
		r, err := http.Get(u)
		if err != nil {
			r, err = client.Get(u)
			if err != nil {
				println("can't connect url")
				return nil
			}
		}
		defer r.Body.Close()
		d, err := ioutil.ReadAll(r.Body)
		if err != nil {
			println("can't load url")
			return nil
		}
		return d
	}

	// 下载第一个页面，获取当前图片和总图片数
	d := read(u)
	if d == nil {
		println("can't load index page")
		return
	}
	b := digi.FindSubmatch(d)
	if b == nil {
		println("can't get page number")
		return
	}
	i, _ := strconv.Atoi(string(b[1]))
	j, _ := strconv.Atoi(string(b[2]))
	println("[", i, "/", j, "]")

	// 依次下载每个图片
	for i <= j {

		// 获取图片的地址
		d = read(u)
		b = addr.FindSubmatch(d)
		if b == nil {
			println("can't find image addr")
			return
		}
		a := `http://www.dm123.cn` + string(b[1])

		// 保存图片
		println("BEGIN " + u)
		d = read(a)
		if d == nil {
			println("can't load image page")
			return
		}
		ioutil.WriteFile(fmt.Sprintf("%s/%03d.jpg", y, i), d, 0)
		println("OVER  " + u)

		// 生成下一页的地址
		i++
		u = x + y + "_" + strconv.Itoa(i) + "." + z
	}

	// 下载结束
	println("END")
}
