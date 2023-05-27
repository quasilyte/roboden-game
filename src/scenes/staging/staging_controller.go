package staging

import (
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
	"github.com/quasilyte/roboden-game/viewport"
)

type Controller struct {
	state *session.State

	backController ge.SceneController

	scene  *ge.Scene
	world  *worldState
	config gamedata.LevelConfig

	uiLayer *uiLayer

	musicPlayer *musicPlayer

	exitNotice        *messageNode
	transitionQueued  bool
	gameFinished      bool
	victoryCheckDelay float64

	camera *cameraManager

	tutorialManager *tutorialManager
	messageManager  *messageManager
	recipeTab       *recipeTabNode

	arenaManager *arenaManager
	nodeRunner   *nodeRunner

	cursor *gameui.CursorNode

	debugInfo        *ge.Label
	debugUpdateDelay float64

	replayActions [][]serverapi.PlayerAction

	// rects []*ge.Rect
}

func NewController(state *session.State, config gamedata.LevelConfig, back ge.SceneController) *Controller {
	return &Controller{
		state:          state,
		backController: back,
		config:         config,
	}
}

func (c *Controller) SetReplayActions(actions [][]serverapi.PlayerAction) {
	c.replayActions = actions
}

func (c *Controller) initTextures() {
	stunnerCreepStats.beamTexture = ge.NewHorizontallyRepeatedTexture(c.scene.LoadImage(assets.ImageStunnerLine), stunnerCreepStats.weapon.AttackRange)
	uberBossCreepStats.beamTexture = ge.NewHorizontallyRepeatedTexture(c.scene.LoadImage(assets.ImageBossLaserLine), uberBossCreepStats.weapon.AttackRange)
}

func (c *Controller) CenterDemoCamera(pos gmath.Vec) {
	c.camera.ToggleCamera(pos)
	c.camera.cinematicSwitchDelay = c.world.localRand.FloatRange(20, 30)
	c.camera.mode = camCinematic
}

func (c *Controller) IsExcitingDemoFrame() (gmath.Vec, bool) {
	pstate := c.world.players[0].GetState()

	if c.world.boss != nil && c.world.boss.bossStage != 0 {
		for _, creep := range c.world.creeps {
			if creep.stats.kind == creepServant {
				return creep.pos, true
			}
		}
	}

	if c.world.boss != nil && c.world.boss.health < c.world.boss.maxHealth*0.6 {
		return c.world.boss.pos, true
	}

	for _, colony := range pstate.colonies {
		// Many tier 3 drones are around?
		if colony.agents.tier3Num >= 4 {
			return colony.pos, true
		}

		// Maybe tether tower is around?
		if c.world.turretDesign == gamedata.TetherBeaconAgentStats {
			for _, turret := range colony.turrets {
				if turret.pos.DistanceTo(colony.pos) <= 260 {
					return colony.pos, true
				}
			}
		}

		// More than 2 mergings are happening?
		// Or maybe there are many elite units?
		// Or perhaps some units performs something spectacular?
		numMerges := 0
		numSuperElites := 0
		numElites := 0
		numSpectacular := 0
		numRepairs := 0
		numRoombaMerges := 0
		numAttacking := 0
		colony.agents.Each(func(a *colonyAgentNode) {
			if a.mode == agentModeMerging {
				numMerges++
				if a.stats == gamedata.RoombaAgentStats {
					numRoombaMerges++
				}
			}
			switch a.rank {
			case 1:
				numElites++
			case 2:
				numSuperElites++
			}
			switch a.mode {
			case agentModeKamikazeAttack, agentModeConsumeDrone, agentModeCloakHide, agentModeCourierFlight:
				numSpectacular++
			case agentModeRepairBase, agentModeRepairTurret:
				numRepairs++
			case agentModeFollow:
				numAttacking++
			}
		})
		if numAttacking >= 4 {
			return colony.pos, true
		}
		if numRoombaMerges >= 1 {
			return colony.pos, true
		}
		if numSpectacular >= 1 {
			return colony.pos, true
		}
		if numRepairs >= 2 {
			return colony.pos, true
		}
		if numSuperElites >= 2 {
			return colony.pos, true
		}
		if numElites >= 10 {
			return colony.pos, true
		}
		if numMerges >= 4 {
			return colony.pos, true
		}

		// Maybe a colony is heavily damaged?
		if colony.health <= (colony.maxHealth*0.75) && colony.health >= (colony.maxHealth*0.3) {
			return colony.pos, true
		}

		// Are we in a middle of a fight?
		danger, _ := calcPosDanger(c.world, pstate, colony.pos, 0.8*colony.PatrolRadius())
		if danger > 70 {
			return colony.pos, true
		}

		// Does teleportation take place?
		if colony.mode == colonyModeTeleporting {
			return colony.pos, true
		}

		// Are we building a new base?
		for _, construction := range c.world.constructions {
			if construction.stats.Kind != constructBase {
				continue
			}
			// Is it too early or too late?
			if construction.progress < 0.1 || construction.progress > 0.9 {
				continue
			}
			if colony.mode != colonyModeNormal {
				continue
			}
			// It'll be boring without resources to finish it.
			if colony.resources < 100 || colony.NumAgents() < 30 {
				continue
			}
			if colony.pos.DistanceTo(construction.pos) > 180 {
				continue
			}
			return colony.pos, true
		}

	}

	// More than 1 base is flying at the same time?
	if len(pstate.colonies) > 1 {
		numFlying := 0
		flying := pstate.colonies[0]
		for _, c := range pstate.colonies {
			if c.mode == colonyModeRelocating {
				numFlying++
				flying = c
			}
		}
		if numFlying > 1 {
			return flying.pos, true
		}
	}

	// Maybe there are 3 or more colonies already?
	if len(pstate.colonies) >= 3 {
		return pstate.colonies[0].pos, true
	}

	return gmath.Vec{}, false
}

