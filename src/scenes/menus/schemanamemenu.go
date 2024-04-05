package menus

import (
	"runtime"

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

// TODO: refarctor all text input-related code.
// There is a big code duplication on every screen that
// expects some text input from the user.
// These screens include: this one, username, lobby (seed), terminal (command input).

type SchemaNameMenu struct {
	state *session.State

	errorSoundDelay float64

	selectedSlot int
	mode         gamedata.Mode

	ui        *eui.SceneObject
	keyboard  *eui.Keyboard
	textInput *widget.TextInput

	scene *ge.Scene
}

func NewSchemaNameMenuController(state *session.State, mode gamedata.Mode, selectedSlot int) *SchemaNameMenu {
	return &SchemaNameMenu{
		state:        state,
		selectedSlot: selectedSlot,
		mode:         mode,
	}
}

func (c *SchemaNameMenu) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()
}

func (c *SchemaNameMenu) Update(delta float64) {
	c.errorSoundDelay = gmath.ClampMin(c.errorSoundDelay-delta, 0)
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *SchemaNameMenu) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	var widgets []eui.Widget

	key := c.state.SchemaDataKey(c.mode, c.selectedSlot)
	if !c.state.CheckGameItem(key) {
		c.scene.Audio().PlaySound(assets.AudioError)
		c.scene.Context().ChangeScene(NewSchemaMenuController(c.state, c.mode))
		return
	}

	var schema gamedata.SavedSchema
	if err := c.state.LoadGameItem(key, &schema); err != nil {
		c.scene.Audio().PlaySound(assets.AudioError)
		c.scene.Context().ChangeScene(NewSchemaMenuController(c.state, c.mode))
		return
	}

	titleLabel := eui.NewCenteredLabel(d.Get("menu.schema_name"), c.state.Resources.Font3)
	rowContainer.AddChild(titleLabel)

	textinput := eui.NewTextInput(uiResources, eui.TextInputConfig{SteamDeck: c.state.Device.IsSteamDeck()},
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(480, 0),
		),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			if args.InputText == "" {
				return
			}
			if !gamedata.IsValidSchemaName(args.InputText) {
				c.scene.Audio().PlaySound(assets.AudioError)
				return
			}
		}),
		widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
			if len(newInputText) == 0 {
				return true, nil
			}
			good := len(newInputText) <= serverapi.MaxNameLength && gamedata.IsValidSchemaName(newInputText)
			if !good && c.errorSoundDelay == 0 {
				c.scene.Audio().PlaySound(assets.AudioError)
				c.errorSoundDelay = 0.2
			}
			return good, nil
		}),
	)
	if schema.Name != "" {
		textinput.SetText(schema.Name)
	}
	rowContainer.AddChild(textinput)
	widgets = append(widgets, textinput)

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

	saveButton := eui.NewButton(uiResources, c.scene, d.Get("menu.save"), func() {
		schema.Name = textinput.GetText()
		c.state.SaveGameItem(key, schema)
		c.scene.Context().ChangeScene(NewSchemaMenuController(c.state, c.mode))
	})
	rowContainer.AddChild(saveButton)
	widgets = append(widgets, saveButton)

	navTree := createSimpleNavTree(widgets)
	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *SchemaNameMenu) back() {
	c.scene.Audio().PlaySound(assets.AudioError)
}

func (c *SchemaNameMenu) openKeyboard() {
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
