package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/session"
)

type objectiveDisplayNode struct {
	bg   *ge.Sprite
	text *ge.Label

	config *session.LevelConfig
}

func newObjectiveDisplay(config *session.LevelConfig) *objectiveDisplayNode {
	return &objectiveDisplayNode{
		config: config,
	}
}

func (d *objectiveDisplayNode) ToggleVisibility() {
	d.bg.Visible = !d.bg.Visible
	d.text.Visible = !d.text.Visible
}

func (d *objectiveDisplayNode) Init(scene *ge.Scene) {
	d.bg = scene.NewSprite(assets.ImageObjectiveDisplay)
	d.bg.Centered = false
	d.bg.Pos.Offset.Y = 408
	scene.AddGraphicsAbove(d.bg, 1)

	d.text = scene.NewLabel(assets.FontTiny)
	d.text.Pos = d.bg.Pos.WithOffset(0, -8)
	d.text.Width = d.bg.FrameWidth
	d.text.Height = d.bg.FrameHeight
	d.text.AlignHorizontal = ge.AlignHorizontalCenter
	d.text.AlignVertical = ge.AlignVerticalCenter
	d.text.Text = d.missionObjectiveText(scene.Dict())
	d.text.ColorScale.SetColor(ge.RGB(0x9dd793))
	scene.AddGraphicsAbove(d.text, 1)
}

func (d *objectiveDisplayNode) missionObjectiveText(dict *langs.Dictionary) string {
	s := dict.Get("ui.mission_objective") + "\n\n"
	if d.config.Tutorial != nil {
		s += dict.Get("objective", d.config.Tutorial.Objective.String())
	} else {
		switch d.config.GameMode {
		case gamedata.ModeClassic:
			s += dict.Get("objective.boss")
		}
	}
	return s
}

func (d *objectiveDisplayNode) Update(delta float64) {}

func (d *objectiveDisplayNode) IsDisposed() bool { return false }
