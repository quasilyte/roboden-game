package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type PlayMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewPlayMenuController(state *session.State) *PlayMenuController {
	return &PlayMenuController{state: state}
}

func (c *PlayMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *PlayMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *PlayMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.play"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.play.classic"), func() {
		c.scene.Context().ChangeScene(NewLobbyMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.play.rush"), func() {
		// c.scene.Context().ChangeScene(NewLobbyMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.play.arena"), func() {
		// c.scene.Context().ChangeScene(NewLobbyMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.play.tutorial"), func() {
		c.state.LevelOptions.Tutorial = true
		c.scene.Context().ChangeScene(staging.NewController(c.state, 0, NewMainMenuController(c.state)))
	}))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *PlayMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
