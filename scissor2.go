package main

import (
	"math"

	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"

	"github.com/crawshaw/balloon/animation"
)

var baseAnimation = &animation.Animation{}

const maxExtend = geom.Pt(72) // TODO

type scissorArm2 struct {
	a        *animation.Animation
	extend   geom.Pt
	numFolds int

	node        *sprite.Node
	arrangement animation.Arrangement
}

func (s *scissorArm2) balloonTravel() (minX, maxX geom.Pt) {
	minX = geom.Pt(s.numFolds)*10 - 5 + 14
	maxX = minX + geom.Pt(s.numFolds)*18
	maxX *= (s.extend / maxExtend)
	return minX, maxX
}

func (s *scissorArm2) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	s.a.Arrange(e, n, t)
	s.arrangement.Arrange(e, n, t)
}

func (s *scissorArm2) moveArm(ar *animation.Arrangement, tween float32) {
	ar.Offset.X += 18 * geom.Pt(tween) * (s.extend / maxExtend)
}

func (s *scissorArm2) moveArmBack(ar *animation.Arrangement, tween float32) {
	ar.Offset.X -= 18 * geom.Pt(tween) * (s.extend / maxExtend)
}

func (s *scissorArm2) rotateArm(ar *animation.Arrangement, tween float32) {
	ar.Rotation += tween * 0.8 * float32(s.extend/maxExtend)
}

func (s *scissorArm2) rotateArmBack(ar *animation.Arrangement, tween float32) {
	ar.Rotation -= tween * 0.8 * float32(s.extend/maxExtend)
}

func newScissorArm2(eng sprite.Engine) *scissorArm2 {
	s := &scissorArm2{
		extend:   maxExtend,
		numFolds: 3,
		node:     new(sprite.Node),
	}
	s.node.Arranger = s
	eng.Register(s.node)

	base := new(sprite.Node)
	eng.Register(base)
	base.Arranger = &animation.Arrangement{
		Pivot:    geom.Point{X: 6, Y: 18},
		Size:     &geom.Point{X: 12, Y: 36},
		Rotation: math.Pi,
		SubTex:   sheet.pad, // TODO: get a better texture
	}
	s.node.AppendChild(base)

	moveArm := animation.TransformerFunc(s.moveArm)
	moveArmBack := animation.TransformerFunc(s.moveArmBack)
	rotateArm := animation.TransformerFunc(s.rotateArm)
	rotateArmBack := animation.TransformerFunc(s.rotateArmBack)

	expanding := make(map[*sprite.Node]animation.Transform)
	contracting := make(map[*sprite.Node]animation.Transform)

	// TODO: have i inverted top and bottom naming here?

	parent := s.node
	for i := 0; i < s.numFolds; i++ {
		offX := geom.Pt(10)
		if i == 0 {
			offX = 5
		}
		b := new(sprite.Node)
		b.Arranger = &animation.Arrangement{Offset: geom.Point{X: offX}}
		eng.Register(b)
		parent.AppendChild(b)
		parent = b

		expanding[b] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: moveArm,
		}
		contracting[b] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: moveArmBack,
		}

		arm := new(sprite.Node)
		eng.Register(arm)
		a := &animation.Arrangement{
			Pivot:    geom.Point{X: 18, Y: 4},
			Size:     &geom.Point{X: 36, Y: 9},
			Rotation: 1.2,
			SubTex:   sheet.arm,
		}
		arm.Arranger = a
		b.AppendChild(arm)

		expanding[arm] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: rotateArmBack,
		}
		contracting[arm] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: rotateArm,
		}
	}

	parent = s.node
	for i := 0; i < s.numFolds; i++ {
		offX := geom.Pt(10)
		if i == 0 {
			offX = 5
		}
		t := new(sprite.Node)
		t.Arranger = &animation.Arrangement{Offset: geom.Point{X: offX}}
		eng.Register(t)
		parent.AppendChild(t)
		parent = t

		expanding[t] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: moveArm,
		}
		contracting[t] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: moveArmBack,
		}

		arm := new(sprite.Node)
		eng.Register(arm)
		a := &animation.Arrangement{
			Pivot:    geom.Point{X: 18, Y: 4},
			Size:     &geom.Point{X: 36, Y: 9},
			Rotation: -1.2,
			SubTex:   sheet.arm,
		}
		arm.Arranger = a
		t.AppendChild(arm)

		expanding[arm] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: rotateArm,
		}
		contracting[arm] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: rotateArmBack,
		}
	}

	// Pad.
	p := new(sprite.Node)
	eng.Register(p)
	p.Arranger = &animation.Arrangement{
		Offset:   geom.Point{X: 12},
		Pivot:    geom.Point{X: 6, Y: 18},
		Size:     &geom.Point{X: 12, Y: 36},
		Rotation: math.Pi,
		SubTex:   sheet.pad,
	}
	parent.AppendChild(p)

	// Balloon.
	p = new(sprite.Node)
	eng.Register(p)
	p.Arranger = &animation.Arrangement{
		Offset: geom.Point{X: 14, Y: -72 + 36/2},
		Pivot:  geom.Point{X: 6, Y: 18},
		Size:   &geom.Point{X: 24, Y: 72},
		SubTex: sheet.balloon,
	}
	parent.AppendChild(p)

	// New top-level node for moving around entire assembly.
	p = &sprite.Node{
		Arranger: &animation.Arrangement{},
	}
	eng.Register(p)
	p.AppendChild(s.node)
	s.node = p

	size := geom.Pt(s.numFolds*10 + 12)

	s.a = &animation.Animation{
		Current: "init",
		States: map[string]animation.State{
			"init": animation.State{},
			"offscreen": animation.State{
				Next: "onscreen",
				Transforms: map[*sprite.Node]animation.Transform{
					s.node: animation.Transform{
						Tween:       clock.EaseInOut,
						Transformer: animation.Move{X: -size},
					},
				},
			},
			"onscreen": animation.State{
				Duration: 60,
				Next:     "closed",
				Transforms: map[*sprite.Node]animation.Transform{
					s.node: animation.Transform{
						Tween:       clock.EaseInOut,
						Transformer: animation.Move{X: size},
					},
				},
			},
			"closed": animation.State{
				Duration: 5,
				Next:     "loading_balloon",
			},
			"loading_balloon": animation.State{
				Next: "ready",
			},
			"ready": animation.State{},
			"expanding": animation.State{
				Duration:   expandTime,
				Next:       "open",
				Transforms: expanding,
			},
			"open": animation.State{
				Duration: 5,
				Next:     "contracting",
			},
			"contracting": animation.State{
				Duration:   5,
				Next:       "closed",
				Transforms: contracting,
			},
		},
	}
	s.a.Transition(0, "offscreen")

	return s
}
