package staging

import (
	"fmt"
	"math"

	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
)

const (
	coreFlightHeight float64 = 50
	maxUpkeepValue   int     = 270
)

type colonyCoreMode int

const (
	colonyModeNormal colonyCoreMode = iota
	colonyModeTakeoff
	colonyModeRelocating
	colonyModeLanding
)

var colonyResourceRectOffsets = []float64{
	18,
	9,
	-1,
}

var pixelsPerResourceRect = []float64{
	4,
	5,
	6,
}

type colonyCoreNode struct {
	sprite       *ge.Sprite
	hatch        *ge.Sprite
	flyingSprite *ge.Sprite
	shadow       *ge.Sprite
	upkeepBar    *ge.Sprite

	scene *ge.Scene

	pos       gmath.Vec
	spritePos gmath.Vec
	height    float64
	maxHealth float64
	health    float64

	mode colonyCoreMode

	waypoint        gmath.Vec
	relocationPoint gmath.Vec

	resourceShortage int
	resources        resourceContainer
	world            *worldState

	agents       []*colonyAgentNode
	combatAgents []*colonyAgentNode

	hasRedMiner              bool
	numServoAgents           int
	availableAgents          []*colonyAgentNode
	availableCombatAgents    []*colonyAgentNode
	availableUniversalAgents []*colonyAgentNode

	planner *colonyActionPlanner

	openHatchTime float64

	realRadius float64

	upkeepDelay float64

	actionDelay      float64
	actionPriorities *weightContainer[colonyPriority]

	resourceRects []*ge.Rect

	factionTagPicker *gmath.RandPicker[factionTag]

	factionWeights *weightContainer[factionTag]

	EventDestroyed gsignal.Event[*colonyCoreNode]
}

type colonyConfig struct {
	Pos gmath.Vec

	Radius float64

	World *worldState
}

func newColonyCoreNode(config colonyConfig) *colonyCoreNode {
	c := &colonyCoreNode{
		world:                    config.World,
		realRadius:               config.Radius,
		agents:                   make([]*colonyAgentNode, 0, 32),
		availableAgents:          make([]*colonyAgentNode, 0, 32),
		combatAgents:             make([]*colonyAgentNode, 0, 20),
		availableCombatAgents:    make([]*colonyAgentNode, 0, 20),
		availableUniversalAgents: make([]*colonyAgentNode, 0, 20),
		maxHealth:                100,
	}
	c.actionPriorities = newWeightContainer(priorityResources, priorityGrowth, priorityEvolution, prioritySecurity)
	c.factionWeights = newWeightContainer(neutralFactionTag, yellowFactionTag, redFactionTag, greenFactionTag, blueFactionTag)
	c.factionWeights.SetWeight(neutralFactionTag, 1.0)
	c.pos = config.Pos
	return c
}

func (c *colonyCoreNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.factionTagPicker = gmath.NewRandPicker[factionTag](scene.Rand())

	c.planner = newColonyActionPlanner(c, scene.Rand())

	c.health = c.maxHealth

	c.sprite = scene.NewSprite(assets.ImageColonyCore)
	c.sprite.Pos.Base = &c.spritePos
	c.sprite.Shader = scene.NewShader(assets.ShaderColonyDamage)
	c.sprite.Shader.SetFloatValue("HP", 1.0)
	c.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageColonyDamageMask)
	c.world.camera.AddGraphicsBelow(c.sprite)

	c.flyingSprite = scene.NewSprite(assets.ImageColonyCoreFlying)
	c.flyingSprite.Pos.Base = &c.spritePos
	c.flyingSprite.Visible = false
	c.flyingSprite.Shader = c.sprite.Shader
	c.world.camera.AddGraphics(c.flyingSprite)

	c.hatch = scene.NewSprite(assets.ImageColonyCoreHatch)
	c.hatch.Pos.Base = &c.spritePos
	c.hatch.Pos.Offset.Y = -20
	c.world.camera.AddGraphics(c.hatch)

	c.upkeepBar = scene.NewSprite(assets.ImageUpkeepBar)
	c.upkeepBar.Pos.Base = &c.spritePos
	c.upkeepBar.Pos.Offset.Y = -5
	c.upkeepBar.Pos.Offset.X = -c.upkeepBar.ImageWidth() * 0.5
	c.upkeepBar.Centered = false
	c.world.camera.AddGraphicsBelow(c.upkeepBar)
	c.updateUpkeepBar(0)

	c.shadow = scene.NewSprite(assets.ImageColonyCoreShadow)
	c.shadow.Pos.Base = &c.spritePos
	c.shadow.Visible = false
	c.world.camera.AddGraphicsBelow(c.shadow)

	c.resourceRects = make([]*ge.Rect, 3)
	for i := range c.resourceRects {
		rect := ge.NewRect(scene.Context(), 6, pixelsPerResourceRect[i])
		rect.Centered = false
		cscale := 0.6 + (0.2 * float64(i))
		rect.FillColorScale.SetRGBA(uint8(float64(0xd6)*cscale), uint8(float64(0x85)*cscale), uint8(float64(0x43)*cscale), 200)
		rect.Pos.Base = &c.spritePos
		rect.Pos.Offset.X -= 3
		rect.Pos.Offset.Y = colonyResourceRectOffsets[i]
		c.resourceRects[i] = rect
		c.world.camera.AddGraphics(rect)
	}
}

