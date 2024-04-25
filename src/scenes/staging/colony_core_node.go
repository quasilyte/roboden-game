package staging

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
)

const (
	maxUpkeepValue   int     = 800
	maxEvoPoints     float64 = 20
	maxEvoGain       float64 = 1.0
	blueEvoThreshold float64 = 18.0
)

type colonyCoreMode int

const (
	colonyModeNormal colonyCoreMode = iota
	colonyModeTakeoff
	colonyModeRelocating
	colonyModeLanding
	colonyModeTeleporting
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
	evoDiode            *ge.Sprite
	resourceRects       []*ge.Sprite
	flyingResourceRects []*ge.Sprite
	otherShader         ge.Shader

	stats *gamedata.ColonyCoreStats

	shadowComponent shadowComponent

	player player
	scene  *ge.Scene

	flashComponent      damageFlashComponent
	hatchFlashComponent damageFlashComponent

	pos           gmath.Vec
	drawOrder     float64
	maxHealth     float64
	health        float64
	teleportDelay float64

	maxSpeed    float64
	maxJumpDist float64

	activatedTeleport *teleporterNode

	id     int
	tether int

	heavyDamageWarningCooldown float64

	mode colonyCoreMode

	rallyPoint             gmath.Vec
	waypoint               gmath.Vec
	relocationPoint        gmath.Vec
	plannedRelocationPoint gmath.Vec

	path pathing.GridPath

	resourceShortage int
	resources        float64
	eliteResources   float64
	evoPoints        float64
	world            *worldState

	agents          *colonyAgentContainer
	roombas         []*colonyAgentNode
	turrets         []*colonyAgentNode
	numTurretsBuilt int

	planner *colonyActionPlanner

	acceleration float64

	openHatchTime float64

	realRadius    float64
	realRadiusSqr float64

	stashTransferDelay  float64
	repairSentinelDelay float64
	freeWorkerDelay     float64
	upkeepDelay         float64
	cloningDelay        float64
	resourceDelay       float64
	captureDelay        float64
	artifactDelay       float64

	attackDelay float64

	actionDelay float64
	priorities  *weightContainer[colonyPriority]

	failedResource     *essenceSourceNode
	failedResourceTick int

	factionTagPicker *gmath.RandPicker[gamedata.FactionTag]

	factionWeights *weightContainer[gamedata.FactionTag]

	EventTurretAccepted    gsignal.Event[*colonyAgentNode]
	EventTeleported        gsignal.Event[*colonyCoreNode]
	EventUnderAttack       gsignal.Event[*colonyCoreNode]
	EventOnDamage          gsignal.Event[targetable]
	EventDestroyed         gsignal.Event[*colonyCoreNode]
	EventPrioritiesChanged gsignal.Event[*colonyCoreNode]
	EventFactionsChanged   gsignal.Event[gamedata.FactionTag]
}

type colonyConfig struct {
	Pos gmath.Vec

	Radius float64

	World *worldState

	Player player
}

