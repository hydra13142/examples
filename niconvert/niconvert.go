package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func tolerate(str string) string {
	return strings.Map(
		func(c rune) rune {
			switch c {
			case '\\':
				return '＼'
			case '|':
				return '｜'
			case '/':
				return '／'
			case '<':
				return '＜'
			case '>':
				return '＞'
			case '"':
				return '＂'
			case ':':
				return '：'
			case '*':
				return '＊'
			case '?':
				return '？'
			}
			return c
		}, strings.TrimSpace(str))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage : url\n")
		return
	}
	proxy, _ := url.Parse("http://127.0.0.1:8087")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	query := url.Values{
		"url": {os.Args[1]},
	}
	data := url.Values{
		"font_name":     {"微软雅黑"},
		"font_size":     {"40"},
		"video_width":   {"640"},
		"video_height":  {"480"},
		"line_count":    {"12"},
		"bottom_margin": {"0"},
		"tune_seconds":  {"0"},
		"url":           {os.Args[1]},
	}
	URL := "http://niconvert.appspot.com/?" + query.Encode()
	resp, err := client.Get(URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	title, _ := regexp.Compile(`视频标题：</label><a\shref="[^"]+">(.+?)</a>`)
	matched := title.FindSubmatch(page)
	if matched == nil {
		fmt.Println("Cannot find title")
		return
	}
	name := tolerate(string(matched[1]))

	resp, err = client.PostForm(URL, data)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	ass, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Create(name + ".ass")
	if err != nil {
		fmt.Println(err)
		return
	}
	file.Write(ass)
	file.Close()
}
