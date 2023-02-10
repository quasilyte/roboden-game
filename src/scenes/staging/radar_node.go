package staging

import (
	"math"

	"github.com/quasilyte/colony-game/assets"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type radarNode struct {
	sprite *ge.Sprite
	wave   *ge.Sprite

	scene *ge.Scene

	nearDist       float64
	nearDistPixels float64
	scaleRatio     float64

	bossSpot *ge.Sprite

	pos       gmath.Vec
	direction gmath.Rad

	world  *worldState
	colony *colonyCoreNode
}

func newRadarNode(world *worldState) *radarNode {
	return &radarNode{
		world:    world,
		nearDist: 1536,
	}
}

func (r *radarNode) SetBase(colony *colonyCoreNode) {
	r.colony = colony
	r.bossSpot.Visible = false
}

func (r *radarNode) IsDisposed() bool { return false }

func (r *radarNode) Init(scene *ge.Scene) {
	r.scene = scene

	r.sprite = scene.NewSprite(assets.ImageRadar)
	r.sprite.Pos.Base = &r.pos
	scene.AddGraphicsAbove(r.sprite, 1)

	r.nearDistPixels = 54.0
	r.scaleRatio = r.nearDistPixels / r.nearDist

	r.pos = gmath.Vec{
		X: r.world.camera.Rect.Width() - r.sprite.ImageWidth()/2 - 8,
		Y: r.sprite.ImageHeight()/2 + 8,
	}

	r.wave = scene.NewSprite(assets.ImageRadarWave)
	r.wave.Pos.Base = &r.pos
	r.wave.Rotation = &r.direction
	scene.AddGraphicsAbove(r.wave, 1)

	r.bossSpot = ge.NewSprite(scene.Context())
	r.bossSpot.Pos.Base = &r.pos
	r.bossSpot.Centered = false
	r.bossSpot.Visible = false
	scene.AddGraphicsAbove(r.bossSpot, 1)
}

func (r *radarNode) Update(delta float64) {
	r.sprite.Visible = r.colony != nil
	r.wave.Visible = r.colony != nil
	if r.bossSpot.Visible && r.colony == nil {
		r.bossSpot.Visible = false
	}
	if r.colony == nil {
		return
	}

	r.direction += gmath.Rad(delta)

	if r.world.boss == nil {
		return
	}
	radarScanDirection := (r.direction.Normalized() + 2*math.Pi)
	bossDirection := r.colony.pos.AngleToPoint(r.world.boss.pos).Normalized() + 2*math.Pi
	if radarScanDirection.AngleDelta2(bossDirection) < 0.1 && !r.bossSpot.Visible {
		r.bossSpot.Visible = true
		r.bossSpot.SetAlpha(1)
		return
	}
	if !r.bossSpot.Visible {
		return
	}
	r.bossSpot.SetAlpha(r.bossSpot.GetAlpha() - float32(delta*0.15))
	if r.bossSpot.GetAlpha() < 0.2 {
		r.bossSpot.Visible = false
		return
	}
	bossDist := r.world.boss.pos.DistanceTo(r.colony.pos)
	if bossDist > r.nearDist {
		// Boss is far away.
		r.bossSpot.Pos.Offset = gmath.RadToVec(bossDirection).Mulf(r.nearDistPixels).Sub(gmath.Vec{X: 2, Y: 2})
		if r.bossSpot.ImageID() != assets.ImageRadarBossFar {
			r.bossSpot.SetImage(r.scene.LoadImage(assets.ImageRadarBossFar))
		}
	} else {
		// Boss is near.
		r.bossSpot.Pos.Offset = gmath.RadToVec(bossDirection).Mulf(bossDist * r.scaleRatio).Sub(gmath.Vec{X: 2, Y: 2})
		if r.bossSpot.ImageID() != assets.ImageRadarBossNear {
			r.bossSpot.SetImage(r.scene.LoadImage(assets.ImageRadarBossNear))
		}
	}
}
