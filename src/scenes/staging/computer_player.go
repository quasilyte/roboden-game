package staging

import (
	"fmt"
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

// TODO:
// - do not start building turrets in the middle of nowhere (beginning of the game)

type computerPlayer struct {
	world *worldState
	state *playerState
	scene *ge.Scene

	choiceGen       *choiceGenerator
	choiceSelection choiceSelection

	colonies []*computerColony

	resourceCards  []int
	growthCards    []int
	evolutionCards []int
	securityCards  []int

	actionDelay      float64
	buildColonyDelay float64
	buildTurretDelay float64

	maxColonies int

	colonyPower           int
	calculatedColonyPower bool
}

type computerColony struct {
	attackDelay  float64
	moveDelay    float64
	factionDelay float64
	specialDelay float64
	defendDelay  float64
	node         *colonyCoreNode
	maxTurrets   int

	attacking int

	howitzerAttacker *creepNode
}

func newComputerPlayer(world *worldState, state *playerState, choiceGen *choiceGenerator) *computerPlayer {
	p := &computerPlayer{
		world:     world,
		state:     state,
		scene:     world.rootScene,
		choiceGen: choiceGen,

		resourceCards:  make([]int, 0, 4),
		growthCards:    make([]int, 0, 4),
		evolutionCards: make([]int, 0, 4),
		securityCards:  make([]int, 0, 4),

		buildColonyDelay: world.rand.FloatRange(60, 3*60),
	}

	switch roll := world.rand.Float(); {
	case roll < 0.05: // 5%
		p.maxColonies = 1
	case roll < 0.25: // 20%
		p.maxColonies = 2
	case roll < 0.80: // 55%
		p.maxColonies = 3
	default: // 20%
		p.maxColonies = 4
	}

	if p.world.debugLogs {
		fmt.Println("max colonies:", p.maxColonies)
	}

	p.world.EventColonyCreated.Connect(p, func(colony *colonyCoreNode) {
		if colony.player != p {
			return
		}
		wrapped := &computerColony{
			node:       colony,
			moveDelay:  p.world.rand.FloatRange(10, 15),
			maxTurrets: p.maxTurretsForColony(),
		}
		colony.EventDestroyed.Connect(p, func(_ *colonyCoreNode) {
			p.colonies = xslices.Remove(p.colonies, wrapped)
		})
		colony.EventOnDamage.Connect(p, func(attacker targetable) {
			creep, ok := attacker.(*creepNode)
			if !ok {
				return
			}
			if creep.stats.Kind == gamedata.CreepHowitzer {
				wrapped.howitzerAttacker = creep
			}
		})
		p.colonies = append(p.colonies, wrapped)
	})

	return p
}

func (p *computerPlayer) maxTurretsForColony() int {
	switch p.world.turretDesign {
	case gamedata.GunpointAgentStats:
		return p.world.rand.IntRange(2, 5)
	case gamedata.BeamTowerAgentStats:
		return p.world.rand.IntRange(1, 4)
	case gamedata.TetherBeaconAgentStats:
		return p.world.rand.IntRange(0, 2)
	default:
		return 0
	}
}

func (p *computerPlayer) Init() {
	p.state.Init(p.world)

	p.choiceGen.EventChoiceReady.Connect(p, func(selection choiceSelection) {
		p.choiceSelection = selection
	})
}

func (p *computerPlayer) GetState() *playerState { return p.state }

func (p *computerPlayer) IsDisposed() bool { return false }

func (p *computerPlayer) Update(computedDelta, delta float64) {
	if p.world.nodeRunner.IsPaused() || len(p.state.colonies) == 0 {
		return
	}

	p.state.selectedColony = p.state.colonies[0]

	p.actionDelay = gmath.ClampMin(p.actionDelay-computedDelta, 0)
	p.buildColonyDelay = gmath.ClampMin(p.buildColonyDelay-computedDelta, 0)
	p.buildTurretDelay = gmath.ClampMin(p.buildTurretDelay-computedDelta, 0)

	for _, c := range p.colonies {
		c.attackDelay = gmath.ClampMin(c.attackDelay-computedDelta, 0)
		c.defendDelay = gmath.ClampMin(c.defendDelay-computedDelta, 0)
		c.moveDelay = gmath.ClampMin(c.moveDelay-computedDelta, 0)
		c.factionDelay = gmath.ClampMin(c.factionDelay-computedDelta, 0)
		c.specialDelay = gmath.ClampMin(c.specialDelay-computedDelta, 0)
	}
}

func (p *computerPlayer) HandleInput() {
	if p.world.nodeRunner.IsPaused() || len(p.state.colonies) == 0 {
		return
	}
	if !p.choiceGen.IsReady() {
		return
	}
	if p.actionDelay != 0 {
		return
	}

	p.calculatedColonyPower = false
	if p.maybeDoAction() {
		p.actionDelay = p.world.rand.FloatRange(1.5, 4)
	} else {
		p.actionDelay = p.world.rand.FloatRange(0.75, 2.0)
	}
}

func (p *computerPlayer) maybeDoColonyAction(colony *computerColony) bool {
	p.state.selectedColony = colony.node

	if colony.attacking != 0 {
		colony.attacking--
		if p.maybeDoAttacking(colony) {
			return true
		}
	}

	if colony.attackDelay == 0 {
		if p.maybeDoAttackAction(colony) {
			colony.attackDelay = p.world.rand.FloatRange(60, 120)
			return true
		}
		colony.attackDelay = p.world.rand.FloatRange(15, 30)
	}

	if colony.defendDelay == 0 {
		if p.maybeDoDefensiveAction(colony) {
			colony.defendDelay = p.world.rand.FloatRange(30, 70)
			return true
		}
		colony.defendDelay = p.world.rand.FloatRange(3, 6)
	}

	// Only defensive actions (like relocation) are allowed
	// for any unrecoverable colony.
	// It will be a waste of a turn to invest cards here.
	if p.colonyCantRecover(colony.node) {
		return false
	}

	if p.buildColonyDelay == 0 && p.choiceSelection.special.special == specialBuildColony {
		if p.maybeBuildColony(colony) {
			p.buildColonyDelay = p.world.rand.FloatRange(80, 6*60)
			return true
		}
		p.buildColonyDelay = p.world.rand.FloatRange(30, 60)
	}

	if p.buildTurretDelay == 0 && p.choiceSelection.special.special == specialBuildGunpoint && len(colony.node.turrets) < colony.maxTurrets {
		if p.maybeBuildTurret(colony) {
			p.buildTurretDelay = p.world.rand.FloatRange(40, 2*90)
			return true
		}
		p.buildTurretDelay = p.world.rand.FloatRange(5, 20)
	}

	if colony.moveDelay == 0 {
		if p.maybeMoveColony(colony) {
			colony.moveDelay = p.world.rand.FloatRange(40.0, 80.0)
			return true
		}
		colony.moveDelay = p.world.rand.FloatRange(5, 10)
	}

	if colony.specialDelay == 0 {
		if p.maybeUseSpecial(colony) {
			colony.specialDelay = p.world.rand.FloatRange(15, 50)
			return true
		}
		colony.specialDelay = p.world.rand.FloatRange(2, 5)
	}

	if colony.factionDelay == 0 {
		if p.maybeChangePriorities(colony) {
			colony.factionDelay = p.world.rand.FloatRange(10, 15)
			return true
		}
		colony.factionDelay = p.world.rand.FloatRange(3, 20)
	}

	return false
}

func (p *computerPlayer) maybeDoAction() bool {
	return nil != randIterate(p.world.rand, p.colonies, func(c *computerColony) bool {
		if c.node.mode != colonyModeNormal {
			return false
		}
		return p.maybeDoColonyAction(c)
	})
}

func (p *computerPlayer) colonyCantRecover(colony *colonyCoreNode) bool {
	return colony.resources < gamedata.WorkerAgentStats.Cost &&
		len(colony.agents.workers) == 0
}

func (p *computerPlayer) maybeDoAttacking(colony *computerColony) bool {
	if p.world.boss == nil {
		return false
	}
	if p.selectedColonyPower() < 100 {
		return false
	}

	dist := p.world.boss.pos.DistanceTo(colony.node.pos)
	if p.choiceSelection.special.special == specialAttack && dist < 0.9*colony.node.AttackRadius() {
		return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
	}
	if dist < 0.8*colony.node.PatrolRadius() {
		return false
	}

	return p.tryExecuteAction(colony.node, -1, p.world.boss.pos.Add(p.world.rand.Offset(-128, 128)))
}

func (p *computerPlayer) maybeDoAttackAction(colony *computerColony) bool {
	if p.world.boss == nil {
		return false
	}
	distSqr := p.world.boss.pos.DistanceSquaredTo(colony.node.pos)
	jumpDistSqr := colony.node.MaxFlyDistanceSqr() + 100
	if distSqr > 2.75*jumpDistSqr {
		return false
	}
	numJumps := distSqr / jumpDistSqr

	colonyPower := p.selectedColonyPower()
	bossDanger, _ := calcPosDanger(p.world, p.state, p.world.boss.pos, 200)
	bossDanger = int(float64(bossDanger) * p.world.rand.FloatRange(0.9, 1.5))
	if colonyPower < bossDanger {
		return false
	}

	colony.attacking = int(math.Ceil(numJumps))
	for _, otherColony := range p.colonies {
		if otherColony.node == colony.node {
			continue
		}
		if otherColony.node.pos.DistanceSquaredTo(p.world.boss.pos) > 1.5*otherColony.node.MaxFlyDistanceSqr() {
			continue
		}
		colonyPower := p.calcColonyPower(otherColony.node)
		if colonyPower < 120 {
			continue
		}
		otherColony.attacking = 1
	}
	return true
}

func (p *computerPlayer) maybeDoDefensiveAction(colony *computerColony) bool {
	if colony.howitzerAttacker != nil {
		if p.maybeHandleHowitzerThreat(colony) {
			return true
		}
	}
	if p.world.boss != nil {
		if p.maybeRetreatFromBoss(colony) {
			return true
		}
	}
	if p.maybeRegroup(colony) {
		return true
	}
	return false
}

func (p *computerPlayer) maybeRegroup(colony *computerColony) bool {
	if len(p.state.colonies) < 2 {
		return false
	}

	for _, other := range p.state.colonies {
		if colony.node == other {
			continue
		}
		if colony.node.pos.DistanceSquaredTo(other.pos) < (200 * 200) {
			return false
		}
	}

	needRegroup := (len(colony.node.agents.fighters) < 4 && colony.node.resources < 150) ||
		p.colonyCantRecover(colony.node) ||
		(colony.node.NumAgents() < 30 && colony.node.health < colony.node.maxHealth*0.75)
	if !needRegroup || p.world.rand.Chance(0.1) {
		return false
	}

	otherColony := randIterate(p.world.rand, p.state.colonies, func(c *colonyCoreNode) bool {
		if c == colony.node {
			return false
		}
		return c.mode == colonyModeNormal &&
			(c.pos.DistanceSquaredTo(colony.node.pos) < 2*colony.node.MaxFlyDistanceSqr())
	})
	if otherColony != nil {
		return p.tryExecuteAction(otherColony, -1, otherColony.pos.Add(p.world.rand.Offset(-96, 96)))
	}

	return false
}

func (p *computerPlayer) maybeRetreatFromBoss(colony *computerColony) bool {
	boss := p.world.boss
	if boss.waypoint.IsZero() {
		return false
	}

	bossDist := boss.pos.DistanceTo(colony.node.pos)
	if bossDist > colony.node.PatrolRadius()+400 {
		return false
	}

	dangerousDist := gamedata.UberBossCreepStats.Weapon.AttackRange + colony.node.PatrolRadius()
	pathDist := pointToLineDistance(colony.node.pos, boss.pos, boss.waypoint)
	if pathDist > dangerousDist {
		return false
	}

	if p.world.rand.Chance(0.9) {
		// Try to get out of the boss movement trajectory.
		bossDir := boss.waypoint.DirectionTo(boss.pos)
		probe1 := bossDir.Rotated(gmath.Rad(p.world.rand.FloatRange(0.45, 1.1))).Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
		danger1, _ := calcPosDanger(p.world, p.state, probe1, colony.node.realRadius)
		probe2 := bossDir.Rotated(gmath.Rad(-p.world.rand.FloatRange(0.45, 1.1))).Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
		danger2, _ := calcPosDanger(p.world, p.state, probe2, colony.node.realRadius)
		if danger1 == danger2 && p.world.rand.Bool() {
			danger1, probe1 = danger2, probe2
		}
		bestProbe := probe1
		lowestDanger := danger1
		if danger2 < danger1 {
			bestProbe = probe2
			lowestDanger = danger2
		}
		if lowestDanger == 0 || lowestDanger < 2*p.selectedColonyPower() {
			return p.tryExecuteAction(colony.node, -1, bestProbe)
		}
	}

	return p.maybeRetreatFrom(colony, boss.pos)
}

func (p *computerPlayer) maybeHandleHowitzerThreat(colony *computerColony) bool {
	howitzer := colony.howitzerAttacker

	// We tried to resolve the issue.
	// If the attacks persist, this field will be re-assigned again.
	colony.howitzerAttacker = nil

	danger, _ := calcPosDanger(p.world, p.state, howitzer.pos, 300)
	requiredPower := int(float64(danger) * p.world.rand.FloatRange(0.9, 1.4))

	colonyPower := p.selectedColonyPower()
	var colonyForAttack *colonyCoreNode
	if colonyPower < requiredPower {
		// Can we find another colony that can take care of it?
		otherColony := randIterate(p.world.rand, p.state.colonies, func(c *colonyCoreNode) bool {
			if c == colony.node {
				return false
			}
			otherColonyPower := p.calcColonyPower(c)
			return otherColonyPower >= requiredPower &&
				c.resources > 40 &&
				c.mode == colonyModeNormal &&
				(c.pos.DistanceSquaredTo(howitzer.pos) < c.MaxFlyDistanceSqr()*1.1)
		})
		if otherColony != nil && p.world.rand.Chance(0.95) {
			colonyForAttack = otherColony
		}
	} else {
		// Have enough power to destroy the howitzer.
		if (colony.node.pos.DistanceSquaredTo(howitzer.pos) < colony.node.MaxFlyDistanceSqr()*1.5) && p.world.rand.Chance(0.75) {
			colonyForAttack = colony.node
		}
	}

	if colonyForAttack != nil {
		return p.tryExecuteAction(colonyForAttack, -1, howitzer.pos.Add(p.world.rand.Offset(-140, 140)))
	}

	return p.maybeRetreatFrom(colony, howitzer.pos)
}

func (p *computerPlayer) maybeRetreatFrom(colony *computerColony, pos gmath.Vec) bool {
	// Retreat to a nearby allied colony?
	if p.world.rand.Chance(0.8) {
		otherColony := randIterate(p.world.rand, p.state.colonies, func(c *colonyCoreNode) bool {
			if c == colony.node {
				return false
			}
			return c.mode == colonyModeNormal &&
				c.NumAgents() >= 30 &&
				c.agents.NumAvailableFighters() >= 5 &&
				(c.pos.DistanceSquaredTo(colony.node.pos) < colony.node.MaxFlyDistanceSqr())
		})
		if otherColony != nil {
			return p.tryExecuteAction(colony.node, -1, otherColony.pos.Add(p.world.rand.Offset(-96, 96)))
		}
	}

	if p.world.rand.Chance(0.8) {
		tp := p.findUsableTeleporter(colony.node, func(tp *teleporterNode) bool {
			danger, _ := calcPosDanger(p.world, p.state, tp.other.pos, colony.node.PatrolRadius()+260)
			return 3*danger < p.selectedColonyPower()
		})
		if tp != nil {
			return p.tryExecuteAction(colony.node, -1, tp.pos)
		}
	}

	// Just run away from the immediate threat.
	if p.world.rand.Chance(0.2) {
		retreatAngle := pos.AngleToPoint(colony.node.pos) + gmath.Rad(p.world.rand.FloatRange(-0.35, 0.35))
		retreatDir := gmath.RadToVec(retreatAngle)
		return p.tryExecuteAction(colony.node, -1, retreatDir.Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos))
	}

	// Try to find a safe spot to retreat to.
	bestScore, _ := calcPosDanger(p.world, p.state, colony.node.pos, colony.node.PatrolRadius()+300)
	var bestScorePos gmath.Vec
	numProbes := p.world.rand.IntRange(3, 5)
	for i := 0; i < numProbes; i++ {
		dir := gmath.RadToVec(p.world.rand.Rad())
		dist := colony.node.MaxFlyDistance()*p.world.rand.FloatRange(0.7, 1.0) + 200
		candidatePos := dir.Mulf(dist).Add(colony.node.pos)
		danger, _ := calcPosDanger(p.world, p.state, candidatePos, colony.node.PatrolRadius()+300)
		if danger < bestScore {
			bestScore = danger
			bestScorePos = candidatePos
		}
	}
	if bestScorePos.IsZero() {
		return false
	}
	return p.tryExecuteAction(colony.node, -1, bestScorePos)
}

