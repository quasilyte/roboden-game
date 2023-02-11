package menus

import (
	"runtime/pprof"

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

	if c.state.CPUProfileWriter != nil {
		pprof.StopCPUProfile()
		if err := c.state.CPUProfileWriter.Close(); err != nil {
			panic(err)
		}
	}
	if c.state.MemProfileWriter != nil {
		pprof.WriteHeapProfile(c.state.MemProfileWriter)
		if err := c.state.MemProfileWriter.Close(); err != nil {
			panic(err)
		}
	}
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

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	titleLabel := eui.NewLabel(uiResources, "Main Menu -> Start Game", normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.LevelOptions

	{
		valueNames := []string{
			"very low",
			"low",
			"normal",
			"rich",
			"very rich",
		}
		var slider gmath.Slider
		slider.SetBounds(0, 4)
		slider.TrySetValue(options.Resources)
		button := eui.NewButtonSelected(uiResources, "Map Resources: "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Resources = slider.Value()
			button.Text().Label = "Map Resources: " + valueNames[slider.Value()]
		})
		rowContainer.AddChild(button)
	}

	{
		valueNames := []string{
			"very easy",
			"easy",
			"normal",
			"hard",
		}
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(options.Difficulty)
		button := eui.NewButtonSelected(uiResources, "Difficulty: "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Difficulty = slider.Value()
			button.Text().Label = "Difficulty: " + valueNames[slider.Value()]
		})
		rowContainer.AddChild(button)
	}

	{
		valueNames := []string{
			"very small",
			"small",
			"normal",
		}
		var slider gmath.Slider
		slider.SetBounds(0, 2)
		slider.TrySetValue(options.WorldSize)
		button := eui.NewButtonSelected(uiResources, "Map Size: "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.WorldSize = slider.Value()
			button.Text().Label = "Map Size: " + valueNames[slider.Value()]
		})
		rowContainer.AddChild(button)
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Go", func() {
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
