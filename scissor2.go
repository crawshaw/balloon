package main

import (
	"code.google.com/p/go.mobile/event"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"

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

func (s *scissorArm2) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	s.arrangement.Arrange(e, n, t)
}

func (s *scissorArm2) moveTop(ar *animation.Arrangement, tween float32) {
	ar.Offset.X += 16 * geom.Pt(tween) // TODO extend ratio
}

func (s *scissorArm2) moveBottom(ar *animation.Arrangement, tween float32) {
	s.moveTop(ar, tween)
}

func (s *scissorArm2) rotateTop(ar *animation.Arrangement, tween float32) {
	ar.Rotation += tween * 0.7 * float32(s.extend/maxExtend)
}

func (s *scissorArm2) rotateBottom(ar *animation.Arrangement, tween float32) {
	ar.Rotation -= tween * 0.7 * float32(s.extend/maxExtend)
}

func newScissorArm2(eng sprite.Engine) *scissorArm2 {
	s := &scissorArm2{
		extend:   maxExtend,
		numFolds: 3,
		node:     new(sprite.Node),
	}
	s.node.Arranger = s
	eng.Register(s.node)

	moveTop := animation.TransformerFunc(s.moveTop)
	rotateTop := animation.TransformerFunc(s.rotateTop)
	/*
		moveBottom := animation.TransformerFunc(s.moveBottom)
		rotateBottom := animation.TransformerFunc(s.rotateBottom)
	*/

	expanding := animation.State{
		Duration:   15,
		Next:       "open",
		Transforms: make(map[*sprite.Node]animation.Transform),
		/*
				top[0]: animation.Transform{
					Tween:       clock.EaseIn,
					Transformer: moveTop,
				},
				top[1]: animation.Transform{
					Tween:       clock.EaseIn,
					Transformer: moveTop,
				},
				top[2]: animation.Transform{
					Tween:       clock.EaseIn,
					Transformer: moveTop,
				},
				topArm[0]: animation.Transform{
					Tween:       clock.EaseIn,
					Transformer: rotateTop,
				},
			},
		*/
	}

	// TODO: have i inverted top and bottom naming here?

	parent := s.node
	for i := 0; i < s.numFolds; i++ {
		b := new(sprite.Node)
		b.Arranger = &animation.Arrangement{
			Offset: geom.Point{X: 10},
		}
		eng.Register(b)
		parent.AppendChild(b)
		parent = b

		expanding.Transforms[b] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: moveTop, // TODO: i == 0 moveFirstTop
		}

		arm := new(sprite.Node)
		eng.Register(arm)
		a := &animation.Arrangement{
			Pivot:    geom.Point{X: 18, Y: 4},
			Size:     &geom.Point{X: 36, Y: 9},
			Rotation: 1.2,
			Texture:  sheet.arm,
		}
		arm.Arranger = a
		b.AppendChild(arm)

		expanding.Transforms[arm] = animation.Transform{
			Tween:       clock.EaseIn,
			Transformer: rotateTop,
		}
	}

	parent = s.node
	for i := 0; i < s.numFolds; i++ {
		t := new(sprite.Node)
		t.Arranger = &animation.Arrangement{
			Offset: geom.Point{X: 10},
		}
		eng.Register(t)
		parent.AppendChild(t)
		parent = t

		arm := new(sprite.Node)
		eng.Register(arm)
		a := &animation.Arrangement{
			Pivot:    geom.Point{X: 18, Y: 4},
			Size:     &geom.Point{X: 36, Y: 9},
			Rotation: -1.2,
			Texture:  sheet.arm,
		}
		arm.Arranger = a
		t.AppendChild(arm)
	}

	s.a = &animation.Animation{
		Current: "closed",
		States: map[string]animation.State{
			"closed":    animation.State{},
			"expanding": expanding,
			"open":      animation.State{
			/*
				Transforms: map[*sprite.Node]animation.Transform{
					top[0]: animation.Transform{},
				},
			*/
			},
			"contracting": animation.State{
				Duration: 15,
				Next:     "closed",
			},
		},
	}

	return s
}

func (s *scissorArm2) touch(t clock.Time, e event.Touch) {
	if s.a.Current == "closed" {
		s.a.Transition(t, "expanding")
	}
}
