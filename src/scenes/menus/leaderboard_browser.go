package menus

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/httpfetch"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

type LeaderboardBrowserController struct {
	state *session.State

	gameMode string

	scene *ge.Scene
}

func NewLeaderboardBrowserController(state *session.State, gameMode string) *LeaderboardBrowserController {
	return &LeaderboardBrowserController{
		state:    state,
		gameMode: gameMode,
	}
}

func (c *LeaderboardBrowserController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *LeaderboardBrowserController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *LeaderboardBrowserController) initUI() {
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.leaderboard")+" -> "+d.Get("menu.leaderboard", c.gameMode), normalFont)
	rowContainer.AddChild(titleLabel)

	rowContainer.AddChild(eui.NewCenteredLabel(d.Get("menu.leaderboard.season")+" "+strconv.Itoa(seasonNumber), smallFont))

	panel := eui.NewPanel(uiResources, 0, 96)

	boardEntries, err := c.getBoardData()
	if err != nil {
		panel.AddChild(eui.NewCenteredLabel(d.Get("menu.leaderboard.fetch_error"), tinyFont))
	} else {
		grid := widget.NewContainer(
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Spacing(24, 4),
				widget.GridLayoutOpts.Columns(4),
				widget.GridLayoutOpts.Stretch([]bool{false, true, false, false}, nil),
			)))

		grid.AddChild(eui.NewLabel("[rank]", tinyFont))
		grid.AddChild(eui.NewLabel("[name]", tinyFont))
		grid.AddChild(eui.NewLabel("[difficulty]", tinyFont))
		grid.AddChild(eui.NewLabel("[score]", tinyFont))

		grid.AddChild(eui.NewLabel("-", tinyFont))
		grid.AddChild(eui.NewLabel("-", tinyFont))
		grid.AddChild(eui.NewLabel("-", tinyFont))
		grid.AddChild(eui.NewLabel("-", tinyFont))

		for _, e := range boardEntries {
			clr := eui.NormalTextColor
			if e.PlayerName == c.state.Persistent.PlayerName {
				clr = eui.CaretColor
			}
			grid.AddChild(eui.NewColoredLabel(strconv.Itoa(e.Rank), tinyFont, clr))
			grid.AddChild(eui.NewColoredLabel(e.PlayerName, tinyFont, clr))
			grid.AddChild(eui.NewColoredLabel(fmt.Sprintf("%d%%", e.Difficulty), tinyFont, clr))
			grid.AddChild(eui.NewColoredLabel(strconv.Itoa(e.Score), tinyFont, clr))
		}
		panel.AddChild(grid)
	}

	rowContainer.AddChild(panel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *LeaderboardBrowserController) getBoardData() ([]serverapi.LeaderboardEntry, error) {
	var u url.URL
	u.Host = c.state.ServerAddress
	u.Scheme = "http"
	u.Path = "roboden/api/get-player-board"
	q := u.Query()
	q.Add("season", strconv.Itoa(seasonNumber))
	q.Add("mode", c.gameMode)
	q.Add("name", c.state.Persistent.PlayerName)
	u.RawQuery = q.Encode()
	fmt.Println(u.String())
	data, err := httpfetch.GetBytes(u.String())
	if err != nil {
		return nil, err
	}
	var entries []serverapi.LeaderboardEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *LeaderboardBrowserController) back() {
	c.scene.Context().ChangeScene(NewLeaderboardMenuController(c.state))
}
