package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	pool            = &sync.Pool{}
	regex           []*regexp.Regexp
	proxy           []*url.URL
)

var disableRedirect = fmt.Errorf("Disable Redirect")

func init() {
	_, regex, proxy, _ = parseRule(loadRule("proxy.txt"))
	pool.New = func() interface{} {
		client := new(http.Client)
		client.Transport = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				for i, rule := range regex {
					if rule.MatchString(req.URL.Host) {
						return proxy[i], nil
					}
				}
				return nil, nil
			},
		}
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return disableRedirect
		}
		return client
	}
}
func main() {
	laddr, _ := net.ResolveTCPAddr("tcp", ":8080")
	listener, _ := net.ListenTCP("tcp", laddr)
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatalf("Accept: %4096v\n", err)
		}
		go func(conn *net.TCPConn) {
			defer conn.Close()
			defer func() {
				err := recover()
				if err != nil {
					log.Fatalf("Panic: %4096v\n", err)
					var c byte
					fmt.Scanf("%c\n", &c)
				}
			}()
			reader := bufio.NewReader(conn)
			codeline, headers, err := readHeader(reader)
			if err != nil {
				// log.Println("ReadHeader:", err)
				return
			}
			if pieces := strings.Fields(codeline); pieces[0] == "CONNECT" {
				serveHTTPS(conn, reader, codeline, headers, pieces[1])
			} else {
				serveHTTP(conn, reader, codeline, headers)
			}
		}(conn)
	}
}
func loadRule(filename string) [][]string {
	var addr, rule string
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	rules := make([][]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] == ' ' || line[0] == '\t' {
			line = strings.TrimSpace(line)
			if len(line) > 0 {
				rule = line
				rules = append(rules, []string{rule, addr})
			}
		} else {
			line = strings.TrimSpace(line)
			if len(line) > 0 {
				addr = line
			}
		}
	}
	return rules
}
func parseRule(rules [][]string) (l int, r []*regexp.Regexp, p []*url.URL, err error) {
	l = len(rules)
	r = make([]*regexp.Regexp, l)
	p = make([]*url.URL, l)
	for i, item := range rules {
		proxy, err := url.Parse(item[1])
		if err != nil {
			return 0, nil, nil, fmt.Errorf("URL syntax error at index %d", i)
		}
		regex, err := regexp.Compile(item[0])
		if err != nil {
			return 0, nil, nil, fmt.Errorf("regex syntax error at index %d", i)
		}
		r[i], p[i] = regex, proxy
	}
	return l, r, p, nil
}
func readHeader(r *bufio.Reader) (codeline, headers string, err error) {
	line, ispr, err := r.ReadLine()
	if ispr || err != nil {
		return "", "", err
	}
	if len(line) == 0 {
		return "", "", fmt.Errorf("not find code-line")
	}
	codeline = string(line) + "\r\n"
	for {
		line, ispr, err = r.ReadLine()
		if err != nil {
			return "", "", err
		}
		headers += string(line)
		if !ispr {
			headers += "\r\n"
		}
		if len(line) == 0 {
			break
		}
	}
	return codeline, headers, nil
}
func serveHTTPS(conn *net.TCPConn, reader *bufio.Reader, codeline, headers, piece string) {
	this := func(target string) string {
		for i, rule := range regex {
			if rule.MatchString(target) {
				return proxy[i].Host
			}
		}
		return target
	}(piece)
	raddr, err := net.ResolveTCPAddr("tcp", this)
	for i := 0; i < 5 && err != nil; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		raddr, err = net.ResolveTCPAddr("tcp", this)
	}
	if err != nil {
		log.Println("Parse Https:", codeline, piece, err)
		return
	}
	tcpconn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		// log.Println("Connect:", err)
		return
	}
	defer tcpconn.Close()
	if this == piece {
		conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
	} else {
		tcpconn.Write([]byte(codeline + headers))
	}
	go func() {
		if _, err := tcpconn.ReadFrom(reader); err != nil {
			// log.Println("ReadFrom Left:", err)
			return
		}
	}()
	if _, err = conn.ReadFrom(tcpconn); err != nil {
		// log.Println("ReadFrom Right:", err)
		return
	}
}
func serveHTTP(conn *net.TCPConn, reader *bufio.Reader, codeline, headers string) {
	req, err := http.ReadRequest(bufio.NewReader(io.MultiReader(strings.NewReader(codeline+headers), reader)))
	if err != nil {
		// log.Println("ReadRequest:", err)
		return
	}
	req.RequestURI = ""
	req.RemoteAddr = ""
	req.Header.Del("Proxy-Connection")
	client := pool.Get().(*http.Client)
	defer pool.Put(client)
	resp, err := client.Do(req)
	if req.Method == "GET" && err != nil {
		for i := 0; i < 5 && err != nil; i++ {
			time.Sleep(time.Duration(i*2) * time.Second)
			resp, err = http.DefaultClient.Do(req)
		}
	}
	if err != nil {
		under, ok := err.(*url.Error)
		if !ok || under.Err != disableRedirect {
			log.Println("DoRequest:", err)
			return
		}
	}
	resp.Header.Del("Connection")
	content := resp.Header.Get("Content-Type")
	piece := strings.Split(req.URL.Path, "/")
	name := piece[len(piece)-1]
	if l := len(name); l >= 5 {
		if name[l-4:l] == ".hlv" || name[l-4:l] == ".flv" || name[l-4:l] == ".mp4" {
			goto save
		}
		if l >= 6 && name[l-5:l] == ".letv" {
			goto save
		}
	}
	if len(content) < 5 || (content[:5] != "video" && content[:3] != "flv") {
		if err = resp.Write(conn); err != nil {
			// log.Println("WriteResponse:", err)
		}
		return
	}
	if name == "smile" {
		req.ParseForm()
		name = "sm" + strings.Split(req.FormValue("m"), ".")[0]
	}
	if strings.Contains(content, "mp4") {
		if !strings.HasSuffix(name, ".mp4") {
			name += ".mp4"
		}
	}
	if strings.Contains(content, "flv") {
		if !strings.HasSuffix(name, ".flv") {
			name += ".flv"
		}
	}
save:
	log.Printf("LOAD %s\n", name)
	b := new(bytes.Buffer)
	if err = resp.Write(io.MultiWriter(b, conn)); err != nil {
		// log.Println("WriteResponse:", err)
		return
	}
	file, _ := os.OpenFile(name, os.O_APPEND|os.O_CREATE, 0x0666)
	file.Write(bytes.SplitN(b.Bytes(), []byte("\r\n\r\n"), 2)[1])
	file.Close()
	log.Printf("OVER %s\n", name)
}
