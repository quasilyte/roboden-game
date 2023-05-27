package staging

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gedraw"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui"
)

const colonyVisionRadius float64 = 500.0

type humanPlayer struct {
	world     *worldState
	state     *playerState
	scene     *ge.Scene
	input     gameinput.Handler
	choiceGen *choiceGenerator

	choiceWindow         *choiceWindowNode
	rpanel               *rpanelNode
	cursor               *gameui.CursorNode
	radar                *radarNode
	colonySelector       *ge.Sprite
	flyingColonySelector *ge.Sprite
	fogOfWar             *ebiten.Image

	exitButtonRect   gmath.Rect
	toggleButtonRect gmath.Rect
}

func newHumanPlayer(world *worldState, state *playerState, h gameinput.Handler, cursor *gameui.CursorNode, choiceGen *choiceGenerator) *humanPlayer {
	return &humanPlayer{
		world:     world,
		state:     state,
		scene:     world.rootScene,
		choiceGen: choiceGen,
		input:     h,
		cursor:    cursor,
	}
}

func (p *humanPlayer) Init() {
	p.state.Init(p.world)

	choicesPos := gmath.Vec{
		X: 960 - 232 - 16,
		Y: 540 - 200 - 16,
	}
	p.choiceWindow = newChoiceWindowNode(choicesPos, p.world, p.input, p.cursor)
	p.world.nodeRunner.AddObject(p.choiceWindow)

	if p.world.config.ExtraUI {
		p.rpanel = newRpanelNode(p.world, p.world.uiLayer)
		p.scene.AddObject(p.rpanel)
	}

	buttonSize := gmath.Vec{X: 32, Y: 36}
	if p.world.config.EnemyBoss {
		p.radar = newRadarNode(p.world, p.world.uiLayer)
		p.world.nodeRunner.AddObject(p.radar)

		toggleButtonOffset := gmath.Vec{X: 155, Y: 491}
		p.toggleButtonRect = gmath.Rect{Min: toggleButtonOffset, Max: toggleButtonOffset.Add(buttonSize)}

		exitButtonOffset := gmath.Vec{X: 211, Y: 491}
		p.exitButtonRect = gmath.Rect{Min: exitButtonOffset, Max: exitButtonOffset.Add(buttonSize)}
	} else {
		buttonsImage := p.scene.NewSprite(assets.ImageRadarlessButtons)
		buttonsImage.Centered = false
		p.world.uiLayer.AddGraphics(buttonsImage)
		buttonsImage.Pos.Offset = gmath.Vec{
			X: 8,
			Y: p.world.camera.Rect.Height() - buttonsImage.ImageHeight() - 8,
		}

		toggleButtonOffset := (gmath.Vec{X: 13, Y: 23}).Add(buttonsImage.Pos.Offset)
		p.toggleButtonRect = gmath.Rect{Min: toggleButtonOffset, Max: toggleButtonOffset.Add(buttonSize)}

		exitButtonOffset := (gmath.Vec{X: 69, Y: 23}).Add(buttonsImage.Pos.Offset)
		p.exitButtonRect = gmath.Rect{Min: exitButtonOffset, Max: exitButtonOffset.Add(buttonSize)}
	}

	p.colonySelector = p.scene.NewSprite(assets.ImageColonyCoreSelector)
	p.world.camera.AddSpriteBelow(p.colonySelector)
	p.flyingColonySelector = p.scene.NewSprite(assets.ImageColonyCoreSelector)
	p.world.camera.AddSpriteSlightlyAbove(p.flyingColonySelector)

	p.selectNextColony(true)
	p.world.camera.CenterOn(p.state.selectedColony.pos)

	if p.world.config.FogOfWar && p.world.config.ExecMode != gamedata.ExecuteSimulation {
		p.fogOfWar = ebiten.NewImage(int(p.world.width), int(p.world.height))
		gedraw.DrawRect(p.fogOfWar, p.world.rect, color.RGBA{A: 255})
		p.world.camera.SetFogOfWar(p.fogOfWar)

		p.updateFogOfWar(p.state.selectedColony.pos)
	}

	p.choiceGen.EventChoiceReady.Connect(p, p.choiceWindow.RevealChoices)
	p.choiceGen.EventChoiceSelected.Connect(p, func(choice selectedChoice) {
		p.choiceWindow.StartCharging(choice.Cooldown, choice.Index)
		if p.rpanel != nil && choice.Index != -1 && choice.Option.special == specialChoiceNone {
			p.rpanel.UpdateMetrics()
		}
	})
}

