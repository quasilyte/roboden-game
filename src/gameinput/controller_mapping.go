package gameinput

type GamepadLayoutKind int

const (
	GamepadLayoutXbox GamepadLayoutKind = iota
	GamepadLayoutPlayStation
	GamepadLayoutNintendoSwitch
)

var (
	gamepadY     = [...]string{"Y", "△", "X"}
	gamepadB     = [...]string{"B", "○", "A"}
	gamepadA     = [...]string{"A", "×", "B"}
	gamepadX     = [...]string{"X", "□", "Y"}
	gamepadStart = [...]string{"START", "START", "+"}
	gamepadBack  = [...]string{"BACK", "SELECT", "-"}
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
	}
	return ""
}
