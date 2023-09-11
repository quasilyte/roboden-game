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

type sortableCameraObject struct {
	graphics  cameraObject
	prevOrder int
	drawOrder *float64
}

type LayerContainer struct {
	belowObjects         layer
	slightlyBelowObjects layer
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

func (c *LayerContainer) AddSpriteSlightlyBelow(s *ge.Sprite) {
	if c.headless {
		return
	}
	c.slightlyBelowObjects.AddSprite(s)
}

func (c *LayerContainer) AddGraphicsBelow(o cameraObject) {
	if c.headless {
		return
	}
	c.belowObjects.Add(o)
}

func (c *LayerContainer) AddSortableGraphics(o cameraObject, order *float64) {
	if c.headless {
		return
	}
	c.objects.AddSortableObject(o, order)
}

func (c *LayerContainer) AddSortableGraphicsSlightlyAbove(o cameraObject, order *float64) {
	if c.headless {
		return
	}
	c.slightlyAboveObjects.AddSortableObject(o, order)
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
	c.slightlyBelowObjects.filter()
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

	shake      int // number of frames
	shakeDelay int // in frames
	shakeIndex uint

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

	if len(l.sortableObjects) != 0 {
		for _, o := range l.sortableObjects {
			if c.isVisible(o.graphics.BoundsRect()) {
				o.graphics.DrawWithOffset(screen, drawOffset)
			}
		}
		l.needSorting = true
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

func (c *Camera) Shake(value int) {
	c.shake = value
	c.shakeIndex += uint(value) * 97
}

func (c *Camera) Draw(screen *ebiten.Image) {
	c.globalRect = c.Rect
	c.globalRect.Min = c.Offset
	c.globalRect.Max = c.globalRect.Max.Add(c.Offset)

	c.Private.belowObjects.filter()
	c.Private.slightlyBelowObjects.filter()
	c.Private.objects.filter()
	c.Private.slightlyAboveObjects.filter()
	c.Private.aboveObjects.filter()

	c.screen.Clear()
	drawOffset := gmath.Vec{
		X: -c.Offset.X,
		Y: -c.Offset.Y,
	}

	rotation := 0.0
	if c.shake > 0 {
		c.shake--
		if c.shakeDelay == 0 {
			c.shakeDelay = 6
			c.shakeIndex++
			drawOffset = drawOffset.Add(shakeOffsets[c.shakeIndex%uint(len(shakeOffsets))])
			switch c.shakeIndex % 3 {
			case 0:
				rotation = 0.005
			case 2:
				rotation = -0.005
			}
		} else {
			c.shakeDelay--
		}
	}

	c.stage.bg.DrawPartialWithOffset(c.screen, c.globalRect, drawOffset)
	c.drawLayer(c.screen, &c.stage.belowObjects, drawOffset)
	c.drawLayer(c.screen, &c.Private.belowObjects, drawOffset)
	c.drawLayer(c.screen, &c.stage.slightlyBelowObjects, drawOffset)
	c.drawLayer(c.screen, &c.Private.slightlyBelowObjects, drawOffset)
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
	if rotation != 0 {
		width := float64(c.screen.Bounds().Dx())
		height := float64(c.screen.Bounds().Dy())
		options.GeoM.Translate(-width*0.5, -height*0.5)
		options.GeoM.Rotate(rotation)
		options.GeoM.Translate(width*0.5, height*0.5)
	}
	screen.DrawImage(c.screen, &options)
}

type layer struct {
	sprites []*ge.Sprite
	objects []cameraObject

	sortableObjects []sortableCameraObject
	needSorting     bool
}

func (l *layer) AddSprite(s *ge.Sprite) {
	l.sprites = append(l.sprites, s)
}

func (l *layer) Add(o cameraObject) {
	l.objects = append(l.objects, o)
}

func (l *layer) AddSortableObject(o cameraObject, order *float64) {
	l.sortableObjects = append(l.sortableObjects, sortableCameraObject{
		graphics:  o,
		drawOrder: order,
	})
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

	if len(l.sortableObjects) != 0 {
		l.filterSortableObjects()
	}
}

func (l *layer) filterSortableObjects() {
	liveSortableObjects := l.sortableObjects[:0]
	changed := false
	for _, o := range l.sortableObjects {
		if o.graphics.IsDisposed() {
			continue
		}
		// Use int-truncated values to sort less often.
		// For instance, if prev value was 1.5 and not it's 1.7,
		// we still consider it to be int(1) and skip the sorting phase.
		order := int(*o.drawOrder)
		if order != o.prevOrder {
			changed = true
		}
		o.prevOrder = order
		liveSortableObjects = append(liveSortableObjects, o)
	}
	l.sortableObjects = liveSortableObjects

	if !l.needSorting {
		return // Nothing else to do here
	}

	l.needSorting = false
	if !changed {
		return // There were no moves, nothing to do here
	}

	// prevOrder holds the current order value at this point.
	xslices.SortStableFunc(l.sortableObjects, func(a, b sortableCameraObject) bool {
		return a.prevOrder < b.prevOrder
	})
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

var shakeOffsets = []gmath.Vec{
	{X: 1, Y: 0},
	{X: 2, Y: 1},
	{X: 1, Y: 3},
	{X: 2, Y: 1},
	{X: 1, Y: 1},
	{X: 2, Y: 0},
	{X: -1, Y: -2},
	{X: -2, Y: 0},
	{X: -2, Y: 1},
	{X: -1, Y: 2},
	{X: 0, Y: 0},
	{X: 1, Y: -1},
	{X: 2, Y: -2},
	{X: 0, Y: -3},
}