func (c *colonyCoreNode) MaxFlyDistance() float64 {
	return 180.0 + (float64(c.numServoAgents) * 30.0)
}

func (c *colonyCoreNode) PatrolRadius() float64 {
	return c.realRadius * (1.0 + c.GetSecurityPriority()*0.25)
}

func (c *colonyCoreNode) GetPos() *gmath.Vec { return &c.pos }

func (c *colonyCoreNode) GetVelocity() gmath.Vec {
	switch c.mode {
	case colonyModeTakeoff, colonyModeRelocating, colonyModeLanding:
		return c.pos.VecTowards(c.waypoint, c.movementSpeed())
	default:
		return gmath.Vec{}
	}
}

func (c *colonyCoreNode) OnHeal(amount float64) {
	if c.health == c.maxHealth {
		return
	}
	c.health = gmath.ClampMax(c.health+amount, c.maxHealth)
	c.updateHealthShader()
}

func (c *colonyCoreNode) OnDamage(damage damageValue, source gmath.Vec) {
	c.health -= damage.health
	if c.health < 0 {
		createAreaExplosion(c.scene, c.world.camera, spriteRect(c.pos, c.sprite))
		c.Destroy()
		return
	}

	c.updateHealthShader()
	if c.scene.Rand().Chance(0.7) {
		c.actionPriorities.AddWeight(prioritySecurity, 0.02)
	}
}

func (c *colonyCoreNode) Destroy() {
	for _, agent := range c.agents {
		agent.OnDamage(damageValue{health: 1000}, gmath.Vec{})
	}

	c.EventDestroyed.Emit(c)
	c.Dispose()
}

func (c *colonyCoreNode) GetEntrancePos() gmath.Vec {
	return c.pos.Add(gmath.Vec{X: -1, Y: -20})
}

func (c *colonyCoreNode) GetStoragePos() gmath.Vec {
	return c.pos.Add(gmath.Vec{X: 1, Y: 0})
}

func (c *colonyCoreNode) GetResourcePriority() float64 {
	return c.actionPriorities.GetWeight(priorityResources)
}

func (c *colonyCoreNode) GetGrowthPriority() float64 {
	return c.actionPriorities.GetWeight(priorityGrowth)
}

func (c *colonyCoreNode) GetEvolutionPriority() float64 {
	return c.actionPriorities.GetWeight(priorityEvolution)
}

func (c *colonyCoreNode) GetSecurityPriority() float64 {
	return c.actionPriorities.GetWeight(prioritySecurity)
}

func (c *colonyCoreNode) CloneAgentNode(a *colonyAgentNode) *colonyAgentNode {
	pos := a.pos.Add(c.scene.Rand().Offset(-4, 4))
	cloned := a.Clone()
	cloned.pos = pos
	c.AcceptAgent(cloned)
	return cloned
}

func (c *colonyCoreNode) NewColonyAgentNode(stats *agentStats, pos gmath.Vec) *colonyAgentNode {
	a := newColonyAgentNode(c, stats, pos)
	c.AcceptAgent(a)
	return a
}

func (c *colonyCoreNode) DetachAgent(a *colonyAgentNode) {
	a.EventDestroyed.Disconnect(c)
	if a.stats.canPatrol {
		c.combatAgents = xslices.Remove(c.combatAgents, a)
	} else {
		c.agents = xslices.Remove(c.agents, a)
	}
}

