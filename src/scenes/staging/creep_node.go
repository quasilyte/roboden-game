package staging

import (
	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
)

type creepKind int

const (
	creepPrimitiveWanderer creepKind = iota
	creepStunner
	creepAssault
	creepTurret
	creepBase
	creepUberBoss
)

type creepNode struct {
	sprite *ge.Sprite
	shadow *ge.Sprite

	scene *ge.Scene

	world *worldState
	stats *creepStats

	pos             gmath.Vec
	waypoint        gmath.Vec
	wasAttacking    bool
	wasRetreating   bool
	spawnedFromBase bool

	specialDelay    float64
	specialModifier float64
	specialTarget   any

	slow   float64
	health float64
	height float64

	attackDelay float64

	EventDestroyed gsignal.Event[*creepNode]
}

func newCreepNode(world *worldState, stats *creepStats, pos gmath.Vec) *creepNode {
	return &creepNode{
		world: world,
		stats: stats,
		pos:   pos,
	}
}

func (c *creepNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.health = c.stats.maxHealth

	c.sprite = scene.NewSprite(c.stats.image)
	c.sprite.Pos.Base = &c.pos
	c.world.camera.AddGraphicsAbove(c.sprite)

	c.height = agentFlightHeight

	if c.stats.shadowImage != assets.ImageNone {
		c.shadow = scene.NewSprite(c.stats.shadowImage)
		c.shadow.Pos.Base = &c.pos
		c.world.camera.AddGraphics(c.shadow)
		c.shadow.Pos.Offset.Y = c.height
		c.shadow.SetAlpha(0.5)
		if c.spawnedFromBase {
			c.shadow.Visible = false
		}
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
	c.slow = gmath.ClampMin(c.slow-delta, 0)
	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	if c.attackDelay == 0 && c.stats.weaponReload != 0 {
		c.attackDelay = c.stats.weaponReload * c.scene.Rand().FloatRange(0.8, 1.2)
		targets := c.findTargets()
		if len(targets) != 0 {
			for _, target := range targets {
				c.doAttack(target)
			}
			playSound(c.scene, c.world.camera, c.stats.attackSound, c.pos)
		}
	}

	switch c.stats.kind {
	case creepPrimitiveWanderer, creepStunner, creepAssault:
		c.updatePrimitiveWanderer(delta)
	case creepUberBoss:
		c.updateUberBoss(delta)
	case creepBase:
		c.updateCreepBase(delta)
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
	return c.stats.shadowImage != assets.ImageNone
}

func (c *creepNode) explode() {
	switch c.stats.kind {
	case creepUberBoss:
		// TODO: big explosion
	case creepTurret, creepBase:
		createAreaExplosion(c.scene, c.world.camera, spriteRect(c.pos, c.sprite))
		scraps := c.world.NewEssenceSourceNode(bigScrapSource, c.pos.Add(gmath.Vec{Y: 7}))
		c.scene.AddObject(scraps)
	default:
		roll := c.scene.Rand().Float()
		if roll < 0.3 {
			createExplosion(c.scene, c.world.camera, c.pos)
		} else {
			var scraps *essenceSourceStats
			if roll > 0.6 {
				scraps = smallScrapSource
			}
			fall := newDroneFallNode(c.world, scraps, c.stats.image, c.shadow.ImageID(), c.pos, agentFlightHeight)
			c.scene.AddObject(fall)
		}
	}
}

func (c *creepNode) OnDamage(damage damageValue, source gmath.Vec) {
	c.health -= damage.health
	if c.health < 0 {
		c.explode()
		c.Destroy()
		return
	}

	c.slow = gmath.ClampMax(c.slow+damage.slow, 5)

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
	if c.stats.projectileImage != assets.ImageNone {
		toPos := snipePos(c.stats.projectileSpeed, c.pos, *target.GetPos(), target.GetVelocity())
		toPos = toPos.Add(c.scene.Rand().Offset(-3, 3))
		p := newProjectileNode(projectileConfig{
			Camera:      c.world.camera,
			Image:       c.stats.projectileImage,
			FromPos:     c.pos.Add(c.stats.fireOffset),
			ToPos:       toPos,
			Target:      target,
			Area:        c.stats.projectileArea,
			Speed:       c.stats.projectileSpeed,
			RotateSpeed: c.stats.projectileRotateSpeed,
			Damage:      c.stats.projectileDamage,
			Explosion:   c.stats.projectileExplosion,
		})
		c.scene.AddObject(p)
		return
	}

	beam := newBeamNode(c.world.camera, ge.Pos{Base: &c.pos}, ge.Pos{Base: target.GetPos()}, c.stats.beamColor)
	beam.width = c.stats.beamWidth
	c.scene.AddObject(beam)
	target.OnDamage(c.stats.projectileDamage, c.pos)
}

func (c *creepNode) retreatFrom(pos gmath.Vec) {
	direction := pos.AngleToPoint(c.pos) + gmath.Rad(c.scene.Rand().FloatRange(-0.2, 0.2))
	dist := c.scene.Rand().FloatRange(300, 500)
	c.setWaypoint(pos.MoveInDirection(dist, direction))
	c.wasAttacking = false
}

func (c *creepNode) findTargets() []projectileTarget {
	targets := c.world.tmpTargetSlice[:0]
	c.world.FindColonyAgent(c.pos, c.stats.attackRange, func(a *colonyAgentNode) bool {
		targets = append(targets, a)
		return len(targets) >= c.stats.maxTargets
	})
	if c.stats.projectileDamage.health == 0 {
		return targets
	}

	if len(targets) >= c.stats.maxTargets {
		return targets
	}
	for _, colony := range c.world.coreConstructions {
		if len(targets) >= c.stats.maxTargets {
			return targets
		}
		if colony.pos.DistanceTo(c.pos) > c.stats.attackRange {
			continue
		}
		targets = append(targets, colony)
	}

	if len(targets) >= c.stats.maxTargets {
		return targets
	}
	for _, colony := range c.world.colonies {
		if len(targets) >= c.stats.maxTargets {
			return targets
		}
		if colony.pos.DistanceTo(c.pos) > c.stats.attackRange {
			continue
		}
		targets = append(targets, colony)
	}

	return targets
}

func (c *creepNode) updatePrimitiveWanderer(delta float64) {
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

func (c *creepNode) maybeSpawnWaste() bool {
	spawnPos := c.pos.Add(gmath.Vec{Y: agentFlightHeight})
	if !posIsFree(c.world, nil, spawnPos, 24) {
		return false
	}
	wastePool := c.world.NewEssenceSourceNode(wasteSource, spawnPos)
	c.scene.AddObject(wastePool)
	wastePool.percengage = 0
	wastePool.resource = 0
	wastePool.updateShader()
	c.specialTarget = wastePool
	c.specialModifier = c.scene.Rand().FloatRange(0.45, 0.95)
	return true
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
	c.shadow.Pos.Offset.Y = c.height + 4
	newShadowAlpha := float32(1.0 - ((c.height / agentFlightHeight) * 0.5))
	c.shadow.SetAlpha(newShadowAlpha)

	c.specialDelay = gmath.ClampMin(c.specialDelay-delta, 0)
	if c.specialDelay == 0 {
		if c.maybeSpawnWaste() {
			c.specialDelay = c.scene.Rand().FloatRange(48, 80)
		} else {
			c.specialDelay = c.scene.Rand().FloatRange(6, 16)
		}
	}

	const wasteSpillHeight float64 = 20
	if c.specialTarget != nil && c.height != wasteSpillHeight {
		c.height -= delta * 5
		c.pos.Y += delta * 5
		if c.height <= wasteSpillHeight {
			c.height = wasteSpillHeight
		}
		return
	}
	if c.specialTarget != nil && c.height == wasteSpillHeight {
		wastePool := c.specialTarget.(*essenceSourceNode)
		if wastePool.percengage >= c.specialModifier {
			c.specialTarget = nil
			return
		}
		wastePool.Add(delta * 4)
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
