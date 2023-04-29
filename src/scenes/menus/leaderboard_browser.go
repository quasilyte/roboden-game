package menus

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/clientkit"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/gtask"
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
	placeholder  *widget.Text
}

func NewLeaderboardBrowserController(state *session.State, gameMode string) *LeaderboardBrowserController {
	return &LeaderboardBrowserController{
		state:    state,
		gameMode: gameMode,
	}
}

func (c *LeaderboardBrowserController) Init(scene *ge.Scene) {
	c.scene = scene
	c.selectedSeason = gamedata.SeasonNumber
	c.initUI()
}

func (c *LeaderboardBrowserController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *LeaderboardBrowserController) getBoardCache() *serverapi.LeaderboardResp {
	switch c.gameMode {
	case "classic":
		return &c.state.Persistent.CachedClassicLeaderboard
	case "arena":
		return &c.state.Persistent.CachedArenaLeaderboard
	case "inf_arena":
		return &c.state.Persistent.CachedInfArenaLeaderboard
	default:
		return nil
	}
}

func (c *LeaderboardBrowserController) initBoard(boardData *serverapi.LeaderboardResp, fetchErr error) {
	uiResources := c.state.Resources.UI

	d := c.scene.Dict()
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	c.rowContainer.RemoveChild(c.placeholder)

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
			Scene:      c.scene,
			Resources:  uiResources,
			Input:      c.state.MainInput,
			Value:      &c.selectedSeason,
			Label:      d.Get("menu.leaderboard.season"),
			ValueNames: seasons,
		})
		c.rowContainer.AddChild(b)
		if fetchErr != nil {
			b.GetWidget().Disabled = true
		}
	}

	if boardData != nil {
		s := fmt.Sprintf("%s: %d", d.Get("menu.leaderboard.num_players"), boardData.NumPlayers)
		c.rowContainer.AddChild(eui.NewCenteredLabel(s, smallFont))
	}

	panel := eui.NewPanel(uiResources, 0, 96)
	c.rowContainer.AddChild(panel)

	if boardData == nil {
		panel.AddChild(eui.NewCenteredLabel(d.Get("menu.leaderboard.fetch_error"), tinyFont))
	} else {
		numColumns := 5
		if c.gameMode == "arena" {
			numColumns = 4
		}

		grid := widget.NewContainer(
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Spacing(24, 4),
				widget.GridLayoutOpts.Columns(numColumns),
				widget.GridLayoutOpts.Stretch([]bool{false, true, false, false, false}, nil),
			)))

		grid.AddChild(eui.NewLabel("[rank]", tinyFont))
		grid.AddChild(eui.NewLabel("[name]", tinyFont))
		grid.AddChild(eui.NewLabel("[difficulty]", tinyFont))
		grid.AddChild(eui.NewLabel("[score]", tinyFont))
		if c.gameMode != "arena" {
			grid.AddChild(eui.NewLabel("[time]", tinyFont))
		}

		grid.AddChild(eui.NewLabel("-", tinyFont))
		grid.AddChild(eui.NewLabel("-", tinyFont))
		grid.AddChild(eui.NewLabel("-", tinyFont))
		grid.AddChild(eui.NewLabel("-", tinyFont))
		if c.gameMode != "arena" {
			grid.AddChild(eui.NewLabel("-", tinyFont))
		}

		for _, e := range boardData.Entries {
			clr := eui.NormalTextColor
			if e.PlayerName == c.state.Persistent.PlayerName {
				clr = eui.CaretColor
			}
			d := time.Duration(e.Time) * time.Second
			grid.AddChild(eui.NewColoredLabel(strconv.Itoa(e.Rank), tinyFont, clr))
			grid.AddChild(eui.NewColoredLabel(e.PlayerName, tinyFont, clr))
			grid.AddChild(eui.NewColoredLabel(fmt.Sprintf("%d%%", e.Difficulty), tinyFont, clr))
			grid.AddChild(eui.NewColoredLabel(strconv.Itoa(e.Score), tinyFont, clr))
			if c.gameMode != "arena" {
				grid.AddChild(eui.NewColoredLabel(timeutil.FormatDurationCompact(d), tinyFont, clr))
			}
		}
		panel.AddChild(grid)
	}

	c.rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))
}

func (c *LeaderboardBrowserController) initUI() {
	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	c.rowContainer = rowContainer
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.leaderboard")+" -> "+d.Get("menu.leaderboard", c.gameMode), normalFont)
	rowContainer.AddChild(titleLabel)

	c.placeholder = eui.NewCenteredLabel(d.Get("menu.leaderboard.placeholder"), tinyFont)
	rowContainer.AddChild(c.placeholder)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)

	var boardData *serverapi.LeaderboardResp
	var fetchErr error
	fetchTask := gtask.StartTask(func(ctx *gtask.TaskContext) {
		boardData, fetchErr = clientkit.GetLeaderboard(c.state, c.gameMode)
		if fetchErr != nil {
			// Try using the cached data.
			cached := c.getBoardCache()
			if len(cached.Entries) != 0 {
				boardData = cached
			}
		} else {
			// Save fetched data to the cache.
			*c.getBoardCache() = *boardData
			c.scene.Context().SaveGameData("save", c.state.Persistent)
		}
	})
	fetchTask.EventCompleted.Connect(nil, func(gsignal.Void) {
		c.initBoard(boardData, fetchErr)
		root.RequestRelayout()
	})
	c.scene.AddObject(fetchTask)
}

func (c *LeaderboardBrowserController) back() {
	c.scene.Context().ChangeScene(NewLeaderboardMenuController(c.state))
}
