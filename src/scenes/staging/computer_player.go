package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type computerPlayer struct {
	world *worldState
	state *playerState
	scene *ge.Scene

	choiceGen       *choiceGenerator
	choiceSelection choiceSelection

	colonies    []*computerColony
	attackGroup []*computerColony

	resourceCards  []int
	growthCards    []int
	evolutionCards []int
	securityCards  []int

	comebackDelay        float64
	captureDelay         float64
	actionDelay          float64
	buildColonyDelay     float64
	buildTurretDelay     float64
	turretCostMultiplier float64

	minRadiusBeforeColony float64

	colonyTargetRadius float64
	maxColonies        int

	colonyPower           int
	calculatedColonyPower bool
	disposed              bool

	hasRepairbots bool
	hasFirebugs   bool
	hasBombers    bool
	hasPrisms     bool
	hasScavengers bool
}

type computerColony struct {
	attackDelay     float64
	attackBaseDelay float64
	moveDelay       float64
	factionDelay    float64
	specialDelay    float64
	defendDelay     float64
	node            *colonyCoreNode
	maxTurrets      int

	dangerLevel int

	attacking int

	retreatPos gmath.Vec
	nextPos    gmath.Vec

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

	switch p.world.turretDesign {
	case gamedata.GunpointAgentStats:
		p.turretCostMultiplier = 0.8
	case gamedata.BeamTowerAgentStats:
		p.turretCostMultiplier = 1.2
	case gamedata.TetherBeaconAgentStats:
		p.turretCostMultiplier = 1.1
	case gamedata.HarvesterAgentStats:
		p.turretCostMultiplier = 0.9
	case gamedata.SiegeAgentStats:
		p.turretCostMultiplier = 0.75
	}

	numColoniesPicker := gmath.NewRandPicker[int](world.rand)
	switch world.coreDesign {
	case gamedata.DenCoreStats:
		p.colonyTargetRadius = 370
		p.minRadiusBeforeColony = 180
		numColoniesPicker.AddOption(1, 0.05)
		numColoniesPicker.AddOption(2, 0.2)
		numColoniesPicker.AddOption(3, 0.55)
		numColoniesPicker.AddOption(4, 0.2)
	case gamedata.ArkCoreStats:
		p.colonyTargetRadius = 260
		p.minRadiusBeforeColony = 160
		p.buildColonyDelay *= 0.85
		numColoniesPicker.AddOption(2, 0.05)
		numColoniesPicker.AddOption(3, 0.2)
		numColoniesPicker.AddOption(4, 0.55)
		numColoniesPicker.AddOption(5, 0.2)
	case gamedata.TankCoreStats:
		p.colonyTargetRadius = 210
		p.minRadiusBeforeColony = 140
		p.buildColonyDelay *= 0.3
		numColoniesPicker.AddOption(4, 0.1)
		numColoniesPicker.AddOption(5, 0.4)
		numColoniesPicker.AddOption(6, 0.3)
		numColoniesPicker.AddOption(7, 0.2)
	default:
		panic("bot can't play on this core design")
	}
	p.maxColonies = numColoniesPicker.Pick()

	p.attackGroup = make([]*computerColony, 0, p.maxColonies)

	if p.world.debugLogs {
		p.world.sessionState.Logf("max colonies: %d", p.maxColonies)
	}

	for _, recipe := range p.world.tier2recipes {
		switch recipe.Result.Kind {
		case gamedata.AgentFirebug:
			p.hasFirebugs = true
		case gamedata.AgentBomber:
			p.hasBombers = true
		case gamedata.AgentPrism:
			p.hasPrisms = true
		case gamedata.AgentRepair:
			p.hasRepairbots = true
		case gamedata.AgentScavenger:
			p.hasScavengers = true
		}
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
	case gamedata.HarvesterAgentStats:
		return p.world.rand.IntRange(1, 2)
	case gamedata.SiegeAgentStats:
		return p.world.rand.IntRange(1, 2)
	default:
		return 0
	}
}

func (p *computerPlayer) Init() {
	p.choiceGen.EventChoiceReady.Connect(p, func(selection choiceSelection) {
		p.choiceSelection = selection
	})
}

func (p *computerPlayer) GetState() *playerState { return p.state }

func (p *computerPlayer) IsDisposed() bool { return p.disposed }

func (p *computerPlayer) Dispose() {
	p.disposed = true
}

func (p *computerPlayer) Update(computedDelta, delta float64) {
	if p.disposed {
		return
	}

	if len(p.state.colonies) == 0 {
		return
	}

	p.state.selectedColony = p.state.colonies[0]

	p.comebackDelay = gmath.ClampMin(p.comebackDelay-delta, 0)
	p.captureDelay = gmath.ClampMin(p.captureDelay-delta, 0)
	p.actionDelay = gmath.ClampMin(p.actionDelay-computedDelta, 0)
	p.buildColonyDelay = gmath.ClampMin(p.buildColonyDelay-computedDelta, 0)
	p.buildTurretDelay = gmath.ClampMin(p.buildTurretDelay-computedDelta, 0)

	for _, c := range p.colonies {
		if c.node.mode == colonyModeNormal {
			c.moveDelay = gmath.ClampMin(c.moveDelay-computedDelta, 0)
		}
		c.attackDelay = gmath.ClampMin(c.attackDelay-computedDelta, 0)
		c.attackBaseDelay = gmath.ClampMin(c.attackBaseDelay-computedDelta, 0)
		c.defendDelay = gmath.ClampMin(c.defendDelay-computedDelta, 0)
		c.factionDelay = gmath.ClampMin(c.factionDelay-computedDelta, 0)
		c.specialDelay = gmath.ClampMin(c.specialDelay-computedDelta, 0)
	}

	p.handleInput()
}

