// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package text implements a sprite.Arranger that can lay out text.
//
// Glyphs are rendered into a shared, reused cache controlled by the Engine
// implementation.
package text

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"

	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"golang.org/x/mobile/f32"
	"golang.org/x/mobile/geom"
	"golang.org/x/mobile/sprite"
	"golang.org/x/mobile/sprite/clock"
)

const colWidth = 128

var cache = make(map[sprite.Engine]*glyphCache)

// a sheet is made up of 128px columns, filled top-to-bottom, left-to-right.
// TODO be cleverer: https://medium.com/@romainguy/androids-font-renderer-c368bbde87d9
type sheet struct {
	s    sprite.Texture
	x, y int // next empty slot
}

type cacheEntry struct {
	glyph        glyphKey
	texture      sprite.SubTex
	offset       image.Point
	advanceWidth float32    // pixels
	time         clock.Time // needed for rendering at time

	next, prev *cacheEntry // linked-list, most recently used at front
}

type glyphKey struct {
	index truetype.Index
	size  geom.Pt
	font  *truetype.Font
}

type glyphCache struct {
	e sprite.Engine
	s *sheet // TODO more sheets
	r *raster.Rasterizer

	glyphBuf   *truetype.GlyphBuf
	cache      map[glyphKey]*cacheEntry
	cacheFront *cacheEntry
	scratch    *image.RGBA // TODO: *image.Alpha
}

func (c *glyphCache) get(glyph glyphKey, t clock.Time) (*cacheEntry, error) {
	entry := c.cache[glyph]
	if entry == nil {
		entry = &cacheEntry{glyph: glyph}
		if err := c.rasterize(entry, t); err != nil {
			return nil, err
		}
		c.cache[glyph] = entry
	} else {
		entry.time = t

		// remove from list
		if entry.prev != nil {
			entry.prev.next = entry.next
		}
		if entry.next != nil {
			entry.next.prev = entry.prev
		}
	}

	// put on front of list
	entry.prev = nil
	entry.next = c.cacheFront
	if c.cacheFront != nil {
		c.cacheFront.prev = entry
		c.cacheFront = entry
	}
	return entry, nil
}

func (c *glyphCache) findSpace(w, h int, t clock.Time) (image.Point, error) {
	if w > colWidth {
		return image.Point{}, fmt.Errorf("text: glyph larger than cache column width: %d", w)
	}
	s := c.s
	sw, sh := s.s.Bounds()
	if h > sh-s.y {
		s.x += colWidth
		s.y = 0
	}
	if s.x >= sw {
		// out of space, clear out old glyphs
		if err := c.clearHalf(t); err != nil {
			return image.Point{}, err
		}
		if h > sh-s.y {
			s.x += colWidth
			s.y = 0
		}
	}
	if w > sw-s.x || h > sh-s.y {
		return image.Point{}, fmt.Errorf("text: no space for glyph w=%d, h=%d", w, h)
	}
	p := image.Point{s.x, s.y}
	s.y += h
	return p, nil
}

func (c *glyphCache) clearHalf(t clock.Time) error {
	e := c.cacheFront
	for e.next != nil {
		e = e.next
	}

	toDelete := len(c.cache) / 2
	deleted := 0
	for e != nil && toDelete > 0 {
		if e.time < t {
			delete(c.cache, e.glyph)
			if e.next != nil {
				e.next.prev = e.prev
			}
			if e.prev != nil {
				e.prev.next = e.next
			}
			deleted++
			toDelete--
		}
		e = e.prev
	}
	if deleted == 0 {
		return fmt.Errorf("text: glyph cache is full (%d items)", len(c.cache))
	}

	// re-render cache
	for e := c.cacheFront; e != nil; e = e.next {
		if err := c.rasterize(e, e.time); err != nil {
			return err
		}
	}
	return nil
}

func (c *glyphCache) rasterize(entry *cacheEntry, t clock.Time) error {
	// Hinting is disabled. We can't pixel snap without knowing where the
	// pixels are. As a bonus, we get a more space efficient glyph cache.
	err := c.glyphBuf.Load(
		entry.glyph.font,
		floatToFix(entry.glyph.size.Px()),
		entry.glyph.index,
		truetype.NoHinting,
	)
	if err != nil {
		return err
	}
	// Calculate the integer-pixel bounds for the glyph.
	xmin := int(+raster.Fix32(c.glyphBuf.B.XMin<<2)) >> 8
	ymin := int(-raster.Fix32(c.glyphBuf.B.YMax<<2)) >> 8
	xmax := int(+raster.Fix32(c.glyphBuf.B.XMax<<2)+0xff) >> 8
	ymax := int(-raster.Fix32(c.glyphBuf.B.YMin<<2)+0xff) >> 8
	if xmin > xmax || ymin > ymax {
		return errors.New("text: negative sized glyph")
	}
	w, h := xmax-xmin, ymax-ymin
	entry.offset = image.Point{xmin, ymin}
	p, err := c.findSpace(w, h, t)
	if err != nil {
		return err
	}
	entry.advanceWidth = fixToFloat(c.glyphBuf.AdvanceWidth)

	// A TrueType's glyph's nodes can have negative co-ordinates, but the
	// rasterizer clips anything left of x=0 or above y=0. xmin and ymin
	// are the pixel offsets, based on the font's FUnit metrics, that let
	// a negative co-ordinate in TrueType space be non-negative in
	// rasterizer space. xmin and ymin are typically <= 0.
	fx := raster.Fix32(-xmin << 8)
	fy := raster.Fix32(-ymin << 8)
	c.r.Clear()
	e0 := 0
	for _, e1 := range c.glyphBuf.End {
		drawContour(c.r, c.glyphBuf.Point[e0:e1], fx, fy)
		e0 = e1
	}
	//a := c.scratch.SubImage(image.Rect(0, 0, w, h)).(*image.Alpha)
	a := c.scratch.SubImage(image.Rect(0, 0, w, h)).(*image.RGBA)
	for i := range a.Pix {
		a.Pix[i] = 0
	}
	//c.r.Rasterize(raster.NewAlphaSrcPainter(a))
	painter := raster.NewRGBAPainter(a)
	painter.SetColor(color.Black)
	c.r.Rasterize(painter)
	entry.texture = sprite.SubTex{
		T: c.s.s,
		R: image.Rect(p.X, p.Y, p.X+w, p.Y+h),
	}
	c.s.s.Upload(entry.texture.R, a)
	return nil
}