func newColonyCoreNode(config colonyConfig) *colonyCoreNode {
	stats := config.World.coreDesign
	c := &colonyCoreNode{
		world:       config.World,
		realRadius:  config.Radius,
		maxHealth:   stats.MaxHealth,
		maxSpeed:    stats.Speed * 3.5,
		maxJumpDist: stats.JumpDist + 150,
		player:      config.Player,
		stats:       stats,
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

func (c *colonyCoreNode) spriteWithAlliance(imageID resource.ImageID) *ge.Sprite {
	if len(c.world.players) > 1 && c.world.config.GameMode != gamedata.ModeReverse {
		img := c.scene.LoadImage(imageID)
		paintedImg := ebiten.NewImage(img.Data.Size())
		var drawOptions ebiten.DrawImageOptions
		paintedImg.DrawImage(img.Data, &drawOptions)
		drawOptions.GeoM.Translate(c.stats.AllianceColorOffset.X, c.stats.AllianceColorOffset.Y)
		if c.player.GetState().id != 0 {
			drawOptions.ColorM.RotateHue(float64(gmath.DegToRad(-70)))
		}
		paintedImg.DrawImage(c.scene.LoadImage(c.stats.AllianceColorImageID()).Data, &drawOptions)
		sprite := ge.NewSprite(c.scene.Context())
		sprite.SetImage(resource.Image{Data: paintedImg})
		return sprite
	}
	return c.scene.NewSprite(imageID)
}

func (c *colonyCoreNode) BoundsRect() gmath.Rect {
	if c.stats == gamedata.TankCoreStats {
		return resizedRect(c.sprite.BoundsRect(), -12)
	}
	return resizedRect(c.sprite.BoundsRect(), -10)
}

func (c *colonyCoreNode) GetRallyPoint() gmath.Vec {
	if c.rallyPoint.IsZero() {
		return c.pos
	}
	return c.rallyPoint
}

func (c *colonyCoreNode) addSpriteToStage(s *ge.Sprite) {
	switch c.stats {
	case gamedata.ArkCoreStats:
		c.world.stage.AddSortableGraphicsSlightlyAbove(s, &c.drawOrder)
	case gamedata.DenCoreStats:
		c.world.stage.AddSprite(s)
	case gamedata.TankCoreStats:
		c.world.stage.AddSortableGraphics(s, &c.drawOrder)
	case gamedata.HiveCoreStats:
		c.world.stage.AddSprite(s)
	default:
		panic("unexpected core design")
	}
}

func (c *colonyCoreNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.freeWorkerDelay = 10

	c.agents = newColonyAgentContainer(scene.Rand())

	c.factionTagPicker = gmath.NewRandPicker[gamedata.FactionTag](scene.Rand())

	c.planner = newColonyActionPlanner(c, scene.Rand())

	c.health = c.maxHealth

	if len(c.world.teleporters) != 0 {
		c.otherShader = scene.NewShader(assets.ShaderColonyTeleport)
	}

	if c.world.graphicsSettings.ShadowsEnabled && c.stats.Shadow != assets.ImageNone {
		c.shadowComponent.Init(c.world, c.stats.Shadow)
		c.shadowComponent.offset = c.stats.ShadowOffsetY
	}

	c.sprite = c.spriteWithAlliance(c.stats.Image)
	c.sprite.Pos.Base = &c.pos
	if c.world.graphicsSettings.AllShadersEnabled {
		c.sprite.Shader = scene.NewShader(assets.ShaderColonyDamage)
		c.sprite.Shader.SetFloatValue("HP", 1.0)
		c.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageColonyDamageMask)
	}
	c.addSpriteToStage(c.sprite)

	c.flyingSprite = c.spriteWithAlliance(c.stats.FlyingImageID())
	c.flyingSprite.Pos.Base = &c.pos
	c.flyingSprite.Visible = false
	if c.world.graphicsSettings.AllShadersEnabled {
		c.flyingSprite.Shader = c.sprite.Shader
	}
	switch c.stats {
	case gamedata.ArkCoreStats, gamedata.DenCoreStats:
		c.world.stage.AddSortableGraphicsSlightlyAbove(c.flyingSprite, &c.drawOrder)
	case gamedata.TankCoreStats:
		c.world.stage.AddSortableGraphics(c.flyingSprite, &c.drawOrder)
	}

	c.hatch = scene.NewSprite(assets.ImageColonyCoreHatch)
	c.hatch.Pos.Base = &c.pos
	c.hatch.Pos.Offset.Y = c.stats.HatchOffsetY + 2
	c.addSpriteToStage(c.hatch)

	c.flashComponent.sprite = c.sprite
	c.hatchFlashComponent.sprite = c.hatch

	c.evoDiode = scene.NewSprite(assets.ImageColonyCoreDiode)
	c.evoDiode.Pos.Base = &c.pos
	c.evoDiode.Pos.Offset = c.stats.DiodeOffset
	c.addSpriteToStage(c.evoDiode)

	if c.stats != gamedata.TankCoreStats {
		// Den, Ark, Hive cores.
		c.resourceRects = make([]*ge.Sprite, 3)
		c.flyingResourceRects = make([]*ge.Sprite, 3)
		makeResourceRects := func(rects []*ge.Sprite, above bool) {
			for i := range rects {
				rect := scene.NewSprite(assets.ImageColonyResourceBar1 + resource.ImageID(i))
				rect.Centered = false
				rect.Visible = false
				rect.Pos.Base = &c.pos
				rect.Pos.Offset.X -= 3
				rect.Pos.Offset.Y = colonyResourceRectOffsets[i]
				rects[i] = rect
				if above {
					c.world.stage.AddSortableGraphicsSlightlyAbove(rect, &c.drawOrder)
				} else {
					c.world.stage.AddSprite(rect)
				}
			}
		}
		makeResourceRects(c.resourceRects, false)
		makeResourceRects(c.flyingResourceRects, true)
	} else {
		// Tank core.
		c.resourceRects = make([]*ge.Sprite, 1)
		rect := scene.NewSprite(assets.ImageTankResourceBar)
		rect.Centered = false
		rect.Visible = false
		rect.Pos.Base = &c.pos
		rect.Pos.Offset.X -= 5
		rect.Pos.Offset.Y = 8
		c.resourceRects[0] = rect
		c.world.stage.AddSortableGraphics(rect, &c.drawOrder)
	}

	switch c.stats {
	case gamedata.ArkCoreStats:
		c.shadowComponent.SetVisibility(true)
		c.relocationPoint = c.pos
		c.enterTakeoffMode()
	case gamedata.HiveCoreStats:
		c.rallyPoint = c.pos
	}

	if c.stats != gamedata.ArkCoreStats {
		c.markCells(c.pos)
	}

	c.drawOrder = c.pos.Y
}

func (c *colonyCoreNode) GetTargetInfo() targetInfo {
	return targetInfo{building: true, flying: c.IsFlying()}
}

func (c *colonyCoreNode) IsFlying() bool {
	switch c.stats {
	case gamedata.ArkCoreStats:
		return true
	case gamedata.TankCoreStats, gamedata.HiveCoreStats:
		return false
	default:
		switch c.mode {
		case colonyModeNormal, colonyModeTeleporting:
			return false
		default:
			return true
		}
	}
}

func (c *colonyCoreNode) MaxFlyDistanceSqr() float64 {
	dist := c.MaxFlyDistance()
	return dist * dist
}

func (c *colonyCoreNode) MaxFlyDistance() float64 {
	if c.stats == gamedata.HiveCoreStats {
		return 750.0 + float64(c.agents.servoNum*60.0)
	}
	return gmath.ClampMax(c.stats.JumpDist+float64(c.agents.servoNum*10.0), c.maxJumpDist)
}

func (c *colonyCoreNode) PatrolRadius() float64 {
	return c.realRadius * (1.0 + c.GetSecurityPriority()*0.25)
}

func (c *colonyCoreNode) AttackRadius() float64 {
	return 1.4*c.PatrolRadius() + 320
}

func (c *colonyCoreNode) GetHitboxRadius() float64 {
	return c.stats.HitboxRadius
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
	multiplier := 1.0 - c.damageReduction()
	healthDamage := damage.Health * multiplier
	c.health -= healthDamage
	if c.health < 0 {
		if c.shadowComponent.height == 0 {
			createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), normalEffectLayer)
		} else {
			shadowImg := c.shadowComponent.GetImageID()
			fall := newDroneFallNode(c.world, nil, c.stats.Image, shadowImg, c.pos, c.shadowComponent.height)
			c.world.nodeRunner.AddObject(fall)
			fall.sprite.Shader = c.sprite.Shader
		}
		c.Destroy()
		return
	}

	if damage.Health != 0 {
		if !damage.HasFlag(gamedata.DmgflagNoFlash) {
			c.flashComponent.SetFlash(0.1)
			c.hatchFlashComponent.SetFlash(0.1)
		}
		if c.heavyDamageWarningCooldown == 0 && c.health <= c.maxHealth*0.75 {
			c.heavyDamageWarningCooldown = 45
			c.EventUnderAttack.Emit(c)
		}
		c.EventOnDamage.Emit(source)
	}

	c.updateHealthShader()
	if c.scene.Rand().Chance(0.7) {
		c.AddPriority(prioritySecurity, 0.02)
	}
}

