package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/pathing"
)

type creepKind int

const (
	creepPrimitiveWanderer creepKind = iota
	creepStunner
	creepAssault
	creepTurret
	creepBase
	creepTank
	creepCrawler
	creepUberBoss
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

	spawnPos        gmath.Vec
	pos             gmath.Vec
	waypoint        gmath.Vec
	wasAttacking    bool
	wasRetreating   bool
	spawnedFromBase bool

	path            pathing.GridPath
	specialDelay    float64
	specialModifier float64

	slow   float64
	health float64
	height float64

	attackDelay float64

	EventDestroyed gsignal.Event[*creepNode]
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

	c.health = c.stats.maxHealth

	c.sprite = scene.NewSprite(c.stats.image)
	c.sprite.Pos.Base = &c.pos
	if c.stats.shadowImage != assets.ImageNone {
		c.world.camera.AddGraphicsAbove(c.sprite)
	} else {
		c.world.camera.AddGraphics(c.sprite)
	}
	if c.stats.kind == creepTank {
		c.sprite.FlipHorizontal = scene.Rand().Bool()
	}
	if c.stats.animSpeed != 0 {
		c.anim = ge.NewRepeatedAnimation(c.sprite, -1)
		c.anim.Tick(scene.Rand().FloatRange(0, 0.7))
		c.anim.SetSecondsPerFrame(c.stats.animSpeed)
	}
	c.flashComponent.sprite = c.sprite

	c.height = agentFlightHeight

	if c.stats.shadowImage != assets.ImageNone && !c.world.isMobile {
		c.shadow = scene.NewSprite(c.stats.shadowImage)
		c.shadow.Pos.Base = &c.pos
		c.world.camera.AddGraphics(c.shadow)
		c.shadow.Pos.Offset.Y = c.height
		c.shadow.SetAlpha(0.5)
		if c.spawnedFromBase {
			c.shadow.Visible = false
		}
	}

	if c.stats.kind == creepUberBoss {
		c.altSprite = scene.NewSprite(assets.ImageUberBossOpen)
		c.altSprite.Visible = false
		c.altSprite.Pos.Base = &c.pos
		c.world.camera.AddGraphics(c.altSprite)
	}
}

func (c *creepNode) Dispose() {
	c.sprite.Dispose()
	if c.shadow != nil {
		c.shadow.Dispose()
	}
}

func (c *creepNode) Destroy() {
	c.EventDestroyed.Emit(c)
	c.Dispose()
}

func (c *creepNode) IsDisposed() bool { return c.sprite.IsDisposed() }

