package main

import (
	"fmt"
	"strings"
)

func Verbose(s string) string {
	ss := strings.FieldsFunc(s, func(r rune) bool { return r == '\r' || r == '\n' })
	for i, t := range ss {
		for j, k := 0, 0; ; j += k + 1 {
			k = strings.IndexByte(t[j:], '#')
			if k < 0 {
				break
			} else if k == 0 {
				ss[i] = t[:j]
				break
			} else if t[j+k-1] != '\\' {
				ss[i] = t[:j+k]
				break
			}
		}
	}
	return strings.Join(strings.Fields(strings.Join(ss, "")), "")
}

func main() {
	fmt.Println(Verbose(`
    M{0,3}?                     # thousands - M{0,3}
    (C(?:[DM]|C{0,2})|DC{0,3})? # hundreds  - 0-300 (C{0,3}), 400 (CD), 900 (CM), or 500-800 (DC{0,3})
    (X(?:[LC]|X{0,2})|LX{0,3})? # tens      - 0- 30 (X{0,3}),  40 (XL),  90 (XC), or  50- 80 (LX{0,3})
    (I(?:[VX]|I{0,2})|VI{0,3})? # ones      - 0-  3 (I{0,3}),   4 (IV),   9 (IX), or   5-  8 (VI{0,3})
	`))
	// M{0,3}?(C(?:[DM]|C{0,2})|DC{0,3})?(X(?:[LC]|X{0,2})|LX{0,3})?(I(?:[VX]|I{0,2})|VI{0,3})?
	fmt.Println(Verbose(`
	([-+]?)             # sign
	(\d+)               # integer
	(?:\.(\d+))?        # fraction
	(?:[eE]([-+]?\d+))? # exponent
	`))
	// ([-+]?)(\d+)(?:\.(\d+))?(?:[eE]([-+]?\d+))?
}
