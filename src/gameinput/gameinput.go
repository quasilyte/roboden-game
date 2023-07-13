package gameinput

import (
	"strings"

	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
)

type PlayerInputMethod int

const (
	InputMethodCombined PlayerInputMethod = iota
	InputMethodKeyboard
	InputMethodGamepad1
	InputMethodGamepad2
)

type Cursor interface {
	ClickPos(input.Action) (gmath.Vec, bool)
}

type Handler struct {
	*input.Handler

	virtualCursorPos gmath.Vec

	virtualCursorSpeedMultiplier float64

	layout       GamepadLayoutKind
	keysReplacer *strings.Replacer

	gamepadConnected bool

	InputMethod PlayerInputMethod

	EventGamepadDisconnected gsignal.Event[gsignal.Void]
}

func (h *Handler) ReplaceKeyNames(s string) string {
	return h.keysReplacer.Replace(s)
}

func (h *Handler) SetGamepadLayout(l GamepadLayoutKind) {
	h.layout = l

	keys := []string{
		"gamepad_back",
		"gamepad_start",
		"gamepad_y",
		"gamepad_b",
		"gamepad_a",
		"gamepad_x",
	}
	pairs := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		pairs = append(pairs, "$"+k, getKeyName(l, k))
	}
	h.keysReplacer = strings.NewReplacer(pairs...)
}

func (h *Handler) GetVirtualCursorSpeedMultiplier() float64 {
	return h.virtualCursorSpeedMultiplier
}

func (h *Handler) SetVirtualCursorSpeed(level int) {
	h.virtualCursorSpeedMultiplier = ([...]float64{
		0.2,
		0.5,
		0.8,
		1.0,
		1.2,
		1.5,
		1.8,
		2.0,
	})[level]
}

func (h *Handler) SetGamepadDeadzoneLevel(level int) {
	value := (0.05 * float64(level)) + 0.055
	h.GamepadDeadzone = value
}

func (h *Handler) Update() {
	gamepadConnected := h.GamepadConnected()
	if gamepadConnected != h.gamepadConnected {
		if !gamepadConnected {
			h.EventGamepadDisconnected.Emit(gsignal.Void{})
		}
		h.gamepadConnected = gamepadConnected
	}
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

func (h *Handler) HasMouseInput() bool {
	switch h.InputMethod {
	case InputMethodCombined, InputMethodKeyboard:
		return true
	default:
		return false
	}
}

func (h *Handler) CanHideMousePointer() bool {
	return h.InputMethod == InputMethodCombined
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

func (h *Handler) DetectInputMode() string {
	switch PlayerInputMethod(h.InputMethod) {
	case InputMethodKeyboard:
		return "keyboard"
	case InputMethodGamepad1, InputMethodGamepad2:
		return "gamepad"
	case InputMethodCombined:
		if h.GamepadConnected() {
			return "gamepad"
		}
		return "keyboard"
	default:
		return "keyboard"
	}
}
