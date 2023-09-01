package menus

import (
	"image"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/descriptions"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
)

type ProfileDroneCollectionMenuController struct {
	state *session.State

	scene *ge.Scene

	recipeIcons map[gamedata.RecipeSubject]*ebiten.Image

	helpRecipe *eui.RecipeView

	helpLabel *widget.Text
}

func NewProfileDroneCollectionMenuController(state *session.State) *ProfileDroneCollectionMenuController {
	return &ProfileDroneCollectionMenuController{state: state}
}

func (c *ProfileDroneCollectionMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.recipeIcons = gameui.GenerateRecipePreviews(c.scene, true)
	c.initUI()
}

func (c *ProfileDroneCollectionMenuController) Update(delta float64) {
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *ProfileDroneCollectionMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	tinyFont := assets.BitmapFont1

	helpLabel := eui.NewLabel("", tinyFont)
	helpLabel.MaxWidth = 340
	c.helpLabel = helpLabel

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.profile")+" -> "+d.Get("menu.profile.dronebook"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	stats := &c.state.Persistent.PlayerStats

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
			widget.GridLayoutOpts.Columns(7),
			widget.GridLayoutOpts.Spacing(4, 4))))
	droneImage := func(drone *gamedata.AgentStats, available bool) *ebiten.Image {
		if !available {
			return c.scene.LoadImage(assets.ImageLock).Data
		}
		img := c.scene.LoadImage(drone.Image)
		return img.Data.SubImage(image.Rectangle{
			Max: image.Point{X: int(img.DefaultFrameWidth), Y: int(img.DefaultFrameHeight)},
		}).(*ebiten.Image)
	}
	drones := gamedata.AllDroneStats()
	droneIsUnlocked := func(d *gamedata.AgentStats) bool {
		switch d.Tier {
		case 2:
			return xslices.Contains(stats.DronesUnlocked, d.Kind.String())
		case 3:
			return xslices.Contains(stats.Tier3DronesSeen, d.Kind.String())
		default:
			return true
		}
	}
	for i := range drones {
		drone := drones[i]
		available := droneIsUnlocked(drone)
		frame := droneImage(drone, available)
		b := eui.NewItemButton(uiResources, frame, tinyFont, "", 0, func() {})
		b.SetDisabled(true)
		b.Widget.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			if available {
				c.helpLabel.Label = descriptions.DroneText(c.scene.Dict(), drone, true, true)
				if drone.Tier == 1 {
					c.helpRecipe.SetImages(nil, nil)
				} else {
					recipe := gamedata.FindRecipeByName(drone.Kind.String())
					c.helpRecipe.SetImages(c.recipeIcons[recipe.Drone1], c.recipeIcons[recipe.Drone2])
				}
			} else if drone.Tier == 2 {
				c.helpLabel.Label = descriptions.LockedDroneText(c.scene.Dict(), &c.state.Persistent.PlayerStats, drone)
				c.helpRecipe.SetImages(nil, nil)
			} else {
				c.helpLabel.Label = d.Get("drone.undiscovered")
				c.helpRecipe.SetImages(nil, nil)
			}
		})
		leftGrid.AddChild(b.Widget)
	}
	leftPanel.AddChild(leftGrid)

	rightPanel := eui.NewTextPanel(uiResources, 380, 0)
	rightPanel.AddChild(helpLabel)

	c.helpRecipe = eui.NewRecipeView(uiResources)
	rightPanel.AddChild(c.helpRecipe.Container)

	rootGrid.AddChild(leftPanel)
	rootGrid.AddChild(rightPanel)

	rowContainer.AddChild(rootGrid)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *ProfileDroneCollectionMenuController) back() {
	c.scene.Context().ChangeScene(NewProfileMenuController(c.state))
}
