package menus

import (
	"fmt"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ProfileAchievementsMenuController struct {
	state *session.State

	scene *ge.Scene
}

func NewProfileAchievementsMenuController(state *session.State) *ProfileAchievementsMenuController {
	return &ProfileAchievementsMenuController{state: state}
}

func (c *ProfileAchievementsMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *ProfileAchievementsMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *ProfileAchievementsMenuController) paintIcon(icon *ebiten.Image) *ebiten.Image {
	painted := ebiten.NewImage(icon.Size())
	var options ebiten.DrawImageOptions
	options.ColorM.Scale(0, 0, 0, 1)
	painted.DrawImage(icon, &options)
	return painted
}

func (c *ProfileAchievementsMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	helpLabel := eui.NewLabel(uiResources, "", tinyFont)
	helpLabel.MaxWidth = 320

	titleLabel := eui.NewCenteredLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.achievements"), normalFont)
	rowContainer.AddChild(titleLabel)

	rootGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))
	leftPanel := eui.NewPanel(uiResources, 0, 0)
	leftGrid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(6),
			widget.GridLayoutOpts.Spacing(4, 4))))
	for i := range gamedata.AchievementList {
		achievement := gamedata.AchievementList[i]
		status := xslices.Find(c.state.Persistent.PlayerStats.Achievements, func(a *session.Achievement) bool {
			return a.Name == achievement.Name
		})
		grade := 0
		img := c.scene.LoadImage(assets.ImageAchievementImpossible).Data
		if status != nil {
			img = c.paintIcon(img)
			if status.Elite {
				grade = 2
			} else {
				grade = 1
			}
		}
		b := eui.NewItemButton(uiResources, img, strings.Repeat(".", grade), func() {})
		b.SetDisabled(true)
		b.Widget.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			var lines []string
			statusText := "(not achived yet)"
			switch grade {
			case 1:
				statusText = "(normal grade)"
			case 2:
				statusText = "(elite grade)"
			}
			lines = append(lines, fmt.Sprintf("%s %s", d.Get("achievement", achievement.Name), statusText))
			lines = append(lines, "")
			lines = append(lines, d.Get("achievement", achievement.Name, "description"))
			helpLabel.Label = strings.Join(lines, "\n")
		})
		if status != nil {
			b.Toggle()
		}
		leftGrid.AddChild(b.Widget)
	}
	leftPanel.AddChild(leftGrid)

	rightPanel := eui.NewPanel(uiResources, 352, 0)
	rightPanel.AddChild(helpLabel)

	rootGrid.AddChild(leftPanel)
	rootGrid.AddChild(rightPanel)

	rowContainer.AddChild(rootGrid)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, "Back", func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileAchievementsMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
