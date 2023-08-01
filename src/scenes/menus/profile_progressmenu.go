package menus

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ProfileProgressMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewProfileProgressMenuController(state *session.State) *ProfileProgressMenuController {
	return &ProfileProgressMenuController{state: state}
}

func (c *ProfileProgressMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ProfileProgressMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileProgressMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(280, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	tinyFont := assets.BitmapFont1

	helpLabel := eui.NewLabel("", tinyFont)
	helpLabel.MaxWidth = 340

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.progress"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	panel := eui.NewTextPanel(uiResources, 0, 0)
	rowContainer.AddChild(panel)

	numDrones := len(gamedata.Tier2agentMergeRecipes)

	stats := c.state.Persistent.PlayerStats

	modesTotal := 4
	modesUnlocked := 0
	if stats.TotalScore >= gamedata.ClassicModeCost {
		modesUnlocked++
	}
	if stats.TotalScore >= gamedata.ArenaModeCost {
		modesUnlocked++
	}
	if stats.TotalScore >= gamedata.InfArenaModeCost {
		modesUnlocked++
	}
	if stats.TotalScore >= gamedata.ReverseModeCost {
		modesUnlocked++
	}

	smallFont := assets.BitmapFont1

	grid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(24, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))
	lines := [][2]string{
		{d.Get("menu.profile.progress.achievements"), fmt.Sprintf("%d/%d", len(stats.Achievements), len(gamedata.AchievementList))},
		{d.Get("menu.profile.progress.cores_unlocked"), fmt.Sprintf("%d/%d", len(stats.CoresUnlocked), len(gamedata.CoreStatsList))},
		{d.Get("menu.profile.progress.turrets_unlocked"), fmt.Sprintf("%d/%d", len(stats.TurretsUnlocked), len(gamedata.TurretStatsList))},
		{d.Get("menu.profile.progress.drones_unlocked"), fmt.Sprintf("%d/%d", len(stats.DronesUnlocked), numDrones)},
		{d.Get("menu.profile.progress.t3drones_seen"), fmt.Sprintf("%d/%d", len(stats.Tier3DronesSeen), len(gamedata.Tier3agentMergeRecipes))},
		{d.Get("menu.profile.progress.modes_unlocked"), fmt.Sprintf("%d/%d", modesUnlocked, modesTotal)},
	}
	for _, pair := range lines {
		grid.AddChild(eui.NewLabel(pair[0], smallFont))
		grid.AddChild(eui.NewLabel(pair[1], smallFont))
	}
	panel.AddChild(grid)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileProgressMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
