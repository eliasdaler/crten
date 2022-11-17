// Copyright (c) 2022 Elias Daler
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of version 3 of the GNU General Public
// License as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// crten - Cathode-Ray Tube ENgine
// Display/render pixel art with a CRT effect.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/f64"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

var ScreenX = 256
var ScreenY = 240

const fontSize = 32

const version = "0.1" // TODO: don't hardcode

type ShaderParam struct {
	Name string
	Val  float32
	Min  float32
	Max  float32
	Step float32
}

//go:embed shaders/crt-lottes.go
var shaderSrc []byte

//go:embed assets/mastershouse.png
var imgFile []byte

//go:embed m5x7.ttf
var fontFile []byte

type galleryImage struct {
	Desc  string
	Image *ebiten.Image
}

type config struct {
	InputPath  string
	OutputPath string
	ConfigPath string
	NoClose    bool
}

type Game struct {
	config config

	shader        *ebiten.Shader
	ShaderParams  []ShaderParam
	DefaultValues []float32

	img *ebiten.Image

	windowSize   f64.Vec2
	defaultScale int

	font font.Face

	currMenuIndex int
	showHelp      bool

	firstCursorMove bool
	cursorAccum     int

	images         []galleryImage
	currImageIndex int

	closeWindowOnTextFrame bool
}

type InputAction int

const (
	InputActionCursorDown InputAction = iota
	InputActionCursorUp
	InputActionValueDown
	InputActionValueUp
	InputActionPrevImage
	InputActionNextImage
	InputActionResetParams
)

func (g *Game) loadFont() error {
	tt, err := opentype.Parse(fontFile)
	if err != nil {
		return err
	}

	const dpi = 72
	g.font, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Game) loadInputConfig() error {
	cs, err := os.ReadFile(g.config.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	var c struct {
		Shader string             `json:"shader"` // currently unused
		Params map[string]float32 `json:"params"`
		Scale  int                `json:"scale"`
	}
	if err := json.Unmarshal(cs, &c); err != nil {
		log.Fatal(err)
	}

	for pn, v := range c.Params {
		found := false
		// not the most efficient way to find a param by name, but whatever
		for i := 0; i < len(g.ShaderParams); i++ {
			p := &g.ShaderParams[i]
			if p.Name != pn {
				continue
			}

			p.Val = v
			found = true
		}
		if !found {
			return fmt.Errorf("[error] unknown parameter name %q", pn)
		}
	}

	if c.Scale != 0 {
		g.defaultScale = c.Scale
	}
	return nil
}

func (g *Game) onInput(a InputAction) {
	switch a {
	case InputActionCursorDown:
		g.currMenuIndex++
		if g.currMenuIndex >= len(g.ShaderParams) {
			g.currMenuIndex = 0
		}
	case InputActionCursorUp:
		g.currMenuIndex--
		if g.currMenuIndex < 0 {
			g.currMenuIndex = len(g.ShaderParams) - 1
		}
	case InputActionValueDown:
		p := &g.ShaderParams[g.currMenuIndex]
		p.Val -= p.Step
		if p.Val < p.Min {
			p.Val = p.Min
		}
	case InputActionValueUp:
		p := &g.ShaderParams[g.currMenuIndex]
		p.Val += p.Step
		if p.Val > p.Max {
			p.Val = p.Max
		}
	case InputActionPrevImage:
		g.currImageIndex--
		if g.currImageIndex < 0 {
			g.currImageIndex = len(g.images) - 1
		}
		g.onImageChanged()
	case InputActionNextImage:
		g.currImageIndex++
		if g.currImageIndex >= len(g.images) {
			g.currImageIndex = 0
		}
		g.onImageChanged()
	case InputActionResetParams:
		for i := 0; i < len(g.ShaderParams); i++ {
			g.ShaderParams[i].Val = g.DefaultValues[i]
		}
	}
}

func (g *Game) onImageChanged() {
	g.img = g.images[g.currImageIndex].Image
	ScreenX, ScreenY = g.img.Size()
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.showHelp = !g.showHelp
	}

	if !g.showHelp {
		return nil
	}

	// R - reset
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.onInput(InputActionResetParams)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.onInput(InputActionPrevImage)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		g.onInput(InputActionNextImage)
	}

	// Down
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.firstCursorMove = true
		g.cursorAccum = 0
		g.onInput(InputActionCursorDown)
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		g.cursorAccum = 0
	}

	// Up
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.firstCursorMove = true
		g.cursorAccum = 0
		g.onInput(InputActionCursorUp)
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		g.cursorAccum = 0
	}

	// Left
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.firstCursorMove = true
		g.cursorAccum = 0
		g.onInput(InputActionValueDown)
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyLeft) {
		g.cursorAccum = 0
	}

	// Right
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.firstCursorMove = true
		g.cursorAccum = 0
		g.onInput(InputActionValueUp)
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyRight) {
		g.cursorAccum = 0
	}

	// Key repeat behaviour for cursor
	if ebiten.IsKeyPressed(ebiten.KeyDown) ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeyLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.cursorAccum++
		delay := 10
		if g.firstCursorMove {
			delay = 30
		}
		if g.cursorAccum == delay {
			g.firstCursorMove = false
			g.cursorAccum = 0
			if ebiten.IsKeyPressed(ebiten.KeyDown) {
				g.onInput(InputActionCursorDown)
			} else if ebiten.IsKeyPressed(ebiten.KeyUp) {
				g.onInput(InputActionCursorUp)
			} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				g.onInput(InputActionValueDown)
			} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
				g.onInput(InputActionValueUp)
			}
		}
	}

	if g.closeWindowOnTextFrame {
		return fmt.Errorf("exiting...")
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.windowSize[0] = float64(outsideWidth)
	g.windowSize[1] = float64(outsideHeight)
	return outsideWidth, outsideHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawCRTImage(screen)

	if g.config.OutputPath != "" { // dump image to file
		g.renderToImage(screen)
		if g.config.NoClose {
			g.closeWindowOnTextFrame = false
		}
		return
	}

	if g.showHelp {
		g.drawUI(screen)
	}
}

