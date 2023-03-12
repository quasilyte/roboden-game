package session

import (
	"fmt"
	"io"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/ge/langs"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/userdevice"
)

type State struct {
	CPUProfile       string
	CPUProfileWriter io.WriteCloser
	MemProfile       string
	MemProfileWriter io.WriteCloser

	Device userdevice.Info

	MainInput *input.Handler

	LevelOptions LevelOptions

	Persistent PersistentData
}

type PersistentData struct {
	Settings GameSettings
}

type LevelOptions struct {
	Resources int

	CreepsDifficulty  int
	BossDifficulty    int
	StartingResources int

	WorldSize int

	Tutorial bool

	Tier2Recipes []gamedata.AgentMergeRecipe
}

type GameSettings struct {
	Lang               string
	MusicVolumeLevel   int
	EffectsVolumeLevel int
	ScrollingSpeed     int
	EdgeScrollRange    int
	Debug              bool
	Graphics           GraphicsSettings
}

type GraphicsSettings struct {
	ShadowsEnabled bool
}

func (state *State) ReloadLanguage(ctx *ge.Context) {
	var id resource.RawID
	lang := state.Persistent.Settings.Lang
	switch lang {
	case "en":
		id = assets.RawDictEn
	case "ru":
		id = assets.RawDictRu
	default:
		panic(fmt.Sprintf("unsupported lang: %q", lang))
	}
	dict, err := langs.ParseDictionary(lang, 4, ctx.Loader.LoadRaw(id).Data)
	if err != nil {
		panic(err)
	}
	ctx.Dict = dict
}
