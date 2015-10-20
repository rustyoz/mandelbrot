package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/samples/flags"
)

var outY, outX int = 1000, 1000
var yMin float64 = -1.0
var yMax float64 = +1.0
var xMin float64 = -1.0
var xMax float64 = 1.0

var cx, cy, zx, zy, new_zx float64

var dxy float64

var n, nx, ny int

var img gxui.Image

var texture gxui.Texture
var window gxui.Window

var d gxui.Driver

func main() {
	gl.StartDriver(appMain)
}

func appMain(driver gxui.Driver) {

	d = driver
	source := image.Image(newMandelbrot())

	theme := flags.CreateTheme(driver)

	mx := source.Bounds().Max

	img = theme.CreateImage()

	window = theme.CreateWindow(mx.X, mx.Y, "Image viewer")
	window.SetScale(flags.DefaultScaleFactor)
	window.AddChild(img)

	rgba := image.NewRGBA(source.Bounds())
	draw.Draw(rgba, source.Bounds(), source, image.ZP, draw.Src)
	texture = driver.CreateTexture(rgba, 1)
	img.SetTexture(texture)

	window.OnClick(windowOnClickHandler)
	window.OnClose(driver.Terminate)
}

func newMandelbrot() image.Image {

	dxy = (yMax - yMin) / (float64(outY) * 2)
	outputRectangle := image.Rect(0, 0, outY, outX)

	outputImage := image.NewRGBA(outputRectangle)
	var maxn int
	var minn int
	minn = 2 ^ 16

	var ns [1024][2014]int
	var hues [1024][2014]float64
	for cy = yMin; cy < yMax; cy += dxy {
		for cx = xMin; cx < xMax; cx += dxy {
			n, nu := MandelbrotPixel(cx, cy)

			px := int((cx - xMin) / (xMax - xMin) * float64(outX))
			py := int((cy - yMin) / (yMax - yMin) * float64(outY))
			ns[px][py] = n
			hues[px][py] = nu
			if n > maxn {
				maxn = n
			}
			if n < minn {
				minn = n
			}

		}
	}

	for px := 0; px < 1024; px++ {
		for py := 0; py < 1024; py++ {

			hue := hues[px][py]
			hue = math.Sin(hue) + 1.0
			hue = hue / 2.0
			r, g, b := hslToRgb(hue, 0.6, 0.5)
			//	n8 := uint8(n)
			//c := color.RGBA{n8, n8, n8, 255}
			c := color.RGBA{r, g, b, 255}
			//fmt.Println(px, py, c)
			outputImage.Set(px, py, c)

		}

	}

	/*fmt.Println(maxn, minn)
	for i := 0; i < 50; i++ {
		fmt.Println(hues[500][i+100])
	}*/
	return outputImage
}

func MandelbrotPixel(x float64, y float64) (n int, nu float64) {
	var zx float64
	var zy float64
	n = 0
	for zx*zx+zy*zy < 4.0 && n != 2^64 {
		new_zx = zx*zx - zy*zy + x
		zy = 2.0*zx*zy + y
		zx = new_zx
		n++
	}

	logzn := math.Log(zx*zx+zy*zy) / 2.0
	nu = math.Log(logzn/math.Log(2.0)) / math.Log(2.0)
	return n, nu
}

func windowOnClickHandler(me gxui.MouseEvent) {
	y := yMax - yMin // mandelbrot axis size
	x := xMax - xMin

	xp := float64(me.WindowPoint.X) // point clicked on screen in pixels
	yp := float64(me.WindowPoint.Y)

	// find point clicked in mandelbrot space
	xm := xMin + (xp/1024)*x
	ym := yMin + (yp/1024)*y

	// scale viewport of mandelbrot space
	if me.Button == gxui.MouseButtonLeft {
		x = x / 2
		y = y / 2
	} else {
		x = x * 2
		y = y * 2

	}

	yMax = ym + y/2
	yMin = ym - y/2
	xMax = xm + x/2
	xMin = xm - x/2

	//fmt.Print(xm, ym)

	//fmt.Print(yMax, yMin, xMax, xMin)

	source := image.Image(newMandelbrot())
	rgba := image.NewRGBA(source.Bounds())
	draw.Draw(rgba, source.Bounds(), source, image.ZP, draw.Src)
	texture = d.CreateTexture(rgba, 1)
	img.SetTexture(texture)

	window.Redraw()
}

// adapted from https://github.com/mjackson/mjijackson.github.com/blob/master/2008/02/rgb-to-hsl-and-rgb-to-hsv-color-model-conversion-algorithms-in-javascript.txt

func hue2rgb(p float64, q float64, t float64) float64 {
	//fmt.Println(p, q, t)
	if t < 0 {
		t = t + 1
	}
	if t > 1.0 {
		t = t - 1.0
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

func hslToRgb(h float64, s float64, l float64) (uint8, uint8, uint8) {
	var rf float64
	var gf float64
	var bf float64

	if s == 0 {
		rf = l
		gf = l
		bf = l // achromatic

	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		rf = hue2rgb(p, q, (h + 1.0/3.0))
		gf = hue2rgb(p, q, h)
		bf = hue2rgb(p, q, (h - 1.0/3.0))
	}

	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}