func (p *computerPlayer) maybeBuildColony(colony *computerColony) bool {
	if len(p.state.colonies) >= p.maxColonies {
		return false
	}

	if colony.node.NumAgents() < 20 || colony.node.realRadius < 180 {
		return false
	}

	if p.world.boss != nil {
		if p.world.boss.pos.DistanceTo(colony.node.pos) < 300 {
			return false
		}
		pathDist := pointToLineDistance(colony.node.pos, p.world.boss.pos, p.world.boss.waypoint)
		if pathDist < 300 {
			return false
		}
	}

	currentResourcesScore, _ := p.calcPosResources(colony.node, colony.node.pos, colony.node.realRadius*0.7)
	canBuild := (float64(currentResourcesScore) >= 200 && (colony.node.resources*p.world.rand.FloatRange(0.8, 1.2)) > 150) ||
		((float64(currentResourcesScore) * p.world.rand.FloatRange(0.8, 1.2)) >= 400) ||
		(colony.node.resources > 50 && colony.node.agents.NumAvailableWorkers() >= 30 && p.world.rand.Chance(0.1))
	if !canBuild {
		return false
	}

	success := p.tryExecuteAction(colony.node, 4, gmath.Vec{})
	if success {
		colony.moveDelay += p.world.rand.FloatRange(30, 100)
	}
	return success
}

