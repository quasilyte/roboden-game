package staging

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"runtime/pprof"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gedraw"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
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

	fogOfWar *ebiten.Image

	musicPlayer *musicPlayer

	exitNotices       []*messageNode
	transitionQueued  bool
	gameFinished      bool
	victoryCheckDelay float64

	camera       *cameraManager
	secondCamera *cameraManager

	tutorialManager *tutorialManager
	messageManagers []*messageManager

	arenaManager *arenaManager
	nodeRunner   *nodeRunner

	debugInfo        *ge.Label
	debugUpdateDelay float64

	replayActions [][]serverapi.PlayerAction

	// rects []*ge.Rect

	EventBeforeLeaveScene gsignal.Event[gsignal.Void]
}

func NewController(state *session.State, config gamedata.LevelConfig, back ge.SceneController) *Controller {
	numScreens := 1
	if config.PlayersMode == serverapi.PmodeTwoPlayers {
		numScreens = 2
	}
	return &Controller{
		state:           state,
		backController:  back,
		config:          config,
		exitNotices:     make([]*messageNode, 0, numScreens),
		messageManagers: make([]*messageManager, 0, numScreens),
	}
}

func (c *Controller) SetReplayActions(actions [][]serverapi.PlayerAction) {
	c.replayActions = actions
}

func (c *Controller) CenterDemoCamera(pos gmath.Vec) {
	c.camera.ToggleCamera(pos)
	c.camera.cinematicSwitchDelay = c.world.localRand.FloatRange(20, 30)
	c.camera.mode = camCinematic
}

func (c *Controller) RenderDemoFrame() *ebiten.Image {
	visible := c.camera.UI.Visible
	c.camera.UI.Visible = false
	img := c.camera.RenderToImage()
	c.camera.UI.Visible = visible
	return img
}