func (c *Controller) GetSimulationResult() (serverapi.GameResults, bool) {
	var result serverapi.GameResults
	if !c.gameFinished {
		return result, false
	}
	for _, p := range c.world.players {
		p, ok := p.(*replayPlayer)
		if ok && len(p.state.replay) != 0 {
			panic(errExcessiveAcions)
		}
	}
	result.Victory = c.world.result.Victory
	result.Score = c.world.result.Score
	result.Time = int(math.Floor(c.world.result.TimePlayed.Seconds()))
	result.Ticks = c.world.result.Ticks
	return result, true
}

func (c *Controller) Init(scene *ge.Scene) {
	scene.Context().Rand.SetSeed((c.config.Seed + 42) * 21917)
	c.scene = scene

	c.initTextures()

	c.musicPlayer = newMusicPlayer(scene)
	c.musicPlayer.Start()

	c.uiLayer = newUILayer()

	if c.state.CPUProfile != "" {
		f, err := os.Create(c.state.CPUProfile)
		if err != nil {
			panic(err)
		}
		c.state.CPUProfileWriter = f
		pprof.StartCPUProfile(f)
	}
	if c.state.MemProfile != "" {
		f, err := os.Create(c.state.MemProfile)
		if err != nil {
			panic(err)
		}
		c.state.MemProfileWriter = f
	}

	var worldSize float64
	switch c.config.WorldSize {
	case 0:
		worldSize = 1856
	case 1:
		worldSize = 2368
	case 2:
		worldSize = 2880
	case 3:
		worldSize = 3392
	}

	viewportWorld := &viewport.World{
		Width:  worldSize,
		Height: worldSize,
	}

	var localRand gmath.Rand
	localRand.SetSeed(time.Now().Unix())

	gameSpeed := 1.0
	switch c.config.GameSpeed {
	case 1:
		gameSpeed = 1.2
	case 2:
		gameSpeed = 1.5
	}
	c.nodeRunner = newNodeRunner(gameSpeed)
	c.nodeRunner.Init(c.scene)

	tier2recipes := make([]gamedata.AgentMergeRecipe, len(c.config.Tier2Recipes))
	for i, droneName := range c.config.Tier2Recipes {
		tier2recipes[i] = gamedata.FindRecipeByName(droneName)
	}

	world := &worldState{
		uiLayer:          c.uiLayer,
		rootScene:        scene,
		nodeRunner:       c.nodeRunner,
		graphicsSettings: c.state.Persistent.Settings.Graphics,
		pathgrid:         pathing.NewGrid(viewportWorld.Width, viewportWorld.Height),
		config:           &c.config,
		gameSettings:     &c.state.Persistent.Settings,
		deviceInfo:       c.state.Device,
		debugLogs:        c.state.Persistent.Settings.DebugLogs,
		rand:             scene.Rand(),
		localRand:        &localRand,
		tmpTargetSlice:   make([]targetable, 0, 20),
		tmpColonySlice:   make([]*colonyCoreNode, 0, 4),
		width:            viewportWorld.Width,
		height:           viewportWorld.Height,
		rect: gmath.Rect{
			Max: gmath.Vec{
				X: viewportWorld.Width,
				Y: viewportWorld.Height,
			},
		},
		innerRect: gmath.Rect{
			Min: gmath.Vec{X: 180, Y: 180},
			Max: gmath.Vec{
				X: viewportWorld.Width - 180,
				Y: viewportWorld.Height - 180,
			},
		},
		tier2recipes: tier2recipes,
		turretDesign: gamedata.FindTurretByName(c.config.TurretDesign),
	}
	world.inputMode = c.state.DetectInputMode()
	world.creepCoordinator = newCreepCoordinator(world)
	world.bfs = pathing.NewGreedyBFS(world.pathgrid.Size())
	c.world = world
	world.Init()

	// Background generation is an expensive operation.
	// Don't do it inside simulation (headless) mode.
	var bg *ge.TiledBackground
	if c.config.ExecMode != gamedata.ExecuteSimulation {
		// Use local rand for the tileset generation.
		// Otherwise, we'll get incorrect results during the simulation.
		bg = ge.NewTiledBackground(scene.Context())
		bg.LoadTilesetWithRand(scene.Context(), &localRand, viewportWorld.Width, viewportWorld.Height, assets.ImageBackgroundTiles, assets.RawTilesJSON)
	}
	{
		cam := viewport.NewCamera(viewportWorld, c.config.ExecMode == gamedata.ExecuteSimulation, 1920/2, 1080/2)
		cam.SetBackground(bg)
		c.camera = newCameraManager(c.world, cam)
		if c.config.ExecMode == gamedata.ExecuteDemo || c.config.ExecMode == gamedata.ExecuteReplay {
			c.camera.InitCinematicMode()
			c.camera.camera.CenterOn(c.world.rect.Center())
		} else {
			c.camera.InitManualMode(c.state.MainInput)
		}
	}
	c.world.camera = c.camera.camera // FIXME
	if c.config.ExecMode == gamedata.ExecuteReplay {
		c.CenterDemoCamera(c.world.rect.Center())
	}

	c.nodeRunner.world = world

	c.nodeRunner.creepCoordinator = world.creepCoordinator

	c.messageManager = newMessageManager(c.world, c.uiLayer)

	c.world.EventColonyCreated.Connect(c, func(colony *colonyCoreNode) {
		colony.EventUnderAttack.Connect(c, func(colony *colonyCoreNode) {
			center := c.camera.camera.Offset.Add(c.camera.camera.Rect.Center())
			if center.DistanceTo(colony.pos) < 250 {
				return
			}
			c.messageManager.AddMessage(queuedMessageInfo{
				text:          scene.Dict().Get("game.notice.base_under_attack"),
				trackedObject: colony,
				timer:         5,
				targetPos:     ge.Pos{Base: colony.GetPos()},
			})
		})
	})

	switch c.config.GameMode {
	case gamedata.ModeArena, gamedata.ModeInfArena:
		c.arenaManager = newArenaManager(world, c.uiLayer)
		c.nodeRunner.AddObject(c.arenaManager)
		c.arenaManager.EventVictory.Connect(c, c.onVictoryTrigger)
	case gamedata.ModeClassic:
		classicManager := newClassicManager(world)
		c.nodeRunner.AddObject(classicManager)
		// TODO: victory trigger should go to the classic manager
	}

	c.cursor = gameui.NewCursorNode(c.state.MainInput, c.camera.camera.Rect)

	c.createPlayers()

	{
		g := newLevelGenerator(scene, bg, c.world)
		g.Generate()
	}

	for _, p := range c.world.players {
		p.Init()
	}

	scene.AddGraphics(c.camera.camera)

	// if c.config.ExtraUI {
	// 	c.rpanel = newRpanelNode(c.world, c.uiLayer)
	// 	scene.AddObject(c.rpanel)
	// }

	// if c.world.IsTutorial() {
	// 	c.tutorialManager = newTutorialManager(c.state.MainInput, c.world, c.uiLayer, c.messageManager)
	// 	c.nodeRunner.AddObject(c.tutorialManager)
	// 	if c.rpanel != nil {
	// 		c.tutorialManager.EventRequestPanelUpdate.Connect(c, c.onPanelUpdateRequested)
	// 	}
	// 	c.tutorialManager.EventTriggerVictory.Connect(c, c.onVictoryTrigger)
	// }

	if c.state.Persistent.Settings.ShowFPS || c.state.Persistent.Settings.ShowTimer {
		c.debugInfo = scene.NewLabel(assets.FontSmall)
		c.debugInfo.ColorScale.SetColor(ge.RGB(0xffffff))
		c.debugInfo.Pos.Offset = gmath.Vec{X: 10, Y: 10}
		c.uiLayer.AddGraphics(c.debugInfo)
	}

	c.camera.camera.SortBelowLayer()

	// {
	// 	cols, rows := c.world.pathgrid.Size()
	// 	for row := 0; row < rows; row++ {
	// 		for col := 0; col < cols; col++ {
	// 			coord := pathing.GridCoord{X: col, Y: row}
	// 			if c.world.pathgrid.CellIsFree(coord) {
	// 				continue
	// 			}
	// 			rect := ge.NewRect(scene.Context(), pathing.CellSize, pathing.CellSize)
	// 			rect.FillColorScale.SetRGBA(200, 50, 50, 100)
	// 			rect.Pos.Offset = c.world.pathgrid.CoordToPos(coord)
	// 			c.camera.AddGraphics(rect)
	// 			c.rects = append(c.rects, rect)
	// 		}
	// 	}
	// 	c.world.EventUnmarked.Connect(c, func(pos gmath.Vec) {
	// 		index := xslices.IndexWhere(c.rects, func(rect *ge.Rect) bool {
	// 			return rect.Pos.Offset == pos
	// 		})
	// 		if index == -1 {
	// 			panic("??")
	// 		}
	// 		c.rects[index].Dispose()
	// 		c.rects[index] = c.rects[len(c.rects)-1]
	// 		c.rects = c.rects[:len(c.rects)-1]
	// 	})
	// 	c.world.EventMarked.Connect(c, func(pos gmath.Vec) {
	// 		rect := ge.NewRect(scene.Context(), pathing.CellSize, pathing.CellSize)
	// 		rect.FillColorScale.SetRGBA(200, 50, 50, 100)
	// 		rect.Pos.Offset = c.world.pathgrid.CoordToPos(c.world.pathgrid.PosToCoord(pos))
	// 		c.camera.AddGraphics(rect)
	// 		c.rects = append(c.rects, rect)
	// 	})
	// }

	{
		c.recipeTab = newRecipeTabNode(c.world)
		c.recipeTab.Visible = false
		c.uiLayer.AddGraphics(c.recipeTab)
		scene.AddObject(c.recipeTab)
	}

	scene.AddGraphics(c.uiLayer)

	scene.AddObject(c.cursor)
}

