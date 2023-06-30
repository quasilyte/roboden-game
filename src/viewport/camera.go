package viewport

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
)

type World struct {
	Width  float64
	Height float64
}

type cameraObject interface {
	DrawWithOffset(dst *ebiten.Image, offset gmath.Vec)
	IsDisposed() bool
	BoundsRect() gmath.Rect
}

type LayerContainer struct {
	belowObjects         layer
	objects              layer
	slightlyAboveObjects layer
	aboveObjects         layer

	headless bool
}

func (c *LayerContainer) AddSprite(s *ge.Sprite) {
	if c.headless {
		return
	}
	c.objects.AddSprite(s)
}

func (c *LayerContainer) AddGraphics(o cameraObject) {
	if c.headless {
		return
	}
	c.objects.Add(o)
}

func (c *LayerContainer) MoveSlightlyAboveSpriteDown(s *ge.Sprite) {
	if c.headless {
		return
	}
	index := xslices.Index(c.slightlyAboveObjects.sprites, s)
	if index == -1 {
		return
	}
	if index == 0 {
		return // Already at the bottom (in terms of the rendering order)
	}
	// Since this "slightly above" layer is only used for colony cores (and their selectors),
	// this slice is very small. If that will change, we'll need to find a different approach.
	copy(c.slightlyAboveObjects.sprites[1:], c.slightlyAboveObjects.sprites[:index])
	c.slightlyAboveObjects.sprites[0] = s
}

func (c *LayerContainer) AddSpriteSlightlyAbove(s *ge.Sprite) {
	if c.headless {
		return
	}
	c.slightlyAboveObjects.AddSprite(s)
}

func (c *LayerContainer) AddSpriteAbove(s *ge.Sprite) {
	if c.headless {
		return
	}
	c.aboveObjects.AddSprite(s)
}

func (c *LayerContainer) AddGraphicsSlightlyAbove(o cameraObject) {
	if c.headless {
		return
	}
	c.slightlyAboveObjects.Add(o)
}

func (c *LayerContainer) AddGraphicsAbove(o cameraObject) {
	if c.headless {
		return
	}
	c.aboveObjects.Add(o)
}

func (c *LayerContainer) AddSpriteBelow(s *ge.Sprite) {
	if c.headless {
		return
	}
	c.belowObjects.AddSprite(s)
}

type CameraStage struct {
	LayerContainer

	bg       *ge.TiledBackground
	fogOfWar *ebiten.Image
}

func (c *CameraStage) Update() {
	if c.headless {
		return
	}

	c.belowObjects.filter()
	c.objects.filter()
	c.slightlyAboveObjects.filter()
	c.aboveObjects.filter()
}

func (c *CameraStage) SetFogOfWar(img *ebiten.Image) {
	c.fogOfWar = img
}

func (c *CameraStage) SetBackground(bg *ge.TiledBackground) {
	c.bg = bg
}

func (c *CameraStage) SortBelowLayer() {
	if c.headless {
		return
	}
	if len(c.belowObjects.objects) != 0 {
		panic("unexpected below objects count")
	}
	sort.SliceStable(c.belowObjects.sprites, func(i, j int) bool {
		shape1 := c.belowObjects.sprites[i].BoundsRect()
		shape2 := c.belowObjects.sprites[j].BoundsRect()
		return shape1.Max.Y < shape2.Max.Y
	})
}

type Camera struct {
	World *World

	stage *CameraStage

	Offset    gmath.Vec
	ScreenPos gmath.Vec

	Rect       gmath.Rect
	globalRect gmath.Rect

	// A layer that is always on top of everything else.
	// It's also position-independent.
	UI *UserInterfaceLayer

	// Objects that are only rendered for this player.
	Private LayerContainer

	screen *ebiten.Image

	disposed bool
}

type UserInterfaceLayer struct {
	belowObjects []ge.SceneGraphics
	objects      []ge.SceneGraphics
	aboveObjects []ge.SceneGraphics

	Visible bool
}

func (l *UserInterfaceLayer) AddGraphicsBelow(o ge.SceneGraphics) {
	if l == nil {
		return
	}
	l.belowObjects = append(l.belowObjects, o)
}

func (l *UserInterfaceLayer) AddGraphics(o ge.SceneGraphics) {
	if l == nil {
		return
	}
	l.objects = append(l.objects, o)
}

func (l *UserInterfaceLayer) AddGraphicsAbove(o ge.SceneGraphics) {
	if l == nil {
		return
	}
	l.aboveObjects = append(l.aboveObjects, o)
}

