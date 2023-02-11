package assets

import (
	"embed"
	"io"

	"github.com/quasilyte/ge"
)

const (
	SoundGroupEffect uint = iota
	SoundGroupMusic
)

func VolumeMultiplier(level int) float64 {
	switch level {
	case 1:
		return 0.05
	case 2:
		return 0.10
	case 3:
		return 0.3
	case 4:
		return 0.55
	case 5:
		return 0.8
	case 6:
		return 1.0
	default:
		return 0
	}
}

func Register(ctx *ge.Context) {
	ctx.Loader.OpenAssetFunc = func(path string) io.ReadCloser {
		f, err := gameAssets.Open("_data/" + path)
		if err != nil {
			ctx.OnCriticalError(err)
		}
		return f
	}

	registerImageResources(ctx)
	registerAudioResource(ctx)
	registerShaderResources(ctx)
	registerFontResources(ctx)
	registerRawResources(ctx)
}

//go:embed all:_data
var gameAssets embed.FS
