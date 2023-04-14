package menus

import (
	"bytes"
	"flag"
	"fmt"
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
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
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
	if c.state.MainInput.ActionIsJustPressed(controls.ActionBack) {
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
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	normalFont := c.scene.Context().Loader.LoadFont(assets.FontNormal).Face
	tinyFont := c.scene.Context().Loader.LoadFont(assets.FontTiny).Face

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.title")+" -> "+d.Get("menu.main.settings")+" -> "+d.Get("menu.options.extra")+" -> "+d.Get("menu.terminal"), normalFont)
	rowContainer.AddChild(titleLabel)

	outputPanel := eui.NewPanel(uiResources, 520, 200)

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
			key:     "debug.logs",
			handler: c.onDebugLogs,
		},
	}

	textinput := eui.NewTextInput(uiResources, normalFont,
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

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *TerminalMenu) back() {
	c.scene.Context().SaveGameData("save", c.state.Persistent)
	c.scene.Context().ChangeScene(NewOptionsExtraMenuController(c.state))
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
	return "The save data is cleared.", nil
}

func (c *TerminalMenu) onSaveInfo(*terminalCommandContext) (string, error) {
	lines := []string{
		fmt.Sprintf("Save file: %q", c.scene.Context().LocateGameData("save")),
	}
	return strings.Join(lines, "\n"), nil
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
