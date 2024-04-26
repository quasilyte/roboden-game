package staging

import (
	"math"
	"strings"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
)

// This tutorial system is not very elegant.
// Instead of having an event-based system it can subscribe to,
// it has to query the game state and compare it with its expectations.
// To avoid too much redundant computations, we only do that once in a while
// with a randomized jitter.
// Also, the tutorial objects can't describe the interactive hints
// in a declarative way, so we'll have to hardcode every one of
// them here in the most adhoc way possible.

type tutorialManager struct {
	input *gameinput.Handler

	scene *ge.Scene

	messageManager *messageManager

	choice selectedChoice

	world        *worldState
	config       *gamedata.LevelConfig
	tutorialStep int
	stepTicks    int

	targetPos ge.Pos
	drone     *colonyAgentNode
	creep     *creepNode

	explainedResourcePool     bool
	explainedBaseConstruction bool
	explainedSecondBase       bool
	explainedFighter          bool
	explainedDestroyer        bool

	attackCountdown float64
	attackNum       int

	nextPressed bool

	alreadyMoved   bool
	nearReasources bool

	waitingForEnterKey bool
	enterKeyTimer      float64

	updateDelay float64

	EventShowRecipeTab        gsignal.Event[gsignal.Void]
	EventEnableChoices        gsignal.Event[gsignal.Void]
	EventEnableSpecialChoices gsignal.Event[gsignal.Void]
	EventForceSpecialChoice   gsignal.Event[specialChoiceKind]
	EventTriggerVictory       gsignal.Event[gsignal.Void]
}

func newTutorialManager(h *gameinput.Handler, world *worldState, messageManager *messageManager) *tutorialManager {
	return &tutorialManager{
		input:           h,
		world:           world,
		config:          world.config,
		updateDelay:     2,
		messageManager:  messageManager,
		attackCountdown: 5 * 60,
	}
}

func (m *tutorialManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.messageManager.EventMainMessageClicked.Connect(nil, func(gsignal.Void) {
		m.OnNextPressed()
	})
}

func (m *tutorialManager) IsDisposed() bool {
	return false
}

func (m *tutorialManager) waitForEnter() {
	m.nextPressed = false
	m.waitingForEnterKey = true
	m.enterKeyTimer = 0
}

func (m *tutorialManager) OnNextPressed() {
	m.nextPressed = true
	m.waitingForEnterKey = false
	m.runUpdateFunc()
}

func (m *tutorialManager) Update(delta float64) {
	if m.drone != nil && m.drone.IsDisposed() {
		m.drone = nil
	}
	if m.creep != nil && m.creep.IsDisposed() {
		m.creep = nil
	}

	m.enterKeyTimer += delta

	m.attackCountdown = gmath.ClampMin(m.attackCountdown-delta, 0)
	if m.attackCountdown == 0 {
		m.attackCountdown = m.world.rand.FloatRange(3.0*60, 5*60)
		m.spawnAttack()
	}

	m.updateDelay = gmath.ClampMin(m.updateDelay-delta, 0)
	if m.updateDelay != 0 {
		return
	}
	m.updateDelay = m.scene.Rand().FloatRange(0.75, 1.25)

	m.stepTicks = gmath.ClampMin(m.stepTicks-1, 0)

	m.choice = selectedChoice{}
	m.runUpdateFunc()
}

func (m *tutorialManager) spawnAttack() {
	m.attackNum++
	var creeps []arenaWaveUnit
	if m.world.rand.Bool() {
		numCreeps := m.world.rand.IntRange(3, 5)
		for i := 0; i < numCreeps; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.WandererCreepStats})
		}
	} else {
		numCreeps := m.world.rand.IntRange(5, 7)
		for i := 0; i < numCreeps; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.EliteCrawlerCreepStats})
		}
	}
	if m.attackNum >= 3 {
		if m.world.rand.Bool() {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.AssaultCreepStats})
		} else {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.BuilderCreepStats})
		}
	}
	creeps = append(creeps, arenaWaveUnit{stats: gamedata.StunnerCreepStats})
	sendCreeps(m.world, arenaWaveGroup{
		side:  m.world.rand.IntRange(0, 3),
		units: creeps,
	})
}

func (m *tutorialManager) runUpdateFunc() {
	if len(m.world.allColonies) == 0 {
		return
	}

	if m.waitingForEnterKey && m.hint() != nil && !m.hint().highlight {
		if m.enterKeyTimer > (30 * 1.2) {
			m.hint().Highlight()
		}
	}

	hintOpen := m.hint() != nil
	if m.maybeCompleteStep() {
		m.tutorialStep++
		if hintOpen && m.hint() != nil {
			m.hint().Dispose()
		}
	}
}

