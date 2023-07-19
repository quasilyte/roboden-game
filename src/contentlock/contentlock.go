package contentlock

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/session"
)

func GetDefaultData() session.PersistentData {
	defaultGamepadSettings := session.GamepadSettings{
		Layout:        int(gameinput.GamepadLayoutXbox),
		DeadzoneLevel: 1,
		CursorSpeed:   3,
	}
	return session.PersistentData{
		// The default settings.
		Settings: session.GameSettings{
			Player1InputMethod: int(gameinput.InputMethodCombined),
			Player2InputMethod: int(gameinput.InputMethodGamepad2),
			EffectsVolumeLevel: 2,
			MusicVolumeLevel:   2,
			ScrollingSpeed:     2,
			EdgeScrollRange:    2,
			HintMode:           2,
			Demo:               true,
			ShowFPS:            false,
			Lang:               inferDefaultLang(),
			Graphics: session.GraphicsSettings{
				ShadowsEnabled:       true,
				AllShadersEnabled:    true,
				FullscreenEnabled:    true,
				CameraShakingEnabled: true,
				VSyncEnabled:         true,
			},
			GamepadSettings: [2]session.GamepadSettings{
				defaultGamepadSettings,
				defaultGamepadSettings,
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
	CoresUnlocked   []string
	DronesUnlocked  []gamedata.ColonyAgentKind
	TurretsUnlocked []gamedata.ColonyAgentKind
}

func Update(state *session.State) *Result {
	result := &Result{}

	stats := &state.Persistent.PlayerStats

	coresUnlocked := map[string]struct{}{}
	for _, name := range stats.CoresUnlocked {
		coresUnlocked[name] = struct{}{}
	}
	for _, core := range gamedata.CoreStatsList {
		if _, ok := coresUnlocked[core.Name]; ok {
			continue
		}
		if core.ScoreCost > stats.TotalScore {
			continue
		}
		result.CoresUnlocked = append(result.CoresUnlocked, core.Name)
		stats.CoresUnlocked = append(stats.CoresUnlocked, core.Name)
	}

	alreadyUnlocked := map[gamedata.ColonyAgentKind]struct{}{}
	for _, name := range stats.DronesUnlocked {
		alreadyUnlocked[gamedata.DroneKindByName[name]] = struct{}{}
	}
	for _, name := range stats.TurretsUnlocked {
		alreadyUnlocked[gamedata.DroneKindByName[name]] = struct{}{}
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
		stats.DronesUnlocked = append(stats.DronesUnlocked, drone.Kind.String())
	}
	for _, turret := range gamedata.TurretStatsList {
		if _, ok := alreadyUnlocked[turret.Kind]; ok {
			continue
		}
		if turret.ScoreCost > stats.TotalScore {
			continue
		}
		result.TurretsUnlocked = append(result.TurretsUnlocked, turret.Kind)
		stats.TurretsUnlocked = append(stats.TurretsUnlocked, turret.Kind.String())
	}

	return result
}