func (p *computerPlayer) maybeBuildTurret(colony *computerColony) bool {
	currentResourcesScore, _ := p.calcPosResources(colony.node, colony.node.pos, colony.node.realRadius*0.7)
	canBuild := (float64(currentResourcesScore) >= 100 && (colony.node.resources*p.world.rand.FloatRange(0.8, 1.2)) > 140) ||
		((float64(currentResourcesScore) * p.world.rand.FloatRange(0.8, 1.2)) >= 250) ||
		(colony.node.resources > 80 && colony.node.agents.NumAvailableWorkers() > 40 && p.world.rand.Chance(0.15))
	if !canBuild {
		return false
	}

	success := p.tryExecuteAction(colony.node, 4, gmath.Vec{})
	if success {
		colony.moveDelay += p.world.rand.FloatRange(20, 35)
	}
	return success
}

func (p *computerPlayer) maybeUseSpecial(colony *computerColony) bool {
	if p.choiceSelection.special.special == specialIncreaseRadius {
		increaseRadius := (colony.node.resources > 100 && colony.node.realRadius < 320) ||
			(colony.node.realRadius < 200 && p.world.rand.Chance(0.5))
		if increaseRadius {
			return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
		}
	}

	if p.choiceSelection.special.special == specialDecreaseRadius {
		if colony.node.resources < 80 && p.world.rand.Chance(0.6) {
			preferredRadius := 400.0
			numDrones := colony.node.NumAgents()
			switch {
			case numDrones < 10:
				preferredRadius = 150
			case numDrones < 20:
				preferredRadius = 200
			case numDrones < 30:
				preferredRadius = 250
			case numDrones < 40:
				preferredRadius = 300
			}
			if colony.node.realRadius > preferredRadius {
				return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
			}
		}
	}

	if p.choiceSelection.special.special == specialAttack {
		if colony.node.resources > 50 && p.world.rand.Chance(0.7) {
			power := p.selectedColonyPower()
			if power > 90 {
				danger, _ := calcPosDanger(p.world, p.state, colony.node.pos, colony.node.AttackRadius()*1.1)
				if danger != 0 && power > 2*danger {
					return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
				}
			}
		}
	}

	return false
}

