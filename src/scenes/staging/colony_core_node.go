package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

const (
	coreFlightHeight   float64 = 50
	maxUpkeepValue     int     = 400
	maxEvoPoints       float64 = 20
	maxEvoGain         float64 = 1.0
	blueEvoThreshold   float64 = 15.0
	maxResources       float64 = 500.0
	maxVisualResources float64 = maxResources - 100.0
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

type colonyCoreNode struct {
	sprite              *ge.Sprite
	hatch               *ge.Sprite
	flyingSprite        *ge.Sprite
	shadow              *ge.Sprite
	evoDiode            *ge.Sprite
	resourceRects       []*ge.Sprite
	flyingResourceRects []*ge.Sprite

	scene *ge.Scene

	flashComponent      damageFlashComponent
	hatchFlashComponent damageFlashComponent

	pos       gmath.Vec
	spritePos gmath.Vec
	height    float64
	maxHealth float64
	health    float64

	mode colonyCoreMode

	waypoint        gmath.Vec
	relocationPoint gmath.Vec

	resourceShortage int
	resources        float64
	eliteResources   float64
	evoPoints        float64
	world            *worldState

	agents  *colonyAgentContainer
	turrets []*colonyAgentNode

	planner *colonyActionPlanner

	openHatchTime float64

	realRadius    float64
	realRadiusSqr float64

	upkeepDelay  float64
	cloningDelay float64

	actionDelay float64
	priorities  *weightContainer[colonyPriority]

	failedResource     *essenceSourceNode
	failedResourceTick int

	factionTagPicker *gmath.RandPicker[gamedata.FactionTag]

	factionWeights *weightContainer[gamedata.FactionTag]

	EventDestroyed         gsignal.Event[*colonyCoreNode]
	EventPrioritiesChanged gsignal.Event[*colonyCoreNode]
}

type colonyConfig struct {
	Pos gmath.Vec

	Radius float64

	World *worldState
}

func newColonyCoreNode(config colonyConfig) *colonyCoreNode {
	c := &colonyCoreNode{
		world:      config.World,
		realRadius: config.Radius,
		maxHealth:  100,
	}
	c.realRadiusSqr = c.realRadius * c.realRadius
	c.priorities = newWeightContainer(priorityResources, priorityGrowth, priorityEvolution, prioritySecurity)
	c.factionWeights = newWeightContainer(
		gamedata.NeutralFactionTag,
		gamedata.YellowFactionTag,
		gamedata.RedFactionTag,
		gamedata.GreenFactionTag,
		gamedata.BlueFactionTag)
	c.factionWeights.SetWeight(gamedata.NeutralFactionTag, 1.0)
	c.pos = config.Pos
	return c
}

func (c *colonyCoreNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.agents = newColonyAgentContainer(scene.Rand())

	c.factionTagPicker = gmath.NewRandPicker[gamedata.FactionTag](scene.Rand())

	c.planner = newColonyActionPlanner(c, scene.Rand())

	c.health = c.maxHealth

	c.sprite = scene.NewSprite(assets.ImageColonyCore)
	c.sprite.Pos.Base = &c.spritePos
	if c.world.graphicsSettings.AllShadersEnabled {
		c.sprite.Shader = scene.NewShader(assets.ShaderColonyDamage)
		c.sprite.Shader.SetFloatValue("HP", 1.0)
		c.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageColonyDamageMask)
	}
	c.world.camera.AddSprite(c.sprite)

	c.flyingSprite = scene.NewSprite(assets.ImageColonyCoreFlying)
	c.flyingSprite.Pos.Base = &c.spritePos
	c.flyingSprite.Visible = false
	if c.world.graphicsSettings.AllShadersEnabled {
		c.flyingSprite.Shader = c.sprite.Shader
	}
	c.world.camera.AddSpriteSlightlyAbove(c.flyingSprite)

	c.hatch = scene.NewSprite(assets.ImageColonyCoreHatch)
	c.hatch.Pos.Base = &c.spritePos
	c.hatch.Pos.Offset.Y = -20
	c.world.camera.AddSprite(c.hatch)

	c.flashComponent.sprite = c.sprite
	c.hatchFlashComponent.sprite = c.hatch

	c.evoDiode = scene.NewSprite(assets.ImageColonyCoreDiode)
	c.evoDiode.Pos.Base = &c.spritePos
	c.evoDiode.Pos.Offset = gmath.Vec{X: -16, Y: -29}
	c.world.camera.AddSprite(c.evoDiode)

	if c.world.graphicsSettings.ShadowsEnabled {
		c.shadow = scene.NewSprite(assets.ImageColonyCoreShadow)
		c.shadow.Pos.Base = &c.spritePos
		c.shadow.Visible = false
		c.world.camera.AddSprite(c.shadow)
	}

	c.resourceRects = make([]*ge.Sprite, 3)
	c.flyingResourceRects = make([]*ge.Sprite, 3)
	makeResourceRects := func(rects []*ge.Sprite, above bool) {
		for i := range rects {
			rect := scene.NewSprite(assets.ImageColonyResourceBar1 + resource.ImageID(i))
			rect.Centered = false
			rect.Visible = false
			rect.Pos.Base = &c.spritePos
			rect.Pos.Offset.X -= 3
			rect.Pos.Offset.Y = colonyResourceRectOffsets[i]
			rects[i] = rect
			if above {
				c.world.camera.AddSpriteSlightlyAbove(rect)
			} else {
				c.world.camera.AddSprite(rect)
			}
		}
	}
	makeResourceRects(c.resourceRects, false)
	makeResourceRects(c.flyingResourceRects, true)
}

