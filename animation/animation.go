package animation

import (
	"fmt"
	"log"

	"code.google.com/p/go.mobile/f32"
	"code.google.com/p/go.mobile/geom"
	"code.google.com/p/go.mobile/sprite"
	"code.google.com/p/go.mobile/sprite/clock"
)

type NodeName string

type Path []int

type State struct {
	Duration   int
	Next       string
	Transforms map[NodeName]Transform
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
	State  string
	States map[string]State
	Nodes  map[NodeName]Path

	root           *sprite.Node
	nodes          map[NodeName]*sprite.Node
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
}

func (a *Animation) Transition(state string) {
	log.Printf("TODO Transition")
}

func (a *Animation) init(root *sprite.Node) error {
	a.root = root
	a.nodes = make(map[NodeName]*sprite.Node)
	for nodeName, path := range a.Nodes {
		n, err := resolve(a.root, path)
		if err != nil {
			return fmt.Errorf("animation.Animation: node %q has invalid path %v, search stopped at %v", nodeName, path, err)
		}
		a.nodes[nodeName] = n
	}
	for stateName, s := range a.States {
		if s.Next != "" {
			if _, exists := a.States[s.Next]; !exists {
				return fmt.Errorf("animation.Animation: state %q transitions to non-existent state %q", stateName, s.Next)
			}
		}
	}
	return nil
}

func resolve(n *sprite.Node, path Path) (*sprite.Node, error) {
	if len(path) == 0 {
		return n, nil
	}
	c := n.FirstChild
	rem := path[0]
	for rem > 0 && c != nil {
		c = c.NextSibling
		rem--
	}
	if c == nil {
		return nil, fmt.Errorf("[%d]", rem)
	}

	c, err := resolve(c, path[1:])
	if err != nil {
		return nil, fmt.Errorf("[%d]%s", path[0], err)
	}
	return c, nil
}

// Arrangement is a sprite Arranger that uses high-level concepts to
// transform a sprite Node.
type Arrangement struct {
	Offset   geom.Point     // distance between parent and pivot
	Pivot    geom.Point     // point on sized, unrotated node
	Size     *geom.Point    // optional bounding rectangle for scaling
	Rotation float32        // radians counter-clockwise
	Texture  sprite.Texture // optional Node Texture

	Transform struct {
		Transform
		T0, T1 clock.Time
	}

	//Transforms []Transform    // active transformations to apply on Arrange

	// TODO: Physics *physics.Physics
}

func (ar *Arrangement) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	ar2 := *ar

	if ar.Transform.Transformer != nil {
		tween := ar.Transform.Tween(ar.Transform.T0, ar.Transform.T1, t)
		ar.Transform.Transformer.Transform(&ar2, tween)
	}
	/*
		for _, a := range ar.Transforms {
			tween := a.Tween(a.T0, a.T1, t)
			a.Transformer.Transform(&ar2, tween)
		}
	*/
	e.SetTexture(n, t, ar2.Texture)
	e.SetTransform(n, t, ar2.Affine())

	//ar.squash(t)
}

// Squash plays through transformas and physics, updating the Arrangement
// and removing any outdated transforms.
//
// TODO: automatically do this? export? if automatic, merge into Arrange.
/*
func (ar *Arrangement) Squash(t clock.Time) {
	remove := 0
	for _, a := range ar.Transforms {
		if t < a.T1 {
			// stop squashing at the first animation that cannot be squashed.
			// animations are not commutative.
			break
		}
		a.Transformer.Transform(ar, 1)
		fmt.Printf("squash: %+v\n", ar)
		remove++
	}
	ar.Transforms = ar.Transforms[remove:]
}
*/

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
	//T0, T1      clock.Time
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

// Move moves the Arrangement offset.
type Move geom.Point

func (m Move) Transform(ar *Arrangement, tween float32) {
	ar.Offset.X += m.X * geom.Pt(tween)
	ar.Offset.Y += m.Y * geom.Pt(tween)
}

func (m Move) String() string {
	return fmt.Sprintf("Move(%s,%s)", m.X, m.Y)
}
