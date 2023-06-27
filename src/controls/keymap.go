package controls

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"

	"github.com/quasilyte/roboden-game/gameinput"
)

const (
	ActionUnknown input.Action = iota

	ActionSkipDemo

	ActionNextTutorialMessage

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

	ActionPing

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

func BindKeymap(ctx *ge.Context) (gameinput.Handler, gameinput.Handler) {
	gamepadKeymap := input.Keymap{
		ActionSkipDemo: {input.KeyGamepadStart},

		ActionPanRight: {input.KeyGamepadLStickRight, input.KeyGamepadRight},
		ActionPanDown:  {input.KeyGamepadLStickDown, input.KeyGamepadDown},
		ActionPanLeft:  {input.KeyGamepadLStickLeft, input.KeyGamepadLeft},
		ActionPanUp:    {input.KeyGamepadLStickUp, input.KeyGamepadUp},

		ActionToggleColony: {input.KeyGamepadL1},

		ActionPing: {input.KeyGamepadLStick},

		ActionShowRecipes: {input.KeyGamepadR2},

		ActionToggleInterface: {input.KeyGamepadL2},

		ActionBack:  {input.KeyGamepadBack},
		ActionPause: {input.KeyGamepadStart},

		ActionMenuFocusRight: {input.KeyGamepadRight},
		ActionMenuFocusDown:  {input.KeyGamepadDown},
		ActionMenuFocusLeft:  {input.KeyGamepadLeft},
		ActionMenuFocusUp:    {input.KeyGamepadUp},

		ActionChoice1:    {input.KeyGamepadY},
		ActionChoice2:    {input.KeyGamepadB},
		ActionChoice3:    {input.KeyGamepadA},
		ActionChoice4:    {input.KeyGamepadX},
		ActionChoice5:    {input.KeyGamepadR1},
		ActionMoveChoice: {input.KeyGamepadRStick},

		ActionMoveCursor: {input.KeyGamepadRStickMotion},

		ActionClick: {input.KeyGamepadRStick},
	}

	mainKeymap := input.Keymap{
		ActionSkipDemo: {input.KeyEnter},

		ActionNextTutorialMessage: {input.KeyEnter},

		ActionPanRight: {input.KeyD, input.KeyRight},
		ActionPanDown:  {input.KeyS, input.KeyDown},
		ActionPanLeft:  {input.KeyA, input.KeyLeft},
		ActionPanUp:    {input.KeyW, input.KeyUp},
		ActionPanAlt:   {input.KeyMouseMiddle},
		ActionPanDrag:  {input.KeyTouchDrag},

		ActionToggleColony: {input.KeyTab},

		ActionPing: {input.KeyWithModifier(input.KeyMouseLeft, input.ModControl)},

		ActionShowRecipes: {input.KeyAlt},

		ActionToggleInterface: {input.KeyF11},

		ActionDebug: {input.KeyBackquote},

		ActionBack:  {input.KeyEscape},
		ActionPause: {input.KeySpace},

		ActionMenuFocusRight: {input.KeyRight},
		ActionMenuFocusDown:  {input.KeyDown},
		ActionMenuFocusLeft:  {input.KeyLeft},
		ActionMenuFocusUp:    {input.KeyUp},

		ActionChoice1:    {input.Key1},
		ActionChoice2:    {input.Key2},
		ActionChoice3:    {input.Key3},
		ActionChoice4:    {input.Key4},
		ActionChoice5:    {input.Key5},
		ActionMoveChoice: {input.KeyMouseRight, input.KeyTouchTap},

		ActionClick: {input.KeyMouseLeft, input.KeyTouchTap},
	}

	for a, keys := range gamepadKeymap {
		mainKeymap[a] = append(mainKeymap[a], keys...)
	}

	primary := gameinput.Handler{
		Handler: ctx.Input.NewHandler(0, mainKeymap),
	}
	second := gameinput.Handler{
		Handler: ctx.Input.NewHandler(1, gamepadKeymap),
	}
	return primary, second
}
