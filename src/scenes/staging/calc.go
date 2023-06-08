package staging

import (
	"math"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

func numCreepsPerCard(state *creepsPlayerState, info creepOptionInfo) int {
	if info.maxUnits == 1 {
		return 1
	}
	techLevel := state.techLevel
	if !info.stats.flying {
		techLevel += 0.2
	}
	extraTech := gmath.Clamp(techLevel-info.minTechLevel, 0, 1.0)
	if extraTech == 0 {
		return 1
	}
	numUnits := gmath.Clamp(1+int(float64(info.maxUnits)*extraTech), 1, info.maxUnits)
	return numUnits
}

func getTurretPower(stats *gamedata.AgentStats) int {
	switch stats {
	case gamedata.GunpointAgentStats:
		return 35
	case gamedata.BeamTowerAgentStats:
		return 45
	default:
		return 0
	}
}

func calcCreepPower(world *worldState, creep *creepNode) int {
	power := 2.5 * float64(creepFragScore(creep.stats))
	if creep.super {
		power *= float64(superCreepCostMultiplier(creep.stats))
	}
	if creep.stats.kind == creepUberBoss {
		power = float64(power) * world.bossHealthMultiplier
	} else {
		power = float64(power) * world.creepHealthMultiplier
	}
	power *= (creep.health / creep.maxHealth) + 0.2
	return int(power)
}

func calcPosDanger(world *worldState, pstate *playerState, pos gmath.Vec, r float64) (int, gmath.Vec) {
	total := 0
	highestDanger := 0
	var mostDangerousPos gmath.Vec
	world.WalkCreeps(pos, r, func(creep *creepNode) bool {
		danger := calcCreepPower(world, creep)
		if danger > highestDanger {
			highestDanger = danger
			mostDangerousPos = creep.pos
		}
		total += danger
		return false
	})
	dangerDecrease := 0
	rSqr := r * r
	if turretPower := getTurretPower(world.turretDesign); turretPower != 0 {
		for _, c := range pstate.colonies {
			for _, turret := range c.turrets {
				if turret.pos.DistanceSquaredTo(pos) < rSqr {
					dangerDecrease += turretPower
				}
			}
		}
	}
	if pstate.hasRoombas {
		for _, c := range pstate.colonies {
			for _, roomba := range c.roombas {
				if roomba.pos.DistanceSquaredTo(pos) < rSqr {
					dangerDecrease += int(gamedata.RoombaAgentStats.Cost)
				}
			}
		}
	}
	total = gmath.ClampMin(total-dangerDecrease, 0)
	return total, mostDangerousPos
}

func multipliedDamage(target targetable, weapon *gamedata.WeaponStats) gamedata.DamageValue {
	damage := weapon.Damage
	if damage.Health != 0 {
		damage.Health *= damageMultiplier(target, weapon)
	}
	return damage
}

func damageMultiplier(target targetable, weapon *gamedata.WeaponStats) float64 {
	if target.IsFlying() {
		return weapon.FlyingTargetDamageMult
	}
	return weapon.GroundTargetDamageMult
}

func superCreepCostMultiplier(stats *creepStats) int {
	switch stats.kind {
	case creepCrawler:
		return 3
	case creepTurret, creepBase, creepCrawlerBase, creepHowitzer, creepDominator, creepServant:
		return 5
	case creepUberBoss:
		return 2
	}
	return 4
}

func creepCost(stats *creepStats, super bool) int {
	fragScore := creepFragScore(stats)
	if super {
		fragScore *= superCreepCostMultiplier(stats)
	}
	return fragScore
}

func creepFragScore(stats *creepStats) int {
	switch stats {
	case crawlerCreepStats:
		return 4
	case eliteCrawlerCreepStats:
		return 6
	case stealthCrawlerCreepStats:
		return 7
	case heavyCrawlerCreepStats:
		return 8

	case wandererCreepStats:
		return 6
	case stunnerCreepStats:
		return 9
	case assaultCreepStats:
		return 15
	case builderCreepStats:
		return 30

	case turretCreepStats:
		return 20

	case servantCreepStats:
		return 30
	case dominatorCreepStats:
		return 60
	case howitzerCreepStats:
		return 85

	case uberBossCreepStats:
		return 200

	default:
		return 0
	}
}

func calcScore(world *worldState) int {
	switch world.config.GameMode {
	case gamedata.ModeInfArena:
		score := world.config.DifficultyScore * 7
		timePlayed := world.result.TimePlayed.Seconds()
		if timePlayed < 5*60 {
			return 0
		}
		timePlayed -= 5 * 60
		baselineTime := 60.0 * 60.0
		multiplier := timePlayed / baselineTime
		return int(math.Round(float64(score) * multiplier))

	case gamedata.ModeArena:
		score := world.config.DifficultyScore * 11
		crystalsCollected := gmath.Percentage(world.result.RedCrystalsCollected, world.numRedCrystals)
		score += crystalsCollected * 3
		var multiplier float64
		if world.result.CreepFragScore != 0 {
			multiplier = float64(world.result.CreepFragScore) / float64(world.result.CreepTotalValue)
		}
		return int(math.Round(float64(score) * multiplier))

	case gamedata.ModeReverse:
		score := world.config.DifficultyScore * 10
		if world.boss != nil {
			score += int((world.boss.health / world.boss.maxHealth) * 500.0)
		}
		multiplier := 1.0 - (0.000347222 * (world.result.TimePlayed.Seconds() / 5))
		if multiplier < 0 {
			multiplier = 0.001
		}
		return int(math.Round(float64(score) * multiplier))

	case gamedata.ModeClassic:
		score := world.config.DifficultyScore * 10
		crystalsCollected := gmath.Percentage(world.result.RedCrystalsCollected, world.numRedCrystals)
		score += crystalsCollected * 3
		multiplier := 1.0 - (0.000347222 * (world.result.TimePlayed.Seconds() / 5))
		if multiplier < 0 {
			multiplier = 0.001
		}
		return int(math.Round(float64(score) * multiplier))

	default:
		return 0
	}
}

func mergeAgents(world *worldState, x, y *colonyAgentNode) *gamedata.AgentStats {
	list := world.tier2recipes
	if x.stats.Tier == 2 {
		list = gamedata.Tier3agentMergeRecipes
	}
	for _, recipe := range list {
		if recipe.Match(x.AsRecipeSubject(), y.AsRecipeSubject()) {
			return recipe.Result
		}
	}
	return nil
}

func agentCloningEnergyCost() float64 {
	return 30.0
}

func agentCloningCost(core *colonyCoreNode, cloner, a *colonyAgentNode) float64 {
	multiplier := 0.85
	return a.stats.Cost * multiplier
}

func resourceScore(core *colonyCoreNode, source *essenceSourceNode) float64 {
	if source.stats.regenDelay != 0 && source.percengage < 0.15 {
		return 0
	}
	if core.failedResource == source {
		return 0
	}
	dist := core.pos.DistanceTo(source.pos)
	maxDist := 1.5 + (core.GetResourcePriority() * 0.5)
	if dist > core.realRadius*maxDist || source.resource == 0 {
		return 0
	}
	distScore := 8.0 - gmath.ClampMax(dist/120, 8.0)
	multiplier := 1.0 + (source.stats.value * 0.1)
	if source.stats.regenDelay == 0 {
		multiplier += 0.3
	}
	if dist > core.realRadius*1.2 {
		multiplier -= 0.5
	}
	if source.percengage <= 0.25 {
		multiplier += 0.35
	} else if source.percengage <= 0.5 {
		multiplier += 0.1
	}
	return gmath.ClampMin(distScore*multiplier, 0.01)
}
