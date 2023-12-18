package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/serverapi"
)

type humanPlayer struct {
	world     *worldState
	state     *playerState
	scene     *ge.Scene
	input     *gameinput.Handler
	choiceGen *choiceGenerator

	tooltipManager *tooltipManager

	recipeTab            *recipeTabNode
	choiceWindow         *choiceWindowNode
	rpanel               *rpanelNode
	cursor               *gameui.CursorNode
	radar                *radarNode
	screenButtons        *screenButtonsNode
	colonySelector       *ge.Sprite
	flyingColonySelector *ge.Sprite
	colonyDestination    *ge.Line
	screenSeparator      *ge.Line

	droneSelectorsUsed int
	droneSelectors     []*ge.Sprite

	choiceCardColony     *colonyCoreNode
	choiceCardIndex      int
	choiceCardHighligh   *ge.Sprite
	choiceCenturionPoint gmath.Vec

	creepsState *creepsPlayerState

	spectator          bool
	permanentSeparator bool
	canPing            bool
	pingDelay          float64

	EventRecipesToggled     gsignal.Event[bool]
	EventPauseRequest       gsignal.Event[gsignal.Void]
	EventExitPressed        gsignal.Event[gsignal.Void]
	EventFastForwardPressed gsignal.Event[gsignal.Void]
	EventPing               gsignal.Event[gmath.Vec]
}

type humanPlayerConfig struct {
	world       *worldState
	state       *playerState
	input       *gameinput.Handler
	cursor      *gameui.CursorNode
	choiceGen   *choiceGenerator
	creepsState *creepsPlayerState
	spectator   bool
}

func newHumanPlayer(config humanPlayerConfig) *humanPlayer {
	canPing := config.world.config.GameMode != gamedata.ModeReverse &&
		config.world.config.PlayersMode == serverapi.PmodeTwoPlayers &&
		config.world.config.ExecMode == gamedata.ExecuteNormal
	p := &humanPlayer{
		world:           config.world,
		state:           config.state,
		scene:           config.world.rootScene,
		choiceGen:       config.choiceGen,
		input:           config.input,
		cursor:          config.cursor,
		creepsState:     config.creepsState,
		canPing:         canPing,
		spectator:       config.spectator,
		choiceCardIndex: -1,
	}
	return p
}

func (p *humanPlayer) addCrawlerFactoryToCreepsRadar(creep *creepNode) {
	p.radar.AddFactory(creep)
	creep.EventDestroyed.Connect(p, func(factory *creepNode) {
		p.radar.RemoveFactory(factory)
	})
}

func (p *humanPlayer) addCenturionToCreepsRadar(creep *creepNode) {
	p.radar.AddCenturion(creep)
	creep.EventDestroyed.Connect(p, func(centurion *creepNode) {
		p.radar.RemoveCenturion(centurion)
	})
}

func (p *humanPlayer) addColonyToCreepsRadar(colony *colonyCoreNode) {
	p.radar.AddColony(colony)
	colony.EventTurretAccepted.Connect(p, func(turret *colonyAgentNode) {
		p.radar.AddTurret(turret)
		turret.EventDestroyed.Connect(p, func(turret *colonyAgentNode) {
			p.radar.RemoveTurret(turret)
		})
	})
	colony.EventDestroyed.Connect(p, func(colony *colonyCoreNode) {
		p.radar.RemoveColony(colony)
	})
}

func (p *humanPlayer) CanPing() bool {
	return p.canPing
}

func (p *humanPlayer) CreateChoiceWindow(disableSpecial bool) {
	p.choiceWindow = newChoiceWindowNode(p.state.camera.Camera, p.world, p.input, p.cursor, p.creepsState != nil)
	p.world.nodeRunner.AddObject(p.choiceWindow)

	p.choiceGen.EventChoiceReady.Connect(p, p.choiceWindow.RevealChoices)
	p.choiceGen.EventChoiceSelected.Connect(p, func(choice selectedChoice) {
		if choice.Index == -1 {
			return
		}
		p.choiceWindow.StartCharging(choice.Cooldown, choice.Index)
		if p.rpanel != nil {
			p.rpanel.UpdateMetrics()
		}
	})

	if disableSpecial {
		p.choiceWindow.SetSpecialChoiceEnabled(false)
	}

	if p.choiceGen.IsReady() {
		p.choiceWindow.RevealChoices(p.choiceGen.GetChoices())
	}
}