func (p *computerPlayer) maybeChangePriorities(colony *computerColony) bool {
	randomCardChance := 0.25
	if colony.node.factionWeights.GetWeight(gamedata.NeutralFactionTag) > 0.5 {
		randomCardChance = 0.7
	}
	if p.world.rand.Chance(randomCardChance) {
		// Use a random card.
		return p.tryExecuteAction(colony.node, p.world.rand.IntRange(0, 3), gmath.Vec{})
	}

	p.resourceCards = p.resourceCards[:0]
	p.growthCards = p.growthCards[:0]
	p.evolutionCards = p.evolutionCards[:0]
	p.securityCards = p.securityCards[:0]
	for i, option := range p.choiceSelection.cards {
		for _, e := range option.effects {
			switch e.priority {
			case priorityResources:
				p.resourceCards = append(p.resourceCards, i)
			case priorityGrowth:
				p.growthCards = append(p.growthCards, i)
			case priorityEvolution:
				p.evolutionCards = append(p.evolutionCards, i)
			case prioritySecurity:
				p.securityCards = append(p.securityCards, i)
			}
		}
	}

	c := colony.node

	if c.resources < 50 && len(p.resourceCards) != 0 {
		increaseResourcesChance := gmath.Clamp(1.0-(p.world.rand.FloatRange(0.8, 1.2)*c.GetResourcePriority()), 0, 1)
		if p.world.rand.Chance(increaseResourcesChance) {
			return p.tryExecuteAction(colony.node, gmath.RandElem(p.world.rand, p.resourceCards), gmath.Vec{})
		}
	}

	if len(p.growthCards) != 0 {
		needMoreGrowth := (c.NumAgents() < 20 && c.resources > 80) ||
			(c.health < c.maxHealth*0.9 && c.resources > 100) ||
			(c.agents.NumAvailableWorkers() < 3 && c.resources > 20) ||
			(2*c.NumAgents() < c.calcUnitLimit()) ||
			(c.resources >= (maxVisualResources * 0.85))
		if needMoreGrowth {
			increaseGrowthChance := gmath.Clamp(0.1+(1.0-(p.world.rand.FloatRange(0.9, 1.2)*c.GetGrowthPriority())), 0, 1)
			if p.world.rand.Chance(increaseGrowthChance) {
				return p.tryExecuteAction(colony.node, gmath.RandElem(p.world.rand, p.growthCards), gmath.Vec{})
			}
		}
	}

	if colony.node.evoPoints < blueEvoThreshold && colony.moveDelay >= 10 && c.NumAgents() >= 20 && len(p.evolutionCards) != 0 {
		needMoreEvolution := (c.agents.tier2Num >= 4 && c.agents.tier3Num < 15)
		if needMoreEvolution {
			increaseElolutionChance := gmath.Clamp(0.1+(1.0-(p.world.rand.FloatRange(0.7, 1.0)*c.GetGrowthPriority())), 0, 1)
			if p.world.rand.Chance(increaseElolutionChance) {
				return p.tryExecuteAction(colony.node, gmath.RandElem(p.world.rand, p.evolutionCards), gmath.Vec{})
			}
		}
	}

	if c.GetSecurityPriority() >= 0.7 {
		if len(p.growthCards) != 0 && (len(c.agents.fighters) < 10 || p.world.rand.Chance(0.3)) {
			return p.tryExecuteAction(colony.node, gmath.RandElem(p.world.rand, p.growthCards), gmath.Vec{})
		}
	}

	return false
}

