package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/viewport"
)

type worldState struct {
	rand      *gmath.Rand
	localRand *gmath.Rand

	rootScene  *ge.Scene
	nodeRunner *nodeRunner

	camera *viewport.Camera

	essenceSources []*essenceSourceNode
	creeps         []*creepNode
	colonies       []*colonyCoreNode
	constructions  []*constructionNode
	walls          []*wallClusterNode
	teleporters    []*teleporterNode

	boss             *creepNode
	creepCoordinator *creepCoordinator

	creepClusterSize       float64
	creepClusterMultiplier float64
	creepClusters          [8][8][]*creepNode
	fallbackCreepCluster   []*creepNode

	graphicsSettings session.GraphicsSettings
	tier2recipes     []gamedata.AgentMergeRecipe
	tier2recipeIndex map[gamedata.RecipeSubject][]gamedata.AgentMergeRecipe
	turretDesign     *gamedata.AgentStats

	debugLogs        bool
	evolutionEnabled bool
	movementEnabled  bool

	width  float64
	height float64
	rect   gmath.Rect

	creepHealthMultiplier float64
	bossHealthMultiplier  float64
	oilRegenMultiplier    float64

	numRedCrystals int

	selectedColony *colonyCoreNode

	pathgrid *pathing.Grid
	bfs      *pathing.GreedyBFS

	result battleResults

	simulation bool
	config     *gamedata.LevelConfig

	projectilePool []*projectileNode

	tmpTargetSlice []targetable
	tmpColonySlice []*colonyCoreNode

	replayActions []serverapi.PlayerAction

	inputMode string

	EventColonyCreated gsignal.Event[*colonyCoreNode]
}

func (w *worldState) Init() {
	w.creepClusterSize = w.width * 0.125
	w.creepClusterMultiplier = 1.0 / w.creepClusterSize
	w.fallbackCreepCluster = make([]*creepNode, 0, 32)
	for y := range w.creepClusters {
		for x := range w.creepClusters {
			w.creepClusters[y][x] = make([]*creepNode, 0, 16)
		}
	}

	w.projectilePool = make([]*projectileNode, 0, 128)
	w.evolutionEnabled = true
	w.movementEnabled = true
	w.simulation = w.config.ExecMode == gamedata.ExecuteSimulation

	factions := []gamedata.FactionTag{
		gamedata.YellowFactionTag,
		gamedata.RedFactionTag,
		gamedata.BlueFactionTag,
		gamedata.GreenFactionTag,
	}
	kinds := []gamedata.ColonyAgentKind{
		gamedata.AgentWorker,
		gamedata.AgentScout,
	}
	w.tier2recipeIndex = make(map[gamedata.RecipeSubject][]gamedata.AgentMergeRecipe)
	for _, f := range factions {
		for _, k := range kinds {
			subject := gamedata.RecipeSubject{Kind: k, Faction: f}
			for _, recipe := range w.tier2recipes {
				if !recipe.Match1(subject) && !recipe.Match2(subject) {
					continue
				}
				w.tier2recipeIndex[subject] = append(w.tier2recipeIndex[subject], recipe)
			}
		}
	}

	w.result.OnlyTier1Military = true
	for _, recipe := range w.tier2recipes {
		if recipe.Result.CanPatrol {
			w.result.OnlyTier1Military = false
			break
		}
	}

	w.creepHealthMultiplier = 0.90 + (float64(w.config.CreepDifficulty) * 0.10)
	w.bossHealthMultiplier = 0.75 + (float64(w.config.BossDifficulty) * 0.25)
	w.oilRegenMultiplier = float64(w.config.OilRegenRate) * 0.5
}

func (w *worldState) newProjectileNode(config projectileConfig) *projectileNode {
	if len(w.projectilePool) != 0 {
		p := w.projectilePool[len(w.projectilePool)-1]
		initProjectileNode(p, config)
		w.projectilePool = w.projectilePool[:len(w.projectilePool)-1]
		return p
	}
	p := &projectileNode{}
	initProjectileNode(p, config)
	return p
}

