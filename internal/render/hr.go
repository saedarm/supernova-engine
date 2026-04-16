package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/saedarm/supernova-engine/internal/physics"
	"github.com/saedarm/supernova-engine/internal/state"
)

// TempToColor converts log10(T_eff) to an approximate spectral color.
func TempToColor(logT float64) color.RGBA {
	t := math.Pow(10, logT)
	switch {
	case t > 30000:
		return color.RGBA{155, 176, 255, 255}
	case t > 10000:
		return color.RGBA{170, 191, 255, 255}
	case t > 7500:
		return color.RGBA{202, 215, 255, 255}
	case t > 6000:
		return color.RGBA{248, 247, 255, 255}
	case t > 5200:
		return color.RGBA{255, 244, 234, 255}
	case t > 3700:
		return color.RGBA{255, 210, 161, 255}
	default:
		return color.RGBA{255, 170, 110, 255}
	}
}

// HRDiagram draws the full Hertzsprung-Russell diagram with evolutionary
// track, reference stars, spectral classes, and the current position dot.
func HRDiagram(screen *ebiten.Image, g *state.Game, ox, oy, w, h float64) {
	st := g.ST()

	ebitenutil.DrawRect(screen, ox, oy, w, h, color.RGBA{5, 5, 15, 255})

	logTMin, logTMax := 3.35, 4.75
	logLMin, logLMax := -5.0, 7.0
	padT, padB, padL, padR := 36.0, 40.0, 52.0, 20.0
	plotW := w - padL - padR
	plotH := h - padT - padB

	// Map functions (temperature axis reversed: hot on left)
	mapX := func(logT float64) float64 { return ox + padL + (1-(logT-logTMin)/(logTMax-logTMin))*plotW }
	mapY := func(logL float64) float64 { return oy + padT + (1-(logL-logLMin)/(logLMax-logLMin))*plotH }

	// Main sequence band
	ms := physics.MainSequenceBand
	for i := 0; i < len(ms)-1; i++ {
		x1, y1 := mapX(ms[i][0]+0.03), mapY(ms[i][1]+0.7)
		x2, y2 := mapX(ms[i+1][0]+0.03), mapY(ms[i+1][1]+0.7)
		x3, y3 := mapX(ms[i+1][0]-0.03), mapY(ms[i+1][1]-0.7)
		x4, y4 := mapX(ms[i][0]-0.03), mapY(ms[i][1]-0.7)
		msCol := color.RGBA{60, 120, 255, 10}
		fillTriangle(screen, x1, y1, x2, y2, x3, y3, msCol)
		fillTriangle(screen, x1, y1, x3, y3, x4, y4, msCol)
	}
	for i := 0; i < len(ms)-1; i++ {
		vector.StrokeLine(screen, float32(mapX(ms[i][0])), float32(mapY(ms[i][1])),
			float32(mapX(ms[i+1][0])), float32(mapY(ms[i+1][1])), 1, color.RGBA{80, 140, 255, 25}, true)
	}

	// Region labels
	ebitenutil.DebugPrintAt(screen, "Giants", int(mapX(3.62))-12, int(mapY(3.5)))
	ebitenutil.DebugPrintAt(screen, "Supergiants", int(mapX(4.0))-16, int(mapY(5.5)))
	ebitenutil.DebugPrintAt(screen, "White Dwarfs", int(mapX(4.05))-16, int(mapY(-3.0)))
	ebitenutil.DebugPrintAt(screen, "Main Seq.", int(mapX(3.92))-10, int(mapY(1.0)))

	// Grid
	for lt := 3.4; lt <= 4.7; lt += 0.1 {
		x := float32(mapX(lt))
		vector.StrokeLine(screen, x, float32(oy+padT), x, float32(oy+h-padB), 0.5, color.RGBA{255, 255, 255, 8}, true)
	}
	for ll := -4.0; ll <= 6; ll += 1 {
		y := float32(mapY(ll))
		vector.StrokeLine(screen, float32(ox+padL), y, float32(ox+w-padR), y, 0.5, color.RGBA{255, 255, 255, 8}, true)
	}

	// Axes
	axisCol := color.RGBA{42, 42, 69, 255}
	vector.StrokeLine(screen, float32(ox+padL), float32(oy+padT), float32(ox+padL), float32(oy+h-padB), 1, axisCol, true)
	vector.StrokeLine(screen, float32(ox+padL), float32(oy+h-padB), float32(ox+w-padR), float32(oy+h-padB), 1, axisCol, true)

	// Y-axis labels
	for ll := -4.0; ll <= 6; ll += 2 {
		ebitenutil.DebugPrintAt(screen, fmtf("10^%d", int(ll)), int(ox)+4, int(mapY(ll))-4)
	}
	ebitenutil.DebugPrintAt(screen, "L/L_sun", int(ox)+2, int(oy)+8)

	// X-axis: spectral classes
	for _, sc := range physics.SpectralClasses {
		x := int(mapX(sc.LogT))
		if float64(x) < ox+padL || float64(x) > ox+w-padR {
			continue
		}
		col := TempToColor(sc.LogT)
		vector.StrokeLine(screen, float32(x), float32(oy+h-padB), float32(x), float32(oy+h-padB+4), 2, col, true)
		ebitenutil.DebugPrintAt(screen, sc.Label, x-3, int(oy+h-padB)+6)
		ebitenutil.DebugPrintAt(screen, fmtf("%.0fK", math.Pow(10, sc.LogT)), x-12, int(oy+h-padB)+18)
	}
	ebitenutil.DebugPrintAt(screen, "<-- Hotter          Surface Temp          Cooler -->", int(ox+padL), int(oy+h)-10)

	// Reference stars
	for _, rs := range physics.ReferenceStars {
		rx, ry := mapX(rs.LogT), mapY(rs.LogL)
		if rx < ox+padL || rx > ox+w-padR || ry < oy+padT || ry > oy+h-padB {
			continue
		}
		vector.DrawFilledCircle(screen, float32(rx), float32(ry), 2, color.RGBA{255, 255, 255, 40}, true)
		ebitenutil.DebugPrintAt(screen, rs.Name, int(rx)+4, int(ry)-8)
	}

	// Full evolutionary track (faint)
	track := st.HRTrack
	for i := 0; i < len(track)-1; i++ {
		if track[i].LogT() == 0 || track[i+1].LogT() == 0 {
			continue
		}
		vector.StrokeLine(screen,
			float32(mapX(track[i].LogT())), float32(mapY(track[i].LogL())),
			float32(mapX(track[i+1].LogT())), float32(mapY(track[i+1].LogL())),
			1, color.RGBA{255, 255, 255, 18}, true)
	}

	// Traversed portion (bright)
	ageFrac := g.AgeFrac()
	prevLT, prevLL := physics.InterpolateHR(track, 0)
	for af := 0.005; af <= ageFrac; af += 0.005 {
		lt, ll := physics.InterpolateHR(track, af)
		if lt == 0 {
			break
		}
		vector.StrokeLine(screen,
			float32(mapX(prevLT)), float32(mapY(prevLL)),
			float32(mapX(lt)), float32(mapY(ll)),
			2, color.RGBA{100, 200, 255, 90}, true)
		prevLT, prevLL = lt, ll
	}

	// Trail dots
	for i, tp := range g.HRTrail {
		if tp.LogT == 0 {
			continue
		}
		opacity := uint8(8 + float64(i)/float64(len(g.HRTrail)+1)*60)
		vector.DrawFilledCircle(screen, float32(mapX(tp.LogT)), float32(mapY(tp.LogL)), 1.5, color.RGBA{100, 200, 255, opacity}, true)
	}

	// Current position (pulsing dot)
	logT, logL := physics.InterpolateHR(track, ageFrac)
	if logT != 0 {
		px, py := mapX(logT), mapY(logL)
		dotCol := TempToColor(logT)
		pulseR := 5 + 2*math.Sin(g.Time*4)

		for ring := 3; ring >= 0; ring-- {
			r := pulseR + float64(ring)*4
			a := uint8(math.Max(0, float64(dotCol.A)/4-float64(ring)*15))
			vector.DrawFilledCircle(screen, float32(px), float32(py), float32(r), color.RGBA{dotCol.R, dotCol.G, dotCol.B, a}, true)
		}
		vector.DrawFilledCircle(screen, float32(px), float32(py), float32(pulseR), dotCol, true)
		vector.DrawFilledCircle(screen, float32(px), float32(py), 2, color.RGBA{255, 255, 255, 200}, true)

		tEff := math.Pow(10, logT)
		lum := math.Pow(10, logL)
		lumStr := fmtf("%.2f", lum)
		if lum > 100 {
			lumStr = fmtf("%.0f", lum)
		}
		ebitenutil.DebugPrintAt(screen, fmtf("T=%.0fK L=%s L_sun", tEff, lumStr), int(px)+int(pulseR)*3+4, int(py)-4)
	}

	// Title
	ebitenutil.DebugPrintAt(screen, "HERTZSPRUNG-RUSSELL DIAGRAM", int(ox+padL)+4, int(oy)+6)
	ebitenutil.DebugPrintAt(screen, fmtf("%s - Evolutionary Track", st.Name), int(ox+padL)+4, int(oy)+20)
}

