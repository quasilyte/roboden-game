package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func RegisterRawResources(ctx *ge.Context) {
	rawResources := map[resource.RawID]resource.RawInfo{
		RawTilesJSON:      {Path: "raw/tiles.json"},
		RawDictEn:         {Path: "raw/en.txt"},
		RawDictTutorialEn: {Path: "raw/en_tutorial.txt"},
		RawDictRu:         {Path: "raw/ru.txt"},
		RawDictTutorialRu: {Path: "raw/ru_tutorial.txt"},
	}

	for id, res := range rawResources {
		ctx.Loader.RawRegistry.Set(id, res)
		ctx.Loader.LoadRaw(id)
	}
}

const (
	RawNone resource.RawID = iota

	RawTilesJSON
	RawDictEn
	RawDictTutorialEn
	RawDictRu
	RawDictTutorialRu
)
