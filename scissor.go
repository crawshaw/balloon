package main

import (
	"fmt"

	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"

	"github.com/crawshaw/balloon/animation"
)

type scissorState int

const (
	scissorClosed scissorState = iota
	scissorExpanding
	scissorContracting
)

type scissorArm struct {
	arrangement animation.Arrangement
	node        *sprite.Node
	numFolds    int
	state       scissorState
}

func newScissorArm(eng sprite.Engine) *scissorArm {
	s := &scissorArm{
		node:     new(sprite.Node),
		numFolds: 3,
	}
	s.node.Arranger = s
	eng.Register(s.node)

	parent := s.node
	for i := 0; i < s.numFolds; i++ {
		bottom := new(sprite.Node)
		bottom.Arranger = &animation.Arrangement{
			Offset: geom.Point{X: 10},
		}
		eng.Register(bottom)
		parent.AppendChild(bottom)
		parent = bottom

		bottomArm := new(sprite.Node)
		eng.Register(bottomArm)
		a := &animation.Arrangement{
			Pivot:    geom.Point{X: 18, Y: 4},
			Size:     &geom.Point{X: 36, Y: 9},
			Rotation: 1.2,
			Texture:  sheet.arm,
		}
		bottomArm.Arranger = a
		bottom.AppendChild(bottomArm)
	}

	parent = s.node
	for i := 0; i < s.numFolds; i++ {
		top := new(sprite.Node)
		top.Arranger = &animation.Arrangement{
			Offset: geom.Point{X: 10},
		}
		eng.Register(top)
		parent.AppendChild(top)
		parent = top

		topArm := new(sprite.Node)
		eng.Register(topArm)
		a := &animation.Arrangement{
			Pivot:    geom.Point{X: 18, Y: 4},
			Size:     &geom.Point{X: 36, Y: 9},
			Rotation: -1.2,
			Texture:  sheet.arm,
		}
		topArm.Arranger = a
		top.AppendChild(topArm)
	}

	return s
}

func (s *scissorArm) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	s.arrangement.Arrange(e, n, t)
}

func (s *scissorArm) expand(t clock.Time) {
	if s.state != scissorClosed {
		return
	}
	s.state = scissorExpanding

	top0, bottom0 := s.node.LastChild, s.node.FirstChild
	fmt.Printf("top0: %+v\nbottom0: %+v\n", top0, bottom0)
	appendTransform(top0, t, animation.Move{8, 0})
	appendTransform(bottom0, t, animation.Move{8, 0})

	n := top0.LastChild
	for i := 0; i < s.numFolds-1; i++ {
		appendTransform(n, t, animation.Move{16, 0})
		n = n.LastChild
	}
	n = top0
	for i := 0; i < s.numFolds; i++ {
		appendTransform(n.FirstChild, t, animation.Rotate(0.7))
		n = n.LastChild
	}

	n = bottom0.LastChild
	for i := 0; i < s.numFolds-1; i++ {
		appendTransform(n, t, animation.Move{16, 0})
		n = n.LastChild
	}
	n = bottom0
	for i := 0; i < s.numFolds; i++ {
		appendTransform(n.FirstChild, t, animation.Rotate(-0.7))
		n = n.LastChild
	}
}

func appendTransform(n *sprite.Node, t clock.Time, transformer animation.Transformer) {
	ar := n.Arranger.(*animation.Arrangement)
	ar.Transform.T0 = t
	ar.Transform.T1 = t + 15
	ar.Transform.Transform = animation.Transform{
		Tween:       clock.Linear,
		Transformer: transformer,
	}
}
