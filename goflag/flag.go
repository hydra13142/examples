package main

import (
	"flag"
	"fmt"
)

func main() {
	var l int
	flag.IntVar(&l, "a", 0, "anumber")
	s := flag.String("b", "<nil>", "a string")
	flag.Parse()
	fmt.Println("number: ", l)
	fmt.Println("string: ", *s)
	fmt.Println("others: ")
	t := flag.NArg()
	for i := 0; i < t; i++ {
		fmt.Println("\t", flag.Arg(t))
	}
}