func (p *humanPlayer) IsDisposed() bool { return false }

func (p *humanPlayer) GetState() *playerState { return p.state }

func (p *humanPlayer) Update(delta float64) {
	p.choiceWindow.Enabled = p.state.selectedColony != nil &&
		p.state.selectedColony.mode == colonyModeNormal

	if p.world.config.FogOfWar && p.world.config.ExecMode != gamedata.ExecuteSimulation {
		for _, colony := range p.state.colonies {
			if !colony.IsFlying() {
				continue
			}
			p.updateFogOfWar(colony.spritePos)
		}
	}

	if p.state.selectedColony != nil {
		flying := p.state.selectedColony.IsFlying()
		p.colonySelector.Visible = !flying
		p.flyingColonySelector.Visible = flying
	}
}

func (p *humanPlayer) HandleInput() {
	selectedColony := p.state.selectedColony

	if p.input.ActionIsJustPressed(controls.ActionToggleColony) {
		p.onToggleButtonClicked()
		return
	}

	if selectedColony != nil && p.world.movementEnabled {
		if pos, ok := p.cursor.ClickPos(controls.ActionMoveChoice); ok {
			globalClickPos := pos.Add(p.world.camera.Offset)
			if globalClickPos.DistanceTo(selectedColony.pos) > 28 {
				if !p.choiceGen.TryExecute(-1, globalClickPos) {
					p.scene.Audio().PlaySound(assets.AudioError)
				}
				return
			}
		}
	}

	if cardIndex := p.choiceWindow.HandleInput(); cardIndex != -1 {
		if !p.choiceGen.TryExecute(cardIndex, gmath.Vec{}) {
			p.scene.Audio().PlaySound(assets.AudioError)
		}
		return
	}

	handledClick := false
	clickPos, hasClick := p.cursor.ClickPos(controls.ActionClick)
	if len(p.state.colonies) > 1 {
		if hasClick {
			clickPos := clickPos.Add(p.world.camera.Offset)
			selectDist := 40.0
			if p.world.deviceInfo.IsMobile {
				selectDist = 80.0
			}
			var closestColony *colonyCoreNode
			closestDist := math.MaxFloat64
			for _, colony := range p.state.colonies {
				if colony == p.state.selectedColony {
					continue
				}
				dist := colony.pos.DistanceTo(clickPos)
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
	// if p.exitButtonRect.Contains(clickPos) {
	// 	p.onExitButtonClicked()
	// 	return
	// }
	if p.toggleButtonRect.Contains(clickPos) {
		p.onToggleButtonClicked()
		return
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
		return
	}
	p.state.selectedColony.EventDestroyed.Connect(p, func(_ *colonyCoreNode) {
		p.selectNextColony(false)
	})
	p.state.selectedColony.EventTeleported.Connect(p, func(colony *colonyCoreNode) {
		p.state.camera.ToggleCamera(colony.pos)
		if p.world.config.FogOfWar && p.world.config.ExecMode != gamedata.ExecuteSimulation {
			p.updateFogOfWar(colony.pos)
		}
	})
	if p.rpanel != nil {
		p.state.selectedColony.EventPrioritiesChanged.Connect(p, func(_ *colonyCoreNode) {
			p.rpanel.UpdateMetrics()
		})
	}
	p.colonySelector.Pos.Base = &p.state.selectedColony.spritePos
	p.flyingColonySelector.Pos.Base = &p.state.selectedColony.spritePos
}

func (p *humanPlayer) updateFogOfWar(pos gmath.Vec) {
	var options ebiten.DrawImageOptions
	options.CompositeMode = ebiten.CompositeModeDestinationOut
	options.GeoM.Translate(pos.X-colonyVisionRadius, pos.Y-colonyVisionRadius)
	p.fogOfWar.DrawImage(p.world.visionCircle, &options)
}

func (p *humanPlayer) onPanelUpdateRequested(gsignal.Void) {
	p.rpanel.UpdateMetrics()
}

func (p *humanPlayer) onToggleButtonClicked() {
	p.selectNextColony(true)
}
