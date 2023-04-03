package contentlock

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/session"
)

func GetDefaultData() session.PersistentData {
	return session.PersistentData{
		// The default settings.
		Settings: session.GameSettings{
			EffectsVolumeLevel: 2,
			MusicVolumeLevel:   2,
			ScrollingSpeed:     2,
			EdgeScrollRange:    2,
			Debug:              false,
			Lang:               inferDefaultLang(),
			Graphics: session.GraphicsSettings{
				ShadowsEnabled:    true,
				AllShadersEnabled: true,
				FullscreenEnabled: true,
			},
		},
	}
}

func inferDefaultLang() string {
	languages := ge.InferLanguages()
	defaultLanguage := "en"
	selectedLanguage := ""
	for _, l := range languages {
		switch l {
		case "en", "ru":
			if selectedLanguage != defaultLanguage {
				selectedLanguage = l
			}
		}
	}
	if selectedLanguage == "" {
		selectedLanguage = defaultLanguage
	}
	return selectedLanguage
}

type Result struct {
	DronesUnlocked []gamedata.ColonyAgentKind
}

func Update(state *session.State) *Result {
	result := &Result{}

	stats := &state.Persistent.PlayerStats

	alreadyUnlocked := map[gamedata.ColonyAgentKind]struct{}{}
	for _, kind := range stats.DronesUnlocked {
		alreadyUnlocked[kind] = struct{}{}
	}
	for _, recipe := range gamedata.Tier2agentMergeRecipes {
		drone := recipe.Result
		if _, ok := alreadyUnlocked[drone.Kind]; ok {
			continue
		}
		if drone.ScoreCost > stats.TotalScore {
			continue
		}
		result.DronesUnlocked = append(result.DronesUnlocked, drone.Kind)
		stats.DronesUnlocked = append(stats.DronesUnlocked, drone.Kind)
	}

	return result
}
