package menus

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/descriptions"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type ReplayMenuController struct {
	state *session.State

	helpLabel *widget.Text

	scene *ge.Scene
}

func NewReplayMenuController(state *session.State) *ReplayMenuController {
	return &ReplayMenuController{state: state}
}

func (c *ReplayMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ReplayMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *ReplayMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	smallFont := assets.Font1

	helpLabel := eui.NewLabel("", smallFont)
	helpLabel.MaxWidth = 268
	c.helpLabel = helpLabel

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})

	navTree := gameui.NewNavTree()
	bottomNavBlock := navTree.NewBlock()
	leftNavBlock := navTree.NewBlock()
	rightNavBlock := navTree.NewBlock()
	var leftButtonElems []*gameui.NavElem
	var rightButtonElems []*gameui.NavElem

	bottomNavBlock.NewElem(backButton)

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.watch_replay"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	rootGrid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))
	leftGrid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(8, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))

	for i := 0; i < 10; i++ {
		key := c.state.ReplayDataKey(i)
		replayExists := c.state.CheckGameItem(key)
		var r session.SavedReplay
		if replayExists {
			if err := c.state.LoadGameItem(key, &r); err != nil {
				replayExists = false
			}
			if !gamedata.IsRunnableReplay(r.Replay) {
				replayExists = false
			}
		}
		label := d.Get("menu.replay.empty")
		if replayExists {
			if i == 0 {
				label = d.Get("menu.replay.last_played")
			} else {
				label = fmt.Sprintf("[%d] %s", i, timeutil.FormatDateISO8601(r.Date, true))
			}
		}
		b := eui.NewSmallButton(uiResources, c.scene, label, func() {
			config := gamedata.MakeLevelConfig(gamedata.ExecuteReplay, r.Replay.Config)
			config.Finalize()
			controller := staging.NewController(c.state, config, NewReplayMenuController(c.state))
			controller.SetReplayActions(r.Replay)
			c.scene.Context().ChangeScene(controller)
		})
		if replayExists {
			b.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
				c.helpLabel.Label = descriptions.ReplayText(d, &r)
			})
		}
		b.GetWidget().Disabled = !replayExists || r.Replay.GameVersion != gamedata.BuildNumber
		b.GetWidget().MinWidth = 220
		leftGrid.AddChild(b)
		if i%2 == 0 {
			leftButtonElems = append(leftButtonElems, leftNavBlock.NewElem(b))
		} else {
			rightButtonElems = append(rightButtonElems, rightNavBlock.NewElem(b))
		}
	}

	rightPanel := eui.NewTextPanel(uiResources, 320, 0)
	rightPanel.AddChild(helpLabel)

	rootGrid.AddChild(leftGrid)
	rootGrid.AddChild(rightPanel)

	rowContainer.AddChild(rootGrid)

	rowContainer.AddChild(eui.NewCenteredLabel(d.Get("menu.replay.notice"), smallFont))

	rowContainer.AddChild(eui.NewTransparentSeparator())

	rowContainer.AddChild(backButton)

	bindNavListNoWrap(leftButtonElems, gameui.NavUp, gameui.NavDown)
	bindNavListNoWrap(rightButtonElems, gameui.NavUp, gameui.NavDown)
	bottomNavBlock.Edges[gameui.NavUp] = leftNavBlock
	leftNavBlock.Edges[gameui.NavDown] = bottomNavBlock
	leftNavBlock.Edges[gameui.NavRight] = rightNavBlock
	rightNavBlock.Edges[gameui.NavDown] = bottomNavBlock
	rightNavBlock.Edges[gameui.NavLeft] = leftNavBlock

	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *ReplayMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
