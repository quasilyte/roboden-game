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
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

type LobbyMenuController struct {
	state *session.State

	config gamedata.LevelConfig
	mode   gamedata.Mode

	droneButtons         []droneButton
	turretButtons        []droneButton
	coreButtons          []coreButton
	pointsAllocatedLabel *widget.Text
	difficultyLabel      *widget.Text

	seedInput *widget.TextInput

	colonyTab *widget.TabBookTab
	worldTab  *widget.TabBookTab

	helpPanel         *widget.Container
	helpLabel         *widget.Text
	helpIcon1         *widget.Graphic
	helpIconSeparator *widget.Text
	helpIcon2         *widget.Graphic

	recipeIcons map[gamedata.RecipeSubject]*ebiten.Image

	scene *ge.Scene
}

type coreButton struct {
	widget *eui.ItemButton
	core   *gamedata.ColonyCoreStats
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

	c.config = *c.getConfigForMode()

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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
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
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	rootGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{false, false}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))

	root.AddChild(rootGrid)

	leftRowsContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{
			VerticalPosition: widget.GridLayoutPositionStart,
		})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	leftRows := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
				StretchVertical:   true,
			}),
			widget.WidgetOpts.MinSize(572, (1080/2)-47),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4),
		)),
	)
	leftRowsContainer.AddChild(leftRows)
	rootGrid.AddChild(leftRowsContainer)

	rightRows := eui.NewRowLayoutContainer(4, []bool{false, true, false})
	rootGrid.AddChild(rightRows)

	tabs := c.createTabs(uiResources)
	leftRows.AddChild(tabs)

	rightRows.AddChild(c.createSeedPanel(uiResources))
	rightRows.AddChild(c.createHelpPanel(uiResources))
	rightRows.AddChild(c.createButtonsPanel(uiResources))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)

	c.updateDifficultyScore(c.calcDifficultyScore())
}

func (c *LobbyMenuController) getConfigForMode() *gamedata.LevelConfig {
	switch c.mode {
	case gamedata.ModeArena:
		return c.state.ArenaLevelConfig
	case gamedata.ModeInfArena:
		return c.state.InfArenaLevelConfig
	case gamedata.ModeClassic:
		return c.state.ClassicLevelConfig
	case gamedata.ModeReverse:
		return c.state.ReverseLevelConfig
	case gamedata.ModeTutorial:
		return c.state.TutorialLevelConfig
	default:
		panic("unexpected game mode")
	}
}

func (c *LobbyMenuController) saveConfig() {
	*c.getConfigForMode() = c.config.Clone()
}

func (c *LobbyMenuController) createButtonsPanel(uiResources *eui.Resources) *widget.Container {
	panel := eui.NewPanel(uiResources, 0, 0)

	d := c.scene.Dict()

	tinyFont := assets.BitmapFont1

	c.difficultyLabel = eui.NewCenteredLabel("Difficulty: 1000%", tinyFont)
	panel.AddChild(c.difficultyLabel)

	panel.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.lobby.go"), func() {
		c.saveConfig()

		if c.config.PlayersMode == serverapi.PmodeSinglePlayer && c.mode == gamedata.ModeReverse {
			c.config.CoreDesign = gamedata.PickColonyDesign(c.state.Persistent.PlayerStats.CoresUnlocked, c.scene.Rand())
			c.config.TurretDesign = gamedata.PickTurretDesign(c.scene.Rand())
			c.config.Tier2Recipes = gamedata.CreateDroneBuild(c.scene.Rand())
		}

		c.config.GameMode = c.mode
		c.config.DronePointsAllocated = c.calcAllocatedPoints()
		if c.seedInput.GetText() != "" {
			seed, err := strconv.ParseInt(c.seedInput.GetText(), 10, 64)
			if err != nil {
				panic(err)
			}
			c.config.Seed = seed
		} else {
			c.config.Seed = c.randomSeed()
		}

		c.config.Finalize()
		c.scene.Context().ChangeScene(staging.NewController(c.state, c.config.Clone(), NewLobbyMenuController(c.state, c.mode)))
	}))

	panel.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	return panel
}

