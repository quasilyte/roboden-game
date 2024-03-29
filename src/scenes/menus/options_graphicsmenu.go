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

	var buttons []eui.Widget

	options := &c.state.Persistent.Settings

	{
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.Graphics.ShadowsEnabled,
			Label:     d.Get("menu.options.graphics.shadows"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	{
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.Graphics.VSyncEnabled,
			Label:     d.Get("menu.options.graphics.vsync"),
			OnPressed: func() {
				ebiten.SetVsyncEnabled(options.Graphics.VSyncEnabled)
			},
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	{
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.Graphics.CameraShakingEnabled,
			Label:     d.Get("menu.options.graphics.camera_shaking"),
			ValueNames: []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	{
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.Graphics.AllShadersEnabled,
			Label:     d.Get("menu.options.graphics.shaders"),
			ValueNames: []string{
				d.Get("menu.option.mandatory"),
				d.Get("menu.option.all"),
			},
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	screenFilterSelect := eui.NewSelectButton(eui.SelectButtonConfig{
		Input:     c.state.MenuInput,
		PlaySound: true,
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
	})
	c.scene.AddObject(screenFilterSelect)
	rowContainer.AddChild(screenFilterSelect.Widget)
	buttons = append(buttons, screenFilterSelect.Widget)

	if c.state.Device.IsDesktop() {
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound: true,
			Resources: uiResources,
			Input:     c.state.MenuInput,
			BoolValue: &options.Graphics.FullscreenEnabled,
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
		})
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
		buttons = append(buttons, b.Widget)
	}

	if runtime.GOARCH != "wasm" {
		values := []string{
			gamedata.SupportedDisplayRatios[0].Name,
			gamedata.SupportedDisplayRatios[1].Name,
			gamedata.SupportedDisplayRatios[2].Name,
			gamedata.SupportedDisplayRatios[3].Name,
			gamedata.SupportedDisplayRatios[4].Name,
			gamedata.SupportedDisplayRatios[5].Name,
		}
		if c.state.Device.IsSteamDeck() {
			// Steam Deck Native display.
			values = append(values, gamedata.SupportedDisplayRatios[6].Name)
		}
		b := eui.NewSelectButton(eui.SelectButtonConfig{
			PlaySound:  true,
			Resources:  uiResources,
			Input:      c.state.MenuInput,
			Value:      &options.Graphics.AspectRatio,
			Label:      d.Get("menu.options.graphics.aspect_ratio"),
			ValueNames: values,
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
		c.scene.AddObject(b)
		rowContainer.AddChild(b.Widget)
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

func (c *OptionsGraphicsMenuController) back() {
	c.state.SaveGameItem("save.json", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsController(c.state))
}
