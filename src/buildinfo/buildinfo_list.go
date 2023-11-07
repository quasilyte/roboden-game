package buildinfo

const (
	TagSteam   = "Steam"
	TagItchio  = "itch.io"
	TagAndroid = "Android"
	TagUnknown = "unknown"
)

func IsValidTag(tag string) bool {
	switch tag {
	case TagSteam, TagItchio, TagAndroid:
		return true
	default:
		return false
	}
}