func (g *Game) drawCRTImage(screen *ebiten.Image) {
	tw, th := g.img.Size()
	l := CalculateLetterBox(g.windowSize, f64.Vec2{float64(tw), float64(th)})

	if g.config.OutputPath != "" {
		l.Scale = float64(g.defaultScale)
	}

	// draw at pixel perfect scale
	// even if we have a bigger screen, we want to stay pixel perfect
	w := tw * int(l.Scale)
	h := th * int(l.Scale)

	sop := &ebiten.DrawRectShaderOptions{}
	sop.GeoM.Scale(l.Scale, l.Scale)
	sop.Uniforms = map[string]any{
		"ScreenSize":  []float32{float32(w), float32(h)},
		"TextureSize": []float32{float32(tw), float32(th)},
	}
	for _, p := range g.ShaderParams {
		sop.Uniforms[p.Name] = p.Val
	}
	sop.Images[0] = g.img

	// recalculate letter box - we are drawing a scaled texture now
	l = CalculateLetterBox(g.windowSize, f64.Vec2{float64(w), float64(h)})
	sop.GeoM.Concat(l.GetTransform())

	screen.Clear()
	screen.DrawRectShader(ScreenX, ScreenY, g.shader, sop)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	st := fontSize / 2

	y := st * 2
	x := 8
	drawTextWithShadow(screen, "crten v"+version, g.font, x, y, color.White)
	y += st
	desc := g.images[g.currImageIndex].Desc
	drawTextWithShadow(screen, "Art: "+desc, g.font, x, y, color.White)
	y += st * 2
	drawTextWithShadow(screen, "[F1 - show/hide UI]", g.font, x+1, y+1, color.White)
	y += st
	drawTextWithShadow(screen, "[Arrow keys - adjust params, R to reset]", g.font, x, y, color.White)
	y += st
	drawTextWithShadow(screen, "[Z/X - switch images]", g.font, x, y, color.White)

	y += st * 2
	for i, p := range g.ShaderParams {
		isSelected := g.currMenuIndex == i
		t := getParamText(p, isSelected)
		drawTextWithShadow(screen, t, g.font, x, y, color.White)
		if isSelected { // hack: highlight cursor and param name
			c := color.RGBA{175, 233, 233, 255}
			drawTextWithShadow(screen, " => "+p.Name, g.font, x, y, c)
		}
		y += st
	}
}