func (w *worldState) GetCellRect(x, y int) gmath.Rect {
	min := gmath.Vec{X: float64(x) * w.creepClusterSize, Y: float64(y) * w.creepClusterSize}
	return gmath.Rect{
		Min: min,
		Max: min.Add(gmath.Vec{X: w.creepClusterSize, Y: w.creepClusterSize}),
	}
}

func (w *worldState) GetPosCell(pos gmath.Vec) (x int, y int) {
	cellX := int(pos.X * w.creepClusterMultiplier)
	cellY := int(pos.Y * w.creepClusterMultiplier)
	return cellX, cellY
}

func (w *worldState) Update() {
	w.fallbackCreepCluster = w.fallbackCreepCluster[:0]
	for y := range w.creepClusters {
		for x := range w.creepClusters[y] {
			w.creepClusters[y][x] = w.creepClusters[y][x][:0]
		}
	}

	for _, creep := range w.creeps {
		x, y := w.GetPosCell(creep.pos)
		if y < len(w.creepClusters) {
			if x < len(w.creepClusters[y]) {
				w.creepClusters[y][x] = append(w.creepClusters[y][x], creep)
				continue
			}
		}
		w.fallbackCreepCluster = append(w.fallbackCreepCluster, creep)
	}
}

func (w *worldState) freeProjectileNode(p *projectileNode) {
	w.projectilePool = append(w.projectilePool, p)
}

func (w *worldState) IsTutorial() bool {
	return w.config.Tutorial != nil
}

func (w *worldState) NewWallClusterNode(config wallClusterConfig) *wallClusterNode {
	n := newWallClusterNode(config)
	w.walls = append(w.walls, n)
	return n
}

func (w *worldState) NewColonyCoreNode(config colonyConfig) *colonyCoreNode {
	n := newColonyCoreNode(config)
	n.EventDestroyed.Connect(nil, func(x *colonyCoreNode) {
		w.colonies = xslices.Remove(w.colonies, x)
	})
	w.colonies = append(w.colonies, n)
	w.EventColonyCreated.Emit(n)
	return n
}

func (w *worldState) NewConstructionNode(pos gmath.Vec, stats *constructionStats) *constructionNode {
	n := newConstructionNode(w, pos, stats)
	n.EventDestroyed.Connect(nil, func(x *constructionNode) {
		w.constructions = xslices.Remove(w.constructions, x)
	})
	w.constructions = append(w.constructions, n)
	return n
}

func (w *worldState) NumActiveCrawlers() int {
	return len(w.creepCoordinator.crawlers)
}

func (w *worldState) EliteCrawlerChance() float64 {
	switch w.config.BossDifficulty {
	case 0:
		return 0
	case 1:
		return 0.2
	case 2:
		return 0.4
	default:
		return 0.65
	}
}

func (w *worldState) MaxActiveCrawlers() int {
	switch w.config.BossDifficulty {
	case 0:
		return 15
	case 1:
		return 25
	case 2:
		return 40
	default:
		return 60
	}
}

func (w *worldState) NewCreepNode(pos gmath.Vec, stats *creepStats) *creepNode {
	n := newCreepNode(w, stats, pos)
	n.EventDestroyed.Connect(nil, func(x *creepNode) {
		w.creeps = xslices.Remove(w.creeps, x)
		if x.stats.kind == creepCrawler {
			w.creepCoordinator.crawlers = xslices.Remove(w.creepCoordinator.crawlers, x)
		}
		w.result.CreepFragScore += x.fragScore
		switch x.stats.kind {
		case creepBase:
			w.result.CreepBasesDestroyed++
		case creepUberBoss:
			if !x.IsFlying() {
				w.result.GroundBossDefeat = true
			}
			w.boss = nil
		default:
			w.result.CreepsDefeated++
		}
	})
	w.creeps = append(w.creeps, n)
	if stats.kind == creepCrawler {
		w.creepCoordinator.crawlers = append(w.creepCoordinator.crawlers, n)
	}
	return n
}

