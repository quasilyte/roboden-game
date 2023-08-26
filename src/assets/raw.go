package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func RegisterRawResources(ctx *ge.Context) {
	rawResources := map[resource.RawID]resource.RawInfo{
		RawTilesJSON:        {Path: "raw/tiles.json"},
		RawForestTilesJSON:  {Path: "raw/forest_tiles.json"},
		RawInfernoTilesJSON: {Path: "raw/inferno_tiles.json"},

		RawDictEn:             {Path: "raw/en.txt"},
		RawDictTutorialEn:     {Path: "raw/en_intro.txt"},
		RawDictAchievementsEn: {Path: "raw/en_achievements.txt"},
		RawDictDronesEn:       {Path: "raw/en_drones.txt"},

		RawDictRu:             {Path: "raw/ru.txt"},
		RawDictTutorialRu:     {Path: "raw/ru_intro.txt"},
		RawDictAchievementsRu: {Path: "raw/ru_achievements.txt"},
		RawDictDronesRu:       {Path: "raw/ru_drones.txt"},
	}

	for id, res := range rawResources {
		ctx.Loader.RawRegistry.Set(id, res)
		ctx.Loader.LoadRaw(id)
	}
}

const (
	RawNone resource.RawID = iota

	RawTilesJSON
	RawForestTilesJSON
	RawInfernoTilesJSON

	RawDictEn
	RawDictTutorialEn
	RawDictAchievementsEn
	RawDictDronesEn

	RawDictRu
	RawDictTutorialRu
	RawDictAchievementsRu
	RawDictDronesRu
)