// DensityPlot draws the Lane-Emden θ(ξ) curve inline.
func DensityPlot(screen *ebiten.Image, g *state.Game, x, y, w, h float64) {
	ebitenutil.DrawRect(screen, x, y, w, h, color.RGBA{10, 10, 24, 255})

	axX, axY := x+24, y+h-14
	vector.StrokeLine(screen, float32(axX), float32(y+4), float32(axX), float32(axY), 1, color.RGBA{50, 50, 70, 255}, true)
	vector.StrokeLine(screen, float32(axX), float32(axY), float32(x+w-4), float32(axY), 1, color.RGBA{50, 50, 70, 255}, true)
	ebitenutil.DebugPrintAt(screen, "theta", int(x), int(y)+2)
	ebitenutil.DebugPrintAt(screen, "xi", int(x+w)-14, int(axY)+2)

	prof := g.Profile
	if prof == nil || len(prof.Theta) == 0 {
		return
	}
	plotW, plotH := w-28, h-22
	var prevPx, prevPy float32
	for i := 0; i < len(prof.Theta); i += 2 {
		frac := prof.Xi[i] / prof.SurfaceXi
		px := float32(axX + frac*plotW)
		py := float32(y + 4 + (1-prof.Theta[i])*plotH)
		if i > 0 {
			vector.StrokeLine(screen, prevPx, prevPy, px, py, 1.5, color.RGBA{68, 170, 255, 200}, true)
		}
		prevPx, prevPy = px, py
	}
	st := g.ST()
	ebitenutil.DebugPrintAt(screen, fmtf("n=%.1f xi1=%.3f", st.PolyIndex, prof.SurfaceXi), int(axX)+20, int(y)+4)
}