func (p *computerPlayer) maybeMoveColony(colony *computerColony) bool {
	// Reason to move 1: resources.
	resourcesReach := colony.node.realRadius*0.4 + 100
	upkeepCost, _ := colony.node.calcUpkeed()
	if colony.node.resources < (maxResources*0.85) && p.world.rand.Chance(0.85) {
		resourcesScore, _ := p.calcPosResources(colony.node, colony.node.pos, resourcesReach)
		minAcceptableResourceScore := p.world.rand.IntRange(0, 25) + (2 * int(upkeepCost))
		if colony.node.resources < (maxResources * 0.33) {
			minAcceptableResourceScore += p.world.rand.IntRange(30, 130) + int(upkeepCost)
		}
		if resourcesScore < minAcceptableResourceScore {
			return p.moveColonyToResources(resourcesScore, colony)
		}
	}

	// Reason to move 2: swarmed by enemies.
	danger, dangerousPos := calcPosDanger(p.world, p.state, colony.node.pos, colony.node.realRadius+100)
	if danger >= p.world.rand.IntRange(700, 1000) {
		p.maybeRetreatFrom(colony, dangerousPos)
	}
	if danger > 150 && colony.node.resources > 200 {
		power := int(float64(p.selectedColonyPower()) * p.world.rand.FloatRange(0.9, 1.1))
		if power < danger {
			p.maybeRetreatFrom(colony, dangerousPos)
		}
	}

	// Reason to move 3: close to the map boundary.
	// This is just inconvenient for the player.
	if !p.world.innerRect.Contains(colony.node.pos) && p.world.rand.Chance(0.9) {
		pos := randomSectorPos(p.world.rand, p.world.innerRect)
		dir := gmath.RadToVec(colony.node.pos.AngleToPoint(pos))
		candidatePos := dir.Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
		danger, _ := calcPosDanger(p.world, p.state, candidatePos, colony.node.PatrolRadius()+100)
		power := p.selectedColonyPower()
		if danger < power/3 {
			return p.tryExecuteAction(colony.node, -1, candidatePos)
		}
	}

	return false
}

