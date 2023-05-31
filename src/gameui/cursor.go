package gameui

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
)

type CursorNode struct {
	sprite *ge.Sprite
	input  gameinput.Handler
	pos    gmath.Vec
	rect   gmath.Rect
}

func NewCursorNode(h gameinput.Handler, rect gmath.Rect) *CursorNode {
	return &CursorNode{
		input: h,
		rect:  rect,
	}
}

func (c *CursorNode) Init(scene *ge.Scene) {
	c.pos = c.rect.Center()
	c.sprite = scene.NewSprite(assets.ImageCursor)
	c.sprite.Pos.Base = &c.pos
	c.sprite.Visible = false
	scene.AddGraphicsAbove(c.sprite, 1)
}

func (c *CursorNode) IsDisposed() bool { return false }

func (c *CursorNode) ClickPos(action input.Action) (gmath.Vec, bool) {
	info, ok := c.input.JustPressedActionInfo(action)
	if !ok {
		return gmath.Vec{}, false
	}
	if c.sprite.Visible && info.IsGamepadEvent() {
		return c.pos, true
	}
	return info.Pos, true
}

func (c *CursorNode) Update(delta float64) {
	if info, ok := c.input.PressedActionInfo(controls.ActionMoveCursor); ok {
		c.sprite.Visible = true
		c.pos.X = gmath.Clamp(c.pos.X+info.Pos.X*delta*640, c.rect.Min.X, c.rect.Max.X)
		c.pos.Y = gmath.Clamp(c.pos.Y+info.Pos.Y*delta*640, c.rect.Min.Y, c.rect.Max.Y)
	}
}
