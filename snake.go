package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

/*
#include <windows.h>
#include <conio.h>

// 使用了WinAPI来移动控制台的光标
void gotoxy(int x,int y)
{
	COORD c;
	c.X=x,c.Y=y;
	SetConsoleCursorPosition(GetStdHandle(STD_OUTPUT_HANDLE),c);
}

// 从键盘获取一次按键，但不显示到控制台
int direct()
{
	return _getch();
}
*/
import "C" // go中可以嵌入C语言的函数

// 表示光标的位置
type loct struct {
	i, j int
}

var (
	area = [20][20]byte{} // 记录了蛇、食物的信息
	lead = byte('R')      // 当前蛇头移动方向
	head = loct{4, 4}     // 当前蛇头位置
	tail = loct{4, 4}     // 当前蛇尾位置
	food = false          // 当前是否有食物
	size = 1              // 当前蛇身长度
)

// 随机生成一个位置，来放置食物
func place() loct {
	k := rand.Int() % 400
	return loct{k / 20, k % 20}
}

// 用来更新控制台的显示，在指定位置写字符，使用错误输出避免缓冲
func draw(p loct, c byte) {
	C.gotoxy(C.int(p.i*2+4), C.int(p.j+2))
	fmt.Fprintf(os.Stderr, "%c", c)
}

func init() {

	// 初始化蛇的位置和方向、首尾；初始化随机数
	area[4][4] = 'H'
	rand.Seed(int64(time.Now().Unix()))

	// 输出初始画面
	fmt.Fprintln(os.Stderr,
		`
  #-----------------------------------------#
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |         *                               |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  |                                         |
  #-----------------------------------------#
`)

	// 我们使用一个单独的go程来捕捉键盘的动作，因为是单独的，不怕阻塞
	go func() {
		for { // 函数只写入lead，外部只读取lead，无需设锁
			switch byte(C.direct()) {
			case 72:
				lead = 'U'
			case 75:
				lead = 'L'
			case 77:
				lead = 'R'
			case 80:
				lead = 'D'
			case 32:
				lead = 'P'
			}
		}
	}()
}

func main() {

	// 主程序
	for {

		// 程序更新周期，400毫秒
		time.Sleep(time.Millisecond * 400)

		// 暂停，还是要有滴
		if lead == 'P' {
			continue
		}

		// 放置食物
		if !food {
			give := place()
			if area[give.i][give.j] == 0 { // 食物只能放在空闲位置
				area[give.i][give.j] = 'F'
				draw(give, '$') // 绘制食物
				food = true
			}
		}

		// 我们在蛇头位置记录它移动的方向
		area[head.i][head.j] = lead

		// 根据lead来移动蛇头
		switch lead {
		case 'U':
			head.j--
		case 'L':
			head.i--
		case 'R':
			head.i++
		case 'D':
			head.j++
		}

		// 判断蛇头是否出界
		if head.i < 0 || head.i >= 20 || head.j < 0 || head.j >= 20 {
			C.gotoxy(0, 23) // 让光标移动到画面下方
			break           // 跳出死循环
		}

		// 获取蛇头位置的原值，来判断是否撞车，或者吃到食物
		eat := area[head.i][head.j]

		if eat == 'F' { // 吃到食物
			food = false

			// 增加蛇的尺寸，并且不移动蛇尾
			size++
		} else if eat == 0 { // 普通移动

			draw(tail, ' ') // 擦除蛇尾

			// 注意我们记录了它移动的方向
			dir := area[tail.i][tail.j]

			// 我们需要擦除蛇尾的记录
			area[tail.i][tail.j] = 0

			// 移动蛇尾
			switch dir {
			case 'U':
				tail.j--
			case 'L':
				tail.i--
			case 'R':
				tail.i++
			case 'D':
				tail.j++
			}
		} else { // 撞车了
			C.gotoxy(0, 23)
			break
		}
		draw(head, '*') // 绘制蛇头
	}

	// 收尾了
	switch {
	case size < 22:
		fmt.Fprint(os.Stderr, "Faild! ")
	case size < 42:
		fmt.Fprint(os.Stderr, "Try your best! ")
	default:
		fmt.Fprint(os.Stderr, "Congratulations! ")
	}
	fmt.Fprintf(os.Stderr, "You've eaten %d $\n", size-1)
}
