package menus

import (
	"fmt"
	"strings"

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
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileProgressMenuController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	helpLabel := eui.NewLabel(uiResources, "", tinyFont)
	helpLabel.MaxWidth = 340

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.progress"), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	numDrones := len(gamedata.Tier2agentMergeRecipes)

	stats := c.state.Persistent.PlayerStats

	modesTotal := 2
	modesUnlocked := 1
	if stats.TotalScore >= gamedata.ArenaModeCost {
		modesUnlocked++
	}

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face
	lines := []string{
		fmt.Sprintf("%s: %d/%d", d.Get("menu.profile.progress.achievements"), len(stats.Achievements), len(gamedata.AchievementList)),
		fmt.Sprintf("%s: %d/%d", d.Get("menu.profile.progress.turrets_unlocked"), len(stats.TurretsUnlocked), len(gamedata.TurretStatsList)),
		fmt.Sprintf("%s: %d/%d", d.Get("menu.profile.progress.drones_unlocked"), len(stats.DronesUnlocked), numDrones),
		fmt.Sprintf("%s: %d/%d", d.Get("menu.profile.progress.t3drones_seen"), len(stats.Tier3DronesSeen), len(gamedata.Tier3agentMergeRecipes)),
		fmt.Sprintf("%s: %d/%d", d.Get("menu.profile.progress.modes_unlocked"), modesUnlocked, modesTotal),
		fmt.Sprintf("%s: %d/%d", d.Get("menu.profile.progress.tutorials_completed"), len(stats.TutorialsCompleted), len(gamedata.Tutorials)),
	}
	rowContainer.AddChild(eui.NewCenteredLabel(uiResources, strings.Join(lines, "\n"), smallFont))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileProgressMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
