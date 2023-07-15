package menus

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/session"
)

func addDemoBackground(state *session.State, scene *ge.Scene) {
	if state.DemoFrame == nil {
		return
	}
	s := ge.NewSprite(scene.Context())
	s.Centered = false
	s.SetColorScale(ge.ColorScale{R: 0.35, G: 0.35, B: 0.35, A: 1})
	s.SetImage(resource.Image{Data: state.DemoFrame})
	scene.AddGraphics(s)
}

func reverseStrings(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}
