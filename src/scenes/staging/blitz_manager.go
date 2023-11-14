package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type blitzManager struct {
	world *worldState

	grenadiersDelay float64
	grenadierWave   int

	attackGroup arenaWaveGroup

	scene *ge.Scene
}

func newBlitzManager(world *worldState) *blitzManager {
	return &blitzManager{
		world: world,
	}
}

func (m *blitzManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.grenadiersDelay = gamedata.BlitzModeSetupTime(m.world.numPlayers) + m.world.rand.FloatRange(1*60.0, 2*60.0)
	m.grenadierWave = 2
}

func (m *blitzManager) IsDisposed() bool {
	return false
}

func (m *blitzManager) Update(delta float64) {
	m.grenadiersDelay = gmath.ClampMin(m.grenadiersDelay-delta, 0)
	if m.grenadiersDelay == 0 {
		m.spawnGrenadiers()
	}
}

func (m *blitzManager) spawnGrenadiers() {
	if !m.world.config.GrenadierCreeps {
		m.grenadiersDelay = 99999999 // ~never
		return
	}

	minGrenadiers := gmath.ClampMax(2+m.grenadierWave, 8)
	maxGrenadiers := gmath.ClampMax(3+m.grenadierWave, 12)
	numGrenadiers := m.world.rand.IntRange(minGrenadiers, maxGrenadiers)
	maxSupers := gmath.ClampMax(1+(m.grenadierWave/3), 4)

	units := m.attackGroup.units[:0]
	for i := 0; i < numGrenadiers; i++ {
		super := m.world.config.SuperCreeps && i < maxSupers && m.grenadierWave > 0
		units = append(units, arenaWaveUnit{
			stats: gamedata.GrenadierCreepStats,
			super: super,
		})
	}

	m.grenadiersDelay = m.world.rand.FloatRange(100, 230)
	if m.grenadiersDelay >= 160 {
		m.grenadierWave++
	}

	m.attackGroup.units = units
	m.attackGroup.side = m.world.rand.IntRange(0, 3)
	sendCreeps(m.world, m.attackGroup)
}

func (m *blitzManager) SpawnInitialCreeps() {
	numBuilders := 5
	numSupers := 0
	if m.world.config.SuperCreeps {
		numSupers = 2
	}
	for i := 0; i < numBuilders; i++ {
		var g arenaWaveGroup
		g.side = m.world.rand.IntRange(0, 3)
		isSuper := i < numSupers
		g.units = []arenaWaveUnit{
			{stats: gamedata.BuilderCreepStats, super: isSuper},
		}
		switch i {
		case 0:
			g.units = append(g.units,
				arenaWaveUnit{stats: gamedata.AssaultCreepStats, super: isSuper},
				arenaWaveUnit{stats: gamedata.AssaultCreepStats, super: isSuper},
			)
		case 1:
			for j := 0; j < 4; j++ {
				g.units = append(g.units, arenaWaveUnit{stats: gamedata.HeavyCrawlerCreepStats})
			}
		}
		sendCreeps(m.world, g)
	}
}
