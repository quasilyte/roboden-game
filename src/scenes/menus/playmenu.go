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

type PlayMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewPlayMenuController(state *session.State) *PlayMenuController {
	return &PlayMenuController{state: state}
}

func (c *PlayMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *PlayMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *PlayMenuController) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(440, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.play"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.play.classic"), func() {
		c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeClassic))
	}))

	score := c.state.Persistent.PlayerStats.TotalScore
	{
		label := d.Get("menu.play.arena")
		unlocked := score >= gamedata.ArenaModeCost
		if !unlocked {
			label += fmt.Sprintf(" [%d/%d]", score, gamedata.ArenaModeCost)
		}
		b := eui.NewButton(uiResources, c.scene, label, func() {
			c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeArena))
		})
		b.GetWidget().Disabled = !unlocked
		rowContainer.AddChild(b)
	}

	{
		label := d.Get("menu.play.inf_arena")
		unlocked := score >= gamedata.InfArenaModeCost
		if !unlocked {
			label += fmt.Sprintf(" [%d/%d]", score, gamedata.InfArenaModeCost)
		}
		b := eui.NewButton(uiResources, c.scene, label, func() {
			c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeInfArena))
		})
		b.GetWidget().Disabled = !unlocked
		rowContainer.AddChild(b)
	}

	{
		label := d.Get("menu.play.reverse")
		unlocked := score >= gamedata.ReverseModeCost
		if !unlocked {
			label += fmt.Sprintf(" [%d/%d]", score, gamedata.ReverseModeCost)
		}
		b := eui.NewButton(uiResources, c.scene, label, func() {
			c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeReverse))
		})
		b.GetWidget().Disabled = !unlocked
		rowContainer.AddChild(b)
	}

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.play.tutorial"), func() {
		c.scene.Context().ChangeScene(NewTutorialMenuController(c.state))
	}))

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *PlayMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