func (c *Controller) IsExcitingDemoFrame() (gmath.Vec, bool) {
	pstate := c.world.players[0].GetState()

	if c.world.boss != nil && c.world.boss.bossStage != 0 {
		for _, creep := range c.world.creeps {
			if creep.stats.Kind == gamedata.CreepServant {
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

		if colony.mode == colonyModeTeleporting {
			return colony.pos, true
		}
		if colony.mode == colonyModeRelocating {
			for _, tp := range c.world.teleporters {
				if tp.pos.DistanceSquaredTo(colony.relocationPoint) < (40 * 40) {
					return tp.pos, true
				}
			}
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
		numCloning := 0
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
			case agentModeMakeClone:
				numCloning++
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
		if numCloning >= 2 {
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

	c.musicPlayer = newMusicPlayer(scene)
	c.musicPlayer.Start()

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
		sessionState:     c.state,
		cameras:          make([]*viewport.Camera, 0, 2),
		stage:            viewport.NewCameraStage(c.config.ExecMode == gamedata.ExecuteSimulation),
		rootScene:        scene,
		nodeRunner:       c.nodeRunner,
		graphicsSettings: c.state.Persistent.Settings.Graphics,
		pathgrid:         pathing.NewGrid(viewportWorld.Width, viewportWorld.Height),
		config:           &c.config,
		gameSettings:     &c.state.Persistent.Settings,
		deviceInfo:       c.state.Device,
		hintsMode:        c.state.Persistent.Settings.HintMode,
		debugLogs:        c.state.Persistent.Settings.DebugLogs,
		droneLabels:      c.state.Persistent.Settings.DebugDroneLabels && c.config.ExecMode == gamedata.ExecuteNormal,
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
	world.inputMode = c.state.GetInput(0).DetectInputMode()
	world.creepCoordinator = newCreepCoordinator(world)
	world.bfs = pathing.NewGreedyBFS(world.pathgrid.Size())
	c.world = world
	world.Init()

	world.EventCheckDefeatState.Connect(c, func(gsignal.Void) {
		c.checkDefeat()
	})

	if c.config.FogOfWar && !c.world.simulation {
		fogOfWar := ebiten.NewImage(int(world.width), int(world.height))
		gedraw.DrawRect(fogOfWar, world.rect, color.RGBA{A: 255})
		c.world.stage.SetFogOfWar(fogOfWar)
		c.fogOfWar = fogOfWar
	}

	// Background generation is an expensive operation.
	// Don't do it inside simulation (headless) mode.
	var bg *ge.TiledBackground
	if c.config.ExecMode != gamedata.ExecuteSimulation {
		// Use local rand for the tileset generation.
		// Otherwise, we'll get incorrect results during the simulation.
		bg = ge.NewTiledBackground(scene.Context())
		bg.LoadTilesetWithRand(scene.Context(), &localRand, viewportWorld.Width, viewportWorld.Height, assets.ImageBackgroundTiles, assets.RawTilesJSON)
	}
	c.world.stage.SetBackground(bg)

	c.camera = c.createCameraManager(viewportWorld, true, c.state.GetInput(0))

	forceCenter := c.config.ExecMode == gamedata.ExecuteReplay ||
		c.config.PlayersMode == serverapi.PmodeTwoBots ||
		c.config.PlayersMode == serverapi.PmodeSingleBot
	if forceCenter {
		c.camera.CenterOn(c.world.rect.Center())
	}

	c.nodeRunner.world = world

	c.nodeRunner.creepCoordinator = world.creepCoordinator

	c.world.EventColonyCreated.Connect(c, func(colony *colonyCoreNode) {
		if c.fogOfWar != nil {
			colony.EventTeleported.Connect(c, func(colony *colonyCoreNode) {
				c.updateFogOfWar(colony.pos)
			})
		}

		if c.config.ExecMode == gamedata.ExecuteNormal && isHumanPlayer(colony.player) {
			colony.EventDestroyed.Connect(c, func(colony *colonyCoreNode) {
				cam := colony.player.GetState().camera
				center := cam.AbsPos(cam.Rect.Center())
				if center.DistanceTo(colony.pos) < 300 {
					return
				}
				c.messageManagers[colony.player.GetState().id].AddMessage(queuedMessageInfo{
					text:          scene.Dict().Get("game.notice.base_destroyed"),
					timer:         8,
					targetPos:     ge.Pos{Offset: *colony.GetPos()},
					forceWorldPos: true,
				})
			})
			colony.EventUnderAttack.Connect(c, func(colony *colonyCoreNode) {
				cam := colony.player.GetState().camera
				center := cam.AbsPos(cam.Rect.Center())
				if center.DistanceTo(colony.pos) < 250 {
					return
				}
				c.messageManagers[colony.player.GetState().id].AddMessage(queuedMessageInfo{
					text:          scene.Dict().Get("game.notice.base_under_attack"),
					trackedObject: colony,
					timer:         8,
					targetPos:     ge.Pos{Base: colony.GetPos()},
				})
			})
		}
	})

	switch c.config.GameMode {
	case gamedata.ModeArena, gamedata.ModeInfArena:
		c.arenaManager = newArenaManager(world)
		c.nodeRunner.AddObject(c.arenaManager)
		c.arenaManager.EventVictory.Connect(c, c.onVictoryTrigger)
	case gamedata.ModeClassic:
		classicManager := newClassicManager(world)
		c.nodeRunner.AddObject(classicManager)
		// TODO: victory trigger should go to the classic manager
	}

	c.createPlayers()

	for _, cam := range c.world.cameras {
		c.messageManagers = append(c.messageManagers, newMessageManager(c.world, cam))
	}

	{
		g := newLevelGenerator(scene, bg, c.world)
		g.Generate()
	}

	for _, p := range c.world.players {
		p.Init()
	}

	for _, cam := range c.world.cameras {
		scene.AddGraphics(cam)
	}

	if c.world.config.GameMode == gamedata.ModeTutorial {
		c.tutorialManager = newTutorialManager(c.state.GetInput(0), c.world, c.messageManagers[0])
		c.nodeRunner.AddObject(c.tutorialManager)
		c.tutorialManager.EventTriggerVictory.Connect(c, c.onVictoryTrigger)
		c.tutorialManager.EventEnableChoices.Connect(c, func(gsignal.Void) {
			c.world.players[0].(*humanPlayer).CreateChoiceWindow(true)
		})
		c.tutorialManager.EventEnableSpecialChoices.Connect(c, func(gsignal.Void) {
			c.world.players[0].(*humanPlayer).EnableSpecialChoices()
		})
		c.tutorialManager.EventForceSpecialChoice.Connect(c, func(kind specialChoiceKind) {
			c.world.players[0].(*humanPlayer).ForceSpecialChoice(kind)
		})
	}

	if c.state.Persistent.Settings.ShowFPS || c.state.Persistent.Settings.ShowTimer {
		if len(c.world.cameras) != 0 {
			c.debugInfo = ge.NewLabel(assets.BitmapFont1)
			c.debugInfo.ColorScale.SetRGBA(0x9d, 0xd7, 0x93, 0xff)
			c.debugInfo.Pos.Offset = gmath.Vec{X: 10, Y: 20}
			c.world.cameras[0].UI.AddGraphics(c.debugInfo)
		}
	}

	c.world.stage.SortBelowLayer()

	if c.fogOfWar != nil {
		for _, colony := range c.world.allColonies {
			c.updateFogOfWar(colony.pos)
		}
	}
}

func (c *Controller) updateFogOfWar(pos gmath.Vec) {
	var options ebiten.DrawImageOptions
	options.CompositeMode = ebiten.CompositeModeDestinationOut
	options.GeoM.Translate(pos.X-colonyVisionRadius, pos.Y-colonyVisionRadius)
	c.fogOfWar.DrawImage(c.world.visionCircle, &options)
}

func (c *Controller) createCameraManager(viewportWorld *viewport.World, main bool, h *gameinput.Handler) *cameraManager {
	cam := c.createCamera(viewportWorld)
	if !main {
		cam.ScreenPos.X = (1920.0 / 2 / 2)
	}
	cm := newCameraManager(c.world, cam)
	if c.config.ExecMode == gamedata.ExecuteDemo {
		cm.InitCinematicMode()
		cm.CenterOn(c.world.rect.Center())
	} else {
		cm.InitManualMode(h)
	}
	c.world.cameras = append(c.world.cameras, cam)
	return cm
}

func (c *Controller) createCamera(viewportWorld *viewport.World) *viewport.Camera {
	// FIXME: hardcoded screen size.
	width := 1920.0 / 2
	height := 1080.0 / 2
	if c.config.PlayersMode == serverapi.PmodeTwoPlayers {
		if c.config.ExecMode != gamedata.ExecuteReplay {
			width /= 2
		}
	}
	cam := viewport.NewCamera(viewportWorld, c.world.stage, width, height)
	return cam
}

func (c *Controller) createPlayers() {
	c.world.players = make([]player, 0, len(c.config.Players))
	hasMouseInput := false
	hasPlayers := false
	isSimulation := c.world.config.ExecMode == gamedata.ExecuteReplay ||
		c.world.config.ExecMode == gamedata.ExecuteSimulation
	for i, pk := range c.config.Players {
		var creepsState *creepsPlayerState
		if i == 0 && c.world.config.GameMode == gamedata.ModeReverse {
			creepsState = newCreepsPlayerState()
			c.world.creepsPlayerState = creepsState
		}
		choiceGen := newChoiceGenerator(c.world, creepsState)
		choiceGen.EventChoiceSelected.Connect(c, c.onChoiceSelected)
		pstate := newPlayerState()
		pstate.id = i

		var p player
		switch pk {
		case gamedata.PlayerHuman:
			hasPlayers = true
			if isSimulation {
				p = newReplayPlayer(c.world, pstate, choiceGen)
				pstate.replay = c.replayActions[i]
			} else {
				playerInput := c.state.GetInput(i)
				if playerInput.HasMouseInput() {
					hasMouseInput = true
				}
				pstate.camera = c.camera
				if i != 0 {
					c.secondCamera = c.createCameraManager(c.camera.World, false, playerInput)
					pstate.camera = c.secondCamera
				}
				cursorRect := pstate.camera.Rect
				cursorRect.Min = cursorRect.Min.Add(pstate.camera.ScreenPos)
				cursorRect.Max = cursorRect.Max.Add(pstate.camera.ScreenPos)
				cursor := gameui.NewCursorNode(playerInput, cursorRect)
				human := newHumanPlayer(humanPlayerConfig{
					world:       c.world,
					state:       pstate,
					input:       playerInput,
					cursor:      cursor,
					choiceGen:   choiceGen,
					creepsState: creepsState,
				})

				if i == 0 {
					human.EventRecipesToggled.Connect(c, func(visible bool) {
						c.world.result.OpenedEvolutionTab = true
						if c.debugInfo != nil {
							c.debugInfo.Visible = !visible
						}
					})
				}
				human.EventPauseRequest.Connect(c, func(gsignal.Void) {
					c.onPausePressed()
				})
				human.EventExitPressed.Connect(c, func(gsignal.Void) {
					c.onExitButtonClicked()
				})
				if human.CanPing() {
					human.EventPing.Connect(c, func(pingPos gmath.Vec) {
						c.scene.Audio().PlaySound(assets.AudioPing)
						dst := c.world.GetPingDst(human)
						c.messageManagers[dst.GetState().id].AddMessage(queuedMessageInfo{
							text:          c.scene.Dict().Get("game.notice.ping"),
							timer:         5,
							targetPos:     ge.Pos{Offset: pingPos},
							forceWorldPos: true,
						})
					})
				}
				p = human
				c.scene.AddObject(cursor)
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

	if c.world.config.ExecMode == gamedata.ExecuteNormal {
		if !hasMouseInput && hasPlayers {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		}
	}
}

func (c *Controller) onVictoryTrigger(gsignal.Void) {
	c.victory()
}

func (c *Controller) onExitButtonClicked() {
	if len(c.exitNotices) != 0 {
		c.leaveScene(c.backController)
		return
	}
	if c.transitionQueued {
		return
	}

	d := c.scene.Dict()
	for _, cam := range c.world.cameras {
		cam.UI.Visible = true
		c.nodeRunner.SetPaused(true)
		exitNotice := newScreenTutorialHintNode(cam, gmath.Vec{}, gmath.Vec{}, d.Get("game.exit.notice", c.world.inputMode))
		c.exitNotices = append(c.exitNotices, exitNotice)
		c.scene.AddObject(exitNotice)
		noticeSize := gmath.Vec{X: exitNotice.width, Y: exitNotice.height}
		noticeCenterPos := cam.Rect.Center().Sub(noticeSize.Mulf(0.5))
		exitNotice.SetPos(noticeCenterPos)
	}
}

func (c *Controller) hasTeleportersAt(pos gmath.Vec) bool {
	for _, tp := range c.world.teleporters {
		if tp.pos.DistanceSquaredTo(pos) < (52 * 52) {
			return true
		}
	}
	return false
}

func (c *Controller) executeAction(choice selectedChoice) bool {
	pstate := choice.Player.GetState()
	selectedColony := pstate.selectedColony

	if c.config.ExecMode == gamedata.ExecuteNormal {
		kind := serverapi.PlayerActionKind(choice.Index + 1)
		if choice.Option.special == specialChoiceMoveColony {
			kind = serverapi.ActionMove
		}
		colonyIndex := -1
		if selectedColony != nil {
			colonyIndex = c.world.GetColonyIndex(selectedColony)
		}
		a := serverapi.PlayerAction{
			Kind:           kind,
			Pos:            [2]float64{choice.Pos.X, choice.Pos.Y},
			SelectedColony: colonyIndex,
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
		case gamedata.HarvesterAgentStats:
			stats = harvesterConstructionStats
		}
		coord := c.world.pathgrid.PosToCoord(selectedColony.pos)
		freeCoord := randIterate(c.world.rand, colonyNearCellOffsets, func(offset pathing.GridCoord) bool {
			probe := coord.Add(offset)
			return c.world.pathgrid.CellIsFree(probe) &&
				!c.hasTeleportersAt(c.world.pathgrid.CoordToPos(probe))
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
			return c.world.CellIsFree2x2(probe) &&
				!c.hasTeleportersAt(p.CoordToPos(probe).Sub(gmath.Vec{X: 16, Y: 16}))
		})
		if !freeCoord.IsZero() {
			pos := p.CoordToPos(coord.Add(freeCoord)).Sub(gmath.Vec{X: 16, Y: 16})
			construction := c.world.NewConstructionNode(choice.Player, pos, gmath.Vec{}, stats)
			c.nodeRunner.AddObject(construction)
			return true
		}
		return false

	case specialSendCreeps:
		c.doSendCreeps()
		return true

	case specialSpawnCrawlers:
		return c.doSpawnCrawlers()

	case specialRally:
		return c.doRally()

	case specialAtomicBomb:
		return c.launchAbomb()

	case specialIncreaseTech:
		c.world.creepsPlayerState.techLevel = gmath.ClampMax(c.world.creepsPlayerState.techLevel+0.1, 2.0)
		return true

	case specialBossAttack:
		return c.doBossAttack()

	default:
		if choice.Option.special > _creepCardFirst && choice.Option.special < _creepCardLast {
			info := creepOptionInfoList[creepCardID(choice.Option.special)]
			return c.world.creepsPlayerState.AddUnits(c.world, choice.Option.direction, info)
		}

		panic("unexpected action ID")
	}
}

func (c *Controller) launchAbomb() bool {
	if len(c.world.allColonies) == 0 {
		return false
	}
	if c.world.boss == nil {
		return false
	}
	target := gmath.RandElem(c.world.rand, c.world.allColonies)
	abomb := c.world.newProjectileNode(projectileConfig{
		World:      c.world,
		Weapon:     gamedata.AtomicBombWeapon,
		Attacker:   c.world.boss,
		ToPos:      target.pos.Add(c.scene.Rand().Offset(-10, 10)).Add(gmath.Vec{Y: 16}),
		Target:     target,
		FireOffset: gmath.Vec{Y: -20},
	})
	c.nodeRunner.AddProjectile(abomb)
	return true
}

func (c *Controller) doRally() bool {
	if c.world.boss == nil {
		return false
	}
	c.scene.Audio().PlaySound(assets.AudioWaveStart)
	c.world.creepCoordinator.Rally(c.world.boss.pos)
	return true
}

func (c *Controller) doSpawnCrawlers() bool {
	if c.world.boss == nil {
		return false
	}
	c.world.boss.specialDelay = 0
	return true
}

func (c *Controller) doBossAttack() bool {
	if c.world.boss == nil {
		return false
	}
	var closestColony *colonyCoreNode
	closestDist := math.MaxFloat64
	for _, colony := range c.world.allColonies {
		dist := colony.pos.DistanceTo(c.world.boss.pos)
		if dist < closestDist {
			closestDist = dist
			closestColony = colony
		}
	}
	if closestColony == nil {
		return false
	}
	dir := closestColony.pos.Add(c.world.rand.Offset(-60, 60)).DirectionTo(c.world.boss.pos)
	targetPos := dir.Mulf(c.world.rand.FloatRange(200, 400)).Add(c.world.boss.pos)
	c.world.boss.waypoint = targetPos
	return true
}

func (c *Controller) doSendCreeps() {
	for dir := range c.world.creepsPlayerState.attackSides {
		cg := c.world.creepsPlayerState.attackSides[dir]
		for i := range cg.groups {
			g := cg.groups[i]
			if len(g.units) == 0 {
				continue
			}
			sendCreeps(c.world, g)
		}
	}

	c.world.creepsPlayerState.ResetGroups()
	c.world.creepsPlayerState.RecalcMaxCost()
}

func (c *Controller) playPlayerSound(p player, sound resource.AudioID) {
	if isHumanPlayer(p) {
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

func (c *Controller) prepareBattleResults() {
	if c.config.ExecMode == gamedata.ExecuteNormal {
		c.world.result.Replay = make([][]serverapi.PlayerAction, len(c.world.players))
		for i, p := range c.world.players {
			c.world.result.Replay[i] = p.GetState().replay
		}
	}

	c.world.result.BossDefeated = c.world.boss == nil

	c.world.result.Ticks = c.nodeRunner.ticks
	c.world.result.TimePlayed = time.Second * time.Duration(c.nodeRunner.timePlayed)
	if c.arenaManager != nil {
		c.world.result.ArenaLevel = c.arenaManager.level
		if c.config.GameMode == gamedata.ModeArena {
			for _, creep := range c.world.creeps {
				if creep.stats == gamedata.DominatorCreepStats {
					c.world.result.DominatorsSurvived++
				}
			}
		}
	}
	c.world.result.Score = calcScore(c.world)
	c.world.result.DifficultyScore = c.config.DifficultyScore
	c.world.result.DronePointsAllocated = c.config.DronePointsAllocated
}

func (c *Controller) defeat() {
	if c.transitionQueued {
		return
	}

	c.transitionQueued = true

	c.prepareBattleResults()
	c.world.result.Victory = false
	c.scene.DelayedCall(2.0, func() {
		c.gameFinished = true
		switch c.config.ExecMode {
		case gamedata.ExecuteNormal:
			c.leaveScene(newResultsController(c.state, &c.config, c.backController, c.world.result, nil))
		case gamedata.ExecuteDemo, gamedata.ExecuteReplay:
			c.leaveScene(c.backController)
		}
	})
}

func (c *Controller) victory() {
	if c.transitionQueued {
		return
	}

	c.transitionQueued = true

	c.scene.Audio().PlaySound(assets.AudioVictory)
	c.prepareBattleResults()
	c.world.result.Victory = true
	c.scene.DelayedCall(5.0, func() {
		c.gameFinished = true
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
		case gamedata.ExecuteDemo, gamedata.ExecuteReplay:
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

func (c *Controller) sharedActionIsJustPressed(a input.Action) bool {
	if c.state.GetInput(0).ActionIsJustPressed(a) {
		return true
	}
	if c.world.config.PlayersMode == serverapi.PmodeTwoPlayers {
		if c.state.GetInput(1).ActionIsJustPressed(a) {
			return true
		}
	}
	return false
}

func (c *Controller) handleInput() {
	switch c.config.ExecMode {
	case gamedata.ExecuteSimulation:
		c.handleReplayActions()
		return
	case gamedata.ExecuteDemo:
		c.handleDemoInput()
		return
	case gamedata.ExecuteReplay:
		c.handleReplayActions()
		// And then do a some more common stuff and return from the function.
	}

	if c.sharedActionIsJustPressed(controls.ActionPause) {
		c.onPausePressed()
		return
	}

	c.camera.HandleInput()
	if c.secondCamera != nil {
		c.secondCamera.HandleInput()
	}

	if c.sharedActionIsJustPressed(controls.ActionBack) {
		c.onExitButtonClicked()
		return
	}

	if c.config.ExecMode == gamedata.ExecuteReplay {
		return
	}

	for _, p := range c.world.players {
		p.HandleInput()
	}

	if c.tutorialManager != nil {
		// Tutorial is a single player only experience.
		if c.state.GetInput(0).ActionIsJustPressed(controls.ActionNextTutorialMessage) {
			c.tutorialManager.OnNextPressed()
		}
	}
}

func (c *Controller) isDefeatState() bool {
	switch c.config.GameMode {
	case gamedata.ModeTutorial, gamedata.ModeClassic, gamedata.ModeArena, gamedata.ModeInfArena:
		for _, p := range c.world.players {
			if len(p.GetState().colonies) == 0 {
				return true
			}
		}

	case gamedata.ModeReverse:
		if c.config.PlayersMode == serverapi.PmodeTwoPlayers {
			colonyPlayer := c.world.players[1]
			if len(colonyPlayer.GetState().colonies) == 0 {
				return true
			}
		}
		if c.world.boss == nil {
			return true
		}
	}

	return false
}

func (c *Controller) checkDefeat() {
	if c.transitionQueued {
		return
	}

	if c.isDefeatState() {
		c.defeat()
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

	case gamedata.ModeArena, gamedata.ModeTutorial:
		// Do nothing. This mode is ended with a trigger.

	case gamedata.ModeInfArena:
		// Do nothing. This mode is endless.

	case gamedata.ModeReverse:
		// In two players mode, the only way to finish a match
		// is to trigger a defeat to either players.
		if c.config.PlayersMode == serverapi.PmodeSinglePlayer {
			colonyPlayer := c.world.players[1]
			victory = len(colonyPlayer.GetState().colonies) == 0
		}
	}

	if victory {
		c.victory()
	}
}

func (c *Controller) onPausePressed() {
	if len(c.exitNotices) != 0 {
		c.nodeRunner.SetPaused(false)
		for _, n := range c.exitNotices {
			n.Dispose()
		}
		c.exitNotices = c.exitNotices[:0]
		return
	}
	c.nodeRunner.SetPaused(!c.nodeRunner.IsPaused())
}

func (c *Controller) GetSessionState() *session.State {
	return c.state
}

func (c *Controller) Update(delta float64) {
	computedDelta := c.nodeRunner.ComputeDelta(delta)

	c.world.stage.Update()
	c.camera.Update(delta)
	if c.secondCamera != nil {
		c.secondCamera.Update(delta)
	}
	for _, mm := range c.messageManagers {
		mm.Update(delta)
	}
	c.musicPlayer.Update(delta)
	c.nodeRunner.Update(computedDelta)

	if !c.nodeRunner.IsPaused() {
		if c.fogOfWar != nil {
			for _, colony := range c.world.allColonies {
				if !colony.IsFlying() {
					continue
				}
				c.updateFogOfWar(colony.pos)
			}
		}

		if !c.transitionQueued {
			c.victoryCheckDelay = gmath.ClampMin(c.victoryCheckDelay-delta, 0)
			if c.victoryCheckDelay == 0 {
				c.victoryCheckDelay = c.scene.Rand().FloatRange(2.0, 3.5)
				c.checkVictory()
			}
		}
	}

	c.handleInput()

	for _, p := range c.world.players {
		p.Update(computedDelta, delta)
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
	c.EventBeforeLeaveScene.Emit(gsignal.Void{})

	if !c.world.simulation {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}

	c.scene.Audio().PauseCurrentMusic()
	c.scene.Context().ChangeScene(controller)
}
