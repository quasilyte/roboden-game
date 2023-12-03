package menus

import (
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/gtask"
	"github.com/quasilyte/roboden-game/session"
)

type GlyphCacheController struct {
	scene         *ge.Scene
	state         *session.State
	spinner       *widget.Text
	spinnerFrames []string
	t             float64
}

func NewGlyphCacheController(state *session.State) *GlyphCacheController {
	return &GlyphCacheController{
		state: state,
	}
}

func (c *GlyphCacheController) Init(scene *ge.Scene) {
	c.scene = scene

	c.initUI()
	c.spawnTask()
	c.spinnerFrames = []string{`\`, `|`, `/`, `--`}
}

func (c *GlyphCacheController) spawnTask() {
	// Caching glyphs can take a lot of time on mobiles, so we do
	// it on a separate goroutine hoping not to block the main app thread.
	initTask := gtask.StartTask(func(ctx *gtask.TaskContext) {
		time.Sleep(time.Second)
		c.state.CacheGlyphs()
	})

	initTask.EventCompleted.Connect(nil, func(gsignal.Void) {
		c.scene.Context().ChangeScene(NewOptionsController(c.state))
	})

	c.scene.AddObject(initTask)
}

func (c *GlyphCacheController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	c.spinner = eui.NewCenteredLabel("--", assets.BitmapFont2)
	rowContainer.AddChild(c.spinner)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *GlyphCacheController) Update(delta float64) {
	c.t += 10 * delta
	c.spinner.Label = c.spinnerFrames[int(c.t)%len(c.spinnerFrames)]
}
