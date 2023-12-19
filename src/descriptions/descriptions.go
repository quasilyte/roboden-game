package descriptions

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

func SchemaText(d *langs.Dictionary, id int, schema *gamedata.SavedSchema) string {
	var lines []string

	title := schema.Name
	if title == "" {
		title = fmt.Sprintf("Schema %d", id+1)
	}

	lines = append(lines, fmt.Sprintf("%s [%s]", title, timeutil.FormatDateISO8601(schema.Date, true)))
	lines = append(lines, "")

	difficulty := gamedata.CalcDifficultyScore(schema.Config, gamedata.CalcAllocatedPoints(schema.Config.Tier2Recipes))
	lines = append(lines, fmt.Sprintf("%s: %d%%", d.Get("menu.schema.difficulty"), difficulty))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.schema.colony"), d.Get("core", schema.Config.CoreDesign)))
	lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.schema.turret"), d.Get("turret", strings.ToLower(schema.Config.TurretDesign))))
	lines = append(lines, "")

	{
		var allDrones []string
		for _, recipe := range schema.Config.Tier2Recipes {
			allDrones = append(allDrones, d.Get("drone", strings.ToLower(recipe)))
		}
		sort.Strings(allDrones)
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.schema.drones"), strings.Join(allDrones, ", ")))
	}

	return strings.Join(lines, "\n")
}

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
	mismatchSuffix := ""
	if r.Replay.GameVersion != gamedata.BuildNumber {
		mismatchSuffix = " [!] " + d.Get("menu.replace.version_mismatch")
	}
	lines = append(lines, fmt.Sprintf("%s: %d%s", d.Get("menu.main.build"), r.Replay.GameVersion, mismatchSuffix))
	lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.lobby.players"), playerModeValues[r.Replay.Config.PlayersMode]))
	if r.Replay.Config.RawGameMode != "inf_arena" {
		resultsKey := r.ResultTag
		lines = append(lines, fmt.Sprintf("%s: %s", d.Get("menu.replay.game_result"), strings.ToLower(d.Get(resultsKey))))
	}
	lines = append(lines, fmt.Sprintf("%s: %d", d.Get("menu.lobby.game_seed"), r.Replay.Config.Seed))
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
	gameSpeedValues := []string{"x1.0", "x1.2", "x1.5", "x2.0"}
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
	textLines := make([]string, 0, 3)
	textLines = append(textLines, d.Get("turret.locked"))
	textLines = append(textLines, "")
	textLines = append(textLines, fmt.Sprintf("%s: %d/%d", d.Get("drone.score_required"), stats.TotalScore, drone.ScoreCost))
	return strings.Join(textLines, "\n")
}

func CoreText(d *langs.Dictionary, core *gamedata.ColonyCoreStats) string {
	textLines := make([]string, 0, 6)
	textLines = append(textLines, fmt.Sprintf("%s\n", d.Get("core", core.Name)))
	textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(core.MobilityRating), d.Get("core.mobility_rating")))
	textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(core.UnitLimitRating), d.Get("core.unit_limit_rating")))
	textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(core.AttackRating), d.Get("drone.attack_rating")))
	textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(core.DefenseRating), d.Get("drone.defense_rating")))
	textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(core.CapacityRating), d.Get("core.capacity_rating")))

	var traits []string
	switch core {
	case gamedata.DenCoreStats:
		traits = append(traits, d.Get("core.ability.crush"))
		traits = append(traits, d.Get("core.ability.build_discount"))
	case gamedata.ArkCoreStats:
		traits = append(traits, d.Get("core.ability.flying"))
		traits = append(traits, d.Get("core.ability.no_teleporters"))
	case gamedata.TankCoreStats:
		traits = append(traits, d.Get("core.ability.cant_fly"))
		traits = append(traits, d.Get("core.ability.weapons"))
	case gamedata.HiveCoreStats:
		traits = append(traits, d.Get("core.ability.cant_move"))
		traits = append(traits, d.Get("core.ability.drone_control"))
		traits = append(traits, d.Get("core.ability.ground_weapons"))
		traits = append(traits, d.Get("core.ability.cheap_drones"))
		traits = append(traits, d.Get("core.ability.robust"))
	}
	if len(traits) != 0 {
		textLines = append(textLines, "")
	}
	for _, t := range traits {
		textLines = append(textLines, "> "+t)
	}

	return strings.Join(textLines, "\n")
}

