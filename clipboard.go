package main

/*
#include <windows.h>
#include <stdlib.h>

char *cpbd()
{
	OpenClipboard(NULL);
	HGLOBAL hMem=GetClipboardData(CF_TEXT);
	char* lpStr = (char*)GlobalLock(hMem);
	CloseClipboard();
	return lpStr;
}
*/
import "C"

// 使用cgo调用mingw的库函数，来读取windows的剪贴板文本数据
func clipboard() string {
	i := C.cpbd()
	s := C.GoString(i)
	return s
}

func main() {
	println(clipboard())
}
