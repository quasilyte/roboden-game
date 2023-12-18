package staging

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
)

type tooltipManager struct {
	world *worldState

	player *humanPlayer

	scene *ge.Scene

	allTips bool

	tooltipTime float64

	message *messageNode

	EventHighlightDrones gsignal.Event[*gamedata.AgentStats]
	EventTooltipClosed   gsignal.Event[gsignal.Void]
}

func newTooltipManager(p *humanPlayer, allTips bool) *tooltipManager {
	return &tooltipManager{
		allTips: allTips,
		world:   p.world,
		player:  p,
	}
}

func (m *tooltipManager) Init(scene *ge.Scene) {
	m.scene = scene
}

func (m *tooltipManager) IsDisposed() bool {
	return false
}

func (m *tooltipManager) Update(delta float64) {
	if m.tooltipTime > 0 {
		m.tooltipTime = gmath.ClampMin(m.tooltipTime-delta, 0)
		if m.tooltipTime == 0 {
			m.OnStopHover()
		}
	}
}

func (m *tooltipManager) OnStopHover() {
	m.removeTooltip()
}

func (m *tooltipManager) OnHover(pos gmath.Vec) {
	d := m.scene.Dict()

	if m.player.recipeTab != nil && m.player.recipeTab.Visible {
		drone := m.player.recipeTab.GetDroneUnderCursor(pos.Sub(m.player.state.camera.ScreenPos))
		if drone != nil {
			count := 0
			if m.player.state.selectedColony != nil {
				flags := searchFighters
				if !drone.CanPatrol {
					flags = searchWorkers
				}
				m.player.state.selectedColony.agents.Find(flags, func(a *colonyAgentNode) bool {
					if a.stats == drone {
						count++
					}
					return false
				})
			}
			hint := fmt.Sprintf("%s (%d)", d.Get("drone", strings.ToLower(drone.Kind.String())), count)
			m.createTooltip(pos, hint)
			m.EventHighlightDrones.Emit(drone)
			return
		}
	}

	if m.player.screenButtons != nil {
		button := m.player.screenButtons.GetChoiceUnderCursor(pos.Sub(m.player.state.camera.ScreenPos))
		var hint string
		var buttonAction input.Action
		switch button {
		case screenButtonToggle:
			hint = d.Get("game.hint.screen_button.toggle")
			buttonAction = controls.ActionToggleColony
		case screenButtonExit:
			hint = d.Get("game.hint.screen_button.exit")
			buttonAction = controls.ActionExit
		case screenButtonFastForward:
			hint = d.Get("game.hint.screen_button.fast_forward")
			buttonAction = controls.ActionToggleFastForward
		}
		if hint != "" {
			keyHint := m.player.input.PrettyActionName(d, buttonAction)
			if keyHint != "" {
				hint += " [" + keyHint + "]"
			}
			m.createTooltip(pos, hint)
			return
		}
	}

	if m.player.rpanel != nil {
		item, v := m.player.rpanel.GetItemUnderCursor(pos.Sub(m.player.state.camera.ScreenPos))
		var hint string
		switch item {
		case rpanelItemResourcesPriority:
			hint = fmt.Sprintf("%s: %d%%", d.Get("game.hint.rpanel.priority_resources"), int(math.Round(v*100)))
		case rpanelItemGrowthPriority:
			hint = fmt.Sprintf("%s: %d%%", d.Get("game.hint.rpanel.priority_growth"), int(math.Round(v*100)))
		case rpanelItemEvolutionPriority:
			hint = fmt.Sprintf("%s: %d%%", d.Get("game.hint.rpanel.priority_evolution"), int(math.Round(v*100)))
		case rpanelItemSecurityPriority:
			hint = fmt.Sprintf("%s: %d%%", d.Get("game.hint.rpanel.priority_security"), int(math.Round(v*100)))
		case rpanelItemFactionDistribution:
			hint = d.Get("game.hint.rpanel.factions")
		case rpanelItemTechProgress:
			hint = fmt.Sprintf(d.Get("game.hint.rpanel.tech_progress_f"), int(math.Round(v*100)))
		case rpanelItemGarrison:
			hint = d.Get("game.hint.rpanel.garrison")
		}
		if hint != "" {
			m.createTooltip(pos, hint)
			return
		}
	}

	if !m.player.spectator && m.player.choiceGen.IsReady() && m.player.choiceWindow != nil {
		choice := m.player.choiceWindow.GetChoiceUnderCursor(pos.Sub(m.player.state.camera.ScreenPos))
		if choice != nil {
			var hint string
			if choice.option.special != specialChoiceNone {
				if choice.option.special > _creepCardFirst && choice.option.special < _creepCardLast {
					side := d.Get(sideName(choice.option.direction))
					info := creepOptionInfoList[creepCardID(choice.option.special)]
					hint = fmt.Sprintf(d.Get("game.hint.action.garrison_f"), side) + "\n" +
						fmt.Sprintf("x%d %s", numCreepsPerCard(m.player.creepsState, info), d.Get("creep", info.stats.NameTag))
				} else {
					key := strings.ToLower(choice.option.special.String())
					hint = d.Get("game.hint.action", key)
				}
			} else {
				if len(choice.option.effects) == 1 {
					hint = fmt.Sprintf(d.Get("game.hint.action.priorities1_f"), d.Get("game.choice", strings.ToLower(choice.option.effects[0].priority.String())))
				} else if len(choice.option.effects) == 2 {
					p1 := d.Get("game.choice", strings.ToLower(choice.option.effects[0].priority.String()))
					p2 := d.Get("game.choice", strings.ToLower(choice.option.effects[1].priority.String()))
					hint = fmt.Sprintf(d.Get("game.hint.action.priorities2_f"), p1, p2)
				}
			}
			if hint != "" {
				m.createTooltip(pos, hint)
				return
			}
		}
	}

	if !m.allTips {
		return
	}

	globalPos := m.player.state.camera.AbsClickPos(pos)
	hint := m.findHoverTargetHint(globalPos)
	if hint != "" {
		m.createTooltip(pos, hint)
		return
	}
}

