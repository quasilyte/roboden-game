package gameui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

func GenerateRecipePreviews(scene *ge.Scene, needT3 bool) map[gamedata.RecipeSubject]*ebiten.Image {
	createSubImage := func(img resource.Image) *ebiten.Image {
		return img.Data.SubImage(image.Rectangle{
			Max: image.Point{
				X: int(img.DefaultFrameWidth),
				Y: int(img.DefaultFrameHeight),
			},
		}).(*ebiten.Image)
	}

	workerFrame := createSubImage(scene.LoadImage(assets.ImageWorkerAgent))
	scoutFrame := createSubImage(scene.LoadImage(assets.ImageScoutAgent))

	diode := scene.LoadImage(assets.ImageFactionDiode).Data
	diodeSize := diode.Bounds().Size()

	recipeIcons := make(map[gamedata.RecipeSubject]*ebiten.Image)

	for _, recipe := range gamedata.Tier2agentMergeRecipes {
		subjects := []gamedata.RecipeSubject{
			recipe.Drone1,
			recipe.Drone2,
		}
		for _, s := range subjects {
			if _, ok := recipeIcons[s]; ok {
				continue
			}

			diodeOffset := gamedata.WorkerAgentStats.DiodeOffset + 1
			droneFrame := workerFrame
			if s.Kind == gamedata.AgentScout {
				droneFrame = scoutFrame
				diodeOffset = gamedata.ScoutAgentStats.DiodeOffset + 2
			}
			frameSize := droneFrame.Bounds().Size()
			img := ebiten.NewImage(32, 32)
			var drawOptions ebiten.DrawImageOptions
			drawOptions.GeoM.Scale(2, 2)
			drawOptions.GeoM.Translate(16-float64(frameSize.X), 16-float64(frameSize.Y))
			img.DrawImage(droneFrame, &drawOptions)
			drawOptions.GeoM.Reset()
			drawOptions.GeoM.Scale(2, 2)
			drawOptions.GeoM.Translate(16-float64(diodeSize.X), 16-float64(diodeSize.Y)+diodeOffset)
			drawOptions.ColorM.ScaleWithColor(gamedata.FactionByTag(s.Faction).Color)
			img.DrawImage(diode, &drawOptions)

			recipeIcons[s] = img
		}
	}

	if !needT3 {
		return recipeIcons
	}

	for _, recipe := range gamedata.Tier3agentMergeRecipes {
		subjects := []gamedata.RecipeSubject{
			recipe.Drone1,
			recipe.Drone2,
		}
		for _, s := range subjects {
			if _, ok := recipeIcons[s]; ok {
				continue
			}

			stats := gamedata.FindRecipeByName(s.Kind.String())
			droneFrame := createSubImage(scene.LoadImage(stats.Result.Image))
			frameSize := droneFrame.Bounds().Size()
			img := ebiten.NewImage(48, 48)
			var drawOptions ebiten.DrawImageOptions
			drawOptions.GeoM.Scale(2, 2)
			drawOptions.GeoM.Translate(24-float64(frameSize.X), 24-float64(frameSize.Y))
			img.DrawImage(droneFrame, &drawOptions)
			drawOptions.GeoM.Reset()
			drawOptions.GeoM.Scale(2, 2)

			recipeIcons[s] = img
		}
	}

	return recipeIcons
}
