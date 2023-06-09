package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

type screenButtonsNode struct {
	toggleButtonRect gmath.Rect
	exitButtonRect   gmath.Rect

	cam *viewport.Camera
	pos gmath.Vec

	dark bool

	EventToggleButtonPressed gsignal.Event[gsignal.Void]
	EventExitButtonPressed   gsignal.Event[gsignal.Void]
}

func newScreenButtonsNode(cam *viewport.Camera, pos gmath.Vec, dark bool) *screenButtonsNode {
	return &screenButtonsNode{
		pos:  pos,
		cam:  cam,
		dark: dark,
	}
}

func (n *screenButtonsNode) Init(scene *ge.Scene) {
	buttonSize := gmath.Vec{X: 34, Y: 34}

	toggleButtonOffset := n.pos.Add(gmath.Vec{X: 12, Y: 24})
	n.toggleButtonRect = gmath.Rect{Min: toggleButtonOffset, Max: toggleButtonOffset.Add(buttonSize)}

	exitButtonOffset := n.pos.Add(gmath.Vec{X: 68, Y: 28})
	n.exitButtonRect = gmath.Rect{Min: exitButtonOffset, Max: exitButtonOffset.Add(buttonSize)}

	img := assets.ImageRadarlessButtons
	if n.dark {
		img = assets.ImageDarkRadarlessButtons
	}
	sprite := scene.NewSprite(img)
	sprite.Pos.Base = &n.pos
	sprite.Centered = false
	n.cam.UI.AddGraphics(sprite)
}

func (n *screenButtonsNode) HandleInput(clickPos gmath.Vec) {
	if n.exitButtonRect.Contains(clickPos) {
		n.EventExitButtonPressed.Emit(gsignal.Void{})
		return
	}
	if n.toggleButtonRect.Contains(clickPos) {
		n.EventToggleButtonPressed.Emit(gsignal.Void{})
		return
	}
}
