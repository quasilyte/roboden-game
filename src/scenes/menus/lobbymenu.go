package menus

import (
	"fmt"
	"image"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type LobbyMenuController struct {
	state *session.State

	droneButtons         []droneButton
	pointsAllocatedLabel *widget.Text

	helpPanel         *widget.Container
	helpLabel         *widget.Text
	helpIcon1         *widget.Graphic
	helpIconSeparator *widget.Text
	helpIcon2         *widget.Graphic

	recipeIcons map[gamedata.RecipeSubject]*ebiten.Image

	scene *ge.Scene
}

type droneButton struct {
	widget *eui.ItemButton
	drone  *gamedata.AgentStats
	recipe gamedata.AgentMergeRecipe
}

func NewLobbyMenuController(state *session.State) *LobbyMenuController {
	return &LobbyMenuController{state: state}
}

func (c *LobbyMenuController) Init(scene *ge.Scene) {
	c.scene = scene

	if c.state.Persistent.Settings.MusicVolumeLevel != 0 {
		scene.Audio().ContinueMusic(assets.AudioMusicTrack3)
	}

	c.prepareRecipeIcons()
	c.initUI()

	if c.state.CPUProfileWriter != nil {
		pprof.StopCPUProfile()
		if err := c.state.CPUProfileWriter.Close(); err != nil {
			panic(err)
		}
	}
	if c.state.MemProfileWriter != nil {
		pprof.WriteHeapProfile(c.state.MemProfileWriter)
		if err := c.state.MemProfileWriter.Close(); err != nil {
			panic(err)
		}
	}
}

func (c *LobbyMenuController) Update(delta float64) {
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *LobbyMenuController) prepareRecipeIcons() {
	workerImage := c.scene.LoadImage(assets.ImageWorkerAgent)
	workerFrame := workerImage.Data.SubImage(image.Rectangle{
		Max: image.Point{X: int(workerImage.DefaultFrameWidth), Y: int(workerImage.DefaultFrameHeight)},
	}).(*ebiten.Image)

	militiaImage := c.scene.LoadImage(assets.ImageMilitiaAgent)
	militiaFrame := militiaImage.Data.SubImage(image.Rectangle{
		Max: image.Point{X: int(militiaImage.DefaultFrameWidth), Y: int(militiaImage.DefaultFrameHeight)},
	}).(*ebiten.Image)

	diode := c.scene.LoadImage(assets.ImageFactionDiode).Data
	diodeSize := diode.Bounds().Size()

	// colorScale.SetColor(gamedata.FactionByTag(a.faction).Color)

	c.recipeIcons = make(map[gamedata.RecipeSubject]*ebiten.Image)
	for _, recipe := range gamedata.Tier2agentMergeRecipes {
		subjects := []gamedata.RecipeSubject{
			recipe.Drone1,
			recipe.Drone2,
		}
		for _, s := range subjects {
			if _, ok := c.recipeIcons[s]; ok {
				continue
			}

			diodeOffset := gamedata.WorkerAgentStats.DiodeOffset + 1
			droneFrame := workerFrame
			if s.Kind == gamedata.AgentMilitia {
				droneFrame = militiaFrame
				diodeOffset = gamedata.MilitiaAgentStats.DiodeOffset + 2
			}
			frameSize := droneFrame.Bounds().Size()
			img := ebiten.NewImage(32, 32)
			var drawOptions ebiten.DrawImageOptions
			drawOptions.GeoM.Scale(2, 2)
			drawOptions.GeoM.Translate(16-float64(frameSize.X), 16-float64(frameSize.Y))
			img.DrawImage(droneFrame, &drawOptions)
			drawOptions.GeoM.Reset()
			drawOptions.GeoM.Scale(2, 2)
			drawOptions.GeoM.Translate(16-float64(diodeSize.X), 16-float64(diodeSize.Y)+diodeOffset)
			drawOptions.ColorM.ScaleWithColor(gamedata.FactionByTag(s.Faction).Color)
			img.DrawImage(diode, &drawOptions)

			c.recipeIcons[s] = img
		}
	}
}

func (c *LobbyMenuController) initUI() {
	uiResources := eui.LoadResources(c.scene.Context().Loader)

	// titleLabel := eui.NewLabel(uiResources, d.Get("menu.main.title")+" -> "+d.Get("menu.main.start_game"), normalFont)
	// rowContainer.AddChild(titleLabel)

	root := eui.NewAnchorContainer()

	rootGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, false}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))

	root.AddChild(rootGrid)

	leftRows := eui.NewRowLayoutContainer(4, nil)
	rootGrid.AddChild(leftRows)
	rightRows := eui.NewRowLayoutContainer(4, []bool{false, true, false, false})
	rootGrid.AddChild(rightRows)

	leftRows.AddChild(c.createTabs(uiResources))

	rightRows.AddChild(c.createSeedPanel(uiResources))
	rightRows.AddChild(c.createHelpPanel(uiResources))
	rightRows.AddChild(c.createButtonsPanel(uiResources))

	// for i := 0; i < 2; i++ {
	// 	img := c.scene.LoadImage(assets.ImageGunpointAgent).Data
	// 	// frame := img.SubImage(image.Rectangle{
	// 	// 	Max: image.Point{X: options := &c.state.LevelOptionsonint(img.DefaultFrameWidth), Y: int(img.DefaultFrameHeight)},
	// 	// }).(*ebiten.Image)
	// 	frame := img
	// 	dronesPanel.AddChild(eui.NewItemButton(uiResources, frame, func() {}))
	// }

	// rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	// rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
	// 	c.state.LevelOptions.Tutorial = false
	// 	c.scene.Context().ChangeScene(staging.NewController(c.state, options.WorldSize, NewLobbyMenuController(c.state)))
	// }))

	// rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
	// 	c.back()
	// }))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *LobbyMenuController) createButtonsPanel(uiResources *eui.Resources) *widget.Container {
	panel := eui.NewPanel(uiResources, 0, 0)

	d := c.scene.Dict()

	options := &c.state.LevelOptions

	panel.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		c.state.LevelOptions.Tutorial = false
		c.scene.Context().ChangeScene(staging.NewController(c.state, options.WorldSize, NewLobbyMenuController(c.state)))
	}))

	panel.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	return panel
}

