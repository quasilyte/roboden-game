package staging

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type tooltipManager struct {
	world *worldState

	player *humanPlayer

	scene *ge.Scene

	allTips bool

	tooltipTime float64

	message *messageNode
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
			m.createTooltip(pos, d.Get("drone", strings.ToLower(drone.Kind.String())))
			return
		}
	}

	if m.player.screenButtons != nil {
		button := m.player.screenButtons.GetChoiceUnderCursor(pos.Sub(m.player.state.camera.ScreenPos))
		var hint string
		switch button {
		case screenButtonToggle:
			hint = d.Get("game.hint.screen_button.toggle")
		case screenButtonExit:
			hint = d.Get("game.hint.screen_button.exit")
		}
		if hint != "" {
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

	if m.player.choiceGen.IsReady() && m.player.choiceWindow != nil {
		choice := m.player.choiceWindow.GetChoiceUnderCursor(pos.Sub(m.player.state.camera.ScreenPos))
		if choice != nil {
			var hint string
			if choice.option.special != specialChoiceNone {
				if choice.option.special > _creepCardFirst && choice.option.special < _creepCardLast {
					side := d.Get(sideName(choice.option.direction))
					hint = fmt.Sprintf(d.Get("game.hint.action.garrison_f"), side)
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

func (m *tooltipManager) findHoverTargetHint(pos gmath.Vec) string {
	d := m.scene.Dict()

	for _, b := range m.world.neutralBuildings {
		if b.pos.DistanceSquaredTo(pos) > (26 * 26) {
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
		case gamedata.TowerArtifactAgentStats:
			tag = "tower"
		}
		return d.Get("game.hint.building", tag) + "\n" + d.Get("game.hint.building_status", status)
	}

	for _, tp := range m.world.teleporters {
		if tp.pos.DistanceSquaredTo(pos) > (26 * 26) {
			continue
		}
		label := "A"
		if tp.id != 0 {
			label = "B"
		}
		return d.Get("game.hint.teleporter") + " " + label
	}

	for _, c := range m.player.state.colonies {
		if c.pos.DistanceSquaredTo(pos) > (26 * 26) {
			continue
		}
		return d.Get("game.hint.colony") + " " + strconv.Itoa(c.id)
	}

	if m.world.boss != nil && m.world.boss.pos.DistanceSquaredTo(pos) < (22*22) {
		boss := m.world.boss
		hpPercentage := boss.health / boss.maxHealth
		return fmt.Sprintf("%s (%d%%)", d.Get("game.hint.dreadnought"), int(math.Round(100*hpPercentage)))
	}

	if m.world.wispLair != nil && m.world.wispLair.pos.DistanceSquaredTo(pos) < (22*22) {
		return d.Get("game.hint.wisp_lair")
	}

	for _, res := range m.world.essenceSources {
		if res.pos.DistanceSquaredTo(pos) > (16 * 16) {
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
