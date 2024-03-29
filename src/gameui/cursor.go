package gameui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
)

type CursorNode struct {
	sprite *ge.Sprite
	input  *gameinput.Handler
	pos    gmath.Vec
	rect   gmath.Rect

	gamepad        bool
	hoverTriggered bool
	stillTime      float64
	hoverPos       gmath.Vec
	prevMousePos   gmath.Vec

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

	switch c.input.InputMethod {
	case gameinput.InputMethodGamepad1, gameinput.InputMethodGamepad2:
		c.gamepad = true
	}
}

func (c *CursorNode) IsDisposed() bool { return false }

func (c *CursorNode) ClickPos(action input.Action) (gmath.Vec, bool) {
	if c.input.InputMethod == gameinput.InputMethodTouch {
		info, ok := c.input.JustPressedActionInfo(action)
		if !ok {
			return gmath.Vec{}, false
		}
		return info.Pos, true
	}

	info, ok := c.input.JustReleasedActionInfo(action)
	if !ok {
		return gmath.Vec{}, false
	}
	if c.sprite.Visible && info.IsGamepadEvent() {
		return c.pos, true
	}
	return info.Pos, true
}

func (c *CursorNode) VirtualCursorIsVisible() bool {
	return c.sprite.Visible
}

func (c *CursorNode) setPreferGamepad(gamepad bool) {
	c.gamepad = gamepad
	c.sprite.Visible = gamepad
	if gamepad {
		ebiten.SetCursorMode(ebiten.CursorModeHidden)
		c.prevMousePos = c.input.MouseCursorPos()
		c.pos = c.prevMousePos
	} else {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}
}

func (c *CursorNode) Update(delta float64) {
	if info, ok := c.input.PressedActionInfo(controls.ActionMoveCursor); ok {
		if !c.gamepad && c.input.CanHideMousePointer() {
			c.setPreferGamepad(true)
		}
		c.sprite.Visible = true
		travelled := (delta * 640) * c.input.GetVirtualCursorSpeedMultiplier()
		c.pos.X = gmath.Clamp(c.pos.X+info.Pos.X*travelled, c.rect.Min.X, c.rect.Max.X)
		c.pos.Y = gmath.Clamp(c.pos.Y+info.Pos.Y*travelled, c.rect.Min.Y, c.rect.Max.Y)
	}

	if !c.EventHover.IsEmpty() {
		pos := c.pos
		maxDistSqr := 8.0 * 8.0
		if !c.gamepad {
			pos = c.input.MouseCursorPos()
			maxDistSqr = 6 * 6
		}
		if c.hoverPos.IsZero() {
			c.hoverPos = pos
		}
		distSqr := pos.DistanceSquaredTo(c.hoverPos)
		if distSqr < maxDistSqr {
			if !c.hoverTriggered {
				c.stillTime += delta
				if c.hoverPos.IsZero() && c.stillTime > 0.15 {
					c.hoverPos = pos
				}
				if c.stillTime > 0.3 {
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
	}

	if c.gamepad && c.input.CanHideMousePointer() {
		if c.input.MouseCursorPos().DistanceSquaredTo(c.prevMousePos) > 10 {
			c.setPreferGamepad(false)
		}
	}
	c.prevMousePos = c.input.MouseCursorPos()
}
