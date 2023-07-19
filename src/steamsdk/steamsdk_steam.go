//go:build steam

package steamsdk

import (
	"errors"
)

import (
	"github.com/hajimehoshi/go-steamworks"
)

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
