// 将bilibili的xml弹幕文件转换为ass字幕文件。
// xml文件中，弹幕的格式如下：
// <d p="32.066,1,25,16777215,1409046965,0,017d3f58,579516441">地板好评</d>
// p的属性为时间、弹幕类型、字体大小、字体颜色、创建时间、？、创建者ID、弹幕ID。
// p的属性中，后4项对ass字幕无用，舍弃。被<d>和</d>包围的是弹幕文本。
// 只处理右往左、上现隐、下现隐三种类型的普通弹幕。
package main

import (
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ass文件的头部
const header = "[Script Info]\r\nScriptType: v4.00+\r\nCollisions: Normal\r\nplayResX: 640\r\nplayResY: 360\r\n\r\n[V4+ Styles]\r\n" +
	"Format: Name, Fontname, Fontsize, primaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\r\n" +
	"Style: Default, Microsoft YaHei, 28, &H00FFFFFF, &H00FFFFFF, &H00000000, &H00000000, 0, 0, 0, 0, 100, 100, 0.00, 0.00, 1, 1, 0, 2, 10, 10, 10, 0\r\n" +
	"\r\n[Events]\r\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\r\n"

// 正则匹配获取弹幕原始信息
var (
	biliz = regexp.MustCompile(`^http://www.bilibili.com/video/av\d+/?$`)
	title = regexp.MustCompile(`<title>(.+?)_[^_]+_[^_]+_bilibili_哔哩哔哩弹幕视频网</title>`)
	addrs = regexp.MustCompile(`cid=(\d+)&`)
	bili  = regexp.MustCompile(`<d\sp="([\d\.]+),([145]),(\d+),(\d+),\d+,\d+,\w+,\d+">([^<>]+?)</d>`)
	acfn  = regexp.MustCompile(`\{"c"\:"([\d\.]+),(\d+),([145]),(\d+),[\w,-]+","m"\:"([^"]+)"\}`)
)

// 错误信息
var (
	errNoXML   = fmt.Errorf("Didn't find the xml address")
	errNoTitle = fmt.Errorf("Didn't find the title")
)

// 用来保管弹幕的信息
type Danmu struct {
	text  string
	time  float64
	kind  byte
	size  int
	color int
}

// 使[]Danmu实现sort.Interface接口，以便排序
type Danmus []Danmu

func (d Danmus) Len() int {
	return len(d)
}
func (d Danmus) Less(i, j int) bool {
	return d[i].time < d[j].time
}
func (d Danmus) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// 返回文本的长度，假设ascii字符都是0.5个字长，其余都是1个字长
func length(s string) float64 {
	l := 0.0
	for _, r := range s {
		if r < 127 {
			l += 0.5
		} else {
			l += 1
		}
	}
	return l
}

// 生成时间点的ass格式表示：`0:00:00.00`
func timespot(f float64) string {
	h, f := math.Modf(f / 3600)
	m, f := math.Modf(f * 60)
	return fmt.Sprintf("%d:%02d:%05.2f", int(h), int(m), f*60)
}

// 读取文件并获取其中的弹幕
func open(r io.Reader) ([]Danmu, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	bi := bili.FindAllSubmatch(data, -1)
	ac := acfn.FindAllSubmatch(data, -1)
	dm := make([]Danmu, len(bi)+len(ac))
	da := dm[len(bi):]
	for i, s := range bi {
		dm[i].time, _ = strconv.ParseFloat(string(s[1]), 64)
		dm[i].kind = s[2][0] - '0'
		dm[i].size, _ = strconv.Atoi(string(s[3]))
		bgr, _ := strconv.Atoi(string(s[4]))
		dm[i].color = ((bgr >> 16) & 255) | (bgr & (255 << 8)) | ((bgr & 255) << 16)
		dm[i].text = string(s[5])
	}
	for i, s := range ac {
		da[i].time, _ = strconv.ParseFloat(string(s[1]), 64)
		da[i].kind = s[3][0] - '0'
		da[i].size, _ = strconv.Atoi(string(s[4]))
		bgr, _ := strconv.Atoi(string(s[2]))
		da[i].color = ((bgr >> 16) & 255) | (bgr & (255 << 8)) | ((bgr & 255) << 16)
		da[i].text = string(s[5])
	}
	return dm, nil
}

// 将弹幕排布并写入w，采用的简单的固定存在时间、最小重叠时间的排布算法
func save(w io.Writer, dans []Danmu) {
	// 将屏幕划分10像素为1行，对应3种字体（行宽20，30，40）
	// 其值表示上一个字幕在该时间点结束
	p1 := make([]float64, 36)
	p2 := make([]float64, 36)
	p3 := make([]float64, 36)
	// 选取连续行中时间最后的
	max := func(x []float64) float64 {
		i := x[0]
		for _, j := range x[1:] {
			if i < j {
				i = j
			}
		}
		return i
	}
	// 将连续行设置为同一时间点
	set := func(x []float64, f float64) {
		for i, _ := range x {
			x[i] = f
		}
	}
	// 找出一个有倾向的、字幕重叠时间最短的行
	find := func(p []float64, f float64, i, d int) int {
		i = (i/d + 1) * d % 36
		m, k := f+10000, 0
		for j := 0; j < 36; j += d {
			t := (i + j) % 36
			if n := max(p[t : t+d]); n <= f {
				k = t
				break
			} else if m > n {
				k = t
				m = n
			}
		}
		return k
	}
	// 对每一条弹幕都进行排布
	for _, dan := range dans {
		s, l := "", length(dan.text)
		if l == 0 {
			continue
		}
		switch {
		case dan.size < 25: // 小字体占据2行
			dan.size, l, s = 2, l*18, "\\fs18"
		case dan.size == 25: // 中字体占据3行
			dan.size, l = 3, l*28
		case dan.size > 25: // 大字体占据4行
			dan.size, l, s = 4, l*38, "\\fs38"
		}
		// 字体色彩：\c&HRRGGBB
		if dan.color != 0x00FFFFFF {
			s += fmt.Sprintf("\\c&H%06X", dan.color)
		}
		switch dan.kind {
		case 1: // 右往左
			j := find(p1, dan.time, 0, dan.size)
			set(p1[j:j+dan.size], dan.time+8)
			h := (j+dan.size)*10 - 1
			s += fmt.Sprintf("\\move(%d,%d,%d,%d)", 640+int(l/2), h, -int(l/2), h)
			fmt.Fprintf(w, "Dialogue: 1,%s,%s,Default,,0000,0000,0000,,{%s}%s\r\n",
				timespot(dan.time+0),
				timespot(dan.time+8), s, dan.text)
		case 4: // 下现隐
			j := find(p2, dan.time, 35, dan.size)
			set(p2[j:j+dan.size], dan.time+4)
			s += fmt.Sprintf("\\pos(%d,%d)", 320, (36-j)*10-1)
			fmt.Fprintf(w, "Dialogue: 2,%s,%s,Default,,0000,0000,0000,,{%s}%s\r\n",
				timespot(dan.time+0),
				timespot(dan.time+4), s, dan.text)
		case 5: // 上现隐
			j := find(p3, dan.time, 35, dan.size)
			set(p3[j:j+dan.size], dan.time+4)
			s += fmt.Sprintf("\\pos(%d,%d)", 320, (j+dan.size)*10-1)
			fmt.Fprintf(w, "Dialogue: 3,%s,%s,Default,,0000,0000,0000,,{%s}%s\r\n",
				timespot(dan.time+0),
				timespot(dan.time+4), s, dan.text)
		}
	}
}

// 生成弹幕文件
func translate(r io.ReadCloser, name string) error {
	defer r.Close()
	// 获取xml数据
	dans, err := open(r)
	if err != nil {
		return err
	}
	// 创建字幕文件
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	// 对弹幕进行排序
	sort.Sort(Danmus(dans))
	// utf-8 bom头
	file.Write([]byte{0xEF, 0xBB, 0xBF})
	// ass文件头
	file.WriteString(header)
	// 写入弹幕信息
	save(file, dans)
	return nil
}

// 识别输入（文件名/网址）并提供弹幕数据
func identify(par string) (r io.ReadCloser, name string, err error) {
	if biliz.MatchString(par) {
		err := func() error {
			resp, err := http.Get(par)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			matched := addrs.FindSubmatch(data)
			if len(matched) == 0 {
				return errNoXML
			}
			par = string(matched[1]) + `.xml`
			matched = title.FindSubmatch(data)
			if len(matched) == 0 {
				return errNoTitle
			}
			name = string(matched[1]) + `.ass`
			resp, err = http.Get(`http://comment.bilibili.com/` + par)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.Header.Get("Content-Encoding") == "deflate" {
				resp.Body = flate.NewReader(resp.Body)
				defer resp.Body.Close()
			}
			file, err := os.Create(par)
			if err != nil {
				return err
			}
			defer file.Close()
			xml, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			file.Write(xml)
			return nil
		}()
		if err != nil {
			return nil, "", err
		}
	} else {
		if n := strings.LastIndex(par, "."); n < 0 {
			name = par + ".ass"
		} else {
			name = par[:n] + ".ass"
		}
	}
	file, err := os.Open(par)
	if err != nil {
		return nil, "", err
	}
	return file, name, nil
}

// 主函数，实现了命令行
func main() {
	if len(os.Args) <= 1 {
		os.Exit(0)
	}
	for _, name := range os.Args[1:] {
		r, name, err := identify(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = translate(r, name)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}
}
