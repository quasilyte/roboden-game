package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type HintScreen struct {
	state *session.State

	config gamedata.LevelConfig

	backController ge.SceneController

	scene *ge.Scene
}

func NewHintScreen(state *session.State, config gamedata.LevelConfig, back ge.SceneController) *HintScreen {
	return &HintScreen{
		state:          state,
		config:         config,
		backController: back,
	}
}

func (c *HintScreen) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *HintScreen) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *HintScreen) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	d := c.scene.Context().Dict

	titleLabel := eui.NewCenteredLabel(d.Get("menu.play", c.config.GameMode.String()), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	{
		l := eui.NewLabel(d.Get("menu.overview", c.config.GameMode.String()), smallFont)
		l.MaxWidth = 640
		rowContainer.AddChild(l)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		c.scene.Context().ChangeScene(staging.NewController(c.state, c.config, c.backController))
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *HintScreen) back() {
	c.scene.Context().ChangeScene(NewPlayMenuController(c.state))
}
