package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type lavaGeyserNode struct {
	sprite *ge.Sprite
	world  *worldState

	pos gmath.Vec

	beamNode       *beamNode
	lineHeight     float64
	lineDecayDelay float64

	fireDelay float64
}

func newLavaGeyserNode(world *worldState, pos gmath.Vec) *lavaGeyserNode {
	return &lavaGeyserNode{
		world: world,
		pos:   pos,
	}
}

func (n *lavaGeyserNode) Init(scene *ge.Scene) {
	n.fireDelay = scene.Rand().FloatRange(15, 30)

	n.sprite = scene.NewSprite(assets.ImageLavaGeyser)
	n.sprite.Pos.Base = &n.pos
	if n.world.localRand.Bool() {
		n.sprite.FlipHorizontal = true
	}
	n.world.stage.AddSprite(n.sprite)

	n.world.MarkPos(n.pos, ptagBlocked)
}

func (n *lavaGeyserNode) IsDisposed() bool {
	return false
}

func (n *lavaGeyserNode) createBurstEffect(pos gmath.Vec) {
	createEffect(n.world, effectConfig{
		Pos:   pos,
		Image: assets.ImageFireBurst,
		Layer: normalEffectLayer,
	})
}

func (n *lavaGeyserNode) dealDamage() {
	damageRect := gmath.Rect{
		Min: n.pos.Sub(gmath.Vec{X: 24, Y: 97}),
		Max: n.pos.Add(gmath.Vec{X: 24, Y: -6}),
	}

	damage := gamedata.DamageValue{
		Health: 25,
	}

	for _, colony := range n.world.allColonies {
		colony.agents.Each(func(a *colonyAgentNode) {
			if !damageRect.Contains(a.pos) {
				return
			}
			n.createBurstEffect(a.pos)
			a.OnDamage(damage, a)
		})
	}

	n.world.WalkCreepsWithRand(nil, n.pos, 128, func(creep *creepNode) bool {
		if !creep.IsFlying() || creep.stats == gamedata.UberBossCreepStats {
			return false
		}
		if !damageRect.Contains(creep.pos) {
			return false
		}
		n.createBurstEffect(creep.pos)
		creep.OnDamage(damage, creep)
		return false
	})
}

func (n *lavaGeyserNode) Update(delta float64) {
	if n.beamNode != nil {
		if n.beamNode.IsDisposed() {
			n.beamNode = nil
		} else {
			toPos := n.beamNode.GetToPos()
			prevDecay := n.lineDecayDelay
			n.lineDecayDelay -= delta
			if prevDecay > 0 && n.lineDecayDelay <= 0 && n.world.localRand.Chance(0.75) {
				// A burst with faster animation.
				createEffect(n.world, effectConfig{
					Pos:            n.pos.Sub(gmath.Vec{Y: 20}),
					Image:          assets.ImageFireBurst,
					Layer:          normalEffectLayer,
					AnimationSpeed: animationSpeedFast,
				})
			}
			if n.lineDecayDelay <= 0 && n.lineHeight >= 10 {
				n.lineHeight = gmath.ClampMin(n.lineHeight-delta*280, 10)
			} else {
				if n.lineHeight < 80 {
					n.lineHeight = gmath.ClampMax(n.lineHeight+310*delta, 80)
					if n.lineHeight == 80 {
						n.createBurstEffect(n.pos.Sub(gmath.Vec{Y: 75}))
						n.dealDamage()
					}
				}
			}
			toPos.Offset.Y = -n.lineHeight - 7
		}
	}

	n.fireDelay = gmath.ClampMin(n.fireDelay-delta, 0)
	if n.fireDelay != 0 {
		return
	}

	n.lineHeight = 30
	n.lineDecayDelay = 0.97
	from := ge.Pos{Base: &n.pos, Offset: gmath.Vec{Y: -7}}
	to := ge.Pos{Base: &n.pos, Offset: gmath.Vec{Y: -17}}
	beam := newTextureBeamNode(n.world, from, to, gamedata.LavaGeyserBeamTexture, 1.8, 0.95)
	n.world.nodeRunner.AddObject(beam)
	n.beamNode = beam

	n.createBurstEffect(n.pos.Sub(gmath.Vec{Y: 11}))

	n.fireDelay = n.world.rand.FloatRange(20, 40)
	playSound(n.world, assets.AudioLavaBurst1, n.pos)
}
