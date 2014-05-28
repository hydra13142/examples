package main

import "fmt"

var iter chan int

func init() {
	iter = make(chan int)
	go func() {
		for i := 2; i < 20000; i++ {
			iter <- i
		}
		close(iter)
	}()
}

func attach(d int, curr <-chan int) chan int {
	next := make(chan int)
	go func() {
		for {
			x, ok := <-curr
			if !ok {
				break
			}
			if x%d != 0 {
				next <- x
			}
		}
		close(next)
	}()
	return next
}

func main() {
	curr := iter
	for {
		d, ok := <-curr
		if !ok {
			break
		}
		fmt.Println(d)
		curr = attach(d, curr)
	}
}
