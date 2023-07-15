//go:build steam

package steamsdk

import (
	"github.com/hajimehoshi/go-steamworks"
)

func UnlockAchievement(name string) bool {
	return steamworks.SteamUserStats().SetAchievement(name)
}