func (c *colonyCoreNode) damageReduction() float64 {
	return c.stats.DamageReduction
}

func (c *colonyCoreNode) Destroy() {
	if c.sprite.IsDisposed() {
		return
	}

	c.agents.Each(func(a *colonyAgentNode) {
		a.OnDamage(gamedata.DamageValue{Health: 1000}, c)
	})
	for _, turret := range c.turrets {
		turret.OnDamage(gamedata.DamageValue{Health: 1000}, c)
	}
	for _, roomba := range c.roombas {
		roomba.OnDamage(gamedata.DamageValue{Health: 1000}, c)
	}
	if !c.IsFlying() {
		c.unmarkCells(c.pos)
	}
	c.EventDestroyed.Emit(c)
	c.Dispose()
}

func (c *colonyCoreNode) GetEntrancePos() gmath.Vec {
	return c.pos.Add(gmath.Vec{X: -1, Y: c.stats.HatchOffsetY})
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
	return cloned
}

func (c *colonyCoreNode) NewColonyAgentNode(stats *gamedata.AgentStats, pos gmath.Vec) *colonyAgentNode {
	if stats.Tier == 3 {
		c.world.result.T3created++
	}
	a := newColonyAgentNode(c, stats, pos)
	c.AcceptAgent(a)
	return a
}

func (c *colonyCoreNode) DetachAgent(a *colonyAgentNode) {
	a.EventDestroyed.Disconnect(c)
	c.agents.Remove(a)
}

func (c *colonyCoreNode) AcceptRoomba(roomba *colonyAgentNode) {
	roomba.EventDestroyed.Connect(c, func(x *colonyAgentNode) {
		c.roombas = xslices.Remove(c.roombas, x)
	})
	c.roombas = append(c.roombas, roomba)
	roomba.colonyCore = c
}

func (c *colonyCoreNode) AddGatheredResources(value float64) {
	c.resources += value
	c.world.result.ResourcesGathered += value
}

func (c *colonyCoreNode) AcceptTurret(turret *colonyAgentNode) {
	turret.EventDestroyed.Connect(c, func(x *colonyAgentNode) {
		if !x.stats.IsNeutral {
			c.numTurretsBuilt--
		}
		if turret.stats != gamedata.HarvesterAgentStats && turret.stats != gamedata.MegaRoombaAgentStats {
			c.world.UnmarkPos(x.pos)
		}
		c.world.turrets = xslices.Remove(c.world.turrets, x)
		c.turrets = xslices.Remove(c.turrets, x)
	})
	if !turret.stats.IsNeutral {
		c.numTurretsBuilt++
	}
	if turret.stats != gamedata.HarvesterAgentStats && turret.stats != gamedata.MegaRoombaAgentStats {
		c.world.MarkPos(turret.pos, ptagBlocked)
	}
	c.world.turrets = append(c.world.turrets, turret)
	c.turrets = append(c.turrets, turret)
	turret.colonyCore = c
	c.EventTurretAccepted.Emit(turret)

	switch turret.stats.Kind {
	case gamedata.AgentHarvester:
		turret.mode = agentModeHarvester
	case gamedata.AgentDroneFactory:
		turret.mode = agentModeRelictDroneFactory
	case gamedata.AgentSiege:
		turret.mode = agentModeSiegeGuard
	case gamedata.AgentMegaRoomba:
		turret.mode = agentModeRelictRoomba
		turret.target = c
	case gamedata.AgentSentinelpoint:
		turret.mode = agentModeSentinelTurret
	default:
		turret.mode = agentModeGuardForever
	}
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
	c.shadowComponent.Dispose()
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
	c.hatchFlashComponent.Update(delta)

	c.updateResourceRects()

	c.stashTransferDelay = gmath.ClampMin(c.stashTransferDelay-delta, 0)
	c.repairSentinelDelay = gmath.ClampMin(c.repairSentinelDelay-delta, 0)
	c.artifactDelay = gmath.ClampMin(c.artifactDelay-delta, 0)
	c.captureDelay = gmath.ClampMin(c.captureDelay-delta, 0)
	c.cloningDelay = gmath.ClampMin(c.cloningDelay-delta, 0)
	c.freeWorkerDelay = gmath.ClampMin(c.freeWorkerDelay-delta, 0)
	c.resourceDelay = gmath.ClampMin(c.resourceDelay-delta, 0)
	c.heavyDamageWarningCooldown = gmath.ClampMin(c.heavyDamageWarningCooldown-delta, 0)

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
	case colonyModeTeleporting:
		c.updateTeleporting(delta)
	}
}

func (c *colonyCoreNode) stopTeleportationEffect() {
	c.otherShader, c.sprite.Shader = c.sprite.Shader, c.otherShader
	c.hatch.Visible = true
	for _, rect := range c.resourceRects {
		rect.Visible = true
	}
}

