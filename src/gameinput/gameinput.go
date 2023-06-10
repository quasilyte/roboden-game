package gameinput

import (
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
)

type Cursor interface {
	ClickPos(input.Action) (gmath.Vec, bool)
}

type Handler struct {
	*input.Handler
	virtualCursorPos gmath.Vec
}

func (h *Handler) SetGamepadDeadzoneLevel(level int) {
	value := (0.05 * float64(level)) + 0.055
	h.GamepadDeadzone = value
}

func (h *Handler) UpdateVirtualCursorPos(pos gmath.Vec) {
	h.virtualCursorPos = pos
}

func (h *Handler) AnyCursorPos() gmath.Vec {
	if !h.virtualCursorPos.IsZero() {
		return h.virtualCursorPos
	}
	return h.CursorPos()
}

func (h *Handler) ClickPos(action input.Action) (gmath.Vec, bool) {
	info, ok := h.JustPressedActionInfo(action)
	if !ok {
		return gmath.Vec{}, false
	}
	if info.IsGamepadEvent() {
		return h.virtualCursorPos, true
	}
	return info.Pos, true
}
