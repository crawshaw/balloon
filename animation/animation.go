package animation

import (
	"fmt"
	"log"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"
)

type State struct {
	Duration   int
	Next       string
	Transforms map[*sprite.Node]Transform
}

// Animation is a state machine for a node tree.
//
// It is implemented as an Arranger on the root node of the tree
// it controls. Using named paths, it assigns transforms and manipulates
// children of the root node.
//
// States can transition automatically after some predefined duration, or
// by calling the Transition method.
type Animation struct {
	Current string
	States  map[string]State

	root           *sprite.Node
	lastTransition clock.Time
}

func (a *Animation) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	if a.root == nil {
		a.init(n)
	}
	if a.root != n {
		// TODO: return error
		log.Printf("animation.Animation: root node changed (%p, %p)", a.root, n)
		return
	}
	s := a.States[a.Current]
	if s.Next != "" && clock.Time(s.Duration) < t-a.lastTransition {
		a.Transition(t, s.Next)
	}
}

func (a *Animation) Transition(t clock.Time, name string) {
	log.Printf("animation: Transition from %q to %q", a.Current, name)
	for n := range a.States[a.Current].Transforms {
		// Squash the final animation state down onto the node.
		ar := n.Arranger.(*Arrangement)
		ar.Transform.Transformer.Transform(ar, 1)
		n.Arranger.(*Arrangement).Transform.Transformer = nil
	}
	s := a.States[name]
	for n, transform := range s.Transforms {
		ar := n.Arranger.(*Arrangement)
		ar.T0 = t
		ar.T1 = t + clock.Time(s.Duration)
		ar.Transform = transform
	}
	a.Current = name
	a.lastTransition = t
}

func (a *Animation) init(root *sprite.Node) error {
	a.root = root
	for stateName, s := range a.States {
		if s.Next != "" {
			if _, exists := a.States[s.Next]; !exists {
				return fmt.Errorf("animation.Animation: state %q transitions to non-existent state %q", stateName, s.Next)
			}
		}
	}
	return nil
}

// Arrangement is a sprite Arranger that uses high-level concepts to
// transform a sprite Node.
type Arrangement struct {
	Offset   geom.Point     // distance between parent and pivot
	Pivot    geom.Point     // point on sized, unrotated node
	Size     *geom.Point    // optional bounding rectangle for scaling
	Rotation float32        // radians counter-clockwise
	Texture  sprite.Texture // optional Node Texture

	T0, T1    clock.Time
	Transform Transform

	// TODO: Physics *physics.Physics
}

func (ar *Arrangement) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	ar2 := *ar

	if ar.Transform.Transformer != nil {
		fn := ar.Transform.Tween
		if fn == nil {
			fn = clock.Linear
		}
		tween := fn(ar.T0, ar.T1, t)
		ar.Transform.Transformer.Transform(&ar2, tween)
	}
	e.SetTexture(n, t, ar2.Texture)
	e.SetTransform(n, t, ar2.Affine())
}

func (ar *Arrangement) Affine() f32.Affine {
	var a f32.Affine
	a.Identity()
	if ar == nil {
		return a
	}
	a.Translate(&a, float32(ar.Offset.X), float32(ar.Offset.Y))
	if ar.Rotation != 0 {
		a.Rotate(&a, ar.Rotation)
	}
	x, y := float32(ar.Pivot.X), float32(ar.Pivot.Y)
	a.Translate(&a, -x, -y)

	if ar.Size != nil {
		a.Scale(&a, float32(ar.Size.X), float32(ar.Size.Y))
	}
	return a
}

type Transform struct {
	Tween       func(t0, t1, t clock.Time) float32
	Transformer Transformer
}

type Transformer interface {
	Transform(ar *Arrangement, tween float32)
}

type TransformerFunc func(ar *Arrangement, tween float32)

func (t TransformerFunc) Transform(ar *Arrangement, tween float32) {
	t(ar, tween)
}

// Rotate rotates counter-clockwise, measured in radians.
type Rotate float32

func (r Rotate) Transform(ar *Arrangement, tween float32) {
	ar.Rotation += tween * float32(r)
}

func (r Rotate) String() string { return fmt.Sprintf("Rotate(%d)", r) }

// Move moves the Arrangement offset.
type Move geom.Point

func (m Move) Transform(ar *Arrangement, tween float32) {
	ar.Offset.X += m.X * geom.Pt(tween)
	ar.Offset.Y += m.Y * geom.Pt(tween)
}

func (m Move) String() string { return fmt.Sprintf("Move(%s,%s)", m.X, m.Y) }
