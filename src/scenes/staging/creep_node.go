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

	scene *ge.Scene

	shadowComponent shadowComponent
	flashComponent  damageFlashComponent

	world *worldState
	stats *gamedata.CreepStats

	aggroTarget targetable

	spawnPos        gmath.Vec
	pos             gmath.Vec
	waypoint        gmath.Vec
	wasAttacking    bool
	wasRetreating   bool
	spawnedFromBase bool
	cloaking        bool
	insideForest    bool
	super           bool

	path            pathing.GridPath
	specialTarget   any
	specialDelay    float64
	specialModifier float64

	aggro     float64
	disarm    float64
	slow      float64
	health    float64
	maxHealth float64
	marked    float64

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

	pathing.MakeGridPath(pathing.DirDown, pathing.DirLeft, pathing.DirUp, pathing.DirUp),
	pathing.MakeGridPath(pathing.DirDown, pathing.DirRight, pathing.DirUp, pathing.DirUp),
}

func newCreepNode(world *worldState, stats *gamedata.CreepStats, pos gmath.Vec) *creepNode {
	return &creepNode{
		world:    world,
		stats:    stats,
		pos:      pos,
		spawnPos: pos,
	}
}

func (c *creepNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.maxHealth = c.stats.MaxHealth
	if c.super {
		if c.stats.Kind == gamedata.CreepUberBoss {
			// Boss only gets a bit of extra health.
			c.maxHealth *= 1.3
		} else {
			c.maxHealth = (c.maxHealth * 2) + 10
		}
	}

	if c.stats.ShadowImage != assets.ImageNone {
		if c.world.graphicsSettings.ShadowsEnabled {
			c.shadowComponent.Init(c.world, c.stats.ShadowImage)
			c.shadowComponent.SetVisibility(!c.spawnedFromBase)
		}
		if c.stats.Kind == gamedata.CreepUberBoss {
			c.shadowComponent.offset = 4
		} else {
			c.shadowComponent.offset = 2
		}
		c.shadowComponent.UpdatePos(c.pos)
		c.shadowComponent.UpdateHeight(c.pos, agentFlightHeight, agentFlightHeight)
	}

	c.sprite = scene.NewSprite(c.stats.Image)
	c.sprite.Pos.Base = &c.pos
	if c.stats.Kind == gamedata.CreepUberBoss {
		c.world.stage.AddSpriteSlightlyAbove(c.sprite)
	} else if c.stats.ShadowImage != assets.ImageNone {
		c.world.stage.AddSpriteAbove(c.sprite)
	} else {
		c.world.stage.AddSprite(c.sprite)
	}
	if c.stats.AnimSpeed != 0 {
		if c.stats.Kind == gamedata.CreepHowitzer {
			// 4 frames for the walk, 1 frame is for the "ready" state.
			c.anim = ge.NewRepeatedAnimation(c.sprite, 4)
		} else {
			c.anim = ge.NewRepeatedAnimation(c.sprite, -1)
		}
		c.anim.SetSecondsPerFrame(c.stats.AnimSpeed)
		if c.super {
			c.anim.SetOffsetY(c.sprite.FrameHeight)
		}
		c.anim.Tick(c.world.localRand.FloatRange(0, 0.7))
	} else if c.super {
		c.sprite.FrameOffset.Y = c.sprite.FrameHeight
	}
	c.flashComponent.sprite = c.sprite

	if c.stats.Kind == gamedata.CreepUberBoss {
		c.altSprite = scene.NewSprite(assets.ImageUberBossDoor)
		c.altSprite.Visible = false
		c.altSprite.Pos.Base = &c.pos
		c.altSprite.Pos.Offset.Y = 13
		c.world.stage.AddSprite(c.altSprite)
		c.maxHealth *= c.world.bossHealthMultiplier
	} else {
		c.maxHealth *= c.world.creepHealthMultiplier
		if c.stats.Kind == gamedata.CreepHowitzer {
			c.altSprite = scene.NewSprite(assets.ImageHowitzerPreparing)
			c.altSprite.Visible = false
			c.altSprite.Pos.Base = &c.pos
			c.altSprite.Pos.Offset.Y = -3
			c.world.stage.AddSprite(c.altSprite)
		}
	}
	switch c.stats.Kind {
	case gamedata.CreepUberBoss:
		if !c.world.simulation && c.world.graphicsSettings.AllShadersEnabled {
			c.sprite.Shader = c.scene.NewShader(assets.ShaderColonyDamage)
			c.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageDreadnoughtDamageMask)
			c.sprite.Shader.SetFloatValue("HP", 1.0)
			c.sprite.Shader.Enabled = false
		}
	case gamedata.CreepFortress:
		c.specialDelay = c.scene.Rand().FloatRange(3.5*60, 4.5*60)
	case gamedata.CreepWispLair:
		c.attackDelay = c.scene.Rand().FloatRange(30, 50)
	case gamedata.CreepWisp:
		c.specialDelay = c.scene.Rand().FloatRange(2, 20)
	case gamedata.CreepServant:
		c.specialDelay = c.scene.Rand().FloatRange(0.5, 3)
	case gamedata.CreepBuilder:
		c.specialDelay = c.scene.Rand().FloatRange(15, 30)
	case gamedata.CreepCrawlerBase:
		c.attackDelay = c.scene.Rand().FloatRange(5, 10)
	case gamedata.CreepTurretConstruction, gamedata.CreepCrawlerBaseConstruction:
		if !c.world.simulation {
			c.sprite.Shader = scene.NewShader(assets.ShaderCreepTurretBuild)
			c.sprite.Shader.SetFloatValue("Time", 0)
		}
	case gamedata.CreepHowitzer:
		pos := ge.Pos{Base: &c.pos, Offset: gmath.Vec{Y: -10}}
		trunk := newHowitzerTrunkNode(c.world.stage, pos)
		c.specialTarget = trunk
		c.world.nodeRunner.AddObject(trunk)
		trunk.SetVisibility(false)
		c.specialDelay = c.scene.Rand().FloatRange(20, 30)
	}

	c.health = c.maxHealth
}