func (m *tutorialManager) OnChoice(choice selectedChoice) {
	m.choice = choice
	m.runUpdateFunc()
	if choice.Option.special == specialChoiceMoveColony {
		m.alreadyMoved = true
	}
}

func (m *tutorialManager) hint() *messageNode {
	return m.messageManager.mainMessage
}

func (m *tutorialManager) explainDrone(drone *colonyAgentNode, textKey string) {
	m.messageManager.AddMessage(queuedMessageInfo{
		targetPos:     ge.Pos{Base: &drone.pos, Offset: gmath.Vec{Y: -4}},
		text:          m.scene.Dict().Get(textKey),
		trackedObject: drone,
		timer:         20,
		onReady: func() {
			if drone.IsDisposed() {
				return
			}
			drone.AssignMode(agentModePosing, gmath.Vec{X: 15}, nil)
		},
	})
}

func (m *tutorialManager) maybeCompleteStep() bool {
	if m.choice.Option.special == specialChoiceMoveColony {
		m.alreadyMoved = true
	}
	if !m.nearReasources {
		colony := m.world.allColonies[0]
		if !colony.waypoint.IsZero() {
			// Maybe the colony was already on the move?
			for _, res := range m.world.essenceSources {
				if res.pos.DistanceSquaredTo(colony.relocationPoint) < (128 * 128) {
					m.nearReasources = true
					break
				}
			}
		} else {
			// Maybe it's already around some resources.
			for _, res := range m.world.essenceSources {
				if res.pos.DistanceSquaredTo(colony.pos) < (128 * 128) {
					m.nearReasources = true
					break
				}
			}
		}
	}

	if !m.explainedResourcePool && m.world.allColonies[0].resources > 120 {
		m.explainedResourcePool = true
		m.messageManager.AddMessage(queuedMessageInfo{
			targetPos:     ge.Pos{Base: &m.world.allColonies[0].pos, Offset: gmath.Vec{X: -3, Y: 18}},
			trackedObject: m.world.allColonies[0],
			text:          m.scene.Dict().Get("tutorial.context.resource_bar"),
			timer:         25,
		})
	}

	if !m.explainedBaseConstruction && len(m.world.constructions) != 0 {
		var colonyConstruction *constructionNode
		for _, c := range m.world.constructions {
			if c.stats == colonyCoreConstructionStats {
				colonyConstruction = c
				break
			}
		}
		if colonyConstruction != nil {
			m.explainedBaseConstruction = true
			m.messageManager.AddMessage(queuedMessageInfo{
				targetPos:     ge.Pos{Base: &colonyConstruction.pos, Offset: gmath.Vec{Y: 6}},
				trackedObject: colonyConstruction,
				text:          m.scene.Dict().Get("tutorial.context.colony_construction"),
				timer:         25,
			})
		}
	}

	if !m.explainedSecondBase && len(m.world.allColonies) > 1 {
		m.explainedSecondBase = true
		pos := ge.Pos{}
		if m.input.InputMethod == gameinput.InputMethodTouch {
			pos = ge.Pos{Offset: gmath.Vec{X: 28, Y: 480}}
		}
		m.messageManager.AddMessage(queuedMessageInfo{
			text:      m.scene.Dict().Get("tutorial.context.second_base", m.world.inputMode),
			timer:     25,
			targetPos: pos,
		})
	}

	if !m.explainedDestroyer || !m.explainedFighter {
		var fighter *colonyAgentNode
		var destroyer *colonyAgentNode
		for _, c := range m.world.allColonies {
			c.agents.Each(func(a *colonyAgentNode) {
				switch a.stats.Kind {
				case gamedata.AgentFighter:
					fighter = a
				case gamedata.AgentDestroyer:
					destroyer = a
				}
			})
		}
		if !m.explainedFighter && fighter != nil {
			m.explainedFighter = true
			m.explainDrone(fighter, "tutorial.context.fighter_drone")
		} else if !m.explainedDestroyer && destroyer != nil {
			m.explainedDestroyer = true
			m.explainDrone(destroyer, "tutorial.context.destroyer_drone")
		}
	}

	d := m.scene.Dict()

	switch m.tutorialStep {
	case 0:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.greeting"))
		m.waitForEnter()
		return true
	case 1:
		return m.nextPressed

	case 2:
		m.targetPos = m.findResourceStash(300)
		m.addHintNode(m.targetPos, d.Get("tutorial.camera", m.world.inputMode))
		m.waitForEnter()
		return true

	case 3:
		return m.nextPressed

	case 4:
		if !m.nearReasources || !m.alreadyMoved {
			m.addHintNode(ge.Pos{}, d.Get("tutorial.move", m.world.inputMode))
		}
		return true

	case 5:
		if m.nearReasources && m.alreadyMoved {
			return true
		}
		if m.choice.Option.special != specialChoiceMoveColony {
			return false
		}
		for _, res := range m.world.essenceSources {
			if res.pos.DistanceSquaredTo(m.choice.Pos) < (128 * 128) {
				return true
			}
		}
		return false

	case 6:
		return !m.world.allColonies[0].IsFlying()

	case 7:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.resources"))
		m.waitForEnter()
		return true
	case 8:
		return m.nextPressed

	case 9:
		m.stepTicks = 20
		return true
	case 10:
		return m.stepTicks == 0

	case 11:
		m.addScreenHintNode(gmath.Vec{X: (m.scene.Context().ScreenWidth - 148) + (36 * 0), Y: 516}, d.Get("tutorial.priorities", m.world.inputMode))
		m.waitForEnter()
		return true
	case 12:
		return m.nextPressed

	case 13:
		s := m.input.ReplaceKeyNames(d.Get("tutorial.enable_choices", m.world.inputMode))
		m.addHintNode(ge.Pos{}, s)
		m.EventEnableChoices.Emit(gsignal.Void{})
		return true
	case 14:
		return len(m.choice.Option.effects) != 0

	case 15:
		m.stepTicks = 10
		return true
	case 16:
		return m.stepTicks == 0

	case 17:
		var creeps []arenaWaveUnit
		for i := 0; i < 7; i++ {
			super := i == 0
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.WandererCreepStats, super: super})
		}
		sendCreeps(m.world, arenaWaveGroup{
			side:  m.world.rand.IntRange(0, 3),
			units: creeps,
		})
		for _, creep := range m.world.creeps {
			if creep.super {
				m.creep = creep
				break
			}
		}
		m.addHintNode(ge.Pos{Base: &m.creep.pos}, d.Get("tutorial.enemy_scouts"))
		return true
	case 18:
		return m.creep == nil

	case 19:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.factions"))
		return true
	case 20:
		return len(m.choice.Option.effects) != 0

	case 21:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.factions2", m.world.inputMode))
		m.waitForEnter()
		m.EventShowRecipeTab.Emit(gsignal.Void{})
		return true
	case 22:
		return m.nextPressed

	case 23:
		m.stepTicks = 20
		return true
	case 24:
		return m.stepTicks == 0

	case 25:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.build_turret", m.world.inputMode))
		m.EventEnableSpecialChoices.Emit(gsignal.Void{})
		m.EventForceSpecialChoice.Emit(specialBuildGunpoint)
		m.stepTicks = 70
		return true
	case 26:
		foundTurret := false
		if m.stepTicks <= 40 {
			for _, c := range m.world.constructions {
				if c.stats.Kind == constructTurret {
					foundTurret = true
					break
				}
			}
			if !foundTurret {
				for _, c := range m.world.allColonies {
					if len(c.turrets) != 0 {
						foundTurret = true
						break
					}
				}
			}
		}
		return m.stepTicks == 0 || foundTurret

	case 27:
		var creeps []arenaWaveUnit
		for i := 0; i < 5; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.CrawlerCreepStats})
		}
		for i := 0; i < 3; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.EliteCrawlerCreepStats})
		}
		for i := 0; i < 2; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.HeavyCrawlerCreepStats})
		}
		spawnPos := sendCreeps(m.world, arenaWaveGroup{
			side:  3,
			units: creeps,
		})
		m.addHintNode(ge.Pos{Offset: spawnPos}, d.Get("tutorial.crawlers_attack"))
		m.stepTicks = 15
		return true
	case 28:
		return m.stepTicks == 0

	case 29:
		m.stepTicks = 40
		return true
	case 30:
		return len(m.world.creeps) == 0 || m.stepTicks == 0

	case 31:
		var creeps []arenaWaveUnit
		for i := 0; i < 3; i++ {
			super := i == 0
			creeps = append(creeps, arenaWaveUnit{
				super: super,
				stats: gamedata.BuilderCreepStats,
			})
		}
		spawnPos := sendCreeps(m.world, arenaWaveGroup{
			side:  m.world.rand.IntRange(0, 3),
			units: creeps,
		})

		m.addHintNode(ge.Pos{Offset: spawnPos}, d.Get("tutorial.builders_attack"))
		m.stepTicks = 25
		return true
	case 32:
		return m.stepTicks == 0

	case 33:
		for _, creep := range m.world.creeps {
			if creep.super && creep.stats.Kind == gamedata.CreepBuilder {
				m.creep = creep
				break
			}
		}
		m.stepTicks = 40
		return true

	case 34:
		return m.stepTicks == 0 || m.creep == nil

	case 35:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.final_attack_warning"))
		m.stepTicks = 40
		return true
	case 36:
		return m.stepTicks == 0

	case 37:
		m.stepTicks = 280
		return true

	case 38:
		return m.stepTicks == 0

	case 39:
		var creeps []arenaWaveUnit
		for i := 0; i < 6; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.WandererCreepStats})
		}
		for i := 0; i < 3; i++ {
			creeps = append(creeps, arenaWaveUnit{stats: gamedata.StunnerCreepStats})
		}
		creeps = append(creeps, arenaWaveUnit{stats: gamedata.BuilderCreepStats})
		creeps = append(creeps, arenaWaveUnit{stats: gamedata.HowitzerCreepStats})
		spawnPos := sendCreeps(m.world, arenaWaveGroup{
			side:  m.world.rand.IntRange(0, 3),
			units: creeps,
		})
		m.addHintNode(ge.Pos{Offset: spawnPos}, d.Get("tutorial.final_attack"))
		m.stepTicks = 15
		return true
	case 40:
		return m.stepTicks == 0

	case 41:
		m.stepTicks = 25
		return true
	case 42:
		return m.stepTicks == 0

	case 43:
		var howitzer *creepNode
		for _, creep := range m.world.creeps {
			if creep.stats.Kind == gamedata.CreepHowitzer {
				howitzer = creep
				break
			}
		}
		m.creep = howitzer
		return true

	case 44:
		if m.creep == nil {
			return true
		}
		m.addHintNode(ge.Pos{Base: &m.creep.pos}, d.Get("tutorial.final_goal"))
		return true
	case 45:
		return m.creep == nil

	case 46:
		m.addHintNode(ge.Pos{}, d.Get("tutorial.final_message"))
		m.waitForEnter()
		return true
	case 47:
		return m.nextPressed

	case 48:
		m.EventTriggerVictory.Emit(gsignal.Void{})
	}

	return false
}

