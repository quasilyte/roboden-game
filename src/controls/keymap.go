package controls

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"

	"github.com/quasilyte/roboden-game/gameinput"
)

const (
	ActionUnknown input.Action = iota

	ActionPanRight
	ActionPanDown
	ActionPanLeft
	ActionPanUp
	ActionPanAlt
	ActionPanDrag

	ActionMenuFocusRight
	ActionMenuFocusDown
	ActionMenuFocusLeft
	ActionMenuFocusUp

	ActionPause

	ActionToggleColony

	ActionShowRecipes

	ActionToggleInterface

	ActionClick

	ActionBack

	ActionDebug

	ActionMoveCursor

	ActionChoice1
	ActionChoice2
	ActionChoice3
	ActionChoice4
	ActionChoice5
	ActionMoveChoice
)

func BindKeymap(ctx *ge.Context) gameinput.Handler {
	keymap := input.Keymap{
		ActionPanRight: {input.KeyD, input.KeyRight, input.KeyGamepadLStickRight, input.KeyGamepadRight},
		ActionPanDown:  {input.KeyS, input.KeyDown, input.KeyGamepadLStickDown, input.KeyGamepadDown},
		ActionPanLeft:  {input.KeyA, input.KeyLeft, input.KeyGamepadLStickLeft, input.KeyGamepadLeft},
		ActionPanUp:    {input.KeyW, input.KeyUp, input.KeyGamepadLStickUp, input.KeyGamepadUp},
		ActionPanAlt:   {input.KeyMouseMiddle},
		ActionPanDrag:  {input.KeyTouchDrag},

		ActionToggleColony: {input.KeyTab, input.KeyGamepadL1},

		ActionShowRecipes: {input.KeyAlt, input.KeyGamepadR2},

		ActionToggleInterface: {input.KeyF11},

		ActionDebug: {input.KeyBackquote},

		ActionBack:  {input.KeyEscape, input.KeyGamepadBack},
		ActionPause: {input.KeySpace, input.KeyGamepadStart},

		ActionMenuFocusRight: {input.KeyRight, input.KeyGamepadRight},
		ActionMenuFocusDown:  {input.KeyDown, input.KeyGamepadDown},
		ActionMenuFocusLeft:  {input.KeyLeft, input.KeyGamepadLeft},
		ActionMenuFocusUp:    {input.KeyUp, input.KeyGamepadUp},

		ActionChoice1:    {input.Key1, input.KeyGamepadY},
		ActionChoice2:    {input.Key2, input.KeyGamepadB},
		ActionChoice3:    {input.Key3, input.KeyGamepadA},
		ActionChoice4:    {input.Key4, input.KeyGamepadX},
		ActionChoice5:    {input.Key5, input.KeyGamepadR1},
		ActionMoveChoice: {input.KeyMouseRight, input.KeyGamepadRStick, input.KeyTouchTap},

		ActionMoveCursor: {input.KeyGamepadRStickMotion},

		ActionClick: {input.KeyMouseLeft, input.KeyGamepadRStick, input.KeyTouchTap},
	}

	return gameinput.Handler{
		Handler: ctx.Input.NewHandler(0, keymap),
	}
}
