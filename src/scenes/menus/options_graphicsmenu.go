package menus

import (
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
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

	// You can get here from a demo screen that is triggered after an aspect ratio change.
	if c.state.Persistent.Settings.MusicVolumeLevel != 0 {
		scene.Audio().ContinueMusic(assets.AudioMusicTrack3)
	}

	c.initUI()
}

func (c *OptionsGraphicsMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *OptionsGraphicsMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(520, 10, nil)
	root.AddChild(rowContainer)

	normalFont := assets.BitmapFont3

	d := c.scene.Dict()
	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.graphics"), normalFont)
	rowContainer.AddChild(titleLabel)

	options := &c.state.Persistent.Settings

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.ShadowsEnabled,
			Label:     d.Get("menu.options.graphics.shadows"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.VSyncEnabled,
			Label:     d.Get("menu.options.graphics.vsync"),
			OnPressed: func() {
				ebiten.SetVsyncEnabled(options.Graphics.VSyncEnabled)
			},
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.CameraShakingEnabled,
			Label:     d.Get("menu.options.graphics.camera_shaking"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		}))
	}

	{
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.AllShadersEnabled,
			Label:     d.Get("menu.options.graphics.shaders"),
			ValueNames: []string{
				d.Get("menu.option.mandatory"),
				d.Get("menu.option.all"),
			},
		}))
	}

	rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
		Input:     c.state.MenuInput,
		Scene:     c.scene,
		Resources: uiResources,
		Value:     &options.Graphics.ScreenFilter,
		Label:     d.Get("menu.options.graphics.screen_filter"),
		ValueNames: []string{
			d.Get("menu.options.screen_filter.normal"),
			d.Get("menu.options.screen_filter.crt"),
			d.Get("menu.options.screen_filter.sharpen"),
			d.Get("menu.options.screen_filter.heavy_sharpen"),
			d.Get("menu.options.screen_filter.hue_minus30"),
			d.Get("menu.options.screen_filter.hue_minus60"),
			d.Get("menu.options.screen_filter.hue_plus30"),
			d.Get("menu.options.screen_filter.hue_plus60"),
		},
	}))

	if c.state.Device.IsDesktop() {
		rowContainer.AddChild(eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Value:     &options.Graphics.FullscreenEnabled,
			Label:     d.Get("menu.options.graphics.fullscreen"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
			OnPressed: func() {
				ebiten.SetFullscreen(options.Graphics.FullscreenEnabled)
				displayRatio := gamedata.SupportedDisplayRatios[options.Graphics.AspectRatio]
				if options.Graphics.FullscreenEnabled {
					ebiten.SetWindowSize(int(displayRatio.Width), int(displayRatio.Height))
				}
			},
		}))
	}

	if runtime.GOARCH != "wasm" {
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			Value:     &options.Graphics.AspectRatio,
			Label:     d.Get("menu.options.graphics.aspect_ratio"),
			ValueNames: []string{
				gamedata.SupportedDisplayRatios[0].Name,
				gamedata.SupportedDisplayRatios[1].Name,
				gamedata.SupportedDisplayRatios[2].Name,
				gamedata.SupportedDisplayRatios[3].Name,
				gamedata.SupportedDisplayRatios[4].Name,
				gamedata.SupportedDisplayRatios[5].Name,
			},
			OnPressed: func() {
				displayRatio := gamedata.SupportedDisplayRatios[options.Graphics.AspectRatio]
				ctx := c.scene.Context()
				ctx.WindowWidth = displayRatio.Width
				ctx.WindowHeight = displayRatio.Height
				ctx.ScreenWidth = displayRatio.Width
				ctx.ScreenHeight = displayRatio.Height
				if !options.Graphics.FullscreenEnabled {
					ebiten.SetWindowSize(int(displayRatio.Width), int(displayRatio.Height))
				}
				c.scene.Context().ChangeScene(NewSplashScreenController(c.state, c))
			},
		})
		rowContainer.AddChild(b)
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *OptionsGraphicsMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
