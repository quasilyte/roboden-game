package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type StagingOptionsMenuController struct {
	state *session.State

	scene *ge.Scene

	backController ge.SceneController
}

func NewStagingOptionsController(state *session.State) *StagingOptionsMenuController {
	return &StagingOptionsMenuController{state: state}
}

func (c *StagingOptionsMenuController) WithBackController(controller ge.SceneController) *StagingOptionsMenuController {
	c.backController = controller
	return c
}

func (c *StagingOptionsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *StagingOptionsMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *StagingOptionsMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	if c.backController != nil {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.continue"), func() {
			c.scene.Context().ChangeScene(c.backController)
		}))
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.sound"), func() {
		c.scene.Context().ChangeScene(NewOptionsSoundMenuController(c.state).WithBackController(c))
	}))

	{
		langOptions := []string{
			"en",
			"ru",
		}
		langIndex := xslices.Index(langOptions, options.Lang)
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:      c.scene,
			Resources:  uiResources,
			Input:      c.state.CombinedInput,
			Value:      &langIndex,
			Label:      "Language/Язык",
			ValueNames: langOptions,
			OnPressed: func() {
				options.Lang = langOptions[langIndex]
				c.state.ReloadLanguage(c.scene.Context())
				c.scene.Context().ChangeScene(c)
			},
		}))
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *StagingOptionsMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(c.backController)
}
