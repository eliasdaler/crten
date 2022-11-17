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

//go:build !js

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func focus() {}

func parseConfig() config {
	inputPath := flag.String("i", "", "input image")
	configPath := flag.String("c", "", "custom config JSON")
	noClose := flag.Bool("noclose", false, "if true, window doesn't close after converting an image")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			`crten v0.1, usage:
	crten IMAGE_FILE - display INPUT_FILE with CRT effect
	crten -i INPUT_FILE [-c CONFIG_FILE] OUTPUT_FILE - renders INPUT_FILE image with CRT effect to OUTPUT_FILE and closes the window
`)
		flag.PrintDefaults()
	}
	flag.Parse()

	var c config
	c.ConfigPath = *configPath
	if *inputPath == "" {
		if flag.NArg() == 1 {
			c.InputPath = flag.Arg(0)
		}
	} else {
		c.InputPath = *inputPath
		if flag.NArg() == 1 {
			c.OutputPath = flag.Arg(0)
		}
	}
	if *noClose {
		c.NoClose = true
	}

	return c
}

func getImages() ([]galleryImage, error) {
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(imgFile))
	if err != nil {
		return nil, err
	}

	return []galleryImage{{Desc: "'Master's House' by Elias Daler (@eliasdaler)", Image: img}}, nil
}
