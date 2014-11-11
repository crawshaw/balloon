// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math/rand"

	"golang.org/x/mobile/event"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"
	"golang.org/x/mobile/sprite/text"

	"github.com/crawshaw/balloon/animation"
)

const expandTime = 10

var game struct {
	balloon   *animation.Arrangement
	nextTouch *event.Touch
	scissor   *scissorArm2
	scoreText *text.String
	score     int
	lives     int

	gophers []*animation.Arrangement

	dropGopher *animation.Arrangement
	dropStart  clock.Time
	dropEnd    clock.Time
	dropSave   clock.Time
	dropSaveY  geom.Pt
}

const initialPause = 120

func startGame() {
	game.score = 0
	game.lives = 3
	game.dropEnd = 0

	addGopher := func(size geom.Pt, subTex sprite.SubTex) {
		gopher := &sprite.Node{
			Arranger: &animation.Arrangement{
				Offset: geom.Point{X: 0, Y: -size},
				Size:   &geom.Point{size, size},
				Pivot:  geom.Point{size / 2, size / 2},
				SubTex: subTex,
			},
		}
		eng.Register(gopher)
		gameScene.AppendChild(gopher)

		game.gophers = append(game.gophers, gopher.Arranger.(*animation.Arrangement))
	}

	addGopher(36, sheet.gopherSwim)
	addGopher(18, sheet.gopherRun)

	b := new(sprite.Node)
	eng.Register(b)
	gameScene.AppendChild(b)
	game.balloon = &animation.Arrangement{
		Pivot:  geom.Point{X: 6, Y: 72},
		Size:   &geom.Point{X: 24, Y: 72},
		SubTex: sheet.balloon,
		Hidden: true,
	}
	b.Arranger = game.balloon
}

func updateGame(t clock.Time) {
	if scene == menuScene {
		return
	}
	if game.lives == 0 {
		scene = overScene
		return
	}

	minX, maxX := game.scissor.balloonTravel()

	if game.dropGopher != nil {
		g := game.dropGopher
		if game.dropSave > 0 && t > game.dropSave {
			// float away
			tween := clock.Linear(game.dropSave, game.dropEnd, t)
			b := game.balloon
			g.Offset.Y = -g.Size.Y*2 + (game.dropSaveY+g.Size.Y*2)*geom.Pt(1-tween)
			b.Offset.X = g.Offset.X
			b.Offset.Y = g.Offset.Y
			b.Hidden = false
		} else {
			// fall to doom
			tween := clock.Linear(game.dropStart, game.dropEnd, t)
			g.Offset.Y = (geom.Height + g.Size.Y*2) * geom.Pt(tween)
		}
	}

	if t > initialPause && t > game.dropEnd {
		if game.dropGopher != nil {
			// push old gopher back to the top.
			if game.dropSave == 0 {
				game.lives--
			}
			game.dropSave = 0
			game.dropGopher.Offset.Y = -game.dropGopher.Size.Y
			game.dropGopher = nil
			game.balloon.Hidden = true
		}

		duration := 80
		if game.score < 90 {
			duration -= game.score
		}
		game.dropStart = t
		game.dropEnd = t + clock.Time(duration)

		num := rand.Intn(len(game.gophers))
		game.dropGopher = game.gophers[num]
		minX := minX + 5 // you have to move the balloon to score
		game.dropGopher.Offset.X = minX + (maxX-minX)*geom.Pt(rand.Float32())
	}

	if game.nextTouch != nil && game.scissor.a.Current == "ready" {
		y := game.nextTouch.Loc.Y

		if t < game.dropEnd {
			// check for intersection of balloon and gopher
			g := game.dropGopher
			startY := game.dropStart
			startX := t
			durationY := game.dropEnd - game.dropStart
			durationX := expandTime

			x := g.Offset.X
			minY := geom.Pt(0)
			maxY := geom.Height + g.Size.Y*2

			gopherT := int((y-minY)/(maxY-minY)*geom.Pt(durationY)) + int(startY)
			balloonT := int((x-minX)/(maxX-minX)*geom.Pt(durationX)) + int(startX)

			diff := gopherT - balloonT
			if diff < 0 {
				diff = -diff
			}
			if diff < 4 {
				game.dropSave = clock.Time(gopherT)
				game.dropSaveY = y
				game.score++
			}
		}

		game.scissor.arrangement.Offset.Y = y
		game.scissor.a.Transition(t, "expanding")
	}
	game.nextTouch = nil

	game.scoreText.Text = fmt.Sprintf("score: %d  lives: %d", game.score, game.lives)
}
