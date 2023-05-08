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
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/descriptions"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type LobbyMenuController struct {
	state *session.State

	config gamedata.LevelConfig
	mode   gamedata.Mode

	droneButtons         []droneButton
	turretButtons        []droneButton
	pointsAllocatedLabel *widget.Text
	difficultyLabel      *widget.Text

	seedInput *widget.TextInput

	helpPanel         *widget.Container
	helpLabel         *widget.Text
	helpIcon1         *widget.Graphic
	helpIconSeparator *widget.Text
	helpIcon2         *widget.Graphic

	recipeIcons map[gamedata.RecipeSubject]*ebiten.Image

	scene *ge.Scene
}

type droneButton struct {
	widget    *eui.ItemButton
	drone     *gamedata.AgentStats
	recipe    gamedata.AgentMergeRecipe
	available bool
}

func NewLobbyMenuController(state *session.State, mode gamedata.Mode) *LobbyMenuController {
	return &LobbyMenuController{
		state: state,
		mode:  mode,
	}
}

func (c *LobbyMenuController) Init(scene *ge.Scene) {
	c.scene = scene

	switch c.mode {
	case gamedata.ModeArena:
		c.config = c.state.ArenaLevelConfig.Clone()
	default:
		c.config = c.state.LevelConfig.Clone()
	}
	c.config.Tutorial = nil

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

	scoutImage := c.scene.LoadImage(assets.ImageScoutAgent)
	scoutFrame := scoutImage.Data.SubImage(image.Rectangle{
		Max: image.Point{X: int(scoutImage.DefaultFrameWidth), Y: int(scoutImage.DefaultFrameHeight)},
	}).(*ebiten.Image)

	diode := c.scene.LoadImage(assets.ImageFactionDiode).Data
	diodeSize := diode.Bounds().Size()

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
			if s.Kind == gamedata.AgentScout {
				droneFrame = scoutFrame
				diodeOffset = gamedata.ScoutAgentStats.DiodeOffset + 2
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
	uiResources := c.state.Resources.UI

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

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)

	c.updateDifficultyScore(c.calcDifficultyScore())
}

func (c *LobbyMenuController) saveConfig() {
	clonedConfig := c.config.Clone()
	switch c.mode {
	case gamedata.ModeArena:
		c.state.ArenaLevelConfig = &clonedConfig
	default:
		c.state.LevelConfig = &clonedConfig
	}
}

