package menus

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/contentlock"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/steamsdk"
)

type TerminalMenu struct {
	state *session.State

	errorSoundDelay float64

	outputLabel *widget.Text

	scene *ge.Scene
}

type terminalCommandContext struct {
	fs         *flag.FlagSet
	parsedArgs any
}

func NewTerminalMenuController(state *session.State) *TerminalMenu {
	return &TerminalMenu{state: state}
}

func (c *TerminalMenu) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *TerminalMenu) Update(delta float64) {
	c.errorSoundDelay = gmath.ClampMin(c.errorSoundDelay-delta, 0)
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *TerminalMenu) commandOutputText(s string) string {
	if s == "" {
		return "<" + c.scene.Dict().Get("menu.option.none") + ">"
	}
	return s
}

func (c *TerminalMenu) setOutput(s string) {
	c.outputLabel.Label = strings.ReplaceAll(s, "\t", "  ")
}

func (c *TerminalMenu) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	tinyFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.settings")+" -> "+d.Get("menu.options.extra")+" -> "+d.Get("menu.terminal"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	outputPanel := eui.NewTextPanel(uiResources, 520, 200)

	normalContainer := eui.NewAnchorContainer()
	outputTitle := eui.NewCenteredLabel(d.Get("menu.terminal.command_output_label"), tinyFont)
	c.outputLabel = eui.NewLabel(c.commandOutputText(""), tinyFont)
	c.outputLabel.MaxWidth = 500
	normalContainer.AddChild(c.outputLabel)
	outputPanel.AddChild(normalContainer)

	type terminalCommand struct {
		key     string
		handler func(*terminalCommandContext) (string, error)
		hidden  bool
	}
	var knownCommands []*terminalCommand
	knownCommands = []*terminalCommand{
		{
			key: "help",
			handler: func(*terminalCommandContext) (string, error) {
				var commands []string
				for _, cmd := range knownCommands {
					if !cmd.hidden {
						commands = append(commands, cmd.key)
					}
				}
				lines := []string{
					"This prompt is a tool to help game testers and developers to work on this game.",
					"List of known commands: " + strings.Join(commands, ", ") + ".",
					"To get command-related help, use `<command> --help`.",
					"For example: `save.delete --help`.",
				}
				return strings.Join(lines, "\n"), nil
			},
		},

		{
			key:     "logs.grep",
			handler: c.onLogsGrep,
		},
		{
			key:     "save.info",
			handler: c.onSaveInfo,
		},
		{
			key:     "save.delete",
			handler: c.onSaveDelete,
		},
		{
			key:     "cheat.add_score",
			handler: c.onCheatAddScore,
			hidden:  true,
		},
		{
			key:     "balance.drones_calc",
			handler: c.onBalanceDronesCalc,
			hidden:  true,
		},
		{
			key:     "debug.logs",
			handler: c.onDebugLogs,
		},
		{
			key:     "debug.drone_labels",
			handler: c.onDebugDroneLabels,
		},
		{
			key:     "replay.dump",
			handler: c.onReplayDump,
		},

		{
			key:     "steam.clear_achievements",
			handler: c.onSteamClearAchievements,
		},
		{
			key:     "steam.list_achievements",
			handler: c.onSteamListAchievements,
		},
	}

	textinput := eui.NewTextInput(uiResources, eui.TextInputConfig{SteamDeck: c.state.SteamInfo.SteamDeck},
		widget.TextInputOpts.Placeholder(d.Get("menu.terminal.placeholder")),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			if args.InputText == "" {
				return
			}
			key, rest, _ := strings.Cut(args.InputText, " ")
			i := xslices.IndexWhere(knownCommands, func(cmd *terminalCommand) bool {
				return cmd.key == key
			})
			if i == -1 {
				c.scene.Audio().PlaySound(assets.AudioError)
				c.setOutput(c.commandOutputText(fmt.Sprintf("unknown %q command", key)))
				return
			}
			cmd := knownCommands[i]
			ctx := &terminalCommandContext{
				fs: flag.NewFlagSet(key, flag.ContinueOnError),
			}
			var cmdUsageOutput bytes.Buffer
			ctx.fs.SetOutput(&cmdUsageOutput)
			out, err := cmd.handler(ctx)
			if err != nil {
				c.scene.Audio().PlaySound(assets.AudioError)
				c.setOutput(c.commandOutputText(err.Error()))
				return
			}
			if ctx.parsedArgs != nil {
				if err := ctx.fs.Parse(strings.Fields(rest)); err != nil {
					c.scene.Audio().PlaySound(assets.AudioError)
					var s string
					if cmdUsageOutput.Len() != 0 {
						s = cmdUsageOutput.String()
					} else {
						s = err.Error()
					}
					c.setOutput(c.commandOutputText(s))
					return
				}
				out, err = cmd.handler(ctx)
				if err != nil {
					c.scene.Audio().PlaySound(assets.AudioError)
					c.setOutput(c.commandOutputText(err.Error()))
					return
				}
			}
			c.maybeGrantAchievement()
			c.setOutput(c.commandOutputText(out))
			c.scene.Audio().PlaySound(assets.AudioChoiceMade)
		}),
		widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
			good := true
			if len(newInputText) > 36 {
				good = false
			}
			if good {
				for _, ch := range newInputText {
					if !unicode.IsPrint(ch) || ch >= utf8.RuneSelf {
						good = false
						break
					}
				}
			}
			if !good && c.errorSoundDelay == 0 {
				c.scene.Audio().PlaySound(assets.AudioError)
				c.errorSoundDelay = 0.2
			}
			return good, nil
		}),
	)
	rowContainer.AddChild(textinput)
	rowContainer.AddChild(outputTitle)
	rowContainer.AddChild(outputPanel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *TerminalMenu) maybeGrantAchievement() {
	c.state.UnlockAchievement(session.Achievement{
		Name:  "terminal",
		Elite: true,
	})
}

