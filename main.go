// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A 2D android "game" with a running gopher.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"code.google.com/p/go.mobile/app"
	"code.google.com/p/go.mobile/app/debug"
	"code.google.com/p/go.mobile/event"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/gl/glutil"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"
	"code.google.com/p/go.mobile/sprite/portable"

	"github.com/crawshaw/balloon/animation"
)

var fb struct {
	sync.Once
	*glutil.Image
}

var sheet struct {
	sheet   sprite.Sheet
	balloon sprite.Texture
	arm     sprite.Texture
	pad     sprite.Texture
}

var (
	start time.Time
	eng   sprite.Engine
	scene *sprite.Node

	scissor *scissorArm2
)

func fbinit() {
	start = time.Now()
	toPx := func(x geom.Pt) int { return int(math.Ceil(float64(geom.Pt(x).Px()))) }
	fb.Image = glutil.NewImage(toPx(geom.Width), toPx(geom.Height))
	eng = portable.Engine(fb.Image.RGBA)
	if err := loadSheet(); err != nil {
		log.Fatal(err)
	}

	scene = new(sprite.Node)
	eng.Register(scene)

	b := new(sprite.Node)
	eng.Register(b)
	scene.AppendChild(b)

	a := &animation.Arrangement{
		Offset:  geom.Point{X: 36, Y: 36},
		Size:    &geom.Point{X: 36, Y: 4 * 36},
		Texture: sheet.balloon,
	}
	b.Arranger = a

	b = new(sprite.Node)
	eng.Register(b)
	scene.AppendChild(b)

	b.Arranger = &animation.Arrangement{
		Offset:  geom.Point{X: 136, Y: 36},
		Size:    &geom.Point{X: 12, Y: 36},
		Texture: sheet.pad,
	}

	scissor = newScissorArm2(eng)
	scissor.arrangement.Offset.Y = 24
	scene.AppendChild(scissor.node)

	Fprint(os.Stdout, scene, NotNilFilter)
}

func loadSheet() error {
	b, err := ioutil.ReadFile(filepath.FromSlash("balloon_sheet.png"))
	if err != nil {
		return err
	}
	m, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return err
	}
	s, err := eng.LoadSheet(m)
	if err != nil {
		return err
	}
	sheet.sheet = s

	sheet.arm, err = eng.LoadTexture(s, image.Rect(0, 0, 194, 42))
	if err != nil {
		return fmt.Errorf("arm: %v", err)
	}
	sheet.balloon, err = eng.LoadTexture(s, image.Rect(0, 42, 148, 634))
	if err != nil {
		return fmt.Errorf("ballon: %v", err)
	}
	sheet.pad, err = eng.LoadTexture(s, image.Rect(194, 0, 294, 286))
	if err != nil {
		return fmt.Errorf("pad: %v", err)
	}

	return nil
}

func main() {
	log.Print("starting spriterun")

	app.Run(app.Callbacks{
		Draw:  drawWindow,
		Touch: touch,
	})
}

func now() clock.Time {
	d := time.Since(start)
	return clock.Time(60 * d / time.Second)
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

func touch(e event.Touch) {
	fmt.Printf("touch: %+v\n", e)
	if e.Type == event.TouchStart {
		scissor.touch(now(), e)
		return
	}
}
