package menus

import (
	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

type UserNameMenu struct {
	state *session.State

	errorSoundDelay float64

	nextController ge.SceneController

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
	if c.state.CombinedInput.ActionIsJustPressed(controls.ActionBack) {
		c.back()
		return
	}
}

func (c *UserNameMenu) isValidChar(ch byte) bool {
	isLetter := func(ch byte) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	}
	isDigit := func(ch byte) bool {
		return ch >= '0' && ch <= '9'
	}
	return isLetter(ch) || isDigit(ch) || ch == ' '
}

func (c *UserNameMenu) isValidUsername(s string) bool {
	nonSpace := 0
	if len(s) > serverapi.MaxNameLength {
		return false
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		isValid := c.isValidChar(ch)
		if !isValid {
			return false
		}
		if ch != ' ' {
			nonSpace++
		}
	}
	return nonSpace != 0
}

func (c *UserNameMenu) initUI() {
	addDemoBackground(c.state, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	smallFont := assets.BitmapFont1

	titleLabel := eui.NewCenteredLabel(d.Get("menu.user_name"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	textinput := eui.NewTextInput(uiResources, assets.BitmapFont2,
		widget.TextInputOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(480, 0),
		),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			if args.InputText == "" {
				return
			}
			if !c.isValidUsername(args.InputText) {
				c.scene.Audio().PlaySound(assets.AudioError)
				return
			}
		}),
		widget.TextInputOpts.Validation(func(newInputText string) (bool, *string) {
			good := len(newInputText) <= serverapi.MaxNameLength && c.isValidUsername(newInputText)
			if !good && c.errorSoundDelay == 0 {
				c.scene.Audio().PlaySound(assets.AudioError)
				c.errorSoundDelay = 0.2
			}
			return good, nil
		}),
	)
	textinput.SetText(c.state.Persistent.PlayerName)
	rowContainer.AddChild(textinput)

	normalContainer := eui.NewAnchorContainer()
	rulesLabel := eui.NewLabel(d.Get("menu.user_name_rules"), smallFont)
	normalContainer.AddChild(rulesLabel)
	rowContainer.AddChild(rulesLabel)

	rowContainer.AddChild(eui.NewSeparator(widget.RowLayoutData{Stretch: true}))

	rowContainer.AddChild(eui.NewButton(uiResources, c.scene, d.Get("menu.save"), func() {
		c.save(textinput.GetText())
		c.next()
	}))

	uiObject := eui.NewSceneObject(root)
	c.scene.AddGraphics(uiObject)
	c.scene.AddObject(uiObject)
}

func (c *UserNameMenu) save(name string) {
	if c.isValidUsername(name) || name == "" {
		c.state.Persistent.PlayerName = name
		c.scene.Context().SaveGameData("save", c.state.Persistent)
	}
}

func (c *UserNameMenu) next() {
	c.scene.Context().ChangeScene(c.nextController)
}

func (c *UserNameMenu) back() {
	c.scene.Audio().PlaySound(assets.AudioError)
}
