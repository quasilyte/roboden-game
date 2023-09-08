package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

//go:generate stringer -type=specialChoiceKind -trimprefix=special
type specialChoiceKind int

const (
	specialChoiceNone specialChoiceKind = iota
	specialIncreaseRadius
	specialDecreaseRadius
	specialBuildGunpoint
	specialBuildColony
	specialAttack
	specialChoiceMoveColony

	// These are the actions for the creeps.
	specialSendCreeps
	specialRally
	specialSpawnCrawlers
	specialBossAttack
	specialIncreaseTech
	specialAtomicBomb
	specialSendCenturions

	// These are also for the creeps.
	_creepCardFirst
	specialBuyCrawlers
	specialBuyWanderers
	specialBuyEliteCrawlers
	specialBuyStunners
	specialBuyStealthCrawlers
	specialBuyHeavyCrawlers
	specialBuyBuilders
	specialBuyCenturions
	specialBuyTemplars
	specialBuyAssaults
	specialBuyDominator
	specialBuyHowitzer
	_creepCardLast
)

type selectedChoice struct {
	Index    int
	Cooldown float64
	Faction  gamedata.FactionTag
	Option   choiceOption
	Pos      gmath.Vec
	Player   player
	Colony   *colonyCoreNode
}

type choiceOption struct {
	effects   []choiceOptionEffect
	special   specialChoiceKind
	direction int
	icon      resource.ImageID
	cost      float64
}

type choiceOptionEffect struct {
	priority colonyPriority
	value    float64
}

type choiceSelection struct {
	cards   []choiceOption
	special choiceOption
}

var specialChoicesTable = [...]choiceOption{
	specialAttack: {
		special: specialAttack,
		cost:    5,
		icon:    assets.ImageActionAttack,
	},

	specialBuildColony: {
		special: specialBuildColony,
		cost:    25,
		icon:    assets.ImageActionBuildColony,
	},
	specialBuildGunpoint: {
		special: specialBuildGunpoint,
		cost:    10,
		icon:    assets.ImageActionBuildTurret,
	},

	specialIncreaseRadius: {
		special: specialIncreaseRadius,
		cost:    15,
		icon:    assets.ImageActionIncreaseRadius,
	},
	specialDecreaseRadius: {
		special: specialDecreaseRadius,
		cost:    4,
		icon:    assets.ImageActionDecreaseRadius,
	},

	specialSendCreeps: {
		special: specialSendCreeps,
		cost:    15,
		icon:    assets.ImageActionSendCreeps,
	},
	specialRally: {
		special: specialRally,
		cost:    25,
		icon:    assets.ImageActionRally,
	},
	specialSpawnCrawlers: {
		special: specialSpawnCrawlers,
		cost:    10,
		icon:    assets.ImageActionSpawnCrawlers,
	},
	specialBossAttack: {
		special: specialBossAttack,
		cost:    15,
		icon:    assets.ImageActionBossAttack,
	},
	specialIncreaseTech: {
		special: specialIncreaseTech,
		cost:    20,
		icon:    assets.ImageActionIncreaseTech,
	},
	specialAtomicBomb: {
		special: specialAtomicBomb,
		cost:    20,
		icon:    assets.ImageActionAbomb,
	},
}

type creepOptionInfo struct {
	special      specialChoiceKind
	stats        *gamedata.CreepStats
	minTechLevel float64
	maxUnits     int
	cooldown     float64
}

