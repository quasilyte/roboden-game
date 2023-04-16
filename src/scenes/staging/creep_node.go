package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
)

type creepKind int

const (
	creepPrimitiveWanderer creepKind = iota
	creepStunner
	creepAssault
	creepDominator
	creepBuilder
	creepTurret
	creepTurretConstruction
	creepCrawlerBaseConstruction
	creepBase
	creepCrawlerBase
	creepCrawler
	creepHowitzer
	creepServant
	creepUberBoss
)

const (
	howitzerIdle = iota
	howitzerMove
	howitzerPreparing
	howitzerReady
	howitzerFoldTurret
)

type creepNode struct {
	anim      *ge.Animation
	sprite    *ge.Sprite
	altSprite *ge.Sprite
	shadow    *ge.Sprite

	scene *ge.Scene

	flashComponent damageFlashComponent

	world *worldState
	stats *creepStats

	aggroTarget targetable

	spawnPos        gmath.Vec
	pos             gmath.Vec
	spritePos       gmath.Vec
	waypoint        gmath.Vec
	wasAttacking    bool
	wasRetreating   bool
	spawnedFromBase bool
	cloaking        bool

	path            pathing.GridPath
	specialTarget   any
	specialDelay    float64
	specialModifier float64

	aggro     float64
	disarm    float64
	slow      float64
	health    float64
	maxHealth float64
	height    float64

	attackDelay float64

	bossStage int
	fragScore int

	EventDestroyed    gsignal.Event[*creepNode]
	EventBuildingStop gsignal.Event[gsignal.Void]
}

var crawlerSpawnPositions = []pathing.GridPath{
	pathing.MakeGridPath(pathing.DirDown),
	pathing.MakeGridPath(pathing.DirDown, pathing.DirLeft),
	pathing.MakeGridPath(pathing.DirDown, pathing.DirRight),
	pathing.MakeGridPath(pathing.DirDown, pathing.DirLeft, pathing.DirUp),
	pathing.MakeGridPath(pathing.DirDown, pathing.DirRight, pathing.DirUp),
}

func newCreepNode(world *worldState, stats *creepStats, pos gmath.Vec) *creepNode {
	return &creepNode{
		world:    world,
		stats:    stats,
		pos:      pos,
		spawnPos: pos,
	}
}

func (c *creepNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.maxHealth = c.stats.maxHealth

	c.sprite = scene.NewSprite(c.stats.image)
	c.sprite.Pos.Base = &c.spritePos
	if c.stats.shadowImage != assets.ImageNone {
		c.world.camera.AddSpriteAbove(c.sprite)
	} else {
		c.world.camera.AddSprite(c.sprite)
	}
	if c.stats.animSpeed != 0 {
		if c.stats.kind == creepHowitzer {
			// 4 frames for the walk, 1 frame is for the "ready" state.
			c.anim = ge.NewRepeatedAnimation(c.sprite, 4)
		} else {
			c.anim = ge.NewRepeatedAnimation(c.sprite, -1)
		}
		c.anim.Tick(scene.Rand().FloatRange(0, 0.7))
		c.anim.SetSecondsPerFrame(c.stats.animSpeed)
	}
	c.flashComponent.sprite = c.sprite

	c.height = agentFlightHeight

	if c.stats.shadowImage != assets.ImageNone && c.world.graphicsSettings.ShadowsEnabled {
		c.shadow = scene.NewSprite(c.stats.shadowImage)
		c.shadow.Pos.Base = &c.spritePos
		c.world.camera.AddSprite(c.shadow)
		c.shadow.Pos.Offset.Y = c.height
		c.shadow.SetAlpha(0.5)
		if c.spawnedFromBase {
			c.shadow.Visible = false
		}
	}

	if c.stats.kind == creepUberBoss {
		c.altSprite = scene.NewSprite(assets.ImageUberBossOpen)
		c.altSprite.Visible = false
		c.altSprite.Pos.Base = &c.spritePos
		c.world.camera.AddSprite(c.altSprite)
		c.maxHealth *= c.world.bossHealthMultiplier
	} else {
		c.maxHealth *= c.world.creepHealthMultiplier
		if c.stats.kind == creepHowitzer {
			c.altSprite = scene.NewSprite(assets.ImageHowitzerPreparing)
			c.altSprite.Visible = false
			c.altSprite.Pos.Base = &c.spritePos
			c.altSprite.Pos.Offset.Y = -3
			c.world.camera.AddSprite(c.altSprite)
		}
	}
	switch c.stats.kind {
	case creepServant:
		c.specialDelay = c.scene.Rand().FloatRange(0.5, 3)
	case creepBuilder:
		// c.specialDelay = c.scene.Rand().FloatRange(15, 30)
	case creepCrawlerBase:
		c.attackDelay = c.scene.Rand().FloatRange(5, 10)
	case creepTurretConstruction, creepCrawlerBaseConstruction:
		c.sprite.Shader = scene.NewShader(assets.ShaderCreepTurretBuild)
	case creepHowitzer:
		pos := ge.Pos{Base: &c.spritePos, Offset: gmath.Vec{Y: -10}}
		trunk := newHowitzerTrunkNode(c.world.camera, pos)
		c.specialTarget = trunk
		c.world.nodeRunner.AddObject(trunk)
		trunk.SetVisibility(false)
	}

	c.health = c.maxHealth
}

