package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

type essenceSourceStats struct {
	image       resource.ImageID
	capacity    gmath.Range[int]
	regenDelay  float64 // 0 for "no regen"
	value       float64 // Resource score per unit
	eliteValue  float64 // Elite resource score per unit
	spritesheet bool
	canRotate   bool
	size        float64
}

var redCrystalSource = &essenceSourceStats{
	image:       assets.ImageEssenceRedCrystalSource,
	capacity:    gmath.MakeRange(1, 1),
	value:       35,
	eliteValue:  3,
	spritesheet: true,
	size:        32,
}

var oilSource = &essenceSourceStats{
	image:       assets.ImageEssenceSource,
	capacity:    gmath.MakeRange(50, 80),
	regenDelay:  7,
	value:       4, // 200-320 total
	spritesheet: true,
	size:        32,
}

var redOilSource = &essenceSourceStats{
	image:       assets.ImageRedEssenceSource,
	capacity:    gmath.MakeRange(60, 80),
	regenDelay:  9,
	value:       5, // 300-400 total
	eliteValue:  0.5,
	spritesheet: true,
	size:        32,
}

var goldSource = &essenceSourceStats{
	image:       assets.ImageEssenceGoldSource,
	capacity:    gmath.MakeRange(25, 40),
	regenDelay:  0, // none
	value:       6, // 150-240 total
	spritesheet: true,
	size:        20,
}

var crystalSource = &essenceSourceStats{
	image:       assets.ImageEssenceCrystalSource,
	capacity:    gmath.MakeRange(10, 20),
	regenDelay:  0,  // none
	value:       16, // 160-320 total
	spritesheet: true,
	size:        16,
}

var ironSource = &essenceSourceStats{
	image:       assets.ImageEssenceIronSource,
	capacity:    gmath.MakeRange(60, 80),
	regenDelay:  0, // none
	value:       2, // 120-160 total
	spritesheet: true,
	size:        18,
}

var smallScrapSource = &essenceSourceStats{
	image:      assets.ImageEssenceSmallScrapSource,
	capacity:   gmath.MakeRange(4, 5),
	regenDelay: 0, // none
	value:      1, // 4-5
	size:       14,
}

var scrapSource = &essenceSourceStats{
	image:      assets.ImageEssenceScrapSource,
	capacity:   gmath.MakeRange(8, 12),
	regenDelay: 0, // none
	value:      1, // 8-12
	size:       16,
}

var smallScrapCreepSource = &essenceSourceStats{
	image:      assets.ImageEssenceSmallScrapCreepSource,
	capacity:   gmath.MakeRange(5, 7),
	regenDelay: 0, // none
	value:      2, // 10-14
	size:       14,
}

var scrapCreepSource = &essenceSourceStats{
	image:      assets.ImageEssenceScrapCreepSource,
	capacity:   gmath.MakeRange(8, 14),
	regenDelay: 0, // none
	value:      2, // 16-28
	size:       16,
}

var bigScrapCreepSource = &essenceSourceStats{
	image:      assets.ImageEssenceBigScrapCreepSource,
	capacity:   gmath.MakeRange(12, 20),
	regenDelay: 0, // none
	value:      2, // 24-40
	size:       20,
}

type essenceSourceNode struct {
	scene *ge.Scene

	camera *viewport.Camera

	stats *essenceSourceStats

	sprite *ge.Sprite

	capacity     int
	resource     int
	percengage   float64
	recoverDelay float64
	added        float64

	rotation gmath.Rad
	pos      gmath.Vec

	EventDestroyed gsignal.Event[*essenceSourceNode]
}

func newEssenceSourceNode(camera *viewport.Camera, stats *essenceSourceStats, pos gmath.Vec) *essenceSourceNode {
	return &essenceSourceNode{
		camera: camera,
		stats:  stats,
		pos:    pos,
	}
}

func (e *essenceSourceNode) Init(scene *ge.Scene) {
	e.scene = scene

	e.sprite = scene.NewSprite(e.stats.image)
	e.sprite.Pos.Base = &e.pos
	e.sprite.Rotation = &e.rotation
	if !e.stats.spritesheet {
		e.sprite.Shader = scene.NewShader(assets.ShaderDissolve)
		e.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageEssenceSourceDissolveMask)
		e.sprite.Shader.Enabled = false
	}
	e.camera.AddSpriteBelow(e.sprite)

	if e.stats.canRotate {
		e.rotation = scene.Rand().Rad()
	} else {
		e.sprite.FlipHorizontal = scene.Rand().Bool()
	}

	e.capacity = scene.Rand().IntRange(e.stats.capacity.Min, e.stats.capacity.Max)
	e.resource = e.capacity
	e.percengage = 1.0
}

func (e *essenceSourceNode) IsDisposed() bool { return e.sprite.IsDisposed() }

func (e *essenceSourceNode) Update(delta float64) {
	if e.stats.regenDelay == 0 {
		return
	}
	e.recoverDelay -= delta
	if e.recoverDelay <= 0 {
		e.recoverDelay = e.stats.regenDelay * e.scene.Rand().FloatRange(0.75, 1.25)
		e.resource = gmath.ClampMax(e.resource+1, e.capacity)
		e.percengage = float64(e.resource) / float64(e.capacity)
		e.updateShader()
	}
}

func (e *essenceSourceNode) Add(delta float64) {
	e.added += delta
	changed := false
	for e.added >= 1 {
		e.added--
		e.resource++
		changed = true
	}
	if changed {
		e.percengage = float64(e.resource) / float64(e.capacity)
		e.updateShader()
	}
}

func (e *essenceSourceNode) Harvest(n int) int {
	n = gmath.ClampMax(n, e.resource)
	e.resource -= n
	e.percengage = float64(e.resource) / float64(e.capacity)
	if e.resource <= 0 && e.stats.regenDelay == 0 {
		e.Destroy()
	} else {
		e.updateShader()
	}
	return n
}

func (e *essenceSourceNode) Dispose() {
	e.sprite.Dispose()
}

func (e *essenceSourceNode) Destroy() {
	e.Dispose()
	e.EventDestroyed.Emit(e)
}

func (e *essenceSourceNode) updateShader() {
	if !e.stats.spritesheet {
		if e.percengage >= 0.85 {
			e.sprite.Shader.Enabled = false
			return
		}
		e.sprite.Shader.Enabled = true
		e.sprite.Shader.SetFloatValue("Time", e.percengage+0.15)
		return
	}
	if e.percengage < 0.01 {
		e.sprite.FrameOffset.X = e.sprite.ImageWidth() - e.sprite.FrameWidth
		return
	}
	frameWidth := int(e.sprite.FrameWidth)
	frameIndex := int(e.sprite.ImageWidth()*(1.0-e.percengage)) / frameWidth
	e.sprite.FrameOffset.X = float64(frameIndex * frameWidth)
}
