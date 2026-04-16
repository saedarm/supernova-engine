package render

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// fmtf is a shortcut for fmt.Sprintf.
func fmtf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// fillTriangle rasterizes a filled triangle using horizontal scanlines.
func fillTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float64, col color.RGBA) {
	// Sort by Y
	if y1 > y2 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}
	if y1 > y3 {
		x1, y1, x3, y3 = x3, y3, x1, y1
	}
	if y2 > y3 {
		x2, y2, x3, y3 = x3, y3, x2, y2
	}
	if y3-y1 < 1 {
		return
	}
	for sy := math.Floor(y1); sy <= math.Ceil(y3); sy += 2 {
		var left, right float64
		if sy < y2 {
			if y2-y1 > 0 {
				t := (sy - y1) / (y2 - y1)
				left = x1 + t*(x2-x1)
			} else {
				left = x1
			}
		} else {
			if y3-y2 > 0 {
				t := (sy - y2) / (y3 - y2)
				left = x2 + t*(x3-x2)
			} else {
				left = x2
			}
		}
		if y3-y1 > 0 {
			t := (sy - y1) / (y3 - y1)
			right = x1 + t*(x3-x1)
		} else {
			right = x1
		}
		if left > right {
			left, right = right, left
		}
		if right-left > 0 {
			ebitenutil.DrawRect(screen, left, sy, right-left, 2, col)
		}
	}
}