var creepOptionInfoList = func() []creepOptionInfo {
	list := []creepOptionInfo{
		{
			maxUnits:     15,
			special:      specialBuyCrawlers,
			minTechLevel: 0,
			stats:        gamedata.CrawlerCreepStats,
		},
		{
			maxUnits:     12,
			special:      specialBuyWanderers,
			minTechLevel: 0,
			stats:        gamedata.WandererCreepStats,
		},
		{
			maxUnits:     11,
			special:      specialBuyEliteCrawlers,
			minTechLevel: 0.1,
			stats:        gamedata.EliteCrawlerCreepStats,
		},
		{
			maxUnits:     8,
			special:      specialBuyStunners,
			minTechLevel: 0.2,
			stats:        gamedata.StunnerCreepStats,
		},
		{
			maxUnits:     10,
			special:      specialBuyStealthCrawlers,
			minTechLevel: 0.3,
			stats:        gamedata.StealthCrawlerCreepStats,
		},
		{
			maxUnits:     8,
			special:      specialBuyHeavyCrawlers,
			minTechLevel: 0.4,
			stats:        gamedata.HeavyCrawlerCreepStats,
		},
		{
			maxUnits:     3,
			special:      specialBuyBuilders,
			minTechLevel: 0.4,
			stats:        gamedata.BuilderCreepStats,
		},
		{
			maxUnits:     2,
			special:      specialBuyCenturions,
			minTechLevel: 0.4,
			stats:        gamedata.CenturionCreepStats,
		},
		{
			maxUnits:     5,
			special:      specialBuyTemplars,
			minTechLevel: 0.5,
			stats:        gamedata.TemplarCreepStats,
		},
		{
			maxUnits:     5,
			special:      specialBuyAssaults,
			minTechLevel: 0.6,
			stats:        gamedata.AssaultCreepStats,
		},
		{
			maxUnits:     2,
			special:      specialBuyDominator,
			minTechLevel: 0.8,
			stats:        gamedata.DominatorCreepStats,
		},
		{
			maxUnits:     1,
			special:      specialBuyHowitzer,
			minTechLevel: 1.0,
			stats:        gamedata.HowitzerCreepStats,
		},
	}

	for i := range list {
		e := &list[i]
		cooldown := float64(creepCost(e.stats, false)) * (float64(e.maxUnits) * 0.4) * 0.75
		switch {
		case e.stats.Kind == gamedata.CreepCenturion:
			cooldown *= 3.0
		case e.stats.Kind == gamedata.CreepHowitzer:
			cooldown *= 1.5
		case e.stats.Kind == gamedata.CreepDominator:
			cooldown *= 1.2
		case e.stats.Kind == gamedata.CreepBuilder:
			cooldown *= 0.85
		case e.stats.Kind == gamedata.CreepStunner:
			cooldown *= 0.9
		case !e.stats.Flying:
			cooldown *= 0.75
		}
		e.cooldown = cooldown
	}

	return list
}()

func creepCardID(k specialChoiceKind) int {
	return int(k-_creepCardFirst) - 1
}

