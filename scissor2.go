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
	s.a.Arrange(e, n, t)
	s.arrangement.Arrange(e, n, t)
}

func (s *scissorArm2) moveArm(ar *animation.Arrangement, tween float32) {
	ar.Offset.X += 16 * geom.Pt(tween) * (s.extend / maxExtend)
}

func (s *scissorArm2) moveArmBack(ar *animation.Arrangement, tween float32) {
	ar.Offset.X -= 16 * geom.Pt(tween) * (s.extend / maxExtend)
}

func (s *scissorArm2) rotateArm(ar *animation.Arrangement, tween float32) {
	ar.Rotation += tween * 0.7 * float32(s.extend/maxExtend)
}

func (s *scissorArm2) rotateArmBack(ar *animation.Arrangement, tween float32) {
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

	moveArm := animation.TransformerFunc(s.moveArm)
	moveArmBack := animation.TransformerFunc(s.moveArmBack)
	rotateArm := animation.TransformerFunc(s.rotateArm)
	rotateArmBack := animation.TransformerFunc(s.rotateArmBack)

	expanding := make(map[*sprite.Node]animation.Transform)
	contracting := make(map[*sprite.Node]animation.Transform)

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
			Texture:  sheet.arm,
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
		t := new(sprite.Node)
		t.Arranger = &animation.Arrangement{
			Offset: geom.Point{X: 10},
		}
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
			Texture:  sheet.arm,
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

	s.a = &animation.Animation{
		Current: "closed",
		States: map[string]animation.State{
			"closed": animation.State{},
			"expanding": animation.State{
				Duration:   10,
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

	return s
}

func (s *scissorArm2) touch(t clock.Time, e event.Touch) {
	if s.a.Current == "closed" {
		s.a.Transition(t, "expanding")
	}
}