func (c *Controller) createPlayers() {
	c.world.players = make([]player, 0, len(c.config.Players))
	isSimulation := c.world.config.ExecMode == gamedata.ExecuteReplay ||
		c.world.config.ExecMode == gamedata.ExecuteSimulation
	for i, pk := range c.config.Players {
		choiceGen := newChoiceGenerator(c.world)
		choiceGen.EventChoiceSelected.Connect(c, c.onChoiceSelected)
		pstate := newPlayerState(c.camera)
		pstate.id = i

		var p player
		switch pk {
		case gamedata.PlayerHuman:
			if isSimulation {
				p = newReplayPlayer(c.world, pstate, choiceGen)
				pstate.replay = c.replayActions[i]
			} else {
				p = newHumanPlayer(c.world, pstate, c.state.MainInput, c.cursor, choiceGen)
			}
		case gamedata.PlayerComputer:
			p = newComputerPlayer(c.world, pstate, choiceGen)
		default:
			panic(fmt.Sprintf("unexpected player kind: %d", pk))
		}

		choiceGen.player = p
		c.nodeRunner.AddObject(choiceGen)
		c.world.players = append(c.world.players, p)
	}
}

func (c *Controller) onVictoryTrigger(gsignal.Void) {
	c.victory()
}

func (c *Controller) onExitButtonClicked() {
	if c.exitNotice != nil {
		c.leaveScene(c.backController)
		return
	}

	d := c.scene.Dict()
	c.uiLayer.Visible = true
	c.nodeRunner.SetPaused(true)
	c.exitNotice = newScreenTutorialHintNode(c.camera.camera, c.uiLayer, gmath.Vec{}, gmath.Vec{}, d.Get("game.exit.notice", c.world.inputMode))
	c.scene.AddObject(c.exitNotice)
	noticeSize := gmath.Vec{X: c.exitNotice.width, Y: c.exitNotice.height}
	noticeCenterPos := c.camera.camera.Rect.Center().Sub(noticeSize.Mulf(0.5))
	c.exitNotice.SetPos(noticeCenterPos)
}

