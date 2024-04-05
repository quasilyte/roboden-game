package menus

import (
	"github.com/quasilyte/ge"
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
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *ProfileMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.profile"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	buttons := []eui.Widget{
		eui.NewButton(uiResources, c.scene, d.Get("menu.profile.achievements"), func() {
			c.scene.Context().ChangeScene(NewProfileAchievementsMenuController(c.state))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.profile.stats"), func() {
			c.scene.Context().ChangeScene(NewProfileStatsMenuController(c.state))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.profile.progress"), func() {
			c.scene.Context().ChangeScene(NewProfileProgressMenuController(c.state))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.profile.dronebook"), func() {
			c.scene.Context().ChangeScene(NewProfileDroneCollectionMenuController(c.state))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.profile.watch_replay"), func() {
			c.scene.Context().ChangeScene(NewReplayMenuController(c.state))
		}),
	}

	for _, b := range buttons {
		rowContainer.AddChild(b)
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	setUserNameButton := eui.NewButton(uiResources, c.scene, d.Get("menu.set_user_name"), func() {
		c.scene.Context().ChangeScene(NewUserNameMenuController(c.state, c))
	})
	rowContainer.AddChild(setUserNameButton)
	buttons = append(buttons, setUserNameButton)

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *ProfileMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