var choiceOptionList = []choiceOption{
	{
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.2},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.2},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: prioritySecurity, value: 0.2},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: priorityEvolution, value: 0.2},
		},
	},

	{
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: priorityGrowth, value: 0.15},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: prioritySecurity, value: 0.15},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.15},
			{priority: prioritySecurity, value: 0.15},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
	{
		effects: []choiceOptionEffect{
			{priority: prioritySecurity, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
}

type choiceState int

const (
	choiceCharging choiceState = iota
	choiceReady
)

type choiceGenerator struct {
	targetValue float64
	value       float64

	world *worldState

	state choiceState

	player player

	shuffledOptions []choiceOption

	beforeSpecialShuffle int
	specialOptionIndex   int
	buildTurret          bool
	increaseRadius       bool
	spawnCrawlers        bool
	specialChoiceKinds   []specialChoiceKind

	forcedSpecialChoice specialChoiceKind

	creepsState *creepsPlayerState

	EventChoiceReady    gsignal.Event[choiceSelection]
	EventChoiceSelected gsignal.Event[selectedChoice]
}

func newChoiceGenerator(world *worldState, creepsState *creepsPlayerState) *choiceGenerator {
	g := &choiceGenerator{
		world:       world,
		creepsState: creepsState,
	}

	if creepsState != nil {
		g.shuffledOptions = make([]choiceOption, 4)

		g.specialChoiceKinds = []specialChoiceKind{
			specialSendCreeps,
			specialBossAttack,
			specialRally,
			specialIncreaseTech,
		}
	} else {
		g.shuffledOptions = make([]choiceOption, len(choiceOptionList))
		copy(g.shuffledOptions, choiceOptionList)

		g.specialChoiceKinds = []specialChoiceKind{
			specialBuildColony,
			specialAttack,
			specialDecreaseRadius,
		}
	}

	return g
}

func (g *choiceGenerator) Init(scene *ge.Scene) {}

func (g *choiceGenerator) IsDisposed() bool { return false }

func (g *choiceGenerator) IsReady() bool {
	return g.state == choiceReady
}

func (g *choiceGenerator) Update(delta float64) {
	if g.state != choiceCharging {
		return
	}

	g.value += delta
	if g.value >= g.targetValue {
		g.generateChoices()
		return
	}
}

func (g *choiceGenerator) TryExecute(colony *colonyCoreNode, cardIndex int, pos gmath.Vec) bool {
	if g.creepsState != nil {
		return g.activateChoice(colony, cardIndex)
	}

	if colony.mode != colonyModeNormal {
		return false
	}
	if cardIndex != -1 {
		return g.activateChoice(colony, cardIndex)
	}
	return g.activateMoveChoice(colony, pos)
}

func (g *choiceGenerator) activateMoveChoice(colony *colonyCoreNode, pos gmath.Vec) bool {
	g.EventChoiceSelected.Emit(selectedChoice{
		Colony: colony,
		Index:  -1,
		Option: choiceOption{special: specialChoiceMoveColony},
		Pos:    pos,
		Player: g.player,
	})
	return true
}

func (g *choiceGenerator) activateChoice(colony *colonyCoreNode, i int) bool {
	if g.state != choiceReady {
		return false
	}

	choice := selectedChoice{
		Colony:  colony,
		Faction: gamedata.FactionTag(i + 1),
		Index:   i,
		Player:  g.player,
	}
	cooldown := 10.0
	if i == 4 {
		// A special action is selected.
		g.forcedSpecialChoice = specialChoiceNone
		choice.Option = specialChoicesTable[g.specialOptionIndex]
		cooldown = choice.Option.cost
		if choice.Option.special == specialIncreaseTech {
			div := 0.6 + (0.1 * float64(g.world.config.TechProgressRate))
			cooldown *= (1.0 + 1.75*g.world.creepsPlayerState.techLevel)
			cooldown /= div
		}
	} else {
		if g.creepsState != nil {
			info := creepOptionInfoList[creepCardID(g.shuffledOptions[i].special)]
			extraTech := g.world.creepsPlayerState.techLevel - info.minTechLevel
			multiplier := 1.0
			if extraTech > 0 {
				multiplier = gmath.ClampMin(1.0-(extraTech*0.25), 0.75)
			}
			cooldown = (g.shuffledOptions[i].cost * multiplier)
		}
		choice.Option = g.shuffledOptions[i]
	}
	choice.Cooldown = cooldown

	g.startCharging(cooldown)

	g.EventChoiceSelected.Emit(choice)
	return true
}

func (g *choiceGenerator) startCharging(targetValue float64) {
	g.value = 0
	g.targetValue = targetValue
	g.state = choiceCharging
}

func (g *choiceGenerator) generateChoicesForCreeps() {
	techLevel := g.creepsState.techLevel
	maxIndexAvailable := len(creepOptionInfoList) - 1
	for i := 0; i < len(creepOptionInfoList); i++ {
		info := creepOptionInfoList[i]
		if info.minTechLevel > (techLevel + gmath.Epsilon) {
			maxIndexAvailable = i - 1
			break
		}
	}

	const numCards = (_creepCardLast - _creepCardFirst) - 1
	const numDirections = 4
	var combinationsSet [numCards][numDirections]bool
	for i := range g.shuffledOptions {
		for {
			creepIndex := g.world.rand.IntRange(0, maxIndexAvailable)
			cardID := creepCardID(creepOptionInfoList[creepIndex].special)
			dir := g.world.rand.IntRange(0, 3)
			if combinationsSet[cardID][dir] {
				continue
			}
			o := &g.shuffledOptions[i]
			o.special = creepOptionInfoList[creepIndex].special
			o.icon = creepOptionInfoList[creepIndex].stats.Image
			o.cost = creepOptionInfoList[creepIndex].cooldown
			o.direction = dir
			combinationsSet[cardID][dir] = true
			break
		}
	}
}

func (g *choiceGenerator) GetChoices() choiceSelection {
	return choiceSelection{
		cards:   g.shuffledOptions[:4],
		special: specialChoicesTable[g.specialOptionIndex],
	}
}

func (g *choiceGenerator) ForceSpecialChoice(kind specialChoiceKind) {
	g.forcedSpecialChoice = kind
	if g.state == choiceReady {
		g.generateChoices()
	}
}

func (g *choiceGenerator) prepareChoiceOptions() {
	if g.creepsState != nil {
		g.generateChoicesForCreeps()
	} else {
		gmath.Shuffle(g.world.rand, g.shuffledOptions)
	}

	if g.beforeSpecialShuffle == 0 {
		g.spawnCrawlers = !g.spawnCrawlers
		g.buildTurret = !g.buildTurret
		g.increaseRadius = !g.increaseRadius
		gmath.Shuffle(g.world.rand, g.specialChoiceKinds)
		g.beforeSpecialShuffle = len(g.specialChoiceKinds)
	}
	g.beforeSpecialShuffle--
	specialIndex := g.beforeSpecialShuffle

	if g.forcedSpecialChoice != specialChoiceNone {
		g.specialOptionIndex = int(g.forcedSpecialChoice)
		return
	}
	specialOptionKind := g.specialChoiceKinds[specialIndex]
	switch specialOptionKind {
	case specialRally:
		if g.spawnCrawlers {
			if g.creepsState.techLevel >= 1.5 {
				specialOptionKind = specialAtomicBomb
			} else {
				specialOptionKind = specialSpawnCrawlers
			}
		}
	case specialBuildColony:
		if g.buildTurret {
			specialOptionKind = specialBuildGunpoint
		}
	case specialDecreaseRadius:
		if g.increaseRadius {
			specialOptionKind = specialIncreaseRadius
		}
	}
	g.specialOptionIndex = int(specialOptionKind)
}

func (g *choiceGenerator) generateChoices() {
	g.state = choiceReady
	g.prepareChoiceOptions()
	g.EventChoiceReady.Emit(g.GetChoices())
}
