package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type projectileNode struct {
	attacker  projectileTarget
	pos       gmath.Vec
	toPos     gmath.Vec
	target    projectileTarget
	fireDelay float64
	weapon    *gamedata.WeaponStats
	world     *worldState

	rotation gmath.Rad

	arcProgressionScaling float64
	arcProgression        float64
	arcStart              gmath.Vec
	arcFrom               gmath.Vec
	arcTo                 gmath.Vec

	sprite *ge.Sprite
}

type projectileTarget interface {
	GetPos() *gmath.Vec
	GetVelocity() gmath.Vec
	OnDamage(damage gamedata.DamageValue, source gmath.Vec)
	IsDisposed() bool
	IsFlying() bool
}

type projectileConfig struct {
	Weapon     *gamedata.WeaponStats
	World      *worldState
	Attacker   projectileTarget
	ToPos      gmath.Vec
	Target     projectileTarget
	FireDelay  float64
	FireOffset gmath.Vec
}

func newProjectileNode(config projectileConfig) *projectileNode {
	p := &projectileNode{
		weapon:    config.Weapon,
		attacker:  config.Attacker,
		pos:       config.Attacker.GetPos().Add(config.Weapon.FireOffset).Add(config.FireOffset),
		toPos:     config.ToPos,
		target:    config.Target,
		fireDelay: config.FireDelay,
		world:     config.World,
	}
	if p.weapon.ArcPower != 0 {
		arcPower := p.weapon.ArcPower
		speed := p.weapon.ProjectileSpeed
		if config.ToPos.Y >= p.pos.Y {
			arcPower *= 0.3
			speed *= 1.5
		}
		dist := p.pos.DistanceTo(p.toPos)
		t := dist / speed
		p.arcProgressionScaling = 1.0 / t
		power := gmath.Vec{Y: dist * arcPower}
		p.arcFrom = p.pos.Add(power)
		p.arcTo = p.toPos.Add(power)
		p.arcStart = p.pos
		p.rotation = -math.Pi / 2
	}
	return p
}

func (p *projectileNode) Init(scene *ge.Scene) {
	if p.weapon.ProjectileRotateSpeed == 0 {
		p.rotation = p.pos.AngleToPoint(p.toPos)
	} else {
		p.rotation = scene.Rand().Rad()
	}

	p.sprite = scene.NewSprite(p.weapon.ProjectileImage)
	p.sprite.Pos.Base = &p.pos
	p.sprite.Rotation = &p.rotation
	p.world.camera.AddSpriteAbove(p.sprite)

	p.sprite.Visible = false

	if p.weapon.Accuracy != 1.0 {
		missChance := 1.0 - p.weapon.Accuracy
		if missChance != 0 && scene.Rand().Chance(missChance) {
			dist := p.pos.DistanceTo(p.toPos)
			// 100 => 25
			// 200 => 50
			// 400 => 100
			offsetValue := gmath.Clamp(dist*0.25, 24, 140)
			p.toPos = p.toPos.Add(scene.Rand().Offset(-offsetValue, offsetValue))
		} else if p.arcProgressionScaling != 0 {
			p.toPos = p.toPos.Add(scene.Rand().Offset(-8, 8))
		}
	}

	if p.fireDelay == 0 && p.weapon.ProjectileFireSound {
		p.playFireSound()
	}
}

func (p *projectileNode) IsDisposed() bool { return p.sprite.IsDisposed() }

func (p *projectileNode) playFireSound() {
	playSound(p.world, p.weapon.AttackSound, p.pos)
}

func (p *projectileNode) Update(delta float64) {
	if p.fireDelay > 0 {
		if p.attacker.IsDisposed() {
			p.Dispose()
			return
		}
		p.fireDelay -= delta
		if p.fireDelay <= 0 {
			p.sprite.Visible = true
			p.pos = p.attacker.GetPos().Add(p.weapon.FireOffset)
			p.arcStart = p.pos
			if p.weapon.ProjectileFireSound {
				p.playFireSound()
			}
		} else {
			return
		}
	}

	travelled := p.weapon.ProjectileSpeed * delta

	if p.arcProgressionScaling == 0 {
		if p.pos.DistanceTo(p.toPos) <= travelled {
			p.detonate()
			return
		}
		p.pos = p.pos.MoveTowards(p.toPos, travelled)
		if p.weapon.ProjectileRotateSpeed != 0 {
			p.rotation += gmath.Rad(delta * p.weapon.ProjectileRotateSpeed)
		}
		p.sprite.Visible = true
		return
	}

	p.arcProgression += delta * p.arcProgressionScaling
	if p.arcProgression >= 1 {
		p.detonate()
		return
	}
	newPos := p.arcStart.CubicInterpolate(p.arcFrom, p.toPos, p.arcTo, p.arcProgression)
	if !p.weapon.RoundProjectile {
		p.rotation = p.pos.AngleToPoint(newPos)
	}
	p.pos = newPos
	p.sprite.Visible = true
}

func (p *projectileNode) Dispose() {
	p.sprite.Dispose()
}

func (p *projectileNode) createExplosion() {
	explosionKind := p.weapon.Explosion
	if explosionKind == gamedata.ProjectileExplosionNone {
		return
	}
	explosionPos := p.pos.Add(p.world.rand.Offset(-4, 4))
	switch explosionKind {
	case gamedata.ProjectileExplosionNormal:
		createExplosion(p.world, p.target.IsFlying(), explosionPos)
	case gamedata.ProjectileExplosionBigVertical:
		createBigVerticalExplosion(p.world, explosionPos)
	case gamedata.ProjectileExplosionCripplerBlaster:
		effect := newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImageCripplerBlasterExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionMilitiaIon:
		p.world.nodeRunner.AddObject(newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImageMilitiaIonExplosion))
	case gamedata.ProjectileExplosionShocker:
		p.world.nodeRunner.AddObject(newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImageShockerExplosion))
	case gamedata.ProjectileExplosionStealthLaser:
		p.world.nodeRunner.AddObject(newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImageStealthLaserExplosion))
	case gamedata.ProjectileExplosionFighterLaser:
		effect := newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImageFighterLaserExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionHeavyCrawlerLaser:
		effect := newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImageHeavyCrawlerLaserExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionPurple:
		soundIndex := p.world.rand.IntRange(0, 2)
		sound := assets.AudioPurpleExplosion1 + resource.AudioID(soundIndex)
		p.world.nodeRunner.AddObject(newEffectNode(p.world.camera, explosionPos, p.target.IsFlying(), assets.ImagePurpleExplosion))
		playSound(p.world, sound, explosionPos)
	}
}

func (p *projectileNode) detonate() {
	p.Dispose()
	if p.target.IsDisposed() {
		return
	}
	if p.toPos.DistanceSquaredTo(*p.target.GetPos()) > p.weapon.ImpactAreaSqr {
		if p.weapon.AlwaysExplodes {
			p.createExplosion()
		}
		return
	}

	dmg := p.weapon.Damage
	if dmg.Health != 0 {
		var multiplier float64
		if p.target.IsFlying() {
			multiplier = p.weapon.FlyingTargetDamageMult
		} else {
			multiplier = p.weapon.GroundTargetDamageMult
		}
		dmg.Health *= multiplier
	}
	p.target.OnDamage(p.weapon.Damage, *p.attacker.GetPos())
	p.createExplosion()
}
