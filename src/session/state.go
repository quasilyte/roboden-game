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
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/userdevice"
)

type State struct {
	CPUProfile       string
	CPUProfileWriter io.WriteCloser
	MemProfile       string
	MemProfileWriter io.WriteCloser

	Device userdevice.Info

	MainInput *input.Handler

	LevelConfig      *LevelConfig
	ArenaLevelConfig *LevelConfig

	Persistent PersistentData

	Resources Resources
}

type PersistentData struct {
	Settings GameSettings

	SeenClassicMode bool
	SeenArenaMode   bool

	PlayerStats PlayerStats
}

type PlayerStats struct {
	Achievements []Achievement

	TurretsUnlocked []string
	DronesUnlocked  []string
	Tier3DronesSeen []string

	TutorialsCompleted []int

	TotalPlayTime time.Duration
	TotalScore    int

	HighestClassicScore           int
	HighestClassicScoreDifficulty int

	HighestArenaScore           int
	HighestArenaScoreDifficulty int

	HighestInfArenaScore           int
	HighestInfArenaScoreDifficulty int
}

type Achievement struct {
	Name  string
	Elite bool
}

type LevelConfig struct {
	Resources int

	ExtraUI bool

	GameMode gamedata.Mode

	AttackActionAvailable      bool
	BuildTurretActionAvailable bool
	RadiusActionAvailable      bool

	FogOfWar     bool
	InfiniteMode bool

	SecondBase  bool
	ExtraDrones []*gamedata.AgentStats

	EliteResources    bool
	EnemyBoss         bool
	InitialCreeps     int
	NumCreepBases     int
	CreepDifficulty   int
	BossDifficulty    int
	ArenaProgression  int
	StartingResources int
	GameSpeed         int

	Seed int64

	WorldSize int

	Tutorial        *gamedata.TutorialData
	DifficultyScore int

	Tier2Recipes         []gamedata.AgentMergeRecipe
	TurretDesign         *gamedata.AgentStats
	DronePointsAllocated int
}

func (options *LevelConfig) Clone() LevelConfig {
	cloned := *options

	cloned.Tier2Recipes = make([]gamedata.AgentMergeRecipe, len(options.Tier2Recipes))
	copy(cloned.Tier2Recipes, options.Tier2Recipes)

	return cloned
}

type Resources struct {
	UI *eui.Resources
}

type GameSettings struct {
	Lang               string
	MusicVolumeLevel   int
	EffectsVolumeLevel int
	ScrollingSpeed     int
	EdgeScrollRange    int
	ShowFPS            bool
	DebugLogs          bool
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
	if err := dict.Load("", ctx.Loader.LoadRaw(id+2).Data); err != nil {
		panic(err)
	}
	if err := dict.Load("", ctx.Loader.LoadRaw(id+3).Data); err != nil {
		panic(err)
	}
	ctx.Dict = dict
}
