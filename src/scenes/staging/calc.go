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
	if !info.stats.Flying {
		techLevel += 0.2
	}
	extraTech := gmath.Clamp(techLevel-info.minTechLevel, 0, 1.0)
	if extraTech == 0 {
		return 1 + info.extraUnits
	}
	numUnits := gmath.Clamp((1+info.extraUnits)+int(math.Round(float64(info.maxUnits)*extraTech)), 1, info.maxUnits)
	return numUnits
}

func getTurretPower(stats *gamedata.AgentStats) int {
	switch stats {
	case gamedata.GunpointAgentStats:
		return 35
	case gamedata.BeamTowerAgentStats:
		return 45
	case gamedata.RepulseTowerAgentStats, gamedata.DroneFactoryAgentStats:
		return 50
	case gamedata.MegaRoombaAgentStats:
		return 55
	default:
		return 0
	}
}

func calcCreepPower(world *worldState, creep *creepNode) int {
	power := 2.3 * float64(creepDangerScore(creep))
	if creep.super {
		power *= float64(superCreepCostMultiplier(creep.stats))
	}
	if creep.stats.Kind == gamedata.CreepUberBoss {
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
	turretPower := 0
	for _, turret := range world.turrets {
		power := getTurretPower(turret.stats)
		if power == 0 {
			continue
		}
		if turret.pos.DistanceSquaredTo(pos) < rSqr {
			dangerDecrease += turretPower
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
		damage.Health *= damageMultiplier(target.GetTargetInfo(), weapon)
	}
	return damage
}

func damageMultiplier(info targetInfo, weapon *gamedata.WeaponStats) float64 {
	var m float64
	if info.flying {
		m = weapon.FlyingTargetDamageMult
	} else {
		m = weapon.GroundTargetDamageMult
	}
	if info.building {
		m *= weapon.BuildingTargetDamageMult
	}
	return m
}

func superCreepCostMultiplier(stats *gamedata.CreepStats) int {
	switch stats.Kind {
	case gamedata.CreepCrawler:
		return 3
	case gamedata.CreepTurret, gamedata.CreepBase, gamedata.CreepCrawlerBase, gamedata.CreepHowitzer, gamedata.CreepDominator, gamedata.CreepServant:
		return 5
	case gamedata.CreepUberBoss, gamedata.CreepBuilder:
		return 2
	}
	return 4
}

func creepCost(stats *gamedata.CreepStats, super bool) int {
	fragScore := creepFragScore(stats)
	if super {
		fragScore *= superCreepCostMultiplier(stats)
	}
	return fragScore
}

func creepDangerScore(creep *creepNode) int {
	score := 0
	switch creep.stats {
	case gamedata.GrenadierCreepStats:
		score = 2
	case gamedata.TemplarCreepStats:
		score = 10
	case gamedata.BuilderCreepStats:
		score = 5
	case gamedata.IonMortarCreepStats:
		score = 6
	default:
		score = creepFragScore(creep.stats)
	}
	return score
}

func creepFragScore(stats *gamedata.CreepStats) int {
	switch stats {
	case gamedata.CrawlerCreepStats:
		return 4
	case gamedata.EliteCrawlerCreepStats:
		return 6
	case gamedata.StealthCrawlerCreepStats:
		return 7
	case gamedata.HeavyCrawlerCreepStats:
		return 8

	case gamedata.WandererCreepStats:
		return 6
	case gamedata.StunnerCreepStats:
		return 9
	case gamedata.TemplarCreepStats:
		return 13
	case gamedata.CenturionCreepStats:
		return 13
	case gamedata.GrenadierCreepStats:
		return 14
	case gamedata.AssaultCreepStats:
		return 15
	case gamedata.BuilderCreepStats:
		return 30

	case gamedata.TurretCreepStats:
		return 18
	case gamedata.IonMortarCreepStats:
		return 14
	case gamedata.FortressCreepStats:
		return 50

	case gamedata.ServantCreepStats:
		return 30
	case gamedata.DominatorCreepStats:
		return 65
	case gamedata.HowitzerCreepStats:
		return 85

	case gamedata.UberBossCreepStats:
		return 235

	default:
		return 0
	}
}

func calcScore(world *worldState) int {
	switch world.config.GameMode {
	case gamedata.ModeTutorial:
		return 500

	case gamedata.ModeInfArena:
		score := world.config.DifficultyScore * 9
		timePlayed := world.result.TimePlayed.Seconds()
		if timePlayed < 5*60 {
			return 0
		}
		timePlayed -= 5 * 60
		baselineTime := 60.0 * 60.0
		multiplier := timePlayed / baselineTime
		return int(math.Round(float64(score)*multiplier)) + (world.config.DifficultyScore / 5)

	case gamedata.ModeArena:
		score := world.config.DifficultyScore * 11
		crystalsCollected := gmath.Percentage(world.result.RedCrystalsCollected, world.numRedCrystals)
		score += crystalsCollected * 3
		var multiplier float64
		if world.result.CreepTotalValue != 0 {
			multiplier = float64(world.result.CreepFragScore) / float64(world.result.CreepTotalValue)
		}
		return int(math.Round(float64(score)*multiplier)) + (world.config.DifficultyScore / 4)

	case gamedata.ModeReverse:
		score := world.config.DifficultyScore * 8
		if world.boss != nil {
			score += int((world.boss.health / world.boss.maxHealth) * 500.0)
		}
		multiplier := 1.0 - (0.000347222 * (world.result.TimePlayed.Seconds() / 5))
		if multiplier < 0 {
			multiplier = 0.001
		}
		return int(math.Round(float64(score)*multiplier)) + (world.config.DifficultyScore / 4)

	case gamedata.ModeClassic:
		score := world.config.DifficultyScore * 10
		crystalsCollected := gmath.Percentage(world.result.RedCrystalsCollected, world.numRedCrystals)
		score += crystalsCollected * 3
		multiplier := 1.0 - (0.000347222 * (world.result.TimePlayed.Seconds() / 5))
		if multiplier < 0 {
			multiplier = 0.001
		}
		return int(math.Round(float64(score)*multiplier)) + (world.config.DifficultyScore / 6)

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
	if source.stats.regenDelay != 0 && source.percengage < 0.15 || source.beingHarvested {
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
	if source.stats == sulfurSource {
		multiplier /= 2
	}
	return gmath.ClampMin(distScore*multiplier, 0.01)
}
