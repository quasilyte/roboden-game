package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"

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
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsMenuController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	d := c.scene.Dict()
	titleLabel := eui.NewLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.sound"), func() {
		c.scene.Context().ChangeScene(NewOptionsSoundMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.graphics"), func() {
		c.scene.Context().ChangeScene(NewOptionsGraphicsMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.camera"), func() {
		c.scene.Context().ChangeScene(NewOptionsCameraMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.controls"), func() {
		c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.options.extra"), func() {
		c.scene.Context().ChangeScene(NewOptionsExtraMenuController(c.state))
	}))

	{
		langOptions := []string{
			"en",
			"ru",
		}
		selectedLangIndex := func() int {
			return xslices.Index(langOptions, options.Lang)
		}
		var slider gmath.Slider
		slider.SetBounds(0, len(langOptions)-1)
		slider.TrySetValue(selectedLangIndex())
		button := eui.NewButtonSelected(uiResources, "Language/Язык: "+langOptions[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Lang = langOptions[slider.Value()]
			c.state.ReloadLanguage(c.scene.Context())
			button.Text().Label = "Language/Язык: " + langOptions[slider.Value()]
			c.scene.Context().ChangeScene(c)
		})
		rowContainer.AddChild(button)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

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
