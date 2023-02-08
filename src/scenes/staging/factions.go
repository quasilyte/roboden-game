package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
)

type faction struct {
	tag   factionTag
	color color.RGBA
}

type factionTag int

const (
	neutralFactionTag factionTag = iota
	yellowFactionTag
	redFactionTag
	greenFactionTag
	blueFactionTag
)

func factionByTag(tag factionTag) *faction {
	switch tag {
	case yellowFactionTag:
		return yellowFaction
	case redFactionTag:
		return redFaction
	case greenFactionTag:
		return greenFaction
	case blueFactionTag:
		return blueFaction
	default:
		return nil
	}
}

var (
	yellowFaction = &faction{
		tag:   yellowFactionTag,
		color: ge.RGB(0xf1e851),
	}

	redFaction = &faction{
		tag:   redFactionTag,
		color: ge.RGB(0xef6a57),
	}

	greenFaction = &faction{
		tag:   greenFactionTag,
		color: ge.RGB(0x92d866),
	}

	blueFaction = &faction{
		tag:   blueFactionTag,
		color: ge.RGB(0x7078db),
	}
)
