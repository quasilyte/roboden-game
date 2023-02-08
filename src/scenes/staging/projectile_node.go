package staging

import (
	"github.com/quasilyte/colony-game/viewport"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type projectileExplosionKind int

const (
	projectileExplosionNone projectileExplosionKind = iota
	projectileExplosionNormal
)

type projectileNode struct {
	image       resource.ImageID
	fromPos     gmath.Vec
	pos         gmath.Vec
	toPos       gmath.Vec
	target      projectileTarget
	speed       float64
	rotateSpeed float64
	area        float64
	damage      damageValue
	explosion   projectileExplosionKind

	rotation gmath.Rad

	scene *ge.Scene

	camera *viewport.Camera
	sprite *ge.Sprite
}

type projectileTarget interface {
	GetPos() gmath.Vec
	OnDamage(damage damageValue, source gmath.Vec)
	IsDisposed() bool
}

type projectileConfig struct {
	Camera      *viewport.Camera
	Image       resource.ImageID
	FromPos     gmath.Vec
	ToPos       gmath.Vec
	Target      projectileTarget
	Area        float64
	Speed       float64
	RotateSpeed float64
	Damage      damageValue
	Explosion   projectileExplosionKind
}

func newProjectileNode(config projectileConfig) *projectileNode {
	return &projectileNode{
		camera:      config.Camera,
		image:       config.Image,
		fromPos:     config.FromPos,
		pos:         config.FromPos,
		toPos:       config.ToPos,
		target:      config.Target,
		area:        config.Area,
		speed:       config.Speed,
		damage:      config.Damage,
		rotateSpeed: config.RotateSpeed,
		explosion:   config.Explosion,
	}
}

func (p *projectileNode) Init(scene *ge.Scene) {
	p.scene = scene
	if p.rotateSpeed == 0 {
		p.rotation = p.pos.AngleToPoint(p.toPos)
	} else {
		p.rotation = scene.Rand().Rad()
	}

	p.sprite = scene.NewSprite(p.image)
	p.sprite.Pos.Base = &p.pos
	p.sprite.Rotation = &p.rotation
	p.camera.AddGraphicsAbove(p.sprite)
}

func (p *projectileNode) IsDisposed() bool { return p.sprite.IsDisposed() }

func (p *projectileNode) Update(delta float64) {
	travelled := p.speed * delta
	if p.pos.DistanceTo(p.toPos) <= travelled {
		p.detonate()
		return
	}
	p.pos = p.pos.MoveTowards(p.toPos, travelled)
	if p.rotateSpeed != 0 {
		p.rotation += gmath.Rad(delta * p.rotateSpeed)
	}
}

func (p *projectileNode) detonate() {
	p.sprite.Dispose()
	if p.target.IsDisposed() {
		return
	}
	if p.toPos.DistanceTo(p.target.GetPos()) > p.area {
		return
	}
	p.target.OnDamage(p.damage, p.fromPos)

	switch p.explosion {
	case projectileExplosionNormal:
		createExplosion(p.scene, p.camera, p.pos.Add(p.scene.Rand().Offset(-3, 3)))
	}
}