func (c *creepNode) updateHealthShader() {
	if c.sprite.Shader.IsNil() {
		return
	}
	percentage := c.health / c.maxHealth
	c.sprite.Shader.SetFloatValue("HP", percentage)
	c.sprite.Shader.Enabled = percentage < 0.95
}

func (c *creepNode) dispose() {
	c.sprite.Dispose()
	c.shadowComponent.Dispose()
	if c.altSprite != nil {
		c.altSprite.Dispose()
	}

	if c.stats.Kind == gamedata.CreepHowitzer {
		trunk := c.specialTarget.(*howitzerTrunkNode)
		trunk.Dispose()
	}
}

func (c *creepNode) Destroy() {
	c.EventDestroyed.Emit(c)
	if c.stats.Kind == gamedata.CreepBuilder && c.specialTarget != nil {
		c.EventBuildingStop.Emit(gsignal.Void{})
	}
	c.dispose()
}

func (c *creepNode) IsDisposed() bool { return c.sprite.IsDisposed() }

func (c *creepNode) Update(delta float64) {
	c.flashComponent.Update(delta)

	c.marked = gmath.ClampMin(c.marked-delta, 0)
	c.slow = gmath.ClampMin(c.slow-delta, 0)
	c.aggro = gmath.ClampMin(c.aggro-delta, 0)
	c.disarm = gmath.ClampMin(c.disarm-delta, 0)
	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	if c.stats.Weapon != nil && c.attackDelay == 0 && c.disarm == 0 && !c.cloaking && !c.insideForest {
		c.attackDelay = c.stats.Weapon.Reload * c.scene.Rand().FloatRange(0.8, 1.2)
		targets := c.findTargets()
		weapon := c.stats.Weapon
		if c.super && c.stats.SuperWeapon != nil {
			weapon = c.stats.SuperWeapon
		}
		if len(targets) != 0 {
			for _, target := range targets {
				c.doAttack(target, weapon)
			}
			if !weapon.ProjectileFireSound {
				playSound(c.world, weapon.AttackSound, c.pos)
			}
		}
	}

	switch c.stats.Kind {
	case gamedata.CreepPrimitiveWanderer, gamedata.CreepStunner, gamedata.CreepAssault:
		c.updatePrimitiveWanderer(delta)
	case gamedata.CreepDominator:
		c.updateDominator(delta)
	case gamedata.CreepBuilder:
		c.updateBuilder(delta)
	case gamedata.CreepUberBoss:
		c.updateUberBoss(delta)
	case gamedata.CreepServant:
		c.updateServant(delta)
	case gamedata.CreepBase:
		c.updateCreepBase(delta)
	case gamedata.CreepCrawlerBase:
		c.updateCreepCrawlerBase(delta)
	case gamedata.CreepCrawler:
		c.updateCrawler(delta)
	case gamedata.CreepHowitzer:
		c.updateHowitzer(delta)
	case gamedata.CreepTurret:
		// Do nothing.
	case gamedata.CreepTurretConstruction, gamedata.CreepCrawlerBaseConstruction:
		c.updateTurretConstruction(delta)
	case gamedata.CreepWisp:
		c.updateWisp(delta)
	case gamedata.CreepWispLair:
		c.updateWispLair(delta)
	case gamedata.CreepFortress:
		c.updateFortress(delta)
	case gamedata.CreepTemplar:
		c.updateTemplar(delta)
	default:
		panic("unexpected creep kind in update()")
	}
}

func (c *creepNode) GetPos() *gmath.Vec { return &c.pos }

func (c *creepNode) GetVelocity() gmath.Vec {
	if c.waypoint.IsZero() {
		return gmath.Vec{}
	}
	return c.pos.VecTowards(c.waypoint, c.movementSpeed())
}

func (c *creepNode) GetTargetInfo() targetInfo {
	return targetInfo{
		building: c.stats.Building,
		flying:   c.IsFlying(),
	}
}

func (c *creepNode) IsFlying() bool {
	if c.stats.Kind == gamedata.CreepUberBoss {
		return !c.altSprite.Visible
	}
	return c.stats.Flying
}

func (c *creepNode) TargetKind() gamedata.TargetKind {
	if c.stats.Kind == gamedata.CreepUberBoss {
		if c.IsFlying() {
			return gamedata.TargetFlying
		}
		return gamedata.TargetGround
	}
	return c.stats.TargetKind
}

