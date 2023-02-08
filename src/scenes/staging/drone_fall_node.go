package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type droneFallNode struct {
	world *worldState
	scene *ge.Scene

	image       resource.ImageID
	shadowImage resource.ImageID

	sprite *ge.Sprite
	shadow *ge.Sprite

	pos      gmath.Vec
	rotation gmath.Rad

	scraps *essenceSourceStats

	height float64
}

func newDroneFallNode(world *worldState, scraps *essenceSourceStats, image, shadow resource.ImageID, pos gmath.Vec, height float64) *droneFallNode {
	return &droneFallNode{
		scraps:      scraps,
		world:       world,
		image:       image,
		shadowImage: shadow,
		pos:         pos,
		height:      height,
	}
}

func (d *droneFallNode) Init(scene *ge.Scene) {
	d.scene = scene

	d.sprite = scene.NewSprite(d.image)
	d.sprite.Pos.Base = &d.pos
	d.sprite.Rotation = &d.rotation
	d.world.camera.AddGraphics(d.sprite)

	d.shadow = scene.NewSprite(d.shadowImage)
	d.shadow.Pos.Base = &d.pos
	d.world.camera.AddGraphicsBelow(d.shadow)

	d.height -= 4
	d.pos.Y += 4
}

func (d *droneFallNode) Destroy() {
	d.sprite.Dispose()
	d.shadow.Dispose()

	createExplosion(d.scene, d.world.camera, d.pos)

	essenceSpawnPos := d.pos.Add(gmath.Vec{Y: 6})
	if d.scraps != nil && posIsFree(d.world, nil, essenceSpawnPos, 48) {
		essence := d.world.NewEssenceSourceNode(d.scraps, essenceSpawnPos)
		d.scene.AddObject(essence)
	}
}

func (d *droneFallNode) IsDisposed() bool { return d.sprite.IsDisposed() }

func (d *droneFallNode) Update(delta float64) {
	const fallSpeed float64 = 60

	d.height -= delta * fallSpeed
	if d.height <= 0 {
		d.Destroy()
		return
	}

	d.pos.Y += delta * fallSpeed
	d.pos.X += d.scene.Rand().FloatRange(-6, 6) * delta

	d.rotation += gmath.Rad(delta * 2)

	d.shadow.Pos.Offset.Y = d.height + 4
	newShadowAlpha := float32(1.0 - ((d.height / agentFlightHeight) * 0.5))
	d.shadow.SetAlpha(newShadowAlpha)
}
