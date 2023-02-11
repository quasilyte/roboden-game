package menus

import (
	"strconv"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type LobbyMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewLobbyMenuController(state *session.State) *LobbyMenuController {
	return &LobbyMenuController{state: state}
}

func (c *LobbyMenuController) Init(scene *ge.Scene) {
	c.scene = scene

	c.initUI()
}

func (c *LobbyMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *LobbyMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, "New Game Options", smallFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.LevelOptions

	{
		var slider gmath.Slider
		slider.SetBounds(0, 4)
		slider.TrySetValue(options.Resources)
		button := eui.NewButtonSelected(uiResources, "Map Resources: "+strconv.Itoa(slider.Value()+1))
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Resources = slider.Value()
			button.Text().Label = "Map Resources: " + strconv.Itoa(slider.Value()+1)
		})
		rowContainer.AddChild(button)
	}

	{
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(options.Difficulty)
		button := eui.NewButtonSelected(uiResources, "Difficulty: "+strconv.Itoa(slider.Value()+1))
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Difficulty = slider.Value()
			button.Text().Label = "Difficulty: " + strconv.Itoa(slider.Value()+1)
		})
		rowContainer.AddChild(button)
	}

	{
		var slider gmath.Slider
		slider.SetBounds(0, 2)
		slider.TrySetValue(options.WorldSize)
		button := eui.NewButtonSelected(uiResources, "Map Size: "+strconv.Itoa(slider.Value()+1))
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.WorldSize = slider.Value()
			button.Text().Label = "Map Size: " + strconv.Itoa(slider.Value()+1)
		})
		rowContainer.AddChild(button)
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Start", func() {
		c.state.LevelOptions.Tutorial = false
		c.scene.Context().ChangeScene(staging.NewController(c.state, options.WorldSize, NewLobbyMenuController(c.state)))
	}))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *LobbyMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
