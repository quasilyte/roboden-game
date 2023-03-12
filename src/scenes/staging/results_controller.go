package staging

import (
	"fmt"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type resultsController struct {
	state *session.State

	scene          *ge.Scene
	backController ge.SceneController

	results battleResults
}

type battleResults struct {
	Victory         bool
	TimePlayed      time.Duration
	SurvivingDrones int

	ResourcesGathered      float64
	EliteResourcesGathered float64
	DronesProduced         int
	CreepsDefeated         int
}

func newResultsController(state *session.State, backController ge.SceneController, results battleResults) *resultsController {
	return &resultsController{
		state:          state,
		backController: backController,
		results:        results,
	}
}

func (c *resultsController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *resultsController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *resultsController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	d := c.scene.Dict()

	titleString := d.Get("menu.results.defeat")
	if c.results.Victory {
		titleString = d.Get("menu.results.victory") + "!"
	}
	titleLabel := eui.NewLabel(uiResources, titleString, smallFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	lines := []string{
		fmt.Sprintf("%s: %v", d.Get("menu.results.time_played"), formatDuration(d, c.results.TimePlayed)),
		fmt.Sprintf("%s: %v", d.Get("menu.results.resources_gathered"), int(c.results.ResourcesGathered)),
		fmt.Sprintf("%s: %v", d.Get("menu.results.drone_survivors"), c.results.SurvivingDrones),
		fmt.Sprintf("%s: %v", d.Get("menu.results.drones_total"), c.results.DronesProduced),
		fmt.Sprintf("%s: %v", d.Get("menu.results.creeps_defeated"), c.results.CreepsDefeated),
	}

	for _, l := range lines {
		label := eui.NewLabel(uiResources, l, smallFont)
		rowContainer.AddChild(label)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby_back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *resultsController) back() {
	c.scene.Context().ChangeScene(c.backController)
}

func formatDuration(dict *langs.Dictionary, d time.Duration) string {
	d = d.Round(time.Second)
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	if hours >= 1 {
		return fmt.Sprintf("%d%s %d%s %d%s",
			hours, dict.Get("game.value.hour"), minutes, dict.Get("game.value.minute"), seconds, dict.Get("game.value.second"))
	}
	if minutes >= 1 {
		return fmt.Sprintf("%d%s %d%s", minutes, dict.Get("game.value.minute"), seconds, dict.Get("game.value.second"))

	}
	return fmt.Sprintf("%d%s", seconds, dict.Get("game.value.second"))
}
