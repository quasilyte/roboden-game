package menus

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type LeaderboardMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewLeaderboardMenuController(state *session.State) *LeaderboardMenuController {
	return &LeaderboardMenuController{state: state}
}

func (c *LeaderboardMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *LeaderboardMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *LeaderboardMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(400, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.leaderboard"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	var buttons = []eui.Widget{
		eui.NewButton(uiResources, c.scene, d.Get("menu.leaderboard.blitz"), func() {
			c.scene.Context().ChangeScene(NewLeaderboardLoadingController(c.state, gamedata.SeasonNumber, "blitz"))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.leaderboard.classic"), func() {
			c.scene.Context().ChangeScene(NewLeaderboardLoadingController(c.state, gamedata.SeasonNumber, "classic"))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.leaderboard.arena"), func() {
			c.scene.Context().ChangeScene(NewLeaderboardLoadingController(c.state, gamedata.SeasonNumber, "arena"))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.leaderboard.reverse"), func() {
			c.scene.Context().ChangeScene(NewLeaderboardLoadingController(c.state, gamedata.SeasonNumber, "reverse"))
		}),
		eui.NewButton(uiResources, c.scene, d.Get("menu.leaderboard.inf_arena"), func() {
			c.scene.Context().ChangeScene(NewLeaderboardLoadingController(c.state, gamedata.SeasonNumber, "inf_arena"))
		}),
	}

	for _, b := range buttons {
		rowContainer.AddChild(b)
	}

	rowContainer.AddChild(eui.NewTransparentSeparator())

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	rowContainer.AddChild(backButton)
	buttons = append(buttons, backButton)

	navTree := createSimpleNavTree(buttons)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *LeaderboardMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