func (c *Controller) executeAction(choice selectedChoice) bool {
	pstate := choice.Player.GetState()
	selectedColony := pstate.selectedColony

	if c.config.ExecMode == gamedata.ExecuteNormal {
		kind := serverapi.PlayerActionKind(choice.Index + 1)
		if choice.Option.special == specialChoiceMoveColony {
			kind = serverapi.ActionMove
		}
		a := serverapi.PlayerAction{
			Kind:           kind,
			Pos:            [2]float64{choice.Pos.X, choice.Pos.Y},
			SelectedColony: c.world.GetColonyIndex(selectedColony),
			Tick:           c.nodeRunner.ticks,
		}
		pstate.replay = append(pstate.replay, a)
	}

	if choice.Option.special == specialChoiceNone {
		switch choice.Faction {
		case gamedata.YellowFactionTag:
			c.world.result.YellowFactionUsed = true
		case gamedata.RedFactionTag:
			c.world.result.RedFactionUsed = true
		case gamedata.GreenFactionTag:
			c.world.result.GreenFactionUsed = true
		case gamedata.BlueFactionTag:
			c.world.result.BlueFactionUsed = true
		}
		selectedColony.factionWeights.AddWeight(choice.Faction, c.world.rand.FloatRange(0.15, 0.25))
		for _, e := range choice.Option.effects {
			// Use priorities.AddWeight directly here to avoid the signal.
			// We'll call UpdateMetrics() below ourselves.
			selectedColony.priorities.AddWeight(e.priority, e.value)
		}
		return true
	}

	var relocationPos gmath.Vec
	switch choice.Option.special {
	case specialAttack:
		c.launchAttack(selectedColony)
		return true
	case specialChoiceMoveColony:
		maxDist := selectedColony.MaxFlyDistance() * c.world.rand.FloatRange(0.9, 1.1)
		clickPos := choice.Pos
		clickDist := selectedColony.pos.DistanceTo(clickPos)
		dist := gmath.ClampMax(clickDist, maxDist)
		relocationVec := selectedColony.pos.VecTowards(clickPos, 1).Mulf(dist)
		relocationPos = correctedPos(c.world.rect, relocationVec.Add(selectedColony.pos), 128)
		return c.launchRelocation(selectedColony, dist, relocationPos)
	case specialIncreaseRadius:
		c.world.result.RadiusIncreases++
		selectedColony.realRadius += c.world.rand.FloatRange(16, 32)
		selectedColony.realRadiusSqr = selectedColony.realRadius * selectedColony.realRadius
		return true
	case specialDecreaseRadius:
		value := c.world.rand.FloatRange(40, 60)
		selectedColony.realRadius = gmath.ClampMin(selectedColony.realRadius-value, 96)
		selectedColony.realRadiusSqr = selectedColony.realRadius * selectedColony.realRadius
		return true
	case specialBuildGunpoint:
		stats := gunpointConstructionStats
		switch c.world.turretDesign {
		case gamedata.BeamTowerAgentStats:
			stats = beamTowerConstructionStats
		case gamedata.TetherBeaconAgentStats:
			stats = tetherBeaconConstructionStats
		}
		coord := c.world.pathgrid.PosToCoord(selectedColony.pos)
		freeCoord := randIterate(c.world.rand, colonyNearCellOffsets, func(offset pathing.GridCoord) bool {
			probe := coord.Add(offset)
			return c.world.pathgrid.CellIsFree(probe)
		})
		if !freeCoord.IsZero() {
			pos := c.world.pathgrid.CoordToPos(coord.Add(freeCoord))
			spriteOffset := roundedPos(c.world.rand.Offset(-3, 3))
			construction := c.world.NewConstructionNode(choice.Player, pos, spriteOffset, stats)
			c.nodeRunner.AddObject(construction)
			return true
		}
		return false

	case specialBuildColony:
		p := c.world.pathgrid
		stats := colonyCoreConstructionStats
		coord := p.PosToCoord(selectedColony.pos)
		freeCoord := randIterate(c.world.rand, colonyNear2x2CellOffsets, func(offset pathing.GridCoord) bool {
			probe := coord.Add(offset)
			return c.world.CellIsFree2x2(probe)
		})
		if !freeCoord.IsZero() {
			pos := p.CoordToPos(coord.Add(freeCoord)).Sub(gmath.Vec{X: 16, Y: 16})
			construction := c.world.NewConstructionNode(choice.Player, pos, gmath.Vec{}, stats)
			c.nodeRunner.AddObject(construction)
			return true
		}
		return false
	}

	return false
}