func (c *creepNode) explode() {
	switch c.stats.Kind {
	case gamedata.CreepWisp:
		createEffect(c.world, effectConfig{
			Pos:   c.pos,
			Image: assets.ImageWispExplosion,
			Layer: aboveEffectLayer,
		})
	case gamedata.CreepWispLair:
		createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), normalEffectLayer)
	case gamedata.CreepUberBoss:
		if c.IsFlying() {
			shadowImg := c.shadowComponent.GetImageID()
			fall := newDroneFallNode(c.world, nil, c.stats.Image, shadowImg, c.pos, c.shadowComponent.height)
			if c.super {
				fall.FrameOffsetY = c.sprite.FrameHeight
			}
			c.world.nodeRunner.AddObject(fall)
			fall.sprite.Shader = c.sprite.Shader
		} else {
			createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), normalEffectLayer)
		}

	case gamedata.CreepTurret, gamedata.CreepBase, gamedata.CreepCrawlerBase, gamedata.CreepHowitzer:
		createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), normalEffectLayer)
		c.world.CreateScrapsAt(bigScrapCreepSource, c.pos.Add(gmath.Vec{Y: 7}))
	case gamedata.CreepFortress:
		createAreaExplosion(c.world, spriteRect(c.pos, c.sprite), normalEffectLayer)
		c.world.CreateScrapsAt(bigScrapCreepSource, c.pos.Add(gmath.Vec{Y: 4}))
		for i := 0; i < 3; i++ {
			c.world.CreateScrapsAt(scrapCreepSource, c.pos.Add(c.world.rand.Offset(-12, 12)))
		}
	case gamedata.CreepTurretConstruction, gamedata.CreepCrawlerBaseConstruction:
		createExplosion(c.world, normalEffectLayer, c.pos)
		c.world.CreateScrapsAt(smallScrapCreepSource, c.pos.Add(gmath.Vec{Y: 2}))
	case gamedata.CreepCrawler:
		createExplosion(c.world, normalEffectLayer, c.pos)
		if c.world.rand.Chance(0.3) {
			c.world.CreateScrapsAt(smallScrapCreepSource, c.pos.Add(gmath.Vec{Y: 2}))
		}
	default:
		roll := c.scene.Rand().Float()
		if roll < 0.3 {
			createExplosion(c.world, aboveEffectLayer, c.pos)
		} else {
			var scraps *essenceSourceStats
			if roll > 0.65 {
				switch c.stats.Tier {
				case 1:
					scraps = smallScrapCreepSource
				case 2:
					scraps = scrapCreepSource
				case 3:
					scraps = bigScrapCreepSource
				}
			}
			shadowImg := c.shadowComponent.GetImageID()
			fall := newDroneFallNode(c.world, scraps, c.stats.Image, shadowImg, c.pos, agentFlightHeight)
			if c.super {
				fall.FrameOffsetY = c.sprite.FrameHeight
			}
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

	if damage.Mark != 0 {
		c.marked = gmath.ClampMax(c.marked+damage.Mark, 12)
	}

	if damage.Slow != 0 {
		slowImmune := c.super && c.stats == gamedata.CrawlerCreepStats
		if !slowImmune {
			c.slow = gmath.ClampMax(c.slow+damage.Slow, 5)
		}
	}

	if damage.Disarm != 0 && c.stats.Disarmable {
		disarmImmune := c.super && c.stats == gamedata.CrawlerCreepStats
		if !disarmImmune && c.scene.Rand().Chance(damage.Disarm) {
			c.disarm = 2.5
			createEffect(c.world, effectConfig{
				Pos:   c.pos,
				Layer: effectLayerFromBool(c.IsFlying()),
				Image: assets.ImageIonZap,
			})
			playIonExplosionSound(c.world, c.pos)
		}
	}

	if damage.Morale != 0 && c.stats.CanBeRepelled {
		if c.wasRetreating {
			return
		}
		retreatChance := damage.Morale
		if c.stats.Kind == gamedata.CreepCrawler {
			retreatChance *= 0.6
		}
		if c.scene.Rand().Chance(retreatChance) {
			c.wasRetreating = true
			c.retreatFrom(*source.GetPos(), 150, 250)
		}
	}

	if c.stats.Kind == gamedata.CreepCrawler {
		if c.specialModifier == crawlerGuard {
			c.specialModifier = crawlerIdle
		}
		if c.specialModifier == crawlerIdle && (c.pos.DistanceTo(*source.GetPos()) > c.stats.Weapon.AttackRange*0.8) && c.world.rand.Chance(0.45) {
			followPos := c.pos.MoveTowards(*source.GetPos(), 64*c.world.rand.FloatRange(0.8, 1.4))
			p := c.world.BuildPath(c.pos, followPos)
			c.specialModifier = crawlerMove
			c.path = p.Steps
			c.waypoint = c.world.pathgrid.AlignPos(c.pos)
		}
		return
	}

	if c.stats.Kind == gamedata.CreepDominator && damage.Health != 0 {
		if c.wasRetreating {
			return
		}
		if c.scene.Rand().Chance(0.4) {
			c.wasRetreating = true
			c.retreatFrom(*source.GetPos(), 100, 150)
		}
	}

	if c.stats.Kind == gamedata.CreepWisp {
		if c.wasRetreating {
			return
		}
		if c.scene.Rand().Chance(0.5) {
			c.wasRetreating = true
			c.specialTarget = nil
			c.retreatFrom(*source.GetPos(), 180, 300)
		}
	}

	if c.stats.Kind == gamedata.CreepUberBoss {
		c.updateHealthShader()
		if a, ok := source.(*colonyAgentNode); ok && a.IsTurret() {
			c.world.result.EnemyColonyDamageFromTurrets += damage.Health
		} else {
			c.world.result.EnemyColonyDamage += damage.Health
		}
		colony := getUnitColony(source)
		// Stage 0: send 2 servants. (very easy and above)
		// Stage 1: send 3 servants. (easy and above)
		// Stage 2: send 4 servants. (normal and above)
		// Stage 3: send 5 servants. (hard)
		maxStage := c.world.config.BossDifficulty
		if c.bossStage <= maxStage {
			hpPercentage := c.health / c.maxHealth
			if hpPercentage < 0.8 && c.bossStage == 0 {
				c.spawnServants(2, colony)
				c.bossStage++
			}
			if hpPercentage < 0.6 && c.bossStage == 1 {
				c.spawnServants(3, colony)
				c.bossStage++
			}
			if hpPercentage < 0.4 && c.bossStage == 2 {
				c.spawnServants(4, colony)
				c.bossStage++
			}
			if hpPercentage < 0.2 && c.bossStage == 3 {
				c.spawnServants(5, colony)
				c.bossStage++
			}
		}
	}
}

func (c *creepNode) spawnServants(n int, colony *colonyCoreNode) {
	startAngle := gmath.DegToRad(180 + 45)
	endAngle := gmath.DegToRad(360 - 45)
	angleDelta := endAngle - startAngle
	angleStep := gmath.Rad(float64(angleDelta) / float64(n-1))
	angle := startAngle
	for i := 0; i < n; i++ {
		super := c.super && (i == 0 || i == n-1)
		dir := gmath.RadToVec(angle)
		spawn := newServantSpawnerNode(c.world, c.pos, dir, colony)
		spawn.super = super
		c.world.nodeRunner.AddObject(spawn)
		angle += angleStep
	}
}

