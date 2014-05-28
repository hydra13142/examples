package main

import (
	"fmt"
	"os"
)

func main() {
	for i, par := range os.Args {
		fmt.Printf("[%02d]: %s\n", i, par)
	}
	os.Args = ParseCmd(`cmdline -w 3 .+\.jpg "A{{index . 0 | print \"%s\"}}.jpg"`)
	for i, par := range os.Args {
		fmt.Printf("[%02d]: %s\n", i, par)
	}
}

// 模拟实现dos命令行的参数解析函数
func ParseCmd(s string) []string {
	var (
		ans         = make([]string, 0, 4)
		med         = make([]byte, 0, 12)
		sp, qt byte = 0, 0
	)
loop:
	for i := 0; i < len(s); {
		switch sp {
		case 0:
			switch s[i] {
			case '\r', '\n': // 换行表示命令结束
				break loop
			case ' ', '\t': // 空格分割参数
				sp = 1
			case '\'', '"', '`': // 引号用于提供包括空格/换行的参数
				qt = s[i]
				sp = 2
			case '\\': // 用于转义
				if i++; i < len(s) {
					if s[i] == '\r' || s[i] == '\n' { // 让较长的命令可以包括多行
						i++
						if i < len(s) && s[i] != s[i-1] && (s[i] == '\r' || s[i] == '\n') {
							i++
						}
						sp = 1
					} else { // 对三种引号进行转义
						if s[i] != '\'' && s[i] != '"' && s[i] != '`' {
							med = append(med, '\\')
						}
						med = append(med, s[i])
					}
				}
			default:
				med = append(med, s[i])
			}
			i++
		case 1: // 忽略空格
			if len(med) > 0 { // 加入已读取的参数
				ans = append(ans, string(med))
				med = make([]byte, 0, 12)
			}
			for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
				i++
			}
			sp = 0
		case 2: // 引号内的读取
			for ; i < len(s) && s[i] != qt; i++ {
				if s[i] == '\\' {
					if i++; i < len(s) { // 引号内，'\'只用于该引号的转义
						if s[i] != qt {
							med = append(med, '\\')
						}
						med = append(med, s[i])
					}
				} else {
					med = append(med, s[i])
				}
			}
			i, sp = i+1, 0
		}
	}
	if len(med) > 0 { // 加入最后一个参数
		ans = append(ans, string(med))
	}
	return ans
}