func (c *Controller) playPlayerSound(p player, sound resource.AudioID) {
	if _, ok := p.(*humanPlayer); ok {
		c.scene.Audio().PlaySound(sound)
	}
}

func (c *Controller) onChoiceSelected(choice selectedChoice) {
	if c.tutorialManager != nil {
		c.tutorialManager.OnChoice(choice)
	}

	if c.executeAction(choice) {
		c.playPlayerSound(choice.Player, assets.AudioChoiceMade)
	} else {
		c.playPlayerSound(choice.Player, assets.AudioError)
	}
}

func (c *Controller) launchAttack(selectedColony *colonyCoreNode) {
	if selectedColony.agents.NumAvailableFighters() == 0 {
		return
	}
	closeTargets := c.world.tmpTargetSlice[:0]
	maxDist := selectedColony.AttackRadius() * c.world.rand.FloatRange(0.95, 1.1)
	for _, creep := range c.world.creeps {
		if len(closeTargets) >= 5 {
			break
		}
		if creep.IsCloaked() {
			continue
		}
		if creep.pos.DistanceTo(selectedColony.pos) > maxDist {
			continue
		}
		closeTargets = append(closeTargets, creep)
	}
	if len(closeTargets) == 0 {
		return
	}
	maxDispatched := gmath.Clamp(int(float64(selectedColony.agents.NumAvailableFighters())*0.6), 1, 15)
	selectedColony.agents.Find(searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
		target := gmath.RandElem(c.world.rand, closeTargets)
		kind := gamedata.TargetGround
		if target.IsFlying() {
			kind = gamedata.TargetFlying
		}
		if !a.CanAttack(kind) {
			return false
		}
		maxDispatched--
		a.AssignMode(agentModeAttack, gmath.Vec{}, target)
		return maxDispatched <= 0
	})
}

