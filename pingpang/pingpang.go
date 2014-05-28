package main

import (
	"fmt"
	"net"
	"sync"
)

var w sync.WaitGroup

func listen() {
	defer w.Done()
	addr, _ := net.ResolveTCPAddr("tcp4", ":12345")
	serv, _ := net.ListenTCP("tcp", addr)
	conn, _ := serv.Accept()
	w.Add(1)
	go func() {
		defer w.Done()
		defer conn.Close()
		x := make([]byte, 8192)
		for i := 0; i < 10; i++ {
			n, _ := conn.Read(x)
			fmt.Println(string(x[:n]))
			conn.Write([]byte("[----pang]"))
		}
	}()
}

func send() {
	defer w.Done()
	addr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:12345")
	conn, _ := net.DialTCP("tcp", nil, addr)
	defer conn.Close()
	x := make([]byte, 8192)
	for i := 0; i < 10; i++ {
		conn.Write([]byte("[ping----]"))
		n, _ := conn.Read(x)
		fmt.Println(string(x[:n]))
	}
}

func main() {
	w.Add(2)
	go listen()
	go send()
	w.Wait()
}
