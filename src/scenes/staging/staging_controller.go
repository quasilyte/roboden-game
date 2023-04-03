package staging

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gedraw"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/viewport"
)

type Controller struct {
	state *session.State

	backController       ge.SceneController
	cameraPanDragPos     gmath.Vec
	cameraPanSpeed       float64
	cameraPanBoundary    float64
	cameraToggleProgress float64
	cameraToggleTarget   gmath.Vec

	startTime time.Time

	colonySelector   *ge.Sprite
	radar            *radarNode
	rpanel           *rpanelNode
	menuButtonRect   gmath.Rect
	toggleButtonRect gmath.Rect

	scene  *ge.Scene
	world  *worldState
	config session.LevelConfig

	choices *choiceWindowNode

	musicPlayer *musicPlayer

	tier3spawnDelay float64
	tier3spawnRate  float64

	transitionQueued  bool
	victoryCheckDelay float64

	camera *viewport.Camera

	tutorialManager *tutorialManager

	visionRadius float64
	fogOfWar     *ebiten.Image
	visionCircle *ebiten.Image

	arenaManager *arenaManager

	cursor *cursorNode

	debugInfo *ge.Label
}

func NewController(state *session.State, config session.LevelConfig, back ge.SceneController) *Controller {
	return &Controller{
		state:          state,
		backController: back,
		config:         config,
	}
}

