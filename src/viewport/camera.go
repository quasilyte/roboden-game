package viewport

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type World struct {
	Width  float64
	Height float64
}

type cameraObject interface {
	ge.SceneGraphics
	BoundsRect() gmath.Rect
}

type Camera struct {
	World *World

	Offset gmath.Vec

	Rect       gmath.Rect
	globalRect gmath.Rect

	bg                   *ge.TiledBackground
	fogOfWar             *ebiten.Image
	belowObjects         layer
	objects              layer
	slightlyAboveObjects layer
	aboveObjects         layer

	screen *ebiten.Image

	disposed bool
}

func NewCamera(w *World, width, height float64) *Camera {
	return &Camera{
		World: w,
		Rect: gmath.Rect{
			Min: gmath.Vec{},
			Max: gmath.Vec{X: width, Y: height},
		},
		screen: ebiten.NewImage(int(w.Width), int(w.Height)),
	}
}

func (c *Camera) Dispose() {
	c.disposed = true
}

func (c *Camera) IsDisposed() bool {
	return c.disposed
}

func (c *Camera) AddSprite(s *ge.Sprite) {
	c.objects.AddSprite(s)
}

func (c *Camera) AddGraphics(o cameraObject) {
	c.objects.Add(o)
}

func (c *Camera) AddSpriteSlightlyAbove(s *ge.Sprite) {
	c.slightlyAboveObjects.AddSprite(s)
}

func (c *Camera) AddSpriteAbove(s *ge.Sprite) {
	c.aboveObjects.AddSprite(s)
}

func (c *Camera) AddGraphicsSlightlyAbove(o cameraObject) {
	c.slightlyAboveObjects.Add(o)
}

func (c *Camera) AddGraphicsAbove(o cameraObject) {
	c.aboveObjects.Add(o)
}

func (c *Camera) AddSpriteBelow(s *ge.Sprite) {
	c.belowObjects.AddSprite(s)
}

func (c *Camera) SortBelowLayer() {
	if len(c.belowObjects.objects) != 0 {
		panic("unexpected below objects count")
	}
	sort.Slice(c.belowObjects.sprites, func(i, j int) bool {
		shape1 := c.belowObjects.sprites[i].BoundsRect()
		shape2 := c.belowObjects.sprites[j].BoundsRect()
		return shape1.Max.Y < shape2.Max.Y
	})
}

func (c *Camera) SetFogOfWar(img *ebiten.Image) {
	c.fogOfWar = img
}

func (c *Camera) SetBackground(bg *ge.TiledBackground) {
	c.bg = bg
}

func (c *Camera) drawLayer(screen *ebiten.Image, l *layer) {
	liveSprites := l.sprites[:0]
	for _, s := range l.sprites {
		if s.IsDisposed() {
			continue
		}
		if c.isVisible(s.BoundsRect()) {
			s.Draw(screen)
		}
		liveSprites = append(liveSprites, s)
	}
	l.sprites = liveSprites

	if len(l.objects) != 0 {
		liveObjects := l.objects[:0]
		for _, o := range l.objects {
			if o.IsDisposed() {
				continue
			}
			if c.isVisible(o.BoundsRect()) {
				o.Draw(screen)
			}
			liveObjects = append(liveObjects, o)
		}
		l.objects = liveObjects
	}
}

func (c *Camera) isVisible(objectRect gmath.Rect) bool {
	cameraRect := c.globalRect

	if objectRect.Max.X < cameraRect.Min.X {
		return false
	}
	if objectRect.Min.X > cameraRect.Max.X {
		return false
	}
	if objectRect.Max.Y < cameraRect.Min.Y {
		return false
	}
	if objectRect.Min.Y > cameraRect.Max.Y {
		return false
	}

	return true
}

func (c *Camera) ContainsPos(pos gmath.Vec) bool {
	globalRect := c.Rect
	globalRect.Min = c.Offset
	globalRect.Max = globalRect.Max.Add(c.Offset)
	return globalRect.Contains(pos)
}

func (c *Camera) checkBounds() {
	c.Offset.X = gmath.Clamp(c.Offset.X, 0, c.World.Width-c.Rect.Width())
	c.Offset.Y = gmath.Clamp(c.Offset.Y, 0, c.World.Height-c.Rect.Height())
}

func (c *Camera) Pan(delta gmath.Vec) {
	if delta.IsZero() {
		return
	}
	c.Offset = c.Offset.Add(delta)
	c.checkBounds()
}

func (c *Camera) CenterOn(pos gmath.Vec) {
	c.Offset = pos.Sub(c.Rect.Center())
	c.checkBounds()
}

func (c *Camera) SetOffset(pos gmath.Vec) {
	c.Offset = pos
	c.checkBounds()
}

func (c *Camera) Draw(screen *ebiten.Image) {
	c.globalRect = c.Rect
	c.globalRect.Min = c.Offset
	c.globalRect.Max = c.globalRect.Max.Add(c.Offset)

	c.screen.Clear()
	c.bg.DrawPartial(c.screen, c.globalRect)
	c.drawLayer(c.screen, &c.belowObjects)
	c.drawLayer(c.screen, &c.objects)
	c.drawLayer(c.screen, &c.slightlyAboveObjects)
	c.drawLayer(c.screen, &c.aboveObjects)
	if c.fogOfWar != nil {
		var options ebiten.DrawImageOptions
		c.screen.DrawImage(c.fogOfWar, &options)
	}

	var options ebiten.DrawImageOptions
	options.GeoM.Translate(-c.Offset.X, -c.Offset.Y)
	screen.DrawImage(c.screen, &options)
}

type layer struct {
	sprites []*ge.Sprite
	objects []cameraObject
}

func (l *layer) Add(o cameraObject) {
	l.objects = append(l.objects, o)
}

func (l *layer) AddSprite(s *ge.Sprite) {
	l.sprites = append(l.sprites, s)
}
