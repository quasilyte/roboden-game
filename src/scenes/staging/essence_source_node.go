package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type essenceSourceStats struct {
	name        string
	image       resource.ImageID
	capacity    gmath.Range[int]
	regenDelay  float64 // 0 for "no regen"
	value       float64 // Resource score per unit
	eliteValue  float64 // Elite resource score per unit
	spritesheet bool
	canRotate   bool
	passable    bool
	scrap       bool
	canDeplete  bool
	size        float64
}

var redCrystalSource = &essenceSourceStats{
	name:        "red_crystal",
	image:       assets.ImageEssenceRedCrystalSource,
	capacity:    gmath.MakeRange(1, 1),
	value:       20,
	eliteValue:  3,
	spritesheet: true,
	canDeplete:  true,
	size:        32,
}

var oilSource = &essenceSourceStats{
	name:        "oil",
	image:       assets.ImageEssenceSource,
	capacity:    gmath.MakeRange(50, 80),
	regenDelay:  7,
	value:       4, // 200-320 total
	spritesheet: true,
	canDeplete:  false,
	size:        32,
}

var redOilSource = &essenceSourceStats{
	name:        "red_oil",
	image:       assets.ImageRedEssenceSource,
	capacity:    gmath.MakeRange(60, 80),
	regenDelay:  9,
	value:       5, // 300-400 total
	eliteValue:  0.5,
	spritesheet: true,
	canDeplete:  false,
	size:        32,
}

var goldSource = &essenceSourceStats{
	name:        "gold",
	image:       assets.ImageEssenceGoldSource,
	capacity:    gmath.MakeRange(25, 40),
	regenDelay:  0, // none
	value:       6, // 150-240 total
	spritesheet: true,
	canDeplete:  true,
	size:        20,
}

var crystalSource = &essenceSourceStats{
	name:        "crystal",
	image:       assets.ImageEssenceCrystalSource,
	capacity:    gmath.MakeRange(15, 20),
	regenDelay:  0,  // none
	value:       16, // 240-320 total
	spritesheet: true,
	canDeplete:  true,
	size:        16,
}

var ironSource = &essenceSourceStats{
	name:        "iron",
	image:       assets.ImageEssenceIronSource,
	capacity:    gmath.MakeRange(105, 150),
	regenDelay:  0,   // none
	value:       1.5, // 160-220 total
	spritesheet: true,
	canDeplete:  true,
	size:        20,
}

var organicSource = &essenceSourceStats{
	name:        "organic",
	image:       assets.ImageOrganicSource,
	capacity:    gmath.MakeRange(20, 25),
	regenDelay:  0,   // none
	value:       4.0, // 80-100 total
	spritesheet: true,
	canDeplete:  false,
	passable:    true,
	size:        20,
}

var smallScrapSource = &essenceSourceStats{
	name:       "scrap",
	image:      assets.ImageEssenceSmallScrapSource,
	capacity:   gmath.MakeRange(4, 5),
	regenDelay: 0, // none
	value:      1, // 4-5
	size:       14,
	canDeplete: true,
	passable:   true,
	scrap:      true,
}

var scrapSource = &essenceSourceStats{
	name:       "scrap",
	image:      assets.ImageEssenceScrapSource,
	capacity:   gmath.MakeRange(8, 12),
	regenDelay: 0, // none
	value:      1, // 8-12
	size:       16,
	canDeplete: true,
	passable:   true,
	scrap:      true,
}

var smallScrapCreepSource = &essenceSourceStats{
	name:       "scrap",
	image:      assets.ImageEssenceSmallScrapCreepSource,
	capacity:   gmath.MakeRange(5, 7),
	regenDelay: 0, // none
	value:      2, // 10-14
	size:       14,
	canDeplete: true,
	passable:   true,
	scrap:      true,
}

var scrapCreepSource = &essenceSourceStats{
	name:       "scrap",
	image:      assets.ImageEssenceScrapCreepSource,
	capacity:   gmath.MakeRange(8, 14),
	regenDelay: 0, // none
	value:      2, // 16-28
	size:       16,
	canDeplete: true,
	passable:   true,
	scrap:      true,
}

var bigScrapCreepSource = &essenceSourceStats{
	name:       "scrap",
	image:      assets.ImageEssenceBigScrapCreepSource,
	capacity:   gmath.MakeRange(12, 20),
	regenDelay: 0, // none
	value:      2, // 24-40
	size:       20,
	canDeplete: true,
	passable:   true,
	scrap:      true,
}

