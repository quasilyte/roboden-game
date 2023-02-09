package staging

import (
	"fmt"

	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/colony-game/controls"
	"github.com/quasilyte/colony-game/session"
	"github.com/quasilyte/colony-game/viewport"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
)

type Controller struct {
	state *session.State

	selectedColony *colonyCoreNode
	colonySelector *ge.Sprite

	scene *ge.Scene
	world *worldState

	choices *choiceWindowNode

	tier3spawnDelay float64
	tier3spawnRate  float64

	camera *viewport.Camera

	debugInfo *ge.Label
}

func NewController(state *session.State) *Controller {
	return &Controller{state: state}
}

func (c *Controller) Init(scene *ge.Scene) {
	viewportWorld := &viewport.World{
		Width:  2880,
		Height: 2880,
	}
	c.scene = scene
	c.camera = viewport.NewCamera(viewportWorld, 1920/2, 1080/2)

	// Start launching tier3 creeps after ~15 minutes.
	c.tier3spawnDelay = scene.Rand().FloatRange(14*60.0, 16*60.0)
	c.tier3spawnRate = 1.0

	world := &worldState{
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

	c.selectNextColony(true)
	c.camera.CenterOn(c.selectedColony.pos)

	scene.AddGraphics(c.camera)

	// {
	// 	creep := c.world.NewCreepNode(c.selectedColony.pos.Add(gmath.Vec{X: 100}), assaultCreepStats)
	// 	scene.AddObject(creep)
	// }

	c.debugInfo = scene.NewLabel(assets.FontSmall)
	c.debugInfo.ColorScale.SetColor(ge.RGB(0xffffff))
	c.debugInfo.Pos.Offset = gmath.Vec{X: 10, Y: 10}
	scene.AddGraphics(c.debugInfo)

	choicesPos := gmath.Vec{
		X: 960 - 224 - 16,
		Y: 540 - 192 - 16,
	}
	c.choices = newChoiceWindowNode(choicesPos, c.state.MainInput)
	c.choices.EventChoiceSelected.Connect(nil, c.onChoiceSelected)
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
	case specialChoiceMoveColony:
		dist := c.world.rand.FloatRange(160, 200)
		clickPos := c.state.MainInput.CursorPos().Add(c.camera.Offset)
		relocationVec = c.selectedColony.pos.VecTowards(clickPos, 1).Mulf(dist)

	case specialIncreaseRadius:
		c.selectedColony.realRadius += c.world.rand.FloatRange(16, 32)
	case specialDecreaseRadius:
		c.selectedColony.realRadius -= c.world.rand.FloatRange(16, 32)
	case specialBuildColony:
		dist := 52.0
		for i := 0; i < 5; i++ {
			constructionPos := c.pickColonyPos(nil, c.selectedColony.pos.Add(c.world.rand.Offset(-dist, dist)), 40, 7)
			if !constructionPos.IsZero() {
				construction := c.world.NewColonyCoreConstructionNode(constructionPos)
				c.scene.AddObject(construction)
				break
			}
			dist += 14.0
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
	c.tier3spawnRate = gmath.ClampMin(c.tier3spawnRate-0.02, 0.5)
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
	c.choices.Enabled = c.selectedColony != nil &&
		c.selectedColony.mode == colonyModeNormal

	c.tier3spawnDelay = gmath.ClampMin(c.tier3spawnDelay-delta, 0)
	if c.tier3spawnDelay == 0 {
		c.spawnTier3Creep()
	}

	mainInput := c.state.MainInput
	var cameraPan gmath.Vec
	const cameraPanSpeed float64 = 8.0
	if mainInput.ActionIsPressed(controls.ActionPanRight) {
		cameraPan.X += cameraPanSpeed
	}
	if mainInput.ActionIsPressed(controls.ActionPanDown) {
		cameraPan.Y += cameraPanSpeed
	}
	if mainInput.ActionIsPressed(controls.ActionPanLeft) {
		cameraPan.X -= cameraPanSpeed
	}
	if mainInput.ActionIsPressed(controls.ActionPanUp) {
		cameraPan.Y -= cameraPanSpeed
	}
	if cameraPan.IsZero() {
		// Mouse cursor can pan the camera too.
		cursor := mainInput.CursorPos()
		if cursor.X > c.camera.Rect.Width()-2 {
			cameraPan.X += cameraPanSpeed
		}
		if cursor.Y > c.camera.Rect.Height()-2 {
			cameraPan.Y += cameraPanSpeed
		}
		if cursor.X < 2 {
			cameraPan.X -= cameraPanSpeed
		}
		if cursor.Y < 2 {
			cameraPan.Y -= cameraPanSpeed
		}
	}
	c.camera.Pan(cameraPan)

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

	// colony := c.selectedColony
	// c.debugInfo.Text = fmt.Sprintf("colony resources: %.2f, workers: %d, warriors: %d lim: %d radius: %d upkeep: %.2f\nresources=%d%% growth=%d%% evolution=%d%% security=%d%%\ngray: %d%% yellow: %d%% red: %d%% green: %d%% blue: %d%%\nfps: %f",
	// 	colony.resources.Essence,
	// 	len(colony.agents),
	// 	len(colony.combatAgents),
	// 	colony.calcUnitLimit(),
	// 	int(colony.realRadius),
	// 	colony.calcUpkeed(),
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

func (c *Controller) selectColony(colony *colonyCoreNode) {
	if c.selectedColony != nil {
		c.selectedColony.EventDestroyed.Disconnect(c)
	}
	c.selectedColony = colony
	if c.selectedColony == nil {
		// TODO: game over.
		fmt.Println("game over")
		c.colonySelector.Visible = false
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
	if center {
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
