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

	if c.state.Persistent.Settings.MusicVolumeLevel != 0 {
		scene.Audio().ContinueMusic(assets.AudioMusicTrack3)
	}

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

	d := c.scene.Dict()

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.start_game"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.LevelOptions

	{
		valueNames := []string{
			d.Get("menu.option.very_low"),
			d.Get("menu.option.low"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.rich"),
			d.Get("menu.option.very_rich"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 4)
		slider.TrySetValue(options.Resources)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.world_resources")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Resources = slider.Value()
			button.Text().Label = d.Get("menu.lobby.world_resources") + ": " + valueNames[slider.Value()]
		})
		rowContainer.AddChild(button)
	}

	{
		valueNames := []string{
			d.Get("menu.option.very_small"),
			d.Get("menu.option.small"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.big"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(options.WorldSize)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.world_size")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.WorldSize = slider.Value()
			button.Text().Label = d.Get("menu.lobby.world_size") + ": " + valueNames[slider.Value()]
		})
		rowContainer.AddChild(button)
	}

	{
		valueNames := []string{
			d.Get("menu.option.very_easy"),
			d.Get("menu.option.easy"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.hard"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(options.Difficulty)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.difficulty")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Difficulty = slider.Value()
			button.Text().Label = d.Get("menu.lobby.difficulty") + valueNames[slider.Value()]
		})
		rowContainer.AddChild(button)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		c.state.LevelOptions.Tutorial = false
		c.scene.Context().ChangeScene(staging.NewController(c.state, options.WorldSize, NewLobbyMenuController(c.state)))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *LobbyMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