type essenceSourceNode struct {
	scene *ge.Scene

	world *worldState

	stats *essenceSourceStats

	sprite *ge.Sprite

	capacity          int
	resource          int
	percengage        float64
	recoverDelay      float64 // a ticker before the next regen
	recoverDelayTimer float64 // how much time it takes to reach a regen tick
	beingHarvested    bool

	rotation gmath.Rad
	pos      gmath.Vec

	EventDestroyed gsignal.Event[*essenceSourceNode]
}

func newEssenceSourceNode(world *worldState, stats *essenceSourceStats, pos gmath.Vec) *essenceSourceNode {
	return &essenceSourceNode{
		world: world,
		stats: stats,
		pos:   pos,
	}
}

func (e *essenceSourceNode) Init(scene *ge.Scene) {
	e.scene = scene

	img := e.stats.image
	switch gamedata.EnvironmentKind(e.world.config.Environment) {
	case gamedata.EnvForest:
		if e.stats == oilSource {
			img++
		}
	}

	e.sprite = scene.NewSprite(img)
	e.sprite.Pos.Base = &e.pos
	e.sprite.Rotation = &e.rotation
	if !e.stats.spritesheet && e.world.graphicsSettings.AllShadersEnabled {
		e.sprite.Shader = scene.NewShader(assets.ShaderDissolve)
		e.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageEssenceSourceDissolveMask)
		e.sprite.Shader.Enabled = false
	}
	e.world.stage.AddSpriteBelow(e.sprite)

	if e.stats.canRotate {
		e.rotation = scene.Rand().Rad()
	} else {
		e.sprite.FlipHorizontal = scene.Rand().Bool()
	}

	e.capacity = scene.Rand().IntRange(e.stats.capacity.Min, e.stats.capacity.Max)
	if e.stats == ironSource && !e.world.config.GoldEnabled {
		// If gold is disabled, iron has doubled capacity.
		e.capacity *= 2
	}
	e.resource = e.capacity
	if e.stats == organicSource {
		e.resource = int(float64(e.resource) * scene.Rand().FloatRange(0.4, 0.9))
		e.percengage = float64(e.resource) / float64(e.capacity)
	} else {
		e.percengage = 1.0
	}
	e.updateShader()
}

func (e *essenceSourceNode) IsDisposed() bool { return e.sprite.IsDisposed() }

func (e *essenceSourceNode) Update(delta float64) {
	if e.recoverDelayTimer == 0 {
		return
	}
	e.recoverDelay -= delta
	if e.recoverDelay <= 0 {
		e.recoverDelay = e.recoverDelayTimer * e.scene.Rand().FloatRange(0.75, 1.25)
		e.resource = gmath.ClampMax(e.resource+1, e.capacity)
		e.percengage = float64(e.resource) / float64(e.capacity)
		e.updateShader()
	}
}

func (e *essenceSourceNode) Restore(n int) {
	e.resource = gmath.ClampMax(e.resource+n, e.capacity)
	e.percengage = float64(e.resource) / float64(e.capacity)
	e.updateShader()
}

func (e *essenceSourceNode) Harvest(n int) int {
	if e.IsDisposed() {
		return 0
	}

	n = gmath.ClampMax(n, e.resource)
	e.resource -= n
	e.percengage = float64(e.resource) / float64(e.capacity)

	if e.resource <= 0 && e.stats.canDeplete {
		e.Destroy()
		if e.stats == redCrystalSource {
			e.world.result.RedCrystalsCollected++
		}
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
		if !e.sprite.Shader.IsNil() {
			if e.percengage >= 0.85 {
				e.sprite.Shader.Enabled = false
				return
			}
			e.sprite.Shader.Enabled = true
			e.sprite.Shader.SetFloatValue("Time", e.percengage+0.15)
		}
		return
	}

	if e.stats == organicSource {
		e.sprite.Visible = e.percengage > 0
	}
	if e.percengage < 0.01 {
		e.sprite.FrameOffset.X = e.sprite.ImageWidth() - e.sprite.FrameWidth
		return
	}
	frameWidth := int(e.sprite.FrameWidth)
	frameIndex := int(e.sprite.ImageWidth()*(1.0-e.percengage)) / frameWidth
	e.sprite.FrameOffset.X = float64(frameIndex * frameWidth)
}
