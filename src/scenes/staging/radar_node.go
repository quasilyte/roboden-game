package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type radarNode struct {
	sprite *ge.Sprite
	wave   *ge.Sprite

	scene *ge.Scene

	nearDist       float64
	nearDistPixels float64
	radius         float64
	diameter       float64
	scaleRatio     float64

	bossSpot *ge.Sprite
	bossPath *ge.Line

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
	r.bossPath.Visible = false
}

func (r *radarNode) IsDisposed() bool { return false }

func (r *radarNode) Init(scene *ge.Scene) {
	r.scene = scene

	r.sprite = scene.NewSprite(assets.ImageRadar)
	r.sprite.Pos.Offset = gmath.Vec{
		X: 8 + r.sprite.ImageWidth()/2,
		Y: 1080/2 - (8 + r.sprite.ImageHeight()/2),
	}
	scene.AddGraphicsAbove(r.sprite, 1)

	r.radius = 55.0
	r.diameter = r.radius * 2
	r.nearDistPixels = r.radius - 1
	r.scaleRatio = r.nearDistPixels / r.nearDist

	r.pos = (gmath.Vec{X: 67, Y: 83}).Add(gmath.Vec{
		X: 8,
		Y: 1080/2 - r.sprite.ImageHeight() - 8,
	})

	r.wave = scene.NewSprite(assets.ImageRadarWave)
	r.wave.Pos.Base = &r.pos
	r.wave.Rotation = &r.direction
	scene.AddGraphicsAbove(r.wave, 1)

	r.bossPath = ge.NewLine(ge.Pos{}, ge.Pos{})
	var pathColor ge.ColorScale
	pathColor.SetColor(ge.RGB(0x91234e))
	r.bossPath.SetColorScale(pathColor)
	r.bossPath.Visible = false
	scene.AddGraphicsAbove(r.bossPath, 1)

	r.bossSpot = ge.NewSprite(scene.Context())
	r.bossSpot.Pos.Base = &r.pos
	r.bossSpot.Centered = false
	r.bossSpot.Visible = false
	scene.AddGraphicsAbove(r.bossSpot, 1)
}

func (r *radarNode) setBossVisibility(visible bool) {
	r.bossSpot.Visible = visible
	r.bossPath.Visible = visible
}

func (r *radarNode) Update(delta float64) {
	r.sprite.Visible = r.colony != nil
	r.wave.Visible = r.colony != nil
	if r.bossSpot.Visible && r.colony == nil {
		r.setBossVisibility(false)
	}
	if r.colony == nil {
		return
	}

	r.direction += gmath.Rad(delta)

	if r.world.boss == nil {
		r.setBossVisibility(false)
		return
	}
	radarScanDirection := (r.direction.Normalized() + 2*math.Pi)
	bossDirection := r.colony.pos.AngleToPoint(r.world.boss.pos).Normalized() + 2*math.Pi
	if radarScanDirection.AngleDelta2(bossDirection) < 0.1 && !r.bossSpot.Visible {
		r.setBossVisibility(true)
		r.bossSpot.SetAlpha(1)
	}
	if !r.bossSpot.Visible {
		return
	}
	r.bossSpot.SetAlpha(r.bossSpot.GetAlpha() - float32(delta*0.15))
	r.bossPath.SetAlpha(r.bossSpot.GetAlpha())
	if r.bossSpot.GetAlpha() < 0.2 {
		r.setBossVisibility(false)
		return
	}
	bossDist := r.world.boss.pos.DistanceTo(r.colony.pos)
	extraOffset := gmath.Vec{X: 2, Y: 2}
	if bossDist > r.nearDist {
		// Boss is far away.
		r.bossSpot.Pos.Offset = gmath.RadToVec(bossDirection).Mulf(r.nearDistPixels).Sub(extraOffset)
		if r.bossSpot.ImageID() != assets.ImageRadarBossFar {
			r.bossSpot.SetImage(r.scene.LoadImage(assets.ImageRadarBossFar))
		}
		r.bossPath.Visible = false
	} else {
		// Boss is near.
		r.bossSpot.Pos.Offset = gmath.RadToVec(bossDirection).Mulf(bossDist * r.scaleRatio).Sub(extraOffset)
		if r.bossSpot.ImageID() != assets.ImageRadarBossNear {
			r.bossSpot.SetImage(r.scene.LoadImage(assets.ImageRadarBossNear))
		}
		r.bossSpot.Visible = true
		startPos := r.bossSpot.Pos.Resolve().Add(extraOffset)
		endPos := r.bossPath.BeginPos.Offset.Add(gmath.RadToVec(r.world.boss.GetVelocity().Angle()).Mulf(r.diameter))
		fromCircleToObject := endPos.Sub(r.pos)
		fromCircleToObject = fromCircleToObject.Mulf(r.radius / fromCircleToObject.Len())
		endPos = r.pos.Add(fromCircleToObject)
		r.bossPath.BeginPos.Offset = startPos
		r.bossPath.EndPos.Offset = endPos
	}
}
