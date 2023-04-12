package assets

import (
	"embed"
	"io"
	"path/filepath"
	"strings"

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

func MakeOpenAssetFunc(ctx *ge.Context, gamedataFolder string) func(path string) io.ReadCloser {
	return func(path string) io.ReadCloser {
		if strings.HasPrefix(path, "$") {
			f, err := openfile(filepath.Join(gamedataFolder, path[len("$"):]))
			if err != nil {
				ctx.OnCriticalError(err)
			}
			return f
		}
		f, err := gameAssets.Open("_data/" + path)
		if err != nil {
			ctx.OnCriticalError(err)
		}
		return f
	}
}

//go:embed all:_data
var gameAssets embed.FS
