package staging

import (
	"github.com/quasilyte/ge"
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
	g.placeResources()
}

func (g *levelGenerator) randomPos(sector gmath.Rect) gmath.Vec {
	return gmath.Vec{
		X: g.scene.Rand().FloatRange(sector.Min.X, sector.Max.X),
		Y: g.scene.Rand().FloatRange(sector.Min.Y, sector.Max.Y),
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

	// type resourceSpawn struct {
	// 	pos   gmath.Vec
	// 	stats *essenceSourceStats
	// }
	// resourceLocations := []resourceSpawn{
	// 	{stats: scrapSource, pos: gmath.Vec{X: 1020, Y: 640}},
	// 	{stats: scrapSource, pos: gmath.Vec{X: 150, Y: 300}},
	// 	{stats: ironSource, pos: gmath.Vec{X: 580, Y: 400}},
	// 	{stats: goldSource, pos: gmath.Vec{X: 290, Y: 470}},
	// 	{stats: oilSource, pos: gmath.Vec{X: 160, Y: 210}},
	// 	{stats: crystalSource, pos: gmath.Vec{X: 90, Y: 270}},
	// 	{stats: goldSource, pos: gmath.Vec{X: 890, Y: 500}},
	// 	{stats: oilSource, pos: gmath.Vec{X: 1050, Y: 350}},
	// 	{stats: ironSource, pos: gmath.Vec{X: 670, Y: 800}},
	// 	{stats: oilSource, pos: gmath.Vec{X: 460, Y: 760}},
	// 	{stats: crystalSource, pos: gmath.Vec{X: 300, Y: 800}},
	// 	{stats: goldSource, pos: gmath.Vec{X: 1600, Y: 900}},
	// 	{stats: goldSource, pos: gmath.Vec{X: 1550, Y: 860}},
	// 	{stats: goldSource, pos: gmath.Vec{X: 1300, Y: 920}},
	// 	{stats: crystalSource, pos: gmath.Vec{X: 1300, Y: 160}},
	// 	{stats: crystalSource, pos: gmath.Vec{X: 1360, Y: 196}},
	// 	{stats: crystalSource, pos: gmath.Vec{X: 1296, Y: 180}},
	// 	{stats: ironSource, pos: gmath.Vec{X: 1190, Y: 400}},
	// 	{stats: ironSource, pos: gmath.Vec{X: 1160, Y: 460}},
	// 	{stats: crystalSource, pos: gmath.Vec{X: 100, Y: 890}},
	// }
	// for _, spawn := range resourceLocations {
	// 	e := g.world.NewEssenceSourceNode(spawn.stats, spawn.pos)
	// 	g.scene.AddObject(e)
	// }
}
