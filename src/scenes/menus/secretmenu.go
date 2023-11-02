package menus

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type SecretMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewSecretMenuController(state *session.State) *SecretMenuController {
	return &SecretMenuController{state: state}
}

func (c *SecretMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	if c.state.UnlockAchievement(session.Achievement{Name: "secret", Elite: true}) {
		c.scene.Context().SaveGameData("save", c.state.Persistent)
	}
	c.initUI()
}

func (c *SecretMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *SecretMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Context().Dict

	smallFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel("???", assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 0, 0)

	label := eui.NewLabel(d.Get("menu.special_text"), smallFont)
	panel.AddChild(label)

	rowContainer.AddChild(panel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *SecretMenuController) back() {
	c.scene.Context().ChangeScene(NewCreditsMenuController(c.state))
}
