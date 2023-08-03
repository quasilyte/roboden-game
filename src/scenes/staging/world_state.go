package staging

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gedraw"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/userdevice"
	"github.com/quasilyte/roboden-game/viewport"
)

type worldState struct {
	rand      *gmath.Rand
	localRand *gmath.Rand

	sessionState *session.State

	rootScene  *ge.Scene
	nodeRunner *nodeRunner

	stage   *viewport.CameraStage
	cameras []*viewport.Camera

	visionCircle *ebiten.Image

	players        []player
	allColonies    []*colonyCoreNode
	essenceSources []*essenceSourceNode
	creeps         []*creepNode
	constructions  []*constructionNode
	walls          []*wallClusterNode
	forests        []*forestClusterNode
	teleporters    []*teleporterNode

	boss              *creepNode
	wispLair          *creepNode
	fortress          *creepNode
	creepCoordinator  *creepCoordinator
	creepsPlayerState *creepsPlayerState

	creepClusterSize       float64
	creepClusterMultiplier float64
	creepClusters          [8][8][]*creepNode
	fallbackCreepCluster   []*creepNode

	graphicsSettings session.GraphicsSettings
	tier2recipes     []gamedata.AgentMergeRecipe
	tier2recipeIndex map[gamedata.RecipeSubject][]gamedata.AgentMergeRecipe
	turretDesign     *gamedata.AgentStats
	coreDesign       *gamedata.ColonyCoreStats

	hasForests           bool
	droneLabels          bool
	debugLogs            bool
	evolutionEnabled     bool
	movementEnabled      bool
	cameraShakingEnabled bool
	screenButtonsEnabled bool

	hintsMode int

	width      float64
	height     float64
	rect       gmath.Rect
	innerRect  gmath.Rect
	spawnAreas []gmath.Rect

	droneHealthMultiplier float64
	creepHealthMultiplier float64
	bossHealthMultiplier  float64
	oilRegenMultiplier    float64

	numRedCrystals int

	gridCounters map[int]uint8
	pathgrid     *pathing.Grid
	bfs          *pathing.GreedyBFS

	result battleResults

	simulation   bool
	seedKind     gamedata.SeedKind
	config       *gamedata.LevelConfig
	gameSettings *session.GameSettings
	deviceInfo   userdevice.Info

	projectilePool []*projectileNode

	tmpTargetSlice  []targetable
	tmpTargetSlice2 []targetable
	tmpColonySlice  []*colonyCoreNode

	inputMode string

	EventCheckDefeatState gsignal.Event[gsignal.Void]
	EventColonyCreated    gsignal.Event[*colonyCoreNode]

	EventCameraShake gsignal.Event[CameraShakeData]
}

type CameraShakeData struct {
	Power int
	Pos   gmath.Vec
}

func (w *worldState) ShakeCamera(power int, pos gmath.Vec) {
	if !w.cameraShakingEnabled {
		return
	}
	w.EventCameraShake.Emit(CameraShakeData{Power: power, Pos: pos})
}

func (w *worldState) Adjust2x2CellPos(pos gmath.Vec, offset float64) gmath.Vec {
	aligned := w.pathgrid.AlignPos2x2(pos)
	if offset == 0 {
		return roundedPos(aligned)
	}
	return roundedPos(aligned.Add(w.rand.Offset(-offset, offset)))
}

func (w *worldState) AdjustCellPos(pos gmath.Vec, offset float64) gmath.Vec {
	aligned := w.pathgrid.AlignPos(pos)
	return roundedPos(aligned.Add(w.rand.Offset(-offset, offset)))
}

func (w *worldState) MarkPos(pos gmath.Vec) {
	w.MarkCell(w.pathgrid.PosToCoord(pos))
}

func (w *worldState) UnmarkPos(pos gmath.Vec) {
	w.UnmarkCell(w.pathgrid.PosToCoord(pos))
}

func (w *worldState) CellIsFree2x2(cell pathing.GridCoord) bool {
	p := w.pathgrid
	return p.CellIsFree(cell) &&
		p.CellIsFree(cell.Add(pathing.GridCoord{X: -1})) &&
		p.CellIsFree(cell.Add(pathing.GridCoord{X: -1, Y: -1})) &&
		p.CellIsFree(cell.Add(pathing.GridCoord{Y: -1}))
}

func (w *worldState) MarkPos2x2(pos gmath.Vec) {
	cell := w.pathgrid.PosToCoord(pos)
	w.MarkCell(cell)
	w.MarkCell(cell.Add(pathing.GridCoord{X: -1}))
	w.MarkCell(cell.Add(pathing.GridCoord{X: -1, Y: -1}))
	w.MarkCell(cell.Add(pathing.GridCoord{Y: -1}))
}

