package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ProfileMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewProfileMenuController(state *session.State) *ProfileMenuController {
	return &ProfileMenuController{state: state}
}

func (c *ProfileMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ProfileMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.profile"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.profile.achievements"), func() {
		c.scene.Context().ChangeScene(NewProfileAchievementsMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.profile.stats"), func() {
		c.scene.Context().ChangeScene(NewProfileStatsMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.profile.progress"), func() {
		c.scene.Context().ChangeScene(NewProfileProgressMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.profile.dronebook"), func() {
		c.scene.Context().ChangeScene(NewProfileDroneCollectionMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
