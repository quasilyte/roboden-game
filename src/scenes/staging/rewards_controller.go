package staging

import (
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type rewardsController struct {
	state *session.State

	scene *ge.Scene

	rewards gameRewards

	backController ge.SceneController

	showDelay float64

	finished     bool
	root         *widget.Container
	rowContainer *widget.Container
	panel        *widget.Container
	grid         *widget.Container
	lines        [][2]string
}

type gameRewards struct {
	newAchievements      []string
	upgradedAchievements []string
	newCores             []string
	newDrones            []gamedata.ColonyAgentKind
	newTurrets           []gamedata.ColonyAgentKind
	newOptions           []string
	newModes             []string
}

func (rewards *gameRewards) IsEmpty() bool {
	return len(rewards.newAchievements) == 0 &&
		len(rewards.upgradedAchievements) == 0 &&
		len(rewards.newCores) == 0 &&
		len(rewards.newDrones) == 0 &&
		len(rewards.newTurrets) == 0 &&
		len(rewards.newOptions) == 0 &&
		len(rewards.newModes) == 0
}

func newRewardsController(state *session.State, rewards gameRewards, backController ge.SceneController) *rewardsController {
	return &rewardsController{
		state:          state,
		rewards:        rewards,
		backController: backController,
	}
}

func (c *rewardsController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
	c.showDelay = 1.5
}

func (c *rewardsController) Update(delta float64) {
	if c.finished && c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}

	if len(c.lines) == 0 {
		return
	}

	c.showDelay -= delta
	if c.showDelay > 0 {
		return
	}
	c.showDelay = 1.5

	explosionSoundIndex := c.scene.Rand().IntRange(0, 4)
	explosionSound := resource.AudioID(int(assets.AudioExplosion1) + explosionSoundIndex)
	c.scene.Audio().PlaySound(explosionSound)

	smallFont := assets.BitmapFont2
	pair := c.lines[0]
	c.lines = c.lines[1:]
	c.grid.AddChild(eui.NewLabel(pair[0], smallFont))
	c.grid.AddChild(eui.NewLabel(pair[1], smallFont, widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter)))
	if c.panel != nil {
		c.rowContainer.AddChild(c.panel)
		c.panel = nil
	}
	if len(c.lines) == 0 {
		c.finished = true
		c.rowContainer.AddChild(eui.NewButton(c.state.Resources.UI, c.scene, c.scene.Dict().Get("menu.lobby_back"), func() {
			c.back()
		}))
	}
	c.root.RequestRelayout()
}

func (c *rewardsController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	c.root = root
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	c.rowContainer = rowContainer
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	rowContainer.AddChild(eui.NewCenteredLabel(d.Get("menu.results.rewards"), assets.BitmapFont3))

	panel := eui.NewTextPanel(uiResources, 580, 0)
	c.panel = panel

	c.grid = widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Spacing(24, 8),
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
		)))

	for _, a := range c.rewards.newAchievements {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.new_achievement"), d.Get("achievement", a)})
	}
	for _, a := range c.rewards.upgradedAchievements {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.upgraded_achievement"), d.Get("achievement", a)})
	}
	for _, name := range c.rewards.newCores {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.new_core"), d.Get("core", name)})
	}
	for _, kind := range c.rewards.newDrones {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.new_drone"), d.Get("drone", strings.ToLower(kind.String()))})
	}
	for _, kind := range c.rewards.newTurrets {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.new_turret"), d.Get("turret", strings.ToLower(kind.String()))})
	}
	for _, id := range c.rewards.newOptions {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.new_option"), d.Get("menu.lobby", id)})
	}
	for _, id := range c.rewards.newModes {
		c.lines = append(c.lines, [2]string{d.Get("menu.results.new_mode"), d.Get("menu.leaderboard", id)})
	}
	panel.AddChild(c.grid)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *rewardsController) back() {
	c.scene.Context().ChangeScene(c.backController)
}