func (c *creepNode) doAttack(target targetable, weapon *gamedata.WeaponStats) {
	if weapon.ProjectileImage != assets.ImageNone {
		burstSize := weapon.BurstSize
		burstDelay := weapon.BurstDelay
		if c.super && c.stats == gamedata.StealthCrawlerCreepStats {
			burstSize += 2
			burstDelay = 0.25
		}
		targetVelocity := target.GetVelocity()
		j := 0
		attacksPerBurst := weapon.AttacksPerBurst
		for i := 0; i < burstSize; i += attacksPerBurst {
			if i+attacksPerBurst > burstSize {
				// This happens only once for the last burst wave
				// if attacks-per-burst are not aligned with burstSize (like with Devourer).
				attacksPerBurst = burstSize - i
			}
			for i := 0; i < attacksPerBurst; i++ {
				burstCorrectedPos := *target.GetPos()
				if i != 0 {
					burstCorrectedPos = burstCorrectedPos.Add(targetVelocity.Mulf(burstDelay))
				}
				toPos := snipePos(weapon.ProjectileSpeed, c.pos, burstCorrectedPos, targetVelocity)
				fireDelay := float64(j) * burstDelay
				p := c.world.newProjectileNode(projectileConfig{
					World:     c.world,
					Weapon:    weapon,
					Attacker:  c,
					ToPos:     toPos.Add(c.scene.Rand().Offset(-4, 4)),
					Target:    target,
					FireDelay: fireDelay,
				})
				c.world.nodeRunner.AddProjectile(p)
			}
			j++
		}
		return
	}

	if c.stats.BeamTexture == nil {
		beam := newBeamNode(c.world, ge.Pos{Base: &c.pos}, ge.Pos{Base: target.GetPos()}, c.stats.BeamColor)
		beam.width = c.stats.BeamWidth
		c.world.nodeRunner.AddObject(beam)
	} else {
		beam := newTextureBeamNode(c.world, ge.Pos{Base: &c.pos}, ge.Pos{Base: target.GetPos()}, c.stats.BeamTexture, c.stats.BeamSlideSpeed, c.stats.BeamOpaqueTime)
		c.world.nodeRunner.AddObject(beam)
		if c.stats.BeamExplosion != assets.ImageNone {
			createEffect(c.world, effectConfig{
				Pos:   target.GetPos().Add(c.world.localRand.Offset(-6, 6)),
				Layer: effectLayerFromBool(target.IsFlying()),
				Image: c.stats.BeamExplosion,
			})
		}
	}

	if c.stats.Kind == gamedata.CreepDominator {
		targetDir := c.pos.DirectionTo(*target.GetPos())
		const deg90rad = 1.5708
		vec1 := targetDir.Rotated(deg90rad).Mulf(4)
		vec2 := targetDir.Rotated(-deg90rad).Mulf(4)

		rearBeam1pos := ge.Pos{Base: &c.pos, Offset: vec1}
		rearBeam1targetPos := ge.Pos{Base: target.GetPos(), Offset: vec2}
		rearBeam1 := newBeamNode(c.world, rearBeam1pos, rearBeam1targetPos, dominatorBeamColorRear)
		c.world.nodeRunner.AddObject(rearBeam1)

		rearBeam2pos := ge.Pos{Base: &c.pos, Offset: vec2}
		rearBeam2targetPos := ge.Pos{Base: target.GetPos(), Offset: vec1}
		rearBeam2 := newBeamNode(c.world, rearBeam2pos, rearBeam2targetPos, dominatorBeamColorRear)
		c.world.nodeRunner.AddObject(rearBeam2)
	}

	target.OnDamage(multipliedDamage(target, weapon), c)
}

func (c *creepNode) retreatFrom(pos gmath.Vec, minRange, maxRange float64) {
	c.SendTo(retreatPos(c.scene.Rand(), c.scene.Rand().FloatRange(minRange, maxRange), c.pos, pos))
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
	switch c.stats.Kind {
	case gamedata.CreepCrawler:
		c.specialModifier = crawlerMove
	case gamedata.CreepHowitzer:
		c.specialModifier = howitzerMove
	}

	if c.stats == gamedata.StealthCrawlerCreepStats {
		c.doCloak()
	}
}

func (c *creepNode) findTargets() []targetable {
	maxTargets := c.stats.Weapon.MaxTargets
	attackRangeSqr := c.stats.Weapon.AttackRangeSqr
	if c.super {
		switch c.stats.Kind {
		case gamedata.CreepAssault, gamedata.CreepPrimitiveWanderer, gamedata.CreepDominator:
			maxTargets++
		case gamedata.CreepTurret:
			attackRangeSqr *= 1.3
		case gamedata.CreepCrawler:
			if c.stats == gamedata.EliteCrawlerCreepStats {
				maxTargets += 3
			} else if c.stats == gamedata.HeavyCrawlerCreepStats {
				attackRangeSqr *= 1.2
			}
		}
	}

	targets := c.world.tmpTargetSlice[:0]
	if c.aggro > 0 && c.aggroTarget != nil {
		if c.aggroTarget.IsDisposed() || c.pos.DistanceSquaredTo(*c.aggroTarget.GetPos()) > c.stats.Weapon.AttackRangeSqr {
			c.aggroTarget = nil
		} else {
			targets = append(targets, c.aggroTarget)
			if len(targets) >= maxTargets {
				return targets
			}
		}
	}

	skipGroundTargets := c.stats.Weapon.TargetFlags&gamedata.TargetGround == 0
	c.world.FindTargetableAgents(c.pos, skipGroundTargets, c.stats.Weapon.AttackRange, func(a *colonyAgentNode) bool {
		targets = append(targets, a)
		return len(targets) >= maxTargets
	})
	if c.stats.Weapon.Damage.Health == 0 {
		return targets
	}

	if len(targets) >= maxTargets {
		return targets
	}

	if !skipGroundTargets {
		for _, colony := range c.world.constructions {
			if len(targets) >= maxTargets {
				return targets
			}
			if colony.pos.DistanceSquaredTo(c.pos) > c.stats.Weapon.AttackRangeSqr {
				continue
			}
			targets = append(targets, colony)
		}
		if len(targets) >= maxTargets {
			return targets
		}
	}

	randIterate(c.world.rand, c.world.allColonies, func(colony *colonyCoreNode) bool {
		if !colony.IsFlying() && skipGroundTargets {
			return false
		}
		if colony.pos.DistanceSquaredTo(c.pos) > c.stats.Weapon.AttackRangeSqr {
			return false
		}
		targets = append(targets, colony)
		return len(targets) >= maxTargets
	})

	return targets
}

