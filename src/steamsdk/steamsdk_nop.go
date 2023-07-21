//go:build !steam

package steamsdk

import (
	"errors"
)

func ClearAchievements(names []string) {}

func UnlockAchievement(name string) bool {
	return false
}

func IsAchievementUnlocked(name string) (bool, error) {
	return false, errors.New("steamsdk is not available")
}