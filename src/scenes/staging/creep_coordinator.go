package staging

import (
	"github.com/quasilyte/gmath"
)

const (
	// Idle is a state of doing nothing.
	// When over, a next state can be selected.
	crawlerIdle = iota
	// Move is a state of running towards a target.
	crawlerMove
)

type creepCoordinator struct {
	world *worldState

	crawlers   []*creepNode
	groupSlice []*creepNode

	scoutingDelay float64
	attackDelay   float64
	scatterDelay  float64
}

func newCreepCoordinator(world *worldState) *creepCoordinator {
	return &creepCoordinator{
		world:        world,
		crawlers:     make([]*creepNode, 0, 16),
		groupSlice:   make([]*creepNode, 0, 4),
		attackDelay:  world.rand.FloatRange(10, 30),
		scatterDelay: world.rand.FloatRange(2*60, 3*60),
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

	if c.attackDelay == 0 {
		c.tryLaunchingAttack()
	}
	if c.scoutingDelay == 0 {
		c.sendScout()
	}
	if c.scatterDelay == 0 {
		c.tryLaunchingScatter()
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

	c.scoutingDelay = c.world.rand.FloatRange(30.0, 50.0)

	scoutingDist := 256 * c.world.rand.FloatRange(1, 2)
	scoutingDest := gmath.RadToVec(c.world.rand.Rad()).Mulf(scoutingDist).Add(scout.pos)
	scout.specialModifier = crawlerMove
	scout.waypoint = c.world.pathgrid.AlignPos(scout.pos)
	p := c.world.BuildPath(scout.waypoint, scoutingDest)
	scout.path = p.Steps
}

func (c *creepCoordinator) tryLaunchingScatter() {
	leader := gmath.RandElem(c.world.rand, c.crawlers)
	if leader.specialModifier != crawlerIdle {
		c.scatterDelay = c.world.rand.FloatRange(4, 10)
		return
	}

	group := c.collectGroup(leader)
	if len(group) < 2 {
		c.scatterDelay = c.world.rand.FloatRange(8, 14)
		return
	}

	c.scatterDelay = c.world.rand.FloatRange(70, 90)

	for _, creep := range group {
		dist := c.world.rand.FloatRange(96, 192)
		targetPos := gmath.RadToVec(c.world.rand.Rad()).Mulf(dist).Add(creep.pos)

		creep.specialModifier = crawlerMove
		p := c.world.BuildPath(creep.pos, targetPos)
		creep.path = p.Steps
		creep.waypoint = c.world.pathgrid.AlignPos(creep.pos)
	}
}

func (c *creepCoordinator) tryLaunchingAttack() {
	// Pick a random unit to start forming a group.
	leader := gmath.RandElem(c.world.rand, c.crawlers)
	if leader.specialModifier != crawlerIdle {
		// Bad leader pick attempt, try later.
		c.attackDelay = c.world.rand.FloatRange(1.2, 2.6)
		return
	}

	const (
		maxAttackRange    float64 = 1024.0
		maxAttackRangeSqr float64 = maxAttackRange * maxAttackRange
	)

	group := c.collectGroup(leader)

	attackRangeSqr := maxAttackRangeSqr * c.world.rand.FloatRange(0.8, 1.2)

	// Now try to find a suitable target.
	var target *colonyCoreNode
	for _, colony := range c.world.colonies {
		if colony.pos.DistanceSquaredTo(leader.pos) > attackRangeSqr {
			continue
		}
		target = colony
		break
	}

	if target == nil {
		// No reachable targets for this group, try later.
		c.attackDelay = c.world.rand.FloatRange(4.5, 6.5)
		return
	}

	// Launch the attack.

	// The next action will be much later.
	c.attackDelay = c.world.rand.FloatRange(30.0, 70.0)

	for _, creep := range group {
		dist := c.world.rand.FloatRange(creep.stats.weapon.AttackRange*0.5, creep.stats.weapon.AttackRange*0.8)
		targetPos := gmath.RadToVec(c.world.rand.Rad()).Mulf(dist).Add(target.pos)

		creep.specialModifier = crawlerMove
		p := c.world.BuildPath(creep.pos, targetPos)
		creep.path = p.Steps
		creep.waypoint = c.world.pathgrid.AlignPos(creep.pos)
	}
}

func (c *creepCoordinator) collectGroup(leader *creepNode) []*creepNode {
	const (
		maxUnitRange    float64 = 196
		maxUnitRangeSqr float64 = maxUnitRange * maxUnitRange
	)

	// Try to build a group of at least 2 units.
	maxGroupSize := c.world.rand.IntRange(2, cap(c.groupSlice))
	group := c.groupSlice[:0]
	group = append(group, leader)
	for _, creep := range c.crawlers {
		if len(group) >= maxGroupSize {
			break
		}
		if creep == leader || creep.specialModifier != crawlerIdle {
			continue
		}
		if creep.pos.DistanceSquaredTo(leader.pos) > maxUnitRangeSqr {
			continue
		}
		group = append(group, creep)
	}

	return group
}
