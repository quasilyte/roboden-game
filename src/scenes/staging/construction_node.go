package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type constructionKind int

const (
	constructBase constructionKind = iota
	constructGunpoint
)

type constructionStats struct {
	ConstructionSpeed float64
	DamageModifier    float64
	Kind              constructionKind
	Image             resource.ImageID
}

var colonyCoreConstructionStats = &constructionStats{
	ConstructionSpeed: 0.01,
	DamageModifier:    0.01,
	Kind:              constructBase,
	Image:             assets.ImageColonyCore,
}

var gunpointConstructionStats = &constructionStats{
	ConstructionSpeed: 0.04,
	DamageModifier:    0.03,
	Kind:              constructGunpoint,
	Image:             assets.ImageGunpointAgent,
}

type constructionNode struct {
	pos              gmath.Vec
	constructPosBase gmath.Vec

	stats *constructionStats

	world *worldState

	progress float64

	attention float64

	maxBuildHeight     float64
	initialBuildHeight float64

	sprite *ge.Sprite

	EventDestroyed gsignal.Event[*constructionNode]
}

func newConstructionNode(world *worldState, pos gmath.Vec, stats *constructionStats) *constructionNode {
	return &constructionNode{
		world: world,
		pos:   pos,
		stats: stats,
	}
}

func (c *constructionNode) Init(scene *ge.Scene) {
	c.sprite = scene.NewSprite(c.stats.Image)
	c.sprite.Pos.Base = &c.pos
	switch c.stats.Kind {
	case constructBase:
		c.sprite.Shader = scene.NewShader(assets.ShaderColonyBuild)
	case constructGunpoint:
		c.sprite.Shader = scene.NewShader(assets.ShaderTurretBuild)
	}
	c.world.camera.AddSpriteBelow(c.sprite)

	c.maxBuildHeight = c.sprite.ImageHeight() * 0.9
	c.initialBuildHeight = c.sprite.ImageHeight() * 0.45
}

func (c *constructionNode) IsDisposed() bool {
	return c.sprite.IsDisposed()
}

func (c *constructionNode) Update(delta float64) {
	c.constructPosBase = c.pos.Add(gmath.Vec{
		Y: (c.maxBuildHeight * (1.0 - c.progress)) - c.initialBuildHeight,
	})
	c.attention = gmath.ClampMin(c.attention-delta, 0)
}

func (c *constructionNode) GetConstructPos() ge.Pos {
	xdelta := c.sprite.ImageWidth() * 0.3
	return ge.Pos{
		Base:   &c.constructPosBase,
		Offset: gmath.Vec{X: c.world.rand.FloatRange(-xdelta, xdelta)},
	}
}

func (c *constructionNode) Destroy() {
	c.EventDestroyed.Emit(c)
	c.Dispose()
}

func (c *constructionNode) IsFlying() bool { return false }

func (c *constructionNode) OnDamage(damage gamedata.DamageValue, source targetable) {
	c.progress -= damage.Health * c.stats.DamageModifier
	xdelta := c.sprite.ImageWidth() * 0.3
	if c.progress < 0 {
		rect := gmath.Rect{
			Min: c.constructPosBase.Sub(gmath.Vec{X: xdelta, Y: 8}),
			Max: c.constructPosBase.Add(gmath.Vec{X: xdelta, Y: 8}),
		}
		createAreaExplosion(c.world, rect, true)
		c.Destroy()
		return
	}
	explosionOffset := c.world.rand.FloatRange(-xdelta, xdelta)
	explosionPos := c.constructPosBase.Add(gmath.Vec{X: explosionOffset, Y: c.world.rand.FloatRange(0, 4)})
	createExplosion(c.world, false, explosionPos)
	c.sprite.Shader.SetFloatValue("Time", c.progress)
}

func (c *constructionNode) Construct(v float64, builder *colonyCoreNode) bool {
	c.progress += v * c.stats.ConstructionSpeed
	if c.progress >= 1 {
		c.done(builder)
		return true
	}
	c.sprite.Shader.SetFloatValue("Time", c.progress)
	return false
}

func (c *constructionNode) GetVelocity() gmath.Vec { return gmath.Vec{} }

func (c *constructionNode) GetPos() *gmath.Vec { return &c.constructPosBase }

func (c *constructionNode) Dispose() {
	c.sprite.Dispose()
}

func (c *constructionNode) done(builder *colonyCoreNode) {
	c.Dispose()
	c.EventDestroyed.Emit(c)

	switch c.stats.Kind {
	case constructGunpoint:
		turret := newColonyAgentNode(builder, gamedata.GunpointAgentStats, c.pos)
		builder.AcceptTurret(turret)
		c.world.nodeRunner.AddObject(turret)
		turret.mode = agentModeGuardForever

	case constructBase:
		core := c.world.NewColonyCoreNode(colonyConfig{
			World:  c.world,
			Radius: 128,
			Pos:    c.pos,
		})
		core.resources = 40
		core.priorities.SetWeight(priorityResources, 0.4)
		core.priorities.SetWeight(priorityGrowth, 0.4)
		core.priorities.SetWeight(prioritySecurity, 0.2)
		c.world.nodeRunner.AddObject(core)
	}
}
