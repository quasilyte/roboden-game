package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
)

type cursorNode struct {
	sprite *ge.Sprite
	input  *input.Handler
	pos    gmath.Vec
	rect   gmath.Rect
}

func newCursorNode(h *input.Handler, rect gmath.Rect) *cursorNode {
	return &cursorNode{
		input: h,
		rect:  rect,
	}
}

func (c *cursorNode) Init(scene *ge.Scene) {
	c.pos = c.rect.Center()
	c.sprite = scene.NewSprite(assets.ImageCursor)
	c.sprite.Pos.Base = &c.pos
	c.sprite.Visible = false
	scene.AddGraphicsAbove(c.sprite, 1)
}

func (c *cursorNode) IsDisposed() bool { return false }

func (c *cursorNode) ClickPos() (gmath.Vec, bool) {
	info, ok := c.input.JustPressedActionInfo(controls.ActionMoveChoice)
	if !ok {
		return gmath.Vec{}, false
	}
	if c.sprite.Visible && info.IsGamepadEvent() {
		return c.pos, true
	}
	return info.Pos, true
}

func (c *cursorNode) Update(delta float64) {
	if info, ok := c.input.PressedActionInfo(controls.ActionMoveCursor); ok {
		c.sprite.Visible = true
		c.pos.X = gmath.Clamp(c.pos.X+info.Pos.X*delta*640, 0, c.rect.Width())
		c.pos.Y = gmath.Clamp(c.pos.Y+info.Pos.Y*delta*640, 0, c.rect.Height())
	}
}
