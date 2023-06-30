package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
)

type builderLaserNode struct {
	scene        *ge.Scene
	world        *worldState
	pos          gmath.Vec
	disposed     bool
	arms         []*ge.Line
	armPositions [5]gmath.Vec
	dstPositions [5]gmath.Vec
	armMoveDelay float64
}

func newBuilderLaserNode(world *worldState, pos gmath.Vec) *builderLaserNode {
	return &builderLaserNode{
		world: world,
		pos:   pos,
	}
}

func (laser *builderLaserNode) IsDisposed() bool {
	return laser.disposed
}

func (laser *builderLaserNode) Dispose() {
	laser.disposed = true
	for _, arm := range laser.arms {
		arm.Dispose()
	}
}

func (laser *builderLaserNode) Init(scene *ge.Scene) {
	laser.scene = scene
	laser.arms = make([]*ge.Line, 5)
	laser.armPositions = [...]gmath.Vec{
		laser.pos.Add(gmath.Vec{X: 0, Y: 14}),
		laser.pos.Add(gmath.Vec{X: -7, Y: 13}),
		laser.pos.Add(gmath.Vec{X: 7, Y: 13}),
		laser.pos.Add(gmath.Vec{X: -13, Y: 11}),
		laser.pos.Add(gmath.Vec{X: 13, Y: 11}),
	}
	laser.dstPositions = [...]gmath.Vec{
		laser.pos.Add(gmath.Vec{X: 0, Y: 16 + 36}),
		laser.pos.Add(gmath.Vec{X: -4, Y: 15 + 34}),
		laser.pos.Add(gmath.Vec{X: 4, Y: 15 + 34}),
		laser.pos.Add(gmath.Vec{X: -8, Y: 13 + 32}),
		laser.pos.Add(gmath.Vec{X: 8, Y: 13 + 32}),
	}
	for i := range laser.arms {
		arm := ge.NewLine(ge.Pos{Base: &laser.armPositions[i]}, ge.Pos{Base: &laser.dstPositions[i]})
		var colorScale ge.ColorScale
		colorScale.SetColor(builderBeamColor)
		arm.SetColorScale(colorScale)
		laser.world.stage.AddGraphics(arm)
		laser.arms[i] = arm
	}
}

func (laser *builderLaserNode) OnBuildingStop(gsignal.Void) {
	laser.Dispose()
}

func (laser *builderLaserNode) Update(delta float64) {
	laser.armMoveDelay = gmath.ClampMin(laser.armMoveDelay-delta, 0)
	if laser.armMoveDelay == 0 {
		laser.armMoveDelay = laser.world.localRand.FloatRange(0.05, 0.1)
		armMoved := laser.world.localRand.IntRange(0, 4)

		laser.arms[armMoved].EndPos.Offset = laser.world.localRand.Offset(-2, 2)
	}
}