func (c *Controller) Init(scene *ge.Scene) {
	c.startTime = time.Now()

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

	if c.state.Persistent.Settings.EdgeScrollRange != 0 {
		c.cameraPanBoundary = 1
		if runtime.GOARCH == "wasm" {
			c.cameraPanBoundary = 8
		}
		c.cameraPanBoundary += 2 * float64(c.state.Persistent.Settings.EdgeScrollRange-1)
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
	c.camera = viewport.NewCamera(viewportWorld, 1920/2, 1080/2)

	// Start launching tier3 creeps after ~15 minutes.
	c.tier3spawnDelay = scene.Rand().FloatRange(14*60.0, 16*60.0)
	c.tier3spawnRate = 1.0

	if c.state.Device.IsMobile {
		switch c.state.Persistent.Settings.ScrollingSpeed {
		case 0:
			c.cameraPanSpeed = 0.5
		case 1:
			c.cameraPanSpeed = 0.8
		case 2:
			// The default speed, x1 factor.
			// This is the most pleasant and convenient to use, but could
			// be too slow for a pro player.
			c.cameraPanSpeed = 1
		case 3:
			// Just a bit faster.
			c.cameraPanSpeed = 1.2
		case 4:
			c.cameraPanSpeed = 2
		}
	} else {
		c.cameraPanSpeed = float64(c.state.Persistent.Settings.ScrollingSpeed+1) * 4
	}

	world := &worldState{
		graphicsSettings: c.state.Persistent.Settings.Graphics,
		debug:            c.state.Persistent.Settings.Debug,
		pathgrid:         pathing.NewGrid(viewportWorld.Width, viewportWorld.Height),
		config:           &c.config,
		camera:           c.camera,
		rand:             scene.Rand(),
		tmpTargetSlice:   make([]projectileTarget, 0, 20),
		tmpColonySlice:   make([]*colonyCoreNode, 0, 4),
		width:            viewportWorld.Width,
		height:           viewportWorld.Height,
		rect: gmath.Rect{
			Max: gmath.Vec{
				X: viewportWorld.Width,
				Y: viewportWorld.Height,
			},
		},
		tier2recipes: c.config.Tier2Recipes,
	}
	world.creepCoordinator = newCreepCoordinator(world)
	world.bfs = pathing.NewGreedyBFS(world.pathgrid.Size())
	c.world = world
	world.Init()

	if c.config.GameMode == gamedata.ModeArena {
		c.arenaManager = newArenaManager(world)
		scene.AddObject(c.arenaManager)
		c.arenaManager.EventVictory.Connect(c, c.onVictoryTrigger)
	}

	bg := ge.NewTiledBackground(scene.Context())
	bg.LoadTileset(scene.Context(), world.width, world.height, assets.ImageBackgroundTiles, assets.RawTilesJSON)
	c.camera.SetBackground(bg)
	g := newLevelGenerator(scene, c.world)
	g.Generate()

	c.colonySelector = scene.NewSprite(assets.ImageColonyCoreSelector)
	c.camera.AddSpriteBelow(c.colonySelector)

	c.cursor = newCursorNode(c.state.MainInput, c.camera.Rect)

	if c.config.EnemyBoss {
		c.radar = newRadarNode(c.world)
		scene.AddObject(c.radar)

		buttonSize := gmath.Vec{X: 32, Y: 36}
		toggleButtonOffset := gmath.Vec{X: 155, Y: 491}
		c.toggleButtonRect = gmath.Rect{Min: toggleButtonOffset, Max: toggleButtonOffset.Add(buttonSize)}

		menuButtonOffset := gmath.Vec{X: 211, Y: 491}
		c.menuButtonRect = gmath.Rect{Min: menuButtonOffset, Max: menuButtonOffset.Add(buttonSize)}
	}

	if c.config.ExtraUI {
		c.rpanel = newRpanelNode(c.world)
		scene.AddObject(c.rpanel)
	}

	scene.AddObject(c.cursor)

	choicesPos := gmath.Vec{
		X: 960 - 232 - 16,
		Y: 540 - 200 - 16,
	}
	c.choices = newChoiceWindowNode(choicesPos, c.world, c.state.MainInput, c.cursor)
	c.choices.EventChoiceSelected.Connect(nil, c.onChoiceSelected)

	c.selectNextColony(true)
	c.camera.CenterOn(c.world.selectedColony.pos)

	scene.AddGraphics(c.camera)

	if c.state.Persistent.Settings.Debug {
		c.debugInfo = scene.NewLabel(assets.FontSmall)
		c.debugInfo.ColorScale.SetColor(ge.RGB(0xffffff))
		c.debugInfo.Pos.Offset = gmath.Vec{X: 10, Y: 10}
		scene.AddGraphicsAbove(c.debugInfo, 1)
	}

	if c.world.IsTutorial() {
		c.tutorialManager = newTutorialManager(c.state.MainInput, c.world)
		scene.AddObject(c.tutorialManager)
		if c.rpanel != nil {
			c.tutorialManager.EventRequestPanelUpdate.Connect(c, c.onPanelUpdateRequested)
		}
		c.tutorialManager.EventTriggerVictory.Connect(c, c.onVictoryTrigger)
	}

	scene.AddObject(c.choices)

	if c.config.FogOfWar {
		c.visionRadius = 500.0

		c.fogOfWar = ebiten.NewImage(int(c.world.width), int(c.world.height))
		gedraw.DrawRect(c.fogOfWar, c.world.rect, color.RGBA{A: 255})
		c.camera.SetFogOfWar(c.fogOfWar)

		c.visionCircle = ebiten.NewImage(int(c.visionRadius*2), int(c.visionRadius*2))
		gedraw.DrawCircle(c.visionCircle, gmath.Vec{X: c.visionRadius, Y: c.visionRadius}, c.visionRadius, color.RGBA{A: 255})

		c.updateFogOfWar(c.world.selectedColony.pos)
	}

	c.camera.SortBelowLayer()

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
	// 		}
	// 	}
	// }
}

func (c *Controller) updateFogOfWar(pos gmath.Vec) {
	var options ebiten.DrawImageOptions
	options.CompositeMode = ebiten.CompositeModeDestinationOut
	options.GeoM.Translate(pos.X-c.visionRadius, pos.Y-c.visionRadius)
	c.fogOfWar.DrawImage(c.visionCircle, &options)
}

func (c *Controller) onPanelUpdateRequested(gsignal.Void) {
	c.rpanel.UpdateMetrics()
}

func (c *Controller) onVictoryTrigger(gsignal.Void) {
	c.victory()
}

func (c *Controller) onMenuButtonClicked() {
	c.leaveScene(c.backController)
}

func (c *Controller) onToggleButtonClicked() {
	c.selectNextColony(true)
}

func (c *Controller) onChoiceSelected(choice selectedChoice) {
	if c.tutorialManager != nil {
		c.tutorialManager.OnChoice(choice)
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
		c.world.selectedColony.factionWeights.AddWeight(choice.Faction, c.world.rand.FloatRange(0.1, 0.2))
		for _, e := range choice.Option.effects {
			// Use priorities.AddWeight directly here to avoid the signal.
			// We'll call UpdateMetrics() below ourselves.
			c.world.selectedColony.priorities.AddWeight(e.priority, e.value)
		}
		if c.rpanel != nil {
			c.rpanel.UpdateMetrics()
		}
		return
	}

	var relocationPos gmath.Vec
	switch choice.Option.special {
	case specialAttack:
		c.launchAttack()
	case specialChoiceMoveColony:
		maxDist := c.world.selectedColony.MaxFlyDistance() * c.world.rand.FloatRange(0.9, 1.1)
		clickPos := choice.Pos
		clickDist := c.world.selectedColony.pos.DistanceTo(clickPos)
		dist := gmath.ClampMax(clickDist, maxDist)
		relocationVec := c.world.selectedColony.pos.VecTowards(clickPos, 1).Mulf(dist)
		relocationPos = relocationVec.Add(c.world.selectedColony.pos)
		c.launchRelocation(c.world.selectedColony, dist, maxDist, relocationPos)
	case specialIncreaseRadius:
		c.world.result.RadiusIncreases++
		c.world.selectedColony.realRadius += c.world.rand.FloatRange(16, 32)
		c.world.selectedColony.realRadiusSqr = c.world.selectedColony.realRadius * c.world.selectedColony.realRadius
	case specialDecreaseRadius:
		value := c.world.rand.FloatRange(30, 40)
		c.world.selectedColony.realRadius = gmath.ClampMin(c.world.selectedColony.realRadius-value, 96)
		c.world.selectedColony.realRadiusSqr = c.world.selectedColony.realRadius * c.world.selectedColony.realRadius
	case specialBuildColony, specialBuildGunpoint:
		// TODO: use a pathing.Grid to find a free cell?
		stats := colonyCoreConstructionStats
		dist := 60.0
		size := 40.0
		if choice.Option.special == specialBuildGunpoint {
			stats = gunpointConstructionStats
			dist = 48.0
			size = 32.0
		} else {
			c.world.result.ColoniesBuilt++
		}
		direction := c.world.rand.Rad()
		for i := 0; i < 18; i++ {
			locationProbe := gmath.RadToVec(direction).Mulf(dist).Add(c.world.selectedColony.pos)
			direction += (2 * math.Pi) / 17
			constructionPos := c.pickColonyPos(nil, locationProbe, size, 4)
			if !constructionPos.IsZero() {
				construction := c.world.NewConstructionNode(constructionPos, stats)
				c.scene.AddObject(construction)
				break
			}
		}
	}
}

func (c *Controller) pickColonyPos(core *colonyCoreNode, pos gmath.Vec, r float64, tries int) gmath.Vec {
	pos = correctedPos(c.world.rect, pos, 0)
	minOffset := -10.0
	maxOffset := 10.0
	for i := 0; i < tries; i++ {
		probe := pos.Add(c.world.rand.Offset(minOffset, maxOffset))
		probe = roundedPos(probe)
		probe = correctedPos(c.world.rect, probe, 98)
		if posIsFree(c.world, core, probe, r) {
			return probe
		}
		minOffset -= 10
		maxOffset += 10
	}
	return gmath.Vec{}
}

func (c *Controller) launchAttack() {
	if c.world.selectedColony.agents.NumAvailableFighters() == 0 {
		return
	}
	closeTargets := c.world.tmpTargetSlice[:0]
	maxDist := gmath.ClampMin(c.world.selectedColony.PatrolRadius()*1.85, 320)
	maxDist *= c.world.rand.FloatRange(0.95, 1.2)
	for _, creep := range c.world.creeps {
		if len(closeTargets) >= 5 {
			break
		}
		if creep.IsCloaked() {
			continue
		}
		if creep.pos.DistanceTo(c.world.selectedColony.pos) > maxDist {
			continue
		}
		closeTargets = append(closeTargets, creep)
	}
	if len(closeTargets) == 0 {
		return
	}
	maxDispatched := gmath.Clamp(int(float64(c.world.selectedColony.agents.NumAvailableFighters())*0.6), 1, 15)
	c.world.selectedColony.agents.Find(searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
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

func (c *Controller) launchRelocation(core *colonyCoreNode, dist, maxDist float64, dst gmath.Vec) {
	dstDir := dst.DirectionTo(core.pos)
	var relocationPoint gmath.Vec
OuterLoop:
	for _, step := range [3]float64{-32.0, 32.0, -16.0} {
		currentDist := dist
		currentPos := dst
		for {
			if posIsFree(c.world, core, currentPos, 48) {
				relocationPoint = currentPos
				break OuterLoop
			}
			leftPos := dstDir.Rotated(-0.2).Mulf(currentDist).Add(core.pos)
			if posIsFree(c.world, core, leftPos, 48) {
				relocationPoint = leftPos
				break OuterLoop
			}
			rightPos := dstDir.Rotated(0.2).Mulf(currentDist).Add(core.pos)
			if posIsFree(c.world, core, rightPos, 48) {
				relocationPoint = rightPos
				break OuterLoop
			}
			currentDist += step
			if currentDist < 0 || currentDist > maxDist || currentPos.DistanceSquaredTo(core.pos) < 32 {
				break
			}
			currentPos = dstDir.Mulf(currentDist).Add(core.pos)
		}
	}
	if !relocationPoint.IsZero() {
		core.doRelocation(roundedPos(relocationPoint))
	} else {
		c.scene.Audio().PlaySound(assets.AudioError)
	}
}

func (c *Controller) spawnTier3Creep() {
	// TODO: move to a creep coordinator?

	c.tier3spawnRate = gmath.ClampMin(c.tier3spawnRate-0.025, 0.4)
	c.tier3spawnDelay = c.scene.Rand().FloatRange(55, 80) * c.tier3spawnRate

	var spawnPos gmath.Vec
	roll := c.scene.Rand().Float()
	if roll < 0.25 {
		spawnPos.X = c.world.width - 4
		spawnPos.Y = c.scene.Rand().FloatRange(0, c.world.height)
	} else if roll < 0.5 {
		spawnPos.X = c.scene.Rand().FloatRange(0, c.world.width)
		spawnPos.Y = c.world.height - 4
	} else if roll < 0.75 {
		spawnPos.X = 4
		spawnPos.Y = c.scene.Rand().FloatRange(0, c.world.height)
	} else {
		spawnPos.X = c.scene.Rand().FloatRange(0, c.world.width)
		spawnPos.Y = 4
	}
	spawnPos = roundedPos(spawnPos)
	creep := c.world.NewCreepNode(spawnPos, assaultCreepStats)
	c.scene.AddObject(creep)
}

func (c *Controller) defeat() {
	if c.transitionQueued {
		return
	}

	c.transitionQueued = true
	c.scene.DelayedCall(2.0, func() {
		c.world.result.Victory = false
		c.prepareBattleResults()
		c.leaveScene(newResultsController(c.state, &c.config, c.backController, c.world.result))
	})
}

func (c *Controller) prepareBattleResults() {
	c.world.result.TimePlayed = time.Since(c.startTime)
	if c.arenaManager != nil {
		c.world.result.ArenaLevel = c.arenaManager.level
		if !c.config.InfiniteMode {
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

	c.scene.DelayedCall(5.0, func() {
		c.world.result.Victory = true
		c.prepareBattleResults()

		t3set := map[gamedata.ColonyAgentKind]struct{}{}
		for _, colony := range c.world.colonies {
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

		c.leaveScene(newResultsController(c.state, &c.config, c.backController, c.world.result))
	})
}

func (c *Controller) handleInput() {
	mainInput := c.state.MainInput
	if !c.state.Device.IsMobile {
		// Camera panning only makes sense on non-mobile devices
		// where we have a keyboard/gamepad or a cursor.
		var cameraPan gmath.Vec
		if mainInput.ActionIsPressed(controls.ActionPanRight) {
			cameraPan.X += c.cameraPanSpeed
		}
		if mainInput.ActionIsPressed(controls.ActionPanDown) {
			cameraPan.Y += c.cameraPanSpeed
		}
		if mainInput.ActionIsPressed(controls.ActionPanLeft) {
			cameraPan.X -= c.cameraPanSpeed
		}
		if mainInput.ActionIsPressed(controls.ActionPanUp) {
			cameraPan.Y -= c.cameraPanSpeed
		}
		if cameraPan.IsZero() {
			if info, ok := mainInput.PressedActionInfo(controls.ActionPanAlt); ok {
				cameraCenter := c.camera.Rect.Center()
				cameraPan = gmath.RadToVec(cameraCenter.AngleToPoint(info.Pos)).Mulf(c.cameraPanSpeed * 0.8)
			}
		}
		if cameraPan.IsZero() && c.cameraPanBoundary != 0 {
			// Mouse cursor can pan the camera too.
			cursor := mainInput.CursorPos()
			if cursor.X > c.camera.Rect.Width()-c.cameraPanBoundary {
				cameraPan.X += c.cameraPanSpeed
			}
			if cursor.Y > c.camera.Rect.Height()-c.cameraPanBoundary {
				cameraPan.Y += c.cameraPanSpeed
			}
			if cursor.X < c.cameraPanBoundary {
				cameraPan.X -= c.cameraPanSpeed
			}
			if cursor.Y < c.cameraPanBoundary {
				cameraPan.Y -= c.cameraPanSpeed
			}
		}
		c.camera.Pan(cameraPan)
	} else {
		// On mobile devices we expect a touch screen support.
		// Instead of panning, we use dragging here.
		if mainInput.ActionIsJustPressed(controls.ActionPanDrag) {
			c.cameraPanDragPos = c.camera.Offset
		}
		if info, ok := mainInput.PressedActionInfo(controls.ActionPanDrag); ok {
			posDelta := info.StartPos.Sub(info.Pos).Mulf(c.cameraPanSpeed)
			newPos := c.cameraPanDragPos.Add(posDelta)
			c.camera.SetOffset(newPos)
		}
	}

	if mainInput.ActionIsJustPressed(controls.ActionBack) {
		c.onMenuButtonClicked()
		return
	}

	if mainInput.ActionIsJustPressed(controls.ActionToggleColony) {
		c.onToggleButtonClicked()
		return
	}

	handledClick := false
	clickPos, hasClick := c.cursor.ClickPos(controls.ActionClick)
	if len(c.world.colonies) > 1 {
		if hasClick {
			clickPos := clickPos.Add(c.camera.Offset)
			selectDist := 40.0
			if c.state.Device.IsMobile {
				selectDist = 80.0
			}
			var closestColony *colonyCoreNode
			closestDist := math.MaxFloat64
			for _, colony := range c.world.colonies {
				if colony == c.world.selectedColony {
					continue
				}
				dist := colony.pos.DistanceTo(clickPos)
				if dist > selectDist {
					continue
				}
				if dist < closestDist {
					closestColony = colony
					closestDist = dist
				}
			}
			if closestColony != nil {
				c.selectColony(closestColony)
				handledClick = true
			}
		}
	}
	if handledClick {
		return
	}
	if c.menuButtonRect.Contains(clickPos) {
		c.onMenuButtonClicked()
		return
	}
	if c.toggleButtonRect.Contains(clickPos) {
		c.onToggleButtonClicked()
		return
	}
	c.choices.HandleInput()
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
			victory = len(c.world.colonies) >= 2
		case gamedata.ObjectiveDestroyCreepBases:
			numBases := 0
			for _, creep := range c.world.creeps {
				if creep.stats.kind == creepBase {
					numBases++
				}
			}
			victory = numBases == 0
		case gamedata.ObjectiveAcquireSuperElite:
			for _, colony := range c.world.colonies {
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
	c.musicPlayer.Update(delta)
	c.world.Update(delta)

	if c.config.FogOfWar {
		for _, colony := range c.world.colonies {
			if !colony.IsFlying() {
				continue
			}
			c.updateFogOfWar(colony.spritePos)
		}
	}

	if !c.cameraToggleTarget.IsZero() {
		c.cameraToggleProgress = gmath.ClampMax(c.cameraToggleProgress+delta, 1)
		c.camera.CenterOn(c.camera.CenterPos().LinearInterpolate(c.cameraToggleTarget, c.cameraToggleProgress))
		if c.cameraToggleProgress >= 0.9 || c.camera.CenterPos().DistanceSquaredTo(c.cameraToggleTarget) < (80*80) {
			c.camera.CenterOn(c.cameraToggleTarget)
			c.cameraToggleTarget = gmath.Vec{}
		}
	}

	if !c.transitionQueued {
		c.victoryCheckDelay = gmath.ClampMin(c.victoryCheckDelay-delta, 0)
		if c.victoryCheckDelay == 0 {
			c.victoryCheckDelay = c.scene.Rand().FloatRange(2.0, 3.5)
			c.checkVictory()
		}
	}

	c.choices.Enabled = c.world.selectedColony != nil &&
		c.world.selectedColony.mode == colonyModeNormal

	if c.config.GameMode == gamedata.ModeClassic {
		// TODO: move somewhere else?
		c.tier3spawnDelay = gmath.ClampMin(c.tier3spawnDelay-delta, 0)
		if c.tier3spawnDelay == 0 {
			c.spawnTier3Creep()
		}
	}

	c.handleInput()

	if c.debugInfo != nil {
		colony := c.world.selectedColony
		numDrones := 0
		droneLimit := 0
		if colony != nil {
			numDrones = colony.NumAgents()
			droneLimit = colony.calcUnitLimit()
		}
		c.debugInfo.Text = fmt.Sprintf("FPS: %.0f Drones: %d/%d",
			ebiten.ActualFPS(),
			numDrones, droneLimit)
	}

}

func (c *Controller) IsDisposed() bool { return false }

func (c *Controller) leaveScene(controller ge.SceneController) {
	c.scene.Audio().PauseCurrentMusic()
	c.scene.Context().ChangeScene(controller)
}

func (c *Controller) selectColony(colony *colonyCoreNode) {
	if c.world.selectedColony == colony {
		return
	}
	if c.world.selectedColony != nil {
		c.scene.Audio().PlaySound(assets.AudioBaseSelect)
		c.world.selectedColony.EventDestroyed.Disconnect(c)
		if c.rpanel != nil {
			c.world.selectedColony.EventPrioritiesChanged.Disconnect(c)
		}
	}
	c.world.selectedColony = colony
	c.choices.selectedColony = colony
	if c.radar != nil {
		c.radar.SetBase(c.world.selectedColony)
	}
	if c.rpanel != nil {
		c.rpanel.SetBase(c.world.selectedColony)
		c.rpanel.UpdateMetrics()
	}
	if c.world.selectedColony == nil {
		c.colonySelector.Visible = false
		c.defeat()
		return
	}
	c.world.selectedColony.EventDestroyed.Connect(c, func(_ *colonyCoreNode) {
		c.selectNextColony(false)
	})
	if c.rpanel != nil {
		c.world.selectedColony.EventPrioritiesChanged.Connect(c, func(_ *colonyCoreNode) {
			c.rpanel.UpdateMetrics()
		})
	}
	c.colonySelector.Pos.Base = &c.world.selectedColony.spritePos
}

func (c *Controller) selectNextColony(center bool) {
	colony := c.findNextColony()
	c.selectColony(colony)
	if center && c.world.selectedColony != nil {
		c.cameraToggleTarget = c.world.selectedColony.pos
		c.cameraToggleProgress = 0
	}
}

func (c *Controller) findNextColony() *colonyCoreNode {
	if len(c.world.colonies) == 0 {
		return nil
	}
	if len(c.world.colonies) == 1 {
		return c.world.colonies[0]
	}
	index := xslices.Index(c.world.colonies, c.world.selectedColony)
	if index == len(c.world.colonies)-1 {
		index = 0
	} else {
		index++
	}
	return c.world.colonies[index]
}
