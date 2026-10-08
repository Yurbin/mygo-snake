//go:build ignore

// 生成应用图标:go run tools/genicon/main.go [输出路径]
// 输出 1024×1024 PNG:绿色圆角底 + 亮薄荷色小蛇 + 红苹果,与游戏内配色一致。

package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/vector"
)

const size = 1024

func main() {
	out := "resources/icon.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// 背景圆角矩形,垂直渐变(#34d399 → #059669)。
	// vector 只支持纯色填充,先画白底圆角矩形做遮罩,再逐像素上渐变色。
	mask := image.NewRGBA(image.Rect(0, 0, size, size))
	r := vector.NewRasterizer(size, size)
	roundedRect(r, 28, 28, size-28, size-28, 232)
	r.Draw(mask, mask.Bounds(), image.NewUniform(color.RGBA{A: 255}), image.Point{})
	top, bot := color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 255}, color.RGBA{R: 0x05, G: 0x96, B: 0x69, A: 255}
	for y := 0; y < size; y++ {
		t := float64(y) / (size - 1)
		c := lerp(top, bot, t)
		for x := 0; x < size; x++ {
			a := mask.RGBAAt(x, y).A
			if a == 0 {
				continue
			}
			c.A = a
			img.SetRGBA(x, y, c)
		}
	}

	// 蛇:S 形粗链(圆串),深色描边 + 薄荷色身体。
	path := [][2]float64{
		{340, 800}, {340, 600}, {640, 600}, {640, 380},
	}
	dense := resample(path, 14)
	for _, p := range dense {
		circle(img, p[0], p[1], 62, color.RGBA{R: 0x06, G: 0x4e, B: 0x3b, A: 255}) // 描边
	}
	for _, p := range dense {
		circle(img, p[0], p[1], 48, color.RGBA{R: 0xd1, G: 0xfa, B: 0xe5, A: 255}) // 身体
	}

	// 蛇头:大圆 + 眼睛,朝上。
	hx, hy := 640.0, 360.0
	circle(img, hx, hy, 96, color.RGBA{R: 0x06, G: 0x4e, B: 0x3b, A: 255})
	circle(img, hx, hy, 82, color.RGBA{R: 0xd1, G: 0xfa, B: 0xe5, A: 255})
	for _, s := range []float64{-1, 1} {
		ex, ey := hx+s*34, hy-26
		circle(img, ex, ey, 24, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 255})
		circle(img, ex, ey-6, 11, color.RGBA{R: 0x05, G: 0x2e, B: 0x16, A: 255})
	}

	// 苹果:红圆 + 高光 + 叶子。
	ax, ay := 830.0, 250.0
	circle(img, ax, ay, 78, color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 255})
	circle(img, ax-24, ay-24, 16, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 178})
	ellipse(img, ax+8, ay-92, 40, 17, -0.5, color.RGBA{R: 0x16, G: 0x5c, B: 0x3a, A: 255})

	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
	println("wrote", out)
}

func roundedRect(r *vector.Rasterizer, x0, y0, x1, y1, rad float32) {
	r.MoveTo(x0+rad, y0)
	r.LineTo(x1-rad, y0)
	r.QuadTo(x1, y0, x1, y0+rad)
	r.LineTo(x1, y1-rad)
	r.QuadTo(x1, y1, x1-rad, y1)
	r.LineTo(x0+rad, y1)
	r.QuadTo(x0, y1, x0, y1-rad)
	r.LineTo(x0, y0+rad)
	r.QuadTo(x0, y0, x0+rad, y0)
	r.ClosePath()
}

func circle(img *image.RGBA, cx, cy, rad float64, c color.RGBA) {
	r := vector.NewRasterizer(size, size)
	const segs = 64
	r.MoveTo(float32(cx+rad), float32(cy))
	for i := 1; i <= segs; i++ {
		a := float64(i) / segs * 2 * math.Pi
		r.LineTo(float32(cx+rad*math.Cos(a)), float32(cy+rad*math.Sin(a)))
	}
	r.ClosePath()
	r.Draw(img, img.Bounds(), image.NewUniform(c), image.Point{})
}

func ellipse(img *image.RGBA, cx, cy, rx, ry, rot float64, c color.RGBA) {
	r := vector.NewRasterizer(size, size)
	const segs = 64
	cr, sr := math.Cos(rot), math.Sin(rot)
	pt := func(t float64) (float32, float32) {
		x, y := rx*math.Cos(t), ry*math.Sin(t)
		return float32(cx + x*cr - y*sr), float32(cy + x*sr + y*cr)
	}
	x0, y0 := pt(0)
	r.MoveTo(x0, y0)
	for i := 1; i <= segs; i++ {
		x, y := pt(float64(i) / segs * 2 * math.Pi)
		r.LineTo(x, y)
	}
	r.ClosePath()
	r.Draw(img, img.Bounds(), image.NewUniform(c), image.Point{})
}

// resample 把折线按 step 间距插值成密集点列(圆串的身体)。
func resample(path [][2]float64, step float64) [][2]float64 {
	var out [][2]float64
	for i := 0; i+1 < len(path); i++ {
		ax, ay := path[i][0], path[i][1]
		bx, by := path[i+1][0], path[i+1][1]
		d := math.Hypot(bx-ax, by-ay)
		n := int(d / step)
		for j := 0; j <= n; j++ {
			t := float64(j) / float64(n)
			out = append(out, [2]float64{ax + (bx-ax)*t, ay + (by-ay)*t})
		}
	}
	out = append(out, path[len(path)-1])
	return out
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	f := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t) }
	return color.RGBA{R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), A: 255}
}
