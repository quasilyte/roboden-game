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
	attacker  targetable
	pos       gmath.Vec
	toPos     gmath.Vec
	target    targetable
	fireDelay float64
	weapon    *gamedata.WeaponStats
	world     *worldState

	trailCounter float64

	rotation gmath.Rad

	arcProgressionScaling float64
	arcProgression        float64
	arcStart              gmath.Vec
	arcFrom               gmath.Vec
	arcTo                 gmath.Vec

	guided   bool
	disposed bool
	sprite   *ge.Sprite
}

type targetable interface {
	GetPos() *gmath.Vec
	GetVelocity() gmath.Vec
	OnDamage(damage gamedata.DamageValue, source targetable)
	IsDisposed() bool
	IsFlying() bool
}

type projectileConfig struct {
	Weapon     *gamedata.WeaponStats
	World      *worldState
	Attacker   targetable
	ToPos      gmath.Vec
	Target     targetable
	FireDelay  float64
	FireOffset gmath.Vec
	Guided     bool
}

func initProjectileNode(p *projectileNode, config projectileConfig) {
	*p = projectileNode{
		weapon:    config.Weapon,
		attacker:  config.Attacker,
		pos:       config.Attacker.GetPos().Add(config.Weapon.FireOffset).Add(config.FireOffset),
		toPos:     config.ToPos,
		target:    config.Target,
		fireDelay: config.FireDelay,
		world:     config.World,
		guided:    config.Guided,
	}
}

