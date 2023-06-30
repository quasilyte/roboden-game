package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type cloningBeamNode struct {
	disposed bool

	world   *worldState
	merging bool

	from *gmath.Vec
	to   ge.Pos

	delay      float64
	soundDelay float64

	lines [3]*ge.Line
}

func newCloningBeamNode(world *worldState, merging bool, from *gmath.Vec, to ge.Pos) *cloningBeamNode {
	return &cloningBeamNode{world: world, merging: merging, from: from, to: to}
}

func (b *cloningBeamNode) Init(scene *ge.Scene) {
	b.lines[0] = ge.NewLine(ge.Pos{Base: b.from}, b.to)
	b.lines[1] = ge.NewLine(ge.Pos{Base: b.from}, b.to)
	b.lines[2] = ge.NewLine(ge.Pos{}, ge.Pos{})

	for i := range b.lines {
		b.world.stage.AddGraphicsAbove(b.lines[i])
		if b.merging {
			b.lines[i].SetColorScaleRGBA(0xa2, 0x4c, 0xba, 255)
		} else {
			b.lines[i].SetColorScaleRGBA(0x33, 0x80, 0xbb, 255)
		}
	}
}

func (b *cloningBeamNode) Update(delta float64) {
	b.delay -= delta
	b.soundDelay -= delta
	if b.delay <= 0 {
		b.delay = 0.06
		offset1 := b.world.localRand.Offset(-6, 6)
		offset2 := b.world.localRand.Offset(-6, 6)
		b.lines[0].EndPos.Offset = b.to.Offset.Add(offset1)
		b.lines[1].EndPos.Offset = b.to.Offset.Add(offset2)
		b.lines[2].BeginPos.Offset = b.lines[0].EndPos.Resolve()
		b.lines[2].EndPos.Offset = b.lines[1].EndPos.Resolve()
	}

	if b.soundDelay <= 0 {
		if b.merging {
			if b.world.localRand.Bool() {
				b.soundDelay = b.world.localRand.FloatRange(0.5, 0.75)
				playSound(b.world, assets.AudioMerging1, *b.from)
			} else {
				b.soundDelay = b.world.localRand.FloatRange(0.55, 0.9)
				playSound(b.world, assets.AudioMerging2, *b.from)
			}
		} else {
			if b.world.localRand.Bool() {
				b.soundDelay = b.world.localRand.FloatRange(0.3, 0.7)
				playSound(b.world, assets.AudioCloning1, *b.from)
			} else {
				b.soundDelay = b.world.localRand.FloatRange(0.25, 0.6)
				playSound(b.world, assets.AudioCloning2, *b.from)
			}
		}
	}
}

func (b *cloningBeamNode) Dispose() {
	b.disposed = true
	for i := range b.lines {
		b.lines[i].Dispose()
	}
}

func (b *cloningBeamNode) IsDisposed() bool {
	return b.disposed
}
