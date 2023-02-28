package viewport

import (
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
	belowObjects         []cameraObject
	objects              []cameraObject
	slightlyAboveObjects []cameraObject
	aboveObjects         []cameraObject

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

func (c *Camera) AddGraphics(o cameraObject) {
	c.objects = append(c.objects, o)
}

func (c *Camera) AddGraphicsSlightlyAbove(o cameraObject) {
	c.slightlyAboveObjects = append(c.slightlyAboveObjects, o)
}

func (c *Camera) AddGraphicsAbove(o cameraObject) {
	c.aboveObjects = append(c.aboveObjects, o)
}

func (c *Camera) AddGraphicsBelow(o cameraObject) {
	c.belowObjects = append(c.belowObjects, o)
}

func (c *Camera) SetBackground(bg *ge.TiledBackground) {
	c.bg = bg
}

func (c *Camera) drawSlice(screen *ebiten.Image, objects []cameraObject) []cameraObject {
	liveObjects := objects[:0]
	for _, o := range objects {
		if o.IsDisposed() {
			continue
		}
		if c.isVisible(o) {
			o.Draw(c.screen)
		}
		liveObjects = append(liveObjects, o)
	}
	return liveObjects
}

func (c *Camera) isVisible(o cameraObject) bool {
	objectRect := o.BoundsRect()
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
	if c.bg != nil {
		c.bg.DrawPartial(c.screen, c.globalRect)
	}
	c.belowObjects = c.drawSlice(c.screen, c.belowObjects)
	c.objects = c.drawSlice(c.screen, c.objects)
	c.slightlyAboveObjects = c.drawSlice(c.screen, c.slightlyAboveObjects)
	c.aboveObjects = c.drawSlice(c.screen, c.aboveObjects)

	var options ebiten.DrawImageOptions
	options.GeoM.Translate(-c.Offset.X, -c.Offset.Y)
	screen.DrawImage(c.screen, &options)
}
