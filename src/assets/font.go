package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
)

func registerFontResources(ctx *ge.Context) {
	fontResources := map[resource.FontID]resource.FontInfo{
		FontTiny:   {Path: "font/DejavuSansMono.ttf", Size: 10},
		FontSmall:  {Path: "font/DejavuSansMono.ttf", Size: 14},
		FontNormal: {Path: "font/DejavuSansMono.ttf", Size: 18},
		FontBig:    {Path: "font/DejavuSansMono.ttf", Size: 22},
	}

	for id, res := range fontResources {
		ctx.Loader.FontRegistry.Set(id, res)
		ctx.Loader.LoadFont(id)
	}
}

const (
	FontSmall resource.FontID = iota
	FontTiny
	FontNormal
	FontBig
)
