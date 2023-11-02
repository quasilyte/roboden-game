package buildinfo

const (
	TagSteam  = "Steam"
	TagItchio = "itch.io"
)

func IsValidTag(tag string) bool {
	switch tag {
	case TagSteam, TagItchio:
		return true
	default:
		return false
	}
}