func (c *creepNode) Dispose() {
	c.sprite.Dispose()
	if c.shadow != nil {
		c.shadow.Dispose()
	}
	if c.altSprite != nil {
		c.altSprite.Dispose()
	}

	if c.stats.kind == creepHowitzer {
		trunk := c.specialTarget.(*howitzerTrunkNode)
		trunk.Dispose()
	}
}

func (c *creepNode) Destroy() {
	c.EventDestroyed.Emit(c)
	if c.stats.kind == creepBuilder && c.specialTarget != nil {
		c.EventBuildingStop.Emit(gsignal.Void{})
	}
	c.Dispose()
}

func (c *creepNode) IsDisposed() bool { return c.sprite.IsDisposed() }

func (c *creepNode) Update(delta float64) {
	c.flashComponent.Update(delta)

	// FIXME: this should be fixed in the ge package.
	c.spritePos.X = math.Round(c.pos.X)
	c.spritePos.Y = math.Round(c.pos.Y)

	c.slow = gmath.ClampMin(c.slow-delta, 0)
	c.aggro = gmath.ClampMin(c.aggro-delta, 0)
	c.disarm = gmath.ClampMin(c.disarm-delta, 0)
	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	if c.stats.weapon != nil && c.attackDelay == 0 && c.disarm == 0 && !c.cloaking {
		c.attackDelay = c.stats.weapon.Reload * c.scene.Rand().FloatRange(0.8, 1.2)
		targets := c.findTargets()
		if len(targets) != 0 {
			for _, target := range targets {
				c.doAttack(target)
			}
			if !c.stats.weapon.ProjectileFireSound {
				playSound(c.world, c.stats.weapon.AttackSound, c.pos)
			}
		}
	}

	switch c.stats.kind {
	case creepPrimitiveWanderer, creepStunner, creepAssault, creepDominator:
		c.updatePrimitiveWanderer(delta)
	case creepBuilder:
		c.updateBuilder(delta)
	case creepUberBoss:
		c.updateUberBoss(delta)
	case creepServant:
		c.updateServant(delta)
	case creepBase:
		c.updateCreepBase(delta)
	case creepCrawlerBase:
		c.updateCreepCrawlerBase(delta)
	case creepCrawler:
		c.updateCrawler(delta)
	case creepHowitzer:
		c.updateHowitzer(delta)
	case creepTurret:
		// Do nothing.
	case creepTurretConstruction, creepCrawlerBaseConstruction:
		c.updateTurretConstruction(delta)
	default:
		panic("unexpected creep kind in update()")
	}
}

func (c *creepNode) GetPos() *gmath.Vec { return &c.spritePos }

func (c *creepNode) GetVelocity() gmath.Vec {
	if c.waypoint.IsZero() {
		return gmath.Vec{}
	}
	return c.pos.VecTowards(c.waypoint, c.movementSpeed())
}

func (c *creepNode) IsFlying() bool {
	switch c.stats.kind {
	case creepCrawler, creepHowitzer:
		return false
	case creepBase, creepTurret, creepCrawlerBase, creepTurretConstruction, creepCrawlerBaseConstruction:
		return false
	case creepUberBoss:
		return !c.altSprite.Visible
	default:
		return true
	}
}

func (c *creepNode) TargetKind() gamedata.TargetKind {
	if c.IsFlying() {
		return gamedata.TargetFlying
	}
	return gamedata.TargetGround
}

func (c *creepNode) explode() {
	switch c.stats.kind {
	case creepUberBoss:
		if c.IsFlying() {
			shadowImg := assets.ImageNone
			if c.shadow != nil {
				shadowImg = c.shadow.ImageID()
			}

			fall := newDroneFallNode(c.world, nil, c.stats.image, shadowImg, c.pos, c.height)
			c.world.nodeRunner.AddObject(fall)
		} else {
			createAreaExplosion(c.world, spriteRect(c.pos, c.altSprite), true)
		}

	case creepTurret, creepBase, creepCrawlerBase, creepHowitzer:
		createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), true)
		scraps := c.world.NewEssenceSourceNode(bigScrapCreepSource, c.pos.Add(gmath.Vec{Y: 7}))
		c.world.nodeRunner.AddObject(scraps)
	case creepTurretConstruction, creepCrawlerBaseConstruction:
		createExplosion(c.world, false, c.pos)
		scraps := c.world.NewEssenceSourceNode(smallScrapCreepSource, c.pos.Add(gmath.Vec{Y: 2}))
		c.world.nodeRunner.AddObject(scraps)
	case creepCrawler:
		createExplosion(c.world, false, c.pos)
		if c.world.rand.Chance(0.3) {
			scraps := c.world.NewEssenceSourceNode(smallScrapCreepSource, c.pos.Add(gmath.Vec{Y: 2}))
			c.world.nodeRunner.AddObject(scraps)
		}
	default:
		roll := c.scene.Rand().Float()
		if roll < 0.3 {
			createExplosion(c.world, true, c.pos)
		} else {
			var scraps *essenceSourceStats
			if roll > 0.65 {
				switch c.stats.tier {
				case 1:
					scraps = smallScrapCreepSource
				case 2:
					scraps = scrapCreepSource
				case 3:
					scraps = bigScrapCreepSource
				}
			}
			shadowImg := assets.ImageNone
			if c.shadow != nil {
				shadowImg = c.shadow.ImageID()
			}

			fall := newDroneFallNode(c.world, scraps, c.stats.image, shadowImg, c.pos, agentFlightHeight)
			c.world.nodeRunner.AddObject(fall)
		}
	}
}

