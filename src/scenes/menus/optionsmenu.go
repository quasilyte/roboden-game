package menus

import (
	"strconv"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/colony-game/controls"
	"github.com/quasilyte/colony-game/gameui/eui"
	"github.com/quasilyte/colony-game/session"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
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
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, "Settings", smallFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		var effectsSlider gmath.Slider
		effectsSlider.SetBounds(0, 6)
		effectsSlider.TrySetValue(options.EffectsVolumeLevel)
		effectsButton := eui.NewButtonSelected(uiResources, "Effects Volume: "+strconv.Itoa(effectsSlider.Value()))
		effectsButton.ClickedEvent.AddHandler(func(args interface{}) {
			effectsSlider.Inc()
			options.EffectsVolumeLevel = effectsSlider.Value()
			effectsButton.Text().Label = "Effects Volume: " + strconv.Itoa(effectsSlider.Value())
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
		musicButton := eui.NewButtonSelected(uiResources, "Music Volume: "+strconv.Itoa(musicSlider.Value()))
		musicButton.ClickedEvent.AddHandler(func(args interface{}) {
			musicSlider.Inc()
			options.MusicVolumeLevel = musicSlider.Value()
			musicButton.Text().Label = "Music Volume: " + strconv.Itoa(musicSlider.Value())
		})
		rowContainer.AddChild(musicButton)
	}

	{
		var scrollSlider gmath.Slider
		scrollSlider.SetBounds(0, 4)
		scrollSlider.TrySetValue(options.ScrollingSpeed)
		scrollButton := eui.NewButtonSelected(uiResources, "Scroll Speed: "+strconv.Itoa(scrollSlider.Value()+1))
		scrollButton.ClickedEvent.AddHandler(func(args interface{}) {
			scrollSlider.Inc()
			options.ScrollingSpeed = scrollSlider.Value()
			scrollButton.Text().Label = "Scroll Speed: " + strconv.Itoa(scrollSlider.Value()+1)
		})
		rowContainer.AddChild(scrollButton)
	}

	{
		debugButton := eui.NewButtonSelected(uiResources, "Debug: "+strconv.FormatBool(options.Debug))
		debugButton.ClickedEvent.AddHandler(func(args interface{}) {
			options.Debug = !options.Debug
			debugButton.Text().Label = "Debug: " + strconv.FormatBool(options.Debug)
		})
		rowContainer.AddChild(debugButton)
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
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
