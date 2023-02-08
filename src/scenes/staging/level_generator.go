package staging

import (
	"fmt"

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
	g.placeCreepBases()
	g.placeCreeps()
	g.placeResources()
	g.placeBoss()
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
		a := core.NewColonyAgentNode(workerAgentStats, core.pos.Add(g.scene.Rand().Offset(-20, 20)))
		// a.faction = blueFactionTag
		g.scene.AddObject(a)
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (g *levelGenerator) placeCreepsCluster(sector gmath.Rect, maxSize int, kind *creepStats) int {
	rand := g.scene.Rand()
	placed := 0
	pos := correctedPos(sector, g.randomPos(sector), 128)
	initialPos := pos
	unitPos := pos
	for i := 0; i < maxSize; i++ {
		if !posIsFree(g.world, nil, pos, 28) || pos.DistanceTo(g.playerSpawn) < 320 {
			break
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
			break
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

	numIron := rand.IntRange(26, 38)
	numScrap := rand.IntRange(8, 10)
	numGold := rand.IntRange(18, 28)
	numCrystals := rand.IntRange(12, 18)
	numOil := rand.IntRange(4, 6)
	numRedOil := rand.IntRange(2, 3)

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
	for numRedOil > 0 {
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numRedOil -= g.placeResourceCluster(sector, 1, redOilSource)
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
			return source.pos.DistanceTo(core.pos) <= core.realRadius &&
				source.stats != scrapSource &&
				source.stats != smallScrapSource
		})
		if !hasResources {
			pos := gmath.RadToVec(rand.Rad()).Mulf(80).Add(core.pos)
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

func (g *levelGenerator) placeCreepBases() {
	// Right now there are always 2 creep bases.
	// One of them will activate later than another.
	// The bases are always located somewhere on the map boundary.
	// Both bases can't be on the same border.
	pad := 128.0
	borderWidth := 440.0
	borders := []gmath.Rect{
		// top border
		{Min: gmath.Vec{X: pad, Y: pad}, Max: gmath.Vec{X: g.world.width - pad, Y: borderWidth + pad}},
		// right border
		{Min: gmath.Vec{X: g.world.width - borderWidth - pad, Y: pad}, Max: gmath.Vec{X: g.world.width - pad, Y: g.world.height - pad}},
		// left border
		{Min: gmath.Vec{X: pad, Y: pad}, Max: gmath.Vec{X: borderWidth + pad, Y: g.world.height - pad}},
		// bottom border
		{Min: gmath.Vec{X: pad, Y: g.world.height - borderWidth - pad}, Max: gmath.Vec{X: g.world.width - pad, Y: g.world.height - pad}},
	}
	gmath.Shuffle(g.scene.Rand(), borders)
	numBases := 2
	for i := 0; i < numBases; i++ {
		border := borders[i]
		var basePos gmath.Vec
		for round := 0; round < 32; round++ {
			posProbe := g.randomPos(border)
			if posIsFree(g.world, nil, posProbe, 48) {
				basePos = posProbe
				break
			}
		}
		if basePos.IsZero() {
			fmt.Println("couldn't deploy creep base", i+1)
			continue
		}
		fmt.Println("deployed a creep base", i+1, "at", basePos, "distance is", basePos.DistanceTo(g.playerSpawn))
		// Base 1: starts right away with level 2.
		// Base 2: has 9 minutes delay before level 1.
		// The first creep
		base := g.world.NewCreepNode(basePos, baseCreepStats)
		if i == 0 {
			base.specialModifier = 1.0
			base.specialDelay = g.scene.Rand().FloatRange(60, 120)
			base.attackDelay = g.scene.Rand().FloatRange(40, 50)
		} else {
			base.specialDelay = (9 * 60.0) * g.scene.Rand().FloatRange(0.9, 1.1)
		}
		g.scene.AddObject(base)
	}
}
