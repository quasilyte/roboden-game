package menus

import (
	"fmt"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/gameui/eui"
)

func bindNavListNoWrap(elems []*gameui.NavElem, prevDir, nextDir gameui.NavDir) {
	for i, e := range elems {
		switch i {
		case 0:
			e.Edges[nextDir] = elems[i+1]
		case len(elems) - 1:
			e.Edges[prevDir] = elems[i-1]
		default:
			e.Edges[nextDir] = elems[i+1]
			e.Edges[prevDir] = elems[i-1]
		}
	}
}

func bindNavList(elems []*gameui.NavElem, prevDir, nextDir gameui.NavDir) {
	for i, e := range elems {
		switch i {
		case 0:
			e.Edges[nextDir] = elems[i+1]
			e.Edges[prevDir] = elems[len(elems)-1]
		case len(elems) - 1:
			e.Edges[nextDir] = elems[0]
			e.Edges[prevDir] = elems[i-1]
		default:
			e.Edges[nextDir] = elems[i+1]
			e.Edges[prevDir] = elems[i-1]
		}
	}
}

func bindNavGrid(elems []*gameui.NavElem, cols, rows int) {
	safeElemGet := func(i int) *gameui.NavElem {
		if i < len(elems) && i >= 0 {
			return elems[i]
		}
		return nil
	}
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			i := row*cols + col
			if i >= len(elems) {
				return
			}
			elem := elems[i]
			if row < rows-1 {
				elem.Edges[gameui.NavDown] = safeElemGet(i + cols)
			}
			if row > 0 {
				elem.Edges[gameui.NavUp] = safeElemGet(i - cols)
			}
			if col < cols-1 {
				elem.Edges[gameui.NavRight] = safeElemGet(i + 1)
			}
			if col > 0 {
				elem.Edges[gameui.NavLeft] = safeElemGet(i - 1)
			}
		}
	}
}

func createSimpleNavTree(widgets []eui.Widget) *gameui.NavTree {
	navTree := gameui.NewNavTree()
	navBlock := navTree.NewBlock()
	elems := make([]*gameui.NavElem, len(widgets))
	for i, w := range widgets {
		elems[i] = navBlock.NewElem(w)
	}
	if len(widgets) > 1 {
		bindNavList(elems, gameui.NavUp, gameui.NavDown)
	}
	return navTree
}

func setupUI(scene *ge.Scene, root *widget.Container, h *gameinput.Handler, navTree *gameui.NavTree) *navController {
	uiObject := eui.NewSceneObject(root)

	var controller *navController
	if navTree != nil {
		controller = &navController{
			input:   h,
			navTree: navTree,
			ui:      uiObject,
		}
		scene.AddObject(controller)
	}

	scene.AddGraphics(uiObject)
	scene.AddObject(uiObject)
	return controller
}

type navController struct {
	input       *gameinput.Handler
	navTree     *gameui.NavTree
	ui          *eui.SceneObject
	lastFocused *gameui.NavElem
}

func (c *navController) FocusBlock(b *gameui.NavBlock) {
	c.lastFocused = b.GetFirstElem()
	if c.lastFocused != nil {
		c.focusElement(c.lastFocused)
	}
}

func (c *navController) Unfocus() {
	c.lastFocused = nil
	c.ui.Unfocus()
}

func (c *navController) IsDisposed() bool { return false }

func (c *navController) Init(scene *ge.Scene) {}

func (c *navController) Update(delta float64) {
	if c.input.ActionIsJustPressed(controls.ActionMenuConfirm) {
		e := c.getFocused()
		if e == nil {
			return
		}
		switch e := e.Widget.(type) {
		case *widget.Button:
			e.Submit()
		default:
			fmt.Printf("unhandled press event: %T\n", e)
		}
	}

	if c.input.ActionIsJustPressed(controls.ActionMenuFocusRight) {
		c.tryFocusing(gameui.NavRight)
	}
	if c.input.ActionIsJustPressed(controls.ActionMenuFocusDown) {
		c.tryFocusing(gameui.NavDown)
	}
	if c.input.ActionIsJustPressed(controls.ActionMenuFocusLeft) {
		c.tryFocusing(gameui.NavLeft)
	}
	if c.input.ActionIsJustPressed(controls.ActionMenuFocusUp) {
		c.tryFocusing(gameui.NavUp)
	}
}

func (c *navController) tryFocusing(d gameui.NavDir) {
	e := c.getFocused()
	if e == nil {
		c.focusElement(c.navTree.GetFirstElem())
		return
	}

	current := e
	var next *gameui.NavElem
	for {
		next = current.Find(d)
		if next == nil {
			return
		}
		if next.Widget.GetWidget().Disabled {
			current = next
			continue
		}
		break
	}

	if next != nil {
		c.focusElement(next)
	}
}

func (c *navController) focusElement(e *gameui.NavElem) {
	if e == nil {
		return
	}
	if focuser, ok := e.Widget.(widget.Focuser); ok {
		focuser.Focus(true)
		c.lastFocused = e

		switch e := e.Widget.(type) {
		case *widget.Button:
			e.CursorEnteredEvent.Fire(&widget.ButtonHoverEventArgs{
				Button: e,
			})
		case *widget.TextInput:
			// TODO: maybe open a screen keyboard?
		default:
			fmt.Printf("unhandled focus event: %T\n", e)
		}
	}
}

func (c *navController) getFocused() *gameui.NavElem {
	if focused := c.ui.GetFocused(); focused != nil {
		return c.navTree.FindElem(focused)
	}
	if c.lastFocused != nil {
		return c.lastFocused
	}
	return nil
}