func (p *humanPlayer) EnableSpecialChoices() {
	p.choiceWindow.SetSpecialChoiceEnabled(true)
}

func (p *humanPlayer) ForceSpecialChoice(kind specialChoiceKind) {
	p.choiceGen.ForceSpecialChoice(kind)
}

func (p *humanPlayer) Init() {
	if p.creepsState == nil && !p.spectator {
		p.colonyDestination = ge.NewLine(ge.Pos{}, ge.Pos{})
		p.colonyDestination.Visible = false
		p.colonyDestination.SetColorScaleRGBA(0x6e, 0x8e, 0xbd, 160)
		if p.world.coreDesign == gamedata.TankCoreStats {
			p.state.camera.Private.AddGraphicsBelow(p.colonyDestination)
		} else {
			p.state.camera.Private.AddGraphics(p.colonyDestination)
		}
	}

	if !p.spectator {
		choiceCardHighligh := p.scene.NewSprite(assets.ImageFloppyHighlight)
		choiceCardHighligh.Visible = false
		p.choiceCardHighligh = choiceCardHighligh
		p.state.camera.UI.AddGraphics(choiceCardHighligh)
	}

	if p.world.hintsMode > 0 {
		ttm := newTooltipManager(p, p.world.hintsMode > 1)
		p.cursor.EventHover.Connect(p, func(hoverPos gmath.Vec) {
			ttm.OnHover(hoverPos)
		})
		p.cursor.EventStopHover.Connect(p, func(gsignal.Void) {
			ttm.OnStopHover()
		})
		p.scene.AddObject(ttm)
		p.tooltipManager = ttm
		ttm.EventHighlightDrones.Connect(nil, func(drone *gamedata.AgentStats) {
			p.highlightDrones(drone)
		})
		ttm.EventTooltipClosed.Connect(nil, func(gsignal.Void) {
			if p.droneSelectorsUsed != 0 {
				p.hideDroneSelectors()
			}
		})
	}

	if !p.spectator && p.world.config.InterfaceMode >= 2 {
		if p.creepsState != nil {
			p.rpanel = newCreepsRpanelNode(p.state.camera.Camera, p.creepsState)
		} else {
			p.rpanel = newRpanelNode(p.state.camera.Camera)
		}
		p.scene.AddObject(p.rpanel)
	}

	buttonsPos := gmath.Vec{X: 137, Y: 470}
	if !p.spectator && p.world.config.EnemyBoss && p.world.config.InterfaceMode >= 1 {
		p.radar = newRadarNode(p.world, p, p.creepsState != nil)
		p.radar.Init(p.world.rootScene)
		if p.creepsState != nil {
			for _, c := range p.world.allColonies {
				p.addColonyToCreepsRadar(c)
			}
			p.world.EventColonyCreated.Connect(p, p.addColonyToCreepsRadar)

			for _, c := range p.world.creeps {
				if c.stats == gamedata.CenturionCreepStats {
					p.addCenturionToCreepsRadar(c)
				}
				if c.stats == gamedata.CrawlerBaseCreepStats {
					p.addCrawlerFactoryToCreepsRadar(c)
				}
			}
			p.world.EventCenturionCreated.Connect(p, p.addCenturionToCreepsRadar)
			p.world.EventCrawlerFactoryCreated.Connect(p, p.addCrawlerFactoryToCreepsRadar)
		}
		if p.creepsState != nil {
			if p.world.mapShape == gamedata.WorldHorizontal {
				buttonsPos.X += 28 * 2
			}
		}
	} else {
		buttonsPos = gmath.Vec{X: 8, Y: 470}
	}
	buttonsPos.Y += p.scene.Context().ScreenHeight - 540

	if len(p.world.cameras) == 1 && p.world.screenButtonsEnabled {
		p.screenButtons = newScreenButtonsNode(p.state.camera.Camera, buttonsPos, p.creepsState != nil)
		p.screenButtons.scaled = p.world.deviceInfo.IsMobile()
		p.screenButtons.Init(p.world.rootScene)
		p.screenButtons.EventToggleButtonPressed.Connect(p, p.onToggleButtonClicked)
		p.screenButtons.EventExitButtonPressed.Connect(p, p.onExitButtonClicked)
		p.screenButtons.EventFastForwardButtonPressed.Connect(p, p.onFastForwardButtonClicked)
	}

	if p.spectator {
		p.state.selectedColony = p.world.allColonies[0]
		p.state.camera.CenterOn(p.state.selectedColony.pos)
	} else if p.creepsState != nil {
		p.state.camera.CenterOn(p.world.boss.pos)
	} else {
		p.colonySelector = p.scene.NewSprite(p.world.coreDesign.SelectorImageID())
		if p.world.coreDesign == gamedata.ArkCoreStats {
			p.state.camera.Private.AddSprite(p.colonySelector)
		} else {
			p.state.camera.Private.AddSpriteBelow(p.colonySelector)
		}
		p.flyingColonySelector = p.scene.NewSprite(p.world.coreDesign.SelectorImageID())
		p.state.camera.Private.AddSpriteSlightlyAbove(p.flyingColonySelector)

		p.selectNextColony(true)
		p.state.camera.CenterOn(p.state.selectedColony.pos)
	}

	if !p.spectator && p.world.config.GameMode != gamedata.ModeTutorial {
		p.CreateChoiceWindow(false)
	}

	if p.creepsState == nil {
		p.recipeTab = newRecipeTabNode(p.world)
		p.recipeTab.Visible = false
		p.state.camera.UI.AddGraphics(p.recipeTab)
		p.scene.AddObject(p.recipeTab)
	}

	if len(p.world.cameras) == 2 && p.state.id == 0 {
		begin := ge.Pos{Offset: gmath.Vec{X: p.scene.Context().ScreenWidth / 2}}
		end := ge.Pos{Offset: gmath.Vec{X: p.scene.Context().ScreenWidth / 2, Y: p.scene.Context().ScreenHeight}}
		p.permanentSeparator = p.scene.Context().ScreenHeight > 540
		p.screenSeparator = ge.NewLine(begin, end)
		p.screenSeparator.SetColorScaleRGBA(0xa1, 0x9a, 0x9e, 255)
		p.screenSeparator.Visible = p.permanentSeparator || p.rpanel == nil || !p.state.camera.UI.Visible
		p.scene.AddGraphicsAbove(p.screenSeparator, 1)
	}

	p.input.EventGamepadDisconnected.Connect(p, func(gsignal.Void) {
		if p.world.nodeRunner.IsPaused() {
			return
		}
		p.EventPauseRequest.Emit(gsignal.Void{})
	})

	// On mobiles, this tab is always opened.
	// Except for the tutorial, where we reveal it later.
	if p.world.config.GameMode != gamedata.ModeTutorial && p.input.InputMethod == gameinput.InputMethodTouch {
		if p.creepsState == nil {
			p.SetRecipeTabVisibility(true)
		}
	}
}

