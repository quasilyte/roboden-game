package gameinput

import (
	"strings"

	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/ge/xslices"
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
	input *input.Handler

	virtualCursorPos gmath.Vec

	virtualCursorSpeedMultiplier float64

	layout       GamepadLayoutKind
	keysReplacer *strings.Replacer

	gamepadConnected bool

	lastTouchPos gmath.Vec

	consumed []input.Action

	InputMethod PlayerInputMethod

	EventGamepadDisconnected gsignal.Event[gsignal.Void]
}

func MakeHandler(m PlayerInputMethod, h *input.Handler) Handler {
	return Handler{
		InputMethod: m,
		input:       h,
	}
}

func (h *Handler) IsClickDevice() bool {
	switch h.InputMethod {
	case InputMethodCombined:
		return !h.input.GamepadConnected()
	case InputMethodKeyboard, InputMethodTouch:
		return true
	default:
		return false
	}
}

func (h *Handler) PrettyActionName(d *langs.Dictionary, a input.Action) string {
	var mask input.DeviceKind
	switch h.InputMethod {
	case InputMethodCombined:
		if h.input.GamepadConnected() {
			mask = input.GamepadDevice
		} else {
			mask = input.KeyboardDevice
		}
	case InputMethodKeyboard:
		mask = input.KeyboardDevice
	case InputMethodGamepad1, InputMethodGamepad2:
		mask = input.GamepadDevice
	}
	names := h.input.ActionKeyNames(a, mask)
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
	h.input.GamepadDeadzone = value
}

func (h *Handler) Update() {
	h.consumed = h.consumed[:0]

	if h.InputMethod == InputMethodTouch {
		if info, ok := h.input.JustPressedActionInfo(controls.ActionClick); ok {
			h.lastTouchPos = info.Pos
		}
		return
	}

	gamepadConnected := h.input.GamepadConnected()
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

func (h *Handler) GamepadConnected() bool {
	return h.input.GamepadConnected()
}

func (h *Handler) MouseCursorPos() gmath.Vec {
	return h.input.CursorPos()
}

func (h *Handler) AnyCursorPos() gmath.Vec {
	if h.InputMethod == InputMethodTouch {
		return h.lastTouchPos
	}
	if !h.virtualCursorPos.IsZero() {
		return h.virtualCursorPos
	}
	return h.input.CursorPos()
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

func (h *Handler) MarkConsumed(action input.Action) {
	if !xslices.Contains(h.consumed, action) {
		h.consumed = append(h.consumed, action)
	}
}

func (h *Handler) PressedActionInfo(action input.Action) (input.EventInfo, bool) {
	if len(h.consumed) != 0 && xslices.Contains(h.consumed, action) {
		return input.EventInfo{}, false
	}
	return h.input.PressedActionInfo(action)
}

func (h *Handler) JustReleasedActionInfo(action input.Action) (input.EventInfo, bool) {
	if len(h.consumed) != 0 && xslices.Contains(h.consumed, action) {
		return input.EventInfo{}, false
	}
	return h.input.JustReleasedActionInfo(action)
}

func (h *Handler) JustPressedActionInfo(action input.Action) (input.EventInfo, bool) {
	if len(h.consumed) != 0 && xslices.Contains(h.consumed, action) {
		return input.EventInfo{}, false
	}
	return h.input.JustPressedActionInfo(action)
}

func (h *Handler) ActionIsPressed(action input.Action) bool {
	if len(h.consumed) != 0 && xslices.Contains(h.consumed, action) {
		return false
	}
	return h.input.ActionIsPressed(action)
}

func (h *Handler) ActionIsJustPressed(action input.Action) bool {
	if len(h.consumed) != 0 && xslices.Contains(h.consumed, action) {
		return false
	}
	return h.input.ActionIsJustPressed(action)
}

func (h *Handler) ClickPos(action input.Action) (gmath.Vec, bool) {
	if len(h.consumed) != 0 && xslices.Contains(h.consumed, action) {
		return gmath.Vec{}, false
	}

	info, ok := h.input.JustPressedActionInfo(action)
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
		if h.input.GamepadConnected() {
			return "gamepad"
		}
		return "keyboard"
	default:
		return "keyboard"
	}
}
