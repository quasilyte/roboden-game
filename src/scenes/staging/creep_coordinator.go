package staging

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

const (
	// Idle is a state of doing nothing.
	// When over, a next state can be selected.
	crawlerIdle = iota
	// Move is a state of running towards a target.
	crawlerMove
	crawlerGuard
)

type creepCoordinator struct {
	world *worldState

	crawlers   []*creepNode
	groupSlice []*creepNode

	scoutingDelay float64
	attackDelay   float64
	scatterDelay  float64
	relocateDelay float64
}

func newCreepCoordinator(world *worldState) *creepCoordinator {
	return &creepCoordinator{
		world:         world,
		crawlers:      make([]*creepNode, 0, 16),
		groupSlice:    make([]*creepNode, 0, 40),
		attackDelay:   world.rand.FloatRange(10, 30),
		scatterDelay:  world.rand.FloatRange(2*60, 3*60),
		relocateDelay: world.rand.FloatRange(1*60, 3*60),
	}
}

func (c *creepCoordinator) Update(delta float64) {
	if len(c.crawlers) == 0 {
		// No units to coordinate, try later.
		return
	}

	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	c.scoutingDelay = gmath.ClampMin(c.scoutingDelay-delta, 0)
	c.scatterDelay = gmath.ClampMin(c.scatterDelay-delta, 0)
	c.relocateDelay = gmath.ClampMin(c.relocateDelay-delta, 0)

	if c.attackDelay == 0 {
		c.tryLaunchingAttack()
	}
	if c.scoutingDelay == 0 {
		c.sendScout()
	}
	if c.scatterDelay == 0 {
		c.tryLaunchingScatter()
	}
	if c.relocateDelay == 0 {
		c.tryLaunchingRelocation()
	}
}

func (c *creepCoordinator) sendScout() {
	scout := gmath.RandElem(c.world.rand, c.crawlers)
	if scout.specialModifier != crawlerIdle {
		c.scoutingDelay = c.world.rand.FloatRange(1.6, 3)
		return
	}

	if c.world.rand.Chance(0.35) {
		c.scoutingDelay = c.world.rand.FloatRange(4, 8.5)
		return
	}

	if c.world.config.GameMode == gamedata.ModeArena {
		c.scoutingDelay = c.world.rand.FloatRange(20.0, 30.0)
	} else {
		c.scoutingDelay = c.world.rand.FloatRange(30.0, 50.0)
	}

	scoutingDist := 320 * c.world.rand.FloatRange(1, 2)
	scoutingDest := gmath.RadToVec(c.world.rand.Rad()).Mulf(scoutingDist).Add(scout.pos)
	scout.specialModifier = crawlerMove
	scout.waypoint = c.world.pathgrid.AlignPos(scout.pos)
	p := c.world.BuildPath(scout.waypoint, scoutingDest)
	scout.path = p.Steps
}

func (c *creepCoordinator) tryLaunchingRelocation() {
	leader := gmath.RandElem(c.world.rand, c.crawlers)
	if leader.specialModifier != crawlerIdle {
		c.relocateDelay = c.world.rand.FloatRange(3, 8)
		return
	}

	group := c.collectGroup(leader.pos, 300, 20)
	if len(group) < 2 {
		c.relocateDelay = c.world.rand.FloatRange(4, 10)
		return
	}

	if c.world.config.GameMode == gamedata.ModeArena {
		c.relocateDelay = c.world.rand.FloatRange(25, 55)
	} else {
		c.relocateDelay = c.world.rand.FloatRange(60, 90)
	}

	targetPos := correctedPos(c.world.rect, randomSectorPos(c.world.rand, c.world.rect), 480)
	for _, creep := range group {
		creepTargetPos := correctedPos(c.world.rect, targetPos.Add(c.world.rand.Offset(-96, 96)), 32)

		creep.specialModifier = crawlerMove
		p := c.world.BuildPath(creep.pos, creepTargetPos)
		creep.path = p.Steps
		creep.waypoint = c.world.pathgrid.AlignPos(creep.pos)
	}
}

func (c *creepCoordinator) tryLaunchingScatter() {
	leader := gmath.RandElem(c.world.rand, c.crawlers)
	if leader.specialModifier != crawlerIdle {
		c.scatterDelay = c.world.rand.FloatRange(4, 10)
		return
	}

	group := c.collectGroup(leader.pos, 300, 10)
	if len(group) < 2 {
		c.scatterDelay = c.world.rand.FloatRange(8, 14)
		return
	}

	if c.world.config.GameMode == gamedata.ModeArena {
		c.scatterDelay = c.world.rand.FloatRange(55, 85)
	} else {
		c.scatterDelay = c.world.rand.FloatRange(70, 90)
	}

	c.scatterCreeps(group)
}

