package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func registerRawResources(ctx *ge.Context) {
	rawResources := map[resource.RawID]resource.RawInfo{
		RawTilesJSON: {Path: "raw/tiles.json"},
	}

	for id, res := range rawResources {
		ctx.Loader.RawRegistry.Set(id, res)
		ctx.Loader.LoadRaw(id)
	}
}

const (
	RawNone resource.RawID = iota

	RawTilesJSON
)
