package buildinfo

const (
	TagSteam   = "Steam"
	TagItchio  = "itch.io"
	TagAndroid = "Android"
)

func IsValidTag(tag string) bool {
	switch tag {
	case TagSteam, TagItchio, TagAndroid:
		return true
	default:
		return false
	}
}
