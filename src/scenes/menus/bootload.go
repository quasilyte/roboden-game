package menus

import (
	"fmt"
	"math"
	"runtime"

	"github.com/ebitenui/ebitenui/widget"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/gtask"
	"github.com/quasilyte/roboden-game/session"
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

	assets.RegisterFontResources(scene.Context(), nil)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

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
		steps := []struct {
			name string
			f    func(*ge.Context, *float64)
		}{
			{"load_images", assets.RegisterImageResources},
			{"load_audio", assets.RegisterAudioResource},
			{"load_music", assets.RegisterMusicResource},
			{"load_shaders", assets.RegisterShaderResources},
			{"load_ui", c.loadUIResources},
			{"load_extra", c.loadExtra},
		}
		ctx.Progress.Total = float64(len(steps))
		for _, step := range steps {
			currentStep++
			currentStepName = step.name
			step.f(scene.Context(), &ctx.Progress.Current)
			runtime.Gosched()
			runtime.GC()
			ctx.Progress.Current = 1.0 * float64(currentStep+1)
		}
	})

	initTask.EventProgress.Connect(nil, func(progress gtask.TaskProgress) {
		p := int(math.Round(progress.Current*100)) - (currentStep * 100)
		progressLabel.Label = fmt.Sprintf("%s: %d%%", d.Get("boot", currentStepName), p)
	})
	initTask.EventCompleted.Connect(nil, func(gsignal.Void) {
		c.scene.Context().ChangeScene(NewMainMenuController(c.state))
	})

	scene.AddObject(initTask)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *BootloadController) loadUIResources(ctx *ge.Context, progress *float64) {
	*progress = 0.1
	c.state.Resources.UI = eui.LoadResources(c.state.Device, c.scene.Context().Loader)
}

func (c *BootloadController) loadExtra(ctx *ge.Context, progress *float64) {
	steps := []struct {
		agent   *gamedata.AgentStats
		imageID resource.ImageID
		length  float64
	}{
		{gamedata.RepairAgentStats, assets.ImageRepairLine, gamedata.RepairAgentStats.SupportRange},
		{gamedata.RechargeAgentStats, assets.ImageRechargerLine, gamedata.RepairAgentStats.SupportRange},
		{gamedata.DefenderAgentStats, assets.ImageDefenderLine, gamedata.DefenderAgentStats.Weapon.AttackRange},
		{gamedata.BeamTowerAgentStats, assets.ImageBeamtowerLine, gamedata.BeamTowerAgentStats.Weapon.AttackRange},
		{gamedata.TetherBeaconAgentStats, assets.ImageTetherLine, gamedata.TetherBeaconAgentStats.SupportRange * 1.5},
	}

	progressPerItem := 1.0 / float64(len(steps))
	for _, step := range steps {
		step.agent.BeamTexture = ge.NewHorizontallyRepeatedTexture(c.scene.LoadImage(step.imageID), step.length)
		*progress += progressPerItem
	}
}

func (c *BootloadController) Update(delta float64) {
}
