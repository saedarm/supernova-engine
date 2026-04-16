package render

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/saedarm/supernova-engine/internal/physics"
	"github.com/saedarm/supernova-engine/internal/state"
)

// Panel draws the left sidebar with all controls, readouts, and status.
func Panel(screen *ebiten.Image, g *state.Game, mouseX, mouseY int) {
	st := g.ST()
	pw := float64(state.PanelWidth)

	// Background
	ebitenutil.DrawRect(screen, 0, 0, pw, float64(state.ScreenHeight), color.RGBA{10, 10, 26, 245})
	ebitenutil.DrawRect(screen, pw-1, 0, 1, float64(state.ScreenHeight), color.RGBA{26, 26, 46, 255})

	y := 8.0

	// Title
	ebitenutil.DebugPrintAt(screen, "=== SUPERNOVA ENGINE ===", 55, int(y))
	y += 14
	ebitenutil.DebugPrintAt(screen, "Lane-Emden + HR Diagram", 48, int(y))
	y += 20

	// Star type selector
	sectionHeader(screen, "STAR TYPE [1-6] (click)", 8, int(y))
	y += 14
	for i, key := range physics.StarTypeOrder {
		t := physics.StarTypes[key]
		sel := key == g.StarTypeKey
		if sel {
			ebitenutil.DrawRect(screen, 8, y, pw-16, 16, color.RGBA{26, 26, 58, 255})
			ebitenutil.DrawRect(screen, 8, y, 2, 16, color.RGBA{68, 170, 255, 255})
		}
		hovered := mouseX >= 8 && mouseX <= int(pw)-8 && mouseY >= int(y) && mouseY < int(y)+16
		if hovered && !sel {
			ebitenutil.DrawRect(screen, 8, y, pw-16, 16, color.RGBA{20, 20, 40, 255})
		}
		marker := "  "
		if sel {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, fmtf("%s%d. %s", marker, i+1, t.Name), 10, int(y)+2)
		y += 18
	}
	y += 4

	ebitenutil.DebugPrintAt(screen, st.Description, 10, int(y))
	y += 14
	ebitenutil.DebugPrintAt(screen, fmtf("Mass: %.1f M_sun  n=%.1f", st.Mass, st.PolyIndex), 10, int(y))
	y += 14
	ebitenutil.DebugPrintAt(screen, fmtf("xi_1=%.3f  Fate: %s", g.Profile.SurfaceXi, st.Fate), 10, int(y))
	y += 18

	// Elements
	sectionHeader(screen, "ELEMENTS [Q W E A S D F] (click)", 8, int(y))
	y += 14
	for i, elKey := range physics.ElementOrder {
		el := physics.Elements[elKey]
		active := g.Elements[elKey]
		bx := float64(10 + i*36)
		c := el.Color
		if active {
			ebitenutil.DrawRect(screen, bx, y, 32, 16, color.RGBA{uint8(c[0] * 60), uint8(c[1] * 60), uint8(c[2] * 60), 255})
			ebitenutil.DrawRect(screen, bx, y, 32, 1, color.RGBA{uint8(c[0] * 200), uint8(c[1] * 200), uint8(c[2] * 200), 255})
			ebitenutil.DrawRect(screen, bx, y+15, 32, 1, color.RGBA{uint8(c[0] * 200), uint8(c[1] * 200), uint8(c[2] * 200), 255})
		} else {
			ebitenutil.DrawRect(screen, bx, y, 32, 16, color.RGBA{17, 17, 17, 255})
		}
		ebitenutil.DebugPrintAt(screen, el.Symbol, int(bx)+6, int(y)+3)
	}
	y += 24

	// Age / Lifecycle
	sectionHeader(screen, "STELLAR AGE", 8, int(y))
	y += 14
	ageFrac := g.AgeFrac()
	ebitenutil.DebugPrintAt(screen, fmtf("%.4f / %.2f Gyr (%.1f%%)", g.Age, st.MaxAge, ageFrac*100), 10, int(y))
	y += 14

	// Progress bar
	barW := pw - 24
	ebitenutil.DrawRect(screen, 10, y, barW, 8, color.RGBA{20, 20, 35, 255})
	fillW := physics.Clamp(ageFrac/1.2, 0, 1) * barW
	barCol := color.RGBA{68, 170, 255, 180}
	if ageFrac > 0.9 {
		barCol = color.RGBA{255, 100, 50, 180}
	}
	ebitenutil.DrawRect(screen, 10, y, fillW, 8, barCol)
	y += 16

	// Play / Reset
	playLabel := "[SPACE] Play"
	if g.IsAging {
		playLabel = "[SPACE] Pause"
	}
	panelButton(screen, 10, y, 68, 16, playLabel, g.IsAging, mouseX, mouseY)
	panelButton(screen, 84, y, 52, 16, "[R] Reset", false, mouseX, mouseY)
	y += 22

	// Speed
	ebitenutil.DebugPrintAt(screen, fmtf("Speed [+/-]: %.1fx", g.AgeSpeed), 10, int(y))
	y += 14
	speeds := []float64{0.5, 1, 2, 5}
	for i, s := range speeds {
		bx := float64(10 + i*42)
		panelButton(screen, bx, y, 38, 16, fmtf("%.1fx", s), g.AgeSpeed == s, mouseX, mouseY)
	}
	y += 22

	// Phase indicator
	phaseCol := color.RGBA{20, 20, 30, 255}
	switch g.Phase {
	case state.PhaseSupernova:
		phaseCol = color.RGBA{60, 25, 0, 255}
	case state.PhaseBlackHole:
		phaseCol = color.RGBA{30, 0, 50, 255}
	case state.PhaseNeutronStar:
		phaseCol = color.RGBA{0, 25, 60, 255}
	}
	ebitenutil.DrawRect(screen, 8, y, pw-16, 30, phaseCol)
	ebitenutil.DebugPrintAt(screen, fmtf("[%s] %s", g.Phase.Icon(), g.Phase), 12, int(y)+2)
	switch g.Phase {
	case state.PhaseSupernova:
		ebitenutil.DebugPrintAt(screen, fmtf("Core collapse: %.0f%%", g.CollapseProgress*100), 12, int(y)+14)
	case state.PhaseBlackHole:
		ebitenutil.DebugPrintAt(screen, "Singularity. Event horizon.", 12, int(y)+14)
	case state.PhaseNeutronStar:
		ebitenutil.DebugPrintAt(screen, "Rapid rotation. B~10^8 T", 12, int(y)+14)
	}
	y += 36

	// View mode
	sectionHeader(screen, "VIEW [Tab]", 8, int(y))
	y += 14
	viewLabels := []string{"Split", "Star Only", "HR Only"}
	for i, label := range viewLabels {
		bx := float64(10 + i*80)
		panelButton(screen, bx, y, 76, 16, label, int(g.View) == i, mouseX, mouseY)
	}
	y += 22

	// HR Position
	sectionHeader(screen, "HR POSITION", 8, int(y))
	y += 14
	logT, logL := g.HRPosition()
	tEff := math.Pow(10, logT)
	lum := math.Pow(10, logL)
	lumStr := fmtf("%.3f", lum)
	if lum > 100 {
		lumStr = fmtf("%.0f", lum)
	}
	ebitenutil.DebugPrintAt(screen, fmtf("T_eff = %.0f K", tEff), 10, int(y))
	y += 12
	ebitenutil.DebugPrintAt(screen, fmtf("L = %s L_sun", lumStr), 10, int(y))
	y += 12
	ebitenutil.DebugPrintAt(screen, fmtf("logT=%.3f  logL=%.3f", logT, logL), 10, int(y))
	y += 18

	// Density plot
	sectionHeader(screen, "DENSITY [P]", 8, int(y))
	y += 14
	if g.ShowDensityPlot {
		DensityPlot(screen, g, 10, y, pw-20, 90)
		y += 96
	} else {
		ebitenutil.DebugPrintAt(screen, "Press P to show theta(xi)", 10, int(y))
		y += 14
	}

	// Controls legend
	y += 4
	sectionHeader(screen, "CONTROLS", 8, int(y))
	y += 14
	for _, l := range []string{
		"1-6  Star type", "SPACE  Age on/off", "R  Reset", "+/-  Speed",
		"[ ]  Brightness", "Tab  View mode", "P  Density plot", "QWEASDF  Elements",
	} {
		ebitenutil.DebugPrintAt(screen, l, 10, int(y))
		y += 12
	}
	y += 4
	ebitenutil.DebugPrintAt(screen, fmtf("Brightness: %.0f%%", g.Brightness*100), 10, int(y))
	y += 12
	ebitenutil.DebugPrintAt(screen, fmtf("FPS: %.0f", ebiten.ActualFPS()), 10, int(y))
}

