package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
)

type rpanelNode struct {
	scene *ge.Scene

	world *worldState

	layerSprite1 *ge.Sprite
	layerSprite2 *ge.Sprite
}

func newRpanelNode(world *worldState) *rpanelNode {
	return &rpanelNode{world: world}
}

func (panel *rpanelNode) IsDisposed() bool { return false }

func (panel *rpanelNode) Init(scene *ge.Scene) {
	panel.scene = scene

	panel.layerSprite1 = scene.NewSprite(assets.ImageRightPanelLayer1)
	panel.layerSprite1.Pos.Offset.X = 782
	panel.layerSprite1.Centered = false
	scene.AddGraphicsAbove(panel.layerSprite1, 1)

	panel.layerSprite2 = scene.NewSprite(assets.ImageRightPanelLayer2)
	panel.layerSprite2.Pos = panel.layerSprite1.Pos
	panel.layerSprite2.Centered = false
	scene.AddGraphicsAbove(panel.layerSprite2, 2)
}

func (panel *rpanelNode) Update(delta float64) {}