func (c *colonyCoreNode) AcceptAgent(a *colonyAgentNode) {
	a.EventDestroyed.Connect(c, func(x *colonyAgentNode) {
		if x.stats.canPatrol {
			c.combatAgents = xslices.Remove(c.combatAgents, x)
		} else {
			c.agents = xslices.Remove(c.agents, x)
		}
	})
	if a.stats.canPatrol {
		c.combatAgents = append(c.combatAgents, a)
	} else {
		c.agents = append(c.agents, a)
	}
	a.colonyCore = c
}

func (c *colonyCoreNode) NumAgents() int {
	return len(c.agents) + len(c.combatAgents)
}

func (c *colonyCoreNode) IsDisposed() bool { return c.sprite.IsDisposed() }

func (c *colonyCoreNode) Dispose() {
	c.sprite.Dispose()
	c.hatch.Dispose()
	c.flyingSprite.Dispose()
	c.shadow.Dispose()
	c.upkeepBar.Dispose()
	for _, rect := range c.resourceRects {
		rect.Dispose()
	}
}

func (c *colonyCoreNode) updateHealthShader() {
	percentage := c.health / c.maxHealth
	c.sprite.Shader.SetFloatValue("HP", percentage)
	c.sprite.Shader.Enabled = percentage < 0.95
}

func (c *colonyCoreNode) Update(delta float64) {
	// FIXME: this should be fixed in the ge package.
	c.spritePos.X = math.Round(c.pos.X)
	c.spritePos.Y = math.Round(c.pos.Y)

	c.updateResourceRects()

	if c.shadow.Visible {
		c.shadow.Pos.Offset.Y = c.height + 4
		newShadowAlpha := float32(1.0 - ((c.height / coreFlightHeight) * 0.5))
		c.shadow.SetAlpha(newShadowAlpha)
	}

	c.processUpkeep(delta)

	switch c.mode {
	case colonyModeTakeoff:
		c.updateTakeoff(delta)
	case colonyModeRelocating:
		c.updateRelocating(delta)
	case colonyModeLanding:
		c.updateLanding(delta)
	case colonyModeNormal:
		c.updateNormal(delta)
	}
}

func (c *colonyCoreNode) movementSpeed() float64 {
	switch c.mode {
	case colonyModeTakeoff, colonyModeLanding:
		return 8.0
	case colonyModeRelocating:
		return 16.0 + (float64(c.numServoAgents) * 3)
	default:
		return 0
	}
}

func (c *colonyCoreNode) updateUpkeepBar(upkeepValue int) {
	percentage := float64(upkeepValue) / float64(maxUpkeepValue)
	c.upkeepBar.FrameWidth = c.upkeepBar.ImageWidth() * percentage
}

func (c *colonyCoreNode) updateResourceRects() {
	const resourcesPerBlock float64 = 100.0
	unallocated := c.resources.Essence
	for i, rect := range c.resourceRects {
		var percentage float64
		if unallocated >= resourcesPerBlock {
			percentage = 1.0
		} else if unallocated <= 0 {
			percentage = 0
		} else {
			percentage = unallocated / resourcesPerBlock
		}
		unallocated -= resourcesPerBlock
		pixels := pixelsPerResourceRect[i]
		rect.Height = percentage * pixels
		rect.Pos.Offset.Y = colonyResourceRectOffsets[i] + (pixels - rect.Height)
		rect.Visible = rect.Height >= 1
	}
}

func (c *colonyCoreNode) calcUnitLimit() int {
	calculated := ((c.realRadius - 80) * 0.3) + 10
	return gmath.Clamp(int(calculated), 10, 128)
}