func (c *creepNode) wandererMovement(delta float64) {
	if c.waypoint.IsZero() {
		c.wasRetreating = false
		// Choose a waypoint.
		if c.wasAttacking && c.scene.Rand().Chance(0.8) {
			// Go away from the colony.
			c.retreatFrom(c.pos, 300, 500)
		} else if c.scene.Rand().Chance(0.4) {
			// Go somewhere near a random colony.
			if len(c.world.allColonies) == 0 {
				return // Waiting for a game over?
			}
			colony := gmath.RandElem(c.scene.Rand(), c.world.allColonies)
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
		if c.spawnedFromBase {
			c.spawnedFromBase = false
			c.shadowComponent.SetVisibility(true)
			c.shadowComponent.UpdateHeight(c.pos, agentFlightHeight, agentFlightHeight)
		}
	}
}

func (c *creepNode) updatePrimitiveWanderer(delta float64) {
	if c.anim != nil {
		c.anim.Tick(delta)
	}
	c.wandererMovement(delta)
}

func (c *creepNode) updateDominator(delta float64) {
	if c.world.boss != nil && c.waypoint.IsZero() {
		if c.world.rand.Chance(0.6) {
			wp := c.world.boss.pos.Add(c.world.rand.Offset(-180, 180))
			c.setWaypoint(wp)
		}
	}

	c.wandererMovement(delta)
}

func (c *creepNode) canBuildHere(pos gmath.Vec) bool {
	const pad float64 = 196
	if pos.X < pad || pos.Y < pad || pos.X > (c.world.width-pad) || pos.Y > (c.world.height-pad) {
		return false
	}
	return posIsFree(c.world, nil, pos, 64)
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
		if turret.stats.Kind == gamedata.CreepTurret || turret.stats.Kind == gamedata.CreepCrawlerBase {
			// Constructed successfully.
			c.specialTarget = nil
			c.specialDelay = c.scene.Rand().FloatRange(70, 120)
			c.EventBuildingStop.Emit(gsignal.Void{})
			return
		}
		return
	}

	if c.waypoint.IsZero() {
		if c.specialDelay == 0 {
			turretPos := c.pos.Add(gmath.Vec{Y: agentFlightHeight})
			if c.canBuildHere(turretPos) {
				// Start building.
				buildingStats := gamedata.TurretConstructionCreepStats
				if c.scene.Rand().Chance(0.35) {
					buildingStats = gamedata.CrawlerBaseConstructionCreepStats
				}
				if c.world.config.IonMortars && buildingStats == gamedata.TurretConstructionCreepStats {
					if c.scene.Rand().Chance(0.4) {
						buildingStats = gamedata.IonMortarConstructionCreepStats
					}
				}
				turret := c.world.NewCreepNode(turretPos, buildingStats)
				turret.super = c.super && c.world.rand.Chance(0.4)
				turret.specialTarget = c
				c.specialTarget = turret
				c.world.nodeRunner.AddObject(turret)
				lasers := newBuilderLaserNode(c.world, c.pos)
				c.EventBuildingStop.Connect(lasers, lasers.OnBuildingStop)
				c.world.nodeRunner.AddObject(lasers)
				return
			}
			// Try finding a better spot nearby.
			for i := 0; i < 5; i++ {
				pos := turretPos.Add(c.world.rand.Offset(-160, 160))
				pos = c.world.AdjustCellPos(correctedPos(c.world.rect, pos, 320), 4)
				if c.canBuildHere(pos) {
					c.waypoint = pos.Sub(gmath.Vec{Y: agentFlightHeight})
					return
				}
			}
		}
		nextWaypoint := correctedPos(c.world.rect, randomSectorPos(c.world.rand, c.world.rect), 400)
		nextWaypoint = c.world.AdjustCellPos(nextWaypoint, 4).Sub(gmath.Vec{Y: agentFlightHeight})
		c.waypoint = nextWaypoint
	}

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
	}
}

func (c *creepNode) crawlerSpawnPos() gmath.Vec {
	return c.pos.Add(gmath.Vec{Y: agentFlightHeight - 20})
}

func (c *creepNode) maybeSpawnCrawlers() bool {
	var minCrawlers int
	var maxCrawlers int
	if c.world.config.GameMode == gamedata.ModeReverse {
		techLevel := c.world.creepsPlayerState.techLevel
		switch {
		case techLevel < 0.1: // 0-10%
			minCrawlers = 2
			maxCrawlers = 2
		case techLevel < 0.2: // 10-20%
			minCrawlers = 3
			maxCrawlers = 4
		case techLevel < 0.3: // 20-30%
			minCrawlers = 4
			maxCrawlers = 5
		case techLevel < 0.5: // 30-50%
			minCrawlers = 5
			maxCrawlers = 6
		case techLevel < 0.7: // 50-70%
			minCrawlers = 6
			maxCrawlers = 7
		default:
			minCrawlers = 7
			maxCrawlers = 7
		}
		minCrawlers = 7
		maxCrawlers = 7
	} else {
		minCrawlers = 3
		maxCrawlers = 4
		switch c.world.config.BossDifficulty {
		case 2:
			minCrawlers = 4
			maxCrawlers = 5
		case 3:
			minCrawlers = 5
			maxCrawlers = 7
		}
	}

	spawnPos := c.crawlerSpawnPos()
	if !posIsFree(c.world, nil, spawnPos, 64) {
		return false
	}

	c.specialModifier = float64(c.scene.Rand().IntRange(minCrawlers, maxCrawlers)) + 1
	return true
}

