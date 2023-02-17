package main

import (
	"flag"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/scenes/menus"
	"github.com/quasilyte/roboden-game/session"
)

func main() {
	state := &session.State{
		LevelOptions: session.LevelOptions{
			Resources:  2,
			Difficulty: 2,
			WorldSize:  2,
		},
		Persistent: session.PersistentData{
			// The default settings.
			Settings: session.GameSettings{
				EffectsVolumeLevel: 2,
				MusicVolumeLevel:   2,
				ScrollingSpeed:     2,
				EdgeScrollRange:    2,
				Debug:              false,
				Lang:               inferDefaultLang(),
			},
		},
	}

	flag.StringVar(&state.MemProfile, "memprofile", "", "collect app heap allocations profile")
	flag.StringVar(&state.CPUProfile, "cpuprofile", "", "collect app cpu profile")
	flag.Parse()

	ctx := ge.NewContext()
	ctx.Rand.SetSeed(time.Now().Unix())
	ctx.GameName = "roboden"
	ctx.WindowTitle = "Roboden"
	ctx.WindowWidth = 1920 / 2
	ctx.WindowHeight = 1080 / 2
	ctx.FullScreen = true

	assets.Register(ctx)
	controls.BindKeymap(ctx, state)

	ctx.LoadGameData("save", &state.Persistent)
	state.ReloadLanguage(ctx)

	if err := ge.RunGame(ctx, menus.NewMainMenuController(state)); err != nil {
		panic(err)
	}
}

func inferDefaultLang() string {
	languages := ge.InferLanguages()
	defaultLanguage := "en"
	selectedLanguage := ""
	for _, l := range languages {
		switch l {
		case "en", "ru":
			if selectedLanguage == "" || selectedLanguage != defaultLanguage {
				selectedLanguage = l
			}
		}
	}
	if selectedLanguage == "" {
		selectedLanguage = defaultLanguage
	}
	return selectedLanguage
}
