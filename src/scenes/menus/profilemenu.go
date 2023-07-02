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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.profile"), assets.BitmapFont3)
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

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.profile.watch_replay"), func() {
		c.scene.Context().ChangeScene(NewReplayMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	if !c.state.Device.IsMobile {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.set_user_name"), func() {
			c.scene.Context().ChangeScene(NewUserNameMenuController(c.state, c))
		}))
	}

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
