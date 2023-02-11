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
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/viewport"
)

type Controller struct {
	state *session.State

	backController    ge.SceneController
	cameraPanSpeed    float64
	cameraPanBoundary float64

	startTime time.Time

	selectedColony *colonyCoreNode
	colonySelector *ge.Sprite
	radar          *radarNode

	scene     *ge.Scene
	world     *worldState
	worldSize int

	choices *choiceWindowNode

	musicPlayer *musicPlayer

	tier3spawnDelay float64
	tier3spawnRate  float64

	camera *viewport.Camera

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
			c.cameraPanBoundary = 6
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
	}

	viewportWorld := &viewport.World{
		Width:  worldSize,
		Height: worldSize,
	}
	c.camera = viewport.NewCamera(viewportWorld, 1920/2, 1080/2)

	// Start launching tier3 creeps after ~15 minutes.
	c.tier3spawnDelay = scene.Rand().FloatRange(14*60.0, 16*60.0)
	c.tier3spawnRate = 1.0

	c.cameraPanSpeed = float64(c.state.Persistent.Settings.ScrollingSpeed+1) * 4

	world := &worldState{
		debug:          c.state.Persistent.Settings.Debug,
		worldSize:      c.worldSize,
		options:        &c.state.LevelOptions,
		camera:         c.camera,
		rand:           scene.Rand(),
		tmpTargetSlice: make([]projectileTarget, 0, 20),
		width:          viewportWorld.Width,
		height:         viewportWorld.Height,
		rect: gmath.Rect{
			Max: gmath.Vec{
				X: viewportWorld.Width,
				Y: viewportWorld.Height,
			},
		},
	}
	c.world = world

	bg := ge.NewTiledBackground()
	bg.LoadTileset(scene.Context(), world.width, world.height, assets.ImageBackgroundTiles, assets.RawTilesJSON)
	c.camera.AddGraphicsBelow(bg)
	g := newLevelGenerator(scene, c.world)
	g.Generate()

	c.colonySelector = scene.NewSprite(assets.ImageColonyCoreSelector)
	c.camera.AddGraphicsBelow(c.colonySelector)

	c.radar = newRadarNode(c.world)
	scene.AddObject(c.radar)

	choicesPos := gmath.Vec{
		X: 960 - 232 - 16,
		Y: 540 - 200 - 16,
	}
	c.choices = newChoiceWindowNode(choicesPos, c.state.MainInput)
	c.choices.EventChoiceSelected.Connect(nil, c.onChoiceSelected)

	c.selectNextColony(true)
	c.camera.CenterOn(c.selectedColony.pos)

	scene.AddGraphics(c.camera)

	if c.state.Persistent.Settings.Debug {
		c.debugInfo = scene.NewLabel(assets.FontSmall)
		c.debugInfo.ColorScale.SetColor(ge.RGB(0xffffff))
		c.debugInfo.Pos.Offset = gmath.Vec{X: 10, Y: 10}
		scene.AddGraphics(c.debugInfo)
	}

	if c.state.LevelOptions.Tutorial {
		tutorial := newTutorialManager(c.state.MainInput, c.choices)
		scene.AddObject(tutorial)
	}

	scene.AddObject(c.choices)
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
		dist := c.selectedColony.MaxFlyDistance() * c.world.rand.FloatRange(0.9, 1.1)
		clickPos := c.state.MainInput.CursorPos().Add(c.camera.Offset)
		relocationVec = c.selectedColony.pos.VecTowards(clickPos, 1).Mulf(dist)
	case specialIncreaseRadius:
		c.selectedColony.realRadius += c.world.rand.FloatRange(16, 32)
	case specialDecreaseRadius:
		value := c.world.rand.FloatRange(16, 32)
		c.selectedColony.realRadius = gmath.ClampMin(c.selectedColony.realRadius-value, 60)
	case specialBuildColony:
		dist := 60.0
		direction := c.world.rand.Rad()
		for i := 0; i < 11; i++ {
			locationProbe := gmath.RadToVec(direction).Mulf(dist).Add(c.selectedColony.pos)
			direction += (2 * math.Pi) / 13
			constructionPos := c.pickColonyPos(nil, locationProbe, 40, 3)
			if !constructionPos.IsZero() {
				construction := c.world.NewColonyCoreConstructionNode(constructionPos)
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
	if len(c.selectedColony.combatAgents) == 0 {
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
	maxDispatched := gmath.Clamp(int(float64(len(c.selectedColony.combatAgents))*0.6), 1, 15)
	for _, agent := range c.selectedColony.combatAgents {
		if agent.mode != agentModeStandby && agent.mode != agentModePatrol {
			continue
		}
		if maxDispatched == 0 {
			break
		}
		maxDispatched--
		target := gmath.RandElem(c.world.rand, closeTargets)
		agent.AssignMode(agentModeAttack, gmath.Vec{}, target)
	}
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

func (c *Controller) Update(delta float64) {
	c.musicPlayer.Update(delta)

	if c.world.boss == nil {
		c.scene.DelayedCall(5.0, func() {
			c.world.result.Victory = true
			c.world.result.TimePlayed = time.Since(c.startTime)
			for _, colony := range c.world.colonies {
				c.world.result.SurvivingDrones += colony.NumAgents()
			}
			c.leaveScene(newResultsController(c.state, c.backController, c.world.result))
		})
	}

	c.choices.Enabled = c.selectedColony != nil &&
		c.selectedColony.mode == colonyModeNormal

	c.tier3spawnDelay = gmath.ClampMin(c.tier3spawnDelay-delta, 0)
	if c.tier3spawnDelay == 0 {
		c.spawnTier3Creep()
	}

	mainInput := c.state.MainInput
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

	if mainInput.ActionIsJustPressed(controls.ActionBack) {
		c.leaveScene(c.backController)
	}

	if mainInput.ActionIsJustPressed(controls.ActionToggleColony) {
		c.selectNextColony(true)
	}
	if len(c.world.colonies) > 1 {
		if info, ok := mainInput.JustPressedActionInfo(controls.ActionClick); ok {
			clickPos := info.Pos.Add(c.camera.Offset)
			for _, colony := range c.world.colonies {
				if colony == c.selectedColony {
					continue
				}
				if colony.pos.DistanceTo(clickPos) > 30 {
					continue
				}
				c.selectColony(colony)
				break
			}
		}
	}

	if c.debugInfo != nil {
		c.debugInfo.Text = fmt.Sprintf("FPS: %f", ebiten.CurrentFPS())
	}
	// colony := c.selectedColony
	// c.debugInfo.Text = fmt.Sprintf("colony resources: %.2f, workers: %d, warriors: %d lim: %d radius: %d\nresources=%d%% growth=%d%% evolution=%d%% security=%d%%\ngray: %d%% yellow: %d%% red: %d%% green: %d%% blue: %d%%\nfps: %f",
	// 	colony.resources.Essence,
	// 	len(colony.agents),
	// 	len(colony.combatAgents),
	// 	colony.calcUnitLimit(),
	// 	int(colony.realRadius),
	// 	int(colony.GetResourcePriority()*100),
	// 	int(colony.GetGrowthPriority()*100),
	// 	int(colony.GetEvolutionPriority()*100),
	// 	int(colony.GetSecurityPriority()*100),
	// 	int(colony.factionWeights.GetWeight(neutralFactionTag)*100),
	// 	int(colony.factionWeights.GetWeight(yellowFactionTag)*100),
	// 	int(colony.factionWeights.GetWeight(redFactionTag)*100),
	// 	int(colony.factionWeights.GetWeight(greenFactionTag)*100),
	// 	int(colony.factionWeights.GetWeight(blueFactionTag)*100),
	// 	ebiten.CurrentFPS())
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
		c.scene.DelayedCall(2.0, func() {
			c.world.result.Victory = false
			c.world.result.TimePlayed = time.Since(c.startTime)
			c.leaveScene(newResultsController(c.state, c.backController, c.world.result))
		})
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
