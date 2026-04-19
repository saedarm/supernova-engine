# Supernova Engine

A standalone Go desktop application that numerically solves the 150-year-old **Lane-Emden equation** to model the density structure of stars, then simulates their full lifecycles — from main sequence through red giant through supernova to their final fate as white dwarfs, neutron stars, or black holes. The star's journey is plotted in real time on a **Hertzsprung-Russell diagram**.

Written in Go. Single binary. Native desktop window. No browser, no JavaScript, no web framework.

```
┌──────────────────────────────────────────────────────────────┐
│ STELLAR VIEWPORT            │  HERTZSPRUNG-RUSSELL DIAGRAM   │
│                             │                                 │
│        ┌──────────┐         │  10⁶                            │
│      ╭─┤ glowing  ├─╮       │         • Supergiants           │
│     ╱  │   star   │  ╲      │  10⁴  ╱                         │
│    │   │ (Lane-   │   │     │     ╱ • • Main Sequence         │
│     ╲  │  Emden   │  ╱      │  10²╱                           │
│      ╰─┤  shells) ├─╯       │   1•━━━━━━━━━━━━                │
│        └──────────┘         │  10⁻² White Dwarfs              │
│                             │       O B A F G K M             │
└──────────────────────────────────────────────────────────────┘
```

---

## Quick Start

**Prerequisites:**
- Go 1.22 or later
- Linux: `sudo apt-get install libgl1-mesa-dev xorg-dev`
- macOS: `xcode-select --install`
- Windows: TDM-GCC recommended if CGo issues arise

**Run:**
```bash
git clone https://github.com/saedarm/supernova-engine
cd supernova-engine
go mod tidy
go run ./cmd/stellarforge
```

**Build a binary:**
```bash
go build -o supernova-engine ./cmd/stellarforge

# Cross-compile:
GOOS=windows GOARCH=amd64 go build -o supernova-engine.exe ./cmd/stellarforge
GOOS=darwin  GOARCH=arm64 go build -o supernova-engine-mac  ./cmd/stellarforge
```

A 1400×850 native window opens showing the star on the left and the HR diagram on the right. Press `1` through `6` to switch star types, `Space` to start aging, `Tab` to switch view modes.

---

## What You're Looking At

### The Star Viewport

The glowing orb you see is not a texture, a 3D mesh, or a pre-rendered image. It is a numerical solution to a differential equation rendered live on every frame. Specifically:

1. When you select a star type, the application solves the Lane-Emden equation for that star's polytropic index using fourth-order Runge-Kutta integration, producing a table of about 1,500 density values from the star's core to its surface.
2. To draw the star, the renderer takes that density profile and stacks 60 to 80 concentric semi-transparent circles. Each circle's color is a blend between the hot core color and the cooler corona color, weighted by the density at that shell. Each circle's opacity is proportional to the local density.
3. Alpha blending all those circles together creates the illusion of a three-dimensional glowing sphere.

Every pixel you see is backed by actual physics.

### The HR Diagram

The **Hertzsprung-Russell diagram** plots surface temperature (x-axis, reversed by convention — hot on the left) against luminosity (y-axis, logarithmic). It's the single most important chart in stellar astrophysics.

When you age a star, a pulsing colored dot traces the star's **evolutionary track** across the diagram in real time. The color of the dot corresponds to the star's actual current surface temperature using standard spectral-class color conventions (blue-white for O-type, yellow-white for G-type like our Sun, orange-red for M-type). A fading trail of ghost dots shows where it's been.

Twelve real reference stars (Sun, Rigel, Betelgeuse, Sirius A, Sirius B, Vega, Aldebaran, Deneb, Polaris, Antares, Spica, Proxima Centauri) are plotted as background markers so you can see where your simulated star sits relative to the real night sky.

### The Control Panel

The left sidebar lets you customize the simulation. You can click the controls with your mouse or use keyboard shortcuts. It shows real-time readouts including the current Lane-Emden surface value ξ₁, the star's effective temperature and luminosity, its current lifecycle phase, and a live FPS counter.