func (c *creepNode) OnDamage(damage gamedata.DamageValue, source targetable) {
	if damage.Health != 0 {
		c.flashComponent.flash = 0.2
	}

	c.health -= damage.Health
	if c.health < 0 {
		c.explode()
		c.Destroy()
		return
	}

	if damage.Aggro != 0 {
		if c.scene.Rand().Chance(damage.Aggro) {
			c.aggroTarget = source
			c.aggro = 3.0
		}
	}

	c.slow = gmath.ClampMax(c.slow+damage.Slow, 5)

	if damage.Disarm != 0 && c.stats.disarmable {
		if c.scene.Rand().Chance(damage.Disarm * 0.1) {
			c.disarm = 2.5
			c.world.nodeRunner.AddObject(newEffectNode(c.world.camera, c.pos, c.IsFlying(), assets.ImageIonZap))
			playIonExplosionSound(c.world, c.pos)
		}
	}

	if c.stats.kind == creepCrawler {
		if c.specialModifier == crawlerGuard {
			c.specialModifier = crawlerIdle
		}
		if c.specialModifier == crawlerIdle && (c.pos.DistanceTo(*source.GetPos()) > c.stats.weapon.AttackRange*0.8) && c.world.rand.Chance(0.45) {
			followPos := c.pos.MoveTowards(*source.GetPos(), 64*c.world.rand.FloatRange(0.8, 1.4))
			p := c.world.BuildPath(c.pos, followPos)
			c.specialModifier = crawlerMove
			c.path = p.Steps
			c.waypoint = c.world.pathgrid.AlignPos(c.pos)
		}
		return
	}

	if damage.Morale != 0 && c.stats.canBeRepelled && c.stats.kind != creepServant && c.stats.kind != creepBuilder {
		if c.wasRetreating {
			return
		}
		if c.scene.Rand().Chance(damage.Morale * 0.15) {
			c.wasAttacking = true
			c.retreatFrom(*source.GetPos())
		}
	}

	if c.stats.kind == creepUberBoss {
		if a, ok := source.(*colonyAgentNode); ok && a.IsTurret() {
			c.world.result.EnemyColonyDamageFromTurrets += damage.Health
		} else {
			c.world.result.EnemyColonyDamage += damage.Health
		}
		if c.IsFlying() {
			// Stage 0: send 2 servants. (very easy and above)
			// Stage 1: send 3 servants. (easy and above)
			// Stage 2: send 4 servants. (normal and above)
			// Stage 3: send 5 servants. (hard)
			maxStage := c.world.config.BossDifficulty
			if c.bossStage <= maxStage {
				hpPercentage := c.health / c.maxHealth
				if hpPercentage < 0.8 && c.bossStage == 0 {
					c.spawnServants(2)
					c.bossStage++
				}
				if hpPercentage < 0.6 && c.bossStage == 1 {
					c.spawnServants(3)
					c.bossStage++
				}
				if hpPercentage < 0.4 && c.bossStage == 2 {
					c.spawnServants(4)
					c.bossStage++
				}
				if hpPercentage < 0.2 && c.bossStage == 3 {
					c.spawnServants(5)
					c.bossStage++
				}
			}
		}
	}
}

func (c *creepNode) spawnServants(n int) {
	startAngle := gmath.DegToRad(180 + 45)
	endAngle := gmath.DegToRad(360 - 45)
	angleDelta := endAngle - startAngle
	angleStep := gmath.Rad(float64(angleDelta) / float64(n-1))
	angle := startAngle
	for i := 0; i < n; i++ {
		dir := gmath.RadToVec(angle)
		spawn := newServantSpawnerNode(c.world, c.pos, dir, c.world.colonies[0])
		c.world.nodeRunner.AddObject(spawn)
		angle += angleStep
	}
}

