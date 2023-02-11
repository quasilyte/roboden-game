package staging

import (
	"fmt"
	"sort"

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

	pendingResources []*essenceSourceNode
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
	if g.world.IsTutorial() {
		g.placePlayers()
		g.placeSecondTutorialBase()
		g.placeResources(0.65)
		g.placeTutorialBoss()
		for _, colony := range g.world.colonies {
			colony.realRadius = 96
			colony.resources = 120
		}
	} else {
		resourceMultipliers := []float64{
			0.4,
			0.75,
			1, // Default
			1.25,
			1.6,
		}
		g.placePlayers()
		g.placeCreepBases()
		g.placeCreeps()
		g.placeResources(resourceMultipliers[g.world.options.Resources])
		g.placeBoss()
	}
}

func (g *levelGenerator) randomPos(sector gmath.Rect) gmath.Vec {
	return gmath.Vec{
		X: g.scene.Rand().FloatRange(sector.Min.X, sector.Max.X),
		Y: g.scene.Rand().FloatRange(sector.Min.Y, sector.Max.Y),
	}
}

func (g *levelGenerator) placeSecondTutorialBase() {
	g.createBase(g.playerSpawn.Add(gmath.Vec{X: -256, Y: 400}))
}

func (g *levelGenerator) placePlayers() {
	g.playerSpawn = g.world.rect.Center()
	g.createBase(g.playerSpawn)
}

func (g *levelGenerator) createBase(pos gmath.Vec) {
	core := g.world.NewColonyCoreNode(colonyConfig{
		World:  g.world,
		Radius: 128,
		Pos:    pos,
	})
	core.actionPriorities.SetWeight(priorityResources, 0.5)
	core.actionPriorities.SetWeight(priorityGrowth, 0.4)
	core.actionPriorities.SetWeight(prioritySecurity, 0.1)
	g.scene.AddObject(core)

	for i := 0; i < 5; i++ {
		a := core.NewColonyAgentNode(workerAgentStats, core.pos.Add(g.scene.Rand().Offset(-20, 20)))
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
		if !posIsFree(g.world, nil, pos, 24) || pos.DistanceTo(g.playerSpawn) < 520 {
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
	pos := correctedPos(sector, g.randomPos(sector), 196)
	initialPos := pos
	for i := 0; i < maxSize; i++ {
		if !posIsFree(g.world, nil, pos, 8) {
			break
		}
		source := g.world.NewEssenceSourceNode(kind, pos)
		g.pendingResources = append(g.pendingResources, source)
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

func (g *levelGenerator) placeResources(resMultiplier float64) {
	rand := g.scene.Rand()

	worldSizeMultipliers := []float64{
		0.8,
		0.9,
		1.0,
	}
	multiplier := resMultiplier * worldSizeMultipliers[g.world.worldSize]
	numIron := int(float64(rand.IntRange(26, 38)) * multiplier)
	numScrap := int(float64(rand.IntRange(8, 10)) * multiplier)
	numGold := int(float64(rand.IntRange(20, 28)) * multiplier)
	numCrystals := int(float64(rand.IntRange(14, 20)) * multiplier)
	numOil := int(float64(rand.IntRange(4, 6)) * multiplier)
	numRedOil := int(float64(rand.IntRange(2, 3)) * multiplier)

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
		if rand.Chance(0.4) {
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
			for i := 0; i < 2; i++ {
				for j := 0; j < 5; j++ {
					pos := gmath.RadToVec(rand.Rad()).Mulf(80).Add(core.pos)
					if !posIsFree(g.world, nil, pos, 14) {
						continue
					}
					essence := g.world.NewEssenceSourceNode(ironSource, pos)
					g.pendingResources = append(g.pendingResources, essence)
					break
				}
			}
		}
	}

	// Now sort all resources by their Y coordinate and only
	// then add them to the scene.
	sort.Slice(g.pendingResources, func(i, j int) bool {
		return g.pendingResources[i].pos.Y < g.pendingResources[j].pos.Y
	})
	for _, source := range g.pendingResources {
		g.scene.AddObject(source)
	}
}

func (g *levelGenerator) placeTutorialBoss() {
	boss := g.world.NewCreepNode(gmath.Vec{X: 256, Y: 256}, uberBossCreepStats)
	g.scene.AddObject(boss)

	boss.OnDamage(damageValue{health: uberBossCreepStats.maxHealth * 0.33}, gmath.Vec{})

	g.world.boss = boss
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

	g.world.boss = boss
}

func (g *levelGenerator) placeCreeps() {
	rand := g.scene.Rand()

	g.sectorSlider.TrySetValue(rand.IntRange(0, len(g.sectors)-1))

	numTurrets := rand.IntRange(4, 5)
	for numTurrets > 0 {
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numTurrets -= g.placeCreepsCluster(sector, 1, turretCreepStats)
	}

	numTanks := rand.IntRange(10, 15)
	for numTanks > 0 {
		sector := g.sectors[g.sectorSlider.Value()]
		g.sectorSlider.Inc()
		numTanks -= g.placeCreepsCluster(sector, 1, tankCreepStats)
	}
}

func (g *levelGenerator) placeCreepBases() {
	if g.world.options.Difficulty == 0 {
		return // Zero bases
	}
	// The bases are always located somewhere on the map boundary.
	// Bases can't be on the same border.
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
	numBases := g.world.options.Difficulty
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
		base := g.world.NewCreepNode(basePos, baseCreepStats)
		if i == 0 {
			base.specialDelay = (9 * 60.0) * g.scene.Rand().FloatRange(0.9, 1.1)
		} else {
			base.specialModifier = 1.0 // Initial level base
			base.specialDelay = g.scene.Rand().FloatRange(60, 120)
			base.attackDelay = g.scene.Rand().FloatRange(40, 50)
		}
		g.scene.AddObject(base)
	}
}