func (c *TerminalMenu) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsExtraMenuController(c.state))
}

func (c *TerminalMenu) achievementNames() []string {
	names := make([]string, len(gamedata.AchievementList))
	for i, a := range gamedata.AchievementList {
		names[i] = a.Name
	}
	return names
}

func (c *TerminalMenu) onSteamListAchievements(ctx *terminalCommandContext) (string, error) {
	if !c.state.SteamInfo.Enabled {
		return "", errors.New("steam is not enabled")
	}
	var allUnlocked []string
	for _, name := range c.achievementNames() {
		unlocked, err := steamsdk.IsAchievementUnlocked(name)
		if err == nil && unlocked {
			allUnlocked = append(allUnlocked, name)
		}
	}
	return strings.Join(allUnlocked, ", "), nil
}

func (c *TerminalMenu) onSteamClearAchievements(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		confirm bool
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.BoolVar(&args.confirm, "confirm", false, "acknowledge the risks, clear the Steam achievements")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	if !args.confirm {
		lines := []string{
			"This operation will clear Steam achievements.",
			"Provide a --confirm flag to perform this operation.",
		}
		return strings.Join(lines, "\n"), nil
	}
	if !c.state.SteamInfo.Enabled {
		return "", errors.New("steam is not enabled")
	}
	steamsdk.ClearAchievements(c.achievementNames())
	return "The Steam achievements are cleared.", nil
}

func (c *TerminalMenu) onSaveDelete(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		confirm bool
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.BoolVar(&args.confirm, "confirm", false, "acknowledge the risks, remove the save file")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	if !args.confirm {
		lines := []string{
			"This operation will remove all saved data, including all setting changes.",
			"Provide a --confirm flag to perform this operation.",
		}
		return strings.Join(lines, "\n"), nil
	}
	c.state.Persistent = contentlock.GetDefaultData()
	contentlock.Update(c.state)
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.state.ReloadLanguage(c.scene.Context())
	return "The save data is cleared.", nil
}

func (c *TerminalMenu) onLogsGrep(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		pattern string
		head    bool
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.StringVar(&args.pattern, "pattern", ".*", "regexp pattern to apply")
		ctx.fs.BoolVar(&args.head, "head", false, "reverse the search direction, scan first log entries first")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	re, err := regexp.Compile(args.pattern)
	if err != nil {
		return "", err
	}
	results := make([]string, 0, 10)
	start := 0
	end := len(c.state.StdoutLogs)
	step := 1
	if !args.head {
		start = len(c.state.StdoutLogs) - 1
		end = -1
		step = -1
	}
	numLines := 0
	for i := start; i != end; i += step {
		if numLines >= 10 {
			break
		}
		l := c.state.StdoutLogs[i]
		if !re.MatchString(l) {
			continue
		}
		lines := strings.Split(l, "\n")
		// We'll reverse lines later, if it's a multi-line text,
		// append it to the result in reversed order.
		for i := len(lines) - 1; i >= 0; i-- {
			subLine := lines[i]
			if numLines >= 10 {
				break
			}
			numLines++
			if i == 0 {
				results = append(results, "> "+subLine)
			} else {
				results = append(results, "  "+subLine)
			}
		}
	}
	if !args.head {
		reverseStrings(results)
	}
	return strings.Join(results, "\n"), nil
}

