package staging

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

func sendCreeps(world *worldState, g arenaWaveGroup) gmath.Vec {
	sector := world.spawnAreas[g.side]
	spawnPos := randomSectorPos(world.rand, sector)
	targetPos := correctedPos(world.rect, randomSectorPos(world.rand, sector), 520)

	spawnRect := gmath.Rect{
		Min: spawnPos.Sub(gmath.Vec{X: 96, Y: 96}),
		Max: spawnPos.Add(gmath.Vec{X: 96, Y: 96}),
	}
	spawnRect.Min.X = gmath.ClampMin(spawnRect.Min.X, sector.Min.X)
	spawnRect.Min.Y = gmath.ClampMin(spawnRect.Min.Y, sector.Min.Y)
	spawnRect.Max.X = gmath.ClampMax(spawnRect.Max.X, sector.Max.X)
	spawnRect.Max.Y = gmath.ClampMax(spawnRect.Max.Y, sector.Max.Y)

	for _, u := range g.units {
		var creepPos gmath.Vec
		spawnDelay := 0.0
		if u.stats.ShadowImage == assets.ImageNone {
			creepPos, spawnDelay = groundCreepSpawnPos(world, spawnRect, u.stats)
			if creepPos.IsZero() {
				continue
			}
		} else {
			creepPos = spawnPos.Add(world.rand.Offset(-60, 60))
		}

		creepTargetPos := targetPos.Add(world.rand.Offset(-64, 64))
		if spawnDelay > 0 {
			spawner := newCreepSpawnerNode(world, spawnDelay, creepPos, creepTargetPos, u.stats)
			spawner.super = u.super
			spawner.fragScore = u.fragScore
			world.nodeRunner.AddObject(spawner)
		} else {
			creep := world.NewCreepNode(creepPos, u.stats)
			creep.super = u.super
			creep.fragScore = u.fragScore
			world.nodeRunner.AddObject(creep)
			creep.SendTo(creepTargetPos)
		}
	}

	return spawnPos
}

func groundCreepSpawnPos(world *worldState, spawnRect gmath.Rect, stats *gamedata.CreepStats) (gmath.Vec, float64) {
	for i := 0; i < 6; i++ {
		spawnDelay := 0.0
		attemptPos := randomSectorPos(world.rand, spawnRect)
		if attemptPos.X <= 0 {
			spawnDelay += (-attemptPos.X) / stats.Speed
			attemptPos.X = 1
		} else if attemptPos.X >= world.width {
			spawnDelay += (attemptPos.X - world.width) / stats.Speed
			attemptPos.X = world.width - 1
		}
		if attemptPos.Y <= 0 {
			spawnDelay += (-attemptPos.Y) / stats.Speed
			attemptPos.Y = 1
		} else if attemptPos.Y >= world.height {
			spawnDelay += (attemptPos.Y - world.height) / stats.Speed
			attemptPos.Y = world.height - 1
		}
		coord := world.pathgrid.PosToCoord(attemptPos)
		if !world.CellIsFree(coord, layerNormal) {
			continue
		}

		return attemptPos, spawnDelay
	}

	return gmath.Vec{}, 0
}
