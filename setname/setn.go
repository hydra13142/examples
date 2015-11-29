package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	begin int
	step  int
	width int
)

func init() {
	flag.IntVar(&begin, "b", 1, "iterator to begin")
	flag.IntVar(&step, "s", 1, "add each time to iterator")
	flag.IntVar(&width, "w", 3, "width trmatchedlate iterator to string")
	flag.Parse()
}

func main() {
	if flag.NArg() < 2 {
		fmt.Println(`usage: executable match-pattern name-pattern`)
		fmt.Println(`note: use '\$' as '$', '$0' is the iterator:`)
		flag.PrintDefaults()
		return
	}
	reg, err := regexp.Compile(flag.Arg(0))
	if err != nil {
		fmt.Println("syntax error in match-pattern")
		return
	}
	temp := strings.Join(flag.Args()[1:], " ")
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if path == "." {
			return nil
		}
		if err == os.ErrPermission {
			return nil
		}
		if err != nil {
			return err
		}
		matched := reg.FindStringSubmatchIndex(path)
		if matched == nil {
			return nil
		}
		media := fmt.Sprintf("%0*d", width, begin) + path
		begin += step
		matched[0], matched[1] = 0, width
		for i := 2; i < len(matched); i++ {
			matched[i] += width
		}
		result := string(reg.ExpandString(nil, temp, media, matched))
		fmt.Println(path, "=>\n\t", result)
		err = os.Rename(path, result)
		if err != nil {
			return err
		}
		if info.IsDir() && path != "." {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}
