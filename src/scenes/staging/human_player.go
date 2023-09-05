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

const colonyVisionRadius float64 = 500.0

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

	creepsState *creepsPlayerState

	inputHandled bool

	canPing   bool
	pingDelay float64

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
}

func newHumanPlayer(config humanPlayerConfig) *humanPlayer {
	canPing := config.world.config.GameMode != gamedata.ModeReverse &&
		config.world.config.PlayersMode == serverapi.PmodeTwoPlayers &&
		config.world.config.ExecMode == gamedata.ExecuteNormal
	p := &humanPlayer{
		world:       config.world,
		state:       config.state,
		scene:       config.world.rootScene,
		choiceGen:   config.choiceGen,
		input:       config.input,
		cursor:      config.cursor,
		creepsState: config.creepsState,
		canPing:     canPing,
	}
	return p
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
	if p.creepsState == nil {
		p.colonyDestination = ge.NewLine(ge.Pos{}, ge.Pos{})
		p.colonyDestination.Visible = false
		p.colonyDestination.SetColorScaleRGBA(0x6e, 0x8e, 0xbd, 160)
		if p.world.coreDesign == gamedata.TankCoreStats {
			p.state.camera.Private.AddGraphicsBelow(p.colonyDestination)
		} else {
			p.state.camera.Private.AddGraphics(p.colonyDestination)
		}
	}

	p.state.Init(p.world)

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

	if p.world.config.InterfaceMode >= 2 {
		if p.creepsState != nil {
			p.rpanel = newCreepsRpanelNode(p.state.camera.Camera, p.creepsState)
		} else {
			p.rpanel = newRpanelNode(p.state.camera.Camera)
		}
		p.scene.AddObject(p.rpanel)
	}

	buttonsPos := gmath.Vec{X: 137, Y: 470}
	if p.world.config.EnemyBoss && p.world.config.InterfaceMode >= 1 {
		p.radar = newRadarNode(p.world, p, p.creepsState != nil)
		p.radar.Init(p.world.rootScene)
		if p.creepsState != nil {
			for _, c := range p.world.allColonies {
				p.addColonyToCreepsRadar(c)
			}
			p.world.EventColonyCreated.Connect(p, p.addColonyToCreepsRadar)
		}
	} else {
		buttonsPos = gmath.Vec{X: 8, Y: 470}
	}

	if len(p.world.cameras) == 1 && p.world.screenButtonsEnabled {
		p.screenButtons = newScreenButtonsNode(p.state.camera.Camera, buttonsPos, p.creepsState != nil)
		p.screenButtons.Init(p.world.rootScene)
		p.screenButtons.EventToggleButtonPressed.Connect(p, p.onToggleButtonClicked)
		p.screenButtons.EventExitButtonPressed.Connect(p, p.onExitButtonClicked)
		p.screenButtons.EventFastForwardButtonPressed.Connect(p, p.onFastForwardButtonClicked)
	}

	if p.creepsState != nil {
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

	if p.world.config.GameMode != gamedata.ModeTutorial {
		p.CreateChoiceWindow(false)
	}

	if p.creepsState == nil {
		p.recipeTab = newRecipeTabNode(p.world)
		p.recipeTab.Visible = false
		p.state.camera.UI.AddGraphics(p.recipeTab)
		p.scene.AddObject(p.recipeTab)
	}

	if len(p.world.cameras) == 2 && p.state.id == 0 {
		begin := ge.Pos{Offset: gmath.Vec{X: (1920 / 4)}}
		end := ge.Pos{Offset: gmath.Vec{X: (1920 / 4), Y: 1080}}
		p.screenSeparator = ge.NewLine(begin, end)
		p.screenSeparator.SetColorScaleRGBA(0xa1, 0x9a, 0x9e, 255)
		p.screenSeparator.Visible = p.rpanel == nil || !p.state.camera.UI.Visible
		p.scene.AddGraphicsAbove(p.screenSeparator, 1)
	}

	p.input.EventGamepadDisconnected.Connect(p, func(gsignal.Void) {
		if p.world.nodeRunner.IsPaused() {
			return
		}
		p.EventPauseRequest.Emit(gsignal.Void{})
	})
}

func (p *humanPlayer) IsDisposed() bool { return false }

func (p *humanPlayer) GetState() *playerState { return p.state }

func (p *humanPlayer) BeforeUpdateStep() {
	p.inputHandled = false
}

func (p *humanPlayer) Update(computedDelta, delta float64) {
	if p.choiceWindow != nil {
		p.choiceWindow.Enabled = p.state.selectedColony != nil &&
			p.state.selectedColony.mode == colonyModeNormal
	}

	if p.canPing {
		p.pingDelay = gmath.ClampMin(p.pingDelay-delta, 0)
	}

	if p.radar != nil {
		p.radar.Update(delta)
	}

	if p.inputHandled {
		return
	}
	p.inputHandled = true
	p.handleInput()

	if p.state.selectedColony != nil {
		flying := p.state.selectedColony.IsFlying()
		p.colonySelector.Visible = !flying
		p.flyingColonySelector.Visible = flying
		p.colonyDestination.Visible = !p.state.selectedColony.relocationPoint.IsZero() &&
			p.state.selectedColony.mode != colonyModeTeleporting
	}
}

func (p *humanPlayer) handleInput() {
	selectedColony := p.state.selectedColony

	p.input.Update()

	if p.canPing && p.pingDelay == 0 {
		if clickPos, ok := p.cursor.ClickPos(controls.ActionPing); ok {
			p.pingDelay = 5.0
			globalClickPos := p.state.camera.AbsClickPos(clickPos)
			p.EventPing.Emit(globalClickPos)
			return
		}
	}

	if p.recipeTab != nil {
		if p.input.ActionIsJustPressed(controls.ActionShowRecipes) {
			p.recipeTab.Visible = !p.recipeTab.Visible
			p.EventRecipesToggled.Emit(p.recipeTab.Visible)
		}
	}

	if p.input.ActionIsJustPressed(controls.ActionToggleColony) {
		p.onToggleButtonClicked(gsignal.Void{})
		return
	}

	if p.input.ActionIsJustPressed(controls.ActionToggleInterface) {
		p.state.camera.UI.Visible = !p.state.camera.UI.Visible
		if p.screenSeparator != nil {
			p.screenSeparator.Visible = p.rpanel == nil || !p.state.camera.UI.Visible
		}
	}

	if (selectedColony != nil || p.creepsState != nil) && p.choiceWindow != nil && p.state.camera.UI.Visible {
		if cardIndex := p.choiceWindow.HandleInput(); cardIndex != -1 {
			if !p.choiceGen.TryExecute(cardIndex, gmath.Vec{}) {
				p.scene.Audio().PlaySound(assets.AudioError)
			}
			return
		}
	}

	handledClick := false
	clickPos, hasClick := p.cursor.ClickPos(controls.ActionClick)
	if len(p.state.colonies) > 1 {
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

	if p.radar != nil {
		requestedCameraPos, ok := p.radar.ResolveClick(clickPos)
		if ok {
			p.state.camera.CenterOn(requestedCameraPos)
			return
		}
	}

	if p.screenButtons != nil && p.state.camera.UI.Visible {
		if p.screenButtons.HandleInput(clickPos) {
			return
		}
	}

	if selectedColony != nil && p.world.movementEnabled {
		if pos, ok := p.cursor.ClickPos(controls.ActionMoveChoice); ok {
			globalClickPos := p.state.camera.AbsClickPos(pos)
			if globalClickPos.DistanceTo(selectedColony.pos) > 28 {
				if !p.choiceGen.TryExecute(-1, globalClickPos) {
					p.scene.Audio().PlaySound(assets.AudioError)
				}
				return
			}
		}
	}
}

func (p *humanPlayer) selectNextColony(center bool) {
	colony := p.findNextColony()
	p.selectColony(colony)
	if center && p.state.selectedColony != nil {
		p.state.camera.ToggleCamera(p.state.selectedColony.pos)
	}
}

func (p *humanPlayer) findNextColony() *colonyCoreNode {
	if len(p.state.colonies) == 0 {
		return nil
	}
	if len(p.state.colonies) == 1 {
		return p.state.colonies[0]
	}
	index := xslices.Index(p.state.colonies, p.state.selectedColony)
	if index == len(p.state.colonies)-1 {
		index = 0
	} else {
		index++
	}
	return p.state.colonies[index]
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
	p.colonyDestination.BeginPos.Base = &p.state.selectedColony.pos
	p.colonyDestination.EndPos.Base = &p.state.selectedColony.relocationPoint
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

	if p.creepsState == nil {
		p.selectNextColony(true)
		return
	}

	if p.world.boss != nil {
		p.state.camera.ToggleCamera(p.world.boss.pos)
	}
}