func (c *LobbyMenuController) createTabs(uiResources *eui.Resources) *widget.TabBook {
	tabs := []*widget.TabBookTab{}

	tabs = append(tabs, c.createColonyTab(uiResources))
	tabs = append(tabs, c.createWorldTab(uiResources))
	tabs = append(tabs, c.createDifficultyTab(uiResources))
	tabs = append(tabs, c.createExtraTab(uiResources))

	t := widget.NewTabBook(
		// widget.TabBookOpts.FlipBookOpts(
		// 	widget.FlipBookOpts.ContainerOpts(
		// 		widget.ContainerOpts.Layout(widget.NewRowLayout(
		// 			// widget.RowLayoutOpts.Padding(widget.Insets{
		// 			// 	Top: 200,
		// 			// }),
		// 			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
		// 		)),
		// 	),
		// ),
		widget.TabBookOpts.Tabs(tabs...),
		widget.TabBookOpts.TabButtonImage(uiResources.TabButton.Image),
		widget.TabBookOpts.TabButtonText(uiResources.TabButton.FontFace, uiResources.TabButton.TextColors),
		widget.TabBookOpts.TabButtonOpts(
			widget.ButtonOpts.TextPadding(uiResources.Button.Padding),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch: true,
				}),
			),
		),
		widget.TabBookOpts.TabButtonSpacing(10),
		widget.TabBookOpts.Spacing(12))

	return t
}

