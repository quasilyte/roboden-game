package staging

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/viewport"
)

type Controller struct {
	state *session.State

	backController    ge.SceneController
	cameraPanDragPos  gmath.Vec
	cameraPanSpeed    float64
	cameraPanBoundary float64

	startTime time.Time

	selectedColony *colonyCoreNode
	colonySelector *ge.Sprite
	radar          *radarNode
	menuButton     *gameui.TextureButton
	toggleButton   *gameui.TextureButton

	scene     *ge.Scene
	world     *worldState
	worldSize int

	choices *choiceWindowNode

	musicPlayer *musicPlayer

	tier3spawnDelay float64
	tier3spawnRate  float64

	transitionQueued bool

	camera *viewport.Camera

	cursor *cursorNode

	debugInfo *ge.Label
}

func NewController(state *session.State, worldSize int, back ge.SceneController) *Controller {
	return &Controller{
		state:          state,
		backController: back,
		worldSize:      worldSize,
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
	switch c.worldSize {
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
		worldSize:        c.worldSize,
		pathgrid:         pathing.NewGrid(viewportWorld.Width, viewportWorld.Height),
		options:          &c.state.LevelOptions,
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
		tier2recipes: c.state.LevelOptions.Tier2Recipes,
	}
	world.creepCoordinator = newCreepCoordinator(world)
	world.bfs = pathing.NewGreedyBFS(world.pathgrid.Size())
	c.world = world
	world.Init()

	bg := ge.NewTiledBackground(scene.Context())
	bg.LoadTileset(scene.Context(), world.width, world.height, assets.ImageBackgroundTiles, assets.RawTilesJSON)
	c.camera.SetBackground(bg)
	g := newLevelGenerator(scene, c.world)
	g.Generate()

	c.colonySelector = scene.NewSprite(assets.ImageColonyCoreSelector)
	c.camera.AddSpriteBelow(c.colonySelector)

	c.cursor = newCursorNode(c.state.MainInput, c.camera.Rect)

	c.menuButton = gameui.NewTextureButton(ge.Pos{Offset: gmath.Vec{X: 76, Y: 12}}, assets.ImageButtonMenu, c.cursor)
	c.menuButton.EventClicked.Connect(c, c.onMenuButtonClicked)
	scene.AddObject(c.menuButton)

	c.toggleButton = gameui.NewTextureButton(ge.Pos{Offset: gmath.Vec{X: 12, Y: 76}}, assets.ImageButtonBaseToggle, c.cursor)
	c.toggleButton.EventClicked.Connect(c, c.onToggleButtonClicked)
	scene.AddObject(c.toggleButton)

	c.radar = newRadarNode(c.world)
	scene.AddObject(c.radar)

	scene.AddObject(c.cursor)

	choicesPos := gmath.Vec{
		X: 960 - 232 - 16,
		Y: 540 - 200 - 16,
	}
	c.choices = newChoiceWindowNode(choicesPos, c.state.MainInput, c.cursor)
	c.choices.EventChoiceSelected.Connect(nil, c.onChoiceSelected)

	c.selectNextColony(true)
	c.camera.CenterOn(c.selectedColony.pos)

	scene.AddGraphics(c.camera)

	if c.state.Persistent.Settings.Debug {
		c.debugInfo = scene.NewLabel(assets.FontSmall)
		c.debugInfo.ColorScale.SetColor(ge.RGB(0xffffff))
		c.debugInfo.Pos.Offset = gmath.Vec{X: 10, Y: 10}
		scene.AddGraphicsAbove(c.debugInfo, 1)
	}

	if c.state.LevelOptions.Tutorial {
		tutorial := newTutorialManager(c.state.MainInput, c.choices)
		scene.AddObject(tutorial)
	}

	scene.AddObject(c.choices)
}

func (c *Controller) onMenuButtonClicked(gsignal.Void) {
	c.leaveScene(c.backController)
}

func (c *Controller) onToggleButtonClicked(gsignal.Void) {
	c.selectNextColony(true)
}

func (c *Controller) onChoiceSelected(choice selectedChoice) {
	if choice.Option.special == specialChoiceNone {
		c.selectedColony.factionWeights.AddWeight(choice.Faction, c.world.rand.FloatRange(0.1, 0.2))
		for _, e := range choice.Option.effects {
			c.selectedColony.actionPriorities.AddWeight(e.priority, e.value)
		}
		return
	}

	var relocationVec gmath.Vec
	switch choice.Option.special {
	case specialAttack:
		c.launchAttack()
	case specialChoiceMoveColony:
		maxDist := c.selectedColony.MaxFlyDistance() * c.world.rand.FloatRange(0.9, 1.1)
		clickPos := choice.Pos
		clickDist := c.selectedColony.pos.DistanceTo(clickPos)
		dist := gmath.ClampMax(clickDist, maxDist)
		relocationVec = c.selectedColony.pos.VecTowards(clickPos, 1).Mulf(dist)
	case specialIncreaseRadius:
		c.selectedColony.realRadius += c.world.rand.FloatRange(16, 32)
		c.selectedColony.realRadiusSqr = c.selectedColony.realRadius * c.selectedColony.realRadius
	case specialDecreaseRadius:
		value := c.world.rand.FloatRange(30, 40)
		c.selectedColony.realRadius = gmath.ClampMin(c.selectedColony.realRadius-value, 96)
		c.selectedColony.realRadiusSqr = c.selectedColony.realRadius * c.selectedColony.realRadius
	case specialBuildColony, specialBuildGunpoint:
		// TODO: use a pathing.Grid to find a free cell?
		stats := colonyCoreConstructionStats
		dist := 60.0
		size := 40.0
		if choice.Option.special == specialBuildGunpoint {
			stats = gunpointConstructionStats
			dist = 48.0
			size = 32.0
		}
		direction := c.world.rand.Rad()
		for i := 0; i < 18; i++ {
			locationProbe := gmath.RadToVec(direction).Mulf(dist).Add(c.selectedColony.pos)
			direction += (2 * math.Pi) / 17
			constructionPos := c.pickColonyPos(nil, locationProbe, size, 4)
			if !constructionPos.IsZero() {
				construction := c.world.NewConstructionNode(constructionPos, stats)
				c.scene.AddObject(construction)
				break
			}
		}
	}

	if !relocationVec.IsZero() {
		c.launchRelocation(c.selectedColony, relocationVec)
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
	if c.selectedColony.agents.NumAvailableFighters() == 0 {
		return
	}
	closeTargets := c.world.tmpTargetSlice[:0]
	maxDist := gmath.ClampMin(c.selectedColony.PatrolRadius()*1.85, 320)
	maxDist *= c.world.rand.FloatRange(0.95, 1.2)
	for _, creep := range c.world.creeps {
		if len(closeTargets) >= 5 {
			break
		}
		if creep.pos.DistanceTo(c.selectedColony.pos) > maxDist {
			continue
		}
		closeTargets = append(closeTargets, creep)
	}
	if len(closeTargets) == 0 {
		return
	}
	maxDispatched := gmath.Clamp(int(float64(c.selectedColony.agents.NumAvailableFighters())*0.6), 1, 15)
	c.selectedColony.agents.Find(searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
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

func (c *Controller) launchRelocation(core *colonyCoreNode, vec gmath.Vec) {
	r := 48.0
	for i := 0; i < 4; i++ {
		probe := core.pos.Add(vec)
		relocationPoint := c.pickColonyPos(core, probe, r, 5)
		if !relocationPoint.IsZero() {
			core.doRelocation(relocationPoint)
			return
		}
		r -= 2
		vec = vec.Mulf(0.85)
	}
	core.doRelocation(core.pos)
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
	c.menuButton.SetVisibility(false)
	c.toggleButton.SetVisibility(false)

	c.transitionQueued = true
	c.scene.DelayedCall(2.0, func() {
		c.world.result.Victory = false
		c.world.result.TimePlayed = time.Since(c.startTime)
		c.leaveScene(newResultsController(c.state, c.backController, c.world.result))
	})
}

func (c *Controller) victory() {
	if c.transitionQueued {
		return
	}
	c.transitionQueued = true
	c.scene.DelayedCall(5.0, func() {
		c.world.result.Victory = true
		c.world.result.TimePlayed = time.Since(c.startTime)
		for _, colony := range c.world.colonies {
			c.world.result.SurvivingDrones += colony.NumAgents()
		}
		c.leaveScene(newResultsController(c.state, c.backController, c.world.result))
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
		c.onMenuButtonClicked(gsignal.Void{})
	}

	if mainInput.ActionIsJustPressed(controls.ActionToggleColony) {
		c.onToggleButtonClicked(gsignal.Void{})
	}

	handledClick := false
	if len(c.world.colonies) > 1 {
		if pos, ok := c.cursor.ClickPos(controls.ActionClick); ok {
			clickPos := pos.Add(c.camera.Offset)
			selectDist := 40.0
			if c.state.Device.IsMobile {
				selectDist = 80.0
			}
			var closestColony *colonyCoreNode
			closestDist := math.MaxFloat64
			for _, colony := range c.world.colonies {
				if colony == c.selectedColony {
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
	if c.menuButton.HandleInput(controls.ActionClick) {
		return
	}
	if c.toggleButton.HandleInput(controls.ActionClick) {
		return
	}
	c.choices.HandleInput()
}

func (c *Controller) Update(delta float64) {
	c.musicPlayer.Update(delta)
	c.world.Update(delta)

	if c.world.boss == nil {
		// TODO: just subscribe to a boss destruction event?
		c.victory()
	}

	c.choices.Enabled = c.selectedColony != nil &&
		c.selectedColony.mode == colonyModeNormal

	c.tier3spawnDelay = gmath.ClampMin(c.tier3spawnDelay-delta, 0)
	if c.tier3spawnDelay == 0 {
		c.spawnTier3Creep()
	}

	c.handleInput()

	if c.debugInfo != nil {
		colony := c.selectedColony
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
	if c.selectedColony == colony {
		return
	}
	if c.selectedColony != nil {
		c.scene.Audio().PlaySound(assets.AudioBaseSelect)
		c.selectedColony.EventDestroyed.Disconnect(c)
	}
	c.selectedColony = colony
	c.choices.selectedColony = colony
	c.radar.SetBase(c.selectedColony)
	if c.selectedColony == nil {
		c.colonySelector.Visible = false
		c.defeat()
		return
	}
	c.selectedColony.EventDestroyed.Connect(c, func(_ *colonyCoreNode) {
		c.selectNextColony(false)
	})
	c.colonySelector.Pos.Base = &c.selectedColony.spritePos
}

func (c *Controller) selectNextColony(center bool) {
	colony := c.findNextColony()
	c.selectColony(colony)
	if center && c.selectedColony != nil {
		c.camera.CenterOn(c.selectedColony.pos)
	}
}

func (c *Controller) findNextColony() *colonyCoreNode {
	if len(c.world.colonies) == 0 {
		return nil
	}
	if len(c.world.colonies) == 1 {
		return c.world.colonies[0]
	}
	index := xslices.Index(c.world.colonies, c.selectedColony)
	if index == len(c.world.colonies)-1 {
		index = 0
	} else {
		index++
	}
	return c.world.colonies[index]
}