func (p *computerPlayer) handleInput() {
	if p.actionDelay != 0 {
		return
	}

	if p.comebackDelay == 0 {
		if p.maybeDoComeback() {
			p.comebackDelay = p.world.rand.FloatRange(140, 200)
			return
		}
		p.comebackDelay = p.world.rand.FloatRange(10, 20)
	}

	p.calculatedColonyPower = false
	if p.maybeDoAction() {
		p.actionDelay = p.world.rand.FloatRange(1.5, 4)
	} else {
		p.actionDelay = p.world.rand.FloatRange(0.75, 2.0)
	}
}

func (p *computerPlayer) findGoodComebackSpot(leaderColony *computerColony, colonyPower int, dist float64) (gmath.Vec, int) {
	var bestPos gmath.Vec
	var bestScore int
	probePos := leaderColony.node.pos.MoveTowards(p.world.rect.Center(), dist)
	randIterate(p.world.rand, comebackProbeOffsets, func(offset gmath.Vec) bool {
		pos := probePos.Add(offset)
		resourceScore, _ := p.calcPosResources(leaderColony.node, pos, 200)
		dangerScore, _ := p.calcPosDangerWithHazards(pos, 250)
		multiplier := float64(colonyPower) / float64(dangerScore)
		score := int(float64(resourceScore) * multiplier)
		if score > bestScore {
			bestScore = score
			bestPos = pos
		}
		return false
	})
	return bestPos, bestScore
}

func (p *computerPlayer) maybeDoComeback() bool {
	// Find a colony group that is on the edge of the map, too afraid
	// to do anything due to the danger levels.
	// Try to show them a chance of breaking out.

	leaderColony := randIterate(p.world.rand, p.colonies, func(c *computerColony) bool {
		if !p.world.innerRect2.Contains(c.node.pos) {
			return false
		}
		if c.dangerLevel < 4 {
			return false
		}
		if c.node.resources > c.node.maxVisualResources()*0.6 {
			return false
		}
		if c.node.mode != colonyModeNormal {
			return false
		}
		resourceScore, _ := p.calcPosResources(c.node, c.node.pos, c.node.realRadius*0.8)
		if resourceScore >= 100 {
			return false
		}
		return true
	})
	if leaderColony == nil {
		return false
	}

	power := p.calcColonyPower(leaderColony.node, gamedata.TargetAny) + 10

	// Try to find a place that worths a shot.
	candidatePos1, candidateScore1 := p.findGoodComebackSpot(leaderColony, power, 1.2*leaderColony.node.MaxFlyDistance())
	candidatePos2, candidateScore2 := p.findGoodComebackSpot(leaderColony, power, 2.3*leaderColony.node.MaxFlyDistance())
	var nextPos gmath.Vec
	bestPos := candidatePos1
	bestScore := candidateScore1
	if candidateScore2 > candidateScore1 {
		bestPos = candidatePos2
		bestScore = candidateScore2
		nextPos = bestPos
	}
	if bestScore < 15 {
		return false
	}

	group := p.attackGroup[:0]
	group = append(group, leaderColony)
	for _, c := range p.colonies {
		if c == leaderColony || c.dangerLevel == 0 {
			continue
		}
		if c.node.pos.DistanceTo(leaderColony.node.pos) < 240 {
			group = append(group, c)
		}
	}

	for i, c := range group {
		c.retreatPos = c.node.pos
		c.dangerLevel = 0
		c.moveDelay *= 2
		pos := bestPos
		if i != 0 {
			pos = bestPos.Add(p.world.rand.Offset(-80, 80))
		}
		c.nextPos = nextPos
		p.executeMoveAction(c.node, pos)
	}

	return true
}