func (p *computerPlayer) selectedColonyPower() int {
	if !p.calculatedColonyPower {
		p.calculatedColonyPower = true
		p.colonyPower = p.calcColonyPower(p.state.selectedColony)
	}
	return p.colonyPower
}

func (p *computerPlayer) calcColonyPower(c *colonyCoreNode) int {
	score := 0
	c.agents.Find(searchFighters, func(a *colonyAgentNode) bool {
		cost := a.stats.Cost
		droneScore := cost
		switch a.rank {
		case 1:
			droneScore += cost * 0.25
		case 2:
			droneScore += cost * 0.5
		}
		if a.faction == gamedata.RedFactionTag {
			droneScore += cost * 0.1
		}
		droneScore *= ((a.health / a.maxHealth) + 0.2) * p.world.droneHealthMultiplier
		score += int(droneScore)
		return false
	})
	return score
}

func (p *computerPlayer) calcPosResources(colony *colonyCoreNode, pos gmath.Vec, r float64) (int, gmath.Vec) {
	resourcesScore := 0
	bestResource := 0
	var bestResourcePos gmath.Vec
	rSqr := r * r
	for _, res := range p.world.essenceSources {
		if res.pos.DistanceSquaredTo(pos) > rSqr {
			continue
		}
		score := int(res.stats.value) * res.resource
		if res.stats == redCrystalSource {
			score *= 3
		}
		if res.stats == redOilSource {
			if colony.agents.hasRedMiner {
				score *= 2
			} else {
				score = 0
			}
		}
		if score > bestResource {
			bestResource = score
			bestResourcePos = res.pos
		}
		resourcesScore += score
	}
	return resourcesScore, bestResourcePos
}