func (c *creepNode) doAttack(target targetable) {
	if c.stats.weapon.ProjectileImage != assets.ImageNone {
		targetVelocity := target.GetVelocity()
		for i := 0; i < c.stats.weapon.BurstSize; i++ {
			burstCorrectedPos := *target.GetPos()
			if i != 0 {
				burstCorrectedPos = burstCorrectedPos.Add(targetVelocity.Mulf(c.stats.weapon.BurstDelay))
			}
			toPos := snipePos(c.stats.weapon.ProjectileSpeed, c.pos, burstCorrectedPos, targetVelocity)
			fireDelay := float64(i) * c.stats.weapon.BurstDelay
			p := newProjectileNode(projectileConfig{
				World:     c.world,
				Weapon:    c.stats.weapon,
				Attacker:  c,
				ToPos:     toPos.Add(c.scene.Rand().Offset(-4, 4)),
				Target:    target,
				FireDelay: fireDelay,
			})
			c.world.nodeRunner.AddObject(p)
		}
		return
	}

	if c.stats.beamTexture == nil {
		beam := newBeamNode(c.world, ge.Pos{Base: &c.spritePos}, ge.Pos{Base: target.GetPos()}, c.stats.beamColor)
		beam.width = c.stats.beamWidth
		c.world.nodeRunner.AddObject(beam)
	} else {
		beam := newTextureBeamNode(c.world, ge.Pos{Base: &c.spritePos}, ge.Pos{Base: target.GetPos()}, c.stats.beamTexture, c.stats.beamSlideSpeed, c.stats.beamOpaqueTime)
		c.world.nodeRunner.AddObject(beam)
	}

	if c.stats.kind == creepDominator {
		targetDir := c.pos.DirectionTo(*target.GetPos())
		const deg90rad = 1.5708
		vec1 := targetDir.Rotated(deg90rad).Mulf(4)
		vec2 := targetDir.Rotated(-deg90rad).Mulf(4)

		rearBeam1pos := ge.Pos{Base: &c.spritePos, Offset: vec1}
		rearBeam1targetPos := ge.Pos{Base: target.GetPos(), Offset: vec2}
		rearBeam1 := newBeamNode(c.world, rearBeam1pos, rearBeam1targetPos, dominatorBeamColorRear)
		c.world.nodeRunner.AddObject(rearBeam1)

		rearBeam2pos := ge.Pos{Base: &c.spritePos, Offset: vec2}
		rearBeam2targetPos := ge.Pos{Base: target.GetPos(), Offset: vec1}
		rearBeam2 := newBeamNode(c.world, rearBeam2pos, rearBeam2targetPos, dominatorBeamColorRear)
		c.world.nodeRunner.AddObject(rearBeam2)
	}

	target.OnDamage(c.stats.weapon.Damage, c)
}

func (c *creepNode) retreatFrom(pos gmath.Vec) {
	c.setWaypoint(retreatPos(c.scene.Rand(), c.scene.Rand().FloatRange(300, 500), c.pos, pos))
	c.wasAttacking = false
}

func (c *creepNode) SendTo(pos gmath.Vec) {
	if c.IsFlying() {
		c.setWaypoint(pos)
		return
	}

	p := c.world.BuildPath(c.pos, pos)
	c.path = p.Steps
	c.waypoint = c.world.pathgrid.AlignPos(c.pos)
	switch c.stats.kind {
	case creepCrawler:
		c.specialModifier = crawlerMove
	case creepHowitzer:
		c.specialModifier = howitzerMove
	}

	if c.stats == stealthCrawlerCreepStats {
		c.doCloak()
	}
}

func (c *creepNode) findTargets() []targetable {
	targets := c.world.tmpTargetSlice[:0]
	if c.aggro > 0 && c.aggroTarget != nil {
		if c.aggroTarget.IsDisposed() || c.pos.DistanceSquaredTo(*c.aggroTarget.GetPos()) > c.stats.weapon.AttackRangeSqr {
			c.aggroTarget = nil
		} else {
			targets = append(targets, c.aggroTarget)
			if len(targets) >= c.stats.weapon.MaxTargets {
				return targets
			}
		}
	}

	c.world.FindColonyAgent(c.pos, c.stats.weapon.AttackRange, func(a *colonyAgentNode) bool {
		targets = append(targets, a)
		return len(targets) >= c.stats.weapon.MaxTargets
	})
	if c.stats.weapon.Damage.Health == 0 {
		return targets
	}

	if len(targets) >= c.stats.weapon.MaxTargets {
		return targets
	}
	if c.stats.weapon.TargetFlags&gamedata.TargetGround == 0 {
		return targets
	}

	for _, colony := range c.world.constructions {
		if len(targets) >= c.stats.weapon.MaxTargets {
			return targets
		}
		if colony.pos.DistanceTo(c.pos) > c.stats.weapon.AttackRange {
			continue
		}
		targets = append(targets, colony)
	}

	if len(targets) >= c.stats.weapon.MaxTargets {
		return targets
	}
	for _, colony := range c.world.colonies {
		if len(targets) >= c.stats.weapon.MaxTargets {
			return targets
		}
		if colony.pos.DistanceTo(c.pos) > c.stats.weapon.AttackRange {
			continue
		}
		targets = append(targets, colony)
	}

	return targets
}

