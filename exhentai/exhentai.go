package main

import (
	"bytes"
	"fmt"
	"github.com/hydra13142/http/glype"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
#include <windows.h>
#include <stdlib.h>
char *cpbd()
{
	OpenClipboard(NULL);
	HGLOBAL hMem=GetClipboardData(CF_TEXT);
	char* lpStr = (char*)GlobalLock(hMem);
	CloseClipboard();
	return lpStr;
}
*/
import "C"

type None struct{}

var (
	// 获取漫画的标题
	title, _ = regexp.Compile(`<h1\sid="gn">(.*?)</h1><h1\sid="gj">(.*?)</h1>`)
	// 获取第一页漫画的地址
	first, _ = regexp.Compile(`<a\shref="([^"]+)"><img\salt="0{0,2}1"`)
	// 获取下一页漫画和当前漫画图片的地址
	addr, _ = regexp.Compile(`href="([^"]+)"><img\s(?:id="img"\s)?src="([^"]+)"\sstyle="`)
	// 获取当前页漫画图片的名称和大小
	attr, _ = regexp.Compile(`[^a]><div>(.+?)\s::\s\d+\sx\s\d+\s::\s(.+?)</div>`)
	// 获取当前页码和总页数
	curr, _ = regexp.Compile(`(\d+)</span>\s/\s<span>(\d+)`)

	glype14 = []string{}
	dailila = []string{}
	mission = []string{}
	records = map[string]None{}
)

// 使用cgo调用mingw的库函数，来读取windows的剪贴板文本数据
func clipboard() string {
	i := C.cpbd()
	s := C.GoString(i)
	return s
}

func getList(f string) []string {
	file, err := os.Open(f)
	if err != nil {
		return nil
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil
	}
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	list := make([]string, len(lines))
	for i, l := range lines {
		list[i] = string(bytes.TrimSpace(l))
	}
	return list
}

func getHistory(f string) error {
	file, err := os.OpenFile(f, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	find, _ := regexp.Compile(`(find|fail|done)\s:\shttp://exhentai\.org/(g/\d+/\w+/)`)
	matched := find.FindAllSubmatch(data, -1)
	for _, every := range matched {
		switch string(every[1]) {
		case "find":
			records[string(every[2])] = None{}
		case "done":
			delete(records, string(every[2]))
		case "fail":
		}
	}
	for one, _ := range records {
		mission = append(mission, one)
		fmt.Print("\a")
	}
	return nil
}

func setDir(dir string) string {
	dir = strings.Map(
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
		}, strings.TrimSpace(dir))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModeDir)
	}
	return dir
}

func getSize(s string) float64 {
	u := strings.Fields(s)
	i, _ := strconv.ParseFloat(u[0], 64)
	switch u[1][0] {
	case 'K', 'k':
		return (i * 1024)
	case 'M', 'm':
		return (i * 1024 * 1024)
	default:
		return (i)
	}
}