func (c *Controller) launchRelocation(core *colonyCoreNode, dist float64, dst gmath.Vec) bool {
	coord := c.world.pathgrid.PosToCoord(dst)
	if c.world.CellIsFree2x2(coord) {
		pos := c.world.pathgrid.CoordToPos(coord).Sub(gmath.Vec{X: 16, Y: 16})
		core.doRelocation(pos)
		return true
	}

	freeCoord := randIterate(c.world.rand, colonyNear2x2CellOffsets, func(offset pathing.GridCoord) bool {
		probe := coord.Add(offset)
		return c.world.CellIsFree2x2(probe)
	})
	if !freeCoord.IsZero() {
		pos := c.world.pathgrid.CoordToPos(coord.Add(freeCoord)).Sub(gmath.Vec{X: 16, Y: 16})
		core.doRelocation(pos)
		return true
	}

	if dist > 160 {
		nextDst := dst.MoveTowards(core.pos, 96)
		return c.launchRelocation(core, dist-96, nextDst)
	}

	return false
}

func (c *Controller) defeat() {
	if c.transitionQueued {
		return
	}

	c.transitionQueued = true

	c.prepareBattleResults()
	c.scene.DelayedCall(2.0, func() {
		c.gameFinished = true
		c.world.result.Victory = false
		if c.config.ExecMode != gamedata.ExecuteSimulation {
			c.leaveScene(newResultsController(c.state, &c.config, c.backController, c.world.result, nil))
		}
	})
}

