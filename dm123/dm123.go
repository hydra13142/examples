package main

import (
	"code.google.com/p/mahonia"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
)

var (
	enc mahonia.Decoder = mahonia.NewDecoder("gbk")
	lck chan int        = make(chan int, 5)
	wgs sync.WaitGroup
)

func loadindex(dir string, idx string) {
	pho, err := regexp.Compile(`href="([^"]+)"><img\sborder="0"`)
	if err != nil {
		return
	}
	nxt, err := regexp.Compile(`href="([^"]+)">下一页`)
	if err != nil {
		return
	}
	nme, err := regexp.Compile(`\w+\.\w+$`)
	if err != nil {
		return
	}
	cln := http.Client{}
	for {
		resp, err := cln.Get(idx)
		if err != nil {
			fmt.Println("Can't open index page")
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Can't load index page content")
			return
		}
		cvtd := enc.ConvertString(string(data))
		out := pho.FindStringSubmatch(cvtd)
		if len(out) > 0 {
			go func(url string) {
				wgs.Add(1)
				get := <-lck
				defer func() {
					wgs.Done()
					lck <- get
				}()
				fmt.Println(url)
				name := nme.FindString(url)
				if name == "" {
					fmt.Println("Can't find image file")
					return
				}
				for i := 0; i < 10; i++ {
					cln := http.Client{}
					resp, err := cln.Get(url)
					if err != nil {
						fmt.Println("Can't connect image file")
						continue
					}
					defer resp.Body.Close()
					data, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println("Can't load image file")
						continue
					}
					file, err := os.Create(dir+"/"+name)
					if err != nil {
						fmt.Println("Can't save image file")
						break
					}
					defer file.Close()
					fmt.Println(name)
					file.Write(data)
					break
				}
			}("http://www.dm123.cn" + out[1])
		}
		out = nxt.FindStringSubmatch(cvtd)
		if len(out) <= 0 {
			break
		}
		idx = out[1]
	}
}

func main() {
	for i := 0; i < 5; i++ {
		lck <- i
	}
	os.Mkdir("47888", os.ModeDir)
	loadindex("47888", "http://www.dm123.cn/pic/cg/2013-08-24/47888.html")
	wgs.Wait()
	close(lck)
}
