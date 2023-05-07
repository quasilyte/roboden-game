package staging

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
)

type uiLayer struct {
	scene *ge.Scene

	Visible bool

	components      []ge.SceneGraphics
	aboveComponents []ge.SceneGraphics
}

func newUILayer() *uiLayer {
	return &uiLayer{Visible: true}
}

func (l *uiLayer) Init(scene *ge.Scene) {
	l.scene = scene
}

func (l *uiLayer) AddGraphics(g ge.SceneGraphics) {
	l.components = append(l.components, g)
}

func (l *uiLayer) AddGraphicsAbove(g ge.SceneGraphics) {
	l.aboveComponents = append(l.aboveComponents, g)
}

func (l *uiLayer) IsDisposed() bool {
	return false
}

func (l *uiLayer) Draw(screen *ebiten.Image) {
	if !l.Visible {
		return
	}

	{
		live := l.components[:0]
		for _, c := range l.components {
			if c.IsDisposed() {
				continue
			}
			c.Draw(screen)
			live = append(live, c)
		}
		l.components = live
	}

	{
		live := l.aboveComponents[:0]
		for _, c := range l.aboveComponents {
			if c.IsDisposed() {
				continue
			}
			c.Draw(screen)
			live = append(live, c)
		}
		l.aboveComponents = live
	}
}