---

## The Physics in Detail

### Lane-Emden Equation

The full equation is:

```
d²θ/dξ² + (2/ξ)(dθ/dξ) + θⁿ = 0
```

Where θ is the dimensionless density (1 at center, 0 at surface), ξ is the dimensionless radius, and **n** is the **polytropic index** — the parameter that distinguishes different types of stars.

The equation comes from combining three physical requirements: **hydrostatic equilibrium** (the condition that gravity and pressure balance at every point in the star), a **polytropic equation of state** (a simple relation between pressure and density), and **mass conservation** (tracking how much mass is enclosed at each radius).

We solve it numerically because the equation has closed-form analytical solutions for only three specific values of n (namely 0, 1, and 5). Every real star requires numerical integration.

### Polytropic Index Values

| n   | Physical Model                                    | Used For                                |
|-----|---------------------------------------------------|-----------------------------------------|
| 1.5 | Convective / electron-degenerate                  | Red dwarfs, white dwarfs, giant envelopes |
| 3.0 | Radiative equilibrium (Eddington standard model)  | Sun-like stars, blue giants, supergiants |
| 5.0 | Plummer model                                     | Galactic dynamics (not real stars)      |

### Runge-Kutta 4th Order (RK4)

For each step of the numerical integration, RK4 evaluates the slope of the solution at four carefully chosen sample points within the step, then takes a weighted average of those slopes to decide where to go next:

```
k1 = f(x, y)              // slope at start
k2 = f(x + h/2, y + h/2·k1)  // slope at midpoint using k1
k3 = f(x + h/2, y + h/2·k2)  // slope at midpoint using k2
k4 = f(x + h,   y + h·k3)    // slope at end using k3

y_next = y + (h/6)·(k1 + 2·k2 + 2·k3 + k4)
```