func Login(client *glype.Client) error {
	user := url.Values{
		"ipb_login_username": {"snake117"},
		"ipb_login_password": {"6ys97pt"},
		"ipb_login_submit":   {"Login!"},
	}
	resp, err := client.Get("http://g.e-hentai.org/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp, err = client.Get("http://e-hentai.org/bounce_login.php?b=d&bt=1-6")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp, err = client.PostForm("http://e-hentai.org/bounce_login.php?b=d&bt=1-6", user)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp, err = client.Get("http://g.e-hentai.org/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	client.Reset()
	resp, err = client.Get("http://exhentai.org/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(data) < 1024 {
		return fmt.Errorf("Your IP is banned")
	}
	return nil
}

func Hentai(web, pxy, adr string) error {
	client, err := glype.New(web, pxy)
	if err != nil {
		return err
	}
	if strings.Contains(adr, "exhentai") {
		err = Login(client)
		if err != nil {
			return err
		}
	} else {
		err = func() error {
			resp, err := client.Get("http://g.e-hentai.org/")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	// 通过网页代理访问漫画的index页
	resp, err := client.Get(adr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("Succeed to load index page\n")
	// 获取漫画的标题并创建文件夹
	matched := title.FindSubmatch(data)
	if len(matched) == 0 {
		return fmt.Errorf("Cannot find title")
	}
	dir := string(matched[1])
	if len(matched[2]) != 0 {
		dir = string(matched[2])
	}
	dir = setDir(dir)
	fmt.Println(dir)
	// 获取漫画的第一页漫画地址
	matched = first.FindSubmatch(data)
	if len(matched) == 0 {
		return fmt.Errorf("Cannot find first page")
	}
	// 单线程顺序下载漫画图片
	Page := glype.Fix(adr, string(matched[1]))
	for i := 0; ; i++ {
		// 累计10次下载失败，退出程序
		if i == 10 {
			return fmt.Errorf("Failed to load 10 times")
		}
		Next, err := func() (string, error) {
			resp, err = client.Get(Page)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			// 获取下一页漫画地址和当前页漫画的图片地址、图片名
			matched = addr.FindSubmatch(data)
			if len(matched) == 0 {
				return "", fmt.Errorf("Cannot find image address")
			}
			Next := string(matched[1])
			Addr := string(matched[2])
			matched = attr.FindSubmatch(data)
			if len(matched) == 0 {
				return "", fmt.Errorf("Cannot find image name")
			}
			Name := string(matched[1])
			Size := string(matched[2])
			// 获取当前页的页码以及漫画的总页数
			matched = curr.FindSubmatch(data)
			if len(matched) == 0 {
				return "", fmt.Errorf("Cannot find current number")
			}
			Curr := string(matched[1]) + " / " + string(matched[2])
			// 下载图片并且保存在新建的目录中
			filename := dir + "/" + Name
			if _, err = os.Stat(filename); err == nil {
				fmt.Printf("%s is already exist (%s)\n", Name, Curr)
				return glype.Fix(Page, Next), nil
			}
			resp, err = client.Get(glype.Fix(Page, Addr))
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			if strings.Split(resp.Header.Get("Content-Type"), "/")[0] != "image" {
				return "", fmt.Errorf("404 File Not Found")
			}
			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			if float64(len(data))*1.01 < getSize(Size) {
				return "", fmt.Errorf("Load image uncompleted")
			}
			file, err := os.Create(filename)
			if err != nil {
				return "", err
			}
			defer file.Close()
			_, err = file.Write(data)
			if err != nil {
				defer os.Remove(filename)
				return "", err
			}
			fmt.Printf("Succeed to load %s (%s)\n", Name, Curr)
			// 返回下一页地址
			return glype.Fix(Page, Next), nil
		}()
		// 如果出现错误，则重新尝试下载
		if err != nil {
			fmt.Println(err)
			if err.Error() == "404 File Not Found" {
				return err
			} else {
				time.Sleep(time.Millisecond * time.Duration(i*500+500))
				continue
			}
		}
		// 如果下一页地址和当前页地址相同，则表明已到结尾，退出
		if Page == Next {
			break
		}
		// 如果返回可用地址，说明当前页下载成功，换下一页下载
		Page = Next
		i = 0
	}
	fmt.Println("OVER")

	return nil
}

func init() {
	glype14 = getList("glypes.txt")
	if glype14 == nil {
		fmt.Println("Can't find glypes.txt")
		os.Exit(1)
	}
	dailila = getList("dailila.txt")
	if dailila == nil {
		fmt.Println("Can't find dailila.txt")
		os.Exit(1)
	}
	if getHistory("fatal.txt") != nil {
		return
	}
	file, _ := os.OpenFile("fatal.txt", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	for _, one := range mission {
		fmt.Fprintln(file, "find : "+"http://exhentai.org/"+one)
	}
	defer file.Close()
}

func main() {

	var (
		err error
		one string
	)

	only := make(chan int, 1)
	done := 0

	go func() {
		find, _ := regexp.Compile(`e[-x]hentai\.org/(g/\d+/\w+/)`)
		for {
			data := clipboard()
			matched := find.FindAllStringSubmatch(data, -1)
			if len(matched) != 0 {
				only <- 1
				file, _ := os.OpenFile("fatal.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
				for _, every := range matched {
					one := every[1]
					_, ok := records[one]
					if !ok {
						records[one] = None{}
						mission = append(mission, one)
						fmt.Fprintln(file, "find : "+"http://exhentai.org/"+one)
						fmt.Print("\a")
					}
				}
				file.Close()
				<-only
			}
			time.Sleep(time.Millisecond * 250)
		}
	}()

	fmt.Println("Start to Download...")
	for {
		only <- 1
		if len(mission) == 0 {
			if done == 0 {
				fmt.Println("Waiting...")
				done = 1
			}
			<-only
			time.Sleep(time.Second * 2)
			continue
		} else {
			done = 0
			one = mission[0]
			mission = mission[1:]
			<-only
		}
		t := int(time.Now().Unix())
		for i := -len(glype14); i < len(dailila); i++ {
			if i < 0 {
				err = Hentai(glype14[(t+i)%len(glype14)], "http://127.0.0.1:8087/", "http://g.e-hentai.org/"+one)
			} else {
				err = Hentai(dailila[(t+i)%len(dailila)], "http://127.0.0.1:8087/", "http://exhentai.org/"+one)
			}
			if err == nil {
				break
			}
		}
		only <- 1
		file, _ := os.OpenFile("fatal.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err == nil {
			fmt.Fprintln(file, "done : "+"http://exhentai.org/"+one)
		} else {
			fmt.Fprintln(file, "fail : "+"http://exhentai.org/"+one)
		}
		file.Close()
		<-only
	}
}
