package menus

import (
	"runtime"
	"strings"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

type UserNameMenu struct {
	state *session.State

	errorSoundDelay float64

	nextController ge.SceneController

	ui        *eui.SceneObject
	keyboard  *eui.Keyboard
	textInput *widget.TextInput

	scene *ge.Scene
}

func NewUserNameMenuController(state *session.State, next ge.SceneController) *UserNameMenu {
	return &UserNameMenu{
		state:          state,
		nextController: next,
	}
}

func (c *UserNameMenu) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *UserNameMenu) Update(delta float64) {
	c.errorSoundDelay = gmath.ClampMin(c.errorSoundDelay-delta, 0)
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *UserNameMenu) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	smallFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.user_name"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	textinput := eui.NewTextInput(uiResources, eui.TextInputConfig{SteamDeck: c.state.Device.IsSteamDeck()},
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(480, 0),
		),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			if args.InputText == "" {
				return
			}
			if !gamedata.IsValidUsername(args.InputText) {
				c.scene.Audio().PlaySound(assets.AudioError)
				return
			}
		}),
		widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
			good := len(newInputText) <= serverapi.MaxNameLength && gamedata.IsValidUsername(newInputText)
			if !good && c.errorSoundDelay == 0 {
				c.scene.Audio().PlaySound(assets.AudioError)
				c.errorSoundDelay = 0.2
			}
			return good, nil
		}),
	)
	textinput.SetText(c.state.Persistent.PlayerName)
	rowContainer.AddChild(textinput)

	c.textInput = textinput
	if runtime.GOOS == "android" {
		c.textInput.GetWidget().FocusEvent.AddHandler(func(args any) {
			e := args.(*widget.WidgetFocusEventArgs)
			if e.Focused {
				if c.keyboard == nil {
					c.openKeyboard()
				}
			}
		})
	}

	panel := eui.NewTextPanel(uiResources, 0, 0)

	normalContainer := eui.NewAnchorContainer()
	rulesLabel := eui.NewLabel(d.Get("menu.user_name_rules"), smallFont)
	normalContainer.AddChild(rulesLabel)
	panel.AddChild(normalContainer)
	rowContainer.AddChild(panel)

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.save"), func() {
		c.save(textinput.GetText())
		c.next()
	}))

	c.ui = eui.NewSceneObject(root)
	c.scene.AddGraphics(c.ui)
	c.scene.AddObject(c.ui)
}

func (c *UserNameMenu) save(name string) {
	name = strings.TrimSpace(name)
	if gamedata.IsValidUsername(name) || name == "" {
		c.state.Persistent.PlayerName = name
		c.state.SaveGameItem("save.json", c.state.Persistent)
	}
}

func (c *UserNameMenu) next() {
	c.scene.Context().ChangeScene(c.nextController)
}

func (c *UserNameMenu) back() {
	c.scene.Audio().PlaySound(assets.AudioError)
}

func (c *UserNameMenu) openKeyboard() {
	k := eui.NewTextKeyboard(eui.KeyboardConfig{
		Resources: c.state.Resources.UI,
		Scene:     c.scene,
		Input:     c.state.MenuInput,
	})
	c.ui.AddWindow(k.Window)

	runeBuf := []rune{0}
	k.EventKey.Connect(nil, func(ch rune) {
		runeBuf[0] = ch
		c.textInput.Insert(runeBuf)
		c.textInput.Focus(true)
	})
	k.EventBackspace.Connect(nil, func(gsignal.Void) {
		c.textInput.Backspace()
		c.textInput.Focus(true)
	})
	k.EventSubmit.Connect(nil, func(gsignal.Void) {
		c.textInput.Submit()
		k.Close()
	})
	k.EventClosed.Connect(nil, func(gsignal.Void) {
		c.keyboard = nil
	})
	c.keyboard = k
	c.scene.AddObject(c.keyboard)
}
