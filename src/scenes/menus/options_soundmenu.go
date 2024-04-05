package menus

import (
	"runtime"

	"github.com/quasilyte/ge"

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
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *OptionsSoundMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.sound"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	var buttons []eui.Widget

	{
		effectsVolumeSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Value:      &options.EffectsVolumeLevel,
			Label:      d.Get("menu.options.effects_volume"),
			ValueNames: []string{"0", "1", "2", "3", "4", "5", "6"},
			OnPressed: func() {
				if options.EffectsVolumeLevel != 0 {
					c.scene.Audio().SetGroupVolume(assets.SoundGroupEffect, assets.VolumeMultiplier(options.EffectsVolumeLevel))
					c.scene.Audio().PlaySound(assets.AudioAssaultShot)
				}
			},
		})
		c.scene.AddObject(effectsVolumeSelect)
		rowContainer.AddChild(effectsVolumeSelect.Widget)
		buttons = append(buttons, effectsVolumeSelect.Widget)
	}

	{
		musicVolumeSelect := eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Value:      &options.MusicVolumeLevel,
			Label:      d.Get("menu.options.music_volume"),
			ValueNames: []string{"0", "1", "2", "3", "4", "5", "6"},
			OnPressed: func() {
				if options.MusicVolumeLevel != 0 {
					c.scene.Audio().SetGroupVolume(assets.SoundGroupMusic, assets.VolumeMultiplier(options.MusicVolumeLevel))
					c.scene.Audio().PauseCurrentMusic()
					c.scene.Audio().PlayMusic(assets.AudioMusicTrack3)
				} else {
					c.scene.Audio().PauseCurrentMusic()
				}
			},
		})
		c.scene.AddObject(musicVolumeSelect)
		rowContainer.AddChild(musicVolumeSelect.Widget)
		buttons = append(buttons, musicVolumeSelect.Widget)
	}

	{
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.XM,
			Label:     d.Get("menu.options.music_player"),
			ValueNames: []string{
				d.Get("menu.option.music_player.ogg"),
				d.Get("menu.option.music_player.xm"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		// Don't allow web platforms to change the music player.
		// The same goes for Androids.
		b.Widget.GetWidget().Disabled = runtime.GOARCH == "wasm" || runtime.GOOS == "android"
		buttons = append(buttons, b.Widget)
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

func (c *OptionsSoundMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