func (c *LobbyMenuController) createTabs(uiResources *eui.Resources) *widget.TabBook {
	tabs := []*widget.TabBookTab{}

	colonyTab := c.createColonyTab(uiResources)
	tabs = append(tabs, colonyTab)
	worldTab := c.createWorldTab(uiResources)
	tabs = append(tabs, worldTab)
	tabs = append(tabs, c.createDifficultyTab(uiResources))
	tabs = append(tabs, c.createExtraTab(uiResources))

	if c.config.RawGameMode == "reverse" {
		c.maybeDisableColonyTab(c.config.PlayersMode != serverapi.PmodeTwoPlayers)
	}

	t := widget.NewTabBook(
		// widget.TabBookOpts.InitialTab(worldTab),
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

func (c *LobbyMenuController) maybeDisableColonyTab(disable bool) {
	if c.config.RawGameMode != "reverse" {
		return
	}
	c.colonyTab.Disabled = disable
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
		disabled := []int{}
		if c.config.RawGameMode == "reverse" {
			disabled = append(disabled, 1, 2, 4) // These combinations are not supported for this mode
		}
		if c.state.Device.IsMobile {
			disabled = append(disabled, 3) // Two players are not available on mobiles
		}
		key := "menu.lobby.players"
		if c.config.RawGameMode == "reverse" {
			key += ".reverse"
		}
		b := c.newOptionButtonWithDisabled(&c.config.PlayersMode, key, disabled, []string{
			d.Get("menu.lobby.player_mode.single_player"),
			d.Get("menu.lobby.player_mode.single_bot"),
			d.Get("menu.lobby.player_mode.player_and_bot"),
			d.Get("menu.lobby.player_mode.two_players"),
			d.Get("menu.lobby.player_mode.two_bots"),
		})
		tab.AddChild(b)
		b.PressedEvent.AddHandler(func(args interface{}) {
			// This handler is called before the config value is changed.
			c.maybeDisableColonyTab(c.config.PlayersMode == serverapi.PmodeTwoPlayers)
		})
	}

	if c.config.RawGameMode != "reverse" {
		disabled := []int{}
		if c.config.RawGameMode == "arena" || c.config.RawGameMode == "inf_arena" {
			disabled = []int{1}
		}
		b := c.newOptionButtonWithDisabled(&c.config.InterfaceMode, "menu.lobby.ui_mode", disabled, []string{
			d.Get("menu.lobby.ui_minimal"),
			d.Get("menu.lobby.ui_radar"),
			d.Get("menu.lobby.ui_full"),
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

	if c.config.RawGameMode != "reverse" {
		b := c.newBoolOptionButton(&c.config.FogOfWar, "menu.lobby.fog_of_war", []string{
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
			"5",
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
		b := c.newBoolOptionButton(&c.config.CreepFortress, "menu.lobby.creep_fortress", []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newBoolOptionButton(&c.config.IonMortars, "menu.lobby.ion_mortars", []string{
			d.Get("menu.option.off"),
			d.Get("menu.option.on"),
		})
		tab.AddChild(b)
	}

	{
		b := c.newOptionButton(&c.config.CreepDifficulty, "menu.lobby.creeps_difficulty", []string{
			"25%",
			"50%",
			"75%",
			"100%",
			"125%",
			"150%",
			"175%",
			"200%",
			"225%",
			"250%",
			"275%",
			"300%",
		})
		tab.AddChild(b)
	}

	if c.mode == gamedata.ModeReverse {
		b := c.newOptionButton(&c.config.DronesPower, "menu.lobby.drones_power", []string{
			"80%",
			"100%",
			"120%",
			"140%",
			"160%",
			"180%",
			"200%",
		})
		tab.AddChild(b)
	}

	if c.mode == gamedata.ModeReverse {
		tab.AddChild(c.newOptionButton(&c.config.TechProgressRate, "menu.lobby.tech_progress_rate", []string{
			"50%",
			"60%",
			"70%",
			"80%",
			"90%",
			"100%",
			"110%",
			"120%",
		}))
	}

	if c.mode == gamedata.ModeClassic || c.mode == gamedata.ModeReverse {
		tab.AddChild(c.newOptionButton(&c.config.BossDifficulty, "menu.lobby.boss_difficulty", []string{
			d.Get("menu.power.weak"),
			d.Get("menu.power.normal"),
			d.Get("menu.power.tough"),
			d.Get("menu.power.very_tough"),
		}))
	}

	if c.mode == gamedata.ModeClassic {
		tab.AddChild(c.newOptionButton(&c.config.CreepSpawnRate, "menu.lobby.creep_spawn_rate", []string{
			"75%",
			"100%",
			"125%",
			"150%",
		}))

		if c.state.Persistent.PlayerStats.TotalScore >= gamedata.ArenaModeCost {
			tab.AddChild(c.newBoolOptionButton(&c.config.SuperCreeps, "menu.lobby.super_creeps", []string{
				d.Get("menu.option.off"),
				d.Get("menu.option.on"),
			}))
		}
	}

	if c.mode == gamedata.ModeArena || c.mode == gamedata.ModeInfArena {
		b := c.newOptionButton(&c.config.ArenaProgression, "menu.lobby.arena_progression", []string{
			"75%",
			"100%",
			"125%",
			"150%",
			"175%",
			"200%",
			"225%",
			"250%",
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

func (c *LobbyMenuController) newOptionButtonWithDisabled(value *int, key string, disabled []int, valueNames []string) *widget.Button {
	return eui.NewSelectButton(eui.SelectButtonConfig{
		Scene:          c.scene,
		Resources:      c.state.Resources.UI,
		Input:          c.state.CombinedInput,
		Value:          value,
		DisabledValues: disabled,
		Label:          c.scene.Dict().Get(key),
		ValueNames:     valueNames,
		OnPressed: func() {
			c.updateDifficultyScore(c.calcDifficultyScore())
		},
		OnHover: func() {
			c.setHelpText(c.optionDescriptionText(key))
		},
	})
}

func (c *LobbyMenuController) newOptionButton(value *int, key string, valueNames []string) widget.PreferredSizeLocateableWidget {
	return c.newOptionButtonWithDisabled(value, key, nil, valueNames)
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

	{
		b := c.newOptionButton(&c.config.Environment, "menu.lobby.environment", []string{
			d.Get("menu.lobby.moon"),
			d.Get("menu.lobby.forest"),
		})
		tab.AddChild(b)
	}

	c.worldTab = tab

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

	tinyFont := assets.BitmapFont1

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

	c.colonyTab = tab

	return tab
}

func (c *LobbyMenuController) calcDifficultyScore() int {
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
	case score < 300:
		tag = d.Get("menu.option.impossible")
	default:
		tag = d.Get("menu.difficulty_score_despair")
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
	panel := eui.NewTextPanel(uiResources, 0, 0)
	c.helpPanel = panel

	tinyFont := assets.BitmapFont1
	normalFont := assets.BitmapFont2

	label := eui.NewLabel("", tinyFont)
	label.MaxWidth = 305
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

func (c *LobbyMenuController) randomSeed() int64 {
	return c.scene.Rand().PositiveInt64()
}

func (c *LobbyMenuController) createSeedPanel(uiResources *eui.Resources) *widget.Container {
	worldSettingsPanel := eui.NewPanel(uiResources, 340, 0)

	tinyFont := assets.BitmapFont1

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

		const maxSeedLen = 18
		textinput := eui.NewTextInput(uiResources, eui.TextInputConfig{SteamDeck: c.state.SteamInfo.SteamDeck},
			widget.TextInputOpts.WidgetOpts(
				widget.WidgetOpts.CursorEnterHandler(func(args *widget.WidgetCursorEnterEventArgs) {
					c.setHelpText(c.optionDescriptionText("menu.lobby.game_seed"))
				}),
			),
			widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
				if len(newInputText) > maxSeedLen {
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
		randSeed := strconv.FormatInt(c.randomSeed(), 10)
		if len(randSeed) >= maxSeedLen {
			randSeed = randSeed[:maxSeedLen]
		}
		textinput.SetText(randSeed)
		grid.AddChild(textinput)
		c.seedInput = textinput
		label := widget.NewLabel(
			widget.LabelOpts.TextOpts(
				widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			),
			widget.LabelOpts.Text(d.Get("menu.lobby.game_seed"), tinyFont, &widget.LabelColor{
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

	for i := range gamedata.CoreStatsList {
		core := gamedata.CoreStatsList[i]
		var b *eui.ItemButton
		available := xslices.Contains(c.state.Persistent.PlayerStats.CoresUnlocked, core.Name)
		var img *ebiten.Image
		if available {
			img = c.scene.LoadImage(core.Image).Data
		} else {
			img = c.scene.LoadImage(assets.ImageLock).Data
		}
		b = eui.NewBigItemButton(uiResources, img, func() {
			if c.config.CoreDesign != core.Name {
				b.Toggle()
				c.onCoreToggled(core)
			}
		})
		b.SetDisabled(!available)
		grid.AddChild(b.Widget)
		b.Widget.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			var s string
			if available {
				s = descriptions.CoreText(c.scene.Dict(), core)
			} else {
				s = descriptions.LockedCoreText(c.scene.Dict(), &c.state.Persistent.PlayerStats, core)
			}
			c.setHelpText(s)
		})
		if c.config.CoreDesign == core.Name {
			b.Toggle()
		}
		c.coreButtons = append(c.coreButtons, coreButton{
			widget: b,
			core:   core,
		})
	}

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
		b = eui.NewItemButton(uiResources, img, nil, "", 0, func() {
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

	smallFont := assets.BitmapFont1

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
		b = eui.NewItemButton(uiResources, frame, smallFont, costLabel, 26, func() {
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
				c.helpLabel.Label = descriptions.DroneText(c.scene.Dict(), drone, false, false)
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
		b := eui.NewItemButton(uiResources, nil, nil, "", 0, func() {})
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

func (c *LobbyMenuController) onCoreToggled(selectedCore *gamedata.ColonyCoreStats) {
	c.config.CoreDesign = selectedCore.Name
	for _, b := range c.coreButtons {
		toggle := (b.core != selectedCore && b.widget.IsToggled())
		if toggle {
			b.widget.Toggle()
		}
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
