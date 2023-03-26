package staging

import (
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/viewport"
)

type worldState struct {
	rand *gmath.Rand

	camera *viewport.Camera

	essenceSources []*essenceSourceNode
	creeps         []*creepNode
	colonies       []*colonyCoreNode
	constructions  []*constructionNode
	walls          []*wallClusterNode

	boss             *creepNode
	creepCoordinator *creepCoordinator

	graphicsSettings session.GraphicsSettings
	tier2recipes     []gamedata.AgentMergeRecipe
	tier2recipeIndex map[gamedata.RecipeSubject][]gamedata.AgentMergeRecipe

	debug            bool
	evolutionEnabled bool
	movementEnabled  bool

	width  float64
	height float64
	rect   gmath.Rect

	creepHealthMultiplier float64
	bossHealthMultiplier  float64

	numRedCrystals int

	selectedColony *colonyCoreNode

	pathgrid *pathing.Grid
	bfs      *pathing.GreedyBFS

	result battleResults

	config *session.LevelConfig

	tmpTargetSlice []projectileTarget
	tmpColonySlice []*colonyCoreNode
}

func (w *worldState) Init() {
	w.evolutionEnabled = true
	w.movementEnabled = true

	factions := []gamedata.FactionTag{
		gamedata.YellowFactionTag,
		gamedata.RedFactionTag,
		gamedata.BlueFactionTag,
		gamedata.GreenFactionTag,
	}
	kinds := []gamedata.ColonyAgentKind{
		gamedata.AgentWorker,
		gamedata.AgentMilitia,
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

	w.creepHealthMultiplier = 1.0 + (float64(w.config.CreepDifficulty-1) * 0.20)
	w.bossHealthMultiplier = 1.0 + (float64(w.config.BossDifficulty-1) * 0.15)
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
		return 20
	case 2:
		return 30
	default:
		return 40
	}
}

func (w *worldState) NewCreepNode(pos gmath.Vec, stats *creepStats) *creepNode {
	n := newCreepNode(w, stats, pos)
	n.EventDestroyed.Connect(nil, func(x *creepNode) {
		w.creeps = xslices.Remove(w.creeps, x)
		if x.stats.kind == creepCrawler {
			w.creepCoordinator.crawlers = xslices.Remove(w.creepCoordinator.crawlers, x)
		}
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
	n.EventDestroyed.Connect(nil, func(x *essenceSourceNode) {
		w.essenceSources = xslices.Remove(w.essenceSources, x)
	})
	w.essenceSources = append(w.essenceSources, n)
	return n
}

func (w *worldState) findColonyAgent(agents []*colonyAgentNode, pos gmath.Vec, r float64, skipIdling bool, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	if len(agents) == 0 {
		return nil
	}
	var slider gmath.Slider
	slider.SetBounds(0, len(agents)-1)
	slider.TrySetValue(w.rand.IntRange(0, len(agents)-1))
	for i := 0; i < len(agents); i++ {
		slider.Inc()
		a := agents[slider.Value()]
		if a.IsCloaked() {
			continue
		}
		if skipIdling && a.mode == agentModeStandby {
			continue
		}
		dist := a.pos.DistanceTo(pos)
		if dist > r {
			continue
		}
		if f(a) {
			return a
		}
	}
	return nil
}

func (w *worldState) Update(delta float64) {
	w.creepCoordinator.Update(delta)
}

func (w *worldState) BuildPath(from, to gmath.Vec) pathing.BuildPathResult {
	return w.bfs.BuildPath(w.pathgrid, w.pathgrid.PosToCoord(from), w.pathgrid.PosToCoord(to))
}

func (w *worldState) FindColonyAgent(pos gmath.Vec, r float64, f func(a *colonyAgentNode) bool) {
	// TODO: use an agent container for turrets too?
	// Randomized order for iteration would be good here.
	// Also, this "find" function is used to collect N units, not a single unit (see its usage).
	for _, c := range w.colonies {
		if len(c.turrets) == 0 {
			continue
		}
		for _, turret := range c.turrets {
			dist := turret.pos.DistanceTo(pos)
			if dist > r {
				continue
			}
			if f(turret) {
				return
			}
		}
	}

	// TODO: use agents container methods here.
	for _, c := range w.colonies {
		skipIdling := c.pos.DistanceTo(pos)*0.3 > r
		if a := w.findColonyAgent(c.agents.fighters, pos, r, skipIdling, f); a != nil {
			return
		}
		if a := w.findColonyAgent(c.agents.workers, pos, r, skipIdling, f); a != nil {
			return
		}
	}
}