The error per step is proportional to h⁵, so halving the step size reduces the error by a factor of 32. Supernova Engine uses h = 0.005 with up to 2,000 steps, terminating when θ reaches zero (the star's surface).

### The Binding Energy Curve and Iron

Every atomic nucleus has a **binding energy per nucleon** — the energy required to pull it apart divided by how many protons and neutrons it contains. When you plot this against atomic number, the curve rises steeply for light elements, peaks at iron-56, and slowly declines for heavier elements.

This single graph determines stellar death. Fusing elements lighter than iron **releases** energy (powering the star). Fusing elements heavier than iron **absorbs** energy. When a massive star's core fills with iron, fusion can no longer sustain the outward pressure that holds gravity at bay. The core collapses in milliseconds, and the star explodes.

### The Chandrasekhar Limit

What a dying star leaves behind depends entirely on the mass of its collapsing core, compared to the **Chandrasekhar limit** (~1.4 solar masses). Subrahmanyan Chandrasekhar derived this limit in 1930 as a 19-year-old during a steamship voyage from India to England.

| Core Mass            | Remnant       | Physics Holding It Up         |
|----------------------|---------------|-------------------------------|
| < 1.4 M☉             | White Dwarf   | Electron degeneracy pressure  |
| 1.4 M☉ – ~3 M☉       | Neutron Star  | Neutron degeneracy pressure   |
| > ~3 M☉              | Black Hole    | Nothing — collapse continues  |

Supernova Engine models all three outcomes. A sun-like star leaves a white dwarf. A blue giant leaves a neutron star (rendered as a pulsar with rotating magnetic beam sweeps). A red supergiant or Wolf-Rayet star leaves a black hole (rendered with accretion disk, event horizon, and gravitational lensing ring).

---

## Project Structure

```
supernova-engine/
├── cmd/stellarforge/
│   └── main.go              Entry point, ebiten.RunGame(), input handling
├── internal/
│   ├── physics/
│   │   ├── solver.go        Lane-Emden RK4 + HR track interpolation
│   │   └── data.go          Star types, elements, reference stars
│   ├── state/
│   │   └── game.go          Simulation state, lifecycle phase logic
│   └── render/
│       ├── star.go          Star viewport + density shell rendering
│       ├── hr.go            Full HR diagram renderer
│       ├── panel.go         Left sidebar control panel
│       └── util.go          Triangle rasterizer, helpers
├── go.mod
└── README.md                (this file)
```

**The package boundary is deliberate:**

- `physics` has **zero** rendering dependencies. You could import it into a CLI tool, a test harness, or a server and solve Lane-Emden equations without ever touching Ebitengine.
- `state` manages the simulation (aging, phase transitions, trail history) without knowing how anything gets drawn.
- `render` is the only package that imports Ebitengine.
- `main` is a thin shell that wires input to state to rendering.

If you ever want to swap Ebitengine for a different renderer (G3N, Raylib, a web frontend), you replace `render/` and leave everything else alone.

---

## Technology Stack

| Layer              | Technology                          |
|--------------------|-------------------------------------|
| Language           | Go 1.22+                            |
| Window / rendering | Ebitengine v2.7 (2D game engine)    |
| Drawing primitives | `ebiten/v2/vector`, `ebitenutil`    |
| Input              | `ebiten/v2/inpututil`               |
| Physics solver     | Pure Go — no external dependencies  |
| Build output       | Single native binary per platform   |

### Why Ebitengine?

Ebitengine is a 2D game engine. It gives you a native OS window via OpenGL (or Metal on macOS, DirectX on Windows) and lets you draw circles, lines, and rectangles at 60 frames per second. It does not do 3D, it does not provide GUI widgets, and it does not ship a scene graph. All of those things are features, not limitations.

The star appears three-dimensional because we stack many semi-transparent 2D circles whose colors and opacities come from the Lane-Emden density solution. This is the same technique used by planetarium software. The HR diagram is lines and dots. The UI panel is rectangles with text. The physics is the interesting part. The rendering tech stays out of the way.

If you want actual 3D rendering — orbit cameras, volumetric ray marching, custom GLSL shaders — you'd swap Ebitengine for G3N or Raylib. That's on the roadmap, not here.

---

## Controls Reference

### Keyboard

| Key         | Action                              |
|-------------|-------------------------------------|
| `1` – `6`   | Select star type                    |
| `Space`     | Toggle aging on/off                 |
| `R`         | Reset age to zero                   |
| `+` / `-`   | Double / halve aging speed          |
| `[` / `]`   | Decrease / increase brightness      |
| `Tab`       | Cycle view mode                     |
| `P`         | Toggle density profile plot         |
| `Q W E`     | Toggle Hydrogen, Helium, Carbon     |
| `A S D F`   | Toggle Nitrogen, Oxygen, Silicon, Iron |

### Mouse

All panel buttons are clickable. Hover highlights, active state is visible.

### View Modes

- **Split** — star on the left, HR diagram on the right (default)
- **Star Only** — full-screen star viewport
- **HR Only** — full-screen HR diagram

---

## Star Types

Each star type has its own polytropic index, mass, temperature range, default elemental composition, and hand-authored HR evolutionary track based on published stellar models.

| Key | Name                | Mass (M☉) | n   | Max Age (Gyr) | Final Fate                     |
|-----|---------------------|-----------|-----|---------------|--------------------------------|
| `1` | Red Dwarf (M-type)  | 0.3       | 1.5 | 100           | White Dwarf                    |
| `2` | Sun-like (G-type)   | 1.0       | 3.0 | 10            | White Dwarf                    |
| `3` | Blue Giant (O/B)    | 15        | 3.0 | 0.05          | Supernova → Neutron Star       |
| `4` | Red Supergiant      | 20        | 3.0 | 0.03          | Supernova → Black Hole         |
| `5` | White Dwarf         | 0.6       | 1.5 | 1000          | Black Dwarf (theoretical)      |
| `6` | Wolf-Rayet Star     | 25        | 3.0 | 0.01          | Supernova → Black Hole         |

## Elements

Toggle the chemical composition of your star. During the main sequence phase, each active element is displayed as a concentric shell indicator at a depth that reflects where fusion of that element occurs in real stars.

| Key | Element  | Role in Stellar Fusion                         |
|-----|----------|------------------------------------------------|
| `Q` | Hydrogen | The main sequence fuel (pp-chain and CNO cycle) |
| `W` | Helium   | Fusion product of H; fuel for triple-alpha process |
| `E` | Carbon   | Triple-alpha fusion product                     |
| `A` | Nitrogen | Catalyst in the CNO cycle                       |
| `S` | Oxygen   | Late-stage fusion product                       |
| `D` | Silicon  | Last element before iron                        |
| `F` | Iron     | The end. Fusing it absorbs energy rather than releasing it. |

---

## Lifecycle Phases

As you age a star, it passes through distinct phases driven by the same age fraction for every star type (though the absolute timescales differ by factors of thousands between red dwarfs and Wolf-Rayets).

| Age Fraction     | Phase                  | What's Happening                             |
|------------------|------------------------|----------------------------------------------|
| 0% – 70%         | Main Sequence          | Stable hydrogen fusion, on the main sequence band |
| 70% – 90%        | Red Giant              | Hydrogen core exhausted, envelope expands    |
| 90% – 100%       | Supernova              | Core collapse (only for massive stars)       |
| 100%+ (low mass) | White Dwarf Remnant    | Cooling forever                              |
| 100%+ (med mass) | Neutron Star / Pulsar  | Rapid rotation, magnetic beam sweeps         |
| 100%+ (high mass)| Black Hole             | Event horizon + accretion disk               |

---

## What Each Rendering Mode Draws

### Main Sequence / Red Giant / White Dwarf

- Lane-Emden density shells (60–80 layered translucent circles)
- Core glow (6 additive brightness layers)
- Corona (4 large faint layers)
- Element composition rings (main sequence only)

### Supernova

- Expanding fireball (phase 1)
- Rebound contraction (phase 2)
- Final core collapse (phase 3)
- 100 ejection particles flying outward at random angles

### Neutron Star / Pulsar

- Tiny, extremely dense core
- Two rotating magnetic beam sweeps, rendered as linear dot trails
- Rapid pulse (20× the normal pulse speed)

### Black Hole

- Multilayer accretion disk (50 rings, color-shifted by ring index)
- Solid black event horizon
- Pulsing gravitational lensing ring (violet)
- Photon sphere (gold, smaller radius)

---

## Roadmap

Ideas for future versions, in rough order of difficulty:

- Save/load star configurations as JSON
- Sound effects using Ebitengine's audio API
- pp-chain and CNO cycle fusion chain visualizations
- Accurate Chandrasekhar limit calculation derived from n=3 Lane-Emden solution
- Binary star systems with Roche lobe overflow
- Migration to 3D rendering via G3N or Raylib
- Custom GLSL shader for volumetric ray-marched star rendering
- Exoplanet orbital visualization for stars in habitable zone

Contributions welcome. The physics package is the easiest place to start if you want to extend the simulation without touching graphics code.

---

## License

MIT. See LICENSE file.

---

## Further Reading

If you want to dig deeper into the physics:

- **Stellar Structure and Evolution** by Rudolf Kippenhahn, Alfred Weigert, and Achim Weiss — the textbook on this subject
- **An Introduction to Modern Astrophysics** by Carroll and Ostlie — more accessible, graduate level
- **The Internal Constitution of the Stars** by Arthur Eddington (1926) — the book that invented most of modern stellar astrophysics, and is surprisingly readable
- **Black Holes, White Dwarfs, and Neutron Stars** by Shapiro and Teukolsky — the definitive treatment of compact objects

Sir Arthur Eddington's book is in the public domain and available at Project Gutenberg. It's worth your time.

---

*Built by [@saedarm](https://github.com/saedarm). Read the blog post about how and why: [I Built a Supernova Engine in Go Using a 150-Year-Old Equation](https://medium.com/@sarouth).*
