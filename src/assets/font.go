package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
)

func registerFontResources(ctx *ge.Context) {
	fontResources := map[resource.FontID]resource.FontInfo{
		FontTiny:   {Path: "font/Retron2000.ttf", Size: 10},
		FontSmall:  {Path: "font/Retron2000.ttf", Size: 14},
		FontNormal: {Path: "font/Retron2000.ttf", Size: 16},
		FontBig:    {Path: "font/Retron2000.ttf", Size: 26},
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
