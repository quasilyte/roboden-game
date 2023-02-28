package staging

import (
	"fmt"

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
}

func newCreepCoordinator(world *worldState) *creepCoordinator {
	return &creepCoordinator{
		world:       world,
		crawlers:    make([]*creepNode, 0, 16),
		groupSlice:  make([]*creepNode, 0, 4),
		attackDelay: world.rand.FloatRange(10, 30),
	}
}

func (c *creepCoordinator) Update(delta float64) {
	c.attackDelay = gmath.ClampMin(c.attackDelay-delta, 0)
	c.scoutingDelay = gmath.ClampMin(c.scoutingDelay-delta, 0)

	if len(c.crawlers) == 0 {
		// No units to coordinate, try later.
		c.attackDelay = c.world.rand.FloatRange(6.0, 10.0)
		c.scoutingDelay += 2 + delta*8
		return
	}

	if c.attackDelay == 0 {
		c.tryLaunchingAttack()
	}
	if c.scoutingDelay == 0 {
		c.sendScout()
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

	fmt.Println("launch creep scout")
	c.scoutingDelay = c.world.rand.FloatRange(10.0, 30.0)

	scoutingDist := 256 * c.world.rand.FloatRange(1, 2)
	scoutingDest := gmath.RadToVec(c.world.rand.Rad()).Mulf(scoutingDist).Add(scout.pos)
	scout.specialModifier = crawlerMove
	scout.waypoint = c.world.pathgrid.AlignPos(scout.pos)
	p := c.world.BuildPath(scout.waypoint, scoutingDest)
	scout.path = p.Steps
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
		maxUnitRange      float64 = 196
		maxUnitRangeSqr   float64 = maxUnitRange * maxUnitRange
		maxAttackRange    float64 = 1024.0
		maxAttackRangeSqr float64 = maxAttackRange * maxAttackRange
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
	fmt.Println("launch creep attack")
	c.attackDelay = c.world.rand.FloatRange(20.0, 55.0)

	for _, creep := range group {
		dist := c.world.rand.FloatRange(creep.stats.weapon.AttackRange*0.5, creep.stats.weapon.AttackRange*0.8)
		targetPos := gmath.RadToVec(c.world.rand.Rad()).Mulf(dist).Add(target.pos)

		creep.specialModifier = crawlerMove
		p := c.world.BuildPath(creep.pos, targetPos)
		creep.path = p.Steps
		creep.waypoint = c.world.pathgrid.AlignPos(creep.pos)
	}
}
