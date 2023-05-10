package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type classicManager struct {
	world *worldState

	spawnAreas []gmath.Rect

	spawnDelayMultiplier float64

	scene *ge.Scene

	tier3spawnDelay float64
	tier3spawnRate  float64

	crawlersDelay float64
}

func newClassicManager(world *worldState) *classicManager {
	return &classicManager{
		world:          world,
		tier3spawnRate: 1,
	}
}

func (m *classicManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.spawnAreas = creepSpawnAreas(m.world)

	m.spawnDelayMultiplier = 0.75 + (0.25 * float64(m.world.config.CreepSpawnRate))

	// 1.1, 1.0, 0.9, 0.8
	firstSpawnDelayMultiplier := 1.1 - (0.1 * float64(m.world.config.CreepSpawnRate))

	// Start launching tier3 creeps after ~15 minutes.
	m.tier3spawnDelay = m.world.rand.FloatRange(15*60.0, 18*60.0) * firstSpawnDelayMultiplier

	// Extra crawlers show up around the 10th minute.
	m.crawlersDelay = m.world.rand.FloatRange(10*60.0, 14*60.0) * firstSpawnDelayMultiplier
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
}

func (m *classicManager) spawnCrawlers() {
	nextAttackDelay := 0.0
	numCreeps := 1
	creepStats := howitzerCreepStats
	if m.world.rand.Chance(0.75) {
		nextAttackDelay = m.world.rand.FloatRange(80, 140) * m.spawnDelayMultiplier
		numCreeps = m.world.rand.IntRange(1, 5) + m.world.config.CreepSpawnRate
		creepStats = stealthCrawlerCreepStats
	} else {
		nextAttackDelay = m.world.rand.FloatRange(210, 250) * m.spawnDelayMultiplier
	}
	m.crawlersDelay = nextAttackDelay

	sector := gmath.RandElem(m.world.rand, m.spawnAreas)
	spawnPos := randomSectorPos(m.world.rand, sector)
	targetPos := correctedPos(m.world.rect, randomSectorPos(m.world.rand, sector), 520)

	for i := 0; i < numCreeps; i++ {
		super := m.world.config.SuperCreeps && m.world.rand.Chance(0.3)
		creepPos, spawnDelay := groundCreepSpawnPos(m.world, spawnPos, creepStats)
		creepTargetPos := targetPos.Add(m.world.rand.Offset(-60, 60))
		if spawnDelay > 0 {
			spawner := newCreepSpawnerNode(m.world, spawnDelay, creepPos, creepTargetPos, creepStats)
			spawner.super = super
			m.world.nodeRunner.AddObject(spawner)
		} else {
			creep := m.world.NewCreepNode(creepPos, creepStats)
			creep.super = super
			m.world.nodeRunner.AddObject(creep)
			creep.SendTo(creepTargetPos)
		}
	}
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
	stats := assaultCreepStats
	if m.world.rand.Chance(0.3) {
		stats = builderCreepStats
	}
	creep := m.world.NewCreepNode(spawnPos, stats)
	creep.super = m.world.config.SuperCreeps && m.world.rand.Chance(superChance)
	m.world.nodeRunner.AddObject(creep)
}
