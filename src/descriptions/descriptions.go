package descriptions

import (
	"fmt"
	"strings"

	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/session"
)

func LockedDroneText(d *langs.Dictionary, stats *session.PlayerStats, drone *gamedata.AgentStats) string {
	textLines := make([]string, 0, 4)
	textLines = append(textLines, d.Get("drone.locked"))
	textLines = append(textLines, "")
	textLines = append(textLines, fmt.Sprintf("%s: %d/%d", d.Get("drone.score_required"), stats.TotalScore, drone.ScoreCost))
	return strings.Join(textLines, "\n")
}

func DroneText(d *langs.Dictionary, drone *gamedata.AgentStats, showTier bool) string {
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

	if showTier {
		textLines = append(textLines, fmt.Sprintf("%s (%s %d)\n", d.Get("drone", key), d.Get("menu.tier"), drone.Tier))
	} else {
		textLines = append(textLines, d.Get("drone", key)+"\n")
	}
	textLines = append(textLines, fmt.Sprintf("%s: %s\n", d.Get("drone.function"), tag))
	textLines = append(textLines, d.Get("drone", key, "description")+"\n")

	if drone.Weapon != nil {
		parts := make([]string, 0, 2)
		if drone.Weapon.TargetFlags&gamedata.TargetGround != 0 {
			p := d.Get("drone.target.ground")
			if drone.Weapon.GroundDamageBonus != 0 {
				p += fmt.Sprintf(" (%d%%)", int(drone.Weapon.GroundDamageBonus*100))
			}
			parts = append(parts, p)
		}
		if drone.Weapon.TargetFlags&gamedata.TargetFlying != 0 {
			p := d.Get("drone.target.flying")
			if drone.Weapon.FlyingDamageBonus != 0 {
				p += fmt.Sprintf(" (%d%%)", int(drone.Weapon.FlyingDamageBonus*100))
			}
			parts = append(parts, p)
		}
		textLines = append(textLines, fmt.Sprintf("%s: %s\n", d.Get("drone.target"), strings.Join(parts, ", ")))
	}

	return strings.Join(textLines, "\n")
}
