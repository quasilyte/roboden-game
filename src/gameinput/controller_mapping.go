package gameinput

type GamepadLayoutKind int

const (
	GamepadLayoutXbox GamepadLayoutKind = iota
	GamepadLayoutPlayStation
	GamepadLayoutNintendoSwitch
	GamepadLayoutSteamDeck
)

var (
	gamepadY     = [...]string{"Y", "△", "X", "Y"}
	gamepadB     = [...]string{"B", "○", "A", "B"}
	gamepadA     = [...]string{"A", "×", "B", "A"}
	gamepadX     = [...]string{"X", "□", "Y", "X"}
	gamepadStart = [...]string{"START", "START", "+", "☰"}
	gamepadBack  = [...]string{"BACK", "SELECT", "-", "❐"}
)

func getKeyName(layout GamepadLayoutKind, key string) string {
	switch key {
	case "gamepad_y":
		return gamepadY[layout]
	case "gamepad_b":
		return gamepadB[layout]
	case "gamepad_a":
		return gamepadA[layout]
	case "gamepad_x":
		return gamepadX[layout]
	case "gamepad_start":
		return gamepadStart[layout]
	case "gamepad_back":
		return gamepadBack[layout]

	case "gamepad_l1":
		return "L1"

	case "escape":
		return "ESC"
	}

	return ""
}
