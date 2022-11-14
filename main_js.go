//go:build js

package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed assets
var assets embed.FS

var isJS = true

func focus() {
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementsByTagName", "canvas").Index(0)
	canvas.Call("focus")
}

func parseConfig() config {
	return config{}
}

func getImages() ([]galleryImage, error) {
	images := []galleryImage{}
	metadata := map[string]string{}
	mc, err := assets.ReadFile("assets/metadata.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(mc, &metadata); err != nil {
		return nil, err
	}

	err = fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".png") {
			cs, err := assets.ReadFile(path)
			if err != nil {
				return err
			}
			img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(cs))
			if err != nil {
				return err
			}
			fmt.Println("???", metadata[path])
			images = append(images, galleryImage{Desc: metadata[path], Image: img})

		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return images, nil
}
