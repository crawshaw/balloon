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
	a      *animation.Animation
	extend geom.Pt
}

func (s *scissorArm2) moveTop(ar *animation.Arrangement, tween float32) {
	ar.Offset.X += 16 * geom.Pt(tween) // TODO extend ratio
}

func (s *scissorArm2) rotateTop(ar *animation.Arrangement, tween float32) {
	ar.Rotation += tween * 0.7 * float32(s.extend/maxExtend)
}

func newScissorArm2(eng sprite.Engine) *scissorArm2 {
	//new(sprite.Node)
	s := &scissorArm2{}
	s.a = &animation.Animation{
		State: "closed",
		States: map[string]animation.State{
			"closed": animation.State{},
			"expanding": animation.State{
				Duration: 15,
				Next:     "open",
				Transforms: map[animation.NodeName]animation.Transform{
					"top0": animation.Transform{
						Tween:       clock.Linear,
						Transformer: animation.TransformerFunc(s.moveTop),
					},
				},
			},
			"open": animation.State{
				Transforms: map[animation.NodeName]animation.Transform{
					"top0": animation.Transform{},
				},
			},
			"contracting": animation.State{
				Duration: 15,
				Next:     "closed",
			},
		},
		Nodes: map[animation.NodeName]animation.Path{},
	}

	return s
}

func (s *scissorArm2) touch(t event.Touch) {
}
