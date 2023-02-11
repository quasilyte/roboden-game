package menus

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type MainMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewMainMenuController(state *session.State) *MainMenuController {
	return &MainMenuController{state: state}
}

func (c *MainMenuController) Init(scene *ge.Scene) {
	c.scene = scene

	scene.Audio().SetGroupVolume(assets.SoundGroupMusic,
		assets.VolumeMultiplier(c.state.Persistent.Settings.MusicVolumeLevel))
	scene.Audio().SetGroupVolume(assets.SoundGroupEffect,
		assets.VolumeMultiplier(c.state.Persistent.Settings.EffectsVolumeLevel))

	if c.state.Persistent.Settings.MusicVolumeLevel != 0 {
		scene.Audio().ContinueMusic(assets.AudioMusicTrack3)
	}

	c.initUI()
}

func (c *MainMenuController) Update(delta float64) {
}

func (c *MainMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer()
	root.AddChild(rowContainer)

	bigFont := c.scene.Context().Loader.LoadFont(assets.FontBig).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	titleLabel := eui.NewLabel(uiResources, "Roboden", bigFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Start Game", func() {
		c.scene.Context().ChangeScene(NewLobbyMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Tutorial", func() {
		c.state.LevelOptions.Tutorial = true
		c.scene.Context().ChangeScene(staging.NewController(c.state, 0, NewMainMenuController(c.state)))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Settings", func() {
		c.scene.Context().ChangeScene(NewOptionsController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Controls", func() {
		c.scene.Context().ChangeScene(NewControlsMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Credits", func() {
		c.scene.Context().ChangeScene(NewCreditsMenuController(c.state))
	}))

	if runtime.GOARCH != "wasm" {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Exit", func() {
			os.Exit(0)
		}))
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	buildVersionLabel := eui.NewLabel(uiResources, fmt.Sprintf("Build %d (a gamejam prototype)", buildNumber), smallFont)
	rowContainer.AddChild(buildVersionLabel)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}
