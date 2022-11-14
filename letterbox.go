package main

import (
	"math"

	"golang.org/x/image/math/f64"

	"github.com/hajimehoshi/ebiten/v2"
)

type LetterBox struct {
	Scale float64
	Pos   f64.Vec2
}

func CalculateLetterBox(windowSize, screenSize f64.Vec2) LetterBox {
	var scale float64
	if windowSize[0] < screenSize[1] || windowSize[1] < screenSize[1] {
		scaleX := float64(windowSize[0]) / float64(screenSize[0])
		scaleY := float64(windowSize[1]) / float64(screenSize[1])
		scale = math.Min(scaleX, scaleY)
	} else {
		scaleX := math.Floor(float64(windowSize[0]) / float64(screenSize[0]))
		scaleY := math.Floor(float64(windowSize[1]) / float64(screenSize[1]))
		scale = math.Min(scaleX, scaleY)
	}

	return LetterBox{
		Scale: scale,
		Pos: f64.Vec2{
			math.Floor(float64(windowSize[0])/2. - float64(screenSize[0])*scale/2.),
			math.Floor(float64(windowSize[1])/2. - float64(screenSize[1])*scale/2.),
		},
	}
}

func (l LetterBox) GetTransform() ebiten.GeoM {
	m := ebiten.GeoM{}
	m.Scale(l.Scale, l.Scale)
	m.Translate(l.Pos[0], l.Pos[1])
	return m
}