func (c *TerminalMenu) onSaveInfo(*terminalCommandContext) (string, error) {
	lines := []string{
		fmt.Sprintf("Save file: %q", c.scene.Context().LocateGameData("save")),
	}
	return strings.Join(lines, "\n"), nil
}

type droneScore struct {
	name         string
	computed     float64
	score        float64
	defenseScore float64
	weaponScore  float64
	pointCost    int
	cost         int
	upkeep       int
}

func (c *TerminalMenu) sortDroneScoreList(list []droneScore) {
	sort.SliceStable(list, func(i, j int) bool {
		score1 := list[i].computed
		score2 := list[j].computed
		return score1 > score2
	})
}

func (c *TerminalMenu) calcDroneScore(drone *gamedata.AgentStats) droneScore {
	defenseScore := 0.0
	maxHealth := drone.MaxHealth
	if drone.Kind == gamedata.AgentDevourer {
		maxHealth += 5 * 6
	}
	defenseScore += 0.4 * maxHealth
	defenseScore += 7 * drone.SelfRepair
	if drone.CanCloak {
		defenseScore += 5
	}

	score := 0.0
	if drone.MaxPayload > 1 {
		extraPayload := float64(drone.MaxPayload - 1)
		if drone.CanGather {
			score += 2.0 * extraPayload
		} else {
			score += 1.0 * extraPayload
		}
	}
	if drone.CanGather {
		score += 5
		score += 0.15 * drone.Speed
		score += 5 * drone.EnergyRegenRateBonus
	} else {
		score += 0.05 * drone.Speed
		score += 2.5 * drone.EnergyRegenRateBonus
	}
	if drone.CanPatrol {
		score += defenseScore
	} else {
		score += 0.4 * defenseScore
	}

	// Rate the special abilities.
	switch drone.Kind {
	case gamedata.AgentCloner, gamedata.AgentRepair:
		score += 18
	case gamedata.AgentRedminer:
		score += 14
	case gamedata.AgentServo:
		score += 12
	case gamedata.AgentCourier, gamedata.AgentTrucker, gamedata.AgentRecharger:
		score += 8
	case gamedata.AgentScarab:
		// Becomes much better if kept safe.
		score += 8
	case gamedata.AgentKamikaze:
		score += 6
	case gamedata.AgentGenerator, gamedata.AgentScavenger, gamedata.AgentMarauder:
		score += 4
	}

	weaponScore := 0.0
	if drone.Weapon != nil {
		projectileScore := 0.0
		healthDamage := drone.Weapon.Damage.Health
		switch drone.Kind {
		case gamedata.AgentPrism:
			// Prisms connect their attacks, so the effective damage
			// is usually higher.
			healthDamage += 6
		}
		projectileScore += 2.5 * healthDamage
		projectileScore += 0.5 * drone.Weapon.Damage.Disarm
		projectileScore += 5.0 * drone.Weapon.Damage.Aggro
		projectileScore += 0.25 * drone.Weapon.Damage.Morale
		projectileScore += 0.5 * drone.Weapon.Damage.Energy
		projectileScore += 0.75 * drone.Weapon.Damage.Slow

		if drone.Weapon.TargetFlags == gamedata.TargetFlying|gamedata.TargetGround {
			// Can attack both kinds of targets.
			// There can be a penalty against some targets, so we take that into account as well.
			flyingMultiplier := 0.6 * drone.Weapon.FlyingTargetDamageMult
			groundMultiplier := 0.4 * drone.Weapon.GroundTargetDamageMult
			projectileScore *= (flyingMultiplier + groundMultiplier)
		} else {
			// Can attack only 1 kind of targets.
			if drone.Weapon.TargetFlags == gamedata.TargetFlying {
				projectileScore *= (0.6 + 0.05)
			} else {
				projectileScore *= (0.4 + 0.05)
			}
		}

		burstScore := projectileScore
		burstSize := drone.Weapon.BurstSize
		if drone.Kind == gamedata.AgentDevourer {
			burstSize += 6
		}
		if drone.Weapon.BurstSize > 1 {
			burstScore *= float64(burstSize)
		}
		if drone.Weapon.MaxTargets > 1 {
			burstScore *= (0.7 * float64(drone.Weapon.MaxTargets))
		}
		shotsPerSecond := 1.0 / drone.Weapon.Reload
		rangeScore := gmath.ClampMin((drone.Weapon.AttackRange-80)*0.04, 0)
		if drone.Weapon.TargetFlags&gamedata.TargetFlying != 0 && drone.Weapon.AttackRange > 250 {
			// Can outrange flying uber boss.
			rangeScore += 2.5
		} else if drone.Weapon.TargetFlags&gamedata.TargetGround != 0 && drone.Weapon.AttackRange > 320 {
			// Can outrange ground turrets.
			rangeScore += 1.5
		}
		weaponScore = 3 * ((burstScore + rangeScore) * shotsPerSecond)
	}

	computedScore := score + weaponScore
	computedScore *= (1.0 - (0.8 * (float64(drone.PointCost) / 20.0)))
	computedScore -= 0.35 * float64(drone.Upkeep)
	computedScore -= 0.1 * float64(drone.Cost)

	return droneScore{
		name:         drone.Kind.String(),
		computed:     computedScore,
		score:        score,
		defenseScore: defenseScore,
		weaponScore:  weaponScore,
		pointCost:    drone.PointCost,
		cost:         int(drone.Cost),
		upkeep:       drone.Upkeep,
	}
}