func (c *colonyCoreNode) calcUpkeed() (float64, int) {
	upkeepTotal := 0
	upkeepDecrease := 0
	c.FindAgent(func(a *colonyAgentNode) bool {
		if a.stats.kind == agentGenerator {
			upkeepDecrease++
		}
		upkeepTotal += a.stats.upkeep
		return false
	})
	upkeepDecrease = gmath.ClampMax(upkeepDecrease, 5)
	upkeepTotal = gmath.ClampMin(upkeepTotal-(upkeepDecrease*10), 0)
	var resourcePrice float64
	switch {
	case upkeepTotal <= 30:
		// 15 workers or ~7 militia
		resourcePrice = 0
	case upkeepTotal <= 45:
		// ~22 workers or ~11 militia
		resourcePrice = 0.5
	case upkeepTotal <= 70:
		// 35 workers or ~17 militia
		resourcePrice = 1.0
	case upkeepTotal <= 95:
		// ~47 workers or ~23 militia
		resourcePrice = 2.0
	case upkeepTotal <= 120:
		// ~60 workers or 30 militia
		resourcePrice = 3.0
	case upkeepTotal <= 150:
		// 75 workers or ~37 militia
		resourcePrice = 4.5
	case upkeepTotal <= 215:
		// ~107 workers or ~53 militia
		resourcePrice = 6.0
	case upkeepTotal < maxUpkeepValue:
		resourcePrice = 8.0
	default:
		resourcePrice = 12.0
	}
	return resourcePrice, upkeepTotal
}

func (c *colonyCoreNode) processUpkeep(delta float64) {
	c.upkeepDelay -= delta
	if c.upkeepDelay > 0 {
		return
	}
	c.upkeepDelay = c.scene.Rand().FloatRange(6.5, 8.5)
	upkeepPrice, upkeepValue := c.calcUpkeed()
	c.updateUpkeepBar(upkeepValue)
	if c.resources.Essence < upkeepPrice {
		c.actionPriorities.AddWeight(priorityResources, 0.04)
		c.resources.Essence = 0
	} else {
		c.resources.Essence -= upkeepPrice
	}
}

func (c *colonyCoreNode) doRelocation(pos gmath.Vec) {
	c.relocationPoint = pos

	c.FindAgent(func(a *colonyAgentNode) bool {
		a.payload = 0
		if a.height != agentFlightHeight {
			a.AssignMode(agentModeAlignStandby, gmath.Vec{}, nil)
		} else {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		}
		return false
	})

	c.mode = colonyModeTakeoff
	c.openHatchTime = 0
	c.shadow.Visible = true
	c.flyingSprite.Visible = true
	c.sprite.Visible = false
	c.hatch.Visible = false
	c.upkeepBar.Visible = false
	c.waypoint = c.pos.Sub(gmath.Vec{Y: coreFlightHeight})
}

func (c *colonyCoreNode) updateTakeoff(delta float64) {
	c.height += delta * c.movementSpeed()
	if c.moveTowards(delta, c.movementSpeed(), c.waypoint) {
		c.height = coreFlightHeight
		c.waypoint = c.relocationPoint.Sub(gmath.Vec{Y: coreFlightHeight})
		c.mode = colonyModeRelocating
	}
}

func (c *colonyCoreNode) updateRelocating(delta float64) {
	if c.moveTowards(delta, c.movementSpeed(), c.waypoint) {
		c.waypoint = c.relocationPoint
		c.mode = colonyModeLanding
	}
}

func (c *colonyCoreNode) updateLanding(delta float64) {
	c.height -= delta * c.movementSpeed()
	if c.moveTowards(delta, c.movementSpeed(), c.waypoint) {
		c.height = 0
		c.mode = colonyModeNormal
		c.flyingSprite.Visible = false
		c.shadow.Visible = false
		c.sprite.Visible = true
		c.hatch.Visible = true
		c.upkeepBar.Visible = true
		playSound(c.scene, c.world.camera, assets.AudioColonyLanded, c.pos)
	}
}

func (c *colonyCoreNode) updateNormal(delta float64) {
	c.actionDelay = gmath.ClampMin(c.actionDelay-delta, 0)
	if c.actionDelay == 0 {
		c.doAction()
	}
	c.openHatchTime = gmath.ClampMin(c.openHatchTime-delta, 0)
	c.hatch.Visible = c.openHatchTime == 0
}

func (c *colonyCoreNode) doAction() {
	if c.resourceShortage >= 5 {
		c.actionPriorities.AddWeight(priorityResources, c.scene.Rand().FloatRange(0.02, 0.05))
		c.resourceShortage -= 5
	}

	action := c.planner.PickAction()
	if action.Kind == actionNone {
		c.actionDelay = c.scene.Rand().FloatRange(0.15, 0.3)
		return
	}
	if c.tryExecutingAction(action) {
		c.actionDelay = c.scene.Rand().FloatRange(action.TimeCost*0.75, action.TimeCost*1.25)
	} else {
		c.actionDelay = c.scene.Rand().FloatRange(0.1, 0.2)
	}
}

