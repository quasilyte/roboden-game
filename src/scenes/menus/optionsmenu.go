package menus

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsController(state *session.State) *OptionsMenuController {
	return &OptionsMenuController{state: state}
}

func (c *OptionsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *OptionsMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.gameplay"), func() {
		c.scene.Context().ChangeScene(NewOptionsGameplayMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.sound"), func() {
		c.scene.Context().ChangeScene(NewOptionsSoundMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.graphics"), func() {
		c.scene.Context().ChangeScene(NewOptionsGraphicsMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.controls"), func() {
		if c.state.Device.IsMobile() {
			c.scene.Context().ChangeScene(NewControlsTouchMenuController(c.state))
		} else {
			c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
		}
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.extra"), func() {
		c.scene.Context().ChangeScene(NewOptionsExtraMenuController(c.state))
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
			Input:      c.state.MenuInput,
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

	rowContainer.AddChild(eui.NewTransparentSeparator())

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
