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
	"runtime"
	"time"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"

	"github.com/crawshaw/balloon/animation"
	"github.com/crawshaw/balloon/text"
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
	overScene *sprite.Node
)

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
	overSceneInit()
	scene = menuScene
}

func menuSceneInit() {
	menuScene = new(sprite.Node)
	eng.Register(menuScene)

	addGopher := func(offsetX, size geom.Pt, subTex sprite.SubTex, duration int) {
		gopher := &sprite.Node{
			Arranger: &animation.Arrangement{
				Offset: geom.Point{X: offsetX, Y: -size},
				Size:   &geom.Point{size, size},
				Pivot:  geom.Point{size / 2, size / 2},
				SubTex: subTex,
			},
		}
		eng.Register(gopher)
		menuScene.AppendChild(gopher)

		gopherAnim := new(sprite.Node)
		eng.Register(gopherAnim)
		menuScene.AppendChild(gopherAnim)
		gopherAnim.Arranger = &animation.Animation{
			Current: "init",
			States: map[string]animation.State{
				"init": animation.State{
					Duration: duration / 4,
					Next:     "falling",
				},
				"falling": animation.State{
					Duration: duration,
					Next:     "reset",
					Transforms: map[*sprite.Node]animation.Transform{
						gopher: animation.Transform{
							Transformer: animation.Move{Y: geom.Height + size*2},
						},
					},
				},
				"reset": animation.State{
					Duration: 0,
					Next:     "falling",
					Transforms: map[*sprite.Node]animation.Transform{
						gopher: animation.Transform{
							Transformer: animation.Move{Y: -geom.Height - size*2},
						},
					},
				},
			},
		}
	}

	addGopher(24, 36, sheet.gopherSwim, 240)
	addGopher(48, 18, sheet.gopherRun, 100)
	addGopher(96, 36, sheet.gopherSwim, 160)

	addText(menuScene, "Gopher Fall!", 20, geom.Point{24, 24})
	addText(menuScene, "Tap to start", 14, geom.Point{48, 48})
}

func overSceneInit() {
	overScene = new(sprite.Node)
	eng.Register(overScene)

	addText(overScene, "GAME OVER", 20, geom.Point{28, 28})
	addText(overScene, "Tap to play again", 14, geom.Point{32, 48})
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

	game.scissor = newScissorArm2(eng)
	game.scissor.arrangement.Offset.Y = 2 * 72
	gameScene.AppendChild(game.scissor.node)

	n1 := new(sprite.Node)
	eng.Register(n1)
	n1.Arranger = &animation.Arrangement{
		Offset: geom.Point{X: 0, Y: geom.Height - 12 - 2},
	}
	gameScene.AppendChild(n1)

	t := new(sprite.Node)
	eng.Register(t)
	n1.AppendChild(t)
	game.scoreText = &text.String{
		Size:  12,
		Color: color.Black,
		Font:  font,
	}
	t.Arranger = game.scoreText

	updateGame(0)

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
	f, err := app.Open("balloon_sheet.png")
	if err != nil {
		return err
	}
	mb, err := ioutil.ReadAll(f)
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
	if e.Type != event.TouchStart {
		return
	}

	switch scene {
	case overScene:
		scene = menuScene
	case menuScene:
		startGame()
		scene = gameScene
	case gameScene:
		game.nextTouch = &e
	default:
		log.Printf("touch in unknown state %v", e)
	}
}

func now() clock.Time {
	d := time.Since(start)
	return clock.Time(60 * d / time.Second)
}
