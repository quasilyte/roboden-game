package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
)

type levelGenerator struct {
	scene *ge.Scene
	world *worldState
}

func newLevelGenerator(scene *ge.Scene, world *worldState) *levelGenerator {
	return &levelGenerator{
		scene: scene,
		world: world,
	}
}

func (g *levelGenerator) Generate() {
	g.placePlayers()
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
	core := g.world.NewColonyCoreNode(colonyConfig{
		World:  g.world,
		Radius: 128,
		Pos:    g.world.rect.Center(),
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

	numIron := rand.IntRange(14, 20)
	numScrap := rand.IntRange(4, 8)
	numGold := rand.IntRange(6, 12)
	numCrystals := rand.IntRange(6, 12)
	numOil := rand.IntRange(5, 8)

	sectors := []gmath.Rect{
		{Min: gmath.Vec{X: 0, Y: 0}, Max: gmath.Vec{X: g.world.width / 2, Y: g.world.height / 2}},
		{Min: gmath.Vec{X: g.world.width / 2, Y: 0}, Max: gmath.Vec{X: g.world.width, Y: g.world.height / 2}},
		{Min: gmath.Vec{X: 0, Y: g.world.height / 2}, Max: gmath.Vec{X: g.world.width / 2, Y: g.world.height}},
		{Min: gmath.Vec{X: g.world.width / 2, Y: g.world.height / 2}, Max: gmath.Vec{X: g.world.width, Y: g.world.height}},
	}
	var sectorSlider gmath.Slider
	sectorSlider.SetBounds(0, len(sectors)-1)
	sectorSlider.TrySetValue(rand.IntRange(0, len(sectors)-1))

	for numIron > 0 {
		clusterSize := rand.IntRange(2, 6)
		sector := sectors[sectorSlider.Value()]
		sectorSlider.Inc()
		numIron -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numIron), ironSource)
	}
	for numScrap > 0 {
		sector := sectors[sectorSlider.Value()]
		sectorSlider.Inc()
		numScrap -= g.placeResourceCluster(sector, 1, scrapSource)
	}
	for numGold > 0 {
		clusterSize := rand.IntRange(1, 3)
		sector := sectors[sectorSlider.Value()]
		sectorSlider.Inc()
		numGold -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numGold), goldSource)
	}
	for numOil > 0 {
		sector := sectors[sectorSlider.Value()]
		sectorSlider.Inc()
		numOil -= g.placeResourceCluster(sector, 1, oilSource)
	}
	for numCrystals > 0 {
		clusterSize := 1
		if rand.Chance(0.3) {
			clusterSize = 2
		}
		sector := sectors[sectorSlider.Value()]
		sectorSlider.Inc()
		numCrystals -= g.placeResourceCluster(sector, gmath.ClampMax(clusterSize, numCrystals), crystalSource)
	}

	// If there are no resources near the colony spawn pos,
	// place something in there.
	for _, core := range g.world.colonies {
		hasResources := xslices.ContainsWhere(g.world.essenceSources, func(source *essenceSourceNode) bool {
			return source.pos.DistanceTo(core.body.Pos) <= core.radius
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