func (c *colonyCoreNode) IsFlying() bool {
	return c.mode != colonyModeNormal
}

func (c *colonyCoreNode) MaxFlyDistance() float64 {
	return 200 + (float64(c.agents.servoNum) * 30.0)
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

func (c *colonyCoreNode) OnDamage(damage gamedata.DamageValue, source targetable) {
	if damage.Health != 0 {
		c.flashComponent.flash = 0.2
		c.hatchFlashComponent.flash = 0.2
	}

	c.health -= damage.Health
	if c.health < 0 {
		if c.height == 0 {
			createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), true)
		} else {
			shadowImg := assets.ImageNone
			if c.shadow != nil {
				shadowImg = c.shadow.ImageID()
			}

			fall := newDroneFallNode(c.world, nil, c.sprite.ImageID(), shadowImg, c.pos, c.height)
			c.world.nodeRunner.AddObject(fall)
		}
		c.Destroy()
		return
	}

	c.updateHealthShader()
	if c.scene.Rand().Chance(0.7) {
		c.AddPriority(prioritySecurity, 0.02)
	}
}

func (c *colonyCoreNode) Destroy() {
	c.agents.Each(func(a *colonyAgentNode) {
		a.OnDamage(gamedata.DamageValue{Health: 1000}, c)
	})
	for _, turret := range c.turrets {
		turret.OnDamage(gamedata.DamageValue{Health: 1000}, c)
	}
	c.EventDestroyed.Emit(c)
	c.Dispose()
}

func (c *colonyCoreNode) GetEntrancePos() gmath.Vec {
	return c.pos.Add(gmath.Vec{X: -1, Y: -22})
}

func (c *colonyCoreNode) GetStoragePos() gmath.Vec {
	return c.pos.Add(gmath.Vec{X: 1, Y: 0})
}

func (c *colonyCoreNode) AddPriority(kind colonyPriority, delta float64) {
	c.priorities.AddWeight(kind, delta)
	c.EventPrioritiesChanged.Emit(c)
}

func (c *colonyCoreNode) GetResourcePriority() float64 {
	return c.priorities.GetWeight(priorityResources)
}

func (c *colonyCoreNode) GetGrowthPriority() float64 {
	return c.priorities.GetWeight(priorityGrowth)
}

