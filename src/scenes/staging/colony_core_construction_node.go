package staging

import (
	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
)

type colonyCoreConstructionNode struct {
	pos          gmath.Vec
	constructPos gmath.Vec

	world *worldState
	scene *ge.Scene

	progress float64

	attention float64

	sprite *ge.Sprite

	EventDestroyed gsignal.Event[*colonyCoreConstructionNode]
}

func newColonyCoreConstructionNode(world *worldState, pos gmath.Vec) *colonyCoreConstructionNode {
	return &colonyCoreConstructionNode{
		world: world,
		pos:   pos,
	}
}

func (c *colonyCoreConstructionNode) Init(scene *ge.Scene) {
	c.scene = scene

	c.sprite = scene.NewSprite(assets.ImageColonyCore)
	c.sprite.Pos.Base = &c.pos
	c.sprite.Shader = scene.NewShader(assets.ShaderColonyBuild)
	c.world.camera.AddGraphicsBelow(c.sprite)
}

func (c *colonyCoreConstructionNode) IsDisposed() bool {
	return c.sprite.IsDisposed()
}

func (c *colonyCoreConstructionNode) Update(delta float64) {
	c.constructPos = c.pos.Add(gmath.Vec{
		Y: (56.0 * (1.0 - c.progress)) - 30,
	})
	c.attention = gmath.ClampMin(c.attention-delta, 0)
}

func (c *colonyCoreConstructionNode) Destroy() {
	c.EventDestroyed.Emit(c)
	c.Dispose()
}

func (c *colonyCoreConstructionNode) OnDamage(damage damageValue, source gmath.Vec) {
	c.progress -= damage.health * 0.02
	if c.progress < 0 {
		rect := gmath.Rect{
			Min: c.constructPos.Sub(gmath.Vec{X: 20, Y: 8}),
			Max: c.constructPos.Add(gmath.Vec{X: 20, Y: 8}),
		}
		createAreaExplosion(c.scene, c.world.camera, rect)
		c.Destroy()
		return
	}
	explosionOffset := c.scene.Rand().FloatRange(-22, 22)
	explosionPos := c.constructPos.Add(gmath.Vec{X: explosionOffset, Y: c.scene.Rand().FloatRange(0, 4)})
	createExplosion(c.scene, c.world.camera, explosionPos)
	c.sprite.Shader.SetFloatValue("Time", c.progress)
}

func (c *colonyCoreConstructionNode) Construct(v float64) *colonyCoreNode {
	c.progress += v * 0.01
	if c.progress >= 1 {
		return c.done()
	}
	c.sprite.Shader.SetFloatValue("Time", c.progress)
	return nil
}

func (c *colonyCoreConstructionNode) GetVelocity() gmath.Vec { return gmath.Vec{} }

func (c *colonyCoreConstructionNode) GetPos() *gmath.Vec { return &c.constructPos }

func (c *colonyCoreConstructionNode) Dispose() {
	c.sprite.Dispose()
}

func (c *colonyCoreConstructionNode) done() *colonyCoreNode {
	c.Dispose()
	core := c.world.NewColonyCoreNode(colonyConfig{
		World:  c.world,
		Radius: 96,
		Pos:    c.pos,
	})
	core.resources.Essence = 20
	core.actionPriorities.SetWeight(priorityResources, 0.3)
	core.actionPriorities.SetWeight(priorityGrowth, 0.50)
	core.actionPriorities.SetWeight(priorityEvolution, 0.1)
	core.actionPriorities.SetWeight(prioritySecurity, 0.1)
	c.scene.AddObject(core)
	return core
}
