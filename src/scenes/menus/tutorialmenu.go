package menus

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type TutorialMenuController struct {
	state  *session.State
	config gamedata.LevelConfig

	scene *ge.Scene

	helpLabel *widget.Text
}

func NewTutorialMenuController(state *session.State) *TutorialMenuController {
	return &TutorialMenuController{state: state}
}

func (c *TutorialMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *TutorialMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *TutorialMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	descriptionText := func(id int) string {
		data := gamedata.Tutorials[id]
		description := d.Get("tutorial.description" + strconv.Itoa(id+1))
		var objective string
		if data.Objective != gamedata.ObjectiveTrigger {
			objective = fmt.Sprintf("%s: %s", d.Get("ui.mission_objective"), strings.ToLower(d.Get("objective", data.Objective.String())))
		} else {
			objective = fmt.Sprintf("%s: %s", d.Get("ui.mission_objective"), strings.ToLower(d.Get(data.ObjectiveKey)))
		}
		rewardText := fmt.Sprintf("%s: %d", d.Get("tutorial.reward"), data.ScoreReward)
		if xslices.Contains(c.state.Persistent.PlayerStats.TutorialsCompleted, id) {
			rewardText += " (" + d.Get("tutorial.reward_claimed") + ")"
		}
		return description + "\n\n" + objective + "\n" + rewardText
	}

	tinyFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.play")+" -> "+d.Get("menu.play.tutorial"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	helpLabel := eui.NewLabel("", tinyFont)
	helpLabel.MaxWidth = 540
	c.helpLabel = helpLabel

	c.config = c.state.TutorialLevelConfig.Clone()

	c.helpLabel.Label = descriptionText(c.config.Tutorial.ID)

	panel := eui.NewPanel(uiResources, 560, 220)
	panel.AddChild(helpLabel)
	rowContainer.AddChild(panel)

	{
		tutorialIndex := c.config.Tutorial.ID
		rowContainer.AddChild(eui.NewSelectButton(eui.SelectButtonConfig{
			Scene:     c.scene,
			Resources: uiResources,
			Input:     c.state.MainInput,
			Value:     &tutorialIndex,
			ValueNames: []string{
				d.Get("tutorial.title1"),
				d.Get("tutorial.title2"),
				d.Get("tutorial.title3"),
				d.Get("tutorial.title4"),
			},
			OnPressed: func() {
				c.config.Tutorial = gamedata.Tutorials[tutorialIndex]
				c.helpLabel.Label = descriptionText(tutorialIndex)
			},
		}))
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		// Clone before overriding any extra options.
		clonedConfig := c.config.Clone()
		c.state.TutorialLevelConfig = &clonedConfig

		tutorial := c.config.Tutorial
		c.config.GameMode = gamedata.ModeTutorial
		c.config.Tier2Recipes = tutorial.Tier2Drones
		c.config.WorldSize = tutorial.WorldSize
		c.config.Resources = tutorial.Resources
		c.config.StartingResources = 0
		c.config.Teleporters = 0
		c.config.InterfaceMode = 2
		c.config.InitialCreeps = tutorial.InitialCreeps
		c.config.EliteResources = tutorial.RedCrystals
		c.config.AttackActionAvailable = tutorial.CanAttack
		c.config.BuildTurretActionAvailable = tutorial.CanBuildTurrets
		c.config.RadiusActionAvailable = tutorial.CanChangeRadius
		c.config.EnemyBoss = tutorial.Boss
		c.config.CreepDifficulty = 0
		c.config.BossDifficulty = 0
		c.config.NumCreepBases = tutorial.NumEnemyBases
		c.config.SecondBase = tutorial.SecondBase
		c.config.ExtraDrones = tutorial.ExtraDrones
		c.config.Seed = tutorial.Seed

		c.scene.Context().ChangeScene(staging.NewController(c.state, c.config.Clone(), NewTutorialMenuController(c.state)))
	}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *TutorialMenuController) back() {
	c.scene.Context().ChangeScene(NewPlayMenuController(c.state))
}
