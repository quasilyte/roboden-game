package session

import (
	"github.com/quasilyte/ge/input"
)

type State struct {
	MainInput *input.Handler

	LevelOptions LevelOptions

	Persistent PersistentData
}

type PersistentData struct {
	Settings GameSettings
}

type LevelOptions struct {
	Resources  int
	Difficulty int

	WorldSize int

	Tutorial bool
}

type GameSettings struct {
	MusicVolumeLevel   int
	EffectsVolumeLevel int
	ScrollingSpeed     int
	Debug              bool
}
