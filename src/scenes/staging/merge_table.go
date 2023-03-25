package staging

import (
	"github.com/quasilyte/roboden-game/gamedata"
)

type mergeTableBits uint64

type mergeTable struct {
	flags mergeTableBits
}

func (t *mergeTable) Update(units *colonyAgentContainer) {
	t.flags = 0
	units.Each(func(a *colonyAgentNode) {
		if a.faction == gamedata.NeutralFactionTag || a.stats.Tier != 1 {
			return
		}
		t.flags |= t.droneToBit(a.faction, a.stats.Kind)
	})
}

func (t *mergeTable) droneToBit(faction gamedata.FactionTag, kind gamedata.ColonyAgentKind) mergeTableBits {
	bitpos := uint64(faction-1) * 2
	if kind == gamedata.AgentMilitia {
		bitpos++
	}
	return 1 << bitpos
}

func (t *mergeTable) CanProduce(recipe gamedata.AgentMergeRecipe) bool {
	return (t.flags&t.droneToBit(recipe.Drone1.Faction, recipe.Drone1.Kind) != 0) &&
		(t.flags&t.droneToBit(recipe.Drone2.Faction, recipe.Drone2.Kind) != 0)
}
