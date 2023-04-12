package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
)

func RegisterFontResources(ctx *ge.Context, progress *float64) {
	fontResources := map[resource.FontID]resource.FontInfo{
		FontTiny:   {Path: "font/Retron2000.ttf", Size: 10},
		FontSmall:  {Path: "font/Retron2000.ttf", Size: 14},
		FontNormal: {Path: "font/Retron2000.ttf", Size: 16},
		FontBig:    {Path: "font/Retron2000.ttf", Size: 26},
	}

	progressPerItem := 1.0 / float64(len(fontResources))
	for id, res := range fontResources {
		ctx.Loader.FontRegistry.Set(id, res)
		ctx.Loader.LoadFont(id)
		if progress != nil {
			*progress += progressPerItem
		}
	}
}

const (
	FontSmall resource.FontID = iota
	FontTiny
	FontNormal
	FontBig
)
