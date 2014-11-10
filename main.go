// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build portable

// A 2D android "game" with a running gopher.
package main

import (
	"log"
	"math"
	"sync"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/app/debug"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/gl/glutil"
	"golang.org/x/mobile/sprite/portable"
)

var fb struct {
	sync.Once
	*glutil.Image
}

func fbinit() {
	start = time.Now()
	toPx := func(x geom.Pt) int { return int(math.Ceil(float64(geom.Pt(x).Px()))) }
	fb.Image = glutil.NewImage(toPx(geom.Width), toPx(geom.Height))
	eng = portable.Engine(fb.Image.RGBA)

	timerInit()
	menuSceneInit()
	gameSceneInit()
	scene = menuScene
}

func main() {
	log.Print("starting spriterun")

	app.Run(app.Callbacks{
		Draw:  drawWindow,
		Touch: touch,
	})
}

func drawWindow() {
	fb.Do(fbinit)

	for i := range fb.Image.RGBA.Pix {
		fb.Image.RGBA.Pix[i] = 0xff // white background
	}
	t := now()
	eng.Render(scene, t)
	fb.Upload()
	fb.Draw(
		geom.Point{},
		geom.Point{geom.Width, 0},
		geom.Point{0, geom.Height},
		fb.Bounds(),
	)
	debug.DrawFPS()
}
