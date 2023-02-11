package session

import (
	"io"

	"github.com/quasilyte/ge/input"
)

type State struct {
	CPUProfile       string
	CPUProfileWriter io.WriteCloser
	MemProfile       string
	MemProfileWriter io.WriteCloser

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
	EdgeScrollRange    int
	Debug              bool
}
