package gameui

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
)

type CursorNode struct {
	sprite  *ge.Sprite
	input   *gameinput.Handler
	prevPos gmath.Vec
	pos     gmath.Vec
	rect    gmath.Rect

	gamepad        bool
	hoverTriggered bool
	stillTime      float64
	hoverPos       gmath.Vec

	EventHover     gsignal.Event[gmath.Vec]
	EventStopHover gsignal.Event[gsignal.Void]
}

func NewCursorNode(h *gameinput.Handler, rect gmath.Rect) *CursorNode {
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
		c.gamepad = true
		c.sprite.Visible = true
		travelled := (delta * 640) * c.input.GetVirtualCursorSpeedMultiplier()
		c.pos.X = gmath.Clamp(c.pos.X+info.Pos.X*travelled, c.rect.Min.X, c.rect.Max.X)
		c.pos.Y = gmath.Clamp(c.pos.Y+info.Pos.Y*travelled, c.rect.Min.Y, c.rect.Max.Y)
	}

	if !c.EventHover.IsEmpty() {
		pos := c.pos
		if !c.gamepad {
			pos = c.input.CursorPos()
		}
		dist := pos.DistanceSquaredTo(c.prevPos)
		if dist < 1 {
			if !c.hoverTriggered {
				c.stillTime += delta
				if c.hoverPos.IsZero() && c.stillTime > 0.3 {
					c.hoverPos = pos
				}
				if c.stillTime > 0.6 {
					c.hoverTriggered = true
					c.EventHover.Emit(c.hoverPos)
				}
			}
		} else {
			if c.hoverTriggered && c.stillTime > 0 {
				c.hoverTriggered = false
				c.EventStopHover.Emit(gsignal.Void{})
			}
			c.stillTime = 0
			c.hoverPos = gmath.Vec{}
		}
		c.prevPos = pos
	}
}