func (p *projectileNode) Init(scene *ge.Scene) {
	if p.weapon.ArcPower != 0 {
		inversed := p.weapon.RandArc && scene.Rand().Bool()

		arcPower := p.weapon.ArcPower
		if p.weapon.RandArc {
			arcPower *= scene.Rand().FloatRange(0.9, 1.5)
		}

		speed := p.weapon.ProjectileSpeed
		if p.weapon.RandArc {
			p.rotation = p.pos.AngleToPoint(p.toPos)
			p.rotation += gmath.Rad(scene.Rand().FloatRange(-0.5, 0.5))
			p.pos = p.pos.MoveInDirection(14, p.rotation)
			speed *= scene.Rand().FloatRange(0.85, 1.25)
		} else {
			p.rotation = -math.Pi / 2
		}

		if inversed {
			if p.toPos.Y <= p.pos.Y {
				arcPower *= 0.3
				speed *= 1.5
			}
		} else {
			if p.toPos.Y >= p.pos.Y {
				arcPower *= 0.3
				speed *= 1.5
			}
		}
		dist := p.pos.DistanceTo(p.toPos)
		t := dist / speed
		p.arcProgressionScaling = 1.0 / t
		power := gmath.Vec{Y: dist * arcPower}
		if inversed {
			p.arcFrom = p.pos.Sub(power)
			p.arcTo = p.toPos.Sub(power)
		} else {
			p.arcFrom = p.pos.Add(power)
			p.arcTo = p.toPos.Add(power)
		}
		p.arcStart = p.pos
	} else if p.weapon.ProjectileRotateSpeed == 0 {
		p.rotation = p.pos.AngleToPoint(p.toPos)
	} else {
		p.rotation = scene.Rand().Rad()
	}

	if !p.world.simulation {
		p.sprite = scene.NewSprite(p.weapon.ProjectileImage)
		p.sprite.Pos.Base = &p.pos
		p.sprite.Rotation = &p.rotation
		p.world.stage.AddSpriteAbove(p.sprite)
		p.sprite.Visible = false
	}

	if p.weapon.Accuracy != 1.0 && !p.guided {
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

func (p *projectileNode) IsDisposed() bool { return p.disposed }

func (p *projectileNode) playFireSound() {
	playSound(p.world, p.weapon.AttackSound, p.pos)
}

func (p *projectileNode) setSpriteVisibility(visible bool) {
	if p.sprite == nil {
		return
	}
	p.sprite.Visible = visible
}

func (p *projectileNode) Update(delta float64) {
	if p.fireDelay > 0 {
		if p.attacker.IsDisposed() {
			p.Dispose()
			return
		}
		p.fireDelay -= delta
		if p.fireDelay <= 0 {
			p.setSpriteVisibility(true)
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

	if !p.world.simulation && p.weapon.TrailEffect != gamedata.ProjectileTrailNone {
		p.trailCounter -= delta
		switch p.weapon.TrailEffect {
		case gamedata.ProjectileTrailSmoke:
			if p.trailCounter <= 0 {
				p.trailCounter = p.world.localRand.FloatRange(0.1, 0.3)
				p.world.nodeRunner.AddObject(newEffectNode(p.world, p.pos, true, assets.ImageProjectileSmoke))
			}
		case gamedata.ProjectileTrailRoomba:
			if p.trailCounter <= 0 {
				p.trailCounter = p.world.localRand.FloatRange(0.1, 0.2)
				effect := newEffectNode(p.world, p.pos, true, assets.ImageRoombaLaserTrail)
				effect.rotation = p.rotation
				p.world.nodeRunner.AddObject(effect)
			}
		case gamedata.ProjectileTrailFire:
			if p.trailCounter <= 0 {
				p.trailCounter = p.world.localRand.FloatRange(0.06, 0.1)
				effect := newEffectNode(p.world, p.pos, true, assets.ImageFireTrail)
				effect.rotation = p.rotation
				p.world.nodeRunner.AddObject(effect)
			}
		}
	}

	if p.arcProgressionScaling == 0 {
		if p.pos.DistanceTo(p.toPos) <= travelled {
			p.detonate()
			return
		}
		p.pos = p.pos.MoveTowards(p.toPos, travelled)
		if p.weapon.ProjectileRotateSpeed != 0 {
			p.rotation += gmath.Rad(delta * p.weapon.ProjectileRotateSpeed)
		}
		p.setSpriteVisibility(true)
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
	p.setSpriteVisibility(true)
}

func (p *projectileNode) Dispose() {
	if p.sprite != nil {
		p.sprite.Dispose()
	}
	p.disposed = true
}

func (p *projectileNode) createExplosion() {
	if p.world.simulation {
		return
	}

	explosionKind := p.weapon.Explosion
	if explosionKind == gamedata.ProjectileExplosionNone {
		return
	}

	above := p.target.IsFlying()
	if !above {
		target, ok := p.target.(*creepNode)
		if ok && target.stats.Kind == gamedata.CreepUberBoss {
			above = true
		}
	}

	explosionPos := p.pos.Add(p.world.localRand.Offset(-4, 4))
	switch explosionKind {
	case gamedata.ProjectileExplosionNormal:
		createExplosion(p.world, above, explosionPos)
	case gamedata.ProjectileExplosionBigVertical:
		createBigVerticalExplosion(p.world, explosionPos)
	case gamedata.ProjectileExplosionAbomb:
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, false, assets.ImageNuclearExplosion))
		playSound(p.world, assets.AudioAbombExplosion, explosionPos)
	case gamedata.ProjectileExplosionCripplerBlaster:
		effect := newEffectNode(p.world, explosionPos, above, assets.ImageCripplerBlasterExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionGreenZap:
		effect := newEffectNode(p.world, explosionPos, above, assets.ImageGreenZap)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionScoutIon:
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, above, assets.ImageScoutIonExplosion))
	case gamedata.ProjectileExplosionRoombaShot:
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, above, assets.ImageRoombaShotExplosion))
	case gamedata.ProjectileExplosionScarab:
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, above, assets.ImageScarabShotExplosion))
	case gamedata.ProjectileExplosionShocker:
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, above, assets.ImageShockerExplosion))
	case gamedata.ProjectileExplosionStealthLaser:
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, above, assets.ImageStealthLaserExplosion))
	case gamedata.ProjectileExplosionCommanderLaser:
		effect := newEffectNode(p.world, explosionPos, above, assets.ImageCommanderShotExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionFighterLaser:
		effect := newEffectNode(p.world, explosionPos, above, assets.ImageFighterLaserExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionHeavyCrawlerLaser:
		effect := newEffectNode(p.world, explosionPos, above, assets.ImageHeavyCrawlerLaserExplosion)
		p.world.nodeRunner.AddObject(effect)
		effect.anim.SetSecondsPerFrame(0.035)
	case gamedata.ProjectileExplosionPurple:
		soundIndex := p.world.localRand.IntRange(0, 2)
		sound := assets.AudioPurpleExplosion1 + resource.AudioID(soundIndex)
		p.world.nodeRunner.AddObject(newEffectNode(p.world, explosionPos, above, assets.ImagePurpleExplosion))
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
		dmg.Health *= damageMultiplier(p.target, p.weapon)
	}
	p.target.OnDamage(dmg, p.attacker)
	p.createExplosion()

	if p.weapon == gamedata.AtomicBombWeapon {
		if len(p.world.allColonies) == 0 {
			p.world.result.AtomicBombVictory = true
		}
	}
}