func sectionHeader(screen *ebiten.Image, title string, x, y int) {
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(state.PanelWidth)-16, 12, color.RGBA{15, 15, 28, 255})
	ebitenutil.DebugPrintAt(screen, strings.ToUpper(title), x+2, y+1)
}

func panelButton(screen *ebiten.Image, x, y float64, w, h float64, label string, active bool, mx, my int) {
	bg := color.RGBA{14, 14, 26, 255}
	border := color.RGBA{26, 26, 46, 255}
	if active {
		bg = color.RGBA{26, 26, 58, 255}
		border = color.RGBA{68, 170, 255, 255}
	}
	hovered := mx >= int(x) && mx <= int(x+w) && my >= int(y) && my < int(y+h)
	if hovered && !active {
		bg = color.RGBA{22, 22, 42, 255}
	}
	ebitenutil.DrawRect(screen, x, y, w, h, bg)
	ebitenutil.DrawRect(screen, x, y, w, 1, border)
	ebitenutil.DrawRect(screen, x, y+h-1, w, 1, border)
	ebitenutil.DrawRect(screen, x, y, 1, h, border)
	ebitenutil.DrawRect(screen, x+w-1, y, 1, h, border)
	maxC := int(w) / 6
	d := label
	if len(d) > maxC {
		d = d[:maxC]
	}
	ebitenutil.DebugPrintAt(screen, d, int(x)+int(w)/2-len(d)*3, int(y)+int(h)/2-4)
}
