package menus

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type LeaderboardBrowserController struct {
	state *session.State

	gameMode string

	selectedSeason int

	scene *ge.Scene

	rowContainer *widget.Container

	boardData *serverapi.LeaderboardResp
	fetchErr  error
}

func NewLeaderboardBrowserController(state *session.State, season int, gameMode string, boardData *serverapi.LeaderboardResp, fetchErr error) *LeaderboardBrowserController {
	return &LeaderboardBrowserController{
		state:          state,
		gameMode:       gameMode,
		boardData:      boardData,
		fetchErr:       fetchErr,
		selectedSeason: season,
	}
}

func (c *LeaderboardBrowserController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *LeaderboardBrowserController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *LeaderboardBrowserController) initUI() {
	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	c.rowContainer = rowContainer
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.leaderboard")+" -> "+d.Get("menu.leaderboard", c.gameMode), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI
	{

		boardData := c.boardData
		fetchErr := c.fetchErr

		d := c.scene.Dict()
		smallFont := assets.BitmapFont1
		tinyFont := assets.BitmapFont1

		{
			numSeasons := c.selectedSeason + 1
			if boardData != nil {
				numSeasons = boardData.NumSeasons
			}
			seasons := make([]string, numSeasons)
			for i := range seasons {
				seasons[i] = strconv.Itoa(i)
			}
			b := eui.NewSelectButton(eui.SelectButtonConfig{
				PlaySound:  true,
				Resources:  uiResources,
				Input:      c.state.MenuInput,
				Value:      &c.selectedSeason,
				Label:      d.Get("menu.leaderboard.season"),
				ValueNames: seasons,
			})
			c.scene.AddObject(b)
			c.rowContainer.AddChild(b.Widget)
			if fetchErr != nil {
				b.Widget.GetWidget().Disabled = true
			}
			// TODO: remove this after season 0 archive is available.
			b.Widget.GetWidget().Disabled = true
		}

		if boardData != nil {
			s := fmt.Sprintf("%s: %d", d.Get("menu.leaderboard.num_players"), boardData.NumPlayers)
			c.rowContainer.AddChild(eui.NewCenteredLabel(s, smallFont))
		}

		panel := eui.NewTextPanel(uiResources, 540, 96)

		if boardData == nil {
			panel.AddChild(eui.NewCenteredLabel(d.Get("menu.leaderboard.fetch_error"), tinyFont))
		} else {
			numColumns := 6
			if c.gameMode == "arena" {
				numColumns = 5
			}

			grid := widget.NewContainer(
				widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch: true,
				})),
				widget.ContainerOpts.Layout(widget.NewGridLayout(
					widget.GridLayoutOpts.Spacing(24, 4),
					widget.GridLayoutOpts.Columns(numColumns),
					widget.GridLayoutOpts.Stretch([]bool{false, false, true, false, false, false}, nil),
				)))

			grid.AddChild(eui.NewLabel("["+d.Get("menu.leaderboard.col_rank")+"]", tinyFont))
			grid.AddChild(eui.NewLabel("["+d.Get("menu.leaderboard.col_platform")+"]", tinyFont))
			grid.AddChild(eui.NewLabel("["+d.Get("menu.leaderboard.col_name")+"]", tinyFont))
			grid.AddChild(eui.NewLabel("["+d.Get("menu.leaderboard.col_difficulty")+"]", tinyFont))
			grid.AddChild(eui.NewLabel("["+d.Get("menu.leaderboard.col_score")+"]", tinyFont))
			if c.gameMode != "arena" {
				grid.AddChild(eui.NewLabel("["+d.Get("menu.leaderboard.col_time")+"]", tinyFont))
			}

			for i := 0; i < numColumns; i++ {
				grid.AddChild(eui.NewLabel("-", tinyFont))
			}

			for _, e := range boardData.Entries {
				clr := eui.NormalTextColor
				if e.PlayerName == c.state.Persistent.PlayerName {
					clr = eui.CaretColor
				}
				d := time.Duration(e.Time) * time.Second
				grid.AddChild(eui.NewColoredLabel(strconv.Itoa(e.Rank), tinyFont, clr))
				clan := e.Platform
				if clan == "" {
					// Empty platforms is an old thing before we were sending this data to the server.
					// Assume these players to come from Steam.
					clan = "Steam"
				}
				grid.AddChild(eui.NewColoredLabel(clan, tinyFont, clr))
				grid.AddChild(eui.NewColoredLabel(e.PlayerName, tinyFont, clr))
				grid.AddChild(eui.NewColoredLabel(fmt.Sprintf("%d%%", e.Difficulty), tinyFont, clr))
				grid.AddChild(eui.NewColoredLabel(strconv.Itoa(e.Score), tinyFont, clr))
				if c.gameMode != "arena" {
					grid.AddChild(eui.NewColoredLabel(timeutil.FormatDurationCompact(d), tinyFont, clr))
				}
			}
			panel.AddChild(grid)
		}

		c.rowContainer.AddChild(panel)
	}

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	c.rowContainer.AddChild(backButton)

	navTree := createSimpleNavTree([]eui.Widget{backButton})
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *LeaderboardBrowserController) back() {
	c.scene.Context().ChangeScene(NewLeaderboardMenuController(c.state))
}
