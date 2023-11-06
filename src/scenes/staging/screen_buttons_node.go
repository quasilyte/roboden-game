package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

type screenButtonKind int

const (
	screenButtonUnknown screenButtonKind = iota
	screenButtonExit
	screenButtonToggle
	screenButtonFastForward
)

type screenButtonsNode struct {
	toggleButtonRect      gmath.Rect
	exitButtonRect        gmath.Rect
	fastForwardButtonRect gmath.Rect

	cam *viewport.Camera
	pos gmath.Vec

	dark   bool
	scaled bool

	EventToggleButtonPressed      gsignal.Event[gsignal.Void]
	EventExitButtonPressed        gsignal.Event[gsignal.Void]
	EventFastForwardButtonPressed gsignal.Event[gsignal.Void]
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
	if n.scaled {
		buttonSize = gmath.Vec{X: 66, Y: 66}
	}

	if n.scaled {
		n.pos = n.pos.Sub(gmath.Vec{Y: 32})
	}

	toggleButtonOffset := n.pos.Add(gmath.Vec{X: 12, Y: 24})
	fastForwardButtonOffset := n.pos.Add(gmath.Vec{X: 68, Y: 24})
	exitButtonOffset := n.pos.Add(gmath.Vec{X: 124, Y: 24})
	if n.scaled {
		fastForwardButtonOffset = n.pos.Add(gmath.Vec{X: 112, Y: 24})
		exitButtonOffset = n.pos.Add(gmath.Vec{X: 212, Y: 24})
	}

	n.toggleButtonRect = gmath.Rect{Min: toggleButtonOffset, Max: toggleButtonOffset.Add(buttonSize)}
	n.exitButtonRect = gmath.Rect{Min: exitButtonOffset, Max: exitButtonOffset.Add(buttonSize)}
	n.fastForwardButtonRect = gmath.Rect{Min: fastForwardButtonOffset, Max: fastForwardButtonOffset.Add(buttonSize)}

	var img resource.ImageID
	switch {
	case !n.dark && n.scaled:
		img = assets.ImageRadarlessButtonsX2
	case n.dark && n.scaled:
		img = assets.ImageDarkRadarlessButtonsX2
	case n.dark:
		img = assets.ImageDarkRadarlessButtons
	default:
		img = assets.ImageRadarlessButtons
	}
	sprite := scene.NewSprite(img)
	sprite.Pos.Base = &n.pos
	sprite.Centered = false
	n.cam.UI.AddGraphics(sprite)
}

func (n *screenButtonsNode) GetChoiceUnderCursor(pos gmath.Vec) screenButtonKind {
	if n.exitButtonRect.Contains(pos) {
		return screenButtonExit
	}
	if n.toggleButtonRect.Contains(pos) {
		return screenButtonToggle
	}
	if n.fastForwardButtonRect.Contains(pos) {
		return screenButtonFastForward
	}
	return screenButtonUnknown
}

func (n *screenButtonsNode) HandleInput(clickPos gmath.Vec) bool {
	if n.exitButtonRect.Contains(clickPos) {
		n.EventExitButtonPressed.Emit(gsignal.Void{})
		return true
	}
	if n.toggleButtonRect.Contains(clickPos) {
		n.EventToggleButtonPressed.Emit(gsignal.Void{})
		return true
	}
	if n.fastForwardButtonRect.Contains(clickPos) {
		n.EventFastForwardButtonPressed.Emit(gsignal.Void{})
		return true
	}
	return false
}