func (c *colonyCoreNode) tryExecutingAction(action colonyAction) bool {
	switch action.Kind {
	case actionMineEssence:
		if len(c.availableAgents) == 0 && len(c.availableUniversalAgents) == 0 {
			return false
		}
		maxNumAgents := gmath.Clamp(len(c.availableAgents)/4, 2, 10)
		minNumAgents := gmath.Clamp(len(c.availableAgents)/10, 2, 6)
		toAssign := c.scene.Rand().IntRange(minNumAgents, maxNumAgents)
		extraAssign := gmath.Clamp(int(c.GetResourcePriority()*10)-1, 0, 10)
		toAssign += extraAssign
		c.pickWorkerUnits(toAssign, func(a *colonyAgentNode) {
			a.AssignMode(agentModeMineEssence, gmath.Vec{}, action.Value.(*essenceSourceNode))
		})
		return true

	case actionRepairBase:
		repairCost := 10.0
		ok := false
		c.pickWorkerUnits(1, func(a *colonyAgentNode) {
			if c.resources.Essence < repairCost {
				return
			}
			if a.AssignMode(agentModeRepairBase, gmath.Vec{}, nil) {
				c.resources.Essence -= repairCost
				ok = true
			}
		})
		return ok

	case actionBuildBase:
		sendCost := 3.0
		maxNumAgents := gmath.Clamp(len(c.availableAgents)/10, 1, 5)
		minNumAgents := gmath.Clamp(len(c.availableAgents)/15, 1, 3)
		toAssign := c.scene.Rand().IntRange(minNumAgents, maxNumAgents)
		c.pickWorkerUnits(toAssign, func(a *colonyAgentNode) {
			if c.resources.Essence < sendCost {
				return
			}
			if a.AssignMode(agentModeBuildBase, gmath.Vec{}, action.Value) {
				c.resources.Essence -= sendCost
			}
		})
		return true

	case actionRecycleAgent:
		a := action.Value.(*colonyAgentNode)
		a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
		return true

	case actionProduceAgent:
		a := c.NewColonyAgentNode(action.Value.(*agentStats), c.GetEntrancePos())
		a.height = 0
		a.faction = c.pickAgentFaction()
		c.scene.AddObject(a)
		c.resources.Essence -= a.stats.cost
		a.AssignMode(agentModeTakeoff, gmath.Vec{}, nil)
		playSound(c.scene, c.world.camera, assets.AudioAgentProduced, c.pos)
		c.openHatchTime = 1.5
		return true

	case actionGetReinforcements:
		wantWorkers := c.scene.Rand().IntRange(1, 3)
		wantWarriors := c.scene.Rand().IntRange(0, 2)
		transferUnit := func(dst, src *colonyCoreNode, a *colonyAgentNode) {
			src.DetachAgent(a)
			dst.AcceptAgent(a)
			a.AssignMode(agentModeAlignStandby, gmath.Vec{}, nil)
		}
		srcColony := action.Value.(*colonyCoreNode)
		srcColony.pickWorkerUnits(wantWorkers, func(a *colonyAgentNode) {
			transferUnit(c, srcColony, a)
		})
		srcColony.pickCombatUnits(wantWarriors, func(a *colonyAgentNode) {
			transferUnit(c, srcColony, a)
		})
		fmt.Println("transfered", wantWarriors, wantWorkers)
		return true

	case actionCloneAgent:
		cloneTarget := action.Value.(*colonyAgentNode)
		cloner := action.Value2.(*colonyAgentNode)
		c.resources.Essence -= agentCloningCost(c, cloner, cloneTarget)
		cloner.AssignMode(agentModeMakeClone, gmath.Vec{}, cloneTarget)
		cloneTarget.AssignMode(agentModeWaitCloning, gmath.Vec{}, cloner)
		return true

	case actionMergeAgents:
		agent1 := action.Value.(*colonyAgentNode)
		agent2 := action.Value2.(*colonyAgentNode)
		agent1.AssignMode(agentModeMerging, gmath.Vec{}, agent2)
		agent2.AssignMode(agentModeMerging, gmath.Vec{}, agent1)
		return true

	case actionSetPatrol:
		numAgents := c.scene.Rand().IntRange(1, 3)
		c.pickCombatUnits(numAgents, func(a *colonyAgentNode) {
			if a.mode == agentModeStandby {
				a.AssignMode(agentModePatrol, gmath.Vec{}, nil)
			}
		})
		return true

	case actionDefenceGarrison:
		attacker := action.Value.(*creepNode)
		numAgents := c.scene.Rand().IntRange(2, 4)
		c.pickCombatUnits(numAgents, func(a *colonyAgentNode) {
			if a.mode == agentModeStandby {
				a.AssignMode(agentModeFollow, gmath.Vec{}, attacker)
			}
		})
		return true

	case actionDefencePatrol:
		attacker := action.Value.(*creepNode)
		numAgents := c.scene.Rand().IntRange(2, 4)
		c.pickCombatUnits(numAgents, func(a *colonyAgentNode) {
			a.AssignMode(agentModeFollow, gmath.Vec{}, attacker)
		})
		return true

	default:
		panic("unexpected action")
	}
}