func drawTextWithShadow(screen *ebiten.Image, str string, font font.Face, x int, y int, c color.Color) {
	text.Draw(screen, str, font, x+1, y+1, color.Black)
	text.Draw(screen, str, font, x, y, c)
}

func getParamText(p ShaderParam, isSelected bool) string {
	pref := "    "
	arrLeft := " < "
	arrRight := " >"
	if isSelected {
		pref = " => "
		if p.Val == p.Min {
			arrLeft = " "
		}
		if p.Val == p.Max {
			arrRight = ""
		}
	} else {
		arrLeft = " "
		arrRight = " "
	}

	var desc string
	if p.Name == "ShadowMask" {
		switch p.Val {
		case 0.0:
			desc = "(None)"
		case 1.0:
			desc = "(Compressed TV style)"
		case 2.0:
			desc = "(Aperture-grille)"
		case 3.0:
			desc = "(Stretched VGA style)"
		case 4.0:
			desc = "(VGA style)"
		default:
			panic("??")
		}
	}

	return fmt.Sprintf("%s%s:%s%.3f%s %s", pref, p.Name, arrLeft, p.Val, arrRight, desc)
}

func (g *Game) renderToImage(screen *ebiten.Image) error {
	log.Printf("saving to %s...\n", g.config.OutputPath)
	f, err := os.Create(g.config.OutputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	g.closeWindowOnTextFrame = true
	return png.Encode(f, screen)
}

func main() {
	config := parseConfig()

	g := &Game{
		config:       config,
		defaultScale: 4,
	}
	focus()

	if config.InputPath != "" {
		f, err := os.Open(config.InputPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		img, _, err := ebitenutil.NewImageFromReader(f)
		if err != nil {
			log.Fatal(err)
		}
		g.images = []galleryImage{{Desc: config.InputPath, Image: img}}
	} else { // no path specified - show default images
		var err error
		g.images, err = getImages()
		if err != nil {
			log.Fatal(err)
		}
	}
	g.onImageChanged()

	shader, err := ebiten.NewShader(shaderSrc)
	if err != nil {
		log.Fatal(err)
	}
	g.shader = shader

	g.ShaderParams = []ShaderParam{
		{"HardScan", -10., -20., 0., 1.},
		{"HardPix", -4., -20., 0., 1.},
		{"WarpX", 0.01, 0.0, 0.125, 0.01},
		{"WarpY", 0.02, 0.0, 0.125, 0.01},
		{"MaskDark", 0.5, 0.0, 2.0, 0.1},
		{"MaskLight", 1.5, 0.0, 2.0, 0.1},
		{"ShadowMask", 0.0, 0.0, 4.0, 1.0},
		{"BrightBoost", 1.0, 0.0, 2.0, 0.05},
		{"HardBloomPix", -1.5, -2.0, -0.5, 0.1},
		{"HardBloomScan", -2.0, -4.0, -1.0, 0.1},
		{"BloomAmount", 0.05, 0.0, 1.0, 0.05},
		{"Shape", 2.0, 0.0, 10.0, 0.05},
	}
	for _, p := range g.ShaderParams {
		g.DefaultValues = append(g.DefaultValues, p.Val)
	}

	if g.config.ConfigPath != "" {
		g.loadInputConfig()
	}

	g.showHelp = true
	if err := g.loadFont(); err != nil {
		log.Fatal(err)
	}

	if g.config.OutputPath != "" {
		ebiten.SetScreenTransparent(true)
		ebiten.SetWindowDecorated(false)
		ebiten.SetWindowFloating(true)
	} else {
		ebiten.SetWindowTitle("CRT shader demo")
		ebiten.SetWindowResizable(false)
	}
	ebiten.SetWindowSize(ScreenX*g.defaultScale, ScreenY*g.defaultScale)

	if err := ebiten.RunGame(g); err != nil {
		if err.Error() == "exiting..." {
			return
		}
		log.Fatal(err)
	}
}
