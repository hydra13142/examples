package main

import "fmt"

// kmp字符串搜索算法
func KMP(s, r string) int {
	l := len(r)
	// 成员n[i]表示既是s[:i]真后缀又是s前缀的最长序列的长度。
	// 真后缀指非自身的非空后缀。如不存在，则置该成员值为0。
	n := func(s string) []int {
		n := make([]int, l)
		// 从n[2]开始算起
		for i, j := 2, 0; i < l; {
			// 已知前面的字符全部匹配
			if s[i-1] == s[j] {
				j++
				n[i] = j
				i++
				continue
			}
			if j == 0 {
				// 申请的n会全部初始化为零
				i++
				continue
			}
			// 类似后缀指针的功用
			j = n[j]
		}
		return n
	}(r)
	// 进行搜索
	i, j := 0, 0
	for i+l < j+len(s) && j < l {
		if s[i] == r[j] {
			i, j = i+1, j+1
			continue
		}
		if j == 0 {
			i++
		} else {
			j = n[j]
		}
	}
	if j == l {
		return i - l
	}
	return -1
}

func main() {
	fmt.Println(KMP("caabccd", "aabcc"))
}
