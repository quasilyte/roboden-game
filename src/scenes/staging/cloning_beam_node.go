package staging

import (
	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/colony-game/viewport"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type cloningBeamNode struct {
	disposed bool

	camera *viewport.Camera
	scene  *ge.Scene

	merging bool

	from *gmath.Vec
	to   ge.Pos

	delay      float64
	soundDelay float64

	lines [3]*ge.Line
}

func newCloningBeamNode(camera *viewport.Camera, merging bool, from *gmath.Vec, to ge.Pos) *cloningBeamNode {
	return &cloningBeamNode{camera: camera, merging: merging, from: from, to: to}
}

func (b *cloningBeamNode) Init(scene *ge.Scene) {
	b.scene = scene

	b.lines[0] = ge.NewLine(ge.Pos{Base: b.from}, b.to)
	b.lines[1] = ge.NewLine(ge.Pos{Base: b.from}, b.to)
	b.lines[2] = ge.NewLine(ge.Pos{}, ge.Pos{})

	for i := range b.lines {
		b.camera.AddGraphicsAbove(b.lines[i])
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
		offset1 := b.scene.Rand().Offset(-6, 6)
		offset2 := b.scene.Rand().Offset(-6, 6)
		b.lines[0].EndPos.Offset = b.to.Offset.Add(offset1)
		b.lines[1].EndPos.Offset = b.to.Offset.Add(offset2)
		b.lines[2].BeginPos.Offset = b.lines[0].EndPos.Resolve()
		b.lines[2].EndPos.Offset = b.lines[1].EndPos.Resolve()
	}

	if b.soundDelay <= 0 {
		if b.merging {
			if b.scene.Rand().Bool() {
				b.soundDelay = b.scene.Rand().FloatRange(0.5, 0.75)
				playSound(b.scene, b.camera, assets.AudioMerging1, *b.from)
			} else {
				b.soundDelay = b.scene.Rand().FloatRange(0.55, 0.9)
				playSound(b.scene, b.camera, assets.AudioMerging2, *b.from)
			}
		} else {
			if b.scene.Rand().Bool() {
				b.soundDelay = b.scene.Rand().FloatRange(0.3, 0.7)
				playSound(b.scene, b.camera, assets.AudioCloning1, *b.from)
			} else {
				b.soundDelay = b.scene.Rand().FloatRange(0.25, 0.6)
				playSound(b.scene, b.camera, assets.AudioCloning2, *b.from)
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
