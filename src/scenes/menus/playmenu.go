package menus

import (
	"fmt"

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

type PlayMenuController struct {
	state *session.State

	scene *ge.Scene

	helpLabel *widget.Text
}

func NewPlayMenuController(state *session.State) *PlayMenuController {
	return &PlayMenuController{state: state}
}

func (c *PlayMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *PlayMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *PlayMenuController) modeDescriptionText(name string, cost int) string {
	d := c.scene.Dict()
	score := c.state.Persistent.PlayerStats.TotalScore
	s := d.Get("menu.overview", name)
	if score >= cost {
		return s
	}
	s += "\n\n"
	s += fmt.Sprintf("%s: %d/%d", d.Get("drone.score_required"), score, cost)
	return s
}

func (c *PlayMenuController) setHelpText(s string) {
	c.helpLabel.Label = s
}

func (c *PlayMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainerWithMinWidth(440, 10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.play"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	rootGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))
	rowContainer.AddChild(rootGrid)

	buttonsContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
	)

	leftPanel := eui.NewPanel(uiResources, 360, 0)
	leftPanel.AddChild(buttonsContainer)
	rootGrid.AddChild(leftPanel)

	helpLabel := eui.NewLabel(d.Get("menu.overview.intro_mission"), assets.BitmapFont1)
	helpLabel.MaxWidth = 320
	c.helpLabel = helpLabel

	rightPanel := eui.NewTextPanel(uiResources, 360, 0)
	rightPanel.AddChild(helpLabel)
	rootGrid.AddChild(rightPanel)

	{
		b := eui.NewButtonWithConfig(uiResources, eui.ButtonConfig{
			Scene: c.scene,
			Text:  d.Get("menu.play.intro_mission"),
			OnPressed: func() {
				back := NewPlayMenuController(c.state)
				config := c.state.TutorialLevelConfig.Clone()
				config.Seed = c.scene.Rand().PositiveInt64()
				c.scene.Context().ChangeScene(staging.NewController(c.state, config, back))
			},
			OnHover: func() { c.setHelpText(d.Get("menu.overview.intro_mission")) },
		})
		buttonsContainer.AddChild(b)
	}

	playerStats := &c.state.Persistent.PlayerStats

	{
		label := d.Get("menu.play.classic")
		b := eui.NewButtonWithConfig(uiResources, eui.ButtonConfig{
			Scene: c.scene,
			Text:  label,
			OnPressed: func() {
				c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeClassic))
			},
			OnHover: func() { c.setHelpText(c.modeDescriptionText("classic", gamedata.ClassicModeCost)) },
		})
		b.GetWidget().Disabled = !xslices.Contains(playerStats.ModesUnlocked, "classic")
		buttonsContainer.AddChild(b)
	}

	{
		label := d.Get("menu.play.arena")
		b := eui.NewButtonWithConfig(uiResources, eui.ButtonConfig{
			Scene: c.scene,
			Text:  label,
			OnPressed: func() {
				c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeArena))
			},
			OnHover: func() { c.setHelpText(c.modeDescriptionText("arena", gamedata.ArenaModeCost)) },
		})
		b.GetWidget().Disabled = !xslices.Contains(playerStats.ModesUnlocked, "arena")
		buttonsContainer.AddChild(b)
	}

	{
		label := d.Get("menu.play.inf_arena")
		b := eui.NewButtonWithConfig(uiResources, eui.ButtonConfig{
			Scene: c.scene,
			Text:  label,
			OnPressed: func() {
				c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeInfArena))
			},
			OnHover: func() { c.setHelpText(c.modeDescriptionText("inf_arena", gamedata.InfArenaModeCost)) },
		})
		b.GetWidget().Disabled = !xslices.Contains(playerStats.ModesUnlocked, "inf_arena")
		buttonsContainer.AddChild(b)
	}

	{
		label := d.Get("menu.play.reverse")
		b := eui.NewButtonWithConfig(uiResources, eui.ButtonConfig{
			Scene: c.scene,
			Text:  label,
			OnPressed: func() {
				c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, gamedata.ModeReverse))
			},
			OnHover: func() { c.setHelpText(c.modeDescriptionText("reverse", gamedata.ReverseModeCost)) },
		})
		b.GetWidget().Disabled = !xslices.Contains(playerStats.ModesUnlocked, "reverse")
		buttonsContainer.AddChild(b)
	}

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
