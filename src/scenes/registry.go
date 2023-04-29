package scenes

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/serverapi"
)

type Registry struct {
	UserNameMenu func(backController ge.SceneController) ge.SceneController

	SubmitScreen func(backController ge.SceneController, replays []serverapi.GameReplay) ge.SceneController
}
