package staging

import (
	"fmt"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type servantWaveNode struct {
	pos   gmath.Vec
	anim  *ge.Animation
	world *worldState

	damageDelay float64
}

func newServantWaveNode(world *worldState, pos gmath.Vec) *servantWaveNode {
	return &servantWaveNode{
		pos:   pos,
		world: world,
	}
}

func (e *servantWaveNode) Init(scene *ge.Scene) {
	s := scene.NewSprite(assets.ImageServantWave)
	s.Pos.Base = &e.pos
	e.world.camera.AddSpriteAbove(s)

	e.anim = ge.NewAnimation(s, -1)
	e.anim.SetSecondsPerFrame(0.03)

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
	const maxRangeSqr float64 = 96 * 96
	for _, colony := range e.world.colonies {
		if colony.realRadius < 196 && colony.pos.DistanceSquaredTo(e.pos) > (maxRangeSqr*2) {
			continue
		}
		colony.agents.Each(func(a *colonyAgentNode) {
			distSqr := a.pos.DistanceSquaredTo(e.pos)
			if distSqr > maxRangeSqr {
				return
			}
			a.OnDamage(gamedata.DamageValue{Health: 4, Slow: 2}, e.pos)
		})
	}
}

func (e *servantWaveNode) Update(delta float64) {
	if e.anim.Tick(delta) {
		if e.damageDelay > 0 {
			fmt.Println("warning: servant: dealing damage after animation is over")
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
