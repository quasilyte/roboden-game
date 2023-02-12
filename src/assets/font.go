package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
)

func registerFontResources(ctx *ge.Context) {
	fontResources := map[resource.FontID]resource.FontInfo{
		FontTiny:   {Path: "font/aesymatt.otf", Size: 10},
		FontSmall:  {Path: "font/aesymatt.otf", Size: 16},
		FontNormal: {Path: "font/aesymatt.otf", Size: 18},
		FontBig:    {Path: "font/aesymatt.otf", Size: 28},
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