func (c *colonyCoreNode) updateTeleporting(delta float64) {
	c.teleportDelay -= delta
	c.sprite.Shader.SetFloatValue("Time", 20-(c.teleportDelay*10))

	if c.teleportDelay <= 0 {
		relocationPoint := c.relocationPoint
		c.relocationPoint = gmath.Vec{}

		if !c.activatedTeleport.CanBeUsedBy(c) {
			c.mode = colonyModeNormal
			c.stopTeleportationEffect()
			return
		}

		needToMove := false
		if c.stats == gamedata.TankCoreStats {
			relocationPoint = c.world.pathgrid.AlignPos(c.pos.Sub(c.activatedTeleport.pos).Add(relocationPoint))
			needToMove = !c.world.PosIsFree(relocationPoint, layerLandColony)
		}

		playSound(c.world, assets.AudioTeleportDone, c.pos)
		playSound(c.world, assets.AudioTeleportDone, relocationPoint)

		c.agents.Each(func(a *colonyAgentNode) {
			switch a.mode {
			case agentModeKamikazeAttack, agentModeBomberAttack, agentModeSentinelPatrol:
				return
			}
			// Create effect at the source pos.
			createEffect(c.world, effectConfig{
				Pos:   a.pos,
				Layer: aboveEffectLayer,
				Image: assets.ImageTeleportEffectSmall,
			})
			a.pos = relocationPoint.Add(c.world.rand.Offset(-38, 38))
			a.shadowComponent.UpdatePos(a.pos)
			// Create effect at the destination pos.
			createEffect(c.world, effectConfig{
				Pos:   a.pos,
				Layer: aboveEffectLayer,
				Image: assets.ImageTeleportEffectSmall,
			})
			a.AssignMode(agentModePosing, gmath.Vec{X: c.world.rand.FloatRange(0.5, 2.5)}, nil)
		})

		createEffect(c.world, effectConfig{
			Pos:   c.pos,
			Layer: normalEffectLayer,
			Image: assets.ImageTeleportEffectBig,
		})

		c.unmarkCells(c.pos)
		c.pos = relocationPoint
		c.drawOrder = c.pos.Y
		c.stopTeleportationEffect()

		createEffect(c.world, effectConfig{
			Pos:   c.pos,
			Layer: normalEffectLayer,
			Image: assets.ImageTeleportEffectBig,
		})

		c.EventTeleported.Emit(c)

		if needToMove {
			wp := gmath.RandElem(c.world.rand, tankColonyTeleportOffsets).Add(c.pos)
			c.sendTo(wp)
			c.mode = colonyModeRelocating
		} else {
			c.markCells(c.pos)
			c.mode = colonyModeNormal
		}
	}
}

