// Пакет converter конвертирует изображение в ASCII-арт.
package converter

import (
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/cubexteam/goascii/internal/charset"
)

// Options — параметры конвертации
type Options struct {
	Width   int
	Invert  bool
	Charset charset.Charset
	Colored bool
}

// Pixel — один символ с RGB-цветом
type Pixel struct {
	Char    rune
	R, G, B uint8
}

// Result — итог конвертации: двумерная сетка символов
type Result struct {
	Rows       [][]Pixel
	OrigWidth  int
	OrigHeight int
}

// Convert загружает изображение из файла и конвертирует в ASCII
func Convert(path string, opts Options) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return convertImage(img, opts), nil
}

// ConvertBytes декодирует изображение из байт и конвертирует в ASCII.
// Используется веб-интерфейсом при загрузке файла через браузер.
func ConvertBytes(data []byte, opts Options) (*Result, error) {
	r := newBytesReader(data)
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return convertImage(img, opts), nil
}

// convertImage выполняет основную логику преобразования
func convertImage(img image.Image, opts Options) *Result {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// Высота с учётом соотношения сторон символа (~0.5 по вертикали)
	height := int(math.Round(float64(origH) / float64(origW) * float64(opts.Width) * 0.45))
	if height < 1 {
		height = 1
	}

	resized := resizeImage(img, opts.Width, height)

	cs := opts.Charset
	rows := make([][]Pixel, height)

	for y := 0; y < height; y++ {
		row := make([]Pixel, opts.Width)
		for x := 0; x < opts.Width; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			lum := luminance(r8, g8, b8)
			ch := mapToChar(lum, cs, opts.Invert)

			row[x] = Pixel{Char: ch, R: r8, G: g8, B: b8}
		}
		rows[y] = row
	}

	return &Result{Rows: rows, OrigWidth: origW, OrigHeight: origH}
}

// resizeImage масштабирует изображение через билинейную интерполяцию.
// Реализовано вручную — внешние пакеты не нужны.
func resizeImage(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	srcB := src.Bounds()
	srcW := srcB.Dx()
	srcH := srcB.Dy()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			fx := float64(x) / float64(w) * float64(srcW)
			fy := float64(y) / float64(h) * float64(srcH)

			x0 := int(fx)
			y0 := int(fy)
			x1 := x0 + 1
			y1 := y0 + 1

			if x1 >= srcW {
				x1 = srcW - 1
			}
			if y1 >= srcH {
				y1 = srcH - 1
			}

			dx := fx - float64(x0)
			dy := fy - float64(y0)

			c00 := toRGBA(src.At(srcB.Min.X+x0, srcB.Min.Y+y0))
			c10 := toRGBA(src.At(srcB.Min.X+x1, srcB.Min.Y+y0))
			c01 := toRGBA(src.At(srcB.Min.X+x0, srcB.Min.Y+y1))
			c11 := toRGBA(src.At(srcB.Min.X+x1, srcB.Min.Y+y1))

			rv := lerp2(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R), dx, dy)
			gv := lerp2(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G), dx, dy)
			bv := lerp2(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B), dx, dy)
			av := lerp2(float64(c00.A), float64(c10.A), float64(c01.A), float64(c11.A), dx, dy)

			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(rv), G: uint8(gv), B: uint8(bv), A: uint8(av),
			})
		}
	}

	return dst
}

// lerp2 — билинейная интерполяция одного канала
func lerp2(c00, c10, c01, c11, dx, dy float64) float64 {
	return c00*(1-dx)*(1-dy) + c10*dx*(1-dy) + c01*(1-dx)*dy + c11*dx*dy
}

// toRGBA конвертирует любой color.Color в color.RGBA
func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

// luminance вычисляет воспринимаемую яркость пикселя (стандарт BT.601)
func luminance(r, g, b uint8) float64 {
	return 0.299*float64(r)/255.0 + 0.587*float64(g)/255.0 + 0.114*float64(b)/255.0
}

// mapToChar сопоставляет яркость символу из набора
func mapToChar(lum float64, cs charset.Charset, invert bool) rune {
	if invert {
		lum = 1.0 - lum
	}
	idx := int(lum * float64(len(cs)-1))
	if idx >= len(cs) {
		idx = len(cs) - 1
	}
	return cs[idx]
}