func (c *creepCoordinator) Rally(pos gmath.Vec) {
	group := c.collectGroup(pos, 400, cap(c.groupSlice))

	maxAttackDistSqr := 1500.0 * 1500.0
	var closestTarget *colonyCoreNode
	var closestDistSqr float64
	for _, colony := range c.world.allColonies {
		distSqr := colony.pos.DistanceSquaredTo(pos)
		if distSqr > maxAttackDistSqr {
			continue
		}
		if closestTarget == nil || distSqr < closestDistSqr {
			closestDistSqr = distSqr
			closestTarget = colony
		}
	}
	if closestTarget != nil {
		c.sendCreepsToAttack(group, closestTarget)
		return
	}
	c.scatterCreeps(group)
}

func (c *creepCoordinator) findColonyToAttack(pos gmath.Vec, r float64) *colonyCoreNode {
	if len(c.world.allColonies) == 0 {
		return nil
	}
	rSqr := r * r
	return randIterate(c.world.rand, c.world.allColonies, func(c *colonyCoreNode) bool {
		return c.pos.DistanceSquaredTo(pos) <= rSqr
	})
}

func (c *creepCoordinator) tryLaunchingAttack() {
	// Pick a random unit to start forming a group.
	leader := gmath.RandElem(c.world.rand, c.crawlers)
	if leader.specialModifier != crawlerIdle {
		// Bad leader pick attempt, try later.
		c.attackDelay = c.world.rand.FloatRange(1.2, 2.6)
		return
	}

	group := c.collectGroup(leader.pos, 300, cap(c.groupSlice))

	maxAttackRange := 1024.0 * c.world.rand.FloatRange(0.8, 1.2)
	if c.world.config.GameMode == gamedata.ModeArena {
		maxAttackRange *= 1.5
	}

	// Now try to find a suitable target.
	target := c.findColonyToAttack(leader.pos, maxAttackRange)
	if target == nil {
		// No reachable targets for this group, try later.
		c.attackDelay = c.world.rand.FloatRange(4.5, 6.5)
		return
	}

	// Launch the attack.

	if c.world.config.GameMode == gamedata.ModeArena {
		c.attackDelay = c.world.rand.FloatRange(25.0, 55.0)
	} else {
		// The next action will be much later.
		c.attackDelay = c.world.rand.FloatRange(30.0, 70.0)
	}

	c.sendCreepsToAttack(group, target)
}

func (c *creepCoordinator) scatterCreeps(group []*creepNode) {
	for _, creep := range group {
		dist := c.world.rand.FloatRange(128, 400)
		targetPos := gmath.RadToVec(c.world.rand.Rad()).Mulf(dist).Add(creep.pos)
		creep.SendTo(targetPos)
	}
}

func (c *creepCoordinator) sendCreepsToAttack(group []*creepNode, target *colonyCoreNode) {
	for _, creep := range group {
		dist := c.world.rand.FloatRange(creep.stats.Weapon.AttackRange*0.5, creep.stats.Weapon.AttackRange*0.8)
		dir := gmath.RadToVec(c.world.rand.Rad()).Mulf(dist)
		targetPos := dir.Add(target.pos)
		if c.world.HasTreesAt(targetPos, 0) {
			// Try to find a better spot.
			targetPos = dir.Mulf(-1).Add(target.pos)
		}
		creep.SendTo(targetPos)
	}
}

func (c *creepCoordinator) collectGroup(pos gmath.Vec, r float64, maxGroupSize int) []*creepNode {
	rSqr := r * r

	if maxGroupSize > cap(c.groupSlice) {
		maxGroupSize = cap(c.groupSlice)
	}
	// Try to build a group of at least 2 units.
	groupSize := c.world.rand.IntRange(2, maxGroupSize)
	group := c.groupSlice[:0]
	for _, creep := range c.crawlers {
		if len(group) >= groupSize {
			break
		}
		if creep.specialModifier != crawlerIdle {
			continue
		}
		if creep.pos.DistanceSquaredTo(pos) > rSqr {
			continue
		}
		group = append(group, creep)
	}

	return group
}
