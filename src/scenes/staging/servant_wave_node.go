package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type servantWaveNode struct {
	pos   gmath.Vec
	anim  *ge.Animation
	owner *creepNode

	damageDelay float64
}

func newServantWaveNode(owner *creepNode) *servantWaveNode {
	return &servantWaveNode{
		pos:   owner.pos,
		owner: owner,
	}
}

func (e *servantWaveNode) Init(scene *ge.Scene) {
	img := assets.ImageServantWave
	if e.owner.super {
		img = assets.ImageSuperServantWave
	}
	s := scene.NewSprite(img)
	s.Pos.Base = &e.pos
	e.owner.world.stage.AddSpriteAbove(s)

	e.anim = ge.NewAnimation(s, -1)
	e.anim.SetAnimationSpan(0.03 * 6)

	e.damageDelay = 0.09
}

func (e *servantWaveNode) IsDisposed() bool {
	return e.anim.IsDisposed()
}

func (e *servantWaveNode) Dispose() {
	e.anim.Sprite().Dispose()
}

func (e *servantWaveNode) dealDamage() {
	// TODO: more efficient way to grab all units around the pos.
	maxRangeSqr := 84.0 * 84.0
	damage := gamedata.DamageValue{Health: 4, Slow: 2}
	damage.Flags |= gamedata.DmgflagNoFlash
	if e.owner.super {
		maxRangeSqr = 112.0 * 112.0
		damage.Slow++
	}
	for _, colony := range e.owner.world.allColonies {
		if colony.realRadius < 196 && colony.pos.DistanceSquaredTo(e.pos) > (maxRangeSqr*2) {
			continue
		}
		colony.agents.Each(func(a *colonyAgentNode) {
			distSqr := a.pos.DistanceSquaredTo(e.pos)
			if distSqr > maxRangeSqr {
				return
			}
			a.OnDamage(damage, e.owner)

			createEffect(e.owner.world, effectConfig{
				Pos:            a.pos,
				Layer:          aboveEffectLayer,
				Image:          assets.ImageServantShotExplosion,
				AnimationSpeed: animationSpeedFast,
			})
		})
	}
}

func (e *servantWaveNode) Update(delta float64) {
	if e.anim.Tick(delta) {
		if e.damageDelay > 0 {
			e.dealDamage()
		}
		e.Dispose()
		return
	}

	if e.damageDelay > 0 {
		e.damageDelay -= delta
		if e.damageDelay <= 0 {
			e.dealDamage()
		}
	}
}