func (c *LobbyMenuController) createExtraTab(uiResources *eui.Resources) *widget.TabBookTab {
	d := c.scene.Dict()

	tab := widget.NewTabBookTab(d.Get("menu.lobby.tab.extra"),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	return tab
}

func (c *LobbyMenuController) createDifficultyTab(uiResources *eui.Resources) *widget.TabBookTab {
	d := c.scene.Dict()

	options := &c.state.LevelOptions

	tab := widget.NewTabBookTab(d.Get("menu.lobby.tab.difficulty"),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	{
		var slider gmath.Slider
		slider.SetBounds(0, 4)
		slider.TrySetValue(options.CreepsDifficulty)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.creeps_difficulty")+": "+strconv.Itoa(slider.Value()))
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.CreepsDifficulty = slider.Value()
			button.Text().Label = d.Get("menu.lobby.creeps_difficulty") + ": " + strconv.Itoa(slider.Value())
		})
		tab.AddChild(button)
	}

	{
		valueNames := []string{
			d.Get("menu.option.very_easy"),
			d.Get("menu.option.easy"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.hard"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(options.BossDifficulty)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.boss_difficulty")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.BossDifficulty = slider.Value()
			button.Text().Label = d.Get("menu.lobby.boss_difficulty") + ": " + valueNames[slider.Value()]
		})
		tab.AddChild(button)
	}

	{
		valueNames := []string{
			d.Get("menu.option.none"),
			d.Get("menu.option.some"),
			d.Get("menu.option.lots"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 2)
		slider.TrySetValue(options.StartingResources)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.starting_resources")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.StartingResources = slider.Value()
			button.Text().Label = d.Get("menu.lobby.starting_resources") + ": " + valueNames[slider.Value()]
		})
		tab.AddChild(button)
	}

	return tab
}

func (c *LobbyMenuController) createWorldTab(uiResources *eui.Resources) *widget.TabBookTab {
	d := c.scene.Dict()

	options := &c.state.LevelOptions

	tab := widget.NewTabBookTab(d.Get("menu.lobby.tab.world"),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	{
		valueNames := []string{
			d.Get("menu.option.very_low"),
			d.Get("menu.option.low"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.rich"),
			d.Get("menu.option.very_rich"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 4)
		slider.TrySetValue(options.Resources)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.world_resources")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.Resources = slider.Value()
			button.Text().Label = d.Get("menu.lobby.world_resources") + ": " + valueNames[slider.Value()]
		})
		tab.AddChild(button)
	}

	{
		valueNames := []string{
			d.Get("menu.option.very_small"),
			d.Get("menu.option.small"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.big"),
		}
		var slider gmath.Slider
		slider.SetBounds(0, 3)
		slider.TrySetValue(options.WorldSize)
		button := eui.NewButtonSelected(uiResources, d.Get("menu.lobby.world_size")+": "+valueNames[slider.Value()])
		button.ClickedEvent.AddHandler(func(args interface{}) {
			slider.Inc()
			options.WorldSize = slider.Value()
			button.Text().Label = d.Get("menu.lobby.world_size") + ": " + valueNames[slider.Value()]
		})
		tab.AddChild(button)
	}

	return tab
}

func (c *LobbyMenuController) createColonyTab(uiResources *eui.Resources) *widget.TabBookTab {
	d := c.scene.Dict()

	tab := widget.NewTabBookTab(d.Get("menu.lobby.tab.colony"),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	tab.AddChild(c.createBasesPanel(uiResources))
	tab.AddChild(c.createTurretsPanel(uiResources))

	label := widget.NewText(
		widget.TextOpts.Text("Points Allocated: 99/99", normalFont, uiResources.Button.TextColors.Idle),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)
	c.pointsAllocatedLabel = label
	tab.AddChild(label)

	tab.AddChild(c.createDronesPanel(uiResources))

	c.updateAllocatedPoints(c.calcAllocatedPoints())

	return tab
}

func (c *LobbyMenuController) updateAllocatedPoints(allocated int) {
	c.pointsAllocatedLabel.Label = fmt.Sprintf("Points Allocated: %d/%d", allocated, gamedata.ClassicModePoints)
}

func (c *LobbyMenuController) calcAllocatedPoints() int {
	total := 0
	for _, b := range c.droneButtons {
		if !b.widget.IsToggled() {
			continue
		}
		total += b.drone.PointCost
	}
	return total
}

func (c *LobbyMenuController) createHelpPanel(uiResources *eui.Resources) *widget.Container {
	panel := eui.NewPanel(uiResources, 0, 0)
	c.helpPanel = panel

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	label := eui.NewLabel(uiResources, "", normalFont)
	label.MaxWidth = 260
	c.helpLabel = label
	panel.AddChild(label)

	{
		iconsContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(widget.Insets{
					Top: 12,
				}),
			)),
		)

		icon1 := widget.NewGraphic()
		c.helpIcon1 = icon1
		iconsContainer.AddChild(icon1)

		separator := widget.NewText(
			widget.TextOpts.Text("", normalFont, uiResources.Button.TextColors.Idle),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		)
		c.helpIconSeparator = separator
		iconsContainer.AddChild(separator)

		icon2 := widget.NewGraphic()
		c.helpIcon2 = icon2
		iconsContainer.AddChild(icon2)

		panel.AddChild(iconsContainer)
	}

	return panel
}

