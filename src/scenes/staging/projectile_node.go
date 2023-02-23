package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
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
	weapon    *weaponStats

	rotation gmath.Rad

	scene *ge.Scene

	camera *viewport.Camera
	sprite *ge.Sprite
}

type projectileTarget interface {
	GetPos() *gmath.Vec
	GetVelocity() gmath.Vec
	OnDamage(damage damageValue, source gmath.Vec)
	IsDisposed() bool
	IsFlying() bool
}

type weaponStats struct {
	MaxTargets            int
	ProjectileImage       resource.ImageID
	ProjectileSpeed       float64
	ProjectileRotateSpeed float64
	ImpactArea            float64
	AttackRange           float64
	Damage                damageValue
	Explosion             projectileExplosionKind
	BurstSize             int
	Reload                float64
	AttackSound           resource.AudioID
	FireOffset            gmath.Vec
}

type projectileConfig struct {
	Weapon    *weaponStats
	Camera    *viewport.Camera
	FromPos   *gmath.Vec
	ToPos     gmath.Vec
	Target    projectileTarget
	FireDelay float64
}

func newProjectileNode(config projectileConfig) *projectileNode {
	return &projectileNode{
		camera:    config.Camera,
		weapon:    config.Weapon,
		fromPos:   config.FromPos,
		pos:       *config.FromPos,
		toPos:     config.ToPos,
		target:    config.Target,
		fireDelay: config.FireDelay,
	}
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
	p.camera.AddGraphicsAbove(p.sprite)
}

func (p *projectileNode) IsDisposed() bool { return p.sprite.IsDisposed() }

func (p *projectileNode) Update(delta float64) {
	if p.fireDelay > 0 {
		p.fireDelay -= delta
		if p.fireDelay <= 0 {
			p.sprite.Visible = true
			p.pos = *p.fromPos
		} else {
			return
		}
	}

	travelled := p.weapon.ProjectileSpeed * delta
	if p.pos.DistanceTo(p.toPos) <= travelled {
		p.detonate()
		return
	}
	p.pos = p.pos.MoveTowards(p.toPos, travelled)
	if p.weapon.ProjectileRotateSpeed != 0 {
		p.rotation += gmath.Rad(delta * p.weapon.ProjectileRotateSpeed)
	}
}

func (p *projectileNode) detonate() {
	p.sprite.Dispose()
	if p.target.IsDisposed() {
		return
	}
	if p.toPos.DistanceTo(*p.target.GetPos()) > p.weapon.ImpactArea {
		return
	}
	p.target.OnDamage(p.weapon.Damage, *p.fromPos)

	switch p.weapon.Explosion {
	case projectileExplosionNormal:
		createExplosion(p.scene, p.camera, p.target.IsFlying(), p.pos.Add(p.scene.Rand().Offset(-3, 3)))
	}
}
