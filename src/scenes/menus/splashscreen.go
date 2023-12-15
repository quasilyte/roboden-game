package menus

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

type SplashScreenController struct {
	state *session.State

	scene *ge.Scene

	darkRect *ge.Rect

	controller *staging.Controller
	simulated  bool

	nextController ge.SceneController
}

func NewSplashScreenController(state *session.State, nextController ge.SceneController) *SplashScreenController {
	return &SplashScreenController{
		state:          state,
		nextController: nextController,
	}
}

func (c *SplashScreenController) Init(scene *ge.Scene) {
	c.scene = scene

	c.scene.Audio().SetGroupVolume(assets.SoundGroupMusic, 0)
	c.scene.Audio().SetGroupVolume(assets.SoundGroupEffect, 0)

	config := c.state.SplashLevelConfig.Clone()
	if scene.Rand().Chance(0.4) {
		config.InitialCreeps = 0
		config.CreepFortress = true
	}
	config.CoordinatorCreeps = scene.Rand().Chance(0.7)
	config.WeatherEnabled = scene.Rand().Chance(0.6)
	config.CoreDesign = gamedata.PickColonyDesign(c.state.Persistent.PlayerStats.CoresUnlocked, scene.Rand())
	config.TurretDesign = gamedata.PickTurretDesign(c.state.Persistent.PlayerStats.TurretsUnlocked, scene.Rand())
	config.Tier2Recipes = gamedata.CreateDroneBuild(scene.Rand())
	config.ExecMode = gamedata.ExecuteDemo
	config.PlayersMode = serverapi.PmodeSingleBot

	envPicker := gmath.NewRandPicker[gamedata.EnvironmentKind](scene.Rand())
	envPicker.AddOption(gamedata.EnvForest, 0.35)
	envPicker.AddOption(gamedata.EnvInferno, 0.30)
	envPicker.AddOption(gamedata.EnvSnow, 0.25)
	envPicker.AddOption(gamedata.EnvMoon, 0.20)
	config.Environment = int(envPicker.Pick())

	if scene.Rand().Chance(0.3) {
		config.IonMortars = true
	} else if scene.Rand().Chance(0.25) {
		config.GrenadierCreeps = true
	}
	config.Seed = scene.Rand().PositiveInt64()
	for i := 0; i < 3; i++ {
		config.ExtraDrones = append(config.ExtraDrones, gamedata.WorkerAgentStats)
	}
	for i := 0; i < 5; i++ {
		d := gamedata.FindRecipeByName(gmath.RandElem(scene.Rand(), config.Tier2Recipes)).Result
		if d.Kind == gamedata.AgentRoomba {
			config.ExtraDrones = append(config.ExtraDrones, gamedata.ScoutAgentStats)
			continue
		}
		config.ExtraDrones = append(config.ExtraDrones, d)
	}
	config.Finalize()
	c.controller = staging.NewController(c.state, config, c.nextController)
	scene.AddObject(c.controller)

	c.controller.EventBeforeLeaveScene.Connect(nil, func(gsignal.Void) {
		// Just in case demo stops by a victory/defeat,
		// make sure that we capture that last frame.
		c.state.BackgroundImage = c.controller.RenderDemoFrame()

		if c.state.UnlockAchievement(session.Achievement{Name: "spectator", Elite: true}) {
			c.state.SaveGameItem("save.json", c.state.Persistent)
		}
	})

	logo := scene.NewSprite(assets.ImageLogo)
	logo.Pos.Offset.X = scene.Context().ScreenWidth / 2
	logo.Pos.Offset.Y = scene.Context().ScreenHeight / 5
	scene.AddGraphics(logo)

	d := scene.Dict()
	input := c.state.MenuInput

	presskeyLabel := ge.NewLabel(assets.BitmapFont2)
	presskeyLabel.Width = scene.Context().WindowWidth
	presskeyLabel.AlignHorizontal = ge.AlignHorizontalCenter
	presskeyLabel.Text = input.ReplaceKeyNames(d.Get("game.splash.presskey", input.DetectInputMode()))
	presskeyLabel.SetColorScaleRGBA(0x9d, 0xd7, 0x93, 0xff)
	presskeyLabel.Pos.Offset.Y = logo.Pos.Offset.Y + 54
	scene.AddGraphics(presskeyLabel)

	c.darkRect = ge.NewRect(scene.Context(), scene.Context().ScreenWidth, scene.Context().ScreenHeight)
	c.darkRect.Centered = false
	c.darkRect.FillColorScale.SetRGBA(0, 0, 0, 0xff)
	scene.AddGraphics(c.darkRect)
}

func (c *SplashScreenController) GetSessionState() *session.State {
	return c.state
}

func (c *SplashScreenController) Update(delta float64) {
	if !c.simulated {
		c.simulated = true
		var cameraPos gmath.Vec
		timeSimulated := 0.0
		maxFrames := 20 * (60 * 60)
		for i := 0; i < maxFrames/(10*60); i++ {
			for j := 0; j < (10 * 60); j++ {
				dt := 1.0 / 60.0
				timeSimulated += dt
				c.controller.Update(1.0 / 60.0)
			}
			if timeSimulated >= 3*60 {
				pos, isExciting := c.controller.IsExcitingDemoFrame()
				if isExciting {
					cameraPos = pos
					break
				}
			}
		}
		c.scene.Audio().SetGroupVolume(assets.SoundGroupMusic, assets.VolumeMultiplier(c.state.Persistent.Settings.MusicVolumeLevel))
		c.scene.Audio().SetGroupVolume(assets.SoundGroupEffect, assets.VolumeMultiplier(c.state.Persistent.Settings.EffectsVolumeLevel))
		if c.state.Persistent.Settings.MusicVolumeLevel != 0 {
			c.scene.Audio().PlayMusic(assets.AudioMusicTrack2)
		}
		c.controller.CenterDemoCamera(cameraPos)
	}

	if c.darkRect != nil {
		c.darkRect.FillColorScale.A -= float32(delta)
		if c.darkRect.FillColorScale.A <= 0.1 {
			c.darkRect.Dispose()
			c.darkRect = nil
		}
	}

	c.handleInput()
}

func (c *SplashScreenController) handleInput() {
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionSkipDemo) {
		c.stopDemo()
		return
	}
}

func (c *SplashScreenController) stopDemo() {
	c.state.BackgroundImage = c.controller.RenderDemoFrame()

	c.scene.Audio().PauseCurrentMusic()
	c.scene.Context().ChangeScene(c.nextController)
}
