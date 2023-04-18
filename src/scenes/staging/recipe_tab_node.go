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
	rect          *ge.Rect
}

func newRecipeTabNode(world *worldState) *recipeTabNode {
	return &recipeTabNode{world: world}
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
		if kind == gamedata.AgentMilitia {
			return gamedata.MilitiaAgentStats
		}
		for _, recipe := range tab.world.config.Tier2Recipes {
			if recipe.Result.Kind == kind {
				return recipe.Result
			}
		}
		panic("should never happen")
	}

	numRecipes := len(tab.world.config.Tier2Recipes)
	droneSeparator := 16
	imageWidth := 30*numRecipes + ((numRecipes - 1) * droneSeparator)
	imageHeight := 15 + 5 + 30
	combined := ebiten.NewImage(imageWidth, imageHeight)
	offsetX := 0.0
	for _, recipe := range tab.world.config.Tier2Recipes {
		drawDrone(combined, droneStatsByKind(recipe.Drone1.Kind), recipe.Drone1.Faction, 15, offsetX, 0)
		drawDrone(combined, droneStatsByKind(recipe.Drone2.Kind), recipe.Drone2.Faction, 15, offsetX+15, 0)
		drawDrone(combined, recipe.Result, gamedata.NeutralFactionTag, 30, offsetX, 15+5)
		offsetX += 30.0 + float64(droneSeparator)
	}
	tab.combinedImage = combined

	tab.rect = ge.NewRect(scene.Context(), float64(imageWidth+(8*2)), float64(imageHeight+(2*2)))
	tab.rect.Centered = false
	tab.rect.Pos.Offset = gmath.Vec{X: 8, Y: 8}
	tab.rect.OutlineColorScale.SetColor(ge.RGB(0x5e5a5d))
	tab.rect.OutlineWidth = 1
	tab.rect.FillColorScale.SetRGBA(0x13, 0x1a, 0x22, 230)
}

func (tab *recipeTabNode) Draw(screen *ebiten.Image) {
	if !tab.Visible {
		return
	}
	if len(tab.world.config.Tier2Recipes) == 0 {
		return
	}

	tab.rect.Draw(screen)

	var options ebiten.DrawImageOptions
	options.GeoM.Translate(tab.rect.Pos.Offset.X+8, tab.rect.Pos.Offset.Y+2)
	screen.DrawImage(tab.combinedImage, &options)
}

func (tab *recipeTabNode) Update(delta float64) {}
