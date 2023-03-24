package menus

import (
	"fmt"
	"strconv"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type TutorialMenuController struct {
	state  *session.State
	config session.LevelConfig

	scene *ge.Scene

	helpLabel *widget.Text
}

func NewTutorialMenuController(state *session.State) *TutorialMenuController {
	return &TutorialMenuController{state: state}
}

func (c *TutorialMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *TutorialMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *TutorialMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	descriptionText := func(id int) string {
		data := gamedata.Tutorials[id]
		description := d.Get("tutorial.description" + strconv.Itoa(id+1))
		rewardText := fmt.Sprintf("%s: %d", d.Get("tutorial.reward"), data.ScoreReward)
		if xslices.Contains(c.state.Persistent.PlayerStats.TutorialsCompleted, id) {
			rewardText += " (" + d.Get("tutorial.reward_claimed") + ")"
		}
		return description + "\n" + rewardText
	}

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.profile")+" -> "+d.Get("menu.play.tutorial"), normalFont)
	rowContainer.AddChild(titleLabel)

	helpLabel := eui.NewLabel(uiResources, "", tinyFont)
	helpLabel.MaxWidth = 540
	c.helpLabel = helpLabel

	c.config = c.state.LevelConfig.Clone()
	if c.config.Tutorial == nil {
		c.config.Tutorial = &session.TutorialData{}
	}

	{
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(c.config.Tutorial.ID)
		button := eui.NewButtonSelected(uiResources, d.Get("tutorial.title"+strconv.Itoa(slider.Value()+1)))
		c.helpLabel.Label = descriptionText(slider.Value())
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			c.config.Tutorial.ID = slider.Value()
			button.Text().Label = d.Get("tutorial.title" + strconv.Itoa(slider.Value()+1))
			c.helpLabel.Label = descriptionText(slider.Value())
		})
		rowContainer.AddChild(button)
	}

	panel := eui.NewPanel(uiResources, 560, 220)
	panel.AddChild(helpLabel)
	rowContainer.AddChild(panel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		// c.state.LevelOptions.Tutorial = nil
		// c.state.LevelOptions.DifficultyScore = c.calcDifficultyScore()
		// c.state.LevelOptions.DronePointsAllocated = c.calcAllocatedPoints()
		// if c.seedInput.InputText != "" {
		// 	seed, err := strconv.ParseInt(c.seedInput.InputText, 10, 64)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	c.state.LevelOptions.Seed = seed
		// } else {
		// 	c.state.LevelOptions.Seed = c.randomSeed()
		// }
		// c.scene.Context().ChangeScene(staging.NewController(c.state, options.WorldSize, NewLobbyMenuController(c.state)))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *TutorialMenuController) back() {
	c.scene.Context().ChangeScene(NewPlayMenuController(c.state))
}
