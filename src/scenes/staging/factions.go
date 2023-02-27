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

// Yellow faction (miners):
// +1 max payload (basically doubles the worker capacity as it's 1 by default)
//
// Red faction (warriors):
// +40% max hp
//
// Green faction (engineers):
// +20% movement speed
// +50% faster building construction
// +50% more efficient building repair (more hp restored)
//
// Blue faction (scientists):
// +80% max energy
// +20% evo points income

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
