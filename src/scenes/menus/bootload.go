package menus

import (
	"fmt"
	"math"
	"runtime"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/gtask"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/steamsdk"
)

type BootloadController struct {
	state *session.State

	scene *ge.Scene
}

func NewBootloadController(state *session.State) *BootloadController {
	return &BootloadController{state: state}
}

func (c *BootloadController) Init(scene *ge.Scene) {
	c.scene = scene

	d := c.scene.Dict()

	smallFont := assets.BitmapFont1
	normalFont := assets.BitmapFont2

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	rowContainer.AddChild(eui.NewCenteredLabel(d.Get("boot.title"), normalFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	progressLabel := eui.NewCenteredLabel("<initializing game user interface>", normalFont)
	rowContainer.AddChild(progressLabel)

	if runtime.GOARCH == "wasm" {
		rowContainer.AddChild(eui.NewCenteredLabel("(!) "+d.Get("boot.wasm"), smallFont))
	}

	currentStepName := ""
	currentStep := -1

	initTask := gtask.StartTask(func(ctx *gtask.TaskContext) {
		type initializationStep struct {
			name string
			f    func(*ge.Context, *assets.Config, *float64)
		}

		config := &assets.Config{
			XM: c.state.Persistent.Settings.XM,
		}

		if c.state.Persistent.Settings.XM {
			c.state.Logf("loading XM music")
		} else {
			c.state.Logf("loading OGG music")
		}

		steps := []initializationStep{
			{name: "load_images", f: assets.RegisterImageResources},
			{name: "load_audio", f: assets.RegisterAudioResource},
			{name: "load_music", f: assets.RegisterMusicResource},
			{name: "load_shaders", f: assets.RegisterShaderResources},
			{name: "load_ui", f: c.loadUIResources},
			{name: "load_extra", f: c.loadExtra},
		}
		if c.state.Device.Steam.Initialized {
			steps = append(steps, initializationStep{
				name: "steam_sync",
				f:    c.steamSync,
			})
		}
		ctx.Progress.Total = float64(len(steps) + 1)
		for _, step := range steps {
			currentStep++
			currentStepName = step.name
			step.f(scene.Context(), config, &ctx.Progress.Current)
			runtime.Gosched()
			runtime.GC()
			ctx.Progress.Current = 1.0 * float64(currentStep+1)
		}
		ctx.Progress.Current++
	})

	initTask.EventProgress.Connect(nil, func(progress gtask.TaskProgress) {
		if progress.Current == progress.Total {
			progressLabel.Label = d.Get("boot.almost_there")
		} else {
			p := int(math.Round(progress.Current*100)) - (currentStep * 100)
			progressLabel.Label = fmt.Sprintf("%s: %d%%", d.Get("boot", currentStepName), p)
		}
	})
	initTask.EventCompleted.Connect(nil, func(gsignal.Void) {
		c.state.AdjustVolumeLevels()

		if c.state.Persistent.FirstLaunch {
			if c.onFirstLaunch() {
				return
			}
		}

		// There is no audio system preload on Android, but we still
		// may want to do it. Play some sound before we show the splashscreen.
		if runtime.GOOS == "android" {
			c.scene.Audio().PlaySound(assets.AudioChoiceReady)
		}

		if c.state.Persistent.Settings.Demo {
			c.scene.Context().ChangeScene(NewSplashScreenController(c.state, NewMainMenuController(c.state)))
		} else {
			c.prepareBackground()
			c.scene.Context().ChangeScene(NewMainMenuController(c.state))
		}
	})

	scene.AddObject(initTask)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *BootloadController) onFirstLaunch() bool {
	if c.state.Device.Steam.Initialized {
		// Infer the player's name from the Steam account info.
		name := steamsdk.PlayerName()
		name = gamedata.CleanUsername(name)
		if gamedata.IsValidUsername(name) {
			c.state.Persistent.PlayerName = name
		}

		// If it's a Steam Deck, set appropriate defaults.
		if c.state.Device.IsSteamDeck() {
			c.state.Persistent.Settings.GamepadSettings[0].Layout = int(gameinput.GamepadLayoutSteamDeck)
			c.state.CombinedInput.SetGamepadLayout(gameinput.GamepadLayoutSteamDeck)
			c.state.FirstGamepadInput.SetGamepadLayout(gameinput.GamepadLayoutSteamDeck)

			// Use 16:10 on Steam Deck instead of default 16:9.
			// This removes the black vertical lines/borders.
			c.state.Persistent.Settings.Graphics.AspectRatio = gamedata.FindDisplayRatio("16:10")
		}
	}

	if c.state.Device.IsMobile() {
		ratioX, ratioY := c.scene.Context().InferDisplayRatio()
		ratioKey := fmt.Sprintf("%d:%d", ratioX, ratioY)
		c.state.Persistent.Settings.Graphics.AspectRatio = gamedata.FindDisplayRatio(ratioKey)
		c.state.Logf("the inferred display ratio is %s", ratioKey)
	}

	{
		displayRatio := gamedata.SupportedDisplayRatios[c.state.Persistent.Settings.Graphics.AspectRatio]
		ctx := c.scene.Context()
		ctx.WindowWidth = displayRatio.Width
		ctx.WindowHeight = displayRatio.Height
		ctx.ScreenWidth = displayRatio.Width
		ctx.ScreenHeight = displayRatio.Height
	}

	c.state.Persistent.FirstLaunch = false
	c.state.SaveGameItem("save.json", c.state.Persistent)

	if c.state.Device.IsSteamDeck() || c.state.Device.IsMobile() {
		// Do not redirect to a controls prompt.
		// We know the layout perfectly well.
		return false
	} else {
		// PC platforms.
		c.scene.Context().ChangeScene(NewControlsPromptController(c.state))
		return true
	}
}

func (c *BootloadController) prepareBackground() {
	bg := ge.NewTiledBackground(c.scene.Context())
	width := c.scene.Context().WindowWidth
	height := c.scene.Context().WindowHeight
	bg.LoadTilesetWithRand(c.scene.Context(), c.scene.Rand(), width, height, assets.ImageBackgroundTiles, assets.RawTilesJSON)
	img := ebiten.NewImage(int(width), int(height))
	bg.Draw(img)
	c.state.BackgroundImage = img
}

func (c *BootloadController) steamSync(ctx *ge.Context, config *assets.Config, progress *float64) {
	if !c.state.Device.Steam.Initialized {
		return
	}

	progressPerItem := 1.0 / float64(len(c.state.Persistent.PlayerStats.Achievements))

	for i, a := range c.state.Persistent.PlayerStats.Achievements {
		*progress += progressPerItem
		unlocked, err := steamsdk.IsAchievementUnlocked(a.Name)
		if err != nil {
			c.state.Logf("check %q achievement (i=%d): %v", a.Name, i, err)
			return
		}
		if !unlocked {
			if !steamsdk.UnlockAchievement(a.Name) {
				c.state.Logf("failed to unlock %q", a.Name)
				return
			}
			c.state.Logf("unlocked %q", a.Name)
		}
	}
}

func (c *BootloadController) loadUIResources(ctx *ge.Context, config *assets.Config, progress *float64) {
	*progress = 0.1
	c.state.Resources.UI = eui.LoadResources(c.state.Device, c.scene.Context().Loader)
}

func (c *BootloadController) loadExtra(ctx *ge.Context, config *assets.Config, progress *float64) {
	steps := []struct {
		dst     **ge.Texture
		imageID resource.ImageID
		length  float64
	}{
		{&gamedata.CourierAgentStats.BeamTexture, assets.ImageCourierLine, 120},
		{&gamedata.RepairAgentStats.BeamTexture, assets.ImageRepairLine, gamedata.RepairAgentStats.SupportRange * 1.4},
		{&gamedata.RechargerAgentStats.BeamTexture, assets.ImageRechargerLine, gamedata.RepairAgentStats.SupportRange * 1.4},
		{&gamedata.DefenderAgentStats.BeamTexture, assets.ImageDefenderLine, gamedata.DefenderAgentStats.Weapon.AttackRange * 1.05},
		{&gamedata.GuardianAgentStats.BeamTexture, assets.ImageDefenderLine, gamedata.GuardianAgentStats.Weapon.AttackRange * 1.05},
		{&gamedata.BeamTowerAgentStats.BeamTexture, assets.ImageBeamtowerLine, gamedata.BeamTowerAgentStats.Weapon.AttackRange * 1.1},
		{&gamedata.RepulseTowerAgentStats.BeamTexture, assets.ImageTempestLine, gamedata.RepulseTowerAgentStats.Weapon.AttackRange * 1.1},
		{&gamedata.TetherBeaconAgentStats.BeamTexture, assets.ImageTetherLine, gamedata.TetherBeaconAgentStats.SupportRange * 1.5},
		{&gamedata.TargeterAgentStats.BeamTexture, assets.ImageTargeterLine, gamedata.TargeterAgentStats.Weapon.AttackRange * 1.15},
		{&gamedata.FirebugAgentStats.BeamTexture, assets.ImageFlamerLine, gamedata.FirebugAgentStats.Weapon.AttackRange * 2},
		{&gamedata.RelictAgentStats.BeamTexture, assets.ImageRelictAgentLine, gamedata.RelictAgentStats.Weapon.AttackRange * 1.5},

		{&gamedata.StunnerCreepStats.BeamTexture, assets.ImageStunnerLine, gamedata.StunnerCreepStats.Weapon.AttackRange * 1.1},
		{&gamedata.TemplarCreepStats.BeamTexture, assets.ImageTemplarLine, gamedata.TemplarCreepStats.Weapon.AttackRange * 1.1},
		{&gamedata.UberBossCreepStats.BeamTexture, assets.ImageBossLaserLine, gamedata.UberBossCreepStats.Weapon.AttackRange * 1.1},

		{&gamedata.LavaGeyserBeamTexture, assets.ImageLavaGeyserLine, 80},
	}

	progressPerItem := 1.0 / float64(len(steps))

	for _, step := range steps {
		*step.dst = ge.NewHorizontallyRepeatedTexture(c.scene.LoadImage(step.imageID), step.length)
		*progress += progressPerItem
	}

	if c.state.Device.IsMobile() {
		alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 <>.,=+-:()[]&@'\"%!?"
		if c.state.Persistent.Settings.Lang == "ru" {
			alphabet = "абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ0123456789 <>.,=+-:()[]&@'\"%!?"
		}
		text.CacheGlyphs(assets.BitmapFont1, alphabet)
		text.CacheGlyphs(assets.BitmapFont2, alphabet)
		text.CacheGlyphs(assets.BitmapFont3, alphabet)
	}
}

func (c *BootloadController) Update(delta float64) {
}