func LockedCoreText(d *langs.Dictionary, stats *session.PlayerStats, core *gamedata.ColonyCoreStats) string {
	textLines := make([]string, 0, 3)
	textLines = append(textLines, d.Get("core.locked"))
	textLines = append(textLines, "")
	textLines = append(textLines, fmt.Sprintf("%s: %d/%d", d.Get("drone.score_required"), stats.TotalScore, core.ScoreCost))
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

func RatingBar(value int) string {
	full := value
	empty := 10 - full
	return strings.Repeat("●", full) + strings.Repeat("◌", empty)
}

func DroneText(d *langs.Dictionary, drone *gamedata.AgentStats, showTier, globalStats bool) string {
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
		textLines = append(textLines, fmt.Sprintf("%s (%s, %s %d)\n", d.Get("drone", key), tag, d.Get("menu.tier"), drone.Tier))
	} else {
		textLines = append(textLines, fmt.Sprintf("%s (%s)\n", d.Get("drone", key), tag))
	}

	var docStats *gamedata.DroneDocs
	if globalStats {
		docStats = &drone.GlobalDocs
	} else {
		docStats = &drone.Docs
	}
	if drone.Weapon != nil {
		damageSuffix := ""
		if drone.Weapon.MaxTargets != 1 {
			damageSuffix = " (" + d.Get("drone.attack_rating_multi") + ")"
		}
		textLines = append(textLines, fmt.Sprintf("%s %s%s", RatingBar(docStats.DamageRating), d.Get("drone.dps_rating"), damageSuffix))
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(docStats.AttackRangeRating), d.Get("drone.attack_range_rating")))
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(docStats.DefenseRating), d.Get("drone.defense_rating")))
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(docStats.UpkeepRating), d.Get("drone.upkeep_cost")))

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
		textLines = append(textLines, "")
		textLines = append(textLines, fmt.Sprintf("%s: %s", d.Get("drone.target"), strings.Join(parts, ", ")))
	} else {
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(0), d.Get("drone.dps_rating")))
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(0), d.Get("drone.attack_range_rating")))
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(docStats.DefenseRating), d.Get("drone.defense_rating")))
		textLines = append(textLines, fmt.Sprintf("%s %s", RatingBar(docStats.UpkeepRating), d.Get("drone.upkeep_cost")))
		textLines = append(textLines, "")
		if drone.Kind == gamedata.AgentBomber {
			textLines = append(textLines, fmt.Sprintf("%s: %s", d.Get("drone.target"), d.Get("drone.target.ground")))
		} else {
			textLines = append(textLines, fmt.Sprintf("%s: %s", d.Get("drone.target"), d.Get("drone.target.none")))
		}
	}

	var traits []string
	switch drone.Kind {
	case gamedata.AgentCloner:
		traits = append(traits, d.Get("drone.ability.cloning"))
	case gamedata.AgentRepair:
		traits = append(traits, d.Get("drone.ability.repair"))
	case gamedata.AgentRecharger:
		traits = append(traits, d.Get("drone.ability.recharge"))
	case gamedata.AgentRedminer:
		traits = append(traits, d.Get("drone.ability.red_oil_scavenge"))
	case gamedata.AgentServo:
		traits = append(traits, d.Get("drone.ability.colony_speed"))
		traits = append(traits, d.Get("drone.ability.colony_jump"))
		traits = append(traits, d.Get("drone.ability.discharged_speed"))
	case gamedata.AgentScavenger:
		traits = append(traits, d.Get("drone.ability.scrap_scavenge"))
	case gamedata.AgentCourier, gamedata.AgentTrucker:
		traits = append(traits, d.Get("drone.ability.courier"))
	case gamedata.AgentFreighter:
		traits = append(traits, d.Get("drone.ability.zero_upkeep"))
	case gamedata.AgentGenerator:
		traits = append(traits, d.Get("drone.ability.upkeep_decrease"))
		traits = append(traits, d.Get("drone.ability.energy_regen"))
	case gamedata.AgentStormbringer:
		traits = append(traits, d.Get("drone.ability.upkeep_decrease"))
	case gamedata.AgentRoomba:
		traits = append(traits, d.Get("drone.ability.ground"))
		traits = append(traits, d.Get("drone.ability.map_patrol"))
	case gamedata.AgentDisintegrator:
		traits = append(traits, d.Get("drone.ability.discharged_after_attack"))
	case gamedata.AgentCommander:
		traits = append(traits, d.Get("drone.ability.group_command"))
		traits = append(traits, d.Get("drone.ability.group_buff"))
		traits = append(traits, d.Get("drone.ability.group_buff_def"))
	case gamedata.AgentTargeter:
		traits = append(traits, d.Get("drone.ability.target_marking"))
	case gamedata.AgentKamikaze:
		traits = append(traits, d.Get("drone.ability.kamikaze"))
	case gamedata.AgentPrism:
		traits = append(traits, d.Get("drone.ability.prism_reflect"))
	case gamedata.AgentScarab:
		traits = append(traits, d.Get("drone.ability.scarab_potential"))
	case gamedata.AgentDevourer:
		traits = append(traits, d.Get("drone.ability.consume_for_heal"))
		traits = append(traits, d.Get("drone.ability.consume_for_power"))
	case gamedata.AgentBomber:
		traits = append(traits, d.Get("drone.ability.bomb_attack"))
		traits = append(traits, d.Get("drone.ability.bomb_aoe"))
	}
	if drone.MaxPayload > 1 {
		traits = append(traits, fmt.Sprintf("%s (%d%%)", d.Get("drone.ability.extra_payload"), 100*drone.MaxPayload))
	}
	if drone.SelfRepair != 0 {
		traits = append(traits, d.Get("drone.ability.self_repair"))
	}
	if drone.Weapon != nil {
		if drone.Weapon.MaxTargets > 1 {
			if drone.Weapon.TargetMaxDist == 0 {
				traits = append(traits, fmt.Sprintf(d.Get("drone.ability.num_targets_f"), drone.Weapon.MaxTargets))
			} else {
				traits = append(traits, fmt.Sprintf(d.Get("drone.ability.num_targets_alt_f"), drone.Weapon.MaxTargets-1))
			}
		}
		if drone.Weapon.Damage.HasFlag(gamedata.DmgflagAggro) {
			traits = append(traits, d.Get("drone.ability.aggro"))
		}
		if drone.Weapon.Damage.Disarm > 0 {
			traits = append(traits, d.Get("drone.ability.disarm"))
		}
		if drone.Weapon.Damage.Slow > 0 {
			traits = append(traits, d.Get("drone.ability.slow"))
		}
		if drone.Weapon.BuildingDamageBonus > 0 {
			traits = append(traits, fmt.Sprintf(d.Get("drone.ability.more_building_damage_f"), int(math.Abs(drone.Weapon.BuildingDamageBonus*100))))
		} else if drone.Weapon.BuildingDamageBonus < 0 {
			traits = append(traits, fmt.Sprintf(d.Get("drone.ability.less_building_damage_f"), int(math.Abs(drone.Weapon.BuildingDamageBonus*100))))
		}
	}
	if drone.CanCloak {
		traits = append(traits, d.Get("drone.ability.cloak_hide"))
		if drone.Kind == gamedata.AgentMarauder {
			traits = append(traits, d.Get("drone.ability.cloak_scavenge"))
		}
	}
	if len(traits) != 0 {
		textLines = append(textLines, "")
	}
	for _, t := range traits {
		textLines = append(textLines, "> "+t)
	}

	return strings.Join(textLines, "\n")
}