func (c *creepNode) wandererMovement(delta float64) {
	if c.waypoint.IsZero() {
		c.wasRetreating = false
		// Choose a waypoint.
		if c.wasAttacking && c.scene.Rand().Chance(0.8) {
			// Go away from the colony.
			c.retreatFrom(c.pos)
		} else if c.scene.Rand().Chance(0.4) {
			// Go somewhere near a random colony.
			if len(c.world.colonies) == 0 {
				return // Waiting for a game over?
			}
			colony := gmath.RandElem(c.scene.Rand(), c.world.colonies)
			c.setWaypoint(colony.pos.Add(c.scene.Rand().Offset(-200, 200)))
			c.wasAttacking = true
		} else if c.scene.Rand().Chance(0.3) {
			// Go to a random screen location.
			pos := gmath.Vec{
				X: c.scene.Rand().FloatRange(0, c.world.width),
				Y: c.scene.Rand().FloatRange(0, c.world.height),
			}
			c.setWaypoint(pos)
			c.wasAttacking = false
		} else {
			c.setWaypoint(c.pos.Add(c.scene.Rand().Offset(-100, 100)))
			c.wasAttacking = false
		}
	}

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
		if c.shadow != nil {
			c.shadow.Visible = true
		}
		c.height = agentFlightHeight
	}
}

func (c *creepNode) updatePrimitiveWanderer(delta float64) {
	if c.anim != nil {
		c.anim.Tick(delta)
	}
	c.wandererMovement(delta)
}

func (c *creepNode) canBuildHere(pos gmath.Vec) bool {
	const pad float64 = 196
	if pos.X < pad || pos.Y < pad || pos.X > (c.world.width-pad) || pos.Y > (c.world.height-pad) {
		return false
	}
	return posIsFree(c.world, nil, pos, 80)
}

func (c *creepNode) updateBuilder(delta float64) {
	c.anim.Tick(delta)
	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)

	// It regenerates 1 health over 5 seconds (*0.2).
	// 12 hp over minute.
	c.health = gmath.ClampMax(c.health+(delta*0.2), c.maxHealth)

	if c.specialTarget != nil {
		// Building in progress.
		turret := c.specialTarget.(*creepNode)
		if turret.IsDisposed() {
			c.specialTarget = nil
			c.specialDelay = c.scene.Rand().FloatRange(30, 50)
			c.EventBuildingStop.Emit(gsignal.Void{})
			return
		}
		if turret.stats.kind == creepTurret || turret.stats.kind == creepCrawlerBase {
			// Constructed successfully.
			c.specialTarget = nil
			c.specialDelay = c.scene.Rand().FloatRange(70, 120)
			c.EventBuildingStop.Emit(gsignal.Void{})
			return
		}
		return
	}

	if c.waypoint.IsZero() {
		turretPos := c.pos.Add(gmath.Vec{Y: agentFlightHeight})
		if c.specialDelay == 0 && c.canBuildHere(turretPos) {
			// Start building.
			buildingStats := turretConstructionCreepStats
			if c.scene.Rand().Chance(0.35) {
				buildingStats = crawlerBaseConstructionCreepStats
			}
			turret := c.world.NewCreepNode(turretPos, buildingStats)
			turret.specialTarget = c
			c.specialTarget = turret
			c.world.nodeRunner.AddObject(turret)
			lasers := newBuilderLaserNode(c.world.camera, c.pos)
			c.EventBuildingStop.Connect(lasers, lasers.OnBuildingStop)
			c.world.nodeRunner.AddObject(lasers)
			return
		}
		c.waypoint = correctedPos(c.world.rect, randomSectorPos(c.world.rand, c.world.rect), 400)
	}

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
	}
}

func (c *creepNode) crawlerSpawnPos() gmath.Vec {
	return c.pos.Add(gmath.Vec{Y: agentFlightHeight - 20})
}

func (c *creepNode) maybeSpawnCrawlers() bool {
	if c.world.NumActiveCrawlers() >= c.world.MaxActiveCrawlers() {
		return false
	}

	spawnPos := c.crawlerSpawnPos()
	if !posIsFree(c.world, nil, spawnPos, 64) {
		return false
	}

	minCrawlers := 2
	maxCrawlers := 3
	switch c.world.config.BossDifficulty {
	case 2:
		maxCrawlers = 5
	case 3:
		minCrawlers = 3
		maxCrawlers = 5
	}

	c.specialModifier = float64(c.scene.Rand().IntRange(minCrawlers, maxCrawlers)) + 1
	return true
}

func (c *creepNode) isNearEnemyBase(dist float64) bool {
	distSqr := dist * dist
	for _, colony := range c.world.colonies {
		if colony.pos.DistanceSquaredTo(c.pos) < distSqr {
			return true
		}
		if !c.cloaking {
			for _, turret := range colony.turrets {
				if turret.pos.DistanceSquaredTo(c.pos) < distSqr {
					return true
				}
			}
		}
	}
	return false
}

