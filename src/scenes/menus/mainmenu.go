package menus

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type MainMenuController struct {
	state *session.State

	cursor *gameui.CursorNode

	scene *ge.Scene
}

func NewMainMenuController(state *session.State) *MainMenuController {
	return &MainMenuController{state: state}
}

func (c *MainMenuController) Init(scene *ge.Scene) {
	c.scene = scene

	c.cursor = gameui.NewCursorNode(&c.state.CombinedInput, scene.Context().WindowRect())
	scene.AddObject(c.cursor)

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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.scene.Audio().PauseCurrentMusic()
		c.scene.Context().ChangeScene(NewSplashScreenController(c.state))
		return
	}
}

func (c *MainMenuController) initUI() {
	addDemoBackground(c.state, c.scene)

	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	logo := widget.NewGraphic(widget.GraphicOpts.Image(c.scene.LoadImage(assets.ImageLogo).Data))
	rowContainer.AddChild(logo)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.play"), func() {
		c.scene.Context().ChangeScene(NewPlayMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.profile"), func() {
		c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.leaderboard"), func() {
		c.scene.Context().ChangeScene(NewLeaderboardMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.settings"), func() {
		c.scene.Context().ChangeScene(NewOptionsController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.credits"), func() {
		c.scene.Context().ChangeScene(NewCreditsMenuController(c.state))
	}))

	if runtime.GOARCH != "wasm" {
		rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.exit"), func() {
			os.Exit(0)
		}))
	}

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	buildVersionLabel := eui.NewCenteredLabel(fmt.Sprintf("%s %d", d.Get("menu.main.build"), gamedata.BuildNumber), assets.BitmapFont1)
	rowContainer.AddChild(buildVersionLabel)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}