func (c *colonyCoreNode) GetEvolutionPriority() float64 {
	return c.priorities.GetWeight(priorityEvolution)
}

func (c *colonyCoreNode) GetSecurityPriority() float64 {
	return c.priorities.GetWeight(prioritySecurity)
}

func (c *colonyCoreNode) CloneAgentNode(a *colonyAgentNode) *colonyAgentNode {
	pos := a.pos.Add(c.scene.Rand().Offset(-4, 4))
	cloned := a.Clone()
	cloned.pos = pos
	c.AcceptAgent(cloned)
	c.world.result.DronesProduced++
	return cloned
}

func (c *colonyCoreNode) NewColonyAgentNode(stats *gamedata.AgentStats, pos gmath.Vec) *colonyAgentNode {
	if stats.Tier == 3 {
		c.world.result.T3created++
	}
	a := newColonyAgentNode(c, stats, pos)
	c.AcceptAgent(a)
	c.world.result.DronesProduced++
	return a
}

func (c *colonyCoreNode) DetachAgent(a *colonyAgentNode) {
	a.EventDestroyed.Disconnect(c)
	c.agents.Remove(a)
}

func (c *colonyCoreNode) AcceptTurret(turret *colonyAgentNode) {
	turret.EventDestroyed.Connect(c, func(x *colonyAgentNode) {
		c.turrets = xslices.Remove(c.turrets, x)
	})
	c.turrets = append(c.turrets, turret)
	turret.colonyCore = c
}

func (c *colonyCoreNode) AcceptAgent(a *colonyAgentNode) {
	a.EventDestroyed.Connect(c, func(x *colonyAgentNode) {
		c.agents.Remove(x)
	})
	c.agents.Add(a)
	a.colonyCore = c
}

func (c *colonyCoreNode) NumAgents() int { return c.agents.TotalNum() }

func (c *colonyCoreNode) IsDisposed() bool { return c.sprite.IsDisposed() }

func (c *colonyCoreNode) Dispose() {
	c.sprite.Dispose()
	c.hatch.Dispose()
	c.flyingSprite.Dispose()
	if c.shadow != nil {
		c.shadow.Dispose()
	}
	c.evoDiode.Dispose()
	for _, rect := range c.resourceRects {
		rect.Dispose()
	}
	for _, rect := range c.flyingResourceRects {
		rect.Dispose()
	}
}

func (c *colonyCoreNode) updateHealthShader() {
	if c.sprite.Shader.IsNil() {
		return
	}
	percentage := c.health / c.maxHealth
	c.sprite.Shader.SetFloatValue("HP", percentage)
	c.sprite.Shader.Enabled = percentage < 0.95
}

func (c *colonyCoreNode) Update(delta float64) {
	c.flashComponent.Update(delta)
	if c.hatch.Visible {
		c.hatchFlashComponent.Update(delta)
	}

	// FIXME: this should be fixed in the ge package.
	c.spritePos.X = math.Round(c.pos.X)
	c.spritePos.Y = math.Round(c.pos.Y)

	c.updateResourceRects()

	c.cloningDelay = gmath.ClampMin(c.cloningDelay-delta, 0)

	if c.shadow != nil && c.shadow.Visible {
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
		return 12 + float64(c.agents.servoNum)
	case colonyModeRelocating:
		return 18.0 + (float64(c.agents.servoNum) * 3)
	default:
		return 0
	}
}

func (c *colonyCoreNode) updateEvoDiode() {
	offset := 0.0
	if c.evoPoints >= blueEvoThreshold {
		offset = c.evoDiode.FrameWidth * 2
	} else if c.evoPoints >= 1 {
		offset = c.evoDiode.FrameWidth * 1
	}
	c.evoDiode.FrameOffset.X = offset
}

