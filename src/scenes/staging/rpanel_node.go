package staging

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/viewport"
)

func setPriorityIconFrame(s *ge.Sprite, priority colonyPriority, faction gamedata.FactionTag) {
	offsetX := float64(priority) * 16.0
	offsetX += float64(faction) * (16.0 * 4)
	s.FrameOffset.X = offsetX
}

type rpanelNode struct {
	scene *ge.Scene

	cam *viewport.Camera

	layerSprite1 *ge.Sprite
	layerSprite2 *ge.Sprite

	colony *colonyCoreNode

	factionRects  []*ge.Rect
	priorityBars  []*ge.Sprite
	priorityIcons []*ge.Sprite
}

func newRpanelNode(cam *viewport.Camera) *rpanelNode {
	return &rpanelNode{
		cam: cam,
	}
}

func (panel *rpanelNode) IsDisposed() bool { return false }

func (panel *rpanelNode) Init(scene *ge.Scene) {
	panel.scene = scene

	cameraWidth := panel.cam.Rect.Width()

	panel.layerSprite1 = scene.NewSprite(assets.ImageRightPanelLayer1)
	panel.layerSprite1.Pos.Offset.X = (cameraWidth - panel.layerSprite1.FrameWidth)
	panel.layerSprite1.Centered = false
	panel.cam.UI.AddGraphicsBelow(panel.layerSprite1)

	priorities := []colonyPriority{
		priorityResources,
		priorityGrowth,
		priorityEvolution,
		prioritySecurity,
	}
	for i, priority := range priorities {
		bar := scene.NewSprite(assets.ImagePriorityBar)
		bar.Pos.Offset = gmath.Vec{X: (cameraWidth - (panel.layerSprite1.FrameWidth - 16)) + ((18 + bar.FrameWidth) * float64(i))}
		bar.Centered = false
		panel.cam.UI.AddGraphics(bar)

		icon := scene.NewSprite(assets.ImagePriorityIcons)
		setPriorityIconFrame(icon, priority, gamedata.NeutralFactionTag)
		icon.Pos.Offset = gmath.Vec{X: (cameraWidth - (panel.layerSprite1.FrameWidth - 16)) + ((18 + bar.FrameWidth) * float64(i))}
		icon.Centered = false
		panel.cam.UI.AddGraphicsAbove(icon)

		panel.priorityBars = append(panel.priorityBars, bar)
		panel.priorityIcons = append(panel.priorityIcons, icon)
	}

	panel.layerSprite2 = scene.NewSprite(assets.ImageRightPanelLayer2)
	panel.layerSprite2.Pos = panel.layerSprite1.Pos
	panel.layerSprite2.Centered = false
	panel.cam.UI.AddGraphicsAbove(panel.layerSprite2)

	colors := [...]color.RGBA{
		gamedata.FactionByTag(gamedata.YellowFactionTag).Color,
		gamedata.FactionByTag(gamedata.RedFactionTag).Color,
		gamedata.FactionByTag(gamedata.GreenFactionTag).Color,
		gamedata.FactionByTag(gamedata.BlueFactionTag).Color,
	}
	for _, clr := range colors {
		rect := ge.NewRect(scene.Context(), 5, 0)
		rect.Centered = false
		rect.Pos.Offset = gmath.Vec{X: (cameraWidth - 8)}
		rect.FillColorScale.SetColor(clr)
		panel.cam.UI.AddGraphicsAbove(rect)
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

	fullPriorityOffset := 445.0
	for i, kv := range panel.colony.priorities.Elems {
		bar := panel.priorityBars[i]
		bar.Pos.Offset.Y = fullPriorityOffset + ((bar.FrameHeight - 8) * (1.0 - kv.Weight))
		icon := panel.priorityIcons[i]
		icon.Pos.Offset.Y = fullPriorityOffset + ((bar.FrameHeight - 8) * (1.0 - kv.Weight)) - icon.FrameHeight - 1
	}
}

func (panel *rpanelNode) Update(delta float64) {}