func (c *LobbyMenuController) createSeedPanel(uiResources *eui.Resources) *widget.Container {
	worldSettingsPanel := eui.NewPanel(uiResources, 292, 0)

	// options := &c.state.LevelOptions

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	{
		grid := widget.NewContainer(
			widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			})),
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
				widget.GridLayoutOpts.Spacing(4, 4),
			)),
		)

		textinput := eui.NewTextInput(uiResources, normalFont, func(s string) {})
		grid.AddChild(textinput)
		label := widget.NewLabel(
			widget.LabelOpts.TextOpts(
				widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			),
			widget.LabelOpts.Text("Game Seed", normalFont, &widget.LabelColor{
				Idle:     uiResources.Button.TextColors.Idle,
				Disabled: uiResources.Button.TextColors.Disabled,
			}),
		)
		grid.AddChild(label)

		worldSettingsPanel.AddChild(grid)
	}

	return worldSettingsPanel
}

func (c *LobbyMenuController) createBasesPanel(uiResources *eui.Resources) *widget.Container {
	panel := eui.NewPanel(uiResources, 0, 0)

	grid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(6),
			widget.GridLayoutOpts.Spacing(4, 4))))

	// TODO: add bases list.
	b := eui.NewBigItemButton(uiResources, c.scene.LoadImage(assets.ImageColonyCore).Data, func() {})
	b.Toggle()
	grid.AddChild(b.Widget)

	panel.AddChild(grid)

	return panel
}

func (c *LobbyMenuController) createTurretsPanel(uiResources *eui.Resources) *widget.Container {
	panel := eui.NewPanel(uiResources, 0, 0)

	grid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(6),
			widget.GridLayoutOpts.Spacing(4, 4))))

	for i, turret := range gamedata.TurretStatsList {
		img := c.scene.LoadImage(turret.Image)
		b := eui.NewItemButton(uiResources, img.Data, "", func() {})
		grid.AddChild(b.Widget)
		if i == 0 {
			b.Toggle()
		}
	}

	panel.AddChild(grid)

	return panel
}