func (c *TerminalMenu) dumpList(tag string, list []droneScore) {
	println(fmt.Sprintf("%s (with weapon):", tag))
	for _, d := range list {
		if d.weaponScore == 0 {
			continue
		}
		s := fmt.Sprintf("> %s [%s] c=%.2f score=%.2f def=%.2f weapon=%.2f (total %.2f) cost=%d upkeep=%d",
			d.name,
			strings.Repeat("*", d.pointCost),
			d.computed,
			d.score, d.defenseScore, d.weaponScore, d.score+d.weaponScore, d.cost, d.upkeep)
		println(s)
	}
	println(fmt.Sprintf("%s (workers):", tag))
	for _, d := range list {
		if d.weaponScore != 0 {
			continue
		}
		s := fmt.Sprintf("> %s [%s] c=%.2f score=%.2f def=%.2f cost=%d upkeep=%d",
			d.name,
			strings.Repeat("*", d.pointCost),
			d.computed,
			d.score, d.defenseScore, d.cost, d.upkeep)
		println(s)
	}
}

func (c *TerminalMenu) onBalanceDronesCalc(*terminalCommandContext) (string, error) {
	var tier2 []droneScore
	for _, recipe := range gamedata.Tier2agentMergeRecipes {
		tier2 = append(tier2, c.calcDroneScore(recipe.Result))
	}
	c.sortDroneScoreList(tier2)
	c.dumpList("tier 2", tier2)

	var tier3 []droneScore
	for _, recipe := range gamedata.Tier3agentMergeRecipes {
		tier3 = append(tier3, c.calcDroneScore(recipe.Result))
	}
	c.sortDroneScoreList(tier3)
	c.dumpList("tier 3", tier3)

	return "Drones balance report is dumped to the console", nil
}

func (c *TerminalMenu) onCheatAddScore(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		amount int
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.IntVar(&args.amount, "amount", 0, "the amount of score to add")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	c.state.Persistent.PlayerStats.TotalScore += args.amount
	contentlock.Update(c.state)
	return fmt.Sprintf("Added %d to the total score.", args.amount), nil
}

func (c *TerminalMenu) onDebugDroneLabels(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		enable bool
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.BoolVar(&args.enable, "enable", false, "whether to enable the debug drone labels")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	oldValue := c.state.Persistent.Settings.DebugLogs
	c.state.Persistent.Settings.DebugDroneLabels = args.enable
	return fmt.Sprintf("Set debug.drone_labels to %v (was %v)", args.enable, oldValue), nil
}

func (c *TerminalMenu) onDebugLogs(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		enable bool
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.BoolVar(&args.enable, "enable", false, "whether to enable the debug logs")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	oldValue := c.state.Persistent.Settings.DebugLogs
	c.state.Persistent.Settings.DebugLogs = args.enable
	return fmt.Sprintf("Set debug.logs to %v (was %v)", args.enable, oldValue), nil
}

func (c *TerminalMenu) onReplayDump(ctx *terminalCommandContext) (string, error) {
	type argsType struct {
		file string
	}
	if ctx.parsedArgs == nil {
		args := &argsType{}
		ctx.parsedArgs = args
		ctx.fs.StringVar(&args.file, "file", "classic_highscore", "which replay file to dump")
		return "", nil
	}
	args := ctx.parsedArgs.(*argsType)
	var replayKey string
	switch args.file {
	case "classic_highscore", "arena_highscore", "inf_arena_highscore":
		replayKey = args.file
	default:
		return "", fmt.Errorf("unknown replay file %q", args.file)
	}
	var replayData serverapi.GameReplay
	if err := c.scene.Context().LoadGameData(replayKey, &replayData); err != nil {
		return "", err
	}
	jsonData, err := json.Marshal(replayData)
	if err != nil {
		return "", err
	}
	println(string(jsonData))
	return fmt.Sprintf("%q replay data is dumped to the console", replayKey), nil
}
