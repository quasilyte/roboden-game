package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsGraphicsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsGraphicsMenuController(state *session.State) *OptionsGraphicsMenuController {
	return &OptionsGraphicsMenuController{state: state}
}

func (c *OptionsGraphicsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsGraphicsMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsGraphicsMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	d := c.scene.Dict()
	titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.graphics"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		sliderOptions := []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, len(sliderOptions)-1)
		if options.Graphics.ShadowsEnabled {
			slider.TrySetValue(1)
		}
		disableShadows := eui.NewButtonSelected(uiResources, d.Get("menu.options.graphics.shadows")+": "+sliderOptions[slider.Value()])
		disableShadows.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Graphics.ShadowsEnabled = slider.Value() != 0
			disableShadows.Text().Label = d.Get("menu.options.graphics.shadows") + ": " + sliderOptions[slider.Value()]
		})
		rowContainer.AddChild(disableShadows)
	}

	{
		sliderOptions := []string{
			d.Get("menu.option.mandatory"),
			d.Get("menu.option.all"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, len(sliderOptions)-1)
		if options.Graphics.AllShadersEnabled {
			slider.TrySetValue(1)
		}
		b := eui.NewButtonSelected(uiResources, d.Get("menu.options.graphics.shaders")+": "+sliderOptions[slider.Value()])
		b.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Graphics.AllShadersEnabled = slider.Value() != 0
			b.Text().Label = d.Get("menu.options.graphics.shaders") + ": " + sliderOptions[slider.Value()]
		})
		rowContainer.AddChild(b)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsGraphicsMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