func (c *creepNode) findHowitzerTarget(rangeMultiplier float64) targetable {
	const minAttackRangeSqe float64 = 160 * 160
	var target targetable
	maxAttackRangeSqr := rangeMultiplier * c.stats.specialWeapon.AttackRangeSqr
	randIterate(c.world.rand, c.world.colonies, func(colony *colonyCoreNode) bool {
		if !colony.IsFlying() {
			distSqr := colony.pos.DistanceSquaredTo(c.pos)
			canAttack := distSqr > minAttackRangeSqe && distSqr < maxAttackRangeSqr
			if canAttack {
				target = colony
				return true
			}
		}
		turretTarget := randIterate(c.world.rand, colony.turrets, func(turret *colonyAgentNode) bool {
			distSqr := turret.pos.DistanceSquaredTo(c.pos)
			return distSqr > minAttackRangeSqe && distSqr < maxAttackRangeSqr
		})
		if turretTarget != nil {
			target = turretTarget
			return true
		}
		return false
	})
	return target
}

func (c *creepNode) updateHowitzer(delta float64) {
	if !c.waypoint.IsZero() {
		c.anim.Tick(delta)
		if c.moveTowards(delta, c.waypoint) {
			if c.path.HasNext() {
				if c.isNearEnemyBase(c.stats.specialWeapon.AttackRange * 0.8) {
					c.path = pathing.GridPath{}
				}
			}
			if c.path.HasNext() {
				d := c.path.Next()
				aligned := c.world.pathgrid.AlignPos(c.pos)
				c.waypoint = posMove(aligned, d).Add(c.world.rand.Offset(-4, 4))
				return
			}
			c.specialDelay = 0
			c.specialModifier = howitzerIdle
			c.waypoint = gmath.Vec{}
			return
		}
	}

	if c.specialModifier == howitzerReady {
		c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
		if c.specialDelay == 0 {
			target := c.findHowitzerTarget(1.0)
			c.specialDelay = c.stats.specialWeapon.Reload * c.world.rand.FloatRange(0.8, 1.2)
			if target != nil && c.world.rand.Chance(0.9) {
				targetPos := *target.GetPos()
				dir := c.pos.AngleToPoint(targetPos).Normalized()
				trunk := c.specialTarget.(*howitzerTrunkNode)
				fireOffset := trunk.SetRotation(dir)
				p := newProjectileNode(projectileConfig{
					World:      c.world,
					Weapon:     c.stats.specialWeapon,
					Attacker:   c,
					ToPos:      targetPos,
					Target:     target,
					FireOffset: fireOffset,
				})
				c.world.nodeRunner.AddObject(p)
			} else if c.world.rand.Chance(0.3) {
				c.specialModifier = howitzerFoldTurret
				c.sprite.Visible = false
				c.altSprite.Visible = true
				c.anim.Mode = ge.AnimationBackward
				c.anim.SetSprite(c.altSprite, -1)
				c.anim.Rewind()
				trunk := c.specialTarget.(*howitzerTrunkNode)
				trunk.SetVisibility(false)
			}
		}
		return
	}

	if c.specialModifier == howitzerFoldTurret {
		if c.anim.Tick(delta) {
			c.specialDelay = c.world.rand.FloatRange(3, 10)
			c.specialModifier = howitzerIdle
			c.altSprite.Visible = false
			c.sprite.Visible = true
			c.anim.Mode = ge.AnimationForward
			c.anim.SetSprite(c.sprite, 4)
			c.anim.Rewind()
		}
		return
	}

	if c.specialModifier == howitzerPreparing {
		if c.anim.Tick(delta) {
			c.specialModifier = howitzerReady
			c.specialDelay = c.world.rand.FloatRange(2, 5)
			c.sprite.Visible = true
			c.sprite.FrameOffset.X = float64(c.sprite.FrameWidth) * 4
			c.altSprite.Visible = false
			trunk := c.specialTarget.(*howitzerTrunkNode)
			trunk.SetVisibility(true)
			trunk.SetRotation(math.Pi + (math.Pi * 0.5))
		}
		return
	}

	if c.specialModifier == howitzerIdle {
		hasTargets := c.findHowitzerTarget(1.1) != nil
		if c.specialDelay == 0 && (hasTargets || c.world.rand.Chance(0.25)) {
			c.specialModifier = howitzerPreparing
			c.anim.SetSprite(c.altSprite, -1)
			c.anim.Rewind()
			c.sprite.Visible = false
			c.altSprite.Visible = true
		} else {
			dist := c.world.rand.FloatRange(96, 256)
			dst := gmath.RadToVec(c.world.rand.Rad()).Mulf(dist).Add(c.pos)
			c.SendTo(dst)
		}
		return
	}
}

func (c *creepNode) updateCrawler(delta float64) {
	if c.waypoint.IsZero() && c.cloaking {
		c.doUncloak()
	}

	if !c.waypoint.IsZero() {
		c.anim.Tick(delta)
		if c.moveTowards(delta, c.waypoint) {
			// To avoid weird cases of walking above colony core or turret,
			// stop if there are any targets in vicinity.
			if c.path.HasNext() {
				if c.isNearEnemyBase(96) {
					c.path = pathing.GridPath{}
				}
			}
			if c.path.HasNext() {
				d := c.path.Next()
				aligned := c.world.pathgrid.AlignPos(c.pos)
				c.waypoint = posMove(aligned, d).Add(c.world.rand.Offset(-4, 4))
				return
			}
			c.specialModifier = crawlerIdle
			c.waypoint = gmath.Vec{}
			return
		}
	}
}

