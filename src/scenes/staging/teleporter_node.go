package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
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
	t.world.stage.AddSpriteBelow(s)

	lights := scene.NewSprite(assets.ImageTeleporterLights)
	lights.Pos.Offset = t.pos
	switch t.id {
	case 1:
		cs := ge.ColorScale{R: 0.5, G: 1.6, B: 0.6, A: 1}
		lights.SetColorScale(cs)
	}
	t.world.stage.AddSpriteBelow(lights)
}

func (t *teleporterNode) IsDisposed() bool { return false }

func (t *teleporterNode) Update(delta float64) {}

func (t *teleporterNode) CanBeUsedBy(user *colonyCoreNode) bool {
	if t.world.coreDesign == gamedata.TankCoreStats {
		return true
	}
	target := t.other
	for _, c := range t.world.allColonies {
		if c != user {
			if c.pos.DistanceSquaredTo(t.pos) < (40 * 40) {
				return false
			}
		}
		if target.pos.DistanceSquaredTo(c.pos) <= (40 * 40) {
			// There is a colony on the other side that blocks the teleporter.
			return false
		}
	}
	return true
}
