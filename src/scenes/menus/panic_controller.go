package menus

import (
	"fmt"
	"strings"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/session"
)

type PanicController struct {
	panicInfo *ge.PanicInfo

	scene *ge.Scene

	state *session.State
}

type recoverableController interface {
	GetSessionState() *session.State
}

func NewPanicController(panicInfo *ge.PanicInfo) *PanicController {
	return &PanicController{
		panicInfo: panicInfo,
	}
}

func (c *PanicController) Init(scene *ge.Scene) {
	c.scene = scene

	traceLines := strings.Split(c.panicInfo.Trace, "\n")
	if len(traceLines) > 28 {
		traceLines = traceLines[:28]
	}
	trimmedTrace := strings.Join(traceLines, "\n")

	if rc, ok := c.panicInfo.Controller.(recoverableController); ok {
		c.state = rc.GetSessionState()
	}

	text := "A critical error has occured.\nThe game can't recover from this error."
	if c.state != nil {
		text = "A critical error has occured.\nPress ENTER to continue."
	}

	fmt.Println(c.panicInfo.Trace)

	errorLabel := ge.NewLabel(assets.BitmapFont1)
	errorLabel.Width = scene.Context().WindowWidth
	errorLabel.Height = scene.Context().WindowHeight
	errorLabel.GrowVertical = ge.GrowVerticalDown
	errorLabel.Pos.Offset = gmath.Vec{X: 16, Y: 16}
	errorLabel.Text = text + "\n\n" + fmt.Sprint(c.panicInfo.Value) + "\n" + trimmedTrace
	errorLabel.ColorScale.SetRGBA(0x9d, 0xd7, 0x93, 0xff)
	scene.AddGraphics(errorLabel)
}

func (c *PanicController) Update(delta float64) {
	if c.state != nil {
		if c.state.MainInput.ActionIsJustPressed(controls.ActionSkipDemo) {
			c.scene.Context().ChangeScene(NewMainMenuController(c.state))
			return
		}
	}
}