func (p *humanPlayer) SetRecipeTabVisibility(visible bool) {
	p.recipeTab.Visible = visible
	p.EventRecipesToggled.Emit(visible)
}

func (p *humanPlayer) IsDisposed() bool { return false }

func (p *humanPlayer) GetState() *playerState { return p.state }

func (p *humanPlayer) AfterUpdateStep() {
	if !p.spectator && p.state.selectedColony != nil {
		flying := p.state.selectedColony.IsFlying()
		p.colonySelector.Visible = !flying
		p.flyingColonySelector.Visible = flying
		p.updateWaypointLine()
	}
}

func (p *humanPlayer) BeforeUpdateStep(delta float64) {
	// This method is called even if the game is paused.
	//
	// When input is being handled here, the player doesn't directly execute the actions
	// right away; instead, they're recorded until the non-paused Update() frame comes.
	//
	// It allows the player to queue actions while on pause.
	// It also allows the game to execute the actions inside the Update() loop as
	// opposed to an out-of-place execution right during the pause.

	p.state.camera.Update(delta)
	p.state.messageManager.Update(delta)

	if !p.spectator {
		if p.world.nodeRunner.IsPaused() {
			p.choiceCardHighligh.Visible = p.choiceCardColony == p.state.selectedColony &&
				p.choiceCardIndex != -1
		} else {
			p.choiceCardHighligh.Visible = false
		}

		if p.canPing {
			p.pingDelay = gmath.ClampMin(p.pingDelay-delta, 0)
		}

		if p.radar != nil {
			p.radar.UpdateCamera()
		}
	}

	p.handleInput()
}

