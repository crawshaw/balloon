// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !portable

package main

import (
	"log"
	"sync"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/app/debug"
	"golang.org/x/mobile/gl"
	"golang.org/x/mobile/sprite/glsprite"
)

var once sync.Once

func glinit() {
	eng = glsprite.Engine()
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
	once.Do(glinit)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
	gl.ClearColor(1, 1, 1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	t := now()
	updateGame(t)
	eng.Render(scene, t)

	debug.DrawFPS()
}