func (c *colonyCoreNode) pickAgentFaction() factionTag {
	c.factionTagPicker.Reset()
	for _, kv := range c.factionWeights.Elems {
		c.factionTagPicker.AddOption(kv.Key, kv.Weight)
	}
	return c.factionTagPicker.Pick()
}

func (c *colonyCoreNode) pickWorkerUnits(n int, f func(a *colonyAgentNode)) {
	c.pick(n, c.availableAgents, func(a *colonyAgentNode) {
		f(a)
		n--
	})
	c.pick(n, c.availableUniversalAgents, f)
}

func (c *colonyCoreNode) pickCombatUnits(n int, f func(a *colonyAgentNode)) {
	c.pick(n, c.availableCombatAgents, func(a *colonyAgentNode) {
		f(a)
		n--
	})
	c.pick(n, c.availableUniversalAgents, f)
}

func (c *colonyCoreNode) pick(n int, agents []*colonyAgentNode, f func(a *colonyAgentNode)) {
	if len(agents) == 0 || n <= 0 {
		return
	}
	var slider gmath.Slider
	slider.SetBounds(0, len(agents)-1)
	slider.TrySetValue(c.scene.Rand().IntRange(0, len(agents)-1))
	n = gmath.ClampMax(n, len(agents))
	for i := 0; i < n; i++ {
		f(agents[slider.Value()])
		slider.Inc()
	}
}

func (c *colonyCoreNode) WalkAgents(f func(a *colonyAgentNode) bool) {
	var slider gmath.Slider
	if len(c.combatAgents) != 0 {
		slider.SetBounds(0, len(c.combatAgents)-1)
		slider.TrySetValue(c.scene.Rand().IntRange(0, len(c.combatAgents)-1))
		for i := 0; i < len(c.combatAgents); i++ {
			if !f(c.combatAgents[slider.Value()]) {
				return
			}
			slider.Inc()
		}
	}
	if len(c.agents) != 0 {
		slider.SetBounds(0, len(c.agents)-1)
		slider.TrySetValue(c.scene.Rand().IntRange(0, len(c.agents)-1))
		for i := 0; i < len(c.agents); i++ {
			if !f(c.agents[slider.Value()]) {
				return
			}
			slider.Inc()
		}
	}
}

func (c *colonyCoreNode) FindRandomAgent(f func(a *colonyAgentNode) bool) *colonyAgentNode {
	var result *colonyAgentNode
	c.WalkAgents(func(a *colonyAgentNode) bool {
		if f(a) {
			result = a
			return false
		}
		return true
	})
	return result
}

func (c *colonyCoreNode) FindAgent(f func(a *colonyAgentNode) bool) *colonyAgentNode {
	for _, a := range c.agents {
		if f(a) {
			return a
		}
	}
	for _, a := range c.combatAgents {
		if f(a) {
			return a
		}
	}
	return nil
}

func (c *colonyCoreNode) moveTowards(delta, speed float64, pos gmath.Vec) bool {
	travelled := speed * delta
	if c.pos.DistanceTo(pos) <= travelled {
		c.pos = pos
		return true
	}
	c.pos = c.pos.MoveTowards(pos, travelled)
	return false
}