func (c *creepNode) Update(delta float64) {
	c.flashComponent.Update(delta)

	c.slow = gmath.ClampMin(c.slow-delta, 0)
	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	if c.attackDelay == 0 && c.stats.weapon != nil {
		c.attackDelay = c.stats.weapon.Reload * c.scene.Rand().FloatRange(0.8, 1.2)
		targets := c.findTargets()
		if len(targets) != 0 {
			for _, target := range targets {
				c.doAttack(target)
			}
			playSound(c.scene, c.world.camera, c.stats.weapon.AttackSound, c.pos)
		}
	}

	switch c.stats.kind {
	case creepPrimitiveWanderer, creepStunner, creepAssault:
		c.updatePrimitiveWanderer(delta)
	case creepUberBoss:
		c.updateUberBoss(delta)
	case creepBase:
		c.updateCreepBase(delta)
	case creepCrawler:
		c.updateCrawler(delta)
	case creepTank:
		c.updateTank(delta)
	case creepTurret:
		// Do nothing.
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

func (c *creepNode) IsFlying() bool {
	return c.shadow != nil
}

func (c *creepNode) TargetKind() targetKind {
	if c.IsFlying() {
		return targetFlying
	}
	return targetGround
}

func (c *creepNode) explode() {
	switch c.stats.kind {
	case creepUberBoss:
		fall := newDroneFallNode(c.world, nil, c.stats.image, c.shadow.ImageID(), c.pos, c.height)
		c.scene.AddObject(fall)
	case creepTurret, creepBase:
		createAreaExplosion(c.scene, c.world.camera, spriteRect(c.pos, c.sprite), true)
		scraps := c.world.NewEssenceSourceNode(bigScrapCreepSource, c.pos.Add(gmath.Vec{Y: 7}))
		c.scene.AddObject(scraps)
	case creepTank:
		createExplosion(c.scene, c.world.camera, false, c.pos)
		scraps := c.world.NewEssenceSourceNode(smallScrapCreepSource, c.pos.Add(gmath.Vec{Y: 2}))
		c.scene.AddObject(scraps)
	case creepCrawler:
		createExplosion(c.scene, c.world.camera, false, c.pos)
		if c.world.rand.Chance(0.3) {
			scraps := c.world.NewEssenceSourceNode(smallScrapCreepSource, c.pos.Add(gmath.Vec{Y: 2}))
			c.scene.AddObject(scraps)
		}
	default:
		roll := c.scene.Rand().Float()
		if roll < 0.3 {
			createExplosion(c.scene, c.world.camera, true, c.pos)
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
			fall := newDroneFallNode(c.world, scraps, c.stats.image, c.shadow.ImageID(), c.pos, agentFlightHeight)
			c.scene.AddObject(fall)
		}
	}
}

func (c *creepNode) OnDamage(damage damageValue, source gmath.Vec) {
	if damage.health != 0 {
		c.flashComponent.flash = 0.2
	}

	c.health -= damage.health
	if c.health < 0 {
		c.explode()
		c.Destroy()
		return
	}

	c.slow = gmath.ClampMax(c.slow+damage.slow, 5)

	if c.stats.kind == creepCrawler {
		if c.specialModifier == crawlerIdle && (c.pos.DistanceTo(source) > c.stats.weapon.AttackRange*0.8) && c.world.rand.Chance(0.45) {
			followPos := c.pos.MoveTowards(source, 64*c.world.rand.FloatRange(0.8, 1.4))
			p := c.world.BuildPath(c.pos, followPos)
			c.specialModifier = crawlerMove
			c.path = p.Steps
			c.waypoint = c.world.pathgrid.AlignPos(c.pos)
		}
		return
	}

	if damage.morale != 0 {
		if c.wasRetreating {
			return
		}
		if c.scene.Rand().Chance(damage.morale * 0.15) {
			c.wasAttacking = true
			c.retreatFrom(source)
		}
	}
}

func (c *creepNode) doAttack(target projectileTarget) {
	if c.stats.weapon.ProjectileImage != assets.ImageNone {
		toPos := snipePos(c.stats.weapon.ProjectileSpeed, c.pos, *target.GetPos(), target.GetVelocity())
		for i := 0; i < c.stats.weapon.BurstSize; i++ {
			fireDelay := float64(i) * c.stats.weapon.BurstDelay
			p := newProjectileNode(projectileConfig{
				Camera:    c.world.camera,
				Weapon:    c.stats.weapon,
				FromPos:   &c.pos,
				ToPos:     toPos.Add(c.scene.Rand().Offset(-4, 4)),
				Target:    target,
				FireDelay: fireDelay,
			})
			c.scene.AddObject(p)
		}
		return
	}

	beam := newBeamNode(c.world.camera, ge.Pos{Base: &c.pos}, ge.Pos{Base: target.GetPos()}, c.stats.beamColor)
	beam.width = c.stats.beamWidth
	c.scene.AddObject(beam)
	target.OnDamage(c.stats.weapon.Damage, c.pos)
}

func (c *creepNode) retreatFrom(pos gmath.Vec) {
	direction := pos.AngleToPoint(c.pos) + gmath.Rad(c.scene.Rand().FloatRange(-0.2, 0.2))
	dist := c.scene.Rand().FloatRange(300, 500)
	c.setWaypoint(pos.MoveInDirection(dist, direction))
	c.wasAttacking = false
}

func (c *creepNode) findTargets() []projectileTarget {
	targets := c.world.tmpTargetSlice[:0]
	c.world.FindColonyAgent(c.pos, c.stats.weapon.AttackRange, func(a *colonyAgentNode) bool {
		targets = append(targets, a)
		return len(targets) >= c.stats.weapon.MaxTargets
	})
	if c.stats.weapon.Damage.health == 0 {
		return targets
	}

	if len(targets) >= c.stats.weapon.MaxTargets {
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

func (c *creepNode) updatePrimitiveWanderer(delta float64) {
	if c.anim != nil {
		c.anim.Tick(delta)
	}

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
		c.shadow.Visible = true
		c.height = agentFlightHeight
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
	switch c.world.options.Difficulty {
	case 2:
		maxCrawlers = 5
	case 3:
		minCrawlers = 3
		maxCrawlers = 5
	}

	c.specialModifier = float64(c.scene.Rand().IntRange(minCrawlers, maxCrawlers)) + 1
	return true
}

func (c *creepNode) updateCrawler(delta float64) {
	if !c.waypoint.IsZero() {
		c.anim.Tick(delta)
		if c.moveTowards(delta, c.waypoint) {
			// To avoid weird cases of walking above colony core or turret,
			// stop if there are any targets in vicinity.
			const stopDistSqr float64 = 96 * 96
			for _, colony := range c.world.colonies {
				if colony.pos.DistanceSquaredTo(c.pos) < stopDistSqr {
					c.path = pathing.GridPath{}
					break
				}
				for _, turret := range colony.turrets {
					if turret.pos.DistanceSquaredTo(c.pos) < stopDistSqr {
						c.path = pathing.GridPath{}
						break
					}
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

func (c *creepNode) updateTank(delta float64) {
	if c.waypoint.IsZero() {
		if c.pos != c.spawnPos {
			c.waypoint = c.spawnPos
			c.specialDelay = c.scene.Rand().FloatRange(1, 4)
		} else {
			offset := gmath.Vec{
				X: c.scene.Rand().FloatRange(-40, 40),
				Y: c.scene.Rand().FloatRange(-4, 4),
			}
			c.waypoint = c.spawnPos.Add(offset)
		}
	}

	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)

	if c.moveTowards(delta, c.waypoint) {
		c.waypoint = gmath.Vec{}
	}
}

func (c *creepNode) updateCreepBase(delta float64) {
	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 && c.specialModifier < 15 {
		c.specialDelay = c.scene.Rand().FloatRange(55, 90)
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
		c.scene.AddObject(creep)
		creep.height = 0
	}
}

func (c *creepNode) updateUberBoss(delta float64) {
	c.anim.Tick(delta)

	if c.world.IsTutorial() {
		return
	}

	c.shadow.Pos.Offset.Y = c.height + 4
	newShadowAlpha := float32(1.0 - ((c.height / agentFlightHeight) * 0.5))
	c.shadow.SetAlpha(newShadowAlpha)

	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 && c.specialModifier == 0 {
		if c.maybeSpawnCrawlers() {
			// Time until the first crawler is spawned.
			c.specialDelay = c.scene.Rand().FloatRange(7, 10)
		} else {
			c.specialDelay = c.scene.Rand().FloatRange(6, 16)
		}
	}

	// It regenerates 1 health over 4 seconds (*0.25).
	// Meaning it's 15 health per minute.
	// In other words, 10 minutes recover 150 health for this guy.
	c.health = gmath.ClampMax(c.health+(delta*0.25), c.stats.maxHealth)

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
					crawlerStats = eliteCrawlerCreepStats
				}
				crawler := c.world.NewCreepNode(spawnPos, crawlerStats)
				crawler.path = crawlerSpawnPositions[int(c.specialModifier-1)]
				crawler.waypoint = crawler.pos
				c.scene.AddObject(crawler)
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

func (c *creepNode) movementSpeed() float64 {
	if c.stats.kind == creepTank && c.specialDelay != 0 {
		return 0
	}
	if c.spawnedFromBase && c.height == 0 {
		return c.stats.speed * 0.5
	}
	multiplier := 1.0
	if c.slow > 0 {
		multiplier = 0.6
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