func (c *colonyCoreNode) updateResourceRects() {
	var slice []*ge.Sprite
	if c.IsFlying() {
		slice = c.flyingResourceRects
	} else {
		slice = c.resourceRects
	}

	const resourcesPerBlock float64 = maxVisualResources / 3
	unallocated := c.resources
	for i, rect := range slice {
		var percentage float64
		if unallocated >= resourcesPerBlock {
			percentage = 1.0
		} else if unallocated <= 0 {
			percentage = 0
		} else {
			percentage = unallocated / resourcesPerBlock
		}
		unallocated -= resourcesPerBlock
		pixels := rect.FrameHeight
		height := math.Round(percentage * pixels)
		rect.FrameTrimTop = pixels - height
		rect.Pos.Offset.Y = colonyResourceRectOffsets[i] + (pixels - height)
		rect.Visible = percentage != 0
	}
}

func (c *colonyCoreNode) calcUnitLimit() int {
	calculated := (gmath.ClampMin(c.realRadius-128, 0) * 0.35) + 10
	growth := c.GetGrowthPriority()
	if growth > 0.1 {
		// 50% growth priority gives 24 extra units to the limit.
		// 80% => 42 extra units
		calculated += (growth - 0.1) * 60
	}
	return gmath.Clamp(int(calculated), 10, 160)
}

func (c *colonyCoreNode) calcUpkeed() (float64, int) {
	upkeepTotal := 0
	upkeepDecrease := 0
	c.agents.Each(func(a *colonyAgentNode) {
		switch a.stats.Kind {
		case gamedata.AgentGenerator, gamedata.AgentStormbringer:
			upkeepDecrease++
		}
		upkeepTotal += a.stats.Upkeep
	})
	for _, turret := range c.turrets {
		upkeepTotal += turret.stats.Upkeep
	}
	upkeepDecrease = gmath.ClampMax(upkeepDecrease, 10)
	upkeepTotal = gmath.ClampMin(upkeepTotal-(upkeepDecrease*15), 0)
	if resourcesPriority := c.GetResourcePriority(); resourcesPriority > 0.2 {
		// <=20 -> 0%
		// 40%  -> 20%
		// 80%  -> 60% (max)
		// 100% -> 60% (max)
		maxPercentageDecrease := gmath.ClampMax(resourcesPriority, 0.6)
		upkeepTotal = int(float64(upkeepTotal) * (1.2 - maxPercentageDecrease))
	}
	var resourcePrice float64
	switch {
	case upkeepTotal <= 30:
		// 15 workers or ~7 militia
		resourcePrice = 0
	case upkeepTotal <= 45:
		// ~22 workers or ~11 militia
		resourcePrice = 1
	case upkeepTotal <= 70:
		// 35 workers or ~17 militia
		resourcePrice = 2.0
	case upkeepTotal <= 95:
		// ~47 workers or ~23 militia
		resourcePrice = 3.0
	case upkeepTotal <= 120:
		// ~60 workers or 30 militia
		resourcePrice = 5.0
	case upkeepTotal <= 150:
		// 75 workers or ~37 militia
		resourcePrice = 7
	case upkeepTotal <= 215:
		// ~107 workers or ~53 militia
		resourcePrice = 9
	case upkeepTotal <= 300:
		resourcePrice = 12
	case upkeepTotal < maxUpkeepValue:
		resourcePrice = 15
	default:
		resourcePrice = 20
	}
	return resourcePrice, upkeepTotal
}

func (c *colonyCoreNode) processUpkeep(delta float64) {
	c.upkeepDelay -= delta
	if c.upkeepDelay > 0 {
		return
	}
	c.eliteResources = gmath.ClampMax(c.eliteResources, 10)
	c.upkeepDelay = c.scene.Rand().FloatRange(6.5, 8.5)
	upkeepPrice, _ := c.calcUpkeed()
	if c.resources < upkeepPrice && c.GetResourcePriority() < 0.75 {
		c.AddPriority(priorityResources, 0.04)
		c.resources = 0
	} else {
		c.resources -= upkeepPrice
	}
}