func (p *computerPlayer) maybeDoColonyAction(colony *computerColony) bool {
	p.state.selectedColony = colony.node

	if !colony.nextPos.IsZero() {
		colony.nextPos = gmath.Vec{}
		if colony.node.pos.DistanceTo(colony.nextPos) > 64 {
			p.executeMoveAction(colony.node, colony.nextPos)
			return true
		}
	}

	// This is not a real "action", the bot just decides whether they're
	// going to send their colonies into combat.
	if colony.attackDelay == 0 {
		if p.maybeStartAttackingDreadnought(colony) {
			colony.attackDelay = p.world.rand.FloatRange(50, 110)
		} else {
			colony.attackDelay = p.world.rand.FloatRange(15, 30)
		}
	}

	if colony.attackBaseDelay == 0 {
		if delay := p.maybeAttackCreepBase(colony); delay != 0 {
			colony.attackBaseDelay = delay
		} else {
			colony.attackBaseDelay = p.world.rand.FloatRange(20, 35)
		}
	}

	// If bot is attacking the dreadnought with this colony,
	// this strategy takes the priority.
	if colony.attacking != 0 && p.world.boss != nil {
		colony.attacking--
		if p.maybeDoAttacking(colony) {
			return true
		}
		if colony.attacking == 0 || colony.node.pos.DistanceTo(p.world.boss.pos) <= 250 {
			colony.attacking = 0
			if p.maybeStayForAttack() {
				delay := p.world.rand.FloatRange(9, 16)
				colony.moveDelay = delay
				colony.defendDelay = delay
				return true
			}
		}
	}

	if colony.defendDelay == 0 {
		if delay := p.maybeDoDefensiveAction(colony); delay != 0 {
			colony.defendDelay = delay * p.world.rand.FloatRange(0.5, 2)
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

	if colony.moveDelay == 0 {
		if delay := p.maybeMoveColony(colony); delay != 0 {
			if colony.node.stats == gamedata.TankCoreStats {
				colony.moveDelay = delay * p.world.rand.FloatRange(0.6, 1.1)
			} else {
				colony.moveDelay = delay * p.world.rand.FloatRange(0.8, 1.4)
			}
			return true
		}
		colony.moveDelay = p.world.rand.FloatRange(5, 10)
	}

	// All actions below require the choices (cards) to be ready.
	if !p.choiceGen.IsReady() {
		return false
	}

	if p.buildColonyDelay == 0 && p.choiceSelection.special.special == specialBuildColony {
		if p.maybeBuildColony(colony) {
			p.buildColonyDelay = p.world.rand.FloatRange(80, 6*60)
			return true
		}
		p.buildColonyDelay = p.world.rand.FloatRange(30, 60)
	}

	if p.buildTurretDelay == 0 && p.choiceSelection.special.special == specialBuildGunpoint && colony.node.numTurretsBuilt < colony.maxTurrets {
		if p.maybeBuildTurret(colony) {
			p.buildTurretDelay = p.world.rand.FloatRange(40, 2*90)
			return true
		}
		p.buildTurretDelay = p.world.rand.FloatRange(5, 20)
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

func (p *computerPlayer) shouldWait(colony *colonyCoreNode) float64 {
	totalCargoValue := 0.0
	eliteResources := false
	t3merging := false
	makingClone := false
	building := false
	repairing := false
	sulfurMiningMinTime := 999.9
	numSulfulMining := 0
	colony.agents.Each(func(a *colonyAgentNode) {
		switch a.mode {
		case agentModeRepairBase, agentModeRepairTurret:
			repairing = true
		case agentModeBuildBuilding, agentModeCaptureBuilding:
			building = true
		case agentModeMineSulfurEssence:
			numSulfulMining++
			if a.dist < sulfurMiningMinTime {
				sulfurMiningMinTime = a.dist
			}
		case agentModeReturn, agentModeResourceTakeoff:
			totalCargoValue += a.cargoValue
			if a.cargoEliteValue != 0 {
				eliteResources = true
			}
		case agentModeMerging:
			if a.stats.Tier == 2 {
				t3merging = true
			}
		case agentModeMakeClone:
			makingClone = true
		}
	})

	if sulfurMiningMinTime <= 7 && colony.resources < 80 {
		return sulfurMiningMinTime + 5
	}
	if numSulfulMining >= 5 && colony.resources < 140 {
		return 6
	}
	if repairing {
		return 2
	}
	if building {
		return 4
	}
	if makingClone {
		return 3
	}
	if eliteResources {
		return 3
	}
	if t3merging {
		return 4
	}
	if totalCargoValue > 50 {
		return 3
	}
	if totalCargoValue > 25 && colony.resources < 50 {
		return 4
	}

	return 0
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
	if colony.stats == gamedata.TankCoreStats {
		return false
	}
	return colony.resources < gamedata.WorkerAgentStats.Cost &&
		len(colony.agents.workers) == 0
}

func (p *computerPlayer) bossTargetKind() gamedata.TargetKind {
	targetKind := gamedata.TargetFlying
	if !p.world.boss.IsFlying() {
		targetKind = gamedata.TargetGround
	}
	return targetKind
}

func (p *computerPlayer) maybeStayForAttack() bool {
	dist := p.world.boss.pos.DistanceTo(p.state.selectedColony.pos)
	if dist > 275 {
		return false
	}
	power := p.maybeAddTankPower(p.state.selectedColony, p.selectedColonyPower(p.bossTargetKind()))
	return power >= 140
}

func (p *computerPlayer) maybeDoAttacking(colony *computerColony) bool {
	// This method is used only when attacking the Dreadnought.

	if p.world.boss == nil {
		return false
	}
	targetKind := gamedata.TargetFlying
	if !p.world.boss.IsFlying() {
		targetKind = gamedata.TargetGround
	}
	if p.maybeAddTankPower(p.state.selectedColony, p.selectedColonyPower(targetKind)) < 90 {
		return false
	}

	dist := p.world.boss.pos.DistanceTo(colony.node.pos)

	if p.choiceGen.IsReady() {
		if p.choiceSelection.special.special == specialAttack && dist < 0.9*colony.node.AttackRadius() {
			return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
		}
	}

	if dist > 0.8*colony.node.PatrolRadius() {
		p.executeMoveAction(colony.node, p.world.boss.pos.Add(p.world.rand.Offset(-128, 128)))
		return true
	}

	return false
}

func (p *computerPlayer) maybeAddTankPower(colony *colonyCoreNode, power int) int {
	if p.world.coreDesign == gamedata.TankCoreStats {
		power += 30 + (2 * colony.NumAgents())
	}
	return power
}

func (p *computerPlayer) maybeAttackCreepBase(colony *computerColony) float64 {
	if !p.world.gameStarted {
		return 0
	}
	if colony.node.agents.TotalNum() < 30 {
		return 0
	}
	if colony.node.agents.TotalNum() < 45 {
		if colony.node.resources < 0.4*colony.node.maxVisualResources() {
			return 0
		}
	}

	isBlitz := p.world.config.GameMode == gamedata.ModeBlitz

	// Be less agressive when having only one colony.
	if !isBlitz && len(p.colonies) == 1 && p.world.rand.Chance(0.65) {
		return 0
	}

	power := p.selectedColonyPower(gamedata.TargetAny)
	if power < 60 {
		return 0
	}

	searchRadius := colony.node.MaxFlyDistance() + colony.node.realRadius
	if isBlitz {
		// Even if it's out of the current range, we can get there in a couple of jumps.
		searchRadius *= 2
	}
	searchRadiusSqr := searchRadius * searchRadius
	enemyBase := p.world.WalkCreeps(colony.node.pos, searchRadius, func(creep *creepNode) bool {
		switch creep.stats.Kind {
		case gamedata.CreepBase, gamedata.CreepCrawlerBase:
			// OK
		default:
			return false
		}
		// Crawler bases are less interesting, so there is a chance to ignore them.
		// For the Blitz mode, they're a winning condition, so this logic won't apply.
		if !isBlitz && creep.stats.Kind == gamedata.CreepCrawlerBase {
			if p.world.rand.Chance(0.4) {
				return false
			}
		}
		distSqr := creep.pos.DistanceSquaredTo(colony.node.pos)
		if distSqr < colony.node.realRadiusSqr {
			return false
		}
		if distSqr > searchRadiusSqr {
			return false
		}
		danger, _ := p.calcPosDanger(creep.pos, 250)
		if int(float64(danger)*1.6) > power {
			return false
		}
		return true
	})
	if enemyBase == nil {
		return 0
	}

	p.executeMoveAction(colony.node, enemyBase.pos.Add(p.world.rand.Offset(-160, 160)))

	if isBlitz {
		if enemyBase.pos.DistanceTo(colony.node.pos) > (colony.node.MaxFlyDistance() + colony.node.realRadius) {
			// It will probably need a second jump soon.
			return p.world.rand.FloatRange(10, 20)
		}
		return p.world.rand.FloatRange(40, 70)
	}

	return p.world.rand.FloatRange(80, 120)
}

func (p *computerPlayer) maybeStartAttackingDreadnought(colony *computerColony) bool {
	if p.world.boss == nil {
		return false
	}
	distSqr := p.world.boss.pos.DistanceSquaredTo(colony.node.pos)
	jumpDistSqr := colony.node.MaxFlyDistanceSqr() + 100
	if distSqr > 2.75*jumpDistSqr {
		return false
	}
	numJumps := distSqr / jumpDistSqr

	totalPower := p.maybeAddTankPower(colony.node, p.selectedColonyPower(gamedata.TargetAny))
	group := p.attackGroup[:0]
	group = append(group, colony)
	for _, otherColony := range p.colonies {
		if otherColony.node == colony.node || otherColony.attacking != 0 {
			continue
		}
		addToGroup := (otherColony.node.stats == gamedata.TankCoreStats && otherColony.node.pos.DistanceTo(colony.node.pos) <= 200) ||
			otherColony.node.pos.DistanceSquaredTo(p.world.boss.pos) <= 1.55*otherColony.node.MaxFlyDistanceSqr()
		if !addToGroup {
			continue
		}
		colonyPower := p.maybeAddTankPower(otherColony.node, p.calcColonyPower(otherColony.node, gamedata.TargetAny))
		if colonyPower < 90 {
			continue
		}
		totalPower += colonyPower
		group = append(group, otherColony)
	}

	bossDanger, _ := p.calcPosDanger(p.world.boss.pos, 200)
	bossDanger = int(float64(bossDanger) * p.world.rand.FloatRange(1.15, 1.55))
	if totalPower < bossDanger {
		return false
	}

	colony.attacking = int(math.Ceil(numJumps))
	for _, otherColony := range group[1:] {
		if otherColony.node.stats == gamedata.TankCoreStats {
			otherColony.attacking = colony.attacking
		} else {
			otherColony.attacking = 1
		}
	}
	return true
}

func (p *computerPlayer) maybeDoDefensiveAction(colony *computerColony) float64 {
	if colony.howitzerAttacker != nil {
		if p.maybeHandleHowitzerThreat(colony) {
			return 30
		}
	}

	if p.world.boss != nil {
		if p.maybeRetreatFromBoss(colony) {
			return 20
		}
	}

	if p.maybeRegroup(colony) {
		return 10
	}

	return 0
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
		p.executeMoveAction(otherColony, otherColony.pos.Add(p.world.rand.Offset(-96, 96)))
		return true
	}

	return false
}

func (p *computerPlayer) maybeRetreatFromBoss(colony *computerColony) bool {
	boss := p.world.boss
	if boss.waypoint.IsZero() {
		return false
	}

	bossDist := boss.pos.DistanceTo(colony.node.pos)
	if bossDist > colony.node.PatrolRadius()+360 {
		return false
	}

	dangerousDist := gamedata.UberBossCreepStats.Weapon.AttackRange + colony.node.PatrolRadius()
	pathDist := pointToLineDistance(colony.node.pos, boss.pos, boss.waypoint)
	if pathDist > dangerousDist {
		return false
	}

	if p.world.rand.Chance(0.65) {
		// Try to get out of the boss movement trajectory.
		bossDir := boss.waypoint.DirectionTo(boss.pos)
		probe1 := bossDir.Rotated(gmath.Rad(p.world.rand.FloatRange(0.45, 1.1))).Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
		danger1, _ := p.calcPosDangerWithHazards(probe1, colony.node.realRadius)
		probe2 := bossDir.Rotated(gmath.Rad(-p.world.rand.FloatRange(0.45, 1.1))).Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
		danger2, _ := p.calcPosDangerWithHazards(probe2, colony.node.realRadius)
		if danger1 == danger2 && p.world.rand.Bool() {
			danger1, probe1 = danger2, probe2
		}
		bestProbe := probe1
		lowestDanger := danger1
		if danger2 < danger1 {
			bestProbe = probe2
			lowestDanger = danger2
		}
		if lowestDanger == 0 || lowestDanger < 2*p.selectedColonyPower(gamedata.TargetAny) {
			p.executeMoveAction(colony.node, bestProbe)
			return true
		}
	}

	return p.maybeRetreatFrom(colony, boss.pos) != 0
}

func (p *computerPlayer) maybeHandleHowitzerThreat(colony *computerColony) bool {
	howitzer := colony.howitzerAttacker

	// We tried to resolve the issue.
	// If the attacks persist, this field will be re-assigned again.
	colony.howitzerAttacker = nil

	danger, _ := p.calcPosDanger(howitzer.pos, 250)
	requiredPower := int(float64(danger) * p.world.rand.FloatRange(0.9, 1.3))

	// A Den colony can crush howitzer just by landing on it.
	tryStomping := false
	if p.world.coreDesign == gamedata.DenCoreStats && howitzer.pos.DistanceTo(colony.node.pos) <= 0.9*colony.node.MaxFlyDistance() {
		tryStomping = true
		requiredPower /= 2
	}

	colonyPower := p.maybeAddTankPower(colony.node, p.selectedColonyPower(gamedata.TargetGround))
	if tryStomping && howitzer.pos.DistanceSquaredTo(colony.node.pos) < colony.node.realRadiusSqr {
		// If it's very close and we can stomp it, be more brave.
		colonyPower += 100
	}
	var colonyForAttack *colonyCoreNode
	if colonyPower < requiredPower {
		// Can we find another colony that can take care of it?
		otherColony := randIterate(p.world.rand, p.colonies, func(cc *computerColony) bool {
			c := cc.node
			if c == colony.node || cc.attacking != 0 {
				return false
			}
			otherColonyPower := p.calcColonyPower(c, gamedata.TargetGround)
			return otherColonyPower >= requiredPower &&
				c.resources > 40 &&
				c.mode == colonyModeNormal &&
				(c.pos.DistanceSquaredTo(howitzer.pos) < c.MaxFlyDistanceSqr()*1.1)
		})
		if otherColony != nil && p.world.rand.Chance(0.95) {
			colonyForAttack = otherColony.node
		}
	} else {
		// Have enough power to destroy the howitzer.
		if (colony.node.pos.DistanceSquaredTo(howitzer.pos) < colony.node.MaxFlyDistanceSqr()*1.5) && p.world.rand.Chance(0.75) {
			colonyForAttack = colony.node
		}
	}

	if colonyForAttack != nil {
		dstPos := howitzer.pos
		if !tryStomping {
			dstPos = dstPos.Add(p.world.rand.Offset(-140, 140))
		}
		p.executeMoveAction(colonyForAttack, dstPos)
		return true
	}

	return p.maybeRetreatFrom(colony, howitzer.pos) != 0
}

func (p *computerPlayer) maybeRetreatFrom(colony *computerColony, pos gmath.Vec) float64 {
	// Retreat to a nearby allied colony?
	if p.world.rand.Chance(0.85) {
		otherColony := randIterate(p.world.rand, p.state.colonies, func(c *colonyCoreNode) bool {
			if c == colony.node {
				return false
			}
			if c.mode != colonyModeNormal || c.NumAgents() < 25 || c.agents.NumAvailableFighters() < 5 {
				return false
			}
			dist := c.pos.DistanceSquaredTo(colony.node.pos)
			return dist < 1.5*colony.node.MaxFlyDistanceSqr() && dist > (200*200)
		})
		if otherColony != nil {
			p.executeMoveAction(colony.node, otherColony.pos.Add(p.world.rand.Offset(-96, 96)))
			return 75
		}
	}

	// Escape via teleporter?
	if p.world.rand.Chance(0.8) {
		tp := p.findUsableTeleporter(colony.node, func(tp *teleporterNode) bool {
			danger, _ := p.calcPosDanger(tp.other.pos, colony.node.PatrolRadius()+260)
			return 3*danger < p.selectedColonyPower(gamedata.TargetAny)
		})
		if tp != nil {
			p.executeMoveAction(colony.node, tp.pos)
			return 55
		}
	}

	// Try to find a safe spot to retreat to.
	currentDanger, _ := p.calcPosDanger(colony.node.pos, colony.node.PatrolRadius()+300)
	currentResourceScore, _ := p.calcPosResources(colony.node, colony.node.pos, colony.node.realRadius*0.7)

	var safestSpotPos gmath.Vec
	safestSpotDanger := currentDanger

	var richestSpotPos gmath.Vec
	richestSpotDanger := currentDanger
	richestSpotScore := currentResourceScore

	numProbes := p.world.rand.IntRange(5, 7)
	for i := 0; i < numProbes; i++ {
		dir := gmath.RadToVec(p.world.rand.Rad())
		dist := colony.node.MaxFlyDistance()*p.world.rand.FloatRange(0.7, 1.2) + 200
		candidatePos := dir.Mulf(dist).Add(colony.node.pos)
		danger, _ := p.calcPosDanger(candidatePos, colony.node.PatrolRadius()+300)
		resourceScore, _ := p.calcPosResources(colony.node, colony.node.pos, colony.node.realRadius)
		if danger < safestSpotDanger {
			safestSpotDanger = danger
			safestSpotPos = candidatePos
		}
		if richestSpotScore < resourceScore && danger < currentDanger {
			richestSpotScore = resourceScore
			richestSpotDanger = danger
			richestSpotPos = candidatePos
		}
	}
	// Safe & richest. Go there and stay for a while.
	if safestSpotPos == richestSpotPos && !safestSpotPos.IsZero() {
		p.executeMoveAction(colony.node, safestSpotPos)
		return 50
	}
	// Safer & richer. Could be a good option.
	if !richestSpotPos.IsZero() && int(1.5*float64(richestSpotDanger)) < currentDanger && richestSpotScore > int(1.5*float64(currentResourceScore)) {
		p.executeMoveAction(colony.node, safestSpotPos)
		return 35
	}
	// Can't choose a rich spot (which would be strategically better).
	// Instead of going for a local safest spot, consider doing a
	// multi-jump relocation to an allied colony.
	if len(p.world.allColonies) > 1 && p.world.rand.Chance(0.6) {
		// Even if they're far away, perhaps we can get there in a couple of jumps.
		otherColony := randIterate(p.world.rand, p.world.allColonies, func(other *colonyCoreNode) bool {
			if other == colony.node || other.agents.TotalNum() < 20 || other.mode != colonyModeNormal {
				return false
			}
			if other.pos.DistanceSquaredTo(colony.node.pos) < (180 * 180) {
				return false
			}
			stepPos := colony.node.pos.MoveTowards(other.pos, 1.1*colony.node.MaxFlyDistance())
			stepDanger, _ := p.calcPosDanger(stepPos, 300)
			return int(0.9*float64(stepDanger)) <= safestSpotDanger
		})
		if otherColony != nil {
			p.executeMoveAction(colony.node, otherColony.pos.Add(p.world.rand.Offset(-96, 96)))
			return 5
		}
	}
	// As a fallback, choose the safest option.
	if !safestSpotPos.IsZero() {
		p.executeMoveAction(colony.node, safestSpotPos)
		return 15
	}

	// Just run away from the immediate threat.
	retreatAngle := pos.AngleToPoint(colony.node.pos) + gmath.Rad(p.world.rand.FloatRange(-0.35, 0.35))
	retreatDir := gmath.RadToVec(retreatAngle)
	p.executeMoveAction(colony.node, retreatDir.Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos))
	return 5
}

func (p *computerPlayer) maybeBuildColony(colony *computerColony) bool {
	if !p.world.gameStarted {
		return false
	}
	if len(p.state.colonies) >= p.maxColonies {
		return false
	}

	if colony.node.NumAgents() < 20 || colony.node.realRadius < p.minRadiusBeforeColony {
		return false
	}

	if p.world.boss != nil {
		if p.world.boss.pos.DistanceTo(colony.node.pos) < 320 {
			return false
		}
		pathDist := pointToLineDistance(colony.node.pos, p.world.boss.pos, p.world.boss.waypoint)
		if pathDist < 280 {
			return false
		}
	}

	currentResourcesScore, _ := p.calcPosResources(colony.node, colony.node.pos, colony.node.realRadius*0.7)
	canBuild := (float64(currentResourcesScore) >= 200 && (colony.node.resources*p.world.rand.FloatRange(0.8, 1.2)) > 170) ||
		((float64(currentResourcesScore) * p.world.rand.FloatRange(0.8, 1.2)) >= 400) ||
		(colony.node.stats == gamedata.TankCoreStats && len(p.state.colonies) < 3 && (colony.node.resources >= colony.node.maxVisualResources()*0.85) && colony.node.agents.NumAvailableWorkers() >= 15) ||
		(colony.node.resources > 100 && colony.node.agents.NumAvailableWorkers() >= 30 && p.world.rand.Chance(0.1))
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
	if !p.world.gameStarted {
		return false
	}

	currentResourcesScore, _ := p.calcPosResources(colony.node, colony.node.pos, colony.node.realRadius*0.7)
	canBuild := (float64(currentResourcesScore) >= 100 && (colony.node.resources*p.world.rand.FloatRange(0.8, 1.2)) > (140*p.turretCostMultiplier)) ||
		((float64(currentResourcesScore) * p.world.rand.FloatRange(0.8, 1.2)) >= 250) ||
		(colony.node.resources > 80 && colony.node.agents.NumAvailableWorkers() > 30 && p.world.rand.Chance(0.15))
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
		increaseRadius := (colony.node.resources > 100 && colony.node.realRadius < p.colonyTargetRadius) ||
			(colony.node.realRadius < 200 && p.world.rand.Chance(0.5))
		if increaseRadius {
			return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
		}
	}

	if p.choiceSelection.special.special == specialDecreaseRadius {
		if p.hasPrisms && colony.node.realRadius > 140 && (float64(colony.node.NumAgents())/float64(colony.node.stats.DroneLimit)) >= 0.65 {
			numUnits := 0
			for _, a := range colony.node.agents.fighters {
				switch a.stats.Kind {
				case gamedata.AgentPrism:
					numUnits++
				}
			}
			if numUnits > 5 {
				// 6 => 0.1
				// 15 => 1.0
				chance := float64(numUnits-5) * 0.1
				if chance >= 1 || p.world.rand.Chance(chance) {
					return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
				}
			}
		}
		if colony.node.resources < 80 && p.world.rand.Chance(0.6) {
			preferredRadius := 0.0
			numDrones := colony.node.NumAgents()
			switch {
			case numDrones < 10:
				preferredRadius = 150
			case numDrones < 20:
				preferredRadius = 180
			}
			if preferredRadius != 0 && colony.node.realRadius > preferredRadius {
				return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
			}
		}
	}

	if p.choiceSelection.special.special == specialAttack {
		if colony.node.resources > 50 && p.world.rand.Chance(0.7) {
			power := p.selectedColonyPower(gamedata.TargetAny)
			if power > 90 {
				danger, _ := p.calcPosDanger(colony.node.pos, colony.node.AttackRadius()*1.1)
				if danger != 0 && power > 2*danger {
					return p.tryExecuteAction(colony.node, 4, gmath.Vec{})
				}
			}
		}

		if p.hasFirebugs || p.hasBombers {
			numUnits := 0
			for _, a := range colony.node.agents.fighters {
				switch a.stats.Kind {
				case gamedata.AgentFirebug, gamedata.AgentBomber:
					numUnits++
				}
			}
			if numUnits > 2 {
				// 3 => 0.1
				// 10 => 0.8
				// 12 => 1.0
				chance := float64(numUnits-2) * 0.1
				if chance >= 1 || p.world.rand.Chance(chance) {
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
			(c.resources >= (c.maxVisualResources() * 0.85))
		if needMoreGrowth {
			increaseGrowthChance := gmath.Clamp(0.1+(1.0-(p.world.rand.FloatRange(0.9, 1.2)*c.GetGrowthPriority())), 0, 1)
			if p.world.rand.Chance(increaseGrowthChance) {
				return p.tryExecuteAction(colony.node, gmath.RandElem(p.world.rand, p.growthCards), gmath.Vec{})
			}
		}
	}

	if colony.node.evoPoints < blueEvoThreshold && colony.moveDelay >= 10 && c.NumAgents() >= 20 && len(p.evolutionCards) != 0 {
		needMoreEvolution := (c.agents.tier2Num >= 4 && c.agents.tier3Num < 15) ||
			(c.agents.tier2Num < 5 && c.GetEvolutionPriority() < 0.05)
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

func (p *computerPlayer) maybeMoveColony(colony *computerColony) float64 {
	resourcesReach := colony.node.realRadius*0.4 + 100
	resourcesScore, _ := p.calcPosResources(colony.node, colony.node.pos, resourcesReach)

	// Reason to move 1: swarmed by enemies.
	colonyPower := int(float64(p.selectedColonyPower(gamedata.TargetAny)) * p.world.rand.FloatRange(0.9, 1.1))
	danger, dangerousPos := p.calcPosDanger(colony.node.pos, colony.node.realRadius+100)
	doRetreat := (danger >= p.world.rand.IntRange(700, 1000)) ||
		(colony.retreatPos.IsZero() && danger > (int(2.2*float64(colonyPower))+15) && colony.node.resources < 120) ||
		(colony.retreatPos.IsZero() && danger > 30 && colonyPower < 40 && resourcesScore < 40 && colony.node.resources < 50)
	if doRetreat {
		if delay := p.maybeRetreatFrom(colony, dangerousPos); delay != 0 {
			colony.retreatPos = colony.node.pos
			return delay
		}
	}

	// Other actions are not as important, so we can wait a bit.
	if delay := p.shouldWait(colony.node); delay != 0 {
		return delay
	}

	// Reason to move 2: resources.
	maxResources := colony.node.stats.ResourcesLimit
	if colony.node.resources < (maxResources*0.8) && p.world.rand.Chance(0.85) {
		upkeepCost, _ := colony.node.calcUpkeed()
		minAcceptableResourceScore := p.world.rand.IntRange(0, 20) + int(upkeepCost)
		if colony.node.resources < (maxResources * 0.33) {
			minAcceptableResourceScore += p.world.rand.IntRange(20, 50) + int(upkeepCost)
		}
		if resourcesScore < minAcceptableResourceScore {
			if delay := p.moveColonyToResources(resourcesScore, colony); delay != 0 {
				colony.retreatPos = gmath.Vec{}
				return delay
			}
		}
	}

	// Reason to move 3: can capture something.
	if len(p.world.neutralBuildings) != 0 && p.captureDelay == 0 && colony.node.resources >= 130 && p.world.rand.Chance(0.4) {
		b := randIterate(p.world.rand, p.world.neutralBuildings, func(b *neutralBuildingNode) bool {
			if b.stats == gamedata.PowerPlantAgentStats {
				// Bot can't utilize a power plant properly.
				return false
			}
			return b.agent == nil && b.pos.DistanceSquaredTo(colony.node.pos) < colony.node.MaxFlyDistanceSqr()+96
		})
		if b != nil {
			danger, _ := p.calcPosDangerWithHazards(colony.node.pos, colony.node.realRadius+100)
			if danger < 2*p.selectedColonyPower(gamedata.TargetAny) {
				p.captureDelay = p.world.rand.FloatRange(50, 100)
				p.executeMoveAction(colony.node, b.pos.Add(p.world.rand.Offset(-128, 128)))
				colony.retreatPos = gmath.Vec{}
				return 70
			} else {
				p.captureDelay = p.world.rand.FloatRange(15, 30)
			}
		}
	}

	// Reason to move 4: close to the map boundary.
	// This is just inconvenient for the player.
	if !p.world.innerRect.Contains(colony.node.pos) && p.world.rand.Chance(0.9) {
		pos := randomSectorPos(p.world.rand, p.world.innerRect)
		dir := gmath.RadToVec(colony.node.pos.AngleToPoint(pos))
		candidatePos := dir.Mulf(colony.node.MaxFlyDistance()).Add(colony.node.pos)
		danger, _ := p.calcPosDanger(candidatePos, colony.node.PatrolRadius()+100)
		power := p.selectedColonyPower(gamedata.TargetAny)
		if danger < power/3 {
			p.executeMoveAction(colony.node, candidatePos)
			return 5
		}
	}

	return 0
}

func (p *computerPlayer) selectedColonyPower(targetFlags gamedata.TargetKind) int {
	if !p.calculatedColonyPower {
		p.calculatedColonyPower = true
		p.colonyPower = p.calcColonyPower(p.state.selectedColony, targetFlags)
	}
	return p.colonyPower
}

func (p *computerPlayer) calcColonyPower(c *colonyCoreNode, targetFlags gamedata.TargetKind) int {
	score := 0
	targetInfo := targetInfo{flying: targetFlags&gamedata.TargetFlying != 0}
	c.agents.Each(func(a *colonyAgentNode) {
		if a.stats.PowerScore == 0 {
			return
		}
		power := a.stats.PowerScore
		if a.stats.Weapon != nil && targetFlags != gamedata.TargetAny {
			if !a.CanAttack(targetFlags) {
				return
			}
			power *= damageMultiplier(targetInfo, a.stats.Weapon)
		}
		droneScore := power
		switch a.rank {
		case 1:
			droneScore += power * 0.25
		case 2:
			droneScore += power * 0.5
		}
		if a.faction == gamedata.RedFactionTag {
			droneScore += power * 0.1
		}
		droneScore *= ((a.health / a.maxHealth) + 0.2) * p.world.dronePowerMultiplier
		score += int(droneScore)
	})
	switch {
	case c.realRadius < 150:
		score = int(float64(score) * 1.2)
	case c.realRadius < 200:
		score = int(float64(score) * 1.1)
	case c.realRadius < 250:
		score = int(float64(score) * 1.05)
	}
	return score
}

func (p *computerPlayer) calcPosResources(colony *colonyCoreNode, pos gmath.Vec, r float64) (int, gmath.Vec) {
	resourcesScore := 0
	bestResource := 0
	var bestResourcePos gmath.Vec
	rSqr := r * r
	for _, res := range p.world.essenceSources {
		if res.beingHarvested {
			continue
		}
		if res.pos.DistanceSquaredTo(pos) > rSqr {
			continue
		}
		score := int(res.stats.value) * res.resource
		if res.stats == redCrystalSource {
			score = (score*2 + 10)
		}
		if res.stats == artifactSource {
			score += 10
		}
		if res.stats == mineralSource {
			switch {
			case colony.agents.tier2plusWorkerNum == 0:
				score = 0
			case colony.agents.tier2plusWorkerNum < 5:
				score /= 2
			}
		}
		if res.stats == redOilSource {
			if colony.agents.hasRedMiner {
				score = (score*2 + 20)
			} else {
				score = 0
			}
		}
		if res.stats.scrap && p.hasScavengers {
			score = int(float64(score)*1.2) + 1
		}
		if res.stats == sulfurSource {
			if p.world.turretDesign == gamedata.TetherBeaconAgentStats {
				score /= 2
			} else {
				score /= 3
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
	waypointPos := colony.node.pos.MoveTowards(pos, colony.node.MaxFlyDistance())
	danger, _ := p.calcPosDangerWithHazards(waypointPos, colony.node.PatrolRadius()+100)
	power := p.selectedColonyPower(gamedata.TargetAny)
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
				danger, _ := p.calcPosDangerWithHazards(checkedSpot, colony.node.PatrolRadius()+260)
				if !colony.retreatPos.IsZero() && checkedSpot.DistanceSquaredTo(colony.retreatPos) < (260*260) {
					danger += 60
				}
				if danger <= maxDanger {
					// Good: can take a location closer to the resource.
					bestScore = score
					bestScorePos = checkedSpot
					continue
				}
				danger, _ = p.calcPosDangerWithHazards(candidatePos, colony.node.PatrolRadius()+260)
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
		danger, _ := p.calcPosDangerWithHazards(tp.other.pos, colony.node.PatrolRadius()+260)
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

func (p *computerPlayer) moveColonyToResources(currentResourcesScore int, colony *computerColony) float64 {
	resourcesReach := colony.node.MaxFlyDistance() + 60
	currentSpotScore := int(float64(currentResourcesScore) * p.world.rand.FloatRange(0.9, 1.25))

	colonyPower := p.selectedColonyPower(gamedata.TargetAny)

	bestScorePos := p.findBestResourcesSpot(colony, colonyPower/2, currentSpotScore+150, resourcesReach)
	if bestScorePos.IsZero() {
		bestScorePos = p.findBestResourcesSpot(colony, int(float64(colonyPower)*0.7), p.world.rand.IntRange(0, 100), 2*resourcesReach)
		if bestScorePos.IsZero() {
			// findBestResourcesSpot() failed twice, and it only does so
			// in case of the danger.
			colony.dangerLevel++
		}
		if currentResourcesScore < 80 && bestScorePos.IsZero() {
			// Try to find at least some new resources spot if we can't find a best spot around us.
			bestScorePos = p.findRandomResourcesSpot(colony)
		}
	}
	if bestScorePos.IsZero() {
		return 0
	}

	colony.dangerLevel = 0

	bestScorePos = correctedPos(p.world.innerRect2, bestScorePos, 0)
	var minDist float64
	switch colony.node.stats {
	case gamedata.DenCoreStats:
		minDist = 180
	case gamedata.TankCoreStats:
		minDist = 100
	case gamedata.ArkCoreStats:
		minDist = 70
	}
	if minDist != 0 && bestScorePos.DistanceTo(colony.node.pos) < minDist {
		return 0
	}

	nextMoveDelay := 55.0
	if bestScorePos.DistanceSquaredTo(colony.node.pos) > 1.25*colony.node.MaxFlyDistanceSqr() {
		nextMoveDelay = 5
	}

	if p.world.coreDesign == gamedata.ArkCoreStats {
		offset := gmath.Vec{X: 64, Y: 24}
		if p.world.rand.Bool() {
			offset.X = -offset.X
		}
		bestScorePos = bestScorePos.Add(offset)
	}
	p.executeMoveAction(colony.node, bestScorePos)
	return nextMoveDelay
}

func (p *computerPlayer) executeMoveAction(colony *colonyCoreNode, pos gmath.Vec) {
	p.tryExecuteAction(colony, -1, pos)
}

func (p *computerPlayer) tryExecuteAction(colony *colonyCoreNode, cardIndex int, pos gmath.Vec) bool {
	return p.choiceGen.TryExecute(colony, cardIndex, pos)
}

func (p *computerPlayer) calcPosDangerWithHazards(pos gmath.Vec, r float64) (int, gmath.Vec) {
	danger, dangerPos := calcPosDanger(p.world, p.state, pos, r)

	if p.world.envKind == gamedata.EnvInferno {
		geyserSafeDist := r * 0.9
		for _, geyser := range p.world.lavaGeysers {
			dist := geyser.pos.DistanceTo(pos)
			if dist < geyserSafeDist {
				// dist=0   => 350
				// dist=100 => 175
				// dist=200 => ~1
				danger += int((gmath.ClampMin(200-dist, 1) / 200.0) * 350)
				if dangerPos.IsZero() {
					dangerPos = geyser.pos
				}
			}
		}
		for _, puddle := range p.world.lavaPuddles {
			if puddle.centerPos.DistanceTo(pos) < r {
				if p.hasRepairbots {
					danger += 4
				} else {
					danger += 10
				}
				if dangerPos.IsZero() {
					dangerPos = puddle.centerPos
				}
			}
		}
	}

	return danger, dangerPos
}

func (p *computerPlayer) calcPosDanger(pos gmath.Vec, r float64) (int, gmath.Vec) {
	return calcPosDanger(p.world, p.state, pos, r)
}

func (p *computerPlayer) findUsableTeleporter(colony *colonyCoreNode, f func(*teleporterNode) bool) *teleporterNode {
	if p.world.coreDesign == gamedata.ArkCoreStats {
		// Can't use teleporters.
		return nil
	}

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