func (w *worldState) NewEssenceSourceNode(stats *essenceSourceStats, pos gmath.Vec) *essenceSourceNode {
	n := newEssenceSourceNode(w, stats, pos)
	if stats.regenDelay != 0 && w.oilRegenMultiplier != 0 {
		// 0.5 => 1.5
		// 1.0 => 1.0
		// 1.5 => 0.5
		n.recoverDelayTimer = (2.0 - w.oilRegenMultiplier) * stats.regenDelay
	}
	n.EventDestroyed.Connect(nil, func(x *essenceSourceNode) {
		w.essenceSources = xslices.Remove(w.essenceSources, x)
	})
	w.essenceSources = append(w.essenceSources, n)
	return n
}

var nearBaseModeTable = [256]bool{
	agentModeAlignStandby:   true,
	agentModeStandby:        true,
	agentModePatrol:         true,
	agentModeRepairBase:     true,
	agentModeRepairTurret:   true,
	agentModeRecycleReturn:  true,
	agentModeRecycleLanding: true,
	agentModeBuildBuilding:  true,
}

func (w *worldState) findColonyAgent(agents []*colonyAgentNode, pos gmath.Vec, r float64, skipIdling bool, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	if len(agents) == 0 {
		return nil
	}

	var slider gmath.Slider
	slider.SetBounds(0, len(agents)-1)
	slider.TrySetValue(w.rand.IntRange(0, len(agents)-1))
	radiusSqr := r * r
	for i := 0; i < len(agents); i++ {
		slider.Inc()
		a := agents[slider.Value()]
		if skipIdling && nearBaseModeTable[byte(a.mode)] {
			continue
		}
		if a.IsCloaked() {
			continue
		}
		distSqr := a.pos.DistanceSquaredTo(pos)
		if distSqr > radiusSqr {
			continue
		}
		if f(a) {
			return a
		}
	}
	return nil
}

func (w *worldState) BuildPath(from, to gmath.Vec) pathing.BuildPathResult {
	return w.bfs.BuildPath(w.pathgrid, w.pathgrid.PosToCoord(from), w.pathgrid.PosToCoord(to))
}

func (w *worldState) FindColonyAgent(pos gmath.Vec, r float64, f func(a *colonyAgentNode) bool) {
	// TODO: use an agent container for turrets too?
	// Randomized order for iteration would be good here.
	// Also, this "find" function is used to collect N units, not a single unit (see its usage).

	radiusSqr := r * r

	// Turrets have the highest targeting priority.
	for _, c := range w.colonies {
		if len(c.turrets) == 0 {
			continue
		}
		for _, turret := range c.turrets {
			distSqr := turret.pos.DistanceSquaredTo(pos)
			if distSqr > radiusSqr {
				continue
			}
			if f(turret) {
				return
			}
		}
	}

	// Roombas have the second priority.
	for _, c := range w.colonies {
		if len(c.roombas) == 0 {
			continue
		}
		for _, roomba := range c.roombas {
			distSqr := roomba.pos.DistanceSquaredTo(pos)
			if distSqr > radiusSqr {
				continue
			}
			if f(roomba) {
				return
			}
		}
	}

	for _, c := range w.colonies {
		skipIdling := false
		dist := c.pos.DistanceTo(pos)
		colonyEffectiveRadius := c.realRadius * 0.8
		if dist > colonyEffectiveRadius {
			skipIdling = (dist - colonyEffectiveRadius) > r
		}
		if a := w.findColonyAgent(c.agents.fighters, pos, r, skipIdling, f); a != nil {
			return
		}
		if a := w.findColonyAgent(c.agents.workers, pos, r, skipIdling, f); a != nil {
			return
		}
	}
}

func (w *worldState) GetColonyIndex(colony *colonyCoreNode) int {
	return xslices.Index(w.colonies, colony)
}
