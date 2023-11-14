package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/clientkit"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/gtask"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

type LeaderboardLoadingController struct {
	state *session.State

	gameMode string

	selectedSeason int

	scene *ge.Scene

	rowContainer *widget.Container
	placeholder  *widget.Text
}

func NewLeaderboardLoadingController(state *session.State, season int, gameMode string) *LeaderboardLoadingController {
	return &LeaderboardLoadingController{
		state:          state,
		gameMode:       gameMode,
		selectedSeason: season,
	}
}

func (c *LeaderboardLoadingController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *LeaderboardLoadingController) Update(delta float64) {}

func (c *LeaderboardLoadingController) getBoardCache() *serverapi.LeaderboardResp {
	switch c.gameMode {
	case "blitz":
		return &c.state.Persistent.CachedBlitzLeaderboard
	case "classic":
		return &c.state.Persistent.CachedClassicLeaderboard
	case "arena":
		return &c.state.Persistent.CachedArenaLeaderboard
	case "inf_arena":
		return &c.state.Persistent.CachedInfArenaLeaderboard
	case "reverse":
		return &c.state.Persistent.CachedReverseLeaderboard
	default:
		return nil
	}
}

func (c *LeaderboardLoadingController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	c.rowContainer = rowContainer
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	tinyFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.leaderboard")+" -> "+d.Get("menu.leaderboard", c.gameMode), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	c.placeholder = eui.NewCenteredLabel(d.Get("menu.leaderboard.placeholder"), tinyFont)
	rowContainer.AddChild(c.placeholder)

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)

	var boardData *serverapi.LeaderboardResp
	var fetchErr error
	fetchTask := gtask.StartTask(func(ctx *gtask.TaskContext) {
		boardData, fetchErr = clientkit.GetLeaderboard(c.state, c.selectedSeason, c.gameMode)
		if fetchErr != nil {
			// Try using the cached data.
			cached := c.getBoardCache()
			if cached.NumSeasons != 0 {
				boardData = cached
			}
		} else {
			// Save fetched data to the cache.
			*c.getBoardCache() = *boardData
			c.state.SaveGameItem("save.json", c.state.Persistent)
		}
	})
	fetchTask.EventCompleted.Connect(nil, func(gsignal.Void) {
		controller := NewLeaderboardBrowserController(c.state, c.selectedSeason, c.gameMode, boardData, fetchErr)
		c.scene.Context().ChangeScene(controller)
	})
	c.scene.AddObject(fetchTask)
}
