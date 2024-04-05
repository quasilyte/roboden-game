package menus

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"

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
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	var buttons []eui.Widget

	options := &c.state.Persistent.Settings

	gameplayButton := eui.NewButton(uiResources, c.scene, d.Get("menu.options.gameplay"), func() {
		c.scene.Context().ChangeScene(NewOptionsGameplayMenuController(c.state))
	})
	rowContainer.AddChild(gameplayButton)
	buttons = append(buttons, gameplayButton)

	soundButton := eui.NewButton(uiResources, c.scene, d.Get("menu.options.sound"), func() {
		c.scene.Context().ChangeScene(NewOptionsSoundMenuController(c.state))
	})
	rowContainer.AddChild(soundButton)
	buttons = append(buttons, soundButton)

	graphicsButton := eui.NewButton(uiResources, c.scene, d.Get("menu.options.graphics"), func() {
		c.scene.Context().ChangeScene(NewOptionsGraphicsMenuController(c.state))
	})
	rowContainer.AddChild(graphicsButton)
	buttons = append(buttons, graphicsButton)

	controlsButton := eui.NewButton(uiResources, c.scene, d.Get("menu.options.controls"), func() {
		if c.state.Device.IsMobile() {
			c.scene.Context().ChangeScene(NewControlsTouchMenuController(c.state))
		} else {
			c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
		}
	})
	rowContainer.AddChild(controlsButton)
	buttons = append(buttons, controlsButton)

	if !c.state.Device.IsMobile() {
		accessibilityButton := eui.NewButton(uiResources, c.scene, d.Get("menu.options.accessibility"), func() {
			c.scene.Context().ChangeScene(NewOptionsAccessibilityMenuController(c.state))
		})
		rowContainer.AddChild(accessibilityButton)
		buttons = append(buttons, accessibilityButton)
	}

	extraButton := eui.NewButton(uiResources, c.scene, d.Get("menu.options.extra"), func() {
		c.scene.Context().ChangeScene(NewOptionsExtraMenuController(c.state))
	})
	rowContainer.AddChild(extraButton)
	buttons = append(buttons, extraButton)

	{
		langOptions := []string{
			"en",
			"ru",
		}
		langIndex := xslices.Index(langOptions, options.Lang)
		langSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound:  true,
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Value:      &langIndex,
			Label:      "Language/Язык",
			ValueNames: langOptions,
			OnPressed: func() {
				options.Lang = langOptions[langIndex]
				c.state.ReloadLanguage(c.scene.Context())
				if c.state.Device.IsMobile() {
					// Show the spinner and cache the glyphs in the background.
					c.scene.Context().ChangeScene(NewGlyphCacheController(c.state))
				} else {
					// Just reload the current scene.
					c.scene.Context().ChangeScene(c)
				}
			},
		})
		c.scene.AddObject(langSelect)
		rowContainer.AddChild(langSelect.Widget)
		buttons = append(buttons, langSelect.Widget)
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *OptionsMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
