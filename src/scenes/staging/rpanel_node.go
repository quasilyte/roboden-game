package staging

import (
	"image/color"
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/viewport"
)

type rpanelItemKind int

const (
	rpanelItemUnknown rpanelItemKind = iota
	rpanelItemResourcesPriority
	rpanelItemGrowthPriority
	rpanelItemEvolutionPriority
	rpanelItemSecurityPriority
	rpanelItemFactionDistribution
	rpanelItemTechProgress
	rpanelItemGarrison
)

var (
	dpadBarColorBright = ge.RGB(0x48d35d)
	dpadBarColorNormal = ge.RGB(0x3ea24d)
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

	// 4 rects for colonies.
	// 1 rect for creeps.
	factionRects []*ge.Rect

	// For colonies.
	colony        *colonyCoreNode
	priorityIcons []*ge.Sprite
	priorityBars  []*ge.Sprite

	// For creeps.
	creepsState *creepsPlayerState
	dpad        *ge.Sprite
	dpadRects   []*ge.Rect
}

func newCreepsRpanelNode(cam *viewport.Camera, creepsState *creepsPlayerState) *rpanelNode {
	return &rpanelNode{
		cam:         cam,
		creepsState: creepsState,
	}
}

func newRpanelNode(cam *viewport.Camera) *rpanelNode {
	return &rpanelNode{
		cam: cam,
	}
}

func (panel *rpanelNode) IsDisposed() bool { return false }

func (panel *rpanelNode) GetItemUnderCursor(pos gmath.Vec) (rpanelItemKind, float64) {
	factionRect := panel.factionRects[0].BoundsRect()
	factionRect.Min.X -= 4
	factionRect.Max.X += 4
	factionRect.Min.Y = 10
	factionRect.Max.Y = 350

	if panel.creepsState != nil {
		if panel.dpad.BoundsRect().Contains(pos) {
			return rpanelItemGarrison, 0
		}
		if factionRect.Contains(pos) {
			return rpanelItemTechProgress, panel.creepsState.techLevel
		}
		return rpanelItemUnknown, 0
	}

	for i, priorityIcon := range panel.priorityIcons {
		rect := priorityIcon.BoundsRect()
		rect.Min = rect.Min.Sub(gmath.Vec{X: 4, Y: 4})
		rect.Max = rect.Max.Add(gmath.Vec{X: 4, Y: 4})
		rect.Max.Y += 20
		if rect.Contains(pos) {
			v := panel.colony.priorities.Elems[i].Weight
			return rpanelItemResourcesPriority + rpanelItemKind(i), v
		}
	}

	if factionRect.Contains(pos) {
		return rpanelItemFactionDistribution, 0
	}

	return rpanelItemUnknown, 0
}

func (panel *rpanelNode) initFactionsForColonies() {
	cameraWidth := panel.cam.Rect.Width()
	colors := [...]color.RGBA{
		gamedata.FactionByTag(gamedata.YellowFactionTag).Color,
		gamedata.FactionByTag(gamedata.RedFactionTag).Color,
		gamedata.FactionByTag(gamedata.GreenFactionTag).Color,
		gamedata.FactionByTag(gamedata.BlueFactionTag).Color,
	}
	for _, clr := range colors {
		rect := ge.NewRect(panel.scene.Context(), 5, 0)
		rect.Centered = false
		rect.Pos.Offset = gmath.Vec{X: (cameraWidth - 8)}
		rect.FillColorScale.SetColor(clr)
		panel.cam.UI.AddGraphicsAbove(rect)
		panel.factionRects = append(panel.factionRects, rect)
	}
}

func (panel *rpanelNode) initFactionsForCreeps() {
	cameraWidth := panel.cam.Rect.Width()

	rect := ge.NewRect(panel.scene.Context(), 5, 0)
	rect.Centered = false
	rect.Pos.Offset = gmath.Vec{X: (cameraWidth - 8)}
	rect.FillColorScale.SetColor(dpadBarColorNormal)
	panel.cam.UI.AddGraphicsAbove(rect)
	panel.factionRects = append(panel.factionRects, rect)

	rect2 := ge.NewRect(panel.scene.Context(), 5, 0)
	rect2.Centered = false
	rect2.Pos.Offset = gmath.Vec{X: (cameraWidth - 8)}
	rect2.FillColorScale.SetColor(dpadBarColorBright)
	panel.cam.UI.AddGraphicsAbove(rect2)
	panel.factionRects = append(panel.factionRects, rect2)
}

func (panel *rpanelNode) initPrioritiesForColonies() {
	cameraWidth := panel.cam.Rect.Width()
	priorities := []colonyPriority{
		priorityResources,
		priorityGrowth,
		priorityEvolution,
		prioritySecurity,
	}
	for i, priority := range priorities {
		bar := panel.scene.NewSprite(assets.ImagePriorityBar)
		bar.Pos.Offset = gmath.Vec{
			X: (cameraWidth - (panel.layerSprite1.FrameWidth - 16)) + ((18 + bar.FrameWidth) * float64(i)),
		}
		bar.Centered = false
		panel.cam.UI.AddGraphics(bar)

		icon := panel.scene.NewSprite(assets.ImagePriorityIcons)
		setPriorityIconFrame(icon, priority, gamedata.NeutralFactionTag)
		icon.Pos.Offset = gmath.Vec{
			X: (cameraWidth - (panel.layerSprite1.FrameWidth - 16)) + ((18 + bar.FrameWidth) * float64(i)),
		}
		icon.Centered = false
		panel.cam.UI.AddGraphicsAbove(icon)

		panel.priorityBars = append(panel.priorityBars, bar)
		panel.priorityIcons = append(panel.priorityIcons, icon)
	}
}

