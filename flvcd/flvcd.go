package main

import (
	"bytes"
	"code.google.com/p/go.text/encoding/simplifiedchinese"
	"code.google.com/p/go.text/transform"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

var (
	flvs, _ = regexp.Compile(`href="(.+?)"\s?target="_blank"\s?(class="link")?\s?onclick=["']_alert\(\);return false`)
	name, _ = regexp.Compile(`document.title\s=\s"([^"]+)"`)
	client  = &http.Client{}
)

func init() {
	client.Jar, _ = cookiejar.New(nil)
}

func Pretend(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.17 (KHTML, like Gecko) Chrome/24.0.1312.57 Safari/537.17")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Accept-Charset", "GBK,utf-8;q=0.7,*;q=0.3")
}

func Decode(s []byte) ([]byte, error) {
	f := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(f)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Detect(addr string) (string, []string, error) {

	req, err := http.NewRequest("GET", "http://www.flvcd.com/parse.php?format=super&kw="+url.QueryEscape(addr), nil)
	if err != nil {
		return "", nil, err
	}
	Pretend(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	file := name.FindSubmatch(data)
	prim, err := Decode(file[1])
	if err != nil {
		return "", nil, err
	}

	list := flvs.FindAllSubmatch(data, 64)
	if list == nil {
		return "", nil, errors.New("Not find video address")
	}
	each := make([]string, len(list))
	for i, m := range list {
		each[i] = string(m[1])
	}

	return string(prim), each, nil
}

func Download(name string, flvs []string) error {

	for i, url := range flvs {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		Pretend(req)
		req.Header.Set("Accept", "q=0.9,*/*;q=0.8")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		l, err := strconv.Atoi(resp.Header.Get("Content-Length"))
		if err != nil {
			return err
		}
		file, err := os.Create(name + "." + strconv.Itoa(i) + ".flv")
		if err != nil {
			return err
		}
		defer file.Close()

		buf := make([]byte, 4096)
		a, n := 0, 0
		fmt.Printf("\r[%.2fMB / %.2fMB]", float64(a)/1048576, float64(l)/1048576)
		for {
			n, err = resp.Body.Read(buf)
			a += n
			if err == nil {
				file.Write(buf[:n])
				fmt.Printf("\r[%.2fMB / %.2fMB]", float64(a)/1048576, float64(l)/1048576)
				continue
			}
			if err == io.EOF {
				file.Write(buf[:n])
				fmt.Printf("\r[%.2fMB / %.2fMB]", float64(a)/1048576, float64(l)/1048576)
				break
			}
			return err
		}
		fmt.Printf("\r\n")
	}
	return nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("usage : url\n")
		return
	}
	a, b, c := Detect(os.Args[1])
	if c != nil {
		fmt.Println(c)
		return
	}
	fmt.Println(a)
	c = Download(a, b)
	if c != nil {
		fmt.Println(c)
	}
}
