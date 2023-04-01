package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/viewport"
)

type projectileNode struct {
	attacker  projectileTarget
	pos       gmath.Vec
	toPos     gmath.Vec
	target    projectileTarget
	fireDelay float64
	weapon    *gamedata.WeaponStats

	rotation gmath.Rad

	arcProgressionScaling float64
	arcProgression        float64
	arcStart              gmath.Vec
	arcFrom               gmath.Vec
	arcTo                 gmath.Vec

	scene *ge.Scene

	camera *viewport.Camera
	sprite *ge.Sprite
}

type projectileTarget interface {
	GetPos() *gmath.Vec
	GetVelocity() gmath.Vec
	OnDamage(damage gamedata.DamageValue, source gmath.Vec)
	IsDisposed() bool
	IsFlying() bool
}

func initWeaponStats(stats *gamedata.WeaponStats) *gamedata.WeaponStats {
	stats.ImpactAreaSqr = stats.ImpactArea * stats.ImpactArea
	return stats
}

type projectileConfig struct {
	Weapon    *gamedata.WeaponStats
	Camera    *viewport.Camera
	Attacker  projectileTarget
	ToPos     gmath.Vec
	Target    projectileTarget
	FireDelay float64
}

func newProjectileNode(config projectileConfig) *projectileNode {
	p := &projectileNode{
		camera:    config.Camera,
		weapon:    config.Weapon,
		attacker:  config.Attacker,
		pos:       config.Attacker.GetPos().Add(config.Weapon.FireOffset),
		toPos:     config.ToPos,
		target:    config.Target,
		fireDelay: config.FireDelay,
	}
	if p.weapon.ArcPower != 0 {
		dist := p.pos.DistanceTo(p.toPos)
		t := dist / p.weapon.ProjectileSpeed
		p.arcProgressionScaling = 1.0 / t
		power := gmath.Vec{Y: dist * p.weapon.ArcPower}
		p.arcFrom = p.pos.Add(power)
		p.arcTo = p.toPos.Add(power)
		p.arcStart = p.pos
		p.rotation = -math.Pi / 2
	}
	return p
}

func (p *projectileNode) Init(scene *ge.Scene) {
	p.scene = scene
	if p.weapon.ProjectileRotateSpeed == 0 {
		p.rotation = p.pos.AngleToPoint(p.toPos)
	} else {
		p.rotation = scene.Rand().Rad()
	}

	p.sprite = scene.NewSprite(p.weapon.ProjectileImage)
	p.sprite.Pos.Base = &p.pos
	p.sprite.Rotation = &p.rotation
	if p.fireDelay > 0 {
		p.sprite.Visible = false
	}
	p.camera.AddSpriteAbove(p.sprite)

	if p.arcProgressionScaling != 0 {
		missChance := 0.1
		if p.target.IsFlying() {
			missChance = 0.4
		}
		if scene.Rand().Chance(missChance) {
			// Most likely will be a miss.
			p.toPos = p.toPos.Add(scene.Rand().Offset(-28, 28))
		}
	}

	if p.fireDelay == 0 && p.weapon.ProjectileFireSound {
		p.playFireSound()
	}
}

func (p *projectileNode) IsDisposed() bool { return p.sprite.IsDisposed() }

func (p *projectileNode) playFireSound() {
	playSound(p.scene, p.camera, p.weapon.AttackSound, p.pos)
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
}

func (p *projectileNode) Dispose() {
	p.sprite.Dispose()
}

func (p *projectileNode) detonate() {
	p.Dispose()
	if p.target.IsDisposed() {
		return
	}
	if p.toPos.DistanceSquaredTo(*p.target.GetPos()) > p.weapon.ImpactAreaSqr {
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

	explosionKind := p.weapon.Explosion
	if explosionKind == gamedata.ProjectileExplosionNone {
		return
	}
	explosionPos := p.pos.Add(p.scene.Rand().Offset(-4, 4))
	switch explosionKind {
	case gamedata.ProjectileExplosionNormal:
		createExplosion(p.scene, p.camera, p.target.IsFlying(), explosionPos)
	case gamedata.ProjectileExplosionCripplerBlaster:
		effect := newEffectNode(p.camera, explosionPos, p.target.IsFlying(), assets.ImageCripplerBlasterExplosion)
		p.scene.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionMilitiaIon:
		p.scene.AddObject(newEffectNode(p.camera, explosionPos, p.target.IsFlying(), assets.ImageMilitiaIonExplosion))
	case gamedata.ProjectileExplosionShocker:
		p.scene.AddObject(newEffectNode(p.camera, explosionPos, p.target.IsFlying(), assets.ImageShockerExplosion))
	case gamedata.ProjectilePurpleExplosion:
		soundIndex := p.scene.Rand().IntRange(0, 2)
		sound := assets.AudioPurpleExplosion1 + resource.AudioID(soundIndex)
		p.scene.AddObject(newEffectNode(p.camera, explosionPos, p.target.IsFlying(), assets.ImagePurpleExplosion))
		playSound(p.scene, p.camera, sound, explosionPos)
	}
}