func (w *worldState) UnmarkPos2x2(pos gmath.Vec) {
	cell := w.pathgrid.PosToCoord(pos)
	w.UnmarkCell(cell)
	w.UnmarkCell(cell.Add(pathing.GridCoord{X: -1}))
	w.UnmarkCell(cell.Add(pathing.GridCoord{X: -1, Y: -1}))
	w.UnmarkCell(cell.Add(pathing.GridCoord{Y: -1}))
}

func (w *worldState) MarkCell(coord pathing.GridCoord) {
	key := w.pathgrid.CoordToIndex(coord)
	if v := w.gridCounters[key]; v == 0 {
		w.pathgrid.MarkCell(coord)
	}
	w.gridCounters[key]++
}

func (w *worldState) UnmarkCell(coord pathing.GridCoord) {
	key := w.pathgrid.CoordToIndex(coord)
	if v := w.gridCounters[key]; v == 1 {
		w.pathgrid.UnmarkCell(coord)
		delete(w.gridCounters, key)
	} else {
		w.gridCounters[key]--
	}
}

func (w *worldState) Init() {
	w.gridCounters = make(map[int]uint8)

	{
		pad := 160.0
		offscreenPad := 160.0
		w.spawnAreas = []gmath.Rect{
			// right border (east)
			{Min: gmath.Vec{X: w.width, Y: pad}, Max: gmath.Vec{X: w.width + offscreenPad, Y: w.height - pad}},
			// bottom border (south)
			{Min: gmath.Vec{X: pad, Y: w.height}, Max: gmath.Vec{X: w.width - pad, Y: w.height + offscreenPad}},
			// left border (west)
			{Min: gmath.Vec{X: -offscreenPad, Y: pad}, Max: gmath.Vec{X: 0, Y: w.height - pad}},
			// top border (north)
			{Min: gmath.Vec{X: pad, Y: -offscreenPad}, Max: gmath.Vec{X: w.width - pad, Y: 0}},
		}
	}

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

	w.droneHealthMultiplier = 0.8 + (float64(w.config.DronesPower) * 0.2)
	w.creepHealthMultiplier = 0.25 + (float64(w.config.CreepDifficulty) * 0.25)
	w.bossHealthMultiplier = 0.7 + (float64(w.config.BossDifficulty) * 0.3)
	w.oilRegenMultiplier = float64(w.config.OilRegenRate) * 0.5

	if w.config.FogOfWar && w.config.ExecMode != gamedata.ExecuteSimulation {
		w.visionCircle = ebiten.NewImage(int(colonyVisionRadius*2), int(colonyVisionRadius*2))
		gedraw.DrawCircle(w.visionCircle, gmath.Vec{X: colonyVisionRadius, Y: colonyVisionRadius}, colonyVisionRadius, color.RGBA{A: 255})
	}
}