func (c *creepNode) isNearEnemyBase(dist float64) bool {
	distSqr := dist * dist
	for _, colony := range c.world.allColonies {
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
	const minAttackRangeSqr float64 = 160 * 160
	var target targetable
	maxAttackRangeSqr := rangeMultiplier * c.stats.SpecialWeapon.AttackRangeSqr
	if c.super {
		maxAttackRangeSqr *= 1.1
	}
	randIterate(c.world.rand, c.world.allColonies, func(colony *colonyCoreNode) bool {
		if !colony.IsFlying() {
			distSqr := colony.pos.DistanceSquaredTo(c.pos)
			canAttack := distSqr > minAttackRangeSqr && distSqr < maxAttackRangeSqr
			if canAttack {
				target = colony
				return true
			}
		}
		turretTarget := randIterate(c.world.rand, colony.turrets, func(turret *colonyAgentNode) bool {
			distSqr := turret.pos.DistanceSquaredTo(c.pos)
			return distSqr > minAttackRangeSqr && distSqr < maxAttackRangeSqr
		})
		if turretTarget != nil {
			target = turretTarget
			return true
		}
		return false
	})
	return target
}

func (c *creepNode) setAnimSprite(s *ge.Sprite, numFrames int) {
	c.anim.SetSprite(s, numFrames)
	if c.super {
		c.anim.SetOffsetY(s.FrameHeight)
	}
}

func (c *creepNode) updateHowitzer(delta float64) {
	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)

	if !c.waypoint.IsZero() {
		c.anim.Tick(delta)
		if c.moveTowards(delta, c.waypoint) {
			if c.specialDelay == 0 && c.path.HasNext() && !c.insideForest {
				if c.isNearEnemyBase(c.stats.SpecialWeapon.AttackRange * 0.8) {
					c.path = pathing.GridPath{}
				}
			}
			if c.path.HasNext() {
				d := c.path.Next()
				aligned := c.world.pathgrid.AlignPos(c.pos)
				nextPos := posMove(aligned, d)
				c.handleForestTransition(nextPos)
				c.waypoint = nextPos.Add(c.world.rand.Offset(-4, 4))
				return
			}
			c.specialModifier = howitzerIdle
			c.waypoint = gmath.Vec{}
			return
		}
	}

	if c.specialModifier == howitzerReady {
		if c.specialDelay == 0 {
			target := c.findHowitzerTarget(1.0)
			c.specialDelay = c.stats.SpecialWeapon.Reload * c.world.rand.FloatRange(0.8, 1.2)
			if target != nil && c.world.rand.Chance(0.9) {
				targetPos := *target.GetPos()
				dir := c.pos.AngleToPoint(targetPos).Normalized()
				trunk := c.specialTarget.(*howitzerTrunkNode)
				fireOffset := trunk.SetRotation(dir)
				p := c.world.newProjectileNode(projectileConfig{
					World:      c.world,
					Weapon:     c.stats.SpecialWeapon,
					Attacker:   c,
					ToPos:      targetPos,
					Target:     target,
					FireOffset: fireOffset,
				})
				c.world.nodeRunner.AddProjectile(p)
			} else if c.world.rand.Chance(0.3) {
				c.specialModifier = howitzerFoldTurret
				c.sprite.Visible = false
				c.altSprite.Visible = true
				c.anim.Mode = ge.AnimationBackward
				c.setAnimSprite(c.altSprite, -1)
				c.anim.Rewind()
				trunk := c.specialTarget.(*howitzerTrunkNode)
				trunk.SetVisibility(false)
			}
		}
		return
	}

	if c.specialModifier == howitzerFoldTurret {
		if c.anim.Tick(delta) {
			c.specialDelay = c.world.rand.FloatRange(5, 15)
			c.specialModifier = howitzerIdle
			c.altSprite.Visible = false
			c.sprite.Visible = true
			c.anim.Mode = ge.AnimationForward
			c.setAnimSprite(c.sprite, 4)
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
		if !c.insideForest && c.specialDelay == 0 && (c.findHowitzerTarget(1.1) != nil || c.world.rand.Chance(0.25)) {
			c.specialModifier = howitzerPreparing
			c.setAnimSprite(c.altSprite, -1)
			c.anim.Rewind()
			c.sprite.Visible = false
			c.altSprite.Visible = true
		} else {
			var dst gmath.Vec
			dist := c.world.rand.FloatRange(96, 256)
			if len(c.world.allColonies) != 0 && c.world.rand.Chance(0.2) {
				colony := gmath.RandElem(c.world.rand, c.world.allColonies)
				dst = colony.pos.DirectionTo(c.pos).Mulf(dist).Add(c.pos)
			} else {
				dst = gmath.RadToVec(c.world.rand.Rad()).Mulf(dist).Add(c.pos)
			}
			c.SendTo(dst)
		}
		return
	}
}

func (c *creepNode) handleForestTransition(nextWaypoint gmath.Vec) {
	if !c.world.hasForests {
		return
	}
	needEffect := false
	switch checkForestState(c.world, c.insideForest, c.pos, nextWaypoint) {
	case forestStateEnter:
		needEffect = true
		c.insideForest = true
		c.sprite.Visible = false
	case forestStateLeave:
		needEffect = true
		c.insideForest = false
		c.sprite.Visible = true
	}

	if needEffect {
		img := assets.ImageDisappearSmokeSmall
		if c.stats.Kind == gamedata.CreepHowitzer {
			img = assets.ImageDisappearSmokeBig
		}
		createEffect(c.world, effectConfig{
			Pos:            c.pos,
			Image:          img,
			AnimationSpeed: animationSpeedVerySlow,
		})
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
			if c.path.HasNext() && !c.insideForest {
				if c.isNearEnemyBase(96) {
					c.path = pathing.GridPath{}
				}
			}
			if c.path.HasNext() {
				d := c.path.Next()
				aligned := c.world.pathgrid.AlignPos(c.pos)
				nextPos := posMove(aligned, d)
				c.handleForestTransition(nextPos)
				c.waypoint = nextPos.Add(c.world.rand.Offset(-4, 4))
				return
			}
			c.specialModifier = crawlerIdle
			c.waypoint = gmath.Vec{}
			return
		}
	}
}

