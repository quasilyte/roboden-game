package menus

import (
	"strconv"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type OptionsCameraMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsCameraMenuController(state *session.State) *OptionsCameraMenuController {
	return &OptionsCameraMenuController{state: state}
}

func (c *OptionsCameraMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsCameraMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsCameraMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	d := c.scene.Dict()
	titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.camera"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		var scrollSlider gmath.Slider
		scrollSlider.SetBounds(0, 4)
		scrollSlider.TrySetValue(options.ScrollingSpeed)
		scrollButton := eui.NewButtonSelected(uiResources, d.Get("menu.options.scroll_speed")+": "+strconv.Itoa(scrollSlider.Value()+1))
		scrollButton.ClickedEvent.AddHandler(func(args interface{}) {
			scrollSlider.Inc()
			options.ScrollingSpeed = scrollSlider.Value()
			scrollButton.Text().Label = d.Get("menu.options.scroll_speed") + ": " + strconv.Itoa(scrollSlider.Value()+1)
		})
		rowContainer.AddChild(scrollButton)
	}

	{
		var scrollSlider gmath.Slider
		scrollSlider.SetBounds(0, 4)
		scrollSlider.TrySetValue(options.EdgeScrollRange)
		scrollButton := eui.NewButtonSelected(uiResources, d.Get("menu.options.edge_scroll_range")+": "+strconv.Itoa(scrollSlider.Value()))
		scrollButton.ClickedEvent.AddHandler(func(args interface{}) {
			scrollSlider.Inc()
			options.EdgeScrollRange = scrollSlider.Value()
			scrollButton.Text().Label = d.Get("menu.options.edge_scroll_range") + ": " + strconv.Itoa(scrollSlider.Value())
		})
		rowContainer.AddChild(scrollButton)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsCameraMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
