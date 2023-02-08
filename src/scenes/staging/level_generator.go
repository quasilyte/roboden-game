package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
)

type levelGenerator struct {
	scene        *ge.Scene
	world        *worldState
	playerSpawn  gmath.Vec
	sectors      []gmath.Rect
	sectorSlider gmath.Slider
}

func newLevelGenerator(scene *ge.Scene, world *worldState) *levelGenerator {
	g := &levelGenerator{
		scene: scene,
		world: world,
	}
	g.sectors = []gmath.Rect{
		{Min: gmath.Vec{X: 0, Y: 0}, Max: gmath.Vec{X: g.world.width / 2, Y: g.world.height / 2}},
		{Min: gmath.Vec{X: g.world.width / 2, Y: 0}, Max: gmath.Vec{X: g.world.width, Y: g.world.height / 2}},
		{Min: gmath.Vec{X: 0, Y: g.world.height / 2}, Max: gmath.Vec{X: g.world.width / 2, Y: g.world.height}},
		{Min: gmath.Vec{X: g.world.width / 2, Y: g.world.height / 2}, Max: gmath.Vec{X: g.world.width, Y: g.world.height}},
	}
	g.sectorSlider.SetBounds(0, len(g.sectors)-1)
	return g
}

func (g *levelGenerator) Generate() {
	g.placePlayers()
	g.placeResources()
	g.placeBoss()
	g.placeCreeps()
}

func (g *levelGenerator) randomPos(sector gmath.Rect) gmath.Vec {
	return gmath.Vec{
		X: g.scene.Rand().FloatRange(sector.Min.X, sector.Max.X),
		Y: g.scene.Rand().FloatRange(sector.Min.Y, sector.Max.Y),
	}
}

func (g *levelGenerator) placePlayers() {
	g.playerSpawn = g.world.rect.Center()
	core := g.world.NewColonyCoreNode(colonyConfig{
		World:  g.world,
		Radius: 128,
		Pos:    g.playerSpawn,
	})
	core.actionPriorities.SetWeight(priorityResources, 0.6)
	core.actionPriorities.SetWeight(priorityGrowth, 0.3)
	core.actionPriorities.SetWeight(prioritySecurity, 0.1)
	g.scene.AddObject(core)

	for i := 0; i < 5; i++ {
		a := core.NewColonyAgentNode(workerAgentStats, core.body.Pos.Add(g.scene.Rand().Offset(-20, 20)))
		// a.faction = blueFactionTag
		g.scene.AddObject(a)
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (g *levelGenerator) placeCreepsCluster(sector gmath.Rect, maxSize int, kind *creepStats) int {
	rand := g.scene.Rand()
	placed := 0
	pos := correctedPos(sector, g.randomPos(sector), 128)
	if pos.DistanceTo(g.playerSpawn) < 256 {
		return 0
	}
	initialPos := pos
	unitPos := pos
	for i := 0; i < maxSize; i++ {
		if !posIsFree(g.world, nil, pos, 28) {
			continue
		}
		creep := g.world.NewCreepNode(pos, kind)
		g.scene.AddObject(creep)
		unitPos = pos
		direction := gmath.RadToVec(rand.Rad()).Mulf(32)
		if rand.Bool() {
			pos = initialPos.Add(direction)
		} else {
			pos = pos.Add(direction)
		}
		placed++
	}
	// Creep groups may have some scraps near them.
	if placed != 0 && rand.Chance(0.7) {
		numScraps := rand.IntRange(1, 2)
		for i := 0; i < numScraps; i++ {
			scrapPos := gmath.RadToVec(rand.Rad()).Mulf(rand.FloatRange(64, 128)).Add(unitPos)
			if posIsFree(g.world, nil, scrapPos, 8) {
				source := g.world.NewEssenceSourceNode(scrapSource, scrapPos)
				g.scene.AddObject(source)
			}
		}
	}
	return placed
}

func (g *levelGenerator) placeResourceCluster(sector gmath.Rect, maxSize int, kind *essenceSourceStats) int {
	rand := g.scene.Rand()
	placed := 0
	pos := correctedPos(sector, g.randomPos(sector), 64)
	initialPos := pos
	for i := 0; i < maxSize; i++ {
		if !posIsFree(g.world, nil, pos, 8) {
			continue
		}
		source := g.world.NewEssenceSourceNode(kind, pos)
		g.scene.AddObject(source)
		direction := gmath.RadToVec(rand.Rad()).Mulf(32)
		if rand.Bool() {
			pos = initialPos.Add(direction)
		} else {
			pos = pos.Add(direction)
		}
		placed++
	}
	return placed
}

func (g *levelGenerator) placeResources() {
	rand := g.scene.Rand()

	numIron := rand.IntRange(18, 26)
	numScrap := rand.IntRange(8, 10)
	numGold := rand.IntRange(10, 16)
	numCrystals := rand.IntRange(8, 14)
	numOil := rand.IntRange(6, 9)

	g.sectorSlider.TrySetValue(rand.IntRange(0, len(g.sectors)-1))

	for numIron > 0 {
		clusterSize := rand.IntRange(2, 6)
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numIron -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numIron), ironSource)
	}
	for numScrap > 0 {
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		clusterSize := 1
		if rand.Chance(0.3) {
			clusterSize = 2
		}
		kind := smallScrapSource
		roll := rand.Float()
		if roll > 0.8 {
			kind = bigScrapSource
		} else if roll > 0.4 {
			kind = scrapSource
		}
		numScrap -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numScrap), kind)
	}
	for numGold > 0 {
		clusterSize := rand.IntRange(1, 3)
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numGold -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numGold), goldSource)
	}
	for numOil > 0 {
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numOil -= g.placeResourceCluster(sector, 1, oilSource)
	}
	for numCrystals > 0 {
		clusterSize := 1
		if rand.Chance(0.3) {
			clusterSize = 2
		}
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numCrystals -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numCrystals), crystalSource)
	}

	// If there are no resources near the colony spawn pos,
	// place something in there.
	for _, core := range g.world.colonies {
		hasResources := xslices.ContainsWhere(g.world.essenceSources, func(source *essenceSourceNode) bool {
			// We don't count scraps as some viable starting resource.
			return source.pos.DistanceTo(core.body.Pos) <= core.realRadius &&
				source.stats != scrapSource &&
				source.stats != smallScrapSource
		})
		if !hasResources {
			pos := gmath.RadToVec(rand.Rad()).Mulf(80).Add(core.body.Pos)
			essence := g.world.NewEssenceSourceNode(ironSource, pos)
			g.scene.AddObject(essence)
		}
	}
}

func (g *levelGenerator) placeBoss() {
	spawnLocations := []gmath.Vec{
		{X: 196, Y: 196},
		{X: g.world.width - 196, Y: 196},
		{X: 196, Y: g.world.height - 196},
		{X: g.world.width - 196, Y: g.world.height - 196},
	}
	pos := gmath.RandElem(g.world.rand, spawnLocations)
	boss := g.world.NewCreepNode(pos, uberBossCreepStats)
	g.scene.AddObject(boss)
}

func (g *levelGenerator) placeCreeps() {
	rand := g.scene.Rand()

	g.sectorSlider.TrySetValue(rand.IntRange(0, len(g.sectors)-1))

	numTurrets := rand.IntRange(3, 4)
	for numTurrets > 0 {
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numTurrets -= g.placeCreepsCluster(sector, 1, turretCreepStats)
	}
}
