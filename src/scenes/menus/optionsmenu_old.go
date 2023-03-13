//go:build ignore
// +build ignore

package menus

import (
	"strconv"

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
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	d := c.scene.Dict()
	titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		var effectsSlider gmath.Slider
		effectsSlider.SetBounds(0, 6)
		effectsSlider.TrySetValue(options.EffectsVolumeLevel)
		effectsButton := eui.NewButtonSelected(uiResources, d.Get("menu.options.effects_volume")+": "+strconv.Itoa(effectsSlider.Value()))
		effectsButton.ClickedEvent.AddHandler(func(args interface{}) {
			effectsSlider.Inc()
			options.EffectsVolumeLevel = effectsSlider.Value()
			effectsButton.Text().Label = d.Get("menu.options.effects_volume") + ": " + strconv.Itoa(effectsSlider.Value())
			if options.EffectsVolumeLevel != 0 {
				c.scene.Audio().SetGroupVolume(assets.SoundGroupEffect, assets.VolumeMultiplier(options.EffectsVolumeLevel))
				c.scene.Audio().PlaySound(assets.AudioAssaultShot)
			}
		})
		rowContainer.AddChild(effectsButton)
	}

	{
		var musicSlider gmath.Slider
		musicSlider.SetBounds(0, 6)
		musicSlider.TrySetValue(options.MusicVolumeLevel)
		musicButton := eui.NewButtonSelected(uiResources, d.Get("menu.options.music_volume")+": "+strconv.Itoa(musicSlider.Value()))
		musicButton.ClickedEvent.AddHandler(func(args interface{}) {
			musicSlider.Inc()
			options.MusicVolumeLevel = musicSlider.Value()
			musicButton.Text().Label = d.Get("menu.options.music_volume") + ": " + strconv.Itoa(musicSlider.Value())
			if options.MusicVolumeLevel != 0 {
				c.scene.Audio().SetGroupVolume(assets.SoundGroupMusic, assets.VolumeMultiplier(options.MusicVolumeLevel))
				c.scene.Audio().PauseCurrentMusic()
				c.scene.Audio().PlayMusic(assets.AudioMusicTrack3)
			} else {
				c.scene.Audio().PauseCurrentMusic()
			}
		})
		rowContainer.AddChild(musicButton)
	}

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

	{
		sliderOptions := []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, len(sliderOptions)-1)
		if options.Debug {
			slider.TrySetValue(1)
		}
		debugButton := eui.NewButtonSelected(uiResources, d.Get("menu.options.debug")+": "+sliderOptions[slider.Value()])
		debugButton.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Debug = slider.Value() != 0
			debugButton.Text().Label = d.Get("menu.options.debug") + ": " + sliderOptions[slider.Value()]
		})
		rowContainer.AddChild(debugButton)
	}
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
