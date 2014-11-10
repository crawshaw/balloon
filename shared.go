// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"
	"golang.org/x/mobile/sprite/text"

	"github.com/crawshaw/balloon/animation"
)

var sheet struct {
	sheet      sprite.Texture
	balloon    sprite.SubTex
	arm        sprite.SubTex
	pad        sprite.SubTex
	gopherSwim sprite.SubTex
	gopherRun  sprite.SubTex
}

var (
	start time.Time
	eng   sprite.Engine
	font  *truetype.Font
)

var (
	scene     *sprite.Node
	gameScene *sprite.Node
	menuScene *sprite.Node
)

var scissor *scissorArm2
var (
	scoreText *text.String
	score     int
	lives     int
)

func writeScore() {
	scoreText.Text = fmt.Sprintf("score: %d  lives: %d", score, lives)
}

func timerInit() {
	start = time.Now()
	if err := loadSheet(); err != nil {
		log.Fatal(err)
	}

	var err error
	font, err = loadFont()
	if err != nil {
		panic(err)
	}

	menuSceneInit()
	gameSceneInit()
	scene = menuScene
}

func menuSceneInit() {
	menuScene = new(sprite.Node)
	eng.Register(menuScene)

	addText(menuScene, "Gopher Run!", 20, geom.Point{24, 24})
	addText(menuScene, "Tap to start", 14, geom.Point{48, 48})
}

func addText(parent *sprite.Node, str string, size geom.Pt, pos geom.Point) {
	p := &sprite.Node{
		Arranger: &animation.Arrangement{
			Offset: pos,
		},
	}
	eng.Register(p)
	parent.AppendChild(p)
	pText := &sprite.Node{
		Arranger: &text.String{
			Size:  size,
			Color: color.Black,
			Font:  font,
			Text:  str,
		},
	}
	eng.Register(pText)
	p.AppendChild(pText)
}

func gameSceneInit() {
	gameScene = new(sprite.Node)
	eng.Register(gameScene)

	scissor = newScissorArm2(eng)
	scissor.arrangement.Offset.Y = 2 * 72
	gameScene.AppendChild(scissor.node)

	n1 := new(sprite.Node)
	eng.Register(n1)
	n1.Arranger = &animation.Arrangement{
		Offset: geom.Point{X: 0, Y: geom.Height - 12 - 2},
	}
	gameScene.AppendChild(n1)

	t := new(sprite.Node)
	eng.Register(t)
	n1.AppendChild(t)
	scoreText = &text.String{
		Size:  12,
		Color: color.Black,
		Font:  font,
	}
	t.Arranger = scoreText

	score = 0
	lives = 3
	writeScore()

	//Fprint(os.Stdout, gameScene, NotNilFilter)
}

func loadFont() (*truetype.Font, error) {
	font := ""
	switch runtime.GOOS {
	case "android":
		font = "/system/fonts/DroidSansMono.ttf"
	case "darwin":
		//font = "/Library/Fonts/Andale Mono.ttf"
		font = "/Library/Fonts/Arial.ttf"
		//font = "/Library/Fonts/儷宋 Pro.ttf"
	case "linux":
		font = "/usr/share/fonts/truetype/droid/DroidSansMono.ttf"
	default:
		return nil, fmt.Errorf("go.mobile/app/debug: unsupported runtime.GOOS %q", runtime.GOOS)
	}
	b, err := ioutil.ReadFile(font)
	if err != nil {
		return nil, err
	}
	return freetype.ParseFont(b)
}

func loadSheet() error {
	mb, err := ioutil.ReadFile(filepath.FromSlash("balloon_sheet.png"))
	if err != nil {
		return err
	}
	m, err := png.Decode(bytes.NewReader(mb))
	if err != nil {
		return err
	}
	t, err := eng.LoadTexture(m)
	if err != nil {
		return err
	}
	sheet.sheet = t

	sheet.arm = sprite.SubTex{t, image.Rect(0, 0, 194, 42)}
	sheet.balloon = sprite.SubTex{t, image.Rect(0, 42, 148, 634)}
	sheet.pad = sprite.SubTex{t, image.Rect(194, 0, 294, 286)}
	sheet.gopherSwim = sprite.SubTex{t, image.Rect(188, 288, 294, 380)}
	sheet.gopherRun = sprite.SubTex{t, image.Rect(194, 380, 240, 440)}

	return nil
}

func touch(e event.Touch) {
	fmt.Printf("touch: %+v\n", e)
	if e.Type == event.TouchStart {
		scissor.touch(now(), e)
		return
	}
}

func now() clock.Time {
	d := time.Since(start)
	return clock.Time(60 * d / time.Second)
}