func (c *creepNode) updateWisp(delta float64) {
	if c.anim != nil {
		c.anim.Tick(delta)
	}

	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)

	// It regenerates 1 health over 2 seconds (*0.5).
	// 30 hp over minute.
	c.health = gmath.ClampMax(c.health+(delta*0.5), c.maxHealth)

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
		if c.specialTarget != nil {
			res := c.specialTarget.(*essenceSourceNode)
			res.Restore(c.world.rand.IntRange(5, 8))
			playSound(c.world, assets.AudioOrganicRestored, res.pos)
			createEffect(c.world, effectConfig{
				Pos:   res.pos,
				Image: assets.ImageOrganicRestored,
			})
			c.specialTarget = nil
			c.specialDelay = c.world.rand.FloatRange(20, 45)
		}
	}

	if c.attackDelay == 0 && len(c.world.allColonies) != 0 {
		const attackRange = 52
		attacking := false
		farFromTargets := true
		for _, colony := range c.world.allColonies {
			// Don't process the unlikely candidates.
			if colony.pos.DistanceSquaredTo(c.pos) > (520 * 520) {
				continue
			}
			farFromTargets = false
			colony.agents.Each(func(a *colonyAgentNode) {
				if a.IsCloaked() {
					return
				}
				if a.pos.DistanceSquaredTo(c.pos) > (attackRange * attackRange) {
					return
				}
				attacking = true
				a.OnDamage(gamedata.DamageValue{Health: 20, Energy: 100}, c)
			})
		}
		if attacking {
			c.attackDelay = c.world.rand.FloatRange(10, 20)
			createEffect(c.world, effectConfig{
				Pos:   c.pos,
				Image: assets.ImageWispShockwave,
				Layer: aboveEffectLayer,
			})
			playSound(c.world, assets.AudioWispShocker, c.pos)
		} else {
			if farFromTargets {
				c.attackDelay = c.world.rand.FloatRange(2.5, 5.75)
			} else {
				c.attackDelay = c.world.rand.FloatRange(0.5, 2.0)
			}
		}
	}

	if c.waypoint.IsZero() {
		c.wasRetreating = false
		if c.specialDelay == 0 && c.world.rand.Chance(0.80) {
			randIterate(c.world.rand, c.world.essenceSources, func(res *essenceSourceNode) bool {
				if res.stats != organicSource {
					return false
				}
				if res.resource == res.capacity {
					return false
				}
				c.specialTarget = res
				c.waypoint = res.pos.Sub(gmath.Vec{Y: agentFlightHeight})
				return true
			})
			if c.specialTarget == nil {
				c.specialDelay = c.world.rand.FloatRange(15, 60)
			}
		} else {
			if c.world.wispLair != nil && c.world.rand.Chance(0.25) {
				c.setWaypoint(c.world.wispLair.pos.Add(c.world.rand.Offset(-64, 64)))
			} else {
				c.waypoint = correctedPos(c.world.rect, randomSectorPos(c.world.rand, c.world.rect), 196)
			}
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
		var resultStats *gamedata.CreepStats
		switch c.stats {
		case gamedata.IonMortarConstructionCreepStats:
			resultStats = gamedata.IonMortarCreepStats
		case gamedata.TurretConstructionCreepStats:
			resultStats = gamedata.TurretCreepStats
			c.world.onCreepTurretBuild()
		case gamedata.CrawlerBaseConstructionCreepStats:
			resultStats = gamedata.CrawlerBaseCreepStats
		}
		result := c.world.NewCreepNode(c.pos, resultStats)
		result.super = c.super
		c.specialTarget.(*creepNode).specialTarget = result
		c.world.nodeRunner.AddObject(result)
		c.Destroy()
		return
	}

	c.specialModifier += delta * 0.02
	if !c.sprite.Shader.IsNil() {
		c.sprite.Shader.SetFloatValue("Time", c.specialModifier)
	}
}

func (c *creepNode) updateFortress(delta float64) {
	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)

	const maxUnits = 12.0
	if c.specialModifier > maxUnits {
		return
	}

	if c.specialDelay != 0 {
		return
	}
	spawnDelay := c.world.rand.FloatRange(25, 40)
	c.specialDelay = spawnDelay
	c.specialModifier++

	spawnPos := c.pos.Sub(gmath.Vec{Y: 16})
	creep := c.world.NewCreepNode(spawnPos, gamedata.TemplarCreepStats)
	creep.waypoint = spawnPos.Sub(gmath.Vec{Y: agentFlightHeight})
	creep.spawnedFromBase = true
	c.world.nodeRunner.AddObject(creep)
	creep.SetHeight(0)
	creep.EventDestroyed.Connect(c, func(arg *creepNode) {
		c.specialModifier--
	})

	createEffect(c.world, effectConfig{
		Pos:   spawnPos,
		Layer: aboveEffectLayer,
		Image: assets.ImageCreepCreatedEffect,
	})
}

func (c *creepNode) updateTemplar(delta float64) {
	if c.world.fortress == nil {
		c.wandererMovement(delta)
		return
	}

	if c.waypoint.IsZero() {
		c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
		if c.specialDelay == 0 {
			offset := gmath.Vec{X: 620, Y: 620}
			sector := gmath.Rect{
				Min: c.world.fortress.pos.Sub(offset),
				Max: c.world.fortress.pos.Add(offset),
			}
			c.setWaypoint(randomSectorPos(c.world.rand, sector))
		}
		return
	}

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
		if c.spawnedFromBase {
			c.spawnedFromBase = false
			c.shadowComponent.SetVisibility(true)
			c.shadowComponent.UpdateHeight(c.pos, agentFlightHeight, agentFlightHeight)
		} else {
			c.specialDelay = c.world.rand.FloatRange(3, 10)
		}
	}
}

func (c *creepNode) updateWispLair(delta float64) {
	// It regenerates 1 health over 5 seconds (*0.2).
	// 12 hp over minute.
	c.health = gmath.ClampMax(c.health+(delta*0.2), c.maxHealth)

	if c.attackDelay != 0 {
		return
	}
	const maxUnits = 15
	if c.specialModifier > maxUnits {
		return
	}

	spawnPos := c.pos.Sub(gmath.Vec{Y: 20})
	c.attackDelay = c.scene.Rand().FloatRange(25, 45)
	c.specialModifier++

	wisp := c.world.NewCreepNode(spawnPos, gamedata.WispCreepStats)
	c.world.nodeRunner.AddObject(wisp)
	wisp.EventDestroyed.Connect(c, func(arg *creepNode) {
		c.specialModifier--
	})
	playSound(c.world, assets.AudioOrganicRestored, spawnPos)
	createEffect(c.world, effectConfig{
		Pos:     spawnPos,
		Image:   assets.ImageWispExplosion,
		Layer:   aboveEffectLayer,
		Reverse: true,
	})
}