func (panel *rpanelNode) initPrioritiesForCreeps() {
	cameraWidth := panel.cam.Rect.Width()

	dpadOffset := gmath.Vec{
		X: (cameraWidth - (panel.layerSprite1.FrameWidth - 85)),
		Y: panel.layerSprite1.FrameHeight - 94 + (panel.scene.Context().ScreenHeight - 540),
	}

	panel.dpad = panel.scene.NewSprite(assets.ImageDarkDPad)
	panel.dpad.Centered = false
	panel.dpad.Pos.Offset = dpadOffset
	panel.cam.UI.AddGraphicsAbove(panel.dpad)

	panel.dpadRects = make([]*ge.Rect, 4)
	for i := range panel.dpadRects {
		rect := ge.NewRect(panel.scene.Context(), 0, 0)
		rect.FillColorScale.SetColor(ge.RGB(0x3ea24d))
		rect.Centered = false
		panel.dpadRects[i] = rect
		panel.cam.UI.AddGraphicsAbove(rect)
	}
}

func (panel *rpanelNode) Init(scene *ge.Scene) {
	panel.scene = scene

	cameraWidth := panel.cam.Rect.Width()

	layer1image := assets.ImageRightPanelLayer1
	layer2image := assets.ImageRightPanelLayer2
	if panel.creepsState != nil {
		layer1image = assets.ImageDarkRightPanelLayer1
		layer2image = assets.ImageDarkRightPanelLayer2
	}

	panel.layerSprite1 = scene.NewSprite(layer1image)
	panel.layerSprite1.Pos.Offset.X = (cameraWidth - panel.layerSprite1.FrameWidth)
	panel.layerSprite1.Pos.Offset.Y += scene.Context().ScreenHeight - 540
	panel.layerSprite1.Centered = false
	panel.cam.UI.AddGraphicsBelow(panel.layerSprite1)

	if panel.creepsState != nil {
		panel.initPrioritiesForCreeps()
	} else {
		panel.initPrioritiesForColonies()
	}

	panel.layerSprite2 = scene.NewSprite(layer2image)
	panel.layerSprite2.Pos = panel.layerSprite1.Pos
	panel.layerSprite2.Centered = false
	panel.cam.UI.AddGraphicsAbove(panel.layerSprite2)

	if panel.creepsState != nil {
		panel.initFactionsForCreeps()
	} else {
		panel.initFactionsForColonies()
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
	if panel.creepsState != nil {
		panel.updateMetricsForCreeps()
	} else {
		panel.updateMetricsForColony()
	}
}

func (panel *rpanelNode) updateMetricsForColony() {
	if panel.colony == nil {
		return
	}

	// Update factions distribution rects.
	topOffset := 8.0 + (panel.scene.Context().ScreenHeight - 540)
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

	fullPriorityOffset := 445.0 + (panel.scene.Context().ScreenHeight - 540)
	for i, kv := range panel.colony.priorities.Elems {
		bar := panel.priorityBars[i]
		bar.Pos.Offset.Y = fullPriorityOffset + ((bar.FrameHeight - 8) * (1.0 - kv.Weight))
		icon := panel.priorityIcons[i]
		icon.Pos.Offset.Y = fullPriorityOffset + ((bar.FrameHeight - 8) * (1.0 - kv.Weight)) - icon.FrameHeight - 1
	}
}

func (panel *rpanelNode) updateMetricsForCreeps() {
	topOffset := 8.0
	totalHeight := 344.0

	{
		rect := panel.factionRects[0]
		rect.Height = gmath.ClampMax(panel.creepsState.techLevel, 1) * totalHeight
		rect.Visible = rect.Height > 0
		rect.Pos.Offset.Y = topOffset + totalHeight - rect.Height
	}
	{
		rect := panel.factionRects[1]
		rect.Height = gmath.Clamp(panel.creepsState.techLevel-1, 0, 1) * totalHeight
		rect.Visible = rect.Height > 0
		rect.Pos.Offset.Y = topOffset + totalHeight - rect.Height
	}

	const rectWidth float64 = 4
	const maxSize float64 = 24
	offsets := [...]gmath.Vec{
		{X: 12, Y: 0},
		{X: 0, Y: 12},
		{X: -(12 - rectWidth), Y: 0},
		{X: 0, Y: -(12 - rectWidth)},
	}
	fullPowerValue := panel.creepsState.maxSideCost
	centerOffset := panel.dpad.Pos.Offset.Add(gmath.Vec{X: 35, Y: 35})
	for dir, rect := range panel.dpadRects {
		isHorizontal := dir%2 == 0
		reversed := dir >= 2
		dirValue := panel.creepsState.attackSides[dir].totalCost
		dirPercentage := float64(dirValue) / float64(fullPowerValue)
		color := dpadBarColorNormal
		if dirPercentage >= 1 {
			dirPercentage = 1
			color = dpadBarColorBright
		}
		size := math.Ceil(maxSize * dirPercentage)
		rect.Pos.Offset = centerOffset.Add(offsets[dir])
		rect.FillColorScale.SetColor(color)
		if isHorizontal {
			rect.Width = size
			rect.Height = rectWidth
			if reversed {
				rect.Pos.Offset = rect.Pos.Offset.Sub(gmath.Vec{X: size})
			}
		} else {
			rect.Width = rectWidth
			rect.Height = size
			if reversed {
				rect.Pos.Offset = rect.Pos.Offset.Sub(gmath.Vec{Y: size})
			}
		}
	}
}

func (panel *rpanelNode) Update(delta float64) {}
