package menus

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ControlsPromptController struct {
	state *session.State

	scene *ge.Scene
}

func NewControlsPromptController(state *session.State) *ControlsPromptController {
	return &ControlsPromptController{state: state}
}

func (c *ControlsPromptController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ControlsPromptController) Update(delta float64) {
	c.state.MenuInput.Update()
}

func (c *ControlsPromptController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("game.onboard.welcome"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	promptText := eui.NewCenteredLabel(d.Get("game.onboard.select_input_method"), c.state.Resources.Font1)
	rowContainer.AddChild(promptText)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.keyboard"), func() {
		c.selectControls(gameinput.InputMethodKeyboard)
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.gamepad"), func() {
		c.selectControls(gameinput.InputMethodGamepad1)
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.controls.auto_infer"), func() {
		c.selectControls(gameinput.InputMethodCombined)
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ControlsPromptController) selectControls(inputMethod gameinput.PlayerInputMethod) {
	c.state.Persistent.Settings.Player1InputMethod = int(inputMethod)
	c.state.ReloadInputs()
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewSplashScreenController(c.state, NewMainMenuController(c.state)))
}