func (c *Controller) prepareBattleResults() {
	if c.config.ExecMode == gamedata.ExecuteNormal {
		c.world.result.Replay = make([][]serverapi.PlayerAction, len(c.world.players))
		for i, p := range c.world.players {
			c.world.result.Replay[i] = p.GetState().replay
		}
	}

	c.world.result.Ticks = c.nodeRunner.ticks
	c.world.result.TimePlayed = time.Second * time.Duration(c.nodeRunner.timePlayed)
	if c.arenaManager != nil {
		c.world.result.ArenaLevel = c.arenaManager.level
		if c.config.GameMode == gamedata.ModeArena {
			for _, creep := range c.world.creeps {
				if creep.stats == dominatorCreepStats {
					c.world.result.DominatorsSurvived++
				}
			}
		}
	}
	c.world.result.Score = calcScore(c.world)
	c.world.result.DifficultyScore = c.config.DifficultyScore
	c.world.result.DronePointsAllocated = c.config.DronePointsAllocated
}

func (c *Controller) victory() {
	if c.transitionQueued {
		return
	}

	c.transitionQueued = true

	c.scene.Audio().PlaySound(assets.AudioVictory)
	c.prepareBattleResults()
	c.scene.DelayedCall(5.0, func() {
		c.gameFinished = true
		c.world.result.Victory = true
		switch c.config.ExecMode {
		case gamedata.ExecuteNormal:
			t3set := map[gamedata.ColonyAgentKind]struct{}{}
			for _, colony := range c.world.players[0].GetState().colonies {
				colony.agents.Each(func(a *colonyAgentNode) {
					if a.stats.Tier != 3 {
						return
					}
					t3set[a.stats.Kind] = struct{}{}
				})
			}
			for k := range t3set {
				c.world.result.Tier3Drones = append(c.world.result.Tier3Drones, k)
			}
			c.leaveScene(newResultsController(c.state, &c.config, c.backController, c.world.result, nil))
		case gamedata.ExecuteDemo:
			c.leaveScene(c.backController)
		}
	})
}

func (c *Controller) handleDemoInput() {
	for _, p := range c.world.players {
		p.HandleInput()
	}
}

func (c *Controller) handleReplayActions() {
	for _, p := range c.world.players {
		p.HandleInput()
	}
}

func (c *Controller) handleInput() {
	mainInput := c.state.MainInput

	switch c.config.ExecMode {
	case gamedata.ExecuteReplay, gamedata.ExecuteSimulation:
		c.handleReplayActions()
		return
	case gamedata.ExecuteDemo:
		c.handleDemoInput()
		return
	}

	if mainInput.ActionIsJustPressed(controls.ActionToggleInterface) {
		c.uiLayer.Visible = !c.uiLayer.Visible
	}

	if mainInput.ActionIsJustPressed(controls.ActionShowRecipes) {
		if c.debugInfo != nil {
			c.debugInfo.Visible = c.recipeTab.Visible
		}
		c.recipeTab.Visible = !c.recipeTab.Visible
		c.world.result.OpenedEvolutionTab = true
	}

	if mainInput.ActionIsJustPressed(controls.ActionBack) {
		c.onExitButtonClicked()
		return
	}

	if mainInput.ActionIsJustPressed(controls.ActionPause) {
		c.nodeRunner.SetPaused(!c.nodeRunner.IsPaused())
		return
	}

	c.camera.HandleInput()

	for _, p := range c.world.players {
		p.HandleInput()
	}
}