func (p *computerPlayer) findRandomResourcesSpot(colony *computerColony) gmath.Vec {
	pos := randomSectorPos(p.world.rand, p.world.innerRect)
	dir := gmath.RadToVec(colony.node.pos.AngleToPoint(pos))
	waypointPos := dir.Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
	danger, _ := calcPosDanger(p.world, p.state, waypointPos, colony.node.PatrolRadius()+100)
	power := p.selectedColonyPower()
	if 2*danger > power {
		return gmath.Vec{}
	}
	score, _ := p.calcPosResources(colony.node, pos, colony.node.realRadius*0.5+120)
	if score < 50 {
		return gmath.Vec{}
	}
	return pos
}

func (p *computerPlayer) findBestResourcesSpot(colony *computerColony, maxDanger, currentScore int, r float64) gmath.Vec {
	// Project several random lines and see whether any of these
	// lead us somewhere good.

	resourcesReach := colony.node.realRadius*0.5 + 120

	bestScore := currentScore
	var bestScorePos gmath.Vec
	currentAngle := p.world.rand.Rad()
	const numProbes = 8
	currentDist := r / 4
	for j := 0; j < 4; j++ {
		for i := 0; i < numProbes; i++ {
			dir := gmath.RadToVec(currentAngle)
			currentAngle += (2 * math.Pi) / numProbes
			dist := currentDist * p.world.rand.FloatRange(0.8, 1.1)
			candidatePos := dir.Mulf(dist).Add(colony.node.pos)
			score, bestResPos := p.calcPosResources(colony.node, candidatePos, resourcesReach)
			if score > bestScore {
				checkedSpot := bestResPos.Add(p.world.rand.Offset(-32, 32))
				danger, _ := calcPosDanger(p.world, p.state, checkedSpot, colony.node.PatrolRadius()+260)
				if danger <= maxDanger {
					// Good: can take a location closer to the resource.
					bestScore = score
					bestScorePos = checkedSpot
					continue
				}
				danger, _ = calcPosDanger(p.world, p.state, candidatePos, colony.node.PatrolRadius()+260)
				if danger <= maxDanger {
					// Could be suboptimal in terms of positioning, but at least it's safer.
					bestScore = score
					bestScorePos = candidatePos
					continue
				}
			}
		}
		currentDist += r / 4
	}

	// Now check if there are any teleporters around.
	// Maybe jumping there could be a good decision.
	p.findUsableTeleporter(colony.node, func(tp *teleporterNode) bool {
		danger, _ := calcPosDanger(p.world, p.state, tp.other.pos, colony.node.PatrolRadius()+260)
		if danger > maxDanger {
			return false
		}
		score, _ := p.calcPosResources(colony.node, tp.other.pos, resourcesReach)
		if score > bestScore {
			bestScore = score
			bestScorePos = tp.pos
		}
		return false
	})

	return bestScorePos
}

