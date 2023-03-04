package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/viewport"
)

type projectileExplosionKind int

const (
	projectileExplosionNone projectileExplosionKind = iota
	projectileExplosionNormal
)

type projectileNode struct {
	fromPos   *gmath.Vec
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
	FromPos   *gmath.Vec
	ToPos     gmath.Vec
	Target    projectileTarget
	FireDelay float64
}

func newProjectileNode(config projectileConfig) *projectileNode {
	p := &projectileNode{
		camera:    config.Camera,
		weapon:    config.Weapon,
		fromPos:   config.FromPos,
		pos:       config.FromPos.Add(config.Weapon.FireOffset),
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
		if scene.Rand().Chance(0.4) {
			// Most likely will be a miss.
			p.toPos = p.toPos.Add(scene.Rand().Offset(-28, 28))
		}
	}
}

func (p *projectileNode) IsDisposed() bool { return p.sprite.IsDisposed() }

func (p *projectileNode) Update(delta float64) {
	if p.fireDelay > 0 {
		p.fireDelay -= delta
		if p.fireDelay <= 0 {
			p.sprite.Visible = true
			p.pos = p.fromPos.Add(p.weapon.FireOffset)
			p.arcStart = p.pos
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

func (p *projectileNode) detonate() {
	p.sprite.Dispose()
	if p.target.IsDisposed() {
		return
	}
	if p.toPos.DistanceSquaredTo(*p.target.GetPos()) > p.weapon.ImpactAreaSqr {
		return
	}
	p.target.OnDamage(p.weapon.Damage, *p.fromPos)

	switch p.weapon.Explosion {
	case gamedata.ProjectileExplosionNormal:
		createExplosion(p.scene, p.camera, p.target.IsFlying(), p.pos.Add(p.scene.Rand().Offset(-3, 3)))
	}
}