func (c *creepNode) updateTurretConstruction(delta float64) {
	// If builder is defeated, this turret should be removed.
	if c.specialTarget == nil || c.specialTarget.(*creepNode).IsDisposed() {
		c.explode()
		c.Destroy()
	}

	if c.specialModifier >= 1 {
		resultStats := turretCreepStats
		if c.stats.kind == creepCrawlerBaseConstruction {
			resultStats = crawlerBaseCreepStats
		}
		result := c.world.NewCreepNode(c.pos, resultStats)
		c.specialTarget.(*creepNode).specialTarget = result
		c.world.nodeRunner.AddObject(result)
		c.Destroy()
		return
	}

	c.specialModifier += delta * 0.02
	c.sprite.Shader.SetFloatValue("Time", c.specialModifier)
}

func (c *creepNode) updateCreepCrawlerBase(delta float64) {
	if c.attackDelay != 0 {
		return
	}
	const maxUnits = 10
	if c.specialModifier > maxUnits {
		return
	}

	spawnPos := c.pos.Add(gmath.Vec{Y: 16})
	dstOffset := gmath.Vec{
		X: c.scene.Rand().FloatRange(-128, 128),
		Y: c.scene.Rand().FloatRange(-80, 80),
	}
	dstPos := spawnPos.Add(gmath.Vec{Y: 96}).Add(dstOffset)
	if !posIsFreeWithFlags(c.world, nil, dstPos, 8, collisionSkipSmallCrawlers) {
		c.attackDelay = c.scene.Rand().FloatRange(2, 5)
		return
	}
	c.attackDelay = c.scene.Rand().FloatRange(25, 55)
	c.specialModifier++

	crawler := c.world.NewCreepNode(spawnPos, crawlerCreepStats)
	crawler.SendTo(dstPos)
	crawler.waypoint = crawler.pos
	c.world.nodeRunner.AddObject(crawler)
	crawler.EventDestroyed.Connect(c, func(arg *creepNode) {
		c.specialModifier--
	})
}

func (c *creepNode) updateCreepBase(delta float64) {
	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 && c.specialModifier < 15 {
		c.specialDelay = c.scene.Rand().FloatRange(80, 120)
		c.specialModifier += 1 // base level up
	}

	level := int(c.specialModifier)
	numSpawned := 0
	switch level {
	case 0:
		// The base is inactive.
		c.sprite.FrameOffset.X = 0
		return
	case 1, 2, 3:
		c.sprite.FrameOffset.X = 32 * 1
		numSpawned = 1
	case 4, 5, 6:
		c.sprite.FrameOffset.X = 32 * 2
		numSpawned = 2
	default:
		c.sprite.FrameOffset.X = 32 * 3
		numSpawned = 3
	}

	if c.attackDelay != 0 {
		return
	}

	spawnDelay := c.scene.Rand().FloatRange(65, 80)
	if level > 11 {
		spawnDelay *= 0.9
	} else if level >= 14 {
		spawnDelay *= 0.75
	}
	c.attackDelay = spawnDelay

	spawnPoints := [...]gmath.Vec{
		c.pos.Add(gmath.Vec{X: -5, Y: -5}),
		c.pos.Add(gmath.Vec{X: 8, Y: -2}),
		c.pos.Add(gmath.Vec{X: -5, Y: 6}),
	}
	waypointOffsets := [...]gmath.Vec{
		{X: -9, Y: -36},
		{X: 8, Y: -30},
		{X: -1, Y: -28},
	}
	tier2chance := 0.0
	if level >= 4 {
		tier2chance = 0.3
	} else if level >= 7 {
		tier2chance = 0.45
	} else if level >= 10 {
		tier2chance = 0.6
	}
	for i := 0; i < numSpawned; i++ {
		spawnPos := spawnPoints[i]
		waypoint := c.pos.Add(waypointOffsets[i])

		stats := wandererCreepStats

		if tier2chance != 0 && c.scene.Rand().Chance(tier2chance) {
			stats = stunnerCreepStats
		}

		creep := c.world.NewCreepNode(spawnPos, stats)
		creep.waypoint = waypoint
		creep.spawnedFromBase = true
		c.world.nodeRunner.AddObject(creep)
		creep.height = 0
	}
}

func (c *creepNode) updateServant(delta float64) {
	c.anim.Tick(delta)

	if c.specialTarget == nil && len(c.world.colonies) == 0 {
		return
	}

	target, ok := c.specialTarget.(*colonyCoreNode)
	if !ok || target == nil || target.IsDisposed() {
		if len(c.world.colonies) == 0 {
			c.specialTarget = nil
			return
		}
		c.specialTarget = gmath.RandElem(c.world.rand, c.world.colonies)
		return
	} else {
		// Fly around the target base.
		const maxDistSqr float64 = 196 * 196
		if c.waypoint.IsZero() || c.waypoint.DistanceSquaredTo(target.pos) > maxDistSqr {
			c.waypoint = target.pos.Add(c.scene.Rand().Offset(-164, 164))
		}
		if c.moveTowards(delta, c.waypoint) {
			c.waypoint = gmath.Vec{}
		}
	}

	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 && c.disarm == 0 {
		c.specialDelay = c.scene.Rand().FloatRange(4, 6)
		wave := newServantWaveNode(c)
		c.world.nodeRunner.AddObject(wave)
		playSound(c.world, assets.AudioServantWave, c.pos)
	}
}