func (c *colonyCoreNode) movementSpeed() float64 {
	switch c.mode {
	case colonyModeTakeoff, colonyModeLanding:
		speed := 13 + float64(c.agents.servoNum) + float64(c.tether*5)
		return gmath.ClampMax(speed, c.maxSpeed)
	case colonyModeRelocating:
		speed := c.stats.Speed + float64(c.agents.servoNum*3) + float64(c.tether*20)
		return gmath.ClampMax(speed, c.maxSpeed) * c.acceleration
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

func (c *colonyCoreNode) maxVisualResources() float64 {
	return c.stats.ResourcesLimit * 0.75
}

func (c *colonyCoreNode) updateTankResourceRect() {
	maxResources := c.maxVisualResources()
	percentage := 1.0
	if c.resources < maxResources {
		percentage = c.resources / maxResources
	}
	rect := c.resourceRects[0]
	pixels := rect.ImageWidth()
	width := math.Round(percentage * pixels)
	rect.FrameWidth = width
	rect.Visible = percentage != 0
}

func (c *colonyCoreNode) updateResourceRects() {
	if c.mode == colonyModeTeleporting {
		return
	}

	if c.stats == gamedata.TankCoreStats {
		c.updateTankResourceRect()
		return
	}

	var slice []*ge.Sprite
	if c.IsFlying() {
		slice = c.flyingResourceRects
	} else {
		slice = c.resourceRects
	}

	resourcesPerBlock := c.maxVisualResources() / 3
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
	// 128 => 10
	// 256 => 61
	// 400 => 118
	calculated := (gmath.ClampMin(c.realRadius-128, 0) * 0.4) + float64(c.stats.StartingDrones)
	growth := c.GetGrowthPriority()
	if growth > 0.1 {
		// 50% growth priority gives 24 extra units to the limit.
		// 80% => 42 extra units
		calculated += (growth - 0.1) * 60
	}
	calculated *= c.stats.DroneLimitScaling
	return gmath.Clamp(int(calculated), c.stats.StartingDrones, c.stats.DroneLimit)
}

func (c *colonyCoreNode) calcUpkeed() (float64, int) {
	upkeepTotal := 0
	upkeepDecrease := 0
	c.agents.Each(func(a *colonyAgentNode) {
		switch a.stats.Kind {
		case gamedata.AgentGenerator, gamedata.AgentStormbringer:
			upkeepDecrease++
		}
		droneUpkeep := a.stats.Upkeep
		if a.stats.CanGather {
			// Rank 1 (elite) worker costs 2 less resources.
			// Rank 2 (super elite) worker costs 4 less resources.
			droneUpkeep = gmath.ClampMin(droneUpkeep-(2*a.rank), 0)
		}
		upkeepTotal += droneUpkeep
	})
	for _, turret := range c.turrets {
		upkeepTotal += turret.stats.Upkeep
	}
	for _, roomba := range c.roombas {
		upkeepTotal += roomba.stats.Upkeep
	}
	upkeepDecrease = gmath.ClampMax(upkeepDecrease, 10)
	upkeepTotal = gmath.ClampMin(upkeepTotal-(upkeepDecrease*20), 0)
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
		// 15 workers or ~7 scouts
		resourcePrice = 0
	case upkeepTotal <= 45:
		// ~22 workers or ~11 scouts
		resourcePrice = 1
	case upkeepTotal <= 70:
		// 35 workers or ~17 scouts
		resourcePrice = 2.0
	case upkeepTotal <= 95:
		// ~47 workers or ~23 scouts
		resourcePrice = 3.0
	case upkeepTotal <= 120:
		// ~60 workers or 30 scouts
		resourcePrice = 5.0
	case upkeepTotal <= 150:
		// 75 workers or ~37 scouts
		resourcePrice = 7
	case upkeepTotal <= 215:
		// ~107 workers or ~53 scouts
		resourcePrice = 9
	case upkeepTotal <= 300:
		resourcePrice = 12
	case upkeepTotal <= 400:
		resourcePrice = 15
	case upkeepTotal <= 500:
		resourcePrice = 20
	case upkeepTotal <= 600:
		resourcePrice = 25
	case upkeepTotal <= 700:
		resourcePrice = 35
	case upkeepTotal <= maxUpkeepValue:
		resourcePrice = 50
	default:
		resourcePrice = 65
	}
	return resourcePrice, upkeepTotal
}

func (c *colonyCoreNode) processUpkeep(delta float64) {
	if c.resources > c.maxVisualResources() {
		c.resources = gmath.ClampMin(c.resources-delta, 0)
	}
	c.upkeepDelay -= delta
	if c.upkeepDelay > 0 {
		return
	}
	c.eliteResources = gmath.ClampMax(c.eliteResources, 10)
	c.upkeepDelay = c.scene.Rand().FloatRange(7.5, 12.5)
	upkeepPrice, _ := c.calcUpkeed()
	if c.resources < upkeepPrice {
		if c.GetResourcePriority() < 0.5 {
			c.AddPriority(priorityResources, 0.03)
		}
		c.resources = 0
	} else {
		c.resources -= upkeepPrice
	}
}

func (c *colonyCoreNode) switchSprite(moving bool) {
	if c.stats != gamedata.TankCoreStats {
		for _, rect := range c.flyingResourceRects {
			rect.Visible = moving
		}
		for _, rect := range c.resourceRects {
			rect.Visible = !moving
		}
	}

	c.flyingSprite.Visible = moving
	c.sprite.Visible = !moving
	c.hatch.Visible = !moving
	c.evoDiode.Visible = !moving
	c.openHatchTime = 0

	if moving {
		c.flashComponent.ChangeSprite(c.flyingSprite)
	} else {
		c.flashComponent.ChangeSprite(c.sprite)
	}

	c.updateResourceRects()
}

func (c *colonyCoreNode) enterTakeoffMode() {
	c.mode = colonyModeTakeoff
	c.switchSprite(true)
	c.waypoint = c.pos.Sub(gmath.Vec{Y: c.stats.FlightHeight})
}

func (c *colonyCoreNode) doRelocation(pos gmath.Vec) bool {
	c.relocationPoint = pos

	c.agents.Each(func(a *colonyAgentNode) {
		a.clearCargo()
		if a.IsCloaked() {
			a.doUncloak()
		}
		switch a.mode {
		case agentModeKamikazeAttack, agentModeFollowCommander, agentModeSentinelPatrol:
			return
		}
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	})

	switch c.stats {
	case gamedata.DenCoreStats:
		c.acceleration = 0.5
		c.unmarkCells(c.pos)
		c.shadowComponent.SetVisibility(true)
		c.enterTakeoffMode()
		return true

	case gamedata.ArkCoreStats:
		c.acceleration = 0.35
		c.mode = colonyModeRelocating
		c.waypoint = c.relocationPoint
		c.switchSprite(true)
		return true

	case gamedata.TankCoreStats:
		c.acceleration = 0.05
		c.sendTo(pos)
		c.unmarkCells(c.pos)
		c.mode = colonyModeRelocating
		createEffect(c.world, effectConfig{Pos: c.pos.Add(gmath.Vec{X: 12, Y: -18}), Image: assets.ImageRoombaSmoke})
		createEffect(c.world, effectConfig{Pos: c.pos.Add(gmath.Vec{X: -12, Y: -18}), Image: assets.ImageRoombaSmoke})
		c.switchSprite(true)
		return true

	case gamedata.HiveCoreStats:
		c.mode = colonyModeRelocating
		return true
	}

	return false
}

func (c *colonyCoreNode) updateTakeoff(delta float64) {
	c.drawOrder = c.pos.Y - 64
	speed := c.movementSpeed()
	height := c.shadowComponent.height + delta*speed
	if c.moveTowards(delta, speed, c.waypoint) {
		height = c.stats.FlightHeight
		switch c.stats {
		case gamedata.DenCoreStats:
			c.waypoint = c.relocationPoint.Sub(gmath.Vec{Y: c.stats.FlightHeight})
			c.mode = colonyModeRelocating
		case gamedata.ArkCoreStats:
			c.enterNormalMode()
		}
	}
	c.shadowComponent.UpdatePos(c.pos)
	c.shadowComponent.UpdateHeight(c.pos, height, c.stats.FlightHeight)
}

func (c *colonyCoreNode) startLanding() {
	c.waypoint = c.relocationPoint
	c.mode = colonyModeLanding
	c.markCells(c.relocationPoint)
}

func (c *colonyCoreNode) canLandAt(coord pathing.GridCoord) bool {
	return c.world.CellIsFree2x2(coord, layerLandColony)
}

func (c *colonyCoreNode) sendTo(pos gmath.Vec) {
	switch c.stats {
	case gamedata.DenCoreStats:
		c.relocationPoint = pos
		c.waypoint = pos.Sub(gmath.Vec{Y: c.stats.FlightHeight})

	case gamedata.ArkCoreStats:
		c.relocationPoint = pos
		c.waypoint = pos

	case gamedata.TankCoreStats:
		p := c.world.BuildPath(c.pos, pos, layerLandColony)
		c.relocationPoint = c.world.pathgrid.CoordToPos(p.Finish)
		c.path = p.Steps
		c.waypoint = c.world.pathgrid.AlignPos(c.pos)
	}
}

func (c *colonyCoreNode) updateRelocating(delta float64) {
	c.processAttack(delta * 0.25)

	if c.stats == gamedata.HiveCoreStats {
		travelled := 50 * delta
		if c.rallyPoint.DistanceTo(c.relocationPoint) <= travelled {
			c.rallyPoint = c.relocationPoint
			c.relocationPoint = gmath.Vec{}
			c.mode = colonyModeNormal
		} else {
			c.rallyPoint = c.rallyPoint.MoveTowards(c.relocationPoint, travelled)
		}
		return
	}

	c.acceleration = gmath.ClampMax(c.acceleration+(delta*0.3), 1)
	if c.moveTowards(delta, c.movementSpeed(), c.waypoint) {
		switch c.stats {
		case gamedata.DenCoreStats:
			// The landing spot could be unavailable by the moment we reach it.
			coord := c.world.pathgrid.PosToCoord(c.relocationPoint)
			if c.canLandAt(coord) {
				c.startLanding()
				return
			}
			newSpot := c.findLandingSpot(coord, 3)
			if !newSpot.IsZero() {
				c.sendTo(newSpot)
			} else {
				c.startLanding()
			}

		case gamedata.ArkCoreStats:
			newSpot := c.findArkHoverSpot()
			if c.pos == newSpot {
				c.enterNormalMode()
			} else {
				c.sendTo(newSpot)
			}

		case gamedata.TankCoreStats:
			if c.path.HasNext() {
				nextPos := nextPathWaypoint(c.world, c.pos, &c.path, layerLandColony)
				c.waypoint = nextPos.Add(c.world.rand.Offset(-3, 3))
				break
			}
			if !c.world.PosIsFree(c.pos, layerLandColony) {
				coord := c.world.pathgrid.PosToCoord(c.pos)
				freeOffset := randIterate(c.world.rand, resourceNearOffsets, func(offset pathing.GridCoord) bool {
					probe := coord.Add(offset)
					return c.world.CellIsFree(probe, layerLandColony)
				})
				if !freeOffset.IsZero() {
					c.sendTo(c.world.pathgrid.CoordToPos(coord.Add(freeOffset)))
					break
				}
			}
			c.enterNormalMode()
			c.markCells(c.pos)
			c.maybeTeleport()
		}
	}

	c.shadowComponent.UpdatePos(c.pos)
	c.drawOrder = c.pos.Y
}

func (c *colonyCoreNode) findArkHoverSpot() gmath.Vec {
	for _, other := range c.world.allColonies {
		if c == other {
			continue
		}
		if other.pos.DistanceSquaredTo(c.pos) < (pathing.CellSize*pathing.CellSize)+5 {
			offset := gmath.RandElem(c.world.rand, colonyNear2x2CellOffsets)
			return c.pos.Add(gmath.Vec{
				X: float64(offset.X) * pathing.CellSize,
				Y: float64(offset.Y) * pathing.CellSize,
			})
		}
	}
	for _, construction := range c.world.constructions {
		if construction.stats != colonyCoreConstructionStats {
			continue
		}
		if construction.pos.DistanceSquaredTo(c.pos) < (8 * 8) {
			offset := gmath.RandElem(c.world.rand, colonyNear2x2CellOffsets)
			return c.pos.Add(gmath.Vec{
				X: float64(offset.X) * pathing.CellSize,
				Y: float64(offset.Y) * pathing.CellSize,
			})
		}
	}
	return c.pos
}

func (c *colonyCoreNode) findLandingSpot(coord pathing.GridCoord, numTries int) gmath.Vec {
	freeCoord := randIterate(c.world.rand, colonyNear2x2CellOffsets, func(offset pathing.GridCoord) bool {
		probe := coord.Add(offset)
		return c.canLandAt(probe)
	})
	if !freeCoord.IsZero() {
		pos := c.world.pathgrid.CoordToPos(coord.Add(freeCoord)).Sub(gmath.Vec{X: 16, Y: 16})
		return pos
	}

	var freePos gmath.Vec
	if numTries > 0 {
		randIterate(c.world.rand, colonyNear2x2CellOffsets, func(offset pathing.GridCoord) bool {
			probe := coord.Add(offset)
			freePos = c.findLandingSpot(probe, numTries-1)
			return !freePos.IsZero()
		})
	}
	return freePos
}

func (c *colonyCoreNode) enterNormalMode() {
	c.waypoint = gmath.Vec{}
	c.relocationPoint = gmath.Vec{}
	c.mode = colonyModeNormal
	c.switchSprite(false)
	c.drawOrder = c.pos.Y
}

func (c *colonyCoreNode) updateLanding(delta float64) {
	c.drawOrder = c.pos.Y - 64
	speed := c.movementSpeed()
	height := c.shadowComponent.height - delta*speed
	if c.moveTowards(delta, speed, c.waypoint) {
		height = 0
		c.enterNormalMode()
		c.shadowComponent.SetVisibility(false)
		playSound(c.world, assets.AudioColonyLanded, c.pos)
		c.createLandingSmokeEffect()
		c.crushCrawlers()
		c.maybeTeleport()
	}
	c.shadowComponent.UpdatePos(c.pos)
	c.shadowComponent.UpdateHeight(c.pos, height, c.stats.FlightHeight)
}

func (c *colonyCoreNode) maybeTeleport() {
	var teleporter *teleporterNode
	for _, tp := range c.world.teleporters {
		if tp.pos.DistanceTo(c.pos) < 34 {
			teleporter = tp
			break
		}
	}
	if teleporter == nil {
		return
	}
	// Is that teleporter already occupied?
	if !teleporter.CanBeUsedBy(c) {
		return
	}

	c.mode = colonyModeTeleporting
	c.teleportDelay = 2
	c.relocationPoint = teleporter.other.pos.Add(teleportOffset)
	c.activatedTeleport = teleporter
	c.otherShader, c.sprite.Shader = c.sprite.Shader, c.otherShader
	c.sprite.Shader.SetFloatValue("Time", c.teleportDelay)
	playSound(c.world, assets.AudioTeleportCharge, c.pos)
	c.hatch.Visible = false
	for _, rect := range c.resourceRects {
		rect.Visible = false
	}
}

func (c *colonyCoreNode) crushCrawlers() {
	const crushRangeSqr = 24.0 * 24.0
	const explodeRangeSqr = 42.0 * 42.0
	crushPos := c.pos.Add(gmath.Vec{Y: 4})

	c.world.WalkCreepsWithRand(nil, crushPos, 48, func(creep *creepNode) bool {
		switch creep.stats.Kind {
		case gamedata.CreepCrawler, gamedata.CreepHowitzer:
			// OK
		default:
			return false
		}

		distSqr := creep.pos.DistanceSquaredTo(crushPos)
		if distSqr > explodeRangeSqr {
			return false // Too far
		}

		if distSqr > crushRangeSqr || creep.stats.Kind == gamedata.CreepHowitzer {
			creep.OnDamage(gamedata.DamageValue{Health: 5000}, c)
			c.world.result.CreepsStomped++
			return false
		}

		// Defeat without an explosion
		creep.Destroy()
		c.world.result.CreepsStomped++
		return false
	})
}

func (c *colonyCoreNode) createLandingSmokeEffect() {
	if c.world.simulation {
		return
	}

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
		e := newEffectNodeFromSprite(c.world, normalEffectLayer, sprite)
		e.flip = effectFlipDisabled
		e.anim.SetAnimationSpan(0.3)
		c.world.nodeRunner.AddObject(e)
	}
}

