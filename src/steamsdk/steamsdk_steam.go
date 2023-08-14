//go:build steam

package steamsdk

import (
	"errors"

	"github.com/hajimehoshi/go-steamworks"
	"github.com/quasilyte/gmath"
)

func PlayerName() string {
	return steamworks.SteamFriends().GetPersonaName()
}

func ShowSteamDeckKeyboard(textFieldRect gmath.Rect) bool {
	x := int32(textFieldRect.Min.X)
	y := int32(textFieldRect.Min.Y)
	width := int32(textFieldRect.Width())
	height := int32(textFieldRect.Height())
	return steamworks.SteamUtils().ShowFloatingGamepadTextInput(steamworks.EFloatingGamepadTextInputMode_ModeNumeric, x, y, width, height)
}

func ClearAchievements(names []string) {
	for _, name := range names {
		steamworks.SteamUserStats().ClearAchievement(name)
	}
}

func UnlockAchievement(name string) bool {
	return steamworks.SteamUserStats().SetAchievement(name)
}

func IsAchievementUnlocked(name string) (bool, error) {
	unlocked, ok := steamworks.SteamUserStats().GetAchievement(name)
	if ok {
		return unlocked, nil
	}
	return false, errors.New("failed to fetch achievement info")
}
