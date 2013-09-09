package main

import (
	"code.google.com/p/freetype-go/freetype"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
)

var (
	bgc = image.White
	red = image.NewUniform(color.NRGBA{255,0,0,255})
	frc = image.Black
	img = image.NewRGBA(image.Rect(0, 0, 1000, 500))
	con = freetype.NewContext()
	str = "Hello World!"
	ori = freetype.Pt(0, 0)
)

func init() {
	data, err := ioutil.ReadFile(`d:\consola.ttf`)
	if err != nil {
		os.Exit(1)
	}
	font, err := freetype.ParseFont(data)
	if err != nil {
		os.Exit(1)
	}
	con.SetDPI(72)
	con.SetFont(font)
	con.SetFontSize(20)
	con.SetSrc(frc)
	con.SetDst(img)
	con.SetClip(img.Bounds())
	draw.Draw(img, img.Bounds(), bgc, image.Point{0,0}, draw.Src)
}

func main() {
	file,err:=os.Create(`out.png`)
	if err != nil {
		return
	}
	defer file.Close()
	
	con.SetDPI(72)
	ori = freetype.Pt(10, 100)
	_, err = con.DrawString(str, ori)
	
	con.SetDPI(108)
	ori = freetype.Pt(10, 200)
	_, err = con.DrawString(str, ori)
	
	con.SetDPI(144)
	ori = freetype.Pt(10, 300)
	_, err = con.DrawString(str, ori)
	
	con.SetDPI(180)
	ori = freetype.Pt(10, 400)
	_, err = con.DrawString(str, ori)
	
	con.SetSrc(red)
	con.SetDPI(72)
	
	con.SetFontSize(20)
	ori = freetype.Pt(510, 100)
	_, err = con.DrawString(str, ori)
	
	con.SetFontSize(30)
	ori = freetype.Pt(510, 200)
	_, err = con.DrawString(str, ori)
	
	con.SetFontSize(40)
	ori = freetype.Pt(510, 300)
	_, err = con.DrawString(str, ori)
	
	con.SetFontSize(50)
	ori = freetype.Pt(510, 400)
	_, err = con.DrawString(str, ori)
	
	png.Encode(file,img)
}