func (c *creepNode) updateCreepCrawlerBase(delta float64) {
	if c.attackDelay != 0 {
		return
	}
	maxUnits := float64(10)
	productionDelay := 1.0
	if c.super {
		maxUnits = 15
		productionDelay = 0.75
	}
	if c.specialModifier > maxUnits {
		return
	}

	spawnPos := c.pos.Add(gmath.Vec{Y: 16})
	dstOffset := gmath.Vec{
		X: c.scene.Rand().FloatRange(-160, 160),
		Y: c.scene.Rand().FloatRange(-160, 160),
	}
	dstPos := spawnPos.Add(dstOffset)
	c.attackDelay = c.scene.Rand().FloatRange(15, 30) * productionDelay
	c.specialModifier++

	crawler := c.world.NewCreepNode(spawnPos, gamedata.CrawlerCreepStats)
	crawler.super = c.super && c.scene.Rand().Chance(0.2)
	crawler.SendTo(dstPos)
	crawler.waypoint = crawler.pos
	c.world.nodeRunner.AddObject(crawler)
	crawler.EventDestroyed.Connect(c, func(arg *creepNode) {
		c.specialModifier--
	})

	createEffect(c.world, effectConfig{
		Pos:   spawnPos,
		Layer: slightlyAboveEffectLayer,
		Image: assets.ImageCreepCreatedEffect,
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

		stats := gamedata.WandererCreepStats

		if tier2chance != 0 && c.scene.Rand().Chance(tier2chance) {
			stats = gamedata.StunnerCreepStats
		}

		creep := c.world.NewCreepNode(spawnPos, stats)
		creep.super = c.super && c.scene.Rand().Chance(0.4)
		creep.waypoint = waypoint
		creep.spawnedFromBase = true
		c.world.nodeRunner.AddObject(creep)
		creep.SetHeight(0)

		createEffect(c.world, effectConfig{
			Pos:   spawnPos,
			Layer: aboveEffectLayer,
			Image: assets.ImageCreepCreatedEffect,
		})
	}
}

func (c *creepNode) changePos(pos gmath.Vec) {
	c.pos = pos
	c.shadowComponent.UpdatePos(pos)
}

func (c *creepNode) SetHeight(h float64) {
	c.shadowComponent.UpdateHeight(c.pos, h, agentFlightHeight)
}

func (c *creepNode) updateServant(delta float64) {
	c.anim.Tick(delta)

	if c.specialTarget == nil && len(c.world.allColonies) == 0 {
		return
	}

	target, ok := c.specialTarget.(*colonyCoreNode)
	if !ok || target == nil || target.IsDisposed() {
		if len(c.world.allColonies) == 0 {
			c.specialTarget = nil
			return
		}
		c.specialTarget = gmath.RandElem(c.world.rand, c.world.allColonies)
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
	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 && c.specialModifier == 0 {
		if c.maybeSpawnCrawlers() {
			// Time until the first crawler is spawned.
			c.specialDelay = c.scene.Rand().FloatRange(7, 10)
		} else {
			c.specialDelay = c.scene.Rand().FloatRange(3, 7)
		}
	}

	// It regenerates 1 health over 4 seconds (*0.25).
	// Meaning it's 15 health per minute.
	// In other words, 10 minutes recover 150 health for this guy.
	c.health = gmath.ClampMax(c.health+(delta*0.25), c.maxHealth)
	c.updateHealthShader()

	const crawlersSpawnHeight float64 = 10
	if c.specialModifier != 0 && c.shadowComponent.height != crawlersSpawnHeight {
		height := c.shadowComponent.height - delta*5
		c.pos.Y += delta * 5
		if height <= crawlersSpawnHeight {
			height = crawlersSpawnHeight
			c.sprite.FrameOffset.X = c.sprite.FrameWidth
			c.altSprite.Visible = true
		}
		c.shadowComponent.UpdateHeight(c.pos, height, agentFlightHeight)
		return
	}
	if c.specialModifier != 0 && c.shadowComponent.height == crawlersSpawnHeight {
		if c.specialDelay == 0 {
			// Time until next crawler is spawned.
			c.specialDelay = c.scene.Rand().FloatRange(4, 6)
			c.specialModifier--
			if c.specialModifier > 0 {
				spawnPos := c.crawlerSpawnPos()
				var crawlerStats *gamedata.CreepStats
				if c.world.seedKind == gamedata.SeedLeet {
					crawlerStats = gamedata.StealthCrawlerCreepStats
				} else {
					crawlerStats = gamedata.CrawlerCreepStats
					eliteChance := c.world.EliteCrawlerChance()
					if c.world.rand.Chance(eliteChance) {
						if c.world.rand.Chance(0.3) {
							crawlerStats = gamedata.HeavyCrawlerCreepStats
						} else {
							crawlerStats = gamedata.EliteCrawlerCreepStats
						}
					}
				}
				crawler := c.world.NewCreepNode(spawnPos, crawlerStats)
				crawler.super = c.super && c.world.rand.Chance(0.5)
				crawler.path = crawlerSpawnPositions[int(c.specialModifier-1)]
				crawler.waypoint = crawler.pos
				c.world.nodeRunner.AddObject(crawler)
			}
		}
		if c.specialModifier == 0 {
			if c.world.config.GameMode == gamedata.ModeReverse {
				c.specialDelay = 60 * 60 * 60 // ~never
			} else {
				c.specialDelay = c.scene.Rand().FloatRange(50, 90)
			}
			c.sprite.FrameOffset.X = 0
			c.altSprite.Visible = false
			return
		}
		return
	}

	if c.shadowComponent.height != agentFlightHeight {
		height := c.shadowComponent.height + delta*5
		c.pos.Y -= delta * 5
		if height >= agentFlightHeight {
			height = agentFlightHeight
		}
		c.shadowComponent.UpdateHeight(c.pos, height, agentFlightHeight)
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

func (c *creepNode) CanBeTargeted() bool {
	return !c.cloaking && !c.insideForest
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
	createEffect(c.world, effectConfig{
		Pos:   c.pos,
		Layer: slightlyAboveEffectLayer,
		Image: assets.ImageCloakWave,
	})
}

func (c *creepNode) movementSpeed() float64 {
	if c.spawnedFromBase {
		return c.stats.Speed * 0.5
	}
	multiplier := 1.0
	if c.slow > 0 {
		multiplier = 0.55
	}
	return c.stats.Speed * multiplier
}

func (c *creepNode) moveTowards(delta float64, pos gmath.Vec) bool {
	pos, reached := moveTowardsWithSpeed(c.pos, pos, delta, c.movementSpeed())
	c.changePos(pos)
	return reached
}