func (m *tutorialManager) processMessageText(s string) string {
	switch m.world.inputMode {
	case "touch":
		d := m.scene.Dict()
		return strings.Replace(s, d.Get("tutorial.next.keyboard"), d.Get("tutorial.next.touch"), 1)
	case "gamepad":
		d := m.scene.Dict()
		return strings.Replace(s, d.Get("tutorial.next.keyboard"), d.Get("tutorial.next.gamepad"), 1)
	default:
		return s // Do nothing
	}
}

func (m *tutorialManager) addScreenHintNode(targetPos gmath.Vec, msg string) {
	m.messageManager.SetMainMessage(queuedMessageInfo{
		text:      m.processMessageText(msg),
		targetPos: ge.Pos{Offset: targetPos},
	})
}

func (m *tutorialManager) addHintNode(targetPos ge.Pos, msg string) {
	m.messageManager.SetMainMessage(queuedMessageInfo{
		text:          m.processMessageText(msg),
		targetPos:     targetPos,
		forceWorldPos: true,
	})
}

func (m *tutorialManager) findResourceStash(minDist float64) ge.Pos {
	var pos gmath.Vec
	closestDist := math.MaxFloat64
	colony := m.world.allColonies[0]

	for _, res := range m.world.essenceSources {
		if res.stats.scrap || res.stats == redOilSource || res.stats == redCrystalSource {
			continue
		}
		dist := res.pos.DistanceTo(colony.pos)
		if dist < minDist {
			continue
		}
		if dist < closestDist {
			for _, lava := range m.world.lavaPuddles {
				if lava.CollidesWith(res.pos, 64) {
					continue
				}
			}
			closestDist = dist
			pos = res.pos
		}
	}

	pos = pos.DirectionTo(colony.pos).Mulf(40).Add(pos)
	return ge.Pos{Offset: pos}
}