func getCache(e sprite.Engine) (*glyphCache, error) {
	c := cache[e]
	if c != nil {
		return c, nil
	}
	const w, h = 1024, 512
	//x, err := e.LoadTexture(image.NewAlpha(image.Rect(0, 0, w, h)))
	x, err := e.LoadTexture(image.NewRGBA(image.Rect(0, 0, w, h)))
	if err != nil {
		return nil, err
	}
	c = &glyphCache{
		e:        e,
		r:        raster.NewRasterizer(0, 0),
		s:        &sheet{s: x},
		glyphBuf: truetype.NewGlyphBuf(),
		scratch:  image.NewRGBA(image.Rect(0, 0, colWidth, h)),
		cache:    make(map[glyphKey]*cacheEntry),
	}
	cache[e] = c
	return c, nil
}

// fixToFloat converts fixed width 26.6 numbers to float32.
func fixToFloat(x int32) float32 {
	return float32(x>>6) + float32(x&0x3f)/0x3f
}

// floatToFix converts float32 to fixed width 26.6 numbers.
func floatToFix(x float32) int32 {
	return int32(x * 64.0)
}

func drawContour(r *raster.Rasterizer, ps []truetype.Point, dx, dy raster.Fix32) {
	if len(ps) == 0 {
		return
	}
	// ps[0] is a truetype.Point measured in FUnits and positive Y going upwards.
	// start is the same thing measured in fixed point units and positive Y
	// going downwards, and offset by (dx, dy)
	start := raster.Point{
		X: dx + raster.Fix32(ps[0].X<<2),
		Y: dy - raster.Fix32(ps[0].Y<<2),
	}
	r.Start(start)
	q0, on0 := start, true
	for _, p := range ps[1:] {
		q := raster.Point{
			X: dx + raster.Fix32(p.X<<2),
			Y: dy - raster.Fix32(p.Y<<2),
		}
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				r.Add1(q)
			} else {
				r.Add2(q0, q)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				mid := raster.Point{
					X: (q0.X + q.X) / 2,
					Y: (q0.Y + q.Y) / 2,
				}
				r.Add2(q0, mid)
			}
		}
		q0, on0 = q, on
	}
	// Close the curve.
	if on0 {
		r.Add1(start)
	} else {
		r.Add2(q0, start)
	}
}

// String is a sprite.Arranger that draws a string.
//
// This arranger owns all child nodes, and rearranges them at will.
type String struct {
	Text  string
	Size  geom.Pt
	Color color.Color
	Font  *truetype.Font
}

func (s *String) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	if s.Font == nil {
		return
	}

	c, err := getCache(e)
	if err != nil {
		log.Fatal(err) // TODO return error
	}
	scale := floatToFix(s.Size.Px())
	b := s.Font.Bounds(scale)
	xmin := +int(b.XMin) >> 6
	ymin := -int(b.YMax) >> 6
	xmax := +int(b.XMax+63) >> 6
	ymax := -int(b.YMin-63) >> 6
	c.r.SetBounds(xmax-xmin, ymax-ymin)

	// TODO reuse nodes
	n.FirstChild = nil
	n.LastChild = nil
	prev, hasPrev := truetype.Index(0), false
	var x float32 // pixels
	for _, rune := range s.Text {
		index := s.Font.Index(rune)
		if hasPrev {
			x += fixToFloat(s.Font.Kerning(scale, prev, index))
		}
		entry, err := c.get(glyphKey{
			index: index,
			size:  s.Size,
			font:  s.Font,
		}, t)
		if err != nil {
			log.Fatal(err) // TODO return error
		}

		// Create a node to represent the glyph.
		glyphNode := new(sprite.Node)
		e.Register(glyphNode)
		n.AppendChild(glyphNode)
		var a f32.Affine
		a.Identity()
		a.Translate(
			&a,
			(x+float32(entry.offset.X))/geom.PixelsPerPt,
			float32(entry.offset.Y)/geom.PixelsPerPt,
		)
		w, h := entry.texture.R.Dx(), entry.texture.R.Dy()
		a.Scale(&a, float32(w)/geom.PixelsPerPt, float32(h)/geom.PixelsPerPt)
		e.SetTransform(glyphNode, a)
		subTex := entry.texture // copy
		/*subTex.T = &sprite.AlphaTexutre{
			T: subTex.T,
			C: s.Color,
		}*/
		e.SetSubTex(glyphNode, subTex)

		x += entry.advanceWidth
		prev, hasPrev = index, true
	}
}