func (p *humanPlayer) Update(computedDelta, delta float64) {
	if p.choiceWindow != nil {
		p.choiceWindow.Enabled = p.state.selectedColony != nil &&
			p.state.selectedColony.mode == colonyModeNormal
	}

	if p.radar != nil {
		p.radar.Update(delta)
	}

	if p.choiceCardIndex != -1 {
		if !p.choiceGen.TryExecute(p.choiceCardColony, p.choiceCardIndex, gmath.Vec{}) {
			p.scene.Audio().PlaySound(assets.AudioError)
		}
		p.choiceCardIndex = -1
		p.choiceCardColony = nil
	}

	for _, colony := range p.state.colonies {
		if colony.plannedRelocationPoint.IsZero() {
			continue
		}
		pos := colony.plannedRelocationPoint
		colony.plannedRelocationPoint = gmath.Vec{}
		if !p.choiceGen.TryExecute(colony, -1, pos) {
			p.scene.Audio().PlaySound(assets.AudioError)
		}
	}

	if !p.choiceCenturionPoint.IsZero() {
		if !p.choiceGen.TryExecute(nil, -1, p.choiceCenturionPoint) {
			p.scene.Audio().PlaySound(assets.AudioError)
		}
		p.choiceCenturionPoint = gmath.Vec{}
	}
}

func (p *humanPlayer) updateWaypointLine() {
	colony := p.state.selectedColony
	if p.world.nodeRunner.IsPaused() {
		dstPos := &colony.relocationPoint
		if dstPos.IsZero() {
			dstPos = &colony.plannedRelocationPoint
		}
		p.colonyDestination.EndPos.Base = dstPos
		p.colonyDestination.Visible = !dstPos.IsZero() &&
			colony.mode != colonyModeTeleporting
	} else {
		p.colonyDestination.EndPos.Base = &colony.relocationPoint
		p.colonyDestination.Visible = !colony.relocationPoint.IsZero() &&
			colony.mode != colonyModeTeleporting
	}
}

func (p *humanPlayer) GetCursor() *gameui.CursorNode {
	return p.cursor
}

func (p *humanPlayer) GetInput() *gameinput.Handler {
	return p.input
}

