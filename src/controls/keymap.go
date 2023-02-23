package controls

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/roboden-game/session"
)

const (
	ActionUnknown input.Action = iota

	ActionPanRight
	ActionPanDown
	ActionPanLeft
	ActionPanUp
	ActionPanAlt

	ActionToggleColony

	ActionClick

	ActionBack

	ActionDebug

	ActionChoice1
	ActionChoice2
	ActionChoice3
	ActionChoice4
	ActionChoice5
	ActionMoveChoice
)

func BindKeymap(ctx *ge.Context, state *session.State) {
	keymap := input.Keymap{
		ActionPanRight: {input.KeyRight, input.KeyGamepadLStickRight, input.KeyGamepadRight},
		ActionPanDown:  {input.KeyDown, input.KeyGamepadLStickDown, input.KeyGamepadDown},
		ActionPanLeft:  {input.KeyLeft, input.KeyGamepadLStickLeft, input.KeyGamepadLeft},
		ActionPanUp:    {input.KeyUp, input.KeyGamepadLStickUp, input.KeyGamepadUp},
		ActionPanAlt:   {input.KeyMouseMiddle},

		ActionToggleColony: {input.KeyTab, input.KeyGamepadL1},

		ActionDebug: {input.KeyBackquote},

		ActionBack: {input.KeyEscape},

		ActionChoice1:    {input.Key1, input.KeyQ, input.KeyGamepadY},
		ActionChoice2:    {input.Key2, input.KeyW, input.KeyGamepadB},
		ActionChoice3:    {input.Key3, input.KeyE, input.KeyGamepadA},
		ActionChoice4:    {input.Key4, input.KeyR, input.KeyGamepadX},
		ActionChoice5:    {input.Key5, input.KeyT, input.KeyGamepadR1},
		ActionMoveChoice: {input.KeyMouseRight},

		ActionClick: {input.KeyMouseLeft},
	}

	state.MainInput = ctx.Input.NewHandler(0, keymap)
}
