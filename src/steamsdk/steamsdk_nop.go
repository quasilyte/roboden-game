//go:build !steam

package steamsdk

func UnlockAchievement(name string) bool {
	return false
}
