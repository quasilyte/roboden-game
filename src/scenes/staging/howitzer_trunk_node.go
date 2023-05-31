package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

type howitzerTrunkNode struct {
	pos    ge.Pos
	stage  *viewport.CameraStage
	sprite *ge.Sprite
}

func newHowitzerTrunkNode(stage *viewport.CameraStage, pos ge.Pos) *howitzerTrunkNode {
	return &howitzerTrunkNode{
		pos:   pos,
		stage: stage,
	}
}

func (trunk *howitzerTrunkNode) Init(scene *ge.Scene) {
	trunk.sprite = scene.NewSprite(assets.ImageHowitzerTrunk)
	trunk.sprite.Pos = trunk.pos
	trunk.stage.AddGraphics(trunk.sprite)
}

func (trunk *howitzerTrunkNode) SetVisibility(visible bool) {
	trunk.sprite.Visible = visible
}

func (trunk *howitzerTrunkNode) SetRotation(angle gmath.Rad) gmath.Vec {
	type frameOption struct {
		maxAngle   gmath.Rad
		frame      float64
		fireOffset gmath.Vec
	}
	options := [...]frameOption{
		{maxAngle: 0.45, frame: 0, fireOffset: gmath.Vec{X: 10, Y: -4}},
		{maxAngle: 1.15, frame: 1, fireOffset: gmath.Vec{X: 8, Y: -1}},
		{maxAngle: math.Pi - 1.15, frame: 2, fireOffset: gmath.Vec{X: 0, Y: 1}},
		{maxAngle: math.Pi - 0.45, frame: 3, fireOffset: gmath.Vec{X: -8, Y: -1}},
		{maxAngle: math.Pi - 0.15, frame: 4, fireOffset: gmath.Vec{X: -10, Y: -4}},

		{maxAngle: math.Pi + 0.45, frame: 5, fireOffset: gmath.Vec{X: -13, Y: -23}},
		{maxAngle: math.Pi + 1.15, frame: 6, fireOffset: gmath.Vec{X: -10, Y: -27}},
		{maxAngle: (2 * math.Pi) - 1.15, frame: 7, fireOffset: gmath.Vec{X: 0, Y: -30}},
		{maxAngle: (2 * math.Pi) - 0.45, frame: 8, fireOffset: gmath.Vec{X: 10, Y: -27}},
		{maxAngle: (2 * math.Pi) - 0.15, frame: 9, fireOffset: gmath.Vec{X: 13, Y: -23}},
	}
	for _, o := range options {
		if angle < o.maxAngle {
			trunk.sprite.FrameOffset.X = trunk.sprite.FrameWidth * o.frame
			return o.fireOffset
		}
	}
	return gmath.Vec{}
}

func (trunk *howitzerTrunkNode) Update(delta float64) {}

func (trunk *howitzerTrunkNode) IsDisposed() bool { return trunk.sprite.IsDisposed() }

func (trunk *howitzerTrunkNode) Dispose() {
	trunk.sprite.Dispose()
}
