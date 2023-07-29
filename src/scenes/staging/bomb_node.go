package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type bombNode struct {
	owner    *colonyAgentNode
	world    *worldState
	sprite   *ge.Sprite
	pos      gmath.Vec
	height   float64
	rotation gmath.Rad
}

func newBombNode(owner *colonyAgentNode, world *worldState) *bombNode {
	return &bombNode{
		owner: owner,
		pos:   owner.pos,
		world: world,
	}
}

func (b *bombNode) Init(scene *ge.Scene) {
	b.rotation = math.Pi / 2

	b.sprite = scene.NewSprite(assets.ImageBomb)
	b.sprite.Rotation = &b.rotation
	b.sprite.Pos.Base = &b.pos
	b.world.stage.AddGraphics(b.sprite)

	b.height = agentFlightHeight
}

func (b *bombNode) IsDisposed() bool {
	return b.sprite.IsDisposed()
}

func (b *bombNode) explode() {
	createEffect(b.world, effectConfig{
		Pos:   b.pos,
		Image: assets.ImageBombExplosion,
		Layer: slightlyAboveEffectLayer,
	})
	playExplosionSound(b.world, b.pos)

	// Bombs deal some extra damage to the dreadnought
	// and a lot of extra damage to buildings.
	const bombMaxDamage = 30.0
	const bombMaxBossDamage = 35.0
	const bombMaxBuildingDamage = 50.0
	const maxRadius = 56
	const maxRadiusSqr = maxRadius * maxRadius
	b.world.WalkCreeps(b.pos, 40, func(creep *creepNode) bool {
		distSqr := b.pos.DistanceSquaredTo(creep.pos)
		if distSqr <= maxRadiusSqr {
			// 40 => 0.5
			// 20 => 0.75
			// 0  => 1.0
			damageMultiplier := 1.0 - ((distSqr * 0.5) / maxRadiusSqr)
			baseDamage := bombMaxDamage
			if creep.stats.Kind == gamedata.CreepUberBoss {
				baseDamage = bombMaxBossDamage
			} else if creep.stats.Building {
				baseDamage = bombMaxBuildingDamage
			}
			creep.OnDamage(gamedata.DamageValue{Health: baseDamage * damageMultiplier}, b.owner)
		}
		return false
	})
}

func (b *bombNode) dispose() {
	b.sprite.Dispose()
}

func (b *bombNode) Update(delta float64) {
	travelled := delta * 200
	b.height -= travelled
	b.pos.Y += travelled
	if b.height <= 0 {
		b.dispose()
		b.explode()
		return
	}
}