func (p *humanPlayer) handleInput() {
	selectedColony := p.state.selectedColony

	p.input.Update()
	p.state.camera.HandleInput()

	if p.world.nodeRunner.exitPrompt {
		if p.input.IsClickDevice() {
			// For these cases, the exit prompt is handled in the staging controller.
			return
		}
	}

	// Pinging is OK even during the pause.
	if p.canPing && p.pingDelay == 0 {
		if clickPos, ok := p.cursor.ClickPos(controls.ActionPing); ok {
			p.pingDelay = 5.0
			globalClickPos := p.state.camera.AbsClickPos(clickPos)
			p.EventPing.Emit(globalClickPos)
			return
		}
	}

	// Recipe tab toggle is OK during the pause.
	if p.recipeTab != nil {
		if p.input.ActionIsJustPressed(controls.ActionShowRecipes) {
			p.SetRecipeTabVisibility(!p.recipeTab.Visible)
		}
	}

	// Colony view toggle is OK during the pause.
	if p.input.ActionIsJustPressed(controls.ActionToggleColony) {
		p.onToggleButtonClicked(gsignal.Void{})
		return
	}

	// Interface on/off toggle is OK during the pause.
	if p.input.ActionIsJustPressed(controls.ActionToggleInterface) {
		p.state.camera.UI.Visible = !p.state.camera.UI.Visible
		if p.screenSeparator != nil {
			p.screenSeparator.Visible = p.permanentSeparator || p.rpanel == nil || !p.state.camera.UI.Visible
		}
	}

	// Card choice is not executed right away. We just remember the index that was selected.
	if !p.spectator && (selectedColony != nil || p.creepsState != nil) && p.choiceWindow != nil && p.state.camera.UI.Visible {
		cardIndex, cardPos := p.choiceWindow.HandleInput()
		if cardIndex != -1 {
			if p.choiceCardIndex != cardIndex || selectedColony != p.choiceCardColony {
				p.choiceCardIndex = cardIndex
				p.choiceCardColony = selectedColony
				p.choiceCardHighligh.Pos = cardPos
			} else {
				p.choiceCardIndex = -1
				p.choiceCardColony = nil
			}
			return
		}
	}

	if p.world.deviceInfo.IsMobile() && p.tooltipManager != nil {
		if p.input.ActionIsJustPressed(controls.ActionPanDrag) {
			p.tooltipManager.OnStopHover()
			return
		}
		if info, ok := p.input.JustPressedActionInfo(controls.ActionShowTooltip); ok {
			p.tooltipManager.OnHover(info.Pos)
			return
		}
	}

	clickPos, hasClick := p.cursor.ClickPos(controls.ActionClick)

	if p.world.deviceInfo.IsMobile() && p.recipeTab != nil && p.recipeTab.Visible {
		if p.recipeTab.ContainsPos(clickPos) {
			return
		}
	}

	if hasClick && p.state.messageManager.HandleInput(clickPos) {
		return
	}

	handledClick := false
	// Selecting a colony by clicking on it is OK during the pause.
	if !p.spectator && len(p.state.colonies) > 1 {
		if hasClick {
			globalClickPos := p.state.camera.AbsClickPos(clickPos)
			selectDist := 40.0
			if p.world.deviceInfo.IsMobile() {
				selectDist = 80.0
			}
			var closestColony *colonyCoreNode
			closestDist := math.MaxFloat64
			for _, colony := range p.state.colonies {
				if colony == p.state.selectedColony {
					continue
				}
				dist := colony.pos.DistanceTo(globalClickPos)
				if dist > selectDist {
					continue
				}
				if dist < closestDist {
					closestColony = colony
					closestDist = dist
				}
			}
			if closestColony != nil {
				p.selectColony(closestColony)
				handledClick = true
			}
		}
	}
	if handledClick {
		return
	}

	// Centering the camera on some spot is OK during the pause.
	if p.radar != nil {
		requestedCameraPos, ok := p.radar.ResolveClick(clickPos)
		if ok {
			p.state.camera.CenterOn(requestedCameraPos)
			return
		}
	}

	// Screen buttons (toggle view, exit, fast forward) are OK during the pause.
	if p.screenButtons != nil && p.state.camera.UI.Visible {
		if p.screenButtons.HandleInput(clickPos) {
			return
		}
	}

	if !p.spectator && selectedColony != nil && selectedColony.relocationPoint.IsZero() && selectedColony.mode == colonyModeNormal {
		if pos, ok := p.cursor.ClickPos(controls.ActionMoveChoice); ok {
			globalClickPos := p.state.camera.AbsClickPos(pos)
			if globalClickPos.DistanceTo(selectedColony.GetRallyPoint()) >= 40 {
				selectedColony.plannedRelocationPoint = globalClickPos
			} else {
				selectedColony.plannedRelocationPoint = gmath.Vec{}
			}
			return
		}
	}

	if p.creepsState != nil && len(p.world.centurions) != 0 {
		pos, ok := p.cursor.ClickPos(controls.ActionMoveChoice)
		if ok && p.world.AllCenturionsReady() {
			globalClickPos := p.state.camera.AbsClickPos(pos)
			p.choiceCenturionPoint = globalClickPos
		}
	}
}

func (p *humanPlayer) selectNextColony(center bool) {
	colony := p.findNextColony(p.state.colonies)
	p.selectColony(colony)
	if center && p.state.selectedColony != nil {
		p.state.camera.ToggleCamera(p.state.selectedColony.pos)
	}
}

