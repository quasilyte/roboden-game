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

type OptionsSoundMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewOptionsSoundMenuController(state *session.State) *OptionsSoundMenuController {
	return &OptionsSoundMenuController{state: state}
}

func (c *OptionsSoundMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *OptionsSoundMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsSoundMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face

	d := c.scene.Dict()
	titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.sound"), normalFont)
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

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsSoundMenuController) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
