package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type classicManager struct {
	world *worldState

	spawnDelayMultiplier float64

	scene *ge.Scene

	tier3spawnDelay float64
	tier3spawnRate  float64

	crawlersDelay float64

	coordinatorsDelay float64

	attackGroup arenaWaveGroup
}

func newClassicManager(world *worldState) *classicManager {
	return &classicManager{
		world:          world,
		tier3spawnRate: 1,
	}
}

func (m *classicManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.spawnDelayMultiplier = 1.0 / (0.75 + (0.25 * float64(m.world.config.CreepSpawnRate)))

	// 1.1, 1.0, 0.9, 0.8, 0.7, 0.6
	firstSpawnDelayMultiplier := 1.1 - (0.1 * float64(m.world.config.CreepSpawnRate))

	// Start launching tier3 creeps after ~15 minutes.
	m.tier3spawnDelay = m.world.rand.FloatRange(15*60.0, 18*60.0) * firstSpawnDelayMultiplier

	// Extra crawlers show up around the 10th minute.
	m.crawlersDelay = m.world.rand.FloatRange(10*60.0, 14*60.0) * firstSpawnDelayMultiplier

	m.coordinatorsDelay = m.world.rand.FloatRange(5, 20)
}

func (m *classicManager) IsDisposed() bool {
	return false
}

func (m *classicManager) Update(delta float64) {
	m.tier3spawnDelay = gmath.ClampMin(m.tier3spawnDelay-delta, 0)
	if m.tier3spawnDelay == 0 {
		m.spawnTier3Creep()
	}
	m.crawlersDelay = gmath.ClampMin(m.crawlersDelay-delta, 0)
	if m.crawlersDelay == 0 {
		m.spawnCrawlers()
	}
	m.coordinatorsDelay = gmath.ClampMin(m.coordinatorsDelay-delta, 0)
	if m.coordinatorsDelay == 0 {
		m.maybeSendCoordinators()
	}
}

func (m *classicManager) maybeSendCoordinators() {
	if m.world.boss == nil {
		m.coordinatorsDelay = 999999 // ~never
		return
	}

	if !m.world.AllCenturionsReady() {
		m.coordinatorsDelay = 5
		return
	}

	stationed := m.world.centurionRallyPointPtr == &m.world.boss.pos
	if stationed {
		if len(m.world.centurions) == 0 {
			m.coordinatorsDelay = m.world.rand.FloatRange(30, 60)
			return
		}
		if len(m.world.centurions) < 3 {
			m.coordinatorsDelay = m.world.rand.FloatRange(20, 40)
			return
		}
	} else {
		if len(m.world.centurions) < 3 {
			m.world.centurionRallyPointPtr = &m.world.boss.pos
			m.coordinatorsDelay = m.world.rand.FloatRange(60, 100)
			return
		}
	}

	bossHealthPercentage := m.world.boss.health / m.world.boss.maxHealth
	if stationed {
		// Keep Coordinators around if it's wounded.
		if bossHealthPercentage < 0.8 {
			m.coordinatorsDelay = m.world.rand.FloatRange(10, 15)
			return
		}
		if m.world.rand.Chance(0.4) {
			// Send Coordinators out for some time.
			m.coordinatorsDelay = m.world.rand.FloatRange(30, 50)
			m.world.centurionRallyPoint = correctedPos(m.world.rect, m.world.boss.pos.Add(m.world.rand.Offset(-300, 300)), 64)
			m.world.centurionRallyPointPtr = &m.world.centurionRallyPoint
			return
		}
		// Keep them around for a bit longer
		m.coordinatorsDelay = m.world.rand.FloatRange(20, 30)
		return
	}

	// Send them back.
	m.world.centurionRallyPointPtr = &m.world.boss.pos
	m.coordinatorsDelay = m.world.rand.FloatRange(50, 80)
}

func (m *classicManager) spawnCrawlers() {
	units := m.attackGroup.units[:0]

	nextAttackDelay := 0.0
	if m.world.rand.Chance(0.75) {
		nextAttackDelay = m.world.rand.FloatRange(80, 140)
		numCreeps := m.world.rand.IntRange(1, 5) + m.world.config.CreepSpawnRate
		for i := 0; i < numCreeps; i++ {
			units = append(units, arenaWaveUnit{
				stats: gamedata.StealthCrawlerCreepStats,
				super: m.world.config.SuperCreeps && m.world.rand.Chance(0.3),
			})
		}
	} else {
		nextAttackDelay = m.world.rand.FloatRange(210, 250)
		units = append(units, arenaWaveUnit{
			stats: gamedata.HowitzerCreepStats,
			super: m.world.config.SuperCreeps && m.world.rand.Chance(0.3),
		})
	}

	m.crawlersDelay = nextAttackDelay * m.spawnDelayMultiplier

	m.attackGroup.units = units
	sendCreeps(m.world, m.attackGroup)
}

func (m *classicManager) spawnTier3Creep() {
	superChance := (1.0 - m.tier3spawnRate) * 0.5
	m.tier3spawnRate = gmath.ClampMin(m.tier3spawnRate-0.02, 0.35)
	m.tier3spawnDelay = (m.world.rand.FloatRange(60, 90) * m.tier3spawnRate) * m.spawnDelayMultiplier

	var spawnPos gmath.Vec
	roll := m.world.rand.Float()
	if roll < 0.25 {
		spawnPos.X = m.world.width - 4
		spawnPos.Y = m.world.rand.FloatRange(0, m.world.height)
	} else if roll < 0.5 {
		spawnPos.X = m.world.rand.FloatRange(0, m.world.width)
		spawnPos.Y = m.world.height - 4
	} else if roll < 0.75 {
		spawnPos.X = 4
		spawnPos.Y = m.world.rand.FloatRange(0, m.world.height)
	} else {
		spawnPos.X = m.world.rand.FloatRange(0, m.world.width)
		spawnPos.Y = 4
	}
	spawnPos = roundedPos(spawnPos)
	stats := gamedata.AssaultCreepStats
	if m.world.rand.Chance(0.3) {
		stats = gamedata.BuilderCreepStats
	}
	creep := m.world.NewCreepNode(spawnPos, stats)
	creep.super = m.world.config.SuperCreeps && m.world.rand.Chance(superChance)
	m.world.nodeRunner.AddObject(creep)
}