func (c *colonyCoreNode) processAttack(delta float64) {
	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	if c.attackDelay != 0 {
		return
	}

	switch c.stats {
	case gamedata.TankCoreStats:
		c.attackWithWeapon(gamedata.TankCoreWeapon1, false)
	case gamedata.HiveCoreStats:
		c.attackWithWeapon(gamedata.HiveMortarWeapon, c.world.rand.Chance(0.4))
	}
}

func (c *colonyCoreNode) attackWithWeapon(weapon *gamedata.WeaponStats, guided bool) {
	targets := findAttackTargets(c.world, c.pos, weapon)
	for _, target := range targets {
		attackWithProjectile(c.world, weapon, c, target, weapon.BurstSize, guided)
	}
	if len(targets) == 0 {
		c.attackDelay = c.world.rand.FloatRange(0.4, 0.9)
	} else {
		c.attackDelay = weapon.Reload * c.world.rand.FloatRange(0.9, 1.1)
	}
}

func (c *colonyCoreNode) updateNormal(delta float64) {
	c.actionDelay = gmath.ClampMin(c.actionDelay-delta, 0)
	if c.actionDelay == 0 {
		c.doAction()
	}
	c.processAttack(delta)
	c.openHatchTime = gmath.ClampMin(c.openHatchTime-delta, 0)
	c.hatch.Visible = c.openHatchTime == 0
}

