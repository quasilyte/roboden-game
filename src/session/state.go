package session

import (
	"fmt"
	"io"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/scenes"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/userdevice"
)

type State struct {
	CPUProfile       string
	CPUProfileWriter io.WriteCloser
	MemProfile       string
	MemProfileWriter io.WriteCloser

	ServerProtocol string
	ServerHost     string
	ServerPath     string

	Device userdevice.Info

	MainInput   gameinput.Handler
	SecondInput gameinput.Handler

	ClassicLevelConfig  *gamedata.LevelConfig
	ArenaLevelConfig    *gamedata.LevelConfig
	InfArenaLevelConfig *gamedata.LevelConfig
	ReverseLevelConfig  *gamedata.LevelConfig
	TutorialLevelConfig *gamedata.LevelConfig

	Persistent PersistentData

	SceneRegistry scenes.Registry

	Resources Resources

	DemoFrame *ebiten.Image

	Context *ge.Context

	SentHighscores bool
}

type PersistentData struct {
	Settings GameSettings

	SeenClassicMode  bool
	SeenArenaMode    bool
	SeenInfArenaMode bool
	SeenReverseMode  bool

	PlayerName string

	NumPendingSubmissions int

	PlayerStats PlayerStats

	CachedClassicLeaderboard  serverapi.LeaderboardResp
	CachedArenaLeaderboard    serverapi.LeaderboardResp
	CachedInfArenaLeaderboard serverapi.LeaderboardResp
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
	ShowTimer          bool
	DebugLogs          bool
	Demo               bool
	SwapGamepads       bool
	Graphics           GraphicsSettings
}

type GraphicsSettings struct {
	ShadowsEnabled    bool
	AllShadersEnabled bool
	FullscreenEnabled bool
}

type SavedReplay struct {
	Date      time.Time
	ResultTag string
	Replay    serverapi.GameReplay
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

func (state *State) DetectInputMode() string {
	inputMode := "keyboard"
	if state.MainInput.GamepadConnected() {
		inputMode = "gamepad"
	}
	return inputMode
}

func (state *State) FindNextReplayIndex() int {
	var minDate time.Time
	minIndex := 0
	for i := 1; i < 10; i++ {
		k := state.ReplayDataKey(i)
		if !state.Context.CheckGameData(k) {
			return i
		}
		var r SavedReplay
		if err := state.Context.LoadGameData(k, &r); err != nil {
			if minIndex == 0 || r.Date.Before(minDate) {
				minDate = r.Date
				minIndex = i
			}
		}
	}
	if minIndex != 0 {
		return minIndex
	}
	return 1
}

func (state *State) ReplayDataKey(i int) string {
	return fmt.Sprintf("saved_replay_%d", i)
}
