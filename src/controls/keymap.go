package controls

import (
	"github.com/quasilyte/colony-game/session"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
)

const (
	ActionUnknown input.Action = iota

	ActionPanRight
	ActionPanDown
	ActionPanLeft
	ActionPanUp

	ActionToggleColony

	ActionChoice1
	ActionChoice2
	ActionChoice3
	ActionChoice4
	ActionChoice5
)

func BindKeymap(ctx *ge.Context, state *session.State) {
	keymap := input.Keymap{
		ActionPanRight: {input.KeyRight, input.KeyGamepadLStickRight, input.KeyGamepadRight},
		ActionPanDown:  {input.KeyDown, input.KeyGamepadLStickDown, input.KeyGamepadDown},
		ActionPanLeft:  {input.KeyLeft, input.KeyGamepadLStickLeft, input.KeyGamepadLeft},
		ActionPanUp:    {input.KeyUp, input.KeyGamepadLStickUp, input.KeyGamepadUp},

		ActionToggleColony: {input.KeyTab, input.KeyGamepadL1},

		ActionChoice1: {input.Key1, input.KeyQ, input.KeyGamepadY},
		ActionChoice2: {input.Key2, input.KeyW, input.KeyGamepadB},
		ActionChoice3: {input.Key3, input.KeyE, input.KeyGamepadA},
		ActionChoice4: {input.Key4, input.KeyR, input.KeyGamepadX},
		ActionChoice5: {input.Key5, input.KeyT, input.KeyA, input.KeyGamepadR1},
	}

	state.MainInput = ctx.Input.NewHandler(0, keymap)
}