func (m *tooltipManager) inHoverRange(hoverPos, objectPos gmath.Vec, objectSize float64) bool {
	multiplier := 1.0
	switch {
	case m.player.cursor.VirtualCursorIsVisible():
		multiplier = 1.2
	case m.player.input.InputMethod == gameinput.InputMethodTouch:
		multiplier = 1.5
	}
	maxDistSqr := (objectSize * objectSize) * multiplier
	return objectPos.DistanceSquaredTo(hoverPos) < maxDistSqr
}

func (m *tooltipManager) findHoverTargetHint(pos gmath.Vec) string {
	d := m.scene.Dict()

	for _, g := range m.world.lavaGeysers {
		if !m.inHoverRange(pos, g.pos, 26) {
			continue
		}
		return d.Get("game.hint.lava_geyser")
	}

	for _, b := range m.world.neutralBuildings {
		if !m.inHoverRange(pos, b.CurrentPos(), 26) {
			continue
		}
		status := "needs_repair"
		if b.agent != nil {
			status = "functioning"
		}
		var tag string
		switch b.stats {
		case gamedata.DroneFactoryAgentStats:
			tag = "drone_factory"
		case gamedata.PowerPlantAgentStats:
			tag = "power_plant"
		case gamedata.RepulseTowerAgentStats:
			tag = "tower"
		case gamedata.MegaRoombaAgentStats:
			tag = "megaroomba"
		}
		return d.Get("game.hint.building", tag) + "\n" + d.Get("game.hint.building_status", status)
	}

	for _, tp := range m.world.teleporters {
		if !m.inHoverRange(pos, tp.pos, 26) {
			continue
		}
		label := "A"
		if tp.id != 0 {
			label = "B"
		}
		return d.Get("game.hint.teleporter") + " " + label
	}

	for _, turret := range m.world.turrets {
		if !m.inHoverRange(pos, turret.pos, 16) {
			continue
		}
		if turret.stats.IsNeutral {
			continue
		}
		return d.Get("turret", strings.ToLower(turret.stats.Kind.String()))
	}

	for _, c := range m.player.state.colonies {
		if !m.inHoverRange(pos, c.pos, 26) {
			continue
		}
		return d.Get("game.hint.colony") + " " + strconv.Itoa(c.id)
	}

	if m.world.boss != nil && m.inHoverRange(pos, m.world.boss.pos, 22) {
		boss := m.world.boss
		hpPercentage := boss.health / boss.maxHealth
		return fmt.Sprintf("%s (%d%%)", d.Get("game.hint.dreadnought"), int(math.Round(100*hpPercentage)))
	}

	{
		creep := m.world.WalkCreepsWithRand(nil, pos, 32, func(creep *creepNode) bool {
			return m.inHoverRange(pos, creep.pos, 12) && creep.stats.Building
		})
		if creep != nil {
			var tag string
			switch creep.stats {
			case gamedata.CrawlerBaseCreepStats:
				tag = "creep.hint.crawler_base"
			case gamedata.BaseCreepStats:
				tag = "creep.hint.base"
			case gamedata.TurretCreepStats:
				tag = "creep.turret"
			case gamedata.IonMortarCreepStats:
				tag = "creep.ion_mortar"
			case gamedata.FortressCreepStats:
				tag = "creep.fortress"
			}
			if tag != "" {
				return d.Get(tag)
			}
		}
	}

	if m.world.wispLair != nil && m.inHoverRange(pos, m.world.wispLair.pos, 22) {
		return d.Get("game.hint.wisp_lair")
	}

	for _, res := range m.world.essenceSources {
		if !m.inHoverRange(pos, res.pos, 18) {
			continue
		}
		return d.Get("game.hint.resource", res.stats.name) + "\n" + d.Get("game.hint.resource", res.stats.name, "value")
	}

	return ""
}

func (m *tooltipManager) removeTooltip() {
	m.tooltipTime = 0
	if m.message != nil {
		m.message.Dispose()
		m.message = nil
		m.EventTooltipClosed.Emit(gsignal.Void{})
	}
}

func (m *tooltipManager) createTooltip(pos gmath.Vec, s string) {
	if m.message != nil {
		m.removeTooltip()
	}

	m.tooltipTime = 5
	camera := m.player.state.camera.Camera

	messagePos := pos.Sub(camera.ScreenPos)
	w, h := estimateMessageBounds(s, 0)
	if w+messagePos.X+16 > camera.Rect.Max.X {
		messagePos.X -= w
	}
	if h+messagePos.Y+16 > camera.Rect.Max.Y {
		messagePos.Y -= h + 16
	} else {
		messagePos.Y += 16
	}

	m.message = newScreenTutorialHintNode(camera, messagePos, gmath.Vec{}, s)
	m.scene.AddObject(m.message)
}
