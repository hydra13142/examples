package main

import (
	"code.google.com/p/go.text/encoding/simplifiedchinese"
	"code.google.com/p/go.text/transform"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	tran   transform.Transformer
	fold   *regexp.Regexp
	page   *regexp.Regexp
	addr   *regexp.Regexp
	titl   *regexp.Regexp
	legal  *strings.Replacer
	wait   = &sync.WaitGroup{}
	pool   = make(chan int, 5)
	client *http.Client
)

func init() {
	fold, _ = regexp.Compile(`<td\salign="center"\s?><a\shref="([^"]+)"\starget="_blank"\s?>([^<>]+)</a>`)
	addr, _ = regexp.Compile(`src="([^"]+)"\s(?:border="0"\s)?(?:alt=""\s)?/>`)
	page, _ = regexp.Compile(`<a\shref="([^"]+)">下一页</a>`)
	titl, _ = regexp.Compile(`<h1>([^<>]+)</h1>`)
	tran = simplifiedchinese.GBK.NewDecoder()
	legal = strings.NewReplacer(
		`/`, `／`,
		`\`, `＼`,
		`<`, `＜`,
		`>`, `＞`,
		`:`, `：`,
		`?`, `？`,
		`*`, `＊`,
		`"`, `＂`,
		`|`, `｜`)
	cookie, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar: cookie,
	}
}

func repair(relate, current string) (absolute string) {
	var (
		a, b, c []string
		scheme  string
	)
	if strings.Contains(relate, "://") {
		return relate
	}
	a = strings.Split(current, "://")
	if len(a) > 1 {
		scheme = a[0] + "://"
	} else {
		scheme = ""
	}
	b = strings.Split(a[len(a)-1], "/")
	if relate[0] == '/' {
		return scheme + b[0] + relate
	}
	if b[len(b)-1] == "" {
		return current + relate
	}
	c = strings.Split(b[len(b)-1], ".")
	if len(c) == 1 {
		return current + "/" + relate
	}
	b[len(b)-1] = relate
	return scheme + strings.Join(b, "/")
}

func fname(addr string) string {
	a := strings.Split(addr, "/")
	return strings.FieldsFunc(a[len(a)-1], func(r rune) bool {
		return r == ';' || r == '?' || r == '#'
	})[0]
}

func readAll(r io.Reader) ([]byte, error) {
	r = transform.NewReader(r, tran)
	return ioutil.ReadAll(r)
}

func findFold(lnk string) error {
	resp, err := client.Get(lnk)
	if err != nil {
		return err
	}
	data, err := readAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	matched := fold.FindAllSubmatch(data, -1)
	for _, each := range matched {
		lnk, dir := string(each[1]), string(each[2])
		err := loadFold(lnk, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadFold(lnk, dir string) error {
	iter := make(chan string)
	over := make(chan int)
	go func() {
		for i := 1; ; i++ {
			select {
			case <-over:
				close(iter)
				close(over)
				return
			case iter <- fmt.Sprintf("%03d", i):
			}
		}
	}()
	dir = legal.Replace(dir)
	os.Mkdir(dir, os.ModeDir)
	for {
		err := func() error {
			fmt.Println(lnk)
			resp, err := client.Get(lnk)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			data, err := readAll(resp.Body)
			if err != nil {
				return err
			}

			matched := addr.FindSubmatch(data)
			if len(matched) >= 1 {
				wait.Add(1)
				go func(lnk, dir, name string) (err error) {
					pool <- 1
					for i := 0; i <= 9; i++ {
						err = loadPicture(lnk, dir, name)
						if err == nil {
							break
						}
					}
					fmt.Println(lnk)
					if err != nil {
						fmt.Println(err)
					}
					time.Sleep(time.Second * 10)
					<-pool
					wait.Done()
					return nil
				}(repair(strings.Join(strings.Split(string(matched[1]), "small"), ""), lnk), dir, <-iter)
			}
			matched = page.FindSubmatch(data)
			if len(matched) < 1 {
				return fmt.Errorf("Can't find next page")
			}
			lnk = string(matched[1])
			return nil
		}()
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	wait.Wait()
	over <- 0
	return nil
}

func loadPicture(lnk, dir, name string) (err error) {
	var (
		end  = make(chan int)
		data []byte
	)
	resp, err := client.Get(lnk)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	go func() {
		data, err = ioutil.ReadAll(resp.Body)
		end <- 1
	}()
	select {
	case <-time.After(time.Minute):
		return fmt.Errorf("Time out while downloading")
	case <-end:
	}
	if err != nil {
		return err
	}
	file, err := os.Create(dir + "/" + name + ".jpg")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func main() {
	/*
		var err error
		err = loadFold(`http://www.dm123.cn/pic/cg/2010-12-08/30470.html`, `総天然色妖怪美少女絵巻`)
		if err != nil {
			fmt.Println(err)
		}
		err = findFold(`http://www.dm123.cn/ecms/pic/cg/index.html`)
		if err != nil {
			fmt.Println(err)
		}
	*/
}
