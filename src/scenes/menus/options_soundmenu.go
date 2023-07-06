package menus

import (
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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *OptionsSoundMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.sound"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.CombinedInput,
			Value:      &options.EffectsVolumeLevel,
			Label:      d.Get("menu.options.effects_volume"),
			ValueNames: []string{"0", "1", "2", "3", "4", "5", "6"},
			OnPressed: func() {
				if options.EffectsVolumeLevel != 0 {
					c.scene.Audio().SetGroupVolume(assets.SoundGroupEffect, assets.VolumeMultiplier(options.EffectsVolumeLevel))
					c.scene.Audio().PlaySound(assets.AudioAssaultShot)
				}
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Resources:  uiResources,
			Input:      c.state.CombinedInput,
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
		}))
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

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