func (p *humanPlayer) findNextColony(colonies []*colonyCoreNode) *colonyCoreNode {
	if len(colonies) == 0 {
		return nil
	}
	if len(colonies) == 1 {
		return colonies[0]
	}
	index := xslices.Index(colonies, p.state.selectedColony)
	if index == len(colonies)-1 {
		index = 0
	} else {
		index++
	}
	return colonies[index]
}

func (p *humanPlayer) selectColony(colony *colonyCoreNode) {
	if p.state.selectedColony == colony {
		return
	}
	if p.state.selectedColony != nil {
		p.scene.Audio().PlaySound(assets.AudioBaseSelect)
		p.state.selectedColony.EventDestroyed.Disconnect(p)
		p.state.selectedColony.EventTeleported.Disconnect(p)
		if p.rpanel != nil {
			p.state.selectedColony.EventPrioritiesChanged.Disconnect(p)
		}
	}
	p.state.selectedColony = colony

	if p.radar != nil {
		p.radar.SetBase(p.state.selectedColony)
	}
	if p.rpanel != nil {
		p.rpanel.SetBase(p.state.selectedColony)
		p.rpanel.UpdateMetrics()
	}
	if p.state.selectedColony == nil {
		p.colonySelector.Visible = false
		p.flyingColonySelector.Visible = false
		p.colonyDestination.Visible = false
		return
	}
	p.state.selectedColony.EventDestroyed.Connect(p, func(_ *colonyCoreNode) {
		p.selectNextColony(false)
	})
	p.state.selectedColony.EventTeleported.Connect(p, func(colony *colonyCoreNode) {
		p.state.camera.ToggleCamera(colony.pos)
	})
	if p.rpanel != nil {
		p.state.selectedColony.EventPrioritiesChanged.Connect(p, func(_ *colonyCoreNode) {
			p.rpanel.UpdateMetrics()
		})
	}
	p.colonySelector.Pos.Base = &p.state.selectedColony.pos
	p.flyingColonySelector.Pos.Base = &p.state.selectedColony.pos
	if p.state.selectedColony.stats == gamedata.HiveCoreStats {
		p.colonyDestination.BeginPos.Base = &p.state.selectedColony.rallyPoint
	} else {
		p.colonyDestination.BeginPos.Base = &p.state.selectedColony.pos
	}
	p.updateWaypointLine()
}

func (p *humanPlayer) highlightDrones(droneStats *gamedata.AgentStats) {
	if p.state.selectedColony == nil {
		return
	}

	p.state.selectedColony.agents.Each(func(drone *colonyAgentNode) {
		if drone.stats != droneStats {
			return
		}
		s := p.makeDroneSelector()
		s.Pos.Base = &drone.pos
		s.Visible = true
	})
}

func (p *humanPlayer) hideDroneSelectors() {
	for _, s := range p.droneSelectors {
		s.Visible = false
	}
	p.droneSelectorsUsed = 0
}

func (p *humanPlayer) makeDroneSelector() *ge.Sprite {
	numSelectors := p.droneSelectorsUsed
	p.droneSelectorsUsed++

	if numSelectors < len(p.droneSelectors) {
		return p.droneSelectors[numSelectors]
	}

	s := p.scene.NewSprite(assets.ImageDroneSelector)
	p.droneSelectors = append(p.droneSelectors, s)
	p.state.camera.Private.AddGraphicsSlightlyAbove(s)
	return s
}

func (p *humanPlayer) onFastForwardButtonClicked(gsignal.Void) {
	p.EventFastForwardPressed.Emit(gsignal.Void{})
}

func (p *humanPlayer) onExitButtonClicked(gsignal.Void) {
	p.EventExitPressed.Emit(gsignal.Void{})
}

func (p *humanPlayer) onToggleButtonClicked(gsignal.Void) {
	if p.tooltipManager != nil {
		p.tooltipManager.removeTooltip()
	}

	if p.spectator {
		colony := p.findNextColony(p.world.allColonies)
		if colony != nil {
			p.state.selectedColony = colony
			p.state.camera.ToggleCamera(p.state.selectedColony.pos)
		}
		return
	}

	if p.creepsState == nil {
		p.selectNextColony(true)
		return
	}

	if p.world.boss != nil {
		p.state.camera.ToggleCamera(p.world.boss.pos)
	}
}
