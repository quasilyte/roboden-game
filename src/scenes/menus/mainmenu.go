package menus

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
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
	c.initTextures()

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

func (c *MainMenuController) initTextures() {
	if gamedata.RepairAgentStats.BeamTexture != nil {
		return
	}
	gamedata.RepairAgentStats.BeamTexture = ge.NewHorizontallyRepeatedTexture(c.scene.LoadImage(assets.ImageRepairLine), gamedata.RepairAgentStats.SupportRange)
	gamedata.RechargeAgentStats.BeamTexture = ge.NewHorizontallyRepeatedTexture(c.scene.LoadImage(assets.ImageRechargerLine), gamedata.RechargeAgentStats.SupportRange)
}

func (c *MainMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	bigFont := c.scene.Context().Loader.LoadFont(assets.FontBig).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("game.title"), bigFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.play"), func() {
		c.scene.Context().ChangeScene(NewPlayMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.main.profile"), func() {
		c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
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

	buildVersionLabel := eui.NewCenteredLabel(uiResources, fmt.Sprintf("%s %d (alpha testing)", d.Get("menu.main.build"), buildNumber), smallFont)
	rowContainer.AddChild(buildVersionLabel)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}
