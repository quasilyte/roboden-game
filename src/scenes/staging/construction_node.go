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
	constructTurret
)

type constructionStats struct {
	ConstructionSpeed float64
	DamageModifier    float64
	Kind              constructionKind
	TurretStats       *gamedata.AgentStats
	Image             resource.ImageID
}

var colonyCoreConstructionStats = &constructionStats{
	ConstructionSpeed: 0.01,
	DamageModifier:    0.01,
	Kind:              constructBase,
}

var harvesterConstructionStats = &constructionStats{
	ConstructionSpeed: 0.03,
	DamageModifier:    0.05,
	Kind:              constructTurret,
	TurretStats:       gamedata.HarvesterAgentStats,
	Image:             assets.ImageHarvesterAgent,
}

var gunpointConstructionStats = &constructionStats{
	ConstructionSpeed: 0.04,
	DamageModifier:    0.03,
	Kind:              constructTurret,
	TurretStats:       gamedata.GunpointAgentStats,
	Image:             assets.ImageGunpointAgent,
}

var beamTowerConstructionStats = &constructionStats{
	ConstructionSpeed: 0.025,
	DamageModifier:    0.04,
	Kind:              constructTurret,
	TurretStats:       gamedata.BeamTowerAgentStats,
	Image:             assets.ImageBeamtowerAgent,
}

var tetherBeaconConstructionStats = &constructionStats{
	ConstructionSpeed: 0.04,
	DamageModifier:    0.03,
	Kind:              constructTurret,
	TurretStats:       gamedata.TetherBeaconAgentStats,
	Image:             assets.ImageTetherBeaconAgent,
}

type constructionNode struct {
	pos              gmath.Vec
	spriteOffset     gmath.Vec
	constructPosBase gmath.Vec

	stats *constructionStats

	player player

	world *worldState

	progress float64

	attention float64

	maxBuildHeight     float64
	initialBuildHeight float64

	sprite *ge.Sprite

	EventDestroyed gsignal.Event[*constructionNode]
}

func newConstructionNode(world *worldState, p player, pos, spriteOffset gmath.Vec, stats *constructionStats) *constructionNode {
	return &constructionNode{
		world:        world,
		pos:          pos,
		spriteOffset: spriteOffset,
		player:       p,
		stats:        stats,
	}
}

func (c *constructionNode) Init(scene *ge.Scene) {
	imageID := c.stats.Image
	if c.stats.Kind == constructBase {
		imageID = c.world.coreDesign.Image
	}
	c.sprite = scene.NewSprite(imageID)
	c.sprite.Pos.Base = &c.pos
	c.sprite.Pos.Offset = c.spriteOffset
	c.maxBuildHeight = c.sprite.ImageHeight() * 0.9
	c.initialBuildHeight = c.sprite.ImageHeight() * 0.45
	if !c.world.simulation {
		switch c.stats.Kind {
		case constructBase:
			if c.world.coreDesign == gamedata.TankCoreStats {
				c.sprite.Shader = scene.NewShader(assets.ShaderSmallColonyBuild)
				c.maxBuildHeight *= 0.55
				c.initialBuildHeight *= 0.7
			} else {
				c.sprite.Shader = scene.NewShader(assets.ShaderColonyBuild)
			}
		case constructTurret:
			c.sprite.Shader = scene.NewShader(assets.ShaderTurretBuild)
		}
		c.sprite.Shader.SetFloatValue("Time", 0)
	}
	c.world.stage.AddSpriteBelow(c.sprite)
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

func (c *constructionNode) GetTargetInfo() targetInfo {
	return targetInfo{building: true, flying: false}
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
		createAreaExplosion(c.world, rect, normalEffectLayer)
		c.Destroy()
		return
	}
	explosionOffset := c.world.rand.FloatRange(-xdelta, xdelta)
	explosionPos := c.constructPosBase.Add(gmath.Vec{X: explosionOffset, Y: c.world.rand.FloatRange(0, 4)})
	createExplosion(c.world, normalEffectLayer, explosionPos)
	c.sprite.Shader.SetFloatValue("Time", c.progress)
}

func (c *constructionNode) Construct(v float64, builder *colonyCoreNode) bool {
	c.progress += v * c.stats.ConstructionSpeed
	if c.progress >= 1 {
		c.done(builder)
		return true
	}
	if !c.sprite.Shader.IsNil() {
		c.sprite.Shader.SetFloatValue("Time", c.progress)
	}
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
	case constructTurret:
		turret := newColonyAgentNode(builder, c.stats.TurretStats, c.pos)
		builder.AcceptTurret(turret)
		c.world.nodeRunner.AddObject(turret)
		turret.sprite.Pos.Offset = c.spriteOffset

	case constructBase:
		c.world.result.ColoniesBuilt++
		core := c.world.NewColonyCoreNode(colonyConfig{
			World:  c.world,
			Radius: 128,
			Pos:    c.pos,
			Player: c.player,
		})
		core.resources = 40
		core.priorities.SetWeight(priorityResources, 0.4)
		core.priorities.SetWeight(priorityGrowth, 0.4)
		core.priorities.SetWeight(prioritySecurity, 0.2)
		c.world.nodeRunner.AddObject(core)
	}
}
