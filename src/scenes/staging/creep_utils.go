package staging

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

func sendCreeps(world *worldState, g arenaWaveGroup) {
	sector := world.spawnAreas[g.side]
	spawnPos := randomSectorPos(world.rand, sector)
	targetPos := correctedPos(world.rect, randomSectorPos(world.rand, sector), 520)

	for _, u := range g.units {
		creepPos := spawnPos
		spawnDelay := 0.0
		if u.stats.shadowImage == assets.ImageNone {
			creepPos, spawnDelay = groundCreepSpawnPos(world, creepPos, u.stats)
			if creepPos.IsZero() {
				continue
			}
		} else {
			creepPos = creepPos.Add(world.rand.Offset(-60, 60))
		}

		creepTargetPos := targetPos.Add(world.rand.Offset(-60, 60))
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
}

func groundCreepSpawnPos(world *worldState, pos gmath.Vec, stats *creepStats) (gmath.Vec, float64) {
	creepPos := pos
	spawnDelay := 0.0
	attemptPos := creepPos.Add(world.rand.Offset(-60, 60))
	for i := 0; i < 4; i++ {
		if attemptPos.X <= 0 {
			spawnDelay = (-attemptPos.X) / stats.speed
			attemptPos.X = 1
		} else if attemptPos.X >= world.width {
			spawnDelay = (attemptPos.X - world.width) / stats.speed
			attemptPos.X = world.width - 1
		}
		if attemptPos.Y <= 0 {
			spawnDelay = (-attemptPos.Y) / stats.speed
			attemptPos.Y = 1
		} else if attemptPos.Y >= world.height {
			spawnDelay = (attemptPos.Y - world.height) / stats.speed
			attemptPos.Y = world.height - 1
		}
		coord := world.pathgrid.PosToCoord(attemptPos)
		if world.pathgrid.CellIsFree(coord) {
			creepPos = attemptPos
			break
		}
	}
	return creepPos, spawnDelay
}