func (w *worldState) HasTreesAt(pos gmath.Vec, r float64) bool {
	for _, forest := range w.forests {
		if forest.CollidesWith(pos, r) {
			return true
		}
	}
	return false
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

func (w *worldState) GetPingDst(src *humanPlayer) *humanPlayer {
	if len(w.players) < 2 {
		return nil
	}
	if w.players[0] == src {
		return w.players[1].(*humanPlayer)
	}
	return w.players[0].(*humanPlayer)
}

func (w *worldState) Update() {
	w.fallbackCreepCluster = w.fallbackCreepCluster[:0]
	for y := range w.creepClusters {
		for x := range w.creepClusters[y] {
			w.creepClusters[y][x] = w.creepClusters[y][x][:0]
		}
	}

	for _, creep := range w.creeps {
		if creep.marked == 0 {
			x, y := w.GetPosCell(creep.pos)
			if y < len(w.creepClusters) {
				if x < len(w.creepClusters[y]) {
					w.creepClusters[y][x] = append(w.creepClusters[y][x], creep)
					continue
				}
			}
		}
		w.fallbackCreepCluster = append(w.fallbackCreepCluster, creep)
	}
}

func (w *worldState) freeProjectileNode(p *projectileNode) {
	w.projectilePool = append(w.projectilePool, p)
}

func (w *worldState) NewWallClusterNode(config wallClusterConfig) *wallClusterNode {
	n := newWallClusterNode(config)
	w.walls = append(w.walls, n)
	return n
}

func (w *worldState) NewColonyCoreNode(config colonyConfig) *colonyCoreNode {
	playerState := config.Player.GetState()
	n := newColonyCoreNode(config)
	n.id = playerState.colonySeq
	playerState.colonySeq++
	n.EventDestroyed.Connect(nil, func(x *colonyCoreNode) {
		w.allColonies = xslices.Remove(w.allColonies, x)
		playerState.colonies = xslices.Remove(playerState.colonies, x)
		w.EventCheckDefeatState.Emit(gsignal.Void{})
	})
	w.allColonies = append(w.allColonies, n)
	playerState.colonies = append(playerState.colonies, n)
	w.EventColonyCreated.Emit(n)
	return n
}

func (w *worldState) NewConstructionNode(p player, pos, spriteOffset gmath.Vec, stats *constructionStats) *constructionNode {
	n := newConstructionNode(w, p, pos, spriteOffset, stats)
	n.EventDestroyed.Connect(nil, func(x *constructionNode) {
		if stats == colonyCoreConstructionStats {
			w.UnmarkPos2x2(x.pos)
		} else {
			w.UnmarkPos(x.pos)
		}
		w.constructions = xslices.Remove(w.constructions, x)
	})
	if stats == colonyCoreConstructionStats {
		w.MarkPos2x2(pos)
	} else {
		w.MarkPos(pos)
	}
	w.constructions = append(w.constructions, n)
	return n
}

func (w *worldState) EliteCrawlerChance() float64 {
	if w.creepsPlayerState != nil {
		return gmath.ClampMax(w.creepsPlayerState.techLevel+0.2, 1.0)
	}

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

func (w *worldState) NewCreepNode(pos gmath.Vec, stats *gamedata.CreepStats) *creepNode {
	n := newCreepNode(w, stats, pos)
	n.EventDestroyed.Connect(nil, func(x *creepNode) {
		if stats.Building {
			w.UnmarkPos(pos)
		}
		w.creeps = xslices.Remove(w.creeps, x)
		if x.stats.Kind == gamedata.CreepCrawler {
			w.creepCoordinator.crawlers = xslices.Remove(w.creepCoordinator.crawlers, x)
		}
		w.result.CreepFragScore += x.fragScore
		switch x.stats.Kind {
		case gamedata.CreepBase:
			w.result.CreepBasesDestroyed++
		case gamedata.CreepWispLair:
			w.wispLair = nil
		case gamedata.CreepFortress:
			w.fortress = nil
		case gamedata.CreepUberBoss:
			if !x.IsFlying() {
				w.result.GroundBossDefeat = true
			}
			w.boss = nil
			w.EventCheckDefeatState.Emit(gsignal.Void{})
		default:
			w.result.CreepsDefeated++
		}
	})
	if stats.Building {
		w.MarkPos(pos)
	}
	w.creeps = append(w.creeps, n)
	if stats.Kind == gamedata.CreepCrawler {
		w.creepCoordinator.crawlers = append(w.creepCoordinator.crawlers, n)
	}
	return n
}

func (w *worldState) CreateScrapsAt(stats *essenceSourceStats, pos gmath.Vec) {
	if w.HasTreesAt(pos, 16) {
		return
	}
	scraps := w.NewEssenceSourceNode(stats, pos)
	w.nodeRunner.AddObject(scraps)
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
		if !stats.passable {
			w.UnmarkPos(x.pos)
		}
		w.essenceSources = xslices.Remove(w.essenceSources, x)
	})
	if !stats.passable {
		w.MarkPos(pos)
	}
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
		// Since normal drones can't be inside forest, this condition will suffice.
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

func (w *worldState) WalkCreeps(pos gmath.Vec, r float64, f func(creep *creepNode) bool) {
	creeps := w.creeps
	if len(creeps) == 0 {
		return
	}

	// Find a sector that contains this pos.
	cellX, cellY := w.GetPosCell(pos)
	cellRect := w.GetCellRect(cellX, cellY)

	// Determine how many sectors we need to consider.
	// In the simplest case, it's a single sector,
	// but sometimes we need to check the adjacent sectors too.
	startX := cellX
	startY := cellY
	endX := cellX
	endY := cellY
	searchRange := r
	leftmostPos := gmath.Vec{X: pos.X - searchRange, Y: pos.Y - searchRange}
	rightmostPos := gmath.Vec{X: pos.X + searchRange, Y: pos.Y + searchRange}
	if leftmostPos.X < cellRect.Min.X {
		delta := cellRect.Min.X - leftmostPos.X
		startX -= int(math.Ceil((searchRange - delta) * w.creepClusterMultiplier))
	}
	if rightmostPos.X > cellRect.Max.X {
		delta := rightmostPos.X - cellRect.Max.X
		endX += int(math.Ceil((searchRange - delta) * w.creepClusterMultiplier))
	}
	if leftmostPos.Y < cellRect.Min.Y {
		delta := cellRect.Min.Y - leftmostPos.Y
		startY -= int(math.Ceil((searchRange - delta) * w.creepClusterMultiplier))
	}
	if rightmostPos.Y > cellRect.Max.Y {
		delta := rightmostPos.Y - cellRect.Max.Y
		endY += int(math.Ceil((searchRange - delta) * w.creepClusterMultiplier))
	}

	startX = gmath.Clamp(startX, 0, 7)
	startY = gmath.Clamp(startY, 0, 7)
	endX = gmath.Clamp(endX, 0, 7)
	endY = gmath.Clamp(endY, 0, 7)
	numStepsX := endX - startX + 1
	numStepsY := endY - startY + 1

	// Now decide the sector traversal order.
	// This is needed to add some randomness to the target selection.
	dx := 1
	if w.rand.Bool() {
		dx = -1
		startX = endX
	}
	dy := 1
	if w.rand.Bool() {
		dy = -1
		startY = endY
	}

	for i, y := 0, startY; i < numStepsY; i, y = i+1, y+dy {
		for j, x := 0, startX; j < numStepsX; j, x = j+1, x+dx {
			clusterCreeps := w.creepClusters[y][x]
			if randIterate(w.rand, clusterCreeps, f) != nil {
				return
			}
		}
	}

	// New creeps are created outside of the map, so they end up
	// in the fallback cluster that includes everything that is out of bounds.
	if len(w.fallbackCreepCluster) != 0 {
		randIterate(w.rand, w.fallbackCreepCluster, f)
	}
}

func (w *worldState) FindTargetableAgents(pos gmath.Vec, skipGround bool, r float64, f func(a *colonyAgentNode) bool) {
	// TODO: use an agent container for turrets too?
	// Also, this "find" function is used to collect N units, not a single unit (see its usage).

	found := false

	if !skipGround {
		radiusSqr := r * r

		// Turrets have the highest targeting priority.
		randIterate(w.rand, w.allColonies, func(c *colonyCoreNode) bool {
			if len(c.turrets) == 0 {
				return false
			}
			for _, turret := range c.turrets {
				if turret.insideForest {
					continue
				}
				distSqr := turret.pos.DistanceSquaredTo(pos)
				if distSqr > radiusSqr {
					continue
				}
				if f(turret) {
					found = true
					return true
				}
			}
			return false
		})
		if found {
			return
		}

		// Roombas have the second priority.
		randIterate(w.rand, w.allColonies, func(c *colonyCoreNode) bool {
			if len(c.roombas) == 0 {
				return false
			}
			for _, roomba := range c.roombas {
				if roomba.insideForest {
					continue
				}
				distSqr := roomba.pos.DistanceSquaredTo(pos)
				if distSqr > radiusSqr {
					continue
				}
				if f(roomba) {
					found = true
					return true
				}
			}
			return false
		})
		if found {
			return
		}
	}

	randIterate(w.rand, w.allColonies, func(c *colonyCoreNode) bool {
		skipIdling := false
		dist := c.pos.DistanceTo(pos)
		colonyEffectiveRadius := c.PatrolRadius()
		if dist > colonyEffectiveRadius {
			skipIdling = (dist - colonyEffectiveRadius) > r
		}
		if a := w.findColonyAgent(c.agents.fighters, pos, r, skipIdling, f); a != nil {
			return true
		}
		if a := w.findColonyAgent(c.agents.workers, pos, r, skipIdling, f); a != nil {
			return true
		}
		return false
	})
}

func (w *worldState) GetColonyIndex(colony *colonyCoreNode) int {
	return xslices.Index(colony.player.GetState().colonies, colony)
}

func (w *worldState) onCreepTurretBuild() {
	if w.config.ExecMode != gamedata.ExecuteNormal {
		return
	}
	if w.config.GameMode != gamedata.ModeReverse {
		return
	}
	if w.config.PlayersMode != serverapi.PmodeSinglePlayer {
		return
	}
	if w.result.GroundControl {
		return
	}
	turrets := 0
	for _, creep := range w.creeps {
		if creep.stats != gamedata.TurretCreepStats {
			continue
		}
		turrets++
	}
	w.result.GroundControl = turrets >= 20
}