func (c *creepNode) updateUberBoss(delta float64) {
	c.anim.Tick(delta)

	if c.shadow != nil {
		c.shadow.Pos.Offset.Y = math.Round(c.height + 4)
		newShadowAlpha := float32(1.0 - ((c.height / agentFlightHeight) * 0.5))
		c.shadow.SetAlpha(newShadowAlpha)
	}

	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 && c.specialModifier == 0 {
		if c.maybeSpawnCrawlers() {
			c.pos.X = math.Round(c.pos.X)
			c.pos.Y = math.Round(c.pos.Y)
			// Time until the first crawler is spawned.
			c.specialDelay = c.scene.Rand().FloatRange(7, 10)
		} else {
			c.specialDelay = c.scene.Rand().FloatRange(6, 16)
		}
	}

	// It regenerates 1 health over 4 seconds (*0.25).
	// Meaning it's 15 health per minute.
	// In other words, 10 minutes recover 150 health for this guy.
	c.health = gmath.ClampMax(c.health+(delta*0.25), c.maxHealth)

	const crawlersSpawnHeight float64 = 10
	if c.specialModifier != 0 && c.height != crawlersSpawnHeight {
		c.height -= delta * 5
		c.pos.Y += delta * 5
		if c.height <= crawlersSpawnHeight {
			c.height = crawlersSpawnHeight
			c.sprite.Visible = false
			c.altSprite.Visible = true
			c.flashComponent.sprite = c.altSprite
		}
		return
	}
	if c.specialModifier != 0 && c.height == crawlersSpawnHeight {
		if c.specialDelay == 0 {
			// Time until next crawler is spawned.
			c.specialDelay = c.scene.Rand().FloatRange(2, 4)
			c.specialModifier--
			if c.specialModifier > 0 {
				spawnPos := c.crawlerSpawnPos()
				crawlerStats := crawlerCreepStats
				eliteChance := c.world.EliteCrawlerChance()
				if c.world.rand.Chance(eliteChance) {
					if c.world.rand.Chance(0.3) {
						crawlerStats = heavyCrawlerCreepStats
					} else {
						crawlerStats = eliteCrawlerCreepStats
					}
				}
				crawler := c.world.NewCreepNode(spawnPos, crawlerStats)
				crawler.path = crawlerSpawnPositions[int(c.specialModifier-1)]
				crawler.waypoint = crawler.pos
				c.world.nodeRunner.AddObject(crawler)
			}
		}
		if c.specialModifier == 0 {
			c.specialDelay = c.scene.Rand().FloatRange(50, 70)
			c.sprite.Visible = true
			c.altSprite.Visible = false
			c.flashComponent.sprite = c.sprite
			return
		}
		return
	}

	if c.height != agentFlightHeight {
		c.height += delta * 5
		c.pos.Y -= delta * 5
		if c.height >= agentFlightHeight {
			c.height = agentFlightHeight
			c.pos.X = math.Round(c.pos.X)
			c.pos.Y = math.Round(c.pos.Y)
		}
		return
	}

	if c.waypoint.IsZero() {
		pos := gmath.Vec{
			X: c.scene.Rand().FloatRange(0, c.world.width),
			Y: c.scene.Rand().FloatRange(0, c.world.height),
		}
		c.waypoint = correctedPos(c.world.rect, pos, 96)
	}

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
	}
}

func (c *creepNode) setWaypoint(pos gmath.Vec) {
	c.waypoint = correctedPos(c.world.rect, pos, 8)
}

func (c *creepNode) IsCloaked() bool {
	return c.cloaking
}

func (c *creepNode) doUncloak() {
	c.cloaking = false
	c.sprite.SetAlpha(1)
}

func (c *creepNode) doCloak() {
	c.cloaking = true
	c.sprite.SetAlpha(0.2)
	c.world.nodeRunner.AddObject(newEffectNode(c.world.camera, c.pos, true, assets.ImageCloakWave))
}

func (c *creepNode) movementSpeed() float64 {
	if c.spawnedFromBase && c.height == 0 {
		return c.stats.speed * 0.5
	}
	multiplier := 1.0
	if c.slow > 0 {
		multiplier = 0.55
	}
	return c.stats.speed * multiplier
}

func (c *creepNode) moveTowards(delta float64, pos gmath.Vec) bool {
	travelled := c.movementSpeed() * delta
	if c.pos.DistanceTo(pos) <= travelled {
		c.pos = pos
		return true
	}
	c.pos = c.pos.MoveTowards(pos, travelled)
	return false
}
