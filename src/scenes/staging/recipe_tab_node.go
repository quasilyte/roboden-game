package staging

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type recipeTabNode struct {
	world *worldState

	Visible       bool
	combinedImage *ebiten.Image

	pos gmath.Vec

	width  float64
	height float64

	rects []recipeTabRect
}

type recipeTabRect struct {
	drone *gamedata.AgentStats
	rect  gmath.Rect
}

func newRecipeTabNode(world *worldState) *recipeTabNode {
	return &recipeTabNode{world: world}
}

func (tab *recipeTabNode) ContainsPos(pos gmath.Vec) bool {
	bounds := gmath.Rect{
		Min: tab.pos,
		Max: tab.pos.Add(gmath.Vec{X: tab.width, Y: tab.height}),
	}
	return bounds.Contains(pos)
}

func (tab *recipeTabNode) GetDroneUnderCursor(pos gmath.Vec) *gamedata.AgentStats {
	for _, r := range tab.rects {
		if r.rect.Contains(pos) {
			return r.drone
		}
	}
	return nil
}

func (tab *recipeTabNode) IsDisposed() bool { return false }

func (tab *recipeTabNode) Init(scene *ge.Scene) {
	if len(tab.world.config.Tier2Recipes) == 0 {
		return
	}

	getDroneFrame := func(img resource.Image) *ebiten.Image {
		return img.Data.SubImage(image.Rectangle{
			Max: image.Point{X: int(img.DefaultFrameWidth), Y: int(img.DefaultFrameHeight)},
		}).(*ebiten.Image)
	}

	diode := scene.LoadImage(assets.ImageFactionDiode).Data
	diodeSize := diode.Bounds().Size()

	extraOffsets := [...]float64{
		gamedata.AgentGenerator:     -3,
		gamedata.AgentRedminer:      -2,
		gamedata.AgentDisintegrator: -1,
		gamedata.AgentRepeller:      -1,
		gamedata.AgentRepair:        -1,
		gamedata.AgentCloner:        -1,
		gamedata.AgentCrippler:      -1,
	}

	drawDrone := func(dst *ebiten.Image, stats *gamedata.AgentStats, faction gamedata.FactionTag, cellWidth, offsetX, offsetY float64) {
		halfWidth := cellWidth * 0.5
		droneImage := scene.LoadImage(stats.Image)
		droneFrame := getDroneFrame(droneImage)
		frameSize := droneFrame.Bounds().Size()
		var drawOptions ebiten.DrawImageOptions
		drawOptions.GeoM.Translate(offsetX, offsetY)
		if int(stats.Kind) < len(extraOffsets) {
			drawOptions.GeoM.Translate(0, extraOffsets[stats.Kind])
		}
		drawOptions.GeoM.Translate(halfWidth-(float64(frameSize.X)*0.5), 15-(float64(frameSize.Y)*0.5))
		dst.DrawImage(droneFrame, &drawOptions)
		if faction != gamedata.NeutralFactionTag {
			drawOptions.GeoM.Reset()
			drawOptions.GeoM.Translate(offsetX, offsetY)
			drawOptions.GeoM.Translate(halfWidth-(float64(diodeSize.X)*0.5), 15-(float64(diodeSize.Y)*0.5)+stats.DiodeOffset)
			drawOptions.ColorM.ScaleWithColor(gamedata.FactionByTag(faction).Color)
			dst.DrawImage(diode, &drawOptions)
		}
	}

	droneStatsByKind := func(kind gamedata.ColonyAgentKind) *gamedata.AgentStats {
		if kind == gamedata.AgentWorker {
			return gamedata.WorkerAgentStats
		}
		if kind == gamedata.AgentScout {
			return gamedata.ScoutAgentStats
		}
		for _, recipe := range tab.world.tier2recipes {
			if recipe.Result.Kind == kind {
				return recipe.Result
			}
		}
		panic("should never happen")
	}

	tab.rects = make([]recipeTabRect, 0, len(tab.world.tier2recipes))
	tab.pos = gmath.Vec{X: 8, Y: 8}

	numRecipes := len(tab.world.config.Tier2Recipes)
	droneSeparator := 4
	droneWidth := 30.0
	imageWidth := int(droneWidth)*numRecipes + ((numRecipes - 1) * droneSeparator)
	imageHeight := 15 + 5 + 30

	tile := ge.NewRect(scene.Context(), float64(droneWidth+2), float64(imageHeight+2))
	tile.Centered = false
	tile.OutlineWidth = 0
	tile.FillColorScale.SetRGBA(0x13, 0x1a, 0x22, 160)

	combined := ebiten.NewImage(imageWidth+2, imageHeight+2)
	offsetX := 0.0
	for i, recipe := range tab.world.tier2recipes {
		rect := gmath.Rect{
			Min: tab.pos.Add(gmath.Vec{X: offsetX}),
			Max: tab.pos.Add(gmath.Vec{X: offsetX + droneWidth + 2, Y: float64(imageHeight + 2)}),
		}
		tile.DrawWithOffset(combined, gmath.Vec{X: offsetX})
		marginX := 1.0
		marginY := 1.0
		drawDrone(combined, droneStatsByKind(recipe.Drone1.Kind), recipe.Drone1.Faction, 15, marginX+offsetX, 0+marginY)
		drawDrone(combined, droneStatsByKind(recipe.Drone2.Kind), recipe.Drone2.Faction, 15, marginX+offsetX+15, 0+marginY)
		drawDrone(combined, recipe.Result, gamedata.NeutralFactionTag, 30, marginX+offsetX, 15+5+marginY)
		tab.rects = append(tab.rects, recipeTabRect{
			drone: recipe.Result,
			rect:  rect,
		})
		offsetX += droneWidth + float64(droneSeparator)

		tab.width += droneWidth
		if i != 0 {
			tab.width += float64(droneSeparator)
		}
	}
	tab.combinedImage = combined

	tab.height = float64(imageHeight)
}

func (tab *recipeTabNode) Draw(screen *ebiten.Image) {
	if !tab.Visible {
		return
	}
	if len(tab.world.config.Tier2Recipes) == 0 {
		return
	}

	var options ebiten.DrawImageOptions
	options.GeoM.Translate(tab.pos.X, tab.pos.Y)
	screen.DrawImage(tab.combinedImage, &options)
}

func (tab *recipeTabNode) Update(delta float64) {}