func (c *colonyCoreNode) doRelocation(pos gmath.Vec) {
	c.relocationPoint = pos

	c.agents.Each(func(a *colonyAgentNode) {
		a.clearCargo()
		if a.IsCloaked() {
			a.doUncloak()
		}
		if a.height != agentFlightHeight {
			a.AssignMode(agentModeAlignStandby, gmath.Vec{}, nil)
		} else {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		}
	})

	c.mode = colonyModeTakeoff
	c.openHatchTime = 0

	if c.shadow != nil {
		c.shadow.Visible = true
	}
	c.flyingSprite.Visible = true
	c.flashComponent.sprite = c.flyingSprite
	c.sprite.Visible = false
	c.hatch.Visible = false
	c.evoDiode.Visible = false
	for _, rect := range c.resourceRects {
		rect.Visible = false
	}
	c.updateResourceRects()
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
		c.flashComponent.sprite = c.sprite
		if c.shadow != nil {
			c.shadow.Visible = false
		}
		for _, rect := range c.flyingResourceRects {
			rect.Visible = false
		}
		c.updateResourceRects()
		c.sprite.Visible = true
		c.hatch.Visible = true
		c.evoDiode.Visible = true
		playSound(c.world, assets.AudioColonyLanded, c.pos)
		c.createLandingSmokeEffect()
		c.crushCrawlers()
	}
}

func (c *colonyCoreNode) crushCrawlers() {
	const crushRangeSqr = 24.0 * 24.0
	const explodeRangeSqr = 38.0 * 38.0
	crushPos := c.pos.Add(gmath.Vec{Y: 4})
	for _, creep := range c.world.creeps {
		if creep.stats.kind != creepCrawler {
			continue
		}
		distSqr := creep.pos.DistanceSquaredTo(crushPos)
		if distSqr <= crushRangeSqr {
			creep.Destroy() // Defeat without an explosion
			c.world.result.CreepsStomped++
			continue
		}
		if distSqr <= explodeRangeSqr {
			creep.OnDamage(gamedata.DamageValue{Health: 1000}, c)
			c.world.result.CreepsStomped++
			continue
		}
	}
}