func (c *colonyCoreNode) doAction() {
	if c.resourceShortage >= 5 && c.GetResourcePriority() < 0.4 {
		c.AddPriority(priorityResources, c.scene.Rand().FloatRange(0.01, 0.03))
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

func (c *colonyCoreNode) unmarkCells(pos gmath.Vec) {
	if c.stats == gamedata.TankCoreStats {
		c.world.UnmarkPos(pos)
	} else {
		c.world.UnmarkPos2x2(pos)
	}
}

func (c *colonyCoreNode) markCells(pos gmath.Vec) {
	if c.stats == gamedata.TankCoreStats {
		c.world.MarkPos(pos, ptagBlocked)
	} else {
		c.world.MarkPos2x2(pos, ptagBlocked)
	}
}

func (c *colonyCoreNode) createEvoBeam(to ge.Pos) {
	if c.world.simulation {
		return
	}
	from := c.evoDiode.Pos
	beam := newBeamNode(c.world, from, to, evoBeamColor)
	beam.width = 2
	c.world.nodeRunner.AddObject(beam)
}

func (c *colonyCoreNode) calcResourceDelay() float64 {
	// 0.1 resource priority: 3.6 delay
	// 0.2 resource priority: 3.2 delay
	// 0.3 resource priority: 2.8 delay
	// 0.5 resource priority: 2.0 delay
	// 0.7 resource priority: 1.2 delay
	return 4.0 - (c.GetResourcePriority() * 4)
}

func (c *colonyCoreNode) addEvoPoints(delta float64) {
	c.evoPoints = gmath.ClampMax(c.evoPoints+delta, maxEvoPoints)
	c.updateEvoDiode()
}

func (c *colonyCoreNode) tryExecutingAction(action colonyAction) bool {
	switch action.Kind {
	case actionAccessResourceStash:
		c.stashTransferDelay += 3
		c.player.GetState().TransferResources(c, action.Value3)
		return true

	case actionConvertEvo:
		target := action.Value.(*neutralBuildingNode)
		c.createEvoBeam(ge.Pos{Base: &target.pos, Offset: gmath.Vec{Y: 13}})
		c.resources += 25
		target.agent.specialDelay = 1.5
		return true

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
				// ~25% more evo points per blue drones.
				evoGain += 0.05
			} else {
				evoGain += 0.04
			}
			return false
		})
		if connectedWorker != nil {
			c.createEvoBeam(ge.Pos{Base: &connectedWorker.pos})
		}
		if connectedFighter != nil {
			c.createEvoBeam(ge.Pos{Base: &connectedFighter.pos})
		}
		// Initial colony radius is 128, minimal radius is 96.
		// Every increase radius action adds ~30 to the radius.
		// * 100 radius => 1.5 (max)
		// * 200 radius => 1.0
		// * 300 radius => 0.5
		// * 400 radius => 0.1 (min)
		evoGainMultiplier := gmath.Clamp(2.0-(c.realRadius/200), 0.1, 1.5)
		c.addEvoPoints(evoGain * evoGainMultiplier)
		return true

	case actionSendCourier:
		courier := action.Value2.(*colonyAgentNode)
		target := action.Value.(*colonyCoreNode)
		if target.resources*1.3 < c.resources && c.resources > 60 {
			const resPerUnit float64 = 12
			courier.payload = gmath.ClampMax(courier.maxPayload(), int(c.resources/resPerUnit))
			cargoValue := float64(courier.payload) * resPerUnit
			courier.cargoValue = cargoValue + 0.5
			c.resources -= cargoValue
		}
		return courier.AssignMode(agentModeCourierFlight, gmath.Vec{}, action.Value)

	case actionMineSulfurEssence:
		if c.agents.NumAvailableWorkers() == 0 {
			return false
		}
		source := action.Value.(*essenceSourceNode)
		toAssign := gmath.ClampMax(3, c.agents.NumAvailableWorkers())
		numAssigned := 0
		var miningNode *sulfurMiningNode
		c.agents.Find(searchWorkers|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
			if a.AssignMode(agentModeMineSulfurEssence, gmath.Vec{}, source) {
				if miningNode == nil {
					miningNode = newSulfurMiningNode(source)
				}
				toAssign--
				numAssigned++
				miningNode.miners = append(miningNode.miners, a)
			}
			return toAssign <= 0
		})
		if numAssigned == 0 && c.failedResource == nil {
			c.failedResource = source
			c.failedResourceTick = 0
		}
		if numAssigned != 0 {
			c.resourceDelay = c.calcResourceDelay() * 0.75
			c.world.nodeRunner.AddObject(miningNode)
		}
		return numAssigned != 0

	case actionGrabArtifact:
		artifact := action.Value.(*essenceSourceNode)
		d := action.Value2.(*colonyAgentNode)
		d.AssignMode(agentModeGrabArtifact, gmath.Vec{}, artifact)
		return true

	case actionMineEssence:
		if c.agents.NumAvailableWorkers() == 0 {
			return false
		}
		source := action.Value.(*essenceSourceNode)
		// 0.1 resource priority: 1
		// 0.2 resource priority: 3
		// 0.3 resource priority: 5
		// 0.5 resource priority: 9
		// 0.7 resource priority: 13
		// 0.8 resource priority: 15 (cap)
		resourcesPriority := c.GetResourcePriority()
		priorityCapacity := gmath.Clamp((resourcesPriority*20)-1, 0, 15)
		// 15 drones => +0
		// 25 drones => +1
		// 35 drones => +2
		// 45 drones => +3
		// 100 drones => +8
		colonySizeBonus := gmath.Clamp((len(c.agents.workers)-15)/10, 0, 10)
		toAssign := int(math.Floor(priorityCapacity)*c.scene.Rand().FloatRange(0.8, 1.3)) + colonySizeBonus
		if toAssign == 0 {
			return false
		}
		toAssign = gmath.ClampMax(toAssign, source.resource)
		numAssigned := 0
		c.agents.Find(searchWorkers|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
			if source.stats == redOilSource && a.stats.Kind != gamedata.AgentRedminer {
				return false
			}
			if source.stats == mineralSource && a.stats.Tier < 2 {
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
		if numAssigned != 0 {
			c.resourceDelay = c.calcResourceDelay()
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
		target := action.Value.(*constructionNode)
		sendCost := target.ConstructionCost()
		if c.stats == gamedata.DenCoreStats {
			sendCost *= 0.5
		}
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

	case actionCaptureBuilding:
		captureCost := 20.0
		ok := false
		if c.resources < captureCost {
			return false
		}
		c.pickWorkerUnits(1, func(a *colonyAgentNode) {
			if a.AssignMode(agentModeCaptureBuilding, gmath.Vec{}, action.Value) {
				c.resources -= captureCost
				ok = true
			}
		})
		return ok

	case actionRecycleAgent:
		a := action.Value.(*colonyAgentNode)
		a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
		return true

	case actionProduceAgent:
		pos := c.GetEntrancePos()
		if c.world.rand.Bool() {
			pos.X++
		}
		a := c.NewColonyAgentNode(action.Value.(*gamedata.AgentStats), pos)
		a.faction = c.pickAgentFaction()
		if c.eliteResources >= 1 {
			c.eliteResources--
			a.rank = 1
		}
		c.world.nodeRunner.AddObject(a)
		a.SetHeight(c.shadowComponent.height)
		c.world.result.DronesProduced++
		c.resources = gmath.ClampMin(c.resources-(a.stats.Cost*c.stats.DroneProductionCost), 0)
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
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
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
			if a.stats == gamedata.CommanderAgentStats {
				return
			}
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
		mode := agentModeMerging
		mergePoint := midpoint(agent1.pos, agent2.pos)
		mergePos1 := mergePoint.Add(c.scene.Rand().Offset(-14, 14))
		mergePos2 := mergePoint.Add(c.scene.Rand().Offset(-14, 14))
		if gamedata.ColonyAgentKind(action.Value4) == gamedata.AgentRoomba {
			mode = agentModeMergingRoomba
			if !posIsFreeWithFlags(c.world, nil, mergePoint.Add(gmath.Vec{Y: agentFlightHeight}), 10, collisionSkipSmallCrawlers|collisionSkipTeleporters) {
				return false
			}
		}
		agent1.AssignMode(mode, mergePos1, agent2)
		agent2.AssignMode(mode, mergePos2, agent1)
		if action.Value3 != 0 {
			c.evoPoints = gmath.ClampMin(c.evoPoints-action.Value3, 0)
			c.updateEvoDiode()
		}
		return true

	case actionAttachToCommander:
		follower := action.Value2.(*colonyAgentNode)
		follower.AssignMode(agentModeFollowCommander, gmath.Vec{}, action.Value)
		return true

	case actionSetPatrol:
		numAgents := c.scene.Rand().IntRange(1, 3)
		c.pickCombatUnits(numAgents, func(a *colonyAgentNode) {
			if a.mode == agentModeStandby {
				a.AssignMode(agentModePatrol, gmath.Vec{}, nil)
			}
		})
		return true

	case actionAssignSentinels:
		c.world.tmpAgentSlice = c.world.tmpAgentSlice[:0]
		c.agents.Find(searchOnlyAvailable|searchRandomized|searchWorkers, func(a *colonyAgentNode) bool {
			ok := a.health >= a.maxHealth*0.2 && a.stats.CanBeSentinel
			if !ok {
				return false
			}
			c.world.tmpAgentSlice = append(c.world.tmpAgentSlice, a)
			return len(c.world.tmpAgentSlice) >= 3
		})
		if len(c.world.tmpAgentSlice) < 3 {
			return false
		}
		turret := action.Value.(*colonyAgentNode)
		turret.SetSentinelWorkers(c.world.tmpAgentSlice[:3])
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