func NewCameraStage(headless bool) *CameraStage {
	stage := &CameraStage{}
	stage.headless = headless
	return stage
}

func NewCamera(w *World, stage *CameraStage, width, height float64) *Camera {
	c := &Camera{
		World: w,
		Rect: gmath.Rect{
			Min: gmath.Vec{},
			Max: gmath.Vec{X: width, Y: height},
		},
		stage: stage,
	}
	c.Private.headless = stage.headless
	if !stage.headless {
		c.UI = &UserInterfaceLayer{
			belowObjects: make([]ge.SceneGraphics, 0, 4),
			objects:      make([]ge.SceneGraphics, 0, 4),
			aboveObjects: make([]ge.SceneGraphics, 0, 4),
			Visible:      true,
		}
	}
	if !stage.headless {
		c.screen = ebiten.NewImage(int(width), int(height))
	}
	return c
}

func (c *Camera) RenderToImage() *ebiten.Image {
	result := ebiten.NewImage(c.screen.Size())
	c.Draw(result)
	return result
}

func (c *Camera) Dispose() {
	c.disposed = true
}

func (c *Camera) AbsClickPos(relativePos gmath.Vec) gmath.Vec {
	return relativePos.Add(c.Offset).Sub(c.ScreenPos)
}

func (c *Camera) AbsPos(relativePos gmath.Vec) gmath.Vec {
	return relativePos.Add(c.Offset)
}

func (c *Camera) IsDisposed() bool {
	return c.disposed
}

func (c *Camera) drawLayer(screen *ebiten.Image, l *layer, drawOffset gmath.Vec) {
	for _, s := range l.sprites {
		if c.isVisible(s.BoundsRect()) {
			s.DrawWithOffset(screen, drawOffset)
		}
	}

	if len(l.objects) != 0 {
		for _, o := range l.objects {
			if c.isVisible(o.BoundsRect()) {
				o.DrawWithOffset(screen, drawOffset)
			}
		}
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

func (c Camera) CenterPos() gmath.Vec {
	return c.Offset.Add(c.Rect.Center())
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

	c.Private.belowObjects.filter()
	c.Private.objects.filter()
	c.Private.slightlyAboveObjects.filter()
	c.Private.aboveObjects.filter()

	c.screen.Clear()
	drawOffset := gmath.Vec{
		X: -c.Offset.X,
		Y: -c.Offset.Y,
	}
	c.stage.bg.DrawPartialWithOffset(c.screen, c.globalRect, drawOffset)
	c.drawLayer(c.screen, &c.stage.belowObjects, drawOffset)
	c.drawLayer(c.screen, &c.Private.belowObjects, drawOffset)
	c.drawLayer(c.screen, &c.stage.objects, drawOffset)
	c.drawLayer(c.screen, &c.Private.objects, drawOffset)
	c.drawLayer(c.screen, &c.stage.slightlyAboveObjects, drawOffset)
	c.drawLayer(c.screen, &c.Private.slightlyAboveObjects, drawOffset)
	c.drawLayer(c.screen, &c.stage.aboveObjects, drawOffset)
	c.drawLayer(c.screen, &c.Private.aboveObjects, drawOffset)
	if c.stage.fogOfWar != nil {
		var options ebiten.DrawImageOptions
		options.GeoM.Translate(drawOffset.X, drawOffset.Y)
		c.screen.DrawImage(c.stage.fogOfWar, &options)
	}
	if c.UI != nil && c.UI.Visible {
		c.UI.belowObjects = drawSlice(c.screen, c.UI.belowObjects)
		c.UI.objects = drawSlice(c.screen, c.UI.objects)
		c.UI.aboveObjects = drawSlice(c.screen, c.UI.aboveObjects)
	}

	var options ebiten.DrawImageOptions
	options.GeoM.Translate(c.ScreenPos.X, c.ScreenPos.Y)
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

func (l *layer) filter() {
	liveSprites := l.sprites[:0]
	for _, s := range l.sprites {
		if s.IsDisposed() {
			continue
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
			liveObjects = append(liveObjects, o)
		}
		l.objects = liveObjects
	}
}

func drawSlice(dst *ebiten.Image, slice []ge.SceneGraphics) []ge.SceneGraphics {
	live := slice[:0]
	for _, o := range slice {
		if o.IsDisposed() {
			continue
		}
		o.Draw(dst)
		live = append(live, o)
	}
	return live
}
