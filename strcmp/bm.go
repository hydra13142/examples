package main

import "fmt"

// bm搜索算法
func BM(s, r string) int {
	l := len(r)
	// n[i]表示byte(i)在s中最后一次出现时距离s末端的字符数
	// 从没出现过的字符都设置对应的值为len(s)
	nc := func(s string) []int {
		n := make([]int, 256)
		for i := 0; i < 256; i++ {
			n[i] = l
		}
		for i := 0; i < l; i++ {
			n[int(s[i])] = l - i - 1
		}
		return n
	}(r)
	// n[i]表示s[i:]在s中前次出现时其末端距离s末端的字符数
	// 没有多次出现过的s[i:]都设置对应的值为len(s)
	ns := func(s string) []int {
		n := make([]int, l+1)
		for i, j := l-1, 1; i >= 0; i-- {
			// 已知后面的字符全部匹配
			if s[i] == s[i-j] {
				n[i] = j
				continue
			}
			// 增加偏移量寻找新的匹配
			for j++; j <= i; j++ {
				k, t := i-j, i
				for t < l && s[k] == s[t] {
					k, t = k+1, t+1
				}
				if t >= l {
					break
				}
			}
			// 后缀长度+偏移量不可能超过总长度
			if j > i {
				for ; i >= 0; i-- {
					n[i] = l
				}
				break
			}
			// 出于简化算法添加
			n[i] = j
		}
		n[l] = l
		return n
	}(r)
	// 进行搜索
	for i := l - 1; i < len(s); {
		a, b := i, l-1
		for b >= 0 && s[a] == r[b] {
			a, b = a-1, b-1
		}
		if b < 0 {
			return i - l + 1
		}
		x, y := nc[s[a]], ns[b+1]
		if a+x > i+y {
			i = a + x
		} else {
			i = i + y
		}
	}
	return -1
}

func main() {
	fmt.Println(BM("caabccd", "aabcc"))
}
