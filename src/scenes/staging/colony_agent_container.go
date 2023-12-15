package staging

import (
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type agentSearchFlags int

const (
	// randomized available fighters workers
	//          0         0        0       0 | bad (empty search flags)
	//          0         0        0       1 | each worker
	//          0         0        1       0 | each fighter
	//          0         0        1       1 | each
	//          0         1        0       0 | bad (searching for no one)
	//          0         1        0       1 | each available worker
	//          0         1        1       0 | each available fighter
	//          0         1        1       1 | each available
	//          1         0        0       0 | bad (searching for no one)
	//          1         0        0       1 | random worker
	//          1         0        1       0 | random fighter
	//          1         0        1       1 | random
	//          1         1        0       0 | bad (searching for no one)
	//          1         1        0       1 | random available worker
	//          1         1        1       0 | random available fighter
	//          1         1        1       1 | random available

	searchWorkers agentSearchFlags = 1 << iota
	searchFighters
	searchOnlyAvailable
	searchRandomized
)

func (flags agentSearchFlags) Workers() bool       { return flags&searchWorkers != 0 }
func (flags agentSearchFlags) Fighters() bool      { return flags&searchFighters != 0 }
func (flags agentSearchFlags) OnlyAvailable() bool { return flags&searchOnlyAvailable != 0 }
func (flags agentSearchFlags) Randomized() bool    { return flags&searchRandomized != 0 }

func (flags agentSearchFlags) Validate() {
	if flags == 0 {
		panic("empty search flags")
	}
	if !flags.Workers() && !flags.Fighters() {
		panic("searching for no one")
	}
}

type colonyAgentContainer struct {
	rand *gmath.Rand

	workers  []*colonyAgentNode
	fighters []*colonyAgentNode // Contains universal drones as well

	universal []*colonyAgentNode // Contains only universal drones; needed only when querying just for workers

	availableWorkers   []*colonyAgentNode
	availableFighters  []*colonyAgentNode // Does not contain universal drones
	availableUniversal []*colonyAgentNode // Can both patrol and gather

	sortTmp [3][]*colonyAgentNode

	hasGatherer        bool
	hasRedMiner        bool
	hasCloner          bool
	hasCourier         bool
	servoNum           int
	tier2Num           int
	tier3Num           int
	tier1workerNum     int
	tier2plusWorkerNum int
}

func newColonyAgentContainer(rand *gmath.Rand) *colonyAgentContainer {
	return &colonyAgentContainer{
		rand:               rand,
		workers:            make([]*colonyAgentNode, 0, 48),
		fighters:           make([]*colonyAgentNode, 0, 32),
		universal:          make([]*colonyAgentNode, 0, 20),
		availableWorkers:   make([]*colonyAgentNode, 0, 48),
		availableFighters:  make([]*colonyAgentNode, 0, 32),
		availableUniversal: make([]*colonyAgentNode, 0, 16),
	}
}

func (c *colonyAgentContainer) Update() {
	c.hasRedMiner = false
	c.hasCloner = false
	c.hasGatherer = false
	c.servoNum = 0
	c.tier2Num = 0
	c.tier3Num = 0
	c.tier1workerNum = 0
	c.tier2plusWorkerNum = 0
	c.availableWorkers = c.availableWorkers[:0]
	c.availableFighters = c.availableFighters[:0]
	c.availableUniversal = c.availableUniversal[:0]

	for _, a := range c.workers {
		if a.stats.CanGather {
			c.hasGatherer = true
		}
		if a.stats.Kind == gamedata.AgentServo {
			c.servoNum++
		}
		switch a.stats.Tier {
		case 1:
			c.tier1workerNum++
		case 2:
			c.tier2Num++
			c.tier2plusWorkerNum++
		case 3:
			c.tier3Num++
			c.tier2plusWorkerNum++
		}
		if a.mode == agentModeStandby {
			c.availableWorkers = append(c.availableWorkers, a)
			switch a.stats.Kind {
			case gamedata.AgentRedminer:
				c.hasRedMiner = true
			case gamedata.AgentCloner:
				c.hasCloner = true
			case gamedata.AgentCourier, gamedata.AgentTrucker:
				c.hasCourier = true
			}
		}
	}

	for _, a := range c.fighters {
		switch a.stats.Tier {
		case 2:
			c.tier2Num++
		case 3:
			c.tier3Num++
		}
		if a.mode == agentModePatrol || a.mode == agentModeStandby {
			if a.stats.CanGather {
				c.availableUniversal = append(c.availableUniversal, a)
			} else {
				c.availableFighters = append(c.availableFighters, a)
			}
		}
	}
}

func (c *colonyAgentContainer) TotalNum() int {
	return len(c.workers) + len(c.fighters)
}

func (c *colonyAgentContainer) NumAvailableWorkers() int {
	return len(c.availableWorkers) + len(c.availableUniversal)
}

func (c *colonyAgentContainer) NumAvailableFighters() int {
	return len(c.availableFighters) + len(c.availableUniversal)
}

func (c *colonyAgentContainer) Add(a *colonyAgentNode) {
	if a.stats.CanPatrol {
		c.fighters = append(c.fighters, a)
		if a.stats.CanGather {
			c.universal = append(c.universal, a)
		}
	} else {
		c.workers = append(c.workers, a)
	}
}

func (c *colonyAgentContainer) Remove(a *colonyAgentNode) {
	if a.stats.CanPatrol {
		c.fighters = xslices.Remove(c.fighters, a)
		if a.stats.CanGather {
			c.universal = xslices.Remove(c.universal, a)
			c.availableUniversal = xslices.Remove(c.availableUniversal, a)
		} else {
			c.availableFighters = xslices.Remove(c.availableFighters, a)
		}
	} else {
		c.workers = xslices.Remove(c.workers, a)
		c.availableWorkers = xslices.Remove(c.availableWorkers, a)
	}
}

// Each is a loop over all colony agents.
// There is no way to handly only some of them.
// The agents are traversed in linear order,
// from workers to fighters.
func (c *colonyAgentContainer) Each(f func(a *colonyAgentNode)) {
	for _, a := range c.workers {
		f(a)
	}
	for _, a := range c.fighters {
		f(a)
	}
}

func (c *colonyAgentContainer) Find(flags agentSearchFlags, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	flags.Validate()
	switch flags {
	case searchWorkers:
		return c.findSlice2(c.workers, c.universal, f)
	case searchFighters:
		// c.fighters has both fighters and universal drones.
		return c.findSlice(c.fighters, f)
	case searchWorkers | searchFighters:
		// These two slices cover 100% of the drones.
		// It's like Each(), but allows an early break.
		return c.findSlice2(c.workers, c.fighters, f)
	case searchWorkers | searchOnlyAvailable:
		return c.findSlice2(c.availableWorkers, c.availableUniversal, f)
	case searchFighters | searchOnlyAvailable:
		// c.availableFighters does not include universal drones, hence the extra c.availableUniversal argument.
		return c.findSlice2(c.availableFighters, c.availableUniversal, f)
	case searchWorkers | searchFighters | searchOnlyAvailable:
		return c.findSlice3(c.availableWorkers, c.availableFighters, c.availableUniversal, f)
	case searchWorkers | searchRandomized:
		return c.randFindSlice2(c.workers, c.universal, f)
	case searchFighters | searchRandomized:
		return c.randFindSlice(c.fighters, f)
	case searchWorkers | searchFighters | searchRandomized:
		return c.randFindSlice2(c.workers, c.fighters, f)
	case searchWorkers | searchOnlyAvailable | searchRandomized:
		return c.randFindSlice2(c.availableWorkers, c.availableUniversal, f)
	case searchFighters | searchOnlyAvailable | searchRandomized:
		return c.randFindSlice2(c.availableFighters, c.availableUniversal, f)
	case searchWorkers | searchFighters | searchOnlyAvailable | searchRandomized:
		return c.randFindSlice3(c.availableWorkers, c.availableFighters, c.availableUniversal, f)
	default:
		panic("unexpected flags combination")
	}
}

func (c *colonyAgentContainer) findSlice(slice []*colonyAgentNode, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	for _, a := range slice {
		if f(a) {
			return a
		}
	}
	return nil
}

func (c *colonyAgentContainer) findSlice2(slice1, slice2 []*colonyAgentNode, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	if c.rand.Bool() {
		slice1, slice2 = slice2, slice1
	}
	if result := c.findSlice(slice1, f); result != nil {
		return result
	}
	return c.findSlice(slice2, f)
}

func (c *colonyAgentContainer) findSlice3(slice1, slice2, slice3 []*colonyAgentNode, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	c.sortTmp[0] = slice1
	c.sortTmp[1] = slice2
	c.sortTmp[2] = slice3
	gmath.Shuffle(c.rand, c.sortTmp[:])
	for _, slice := range c.sortTmp[:] {
		if result := c.findSlice(slice, f); result != nil {
			return result
		}
	}
	return nil
}

func (c *colonyAgentContainer) randFindSlice2(slice1, slice2 []*colonyAgentNode, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	if c.rand.Bool() {
		slice1, slice2 = slice2, slice1
	}
	if result := c.randFindSlice(slice1, f); result != nil {
		return result
	}
	return c.randFindSlice(slice2, f)
}

func (c *colonyAgentContainer) randFindSlice3(slice1, slice2, slice3 []*colonyAgentNode, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	c.sortTmp[0] = slice1
	c.sortTmp[1] = slice2
	c.sortTmp[2] = slice3
	gmath.Shuffle(c.rand, c.sortTmp[:])
	for _, slice := range c.sortTmp[:] {
		if result := c.randFindSlice(slice, f); result != nil {
			return result
		}
	}
	return nil
}

func (c *colonyAgentContainer) randFindSlice(agents []*colonyAgentNode, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	if len(agents) == 0 {
		return nil
	}
	var slider gmath.Slider
	slider.SetBounds(0, len(agents)-1)
	slider.TrySetValue(c.rand.IntRange(0, len(agents)-1))
	for i := 0; i < len(agents); i++ {
		a := agents[slider.Value()]
		if f(a) {
			return a
		}
		slider.Inc()
	}
	return nil
}
