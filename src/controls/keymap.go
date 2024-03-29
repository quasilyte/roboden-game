package controls

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
)

const (
	ActionUnknown input.Action = iota

	ActionSkipDemo

	ActionNextTutorialMessage

	ActionShowTooltip

	ActionPanRight
	ActionPanDown
	ActionPanLeft
	ActionPanUp
	ActionPanAlt
	ActionPanDrag

	ActionPause

	ActionPing

	ActionShowRecipes

	ActionToggleColony
	ActionToggleInterface
	ActionToggleFastForward
	ActionToggleFastForwardAlt

	ActionClick

	ActionExit
	ActionExitConfirm
	ActionMenuBack
	ActionMenuConfirm
	ActionMenuFocusRight
	ActionMenuFocusDown
	ActionMenuFocusLeft
	ActionMenuFocusUp
	ActionMenuTabRight
	ActionMenuTabLeft

	ActionDebug

	ActionMoveCursor
	ActionTestLeftStick

	ActionChoice1
	ActionChoice2
	ActionChoice3
	ActionChoice4
	ActionChoice5
	ActionMoveChoice
)

type KeymapSet struct {
	TouchKeymap         input.Keymap
	CombinedKeymap      input.Keymap
	KeyboardKeymap      input.Keymap
	FirstGamepadKeymap  input.Keymap
	SecondGamepadKeymap input.Keymap
}

func BindKeymap(ctx *ge.Context) KeymapSet {
	touchKeymap := input.Keymap{
		ActionSkipDemo: {input.KeyTouchTap},

		ActionMoveChoice: {input.KeyTouchTap},

		ActionClick: {input.KeyTouchTap},

		ActionPanDrag: {input.KeyTouchDrag},

		ActionShowTooltip: {input.KeyTouchLongTap},
	}

	gamepadKeymap := input.Keymap{
		ActionSkipDemo: {input.KeyGamepadStart},

		ActionNextTutorialMessage: {input.KeyGamepadLStick},

		ActionPanRight: {input.KeyGamepadLStickRight, input.KeyGamepadRight},
		ActionPanDown:  {input.KeyGamepadLStickDown, input.KeyGamepadDown},
		ActionPanLeft:  {input.KeyGamepadLStickLeft, input.KeyGamepadLeft},
		ActionPanUp:    {input.KeyGamepadLStickUp, input.KeyGamepadUp},

		ActionToggleColony: {input.KeyGamepadL1},

		ActionPing:                 {input.KeyGamepadLStick},
		ActionToggleFastForwardAlt: {input.KeyGamepadLStick},

		ActionShowRecipes: {input.KeyGamepadR2},

		ActionToggleInterface: {input.KeyGamepadL2},

		ActionExit:        {input.KeyGamepadBack},
		ActionExitConfirm: {input.KeyGamepadBack},
		ActionMenuBack:    {input.KeyGamepadBack, input.KeyGamepadB},
		ActionPause:       {input.KeyGamepadStart, input.KeyGamepadHome},

		ActionMenuConfirm:    {input.KeyGamepadA},
		ActionMenuFocusRight: {input.KeyGamepadRight},
		ActionMenuFocusDown:  {input.KeyGamepadDown},
		ActionMenuFocusLeft:  {input.KeyGamepadLeft},
		ActionMenuFocusUp:    {input.KeyGamepadUp},
		ActionMenuTabRight:   {input.KeyGamepadR1},
		ActionMenuTabLeft:    {input.KeyGamepadL1},

		ActionChoice1:    {input.KeyGamepadY},
		ActionChoice2:    {input.KeyGamepadB},
		ActionChoice3:    {input.KeyGamepadA},
		ActionChoice4:    {input.KeyGamepadX},
		ActionChoice5:    {input.KeyGamepadR1},
		ActionMoveChoice: {input.KeyGamepadRStick},

		ActionTestLeftStick: {input.KeyGamepadLStickMotion},
		ActionMoveCursor:    {input.KeyGamepadRStickMotion},

		ActionClick: {input.KeyGamepadRStick},
	}

	keyboardKeymap := input.Keymap{
		ActionSkipDemo: {input.KeyEnter},

		ActionNextTutorialMessage: {input.KeyEnter},

		ActionPanRight: {input.KeyD, input.KeyRight},
		ActionPanDown:  {input.KeyS, input.KeyDown},
		ActionPanLeft:  {input.KeyA, input.KeyLeft},
		ActionPanUp:    {input.KeyW, input.KeyUp},
		ActionPanAlt:   {input.KeyMouseMiddle},
		ActionPanDrag:  {input.KeyTouchDrag},

		ActionToggleColony: {input.KeyTab},

		ActionToggleFastForward: {input.KeyF},

		ActionPing: {input.KeyWithModifier(input.KeyMouseLeft, input.ModControl)},

		ActionShowRecipes: {input.KeyAlt},

		ActionToggleInterface: {input.KeyF11},

		ActionDebug: {input.KeyBackquote},

		ActionExit:     {input.KeyEscape},
		ActionMenuBack: {input.KeyEscape},
		ActionPause:    {input.KeySpace},

		ActionMenuConfirm:    {}, // TODO: KeyEnter? It clashes with ebitenui default submit though
		ActionMenuFocusRight: {input.KeyRight},
		ActionMenuFocusDown:  {input.KeyDown},
		ActionMenuFocusLeft:  {input.KeyLeft},
		ActionMenuFocusUp:    {input.KeyUp},
		ActionMenuTabRight:   {input.KeyTab},
		ActionMenuTabLeft:    {input.KeyWithModifier(input.KeyTab, input.ModShift)},

		ActionChoice1:    {input.Key1},
		ActionChoice2:    {input.Key2},
		ActionChoice3:    {input.Key3},
		ActionChoice4:    {input.Key4},
		ActionChoice5:    {input.Key5},
		ActionMoveChoice: {input.KeyMouseRight},

		ActionClick: {input.KeyMouseLeft},
	}

	mainKeymap := input.Keymap{
		ActionMoveChoice: {input.KeyTouchTap},

		ActionClick: {input.KeyTouchTap},
	}

	for a, keys := range gamepadKeymap {
		mainKeymap[a] = append(mainKeymap[a], keys...)
	}
	for a, keys := range keyboardKeymap {
		mainKeymap[a] = append(mainKeymap[a], keys...)
	}

	return KeymapSet{
		TouchKeymap:         touchKeymap,
		CombinedKeymap:      mainKeymap,
		KeyboardKeymap:      keyboardKeymap,
		FirstGamepadKeymap:  gamepadKeymap,
		SecondGamepadKeymap: gamepadKeymap,
	}
}
