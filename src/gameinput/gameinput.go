package gameinput

import (
	"strings"

	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/controls"
)

type WheelScrollStyle int

const (
	WheelScrollDrag WheelScrollStyle = iota
	WheelScrollFloat
)

type PlayerInputMethod int

const (
	InputMethodCombined PlayerInputMethod = iota
	InputMethodKeyboard
	InputMethodGamepad1
	InputMethodGamepad2
	InputMethodTouch
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

	lastTouchPos gmath.Vec

	InputMethod PlayerInputMethod

	EventGamepadDisconnected gsignal.Event[gsignal.Void]
}

func (h *Handler) PrettyActionName(d *langs.Dictionary, a input.Action) string {
	var mask input.DeviceKind
	switch h.InputMethod {
	case InputMethodCombined:
		if h.GamepadConnected() {
			mask = input.GamepadDevice
		} else {
			mask = input.KeyboardDevice
		}
	case InputMethodKeyboard:
		mask = input.KeyboardDevice
	case InputMethodGamepad1, InputMethodGamepad2:
		mask = input.GamepadDevice
	}
	names := h.ActionKeyNames(a, mask)
	if len(names) == 0 {
		return ""
	}

	n := names[0]
	if pretty := getKeyName(h.layout, n); pretty != "" && pretty != n {
		// A button with a well-known name or symbol (e.g. START, ESC).
		// No extra translation is needed.
		return pretty
	}
	if mask == input.KeyboardDevice {
		// Most of the keyboard buttons can be spelled as is.
		return strings.ToUpper(n)
	}

	return ""
}

func (h *Handler) ReplaceKeyNames(s string) string {
	if h.keysReplacer == nil {
		// Keyboard input.
		return s
	}
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
	if h.InputMethod == InputMethodTouch {
		if info, ok := h.JustPressedActionInfo(controls.ActionClick); ok {
			h.lastTouchPos = info.Pos
		}
		return
	}

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
	if h.InputMethod == InputMethodTouch {
		return h.lastTouchPos
	}
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
	case InputMethodTouch:
		return "touch"
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