func (c *Controller) checkVictory() {
	if c.transitionQueued {
		return
	}

	victory := false
	switch c.config.GameMode {
	case gamedata.ModeClassic:
		victory = c.world.boss == nil

	case gamedata.ModeArena:
		// Do nothing. This mode is endless.

	case gamedata.ModeTutorial:
		switch c.config.Tutorial.Objective {
		case gamedata.ObjectiveBoss:
			victory = c.world.boss == nil
		case gamedata.ObjectiveBuildBase:
			victory = len(c.world.allColonies) >= 2
		case gamedata.ObjectiveDestroyCreepBases:
			numBases := 0
			for _, creep := range c.world.creeps {
				if creep.stats.kind == creepBase {
					numBases++
				}
			}
			victory = numBases == 0
		case gamedata.ObjectiveAcquireSuperElite:
			for _, colony := range c.world.allColonies {
				superElite := colony.agents.Find(searchFighters|searchWorkers, func(a *colonyAgentNode) bool {
					return a.rank == 2
				})
				if superElite != nil {
					victory = true
					break
				}
			}
		}
	}

	if victory {
		c.victory()
	}
}

func (c *Controller) Update(delta float64) {
	computedDelta := c.nodeRunner.ComputeDelta(delta)

	c.camera.Update(delta)
	c.musicPlayer.Update(delta)
	c.messageManager.Update(delta)
	c.nodeRunner.Update(computedDelta)

	// if c.exitNotice != nil {
	// 	if c.state.MainInput.ActionIsJustPressed(controls.ActionPause) {
	// 		c.nodeRunner.SetPaused(false)
	// 		c.exitNotice.Dispose()
	// 		c.exitNotice = nil
	// 	}
	// 	clickPos, hasClick := c.cursor.ClickPos(controls.ActionClick)
	// 	exitPressed := (hasClick && c.exitButtonRect.Contains(clickPos)) ||
	// 		c.state.MainInput.ActionIsJustPressed(controls.ActionBack)
	// 	if exitPressed {
	// 		c.onExitButtonClicked()
	// 	}
	// 	return
	// }

	if !c.transitionQueued && !c.nodeRunner.IsPaused() {
		c.victoryCheckDelay = gmath.ClampMin(c.victoryCheckDelay-delta, 0)
		if c.victoryCheckDelay == 0 {
			c.victoryCheckDelay = c.scene.Rand().FloatRange(2.0, 3.5)
			c.checkVictory()
		}
	}

	c.handleInput()

	for _, p := range c.world.players {
		p.Update(computedDelta)
	}

	if c.debugInfo != nil {
		c.updateDebug(delta)
	}
}

func (c *Controller) updateDebug(delta float64) {
	c.debugUpdateDelay -= delta
	if c.debugUpdateDelay > 0 {
		return
	}
	c.debugUpdateDelay = 0.5

	settings := &c.state.Persistent.Settings
	switch {
	case settings.ShowFPS && settings.ShowTimer:
		c.debugInfo.Text = fmt.Sprintf("Time: %s FPS: %.0f TPS: %.0f", timeutil.FormatDurationCompact(time.Second*time.Duration(c.nodeRunner.timePlayed)), ebiten.ActualFPS(), ebiten.ActualTPS())
	case settings.ShowFPS:
		c.debugInfo.Text = fmt.Sprintf("FPS: %.0f TPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS())
	case settings.ShowTimer:
		c.debugInfo.Text = fmt.Sprintf("Time: %s", timeutil.FormatDurationCompact(time.Second*time.Duration(c.nodeRunner.timePlayed)))
	}
}

func (c *Controller) IsDisposed() bool { return false }

func (c *Controller) leaveScene(controller ge.SceneController) {
	c.scene.Audio().PauseCurrentMusic()
	c.scene.Context().ChangeScene(controller)
}
