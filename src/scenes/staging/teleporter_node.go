package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

var teleportOffset = gmath.Vec{Y: -8}

type teleporterNode struct {
	id    int
	world *worldState
	pos   gmath.Vec
	other *teleporterNode
}

func (t *teleporterNode) Init(scene *ge.Scene) {
	s := scene.NewSprite(assets.ImageTeleporter)
	s.Pos.Offset = t.pos
	t.world.camera.AddSpriteBelow(s)

	lights := scene.NewSprite(assets.ImageTeleporterLights)
	lights.Pos.Offset = t.pos
	switch t.id {
	case 1:
		lights.SetHue(-math.Pi)
	}
	t.world.camera.AddSpriteBelow(lights)
}

func (t *teleporterNode) IsDisposed() bool { return false }

func (t *teleporterNode) Update(delta float64) {}
