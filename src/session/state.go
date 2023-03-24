package session

import (
	"fmt"
	"io"
	"time"

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

	LevelConfig *LevelConfig

	Persistent PersistentData
}

type PersistentData struct {
	Settings GameSettings

	PlayerStats PlayerStats
}

type PlayerStats struct {
	Achievements []Achievement

	TurretsUnlocked []gamedata.ColonyAgentKind
	DronesUnlocked  []gamedata.ColonyAgentKind
	Tier3DronesSeen []gamedata.ColonyAgentKind

	TutorialsCompleted []int

	TotalPlayTime          time.Duration
	TotalScore             int
	HighestScore           int
	HighestScoreDifficulty int
}

type Achievement struct {
	Name  string
	Elite bool
}

type LevelConfig struct {
	Resources int

	ExtraUI bool

	NumCreepBases     int
	CreepDifficulty   int
	BossDifficulty    int
	StartingResources int

	Seed int64

	WorldSize int

	Tutorial        *TutorialData
	DifficultyScore int

	Tier2Recipes         []gamedata.AgentMergeRecipe
	DronePointsAllocated int
}

func (options *LevelConfig) Clone() LevelConfig {
	cloned := *options

	cloned.Tier2Recipes = make([]gamedata.AgentMergeRecipe, len(options.Tier2Recipes))
	copy(cloned.Tier2Recipes, options.Tier2Recipes)

	return cloned
}

type TutorialData struct {
	ID int
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
	ShadowsEnabled    bool
	AllShadersEnabled bool
	FullscreenEnabled bool
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
	if err := dict.Load("", ctx.Loader.LoadRaw(id+1).Data); err != nil {
		panic(err)
	}
	ctx.Dict = dict
}
