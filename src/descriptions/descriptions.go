package descriptions

import (
	"fmt"
	"strings"
	"time"

	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

func ReplayText(d *langs.Dictionary, r *session.SavedReplay) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%s [%s]", d.Get("menu.play", r.Replay.Config.RawGameMode), timeutil.FormatDateISO8601(r.Date, true)))
	lines = append(lines, "")
	playerModeValues := []string{
		d.Get("menu.lobby.player_mode.single_player"),
		d.Get("menu.lobby.player_mode.single_bot"),
		d.Get("menu.lobby.player_mode.player_and_bot"),
		d.Get("menu.lobby.player_mode.two_players"),
		d.Get("menu.lobby.player_mode.two_bots"),
	}
	lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.lobby.players"), playerModeValues[r.Replay.Config.PlayersMode]))
	if r.Replay.Config.RawGameMode != "inf_arena" {
		resultsKey := r.ResultTag
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.replay.game_result"), strings.ToLower(d.Get(resultsKey))))
	}
	lines = append(lines, fmt.Sprintf("%s: %d%%", d.Get("menu.lobby.tab.difficulty"), r.Replay.Config.DifficultyScore))

	showScore := false
	switch r.Replay.Config.RawGameMode {
	case "arena", "classic":
		showScore = r.Replay.Results.Victory
	case "inf_arena":
		showScore = true
	case "reverse":
		showScore = r.Replay.Config.PlayersMode == serverapi.PmodeSinglePlayer
	}
	if showScore {
		lines = append(lines, fmt.Sprintf("%s: %d", d.Get("menu.results.score"), r.Replay.Results.Score))
	}

	timePlayed := time.Second * time.Duration(r.Replay.Results.Time)
	lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.results.time_played"), timeutil.FormatDurationCompact(timePlayed)))
	gameSpeedValues := []string{"x1.0", "x1.2", "x1.5"}
	lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.lobby.game_speed"), gameSpeedValues[r.Replay.Config.GameSpeed]))

	return strings.Join(lines, "\n")
}

func LockedDroneText(d *langs.Dictionary, stats *session.PlayerStats, drone *gamedata.AgentStats) string {
	textLines := make([]string, 0, 4)
	textLines = append(textLines, d.Get("drone.locked"))
	textLines = append(textLines, "")
	textLines = append(textLines, fmt.Sprintf("%s: %d/%d", d.Get("drone.score_required"), stats.TotalScore, drone.ScoreCost))
	return strings.Join(textLines, "\n")
}

func LockedTurretText(d *langs.Dictionary, stats *session.PlayerStats, drone *gamedata.AgentStats) string {
	textLines := make([]string, 0, 4)
	textLines = append(textLines, d.Get("turret.locked"))
	textLines = append(textLines, "")
	textLines = append(textLines, fmt.Sprintf("%s: %d/%d", d.Get("drone.score_required"), stats.TotalScore, drone.ScoreCost))
	return strings.Join(textLines, "\n")
}

func TurretText(d *langs.Dictionary, turret *gamedata.AgentStats) string {
	key := strings.ToLower(turret.Kind.String())

	textLines := make([]string, 0, 6)

	textLines = append(textLines, d.Get("turret", key)+"\n")
	textLines = append(textLines, d.Get("turret", key, "description")+"\n")

	if turret.Weapon != nil {
		parts := make([]string, 0, 2)
		if turret.Weapon.TargetFlags&gamedata.TargetGround != 0 {
			p := d.Get("drone.target.ground")
			if turret.Weapon.GroundDamageBonus != 0 {
				p += fmt.Sprintf(" (%d%%)", int(turret.Weapon.GroundDamageBonus*100))
			}
			parts = append(parts, p)
		}
		if turret.Weapon.TargetFlags&gamedata.TargetFlying != 0 {
			p := d.Get("drone.target.flying")
			if turret.Weapon.FlyingDamageBonus != 0 {
				p += fmt.Sprintf(" (%d%%)", int(turret.Weapon.FlyingDamageBonus*100))
			}
			parts = append(parts, p)
		}
		textLines = append(textLines, fmt.Sprintf("%s: %s\n", d.Get("drone.target"), strings.Join(parts, ", ")))
	}

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
	default:
		if drone.Kind == gamedata.AgentRoomba {
			tag = d.Get("drone", "kind", "military")
		}
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
