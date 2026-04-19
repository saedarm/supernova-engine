// Supernova Engine — Lane-Emden Equation Solver & Visualizer
//
// Numerically solves the Lane-Emden equation for different polytropic
// indices and renders the resulting stellar density profile as a
// glowing star built from nested translucent shells.
//
// Run:   go mod tidy && go run .
// Build: go build -o supernova-engine .
package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/saedarm/supernova-engine/internal/physics"
	"github.com/saedarm/supernova-engine/internal/render"
)

const (
	screenWidth  = 1200
	screenHeight = 750
)

type App struct {
	starTypeKey string
	customN     float64
	brightness  float64
	time        float64
	profile     *physics.DensityProfile
	bgStars     [][3]float64
	mouseX      int
	mouseY      int
}

func newApp() *App {
	a := &App{
		starTypeKey: "sun_like",
		customN:     2.0,
		brightness:  1.0,
	}

	// Deterministic background starfield
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 300; i++ {
		a.bgStars = append(a.bgStars, [3]float64{
			r.Float64() * screenWidth,
			r.Float64() * screenHeight,
			r.Float64()*0.6 + 0.1,
		})
	}

	a.recalc()
	return a
}

func (a *App) recalc() {
	n := physics.StarTypes[a.starTypeKey].PolyIndex
	if a.starTypeKey == "custom" {
		n = a.customN
	}
	a.profile = physics.SolveLaneEmden(n, 2000)
}

func (a *App) Update() error {
	a.time += 1.0 / 60.0
	a.mouseX, a.mouseY = ebiten.CursorPosition()

	changed := false

	// Star type selection: keys 1-6
	for i, key := range physics.StarTypeOrder {
		if inpututil.IsKeyJustPressed(ebiten.Key1 + ebiten.Key(i)) {
			a.starTypeKey = key
			changed = true
		}
	}

	// Custom n adjustment: +/-
	if a.starTypeKey == "custom" {
		if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
			a.customN = math.Min(a.customN+0.1, 4.9)
			changed = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
			a.customN = math.Max(a.customN-0.1, 0.5)
			changed = true
		}
	}

	// Brightness: [ ]
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		a.brightness = math.Min(a.brightness+0.1, 2.0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		a.brightness = math.Max(a.brightness-0.1, 0.1)
	}

	// Mouse clicks on panel — star type rows
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && a.mouseX < render.PanelWidth {
		for i, key := range physics.StarTypeOrder {
			by := 46 + i*18
			if a.mouseY >= by && a.mouseY < by+16 && a.mouseX >= 8 && a.mouseX <= render.PanelWidth-8 {
				a.starTypeKey = key
				changed = true
			}
		}
	}

	if changed {
		a.recalc()
	}

	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{6, 6, 14, 255})

	st := physics.StarTypes[a.starTypeKey]
	viewX := float64(render.PanelWidth)
	viewW := float64(screenWidth) - viewX
	starH := float64(screenHeight) * 0.6
	plotH := float64(screenHeight) - starH

	// Star viewport (top)
	render.Star(screen, a.profile, st, a.bgStars, a.brightness, a.time,
		viewX, 0, viewW, starH)

	// Density plot (bottom)
	plotST := st
	if a.starTypeKey == "custom" {
		plotST = &physics.StarType{
			Name: "Custom", PolyIndex: a.customN, Mass: 1.0,
			BaseColor: st.BaseColor, CoronaColor: st.CoronaColor,
		}
	}
	render.DensityPlot(screen, a.profile, plotST, viewX, starH, viewW, plotH)

	// Divider
	ebitenutil.DrawRect(screen, viewX, starH-0.5, viewW, 1, color.RGBA{26, 26, 46, 255})

	// Panel
	render.Panel(screen, a.starTypeKey, a.customN, a.brightness,
		a.profile, a.mouseX, a.mouseY)
}

func (a *App) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Supernova Engine — Lane-Emden Equation Solver")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)

	if err := ebiten.RunGame(newApp()); err != nil {
		log.Fatal(err)
	}
}
