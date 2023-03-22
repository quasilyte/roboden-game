package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type rpanelNode struct {
	scene *ge.Scene

	world *worldState

	layerSprite1 *ge.Sprite
	layerSprite2 *ge.Sprite

	colony *colonyCoreNode

	factionRects []*ge.Rect
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

	colors := [...]color.RGBA{
		gamedata.FactionByTag(gamedata.YellowFactionTag).Color,
		gamedata.FactionByTag(gamedata.RedFactionTag).Color,
		gamedata.FactionByTag(gamedata.GreenFactionTag).Color,
		gamedata.FactionByTag(gamedata.BlueFactionTag).Color,
	}
	for _, clr := range colors {
		rect := ge.NewRect(scene.Context(), 5, 0)
		rect.Centered = false
		rect.Pos.Offset = gmath.Vec{X: 952}
		rect.FillColorScale.SetColor(clr)
		scene.AddGraphicsAbove(rect, 2)
		panel.factionRects = append(panel.factionRects, rect)
	}
}

func (panel *rpanelNode) SetBase(colony *colonyCoreNode) {
	panel.colony = colony

	if panel.colony == nil {
		for _, rect := range panel.factionRects {
			rect.Visible = false
		}
	}
}

func (panel *rpanelNode) UpdateMetrics() {
	if panel.colony == nil {
		return
	}

	// Update factions distribution rects.
	topOffset := 8.0
	totalHeight := 344.0
	height := topOffset
	for i, kv := range panel.colony.factionWeights.Elems {
		factionHeight := kv.Weight * totalHeight
		if kv.Key != gamedata.NeutralFactionTag {
			rect := panel.factionRects[i-1]
			rect.Height = factionHeight
			rect.Pos.Offset.Y = height
		}
		height += factionHeight
	}
}

func (panel *rpanelNode) Update(delta float64) {}