func (p *computerPlayer) moveColonyToResources(currentResourcesScore int, colony *computerColony) bool {
	resourcesReach := colony.node.MaxFlyDistance() + 60
	currentSpotScore := int(float64(currentResourcesScore) * p.world.rand.FloatRange(0.9, 1.25))

	colonyPower := p.selectedColonyPower()

	bestScorePos := p.findBestResourcesSpot(colony, colonyPower/2, currentSpotScore+150, resourcesReach)
	if bestScorePos.IsZero() {
		bestScorePos = p.findBestResourcesSpot(colony, int(float64(colonyPower)*0.7), p.world.rand.IntRange(0, 100), 2*resourcesReach)
		if currentResourcesScore < 80 && bestScorePos.IsZero() {
			// Try to find at least some new resources spot if we can't find a best spot around us.
			bestScorePos = p.findRandomResourcesSpot(colony)
		}
	}
	if bestScorePos.IsZero() {
		return false
	}
	return p.tryExecuteAction(colony.node, -1, bestScorePos)
}

func (p *computerPlayer) tryExecuteAction(colony *colonyCoreNode, cardIndex int, pos gmath.Vec) bool {
	prevSelected := p.state.selectedColony
	p.state.selectedColony = colony
	result := p.choiceGen.TryExecute(cardIndex, pos)
	p.state.selectedColony = prevSelected
	return result
}

func (p *computerPlayer) findUsableTeleporter(colony *colonyCoreNode, f func(*teleporterNode) bool) *teleporterNode {
	return randIterate(p.world.rand, p.world.teleporters, func(tp *teleporterNode) bool {
		if tp.pos.DistanceSquaredTo(colony.pos) > 0.9*colony.MaxFlyDistanceSqr() {
			return false
		}
		if !tp.CanBeUsedBy(colony) {
			return false
		}
		return f(tp)
	})
}