func (c *colonyCoreNode) createLandingSmokeEffect() {
	type effectInfo struct {
		image  resource.ImageID
		offset gmath.Vec
		flip   bool
	}
	effects := [...]effectInfo{
		{image: assets.ImageSmokeDown, offset: gmath.Vec{Y: 36}},
		{image: assets.ImageSmokeSideDown, offset: gmath.Vec{X: 16, Y: 34}},
		{image: assets.ImageSmokeSideDown, offset: gmath.Vec{X: -16, Y: 34}, flip: true},
		{image: assets.ImageSmokeSide, offset: gmath.Vec{X: 30, Y: 28}},
		{image: assets.ImageSmokeSide, offset: gmath.Vec{X: -30, Y: 28}, flip: true},
	}
	for _, info := range effects {
		sprite := c.scene.NewSprite(info.image)
		sprite.FlipHorizontal = info.flip
		sprite.Pos.Offset = c.pos.Add(info.offset)
		e := newEffectNodeFromSprite(c.world.camera, false, sprite)
		e.anim.SetAnimationSpan(0.3)
		c.world.nodeRunner.AddObject(e)
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
	if c.resourceShortage >= 5 && c.GetResourcePriority() < 0.5 {
		c.AddPriority(priorityResources, c.scene.Rand().FloatRange(0.02, 0.05))
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
		c.actionDelay = c.scene.Rand().FloatRange(0.15, 0.25)
	}
}

func (c *colonyCoreNode) tryExecutingAction(action colonyAction) bool {
	switch action.Kind {
	case actionGenerateEvo:
		evoGain := 0.0
		var connectedWorker *colonyAgentNode
		var connectedFighter *colonyAgentNode
		c.agents.Find(searchWorkers|searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
			if evoGain >= maxEvoGain {
				return true
			}
			if a.stats.Tier != 2 {
				return false
			}
			if a.stats.CanPatrol {
				if connectedFighter == nil {
					connectedFighter = a
				}
			} else {
				if connectedWorker == nil {
					connectedWorker = a
				}
			}
			if a.faction == gamedata.BlueFactionTag {
				// 20% more evo points per blue drones.
				evoGain += 0.12
			} else {
				evoGain += 0.1
			}
			return false
		})
		if connectedWorker != nil {
			beam := newBeamNode(c.world, c.evoDiode.Pos, ge.Pos{Base: &connectedWorker.pos}, evoBeamColor)
			beam.width = 2
			c.world.nodeRunner.AddObject(beam)
		}
		if connectedFighter != nil {
			beam := newBeamNode(c.world, c.evoDiode.Pos, ge.Pos{Base: &connectedFighter.pos}, evoBeamColor)
			beam.width = 2
			c.world.nodeRunner.AddObject(beam)
		}
		c.evoPoints = gmath.ClampMax(c.evoPoints+evoGain, maxEvoPoints)
		c.updateEvoDiode()
		return true

	case actionSendCourier:
		courier := action.Value2.(*colonyAgentNode)
		target := action.Value.(*colonyCoreNode)
		if target.resources*1.2 < c.resources && c.resources > 60 {
			courier.payload = courier.maxPayload()
			courier.cargoValue = float64(courier.payload) * 10
			c.resources -= courier.cargoValue
		}
		return courier.AssignMode(agentModeCourierFlight, gmath.Vec{}, action.Value)

	case actionMineEssence:
		if c.agents.NumAvailableWorkers() == 0 {
			return false
		}
		source := action.Value.(*essenceSourceNode)
		maxNumAgents := gmath.Clamp(c.agents.NumAvailableWorkers()/4, 2, 10)
		minNumAgents := gmath.Clamp(c.agents.NumAvailableWorkers()/10, 2, 6)
		toAssign := c.scene.Rand().IntRange(minNumAgents, maxNumAgents)
		extraAssign := gmath.Clamp(int(c.GetResourcePriority()*10)-1, 0, 10)
		toAssign += extraAssign
		toAssign = gmath.ClampMax(toAssign, source.resource)
		numAssigned := 0
		c.agents.Find(searchWorkers|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
			if source.stats == redOilSource && a.stats.Kind != gamedata.AgentRedminer {
				return false
			}
			toAssign--
			if a.AssignMode(agentModeMineEssence, gmath.Vec{}, source) {
				numAssigned++
			}
			return toAssign <= 0
		})
		if numAssigned == 0 && c.failedResource == nil {
			c.failedResource = source
			c.failedResourceTick = 0
		}
		return numAssigned != 0

	case actionRepairTurret:
		repairCost := 4.0
		ok := false
		if c.resources < repairCost {
			return false
		}
		c.pickWorkerUnits(1, func(a *colonyAgentNode) {
			if a.AssignMode(agentModeRepairTurret, gmath.Vec{}, action.Value) {
				c.resources -= repairCost
				ok = true
			}
		})
		return ok

	case actionRepairBase:
		repairCost := 7.0
		ok := false
		if c.resources < repairCost {
			return false
		}
		c.pickWorkerUnits(1, func(a *colonyAgentNode) {
			if a.AssignMode(agentModeRepairBase, gmath.Vec{}, nil) {
				c.resources -= repairCost
				ok = true
			}
		})
		return ok

	case actionBuildBuilding:
		sendCost := 3.0
		maxNumAgents := gmath.Clamp(c.agents.NumAvailableWorkers()/10, 1, 6)
		minNumAgents := gmath.Clamp(c.agents.NumAvailableWorkers()/15, 1, 3)
		toAssign := c.scene.Rand().IntRange(minNumAgents, maxNumAgents)
		// TODO: prefer green workers?
		c.pickWorkerUnits(toAssign, func(a *colonyAgentNode) {
			if c.resources < sendCost {
				return
			}
			if a.AssignMode(agentModeBuildBuilding, gmath.Vec{}, action.Value) {
				c.resources -= sendCost
			}
		})
		return true

	case actionRecycleAgent:
		a := action.Value.(*colonyAgentNode)
		a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
		return true

	case actionProduceAgent:
		a := c.NewColonyAgentNode(action.Value.(*gamedata.AgentStats), c.GetEntrancePos())
		a.faction = c.pickAgentFaction()
		if c.eliteResources >= 1 {
			c.eliteResources--
			a.rank = 1
		}
		a.height = 0
		c.world.nodeRunner.AddObject(a)
		c.resources -= a.stats.Cost
		a.AssignMode(agentModeTakeoff, gmath.Vec{}, nil)
		playSound(c.world, assets.AudioAgentProduced, c.pos)
		c.openHatchTime = 1.5
		return true

	case actionGetReinforcements:
		wantWorkers := c.scene.Rand().IntRange(2, 4)
		wantWarriors := c.scene.Rand().IntRange(1, 2)
		transferUnit := func(dst, src *colonyCoreNode, a *colonyAgentNode) {
			src.DetachAgent(a)
			dst.AcceptAgent(a)
			a.AssignMode(agentModeAlignStandby, gmath.Vec{}, nil)
		}
		srcColony := action.Value.(*colonyCoreNode)
		workersSent := 0
		srcColony.pickWorkerUnits(wantWorkers, func(a *colonyAgentNode) {
			workersSent++
			transferUnit(c, srcColony, a)
		})
		if workersSent == 0 {
			return false
		}
		srcColony.pickCombatUnits(wantWarriors, func(a *colonyAgentNode) {
			transferUnit(c, srcColony, a)
		})
		return true

	case actionCloneAgent:
		c.cloningDelay = 6.5
		cloneTarget := action.Value.(*colonyAgentNode)
		cloner := action.Value2.(*colonyAgentNode)
		c.resources -= agentCloningCost(c, cloner, cloneTarget)
		cloner.AssignMode(agentModeMakeClone, gmath.Vec{}, cloneTarget)
		cloneTarget.AssignMode(agentModeWaitCloning, gmath.Vec{}, cloner)
		return true

	case actionMergeAgents:
		agent1 := action.Value.(*colonyAgentNode)
		agent2 := action.Value2.(*colonyAgentNode)
		agent1.AssignMode(agentModeMerging, gmath.Vec{}, agent2)
		agent2.AssignMode(agentModeMerging, gmath.Vec{}, agent1)
		if action.Value3 != 0 {
			c.evoPoints = gmath.ClampMin(c.evoPoints-action.Value3, 0)
			c.updateEvoDiode()
		}
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
			if a.mode == agentModeStandby && a.CanAttack(attacker.TargetKind()) {
				a.AssignMode(agentModeFollow, gmath.Vec{}, attacker)
			}
		})
		return true

	case actionDefencePatrol:
		attacker := action.Value.(*creepNode)
		numAgents := c.scene.Rand().IntRange(2, 4)
		c.pickCombatUnits(numAgents, func(a *colonyAgentNode) {
			if a.CanAttack(attacker.TargetKind()) {
				a.AssignMode(agentModeFollow, gmath.Vec{}, attacker)
			}
		})
		return true

	default:
		panic("unexpected action")
	}
}

func (c *colonyCoreNode) pickAgentFaction() gamedata.FactionTag {
	c.factionTagPicker.Reset()
	for _, kv := range c.factionWeights.Elems {
		c.factionTagPicker.AddOption(kv.Key, kv.Weight)
	}
	return c.factionTagPicker.Pick()
}

func (c *colonyCoreNode) pickWorkerUnits(n int, f func(a *colonyAgentNode)) {
	c.agents.Find(searchWorkers|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
		f(a)
		n--
		return n <= 0
	})
}

func (c *colonyCoreNode) pickCombatUnits(n int, f func(a *colonyAgentNode)) {
	c.agents.Find(searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
		f(a)
		n--
		return n == 0
	})
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
