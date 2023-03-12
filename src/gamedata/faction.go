package gamedata

import (
	"image/color"

	"github.com/quasilyte/ge"
)

type faction struct {
	Tag   FactionTag
	Color color.RGBA
}

//go:generate stringer -type=FactionTag
type FactionTag int

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
	NeutralFactionTag FactionTag = iota
	YellowFactionTag
	RedFactionTag
	GreenFactionTag
	BlueFactionTag
)

func FactionByTag(tag FactionTag) *faction {
	switch tag {
	case YellowFactionTag:
		return yellowFaction
	case RedFactionTag:
		return redFaction
	case GreenFactionTag:
		return greenFaction
	case BlueFactionTag:
		return blueFaction
	default:
		return nil
	}
}

var (
	yellowFaction = &faction{
		Tag:   YellowFactionTag,
		Color: ge.RGB(0xf1e851),
	}

	redFaction = &faction{
		Tag:   RedFactionTag,
		Color: ge.RGB(0xef6a57),
	}

	greenFaction = &faction{
		Tag:   GreenFactionTag,
		Color: ge.RGB(0x92d866),
	}

	blueFaction = &faction{
		Tag:   BlueFactionTag,
		Color: ge.RGB(0x7078db),
	}
)
