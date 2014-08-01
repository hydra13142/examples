package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	cmd *exec.Cmd
	mks bool
)

func main() {
	flag.BoolVar(&mks, "b", false, " to build, not to run")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("usage a code file")
		return
	}
	part := strings.Split(flag.Arg(0), ".")
	each := len(part)
	if each <= 1 {
		fmt.Println("cannot get file type")
		return
	}
	attr := part[each-1]
	obey := strings.Join(part[:each-1], ".") + `.exe`
	if mks {
		switch attr {
		case "c":
			cmd = exec.Command("gcc", "-O3", "-o", obey, flag.Arg(0))
		case "cpp":
			cmd = exec.Command("g++", "-O3", "-o", obey, flag.Arg(0))
		case "go":
			cmd = exec.Command("gofmt", "-w", flag.Arg(0))
			cmd.Run()
			cmd = exec.Command("go", "build", flag.Arg(0))
		default:
			fmt.Println("cannot build this file")
			return
		}
	} else {
		switch attr {
		case "c":
			cmd = exec.Command("gcc", "-W", "-Wall", "-o", obey, flag.Arg(0))
		case "cpp":
			cmd = exec.Command("g++", "-W", "-Wall", "-o", obey, flag.Arg(0))
		case "go":
			cmd = exec.Command("gofmt", "-w", flag.Arg(0))
			cmd.Run()
			cmd = exec.Command("go", "run", flag.Arg(0))
		case "lsp":
			cmd = exec.Command("lisp", flag.Arg(0))
		case "py":
			cmd = exec.Command("python", flag.Arg(0))
		default:
			fmt.Println("cannot run this file")
			return
		}
		if attr == "c" || attr == "cpp" {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			cmd = exec.Command(obey)
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = filepath.Dir(flag.Arg(0))
	cmd.Run()
	if !mks {
		if attr == "c" || attr == "cpp" {
			os.Remove(obey)
		}
	}
}
