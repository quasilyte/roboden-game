package staging

import (
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

	world := &worldState{
		camera:        c.camera,
		rand:          scene.Rand(),
		tmpAgentSlice: make([]*colonyAgentNode, 0, 8),
		width:         viewportWorld.Width,
		height:        viewportWorld.Height,
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

	c.selectedColony = world.colonies[0]
	c.camera.CenterOn(c.selectedColony.body.Pos)

	c.colonySelector = scene.NewSprite(assets.ImageColonyCoreSelector)
	c.colonySelector.Pos.Base = &c.selectedColony.body.Pos
	c.camera.AddGraphicsBelow(c.colonySelector)

	scene.AddGraphics(c.camera)

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
	var relocationRect gmath.Rect
	switch choice.Option.special {
	case specialChoiceMoveNorth:
		relocationVec = gmath.Vec{Y: c.world.rand.FloatRange(-120, -240)}
		relocationRect.Max = gmath.Vec{X: c.world.width, Y: c.selectedColony.body.Pos.Y}
	case specialChoiceMoveEast:
		relocationVec = gmath.Vec{X: c.world.rand.FloatRange(150, 300)}
		relocationRect.Min = gmath.Vec{X: c.selectedColony.body.Pos.X}
		relocationRect.Max = gmath.Vec{X: c.world.width, Y: c.world.height}
	case specialChoiceMoveSouth:
		relocationVec = gmath.Vec{Y: c.world.rand.FloatRange(120, 240)}
		relocationRect.Min = gmath.Vec{Y: c.selectedColony.body.Pos.Y}
		relocationRect.Max = gmath.Vec{X: c.world.width, Y: c.world.height}
	case specialChoiceMoveWest:
		relocationVec = gmath.Vec{X: c.world.rand.FloatRange(-150, -300)}
		relocationRect.Max = gmath.Vec{X: c.selectedColony.body.Pos.X, Y: c.world.height}
	case specialIncreaseRadius:
		c.selectedColony.realRadius += c.world.rand.FloatRange(16, 32)
	case specialDecreaseRadius:
		c.selectedColony.realRadius -= c.world.rand.FloatRange(16, 32)
	case specialBuildColony:
		dist := 52.0
		for i := 0; i < 5; i++ {
			constructionPos := c.pickColonyPos(nil, c.selectedColony.body.Pos.Add(c.world.rand.Offset(-dist, dist)), 40, 7)
			if !constructionPos.IsZero() {
				construction := c.world.NewColonyCoreConstructionNode(constructionPos)
				c.scene.AddObject(construction)
				break
			}
			dist += 14.0
		}
	}
	if !relocationVec.IsZero() {
		c.launchRelocation(c.selectedColony, relocationVec, relocationRect)
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

func (c *Controller) launchRelocation(core *colonyCoreNode, vec gmath.Vec, rect gmath.Rect) {
	r := 48.0
	for i := 0; i < 4; i++ {
		probe := core.body.Pos.Add(vec)
		relocationPoint := c.pickColonyPos(core, probe, r, 5)
		if !relocationPoint.IsZero() && rect.Contains(relocationPoint) {
			core.doRelocation(relocationPoint)
			return
		}
		r -= 2
		vec = vec.Mulf(0.85)
	}
	core.doRelocation(core.body.Pos)
}

func (c *Controller) Update(delta float64) {
	c.choices.Enabled = c.selectedColony != nil &&
		c.selectedColony.mode == colonyModeNormal

	mainInput := c.state.MainInput
	var cameraPan gmath.Vec
	if mainInput.ActionIsPressed(controls.ActionPanRight) {
		cameraPan.X += 4
	}
	if mainInput.ActionIsPressed(controls.ActionPanDown) {
		cameraPan.Y += 4
	}
	if mainInput.ActionIsPressed(controls.ActionPanLeft) {
		cameraPan.X -= 4
	}
	if mainInput.ActionIsPressed(controls.ActionPanUp) {
		cameraPan.Y -= 4
	}
	c.camera.Pan(cameraPan)

	if mainInput.ActionIsJustPressed(controls.ActionToggleColony) {
		c.selectNextColony()
	}

	// colony := c.selectedColony
	// c.debugInfo.Text = fmt.Sprintf("colony resources: %.2f, workers: %d, warriors: %d lim: %d radius: %d upkeep: %.2f\nresources=%d%% growth=%d%% evolution=%d%% security=%d%%\ngray: %d%% yellow: %d%% red: %d%% green: %d%% blue: %d%%\nfps: %f",
	// 	colony.resources.Essence,
	// 	len(colony.agents),
	// 	len(colony.combatAgents),
	// 	colony.calcUnitLimit(),
	// 	int(colony.radius),
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

func (c *Controller) selectNextColony() {
	c.selectedColony = c.findNextColony()
	c.colonySelector.Pos.Base = &c.selectedColony.body.Pos
	c.camera.CenterOn(c.selectedColony.body.Pos)
}

func (c *Controller) findNextColony() *colonyCoreNode {
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