func (c *LobbyMenuController) createDronesPanel(uiResources *eui.Resources) *widget.Container {
	dronesPanel := eui.NewPanel(uiResources, 0, 0)

	d := c.scene.Dict()

	options := &c.state.LevelOptions

	grid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(8),
			widget.GridLayoutOpts.Spacing(4, 4))))

	maxNumDrones := 8 * 3
	for i := range gamedata.Tier2agentMergeRecipes {
		recipe := gamedata.Tier2agentMergeRecipes[i]
		img := c.scene.LoadImage(recipe.Result.Image)
		frame := img.Data.SubImage(image.Rectangle{
			Max: image.Point{X: int(img.DefaultFrameWidth), Y: int(img.DefaultFrameHeight)},
		}).(*ebiten.Image)
		costLabel := strings.Repeat(".", recipe.Result.PointCost)
		drone := recipe.Result
		var b *eui.ItemButton
		b = eui.NewItemButton(uiResources, frame, costLabel, func() {
			b.Toggle()
			c.onDroneToggled()
			c.updateTier2Recipes()
		})
		grid.AddChild(b.Widget)
		if xslices.Contains(options.Tier2Recipes, recipe) {
			b.Toggle()
		}
		c.droneButtons = append(c.droneButtons, droneButton{
			widget: b,
			drone:  drone,
			recipe: recipe,
		})
		b.Widget.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			tag := ""
			switch {
			case drone.CanGather && drone.CanPatrol:
				tag = d.Get("drone", "kind", "universal")
			case drone.CanGather:
				tag = d.Get("drone", "kind", "worker")
			case drone.CanPatrol:
				tag = d.Get("drone", "kind", "military")
			}
			key := strings.ToLower(drone.Kind.String())

			textLines := make([]string, 0, 6)

			textLines = append(textLines, d.Get("drone", key)+"\n")
			textLines = append(textLines, fmt.Sprintf("%s: %s\n", d.Get("drone.function"), tag))
			textLines = append(textLines, d.Get("drone", key, "description")+"\n")

			if drone.Weapon != nil {
				targets := ""
				switch {
				case drone.Weapon.TargetFlags&gamedata.TargetFlying != 0 && drone.Weapon.TargetFlags&gamedata.TargetGround != 0:
					targets = d.Get("drone.target.ground") + ", " + d.Get("drone.target.flying")
				case drone.Weapon.TargetFlags&gamedata.TargetFlying != 0:
					targets = d.Get("drone.target.flying")
				case drone.Weapon.TargetFlags&gamedata.TargetGround != 0:
					targets = d.Get("drone.target.ground")
				}
				textLines = append(textLines, fmt.Sprintf("%s: %s\n", d.Get("drone.target"), targets))
			}

			c.helpLabel.Label = strings.Join(textLines, "\n")

			c.helpIcon1.Image = c.recipeIcons[recipe.Drone1]
			c.helpIconSeparator.Label = "+"
			c.helpIcon2.Image = c.recipeIcons[recipe.Drone2]
			c.helpPanel.RequestRelayout()
		})
	}
	c.onDroneToggled()

	// Pad the remaining space with disabled buttons.
	for i := len(gamedata.Tier2agentMergeRecipes); i < maxNumDrones; i++ {
		b := eui.NewItemButton(uiResources, nil, "", func() {})
		b.SetDisabled(true)
		grid.AddChild(b.Widget)
	}

	dronesPanel.AddChild(grid)

	return dronesPanel
}

func (c *LobbyMenuController) updateTier2Recipes() {
	options := &c.state.LevelOptions

	options.Tier2Recipes = options.Tier2Recipes[:0]
	for _, b := range c.droneButtons {
		if !b.widget.IsToggled() {
			continue
		}
		options.Tier2Recipes = append(options.Tier2Recipes, b.recipe)
	}
}

func (c *LobbyMenuController) onDroneToggled() {
	allocated := c.calcAllocatedPoints()
	pointsLeft := gamedata.ClassicModePoints - allocated
	c.updateAllocatedPoints(allocated)
	for _, b := range c.droneButtons {
		if b.widget.IsToggled() {
			continue
		}
		b.widget.SetDisabled(b.drone.PointCost > pointsLeft)
	}
}

func (c *LobbyMenuController) back() {
	c.scene.Context().ChangeScene(NewMainMenuController(c.state))
}