func (c *LobbyMenuController) createButtonsPanel(uiResources *eui.Resources) *widget.Container {
	panel := eui.NewPanel(uiResources, 0, 0)

	d := c.scene.Dict()

	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	c.difficultyLabel = eui.NewCenteredLabel("Difficulty: 1000%", tinyFont)
	panel.AddChild(c.difficultyLabel)

	panel.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		c.saveConfig()

		c.config.GameMode = c.mode
		c.config.DifficultyScore = c.calcDifficultyScore()
		c.config.DronePointsAllocated = c.calcAllocatedPoints()
		if c.seedInput.InputText != "" {
			seed, err := strconv.ParseInt(c.seedInput.InputText, 10, 64)
			if err != nil {
				panic(err)
			}
			c.config.Seed = seed
		} else {
			c.config.Seed = c.randomSeed()
		}

		var seenFlag *bool
		switch c.mode {
		case gamedata.ModeClassic:
			seenFlag = &c.state.Persistent.SeenClassicMode
		case gamedata.ModeArena:
			seenFlag = &c.state.Persistent.SeenArenaMode
		}
		if !*seenFlag {
			*seenFlag = true
			c.scene.Context().SaveGameData("save", c.state.Persistent)
			c.scene.Context().ChangeScene(NewHintScreen(c.state, c.config.Clone(), NewLobbyMenuController(c.state, c.mode)))
		} else {
			// replayData, err := os.ReadFile("./replay.json")
			// if err != nil {
			// 	panic(err)
			// }
			// var replay serverapi.GameReplay
			// if err := json.Unmarshal(replayData, &replay); err != nil {
			// 	panic(err)
			// }
			// config := gamedata.LevelConfig{
			// 	ReplayLevelConfig: replay.Config,
			// 	GameMode: gamedata.ModeArena,
			// 	ExecMode: gamedata.ExecuteReplay,
			// 	AttackActionAvailable: true,
			// 	BuildTurretActionAvailable: true,
			// 	RadiusActionAvailable: true,
			// 	EliteResources: true,
			// 	EnemyBoss: replay.Config.RawGameMode == "classic",
			// }
			// controller := staging.NewController(c.state, c.config.Clone(), NewLobbyMenuController(c.state, c.mode))
			// controller.SetReplayActions(replay.Actions)
			// c.scene.Context().ChangeScene(controller)

			c.scene.Context().ChangeScene(staging.NewController(c.state, c.config.Clone(), NewLobbyMenuController(c.state, c.mode)))
		}
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

	{
		b := c.newBoolOptionButton(&c.config.ExtraUI, "menu.lobby.ui_mode", []string{
			d.Get("menu.lobby.ui_immersive"),
			d.Get("menu.lobby.ui_informative"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.Teleporters, "menu.lobby.num_teleporters", []string{
			"0",
			"1",
			"2",
		})
		tab.AddChild(b)
	}

	{
		b := c.newBoolOptionButton(&c.config.FogOfWar, "menu.lobby.fog_of_war", []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		})
		tab.AddChild(b)
	}

	if c.mode == gamedata.ModeArena {
		b := c.newBoolOptionButton(&c.config.InfiniteMode, "menu.lobby.infinite_mode", []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.GameSpeed, "menu.lobby.game_speed", []string{
			"x1",
			"x1.2",
			"x1.5",
		})
		tab.AddChild(b)
	}

	return tab
}

func (c *LobbyMenuController) createDifficultyTab(uiResources *eui.Resources) *widget.TabBookTab {
	d := c.scene.Dict()

	tab := widget.NewTabBookTab(d.Get("menu.lobby.tab.difficulty"),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	if c.mode == gamedata.ModeClassic {
		b := c.newOptionButton(&c.config.NumCreepBases, "menu.lobby.num_creep_bases", []string{
			"0",
			"1",
			"2",
			"3",
			"4",
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.InitialCreeps, "menu.lobby.initial_creeps", []string{
			d.Get("menu.option.none"),
			d.Get("menu.option.some"),
			d.Get("menu.option.lots"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.CreepDifficulty, "menu.lobby.creeps_difficulty", []string{
			"90%",
			"100%",
			"110%",
			"120%",
			"130%",
			"140%",
			"150%",
			"160%",
		})
		tab.AddChild(b)
	}

	if c.mode == gamedata.ModeClassic {
		tab.AddChild(c.newOptionButton(&c.config.CreepSpawnRate, "menu.lobby.creep_spawn_rate", []string{
			"75%",
			"100%",
			"125%",
			"150%",
		}))

		tab.AddChild(c.newOptionButton(&c.config.BossDifficulty, "menu.lobby.boss_difficulty", []string{
			d.Get("menu.option.easy"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.hard"),
			d.Get("menu.option.very_hard"),
		}))

		if c.state.Persistent.PlayerStats.TotalScore >= gamedata.ArenaModeCost {
			tab.AddChild(c.newBoolOptionButton(&c.config.SuperCreeps, "menu.lobby.super_creeps", []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			}))
		}
	}

	if c.mode == gamedata.ModeArena {
		b := c.newOptionButton(&c.config.ArenaProgression, "menu.lobby.arena_progression", []string{
			"80%",
			"100%",
			"120%",
			"140%",
			"160%",
			"180%",
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.StartingResources, "menu.lobby.starting_resources", []string{
			d.Get("menu.option.none"),
			d.Get("menu.option.some"),
			d.Get("menu.option.lots"),
		})
		tab.AddChild(b)
	}

	return tab
}

func (c *LobbyMenuController) optionDescriptionText(key string) string {
	d := c.scene.Dict()
	return fmt.Sprintf("%s\n\n%s", d.Get(key), d.Get(key, "description"))
}

func (c *LobbyMenuController) newBoolOptionButton(value *bool, key string, valueNames []string) widget.PreferredSizeLocateableWidget {
	return eui.NewBoolSelectButton(eui.BoolSelectButtonConfig{
		Scene:      c.scene,
		Resources:  c.state.Resources.UI,
		Value:      value,
		Label:      c.scene.Dict().Get(key),
		ValueNames: valueNames,
		OnPressed: func() {
			c.updateDifficultyScore(c.calcDifficultyScore())
		},
		OnHover: func() {
			c.setHelpText(c.optionDescriptionText(key))
		},
	})
}

func (c *LobbyMenuController) newOptionButton(value *int, key string, valueNames []string) widget.PreferredSizeLocateableWidget {
	return eui.NewSelectButton(eui.SelectButtonConfig{
		Scene:      c.scene,
		Resources:  c.state.Resources.UI,
		Input:      c.state.MainInput,
		Value:      value,
		Label:      c.scene.Dict().Get(key),
		ValueNames: valueNames,
		OnPressed: func() {
			c.updateDifficultyScore(c.calcDifficultyScore())
		},
		OnHover: func() {
			c.setHelpText(c.optionDescriptionText(key))
		},
	})
}

func (c *LobbyMenuController) createWorldTab(uiResources *eui.Resources) *widget.TabBookTab {
	d := c.scene.Dict()

	tab := widget.NewTabBookTab(d.Get("menu.lobby.tab.world"),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
		widget.ContainerOpts.AutoDisableChildren(),
	)

	{
		b := c.newOptionButton(&c.config.Resources, "menu.lobby.world_resources", []string{
			d.Get("menu.option.very_low"),
			d.Get("menu.option.low"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.rich"),
			d.Get("menu.option.very_rich"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.WorldSize, "menu.lobby.world_size", []string{
			d.Get("menu.option.very_small"),
			d.Get("menu.option.small"),
			d.Get("menu.option.normal"),
			d.Get("menu.option.big"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.OilRegenRate, "menu.lobby.oil_regen_rate", []string{
			"0%",
			"50%",
			"100%",
			"150%",
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.Terrain, "menu.lobby.land", []string{
			d.Get("menu.lobby.land_flat"),
			d.Get("menu.lobby.land_normal"),
			d.Get("menu.lobby.land_mountains"),
		})
		tab.AddChild(b)
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

	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	tab.AddChild(c.createBasesPanel(uiResources))
	tab.AddChild(c.createTurretsPanel(uiResources))

	label := widget.NewText(
		widget.TextOpts.Text("Points Allocated: 99/99", tinyFont, uiResources.Button.TextColors.Idle),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)
	c.pointsAllocatedLabel = label
	tab.AddChild(label)

	tab.AddChild(c.createDronesPanel(uiResources))

	c.updateAllocatedPoints(c.calcAllocatedPoints())

	return tab
}

func (c *LobbyMenuController) calcDifficultyScore() int {
	switch c.mode {
	case gamedata.ModeArena:
		if c.config.InfiniteMode {
			c.config.RawGameMode = "inf_arena"
		} else {
			c.config.RawGameMode = "arena"
		}
	case gamedata.ModeClassic:
		c.config.RawGameMode = "classic"
	}
	return gamedata.CalcDifficultyScore(c.config.ReplayLevelConfig, c.calcAllocatedPoints())
}

func (c *LobbyMenuController) updateAllocatedPoints(allocated int) {
	c.pointsAllocatedLabel.Label = fmt.Sprintf("%s: %d/%d", c.scene.Dict().Get("menu.lobby.points_allocated"), allocated, gamedata.ClassicModePoints)
}

func (c *LobbyMenuController) updateDifficultyScore(score int) {
	d := c.scene.Dict()
	var tag string
	switch {
	case score < 40:
		tag = d.Get("menu.option.very_easy")
	case score < 80:
		tag = d.Get("menu.option.easy")
	case score < 120:
		tag = d.Get("menu.option.normal")
	case score < 160:
		tag = d.Get("menu.option.hard")
	case score < 200:
		tag = d.Get("menu.option.very_hard")
	default:
		tag = d.Get("menu.option.impossible")
	}
	c.difficultyLabel.Label = fmt.Sprintf("%s: %d%% (%s)", c.scene.Dict().Get("menu.lobby.tab.difficulty"), score, tag)
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

	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	label := eui.NewLabel("", tinyFont)
	label.MaxWidth = 320
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
			widget.TextOpts.Text("", tinyFont, uiResources.Button.TextColors.Idle),
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

func (c *LobbyMenuController) randomSeed() int64 {
	return int64(c.scene.Rand().IntRange(0, 1e15-1))
}

func (c *LobbyMenuController) createSeedPanel(uiResources *eui.Resources) *widget.Container {
	worldSettingsPanel := eui.NewPanel(uiResources, 340, 0)

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	d := c.scene.Dict()

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

		textinput := eui.NewTextInput(uiResources, normalFont,
			widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
				if len(newInputText) > 15 {
					return false, nil
				}
				onlyDigits := true
				for _, ch := range newInputText {
					if ch >= '0' && ch <= '9' {
						continue
					}
					onlyDigits = false
					break
				}
				return onlyDigits, nil
			}))
		textinput.InputText = strconv.FormatInt(c.randomSeed(), 10)
		c.seedInput = textinput
		grid.AddChild(textinput)
		label := widget.NewLabel(
			widget.LabelOpts.TextOpts(
				widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			),
			widget.LabelOpts.Text(d.Get("menu.lobby.game_seed"), normalFont, &widget.LabelColor{
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

func (c *LobbyMenuController) setHelpText(s string) {
	c.helpLabel.Label = s
	c.helpIcon1.Image = nil
	c.helpIconSeparator.Label = ""
	c.helpIcon2.Image = nil
	c.helpPanel.RequestRelayout()
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

	for i := range gamedata.TurretStatsList {
		turret := gamedata.TurretStatsList[i]
		available := xslices.Contains(c.state.Persistent.PlayerStats.TurretsUnlocked, turret.Kind.String())
		var img *ebiten.Image
		if available {
			img = c.scene.LoadImage(turret.Image).Data
		} else {
			img = c.scene.LoadImage(assets.ImageLock).Data
		}
		var b *eui.ItemButton
		b = eui.NewItemButton(uiResources, img, nil, "", func() {
			if c.config.TurretDesign != turret.Kind.String() {
				b.Toggle()
				c.onTurretToggled(turret)
			}
		})
		b.SetDisabled(!available)
		b.Widget.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			var s string
			if available {
				s = descriptions.TurretText(c.scene.Dict(), turret)
			} else {
				s = descriptions.LockedTurretText(c.scene.Dict(), &c.state.Persistent.PlayerStats, turret)
			}
			c.setHelpText(s)
		})
		c.turretButtons = append(c.turretButtons, droneButton{
			widget: b,
			drone:  turret,
		})
		grid.AddChild(b.Widget)
		if c.config.TurretDesign == turret.Kind.String() {
			b.Toggle()
		}
	}

	panel.AddChild(grid)

	return panel
}

func (c *LobbyMenuController) createDronesPanel(uiResources *eui.Resources) *widget.Container {
	dronesPanel := eui.NewPanel(uiResources, 0, 0)

	smallFont := c.scene.Context().Loader.LoadFont(assets.FontSmall).Face

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
		drone := recipe.Result
		available := xslices.Contains(c.state.Persistent.PlayerStats.DronesUnlocked, drone.Kind.String())
		costLabel := ""
		var frame *ebiten.Image
		if available {
			costLabel = strings.Repeat(".", recipe.Result.PointCost)
			img := c.scene.LoadImage(recipe.Result.Image)
			frame = img.Data.SubImage(image.Rectangle{
				Max: image.Point{X: int(img.DefaultFrameWidth), Y: int(img.DefaultFrameHeight)},
			}).(*ebiten.Image)
		} else {
			frame = c.scene.LoadImage(assets.ImageLock).Data
		}
		var b *eui.ItemButton
		b = eui.NewItemButton(uiResources, frame, smallFont, costLabel, func() {
			b.Toggle()
			c.onDroneToggled()
			c.updateTier2Recipes()
		})
		grid.AddChild(b.Widget)
		if xslices.Contains(c.config.Tier2Recipes, recipe.Result.Kind.String()) {
			b.Toggle()
		}
		c.droneButtons = append(c.droneButtons, droneButton{
			widget:    b,
			drone:     drone,
			recipe:    recipe,
			available: available,
		})
		b.Widget.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			if available {
				c.helpLabel.Label = descriptions.DroneText(c.scene.Dict(), drone, false)
			} else {
				c.helpLabel.Label = descriptions.LockedDroneText(c.scene.Dict(), &c.state.Persistent.PlayerStats, drone)
			}
			if available {
				c.helpIcon1.Image = c.recipeIcons[recipe.Drone1]
				c.helpIconSeparator.Label = "+"
				c.helpIcon2.Image = c.recipeIcons[recipe.Drone2]
			} else {
				c.helpIcon1.Image = nil
				c.helpIconSeparator.Label = ""
				c.helpIcon2.Image = nil
			}
			c.helpPanel.RequestRelayout()
		})
	}
	c.onDroneToggled()

	// Pad the remaining space with disabled buttons.
	for i := len(gamedata.Tier2agentMergeRecipes); i < maxNumDrones; i++ {
		b := eui.NewItemButton(uiResources, nil, nil, "", func() {})
		b.SetDisabled(true)
		grid.AddChild(b.Widget)
	}

	dronesPanel.AddChild(grid)

	return dronesPanel
}

func (c *LobbyMenuController) updateTier2Recipes() {
	c.config.Tier2Recipes = c.config.Tier2Recipes[:0]
	for _, b := range c.droneButtons {
		if !b.widget.IsToggled() {
			continue
		}
		c.config.Tier2Recipes = append(c.config.Tier2Recipes, b.recipe.Result.Kind.String())
	}
}

func (c *LobbyMenuController) onTurretToggled(selectedTurret *gamedata.AgentStats) {
	c.config.TurretDesign = selectedTurret.Kind.String()
	for _, b := range c.turretButtons {
		toggle := (b.drone != selectedTurret && b.widget.IsToggled())
		if toggle {
			b.widget.Toggle()
		}
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
		b.widget.SetDisabled(!b.available || b.drone.PointCost > pointsLeft)
	}
	if c.difficultyLabel != nil {
		c.updateDifficultyScore(c.calcDifficultyScore())
	}
}

func (c *LobbyMenuController) back() {
	c.saveConfig()
	c.scene.Context().ChangeScene(NewPlayMenuController(c.state))
}
