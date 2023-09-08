package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type radarColonySpot struct {
	colony *colonyCoreNode
	spot   *ge.Sprite
}

type radarTurretSpot struct {
	turret *colonyAgentNode
	spot   *ge.Sprite
}

type radarNode struct {
	sprite *ge.Sprite
	wave   *ge.Sprite

	scene  *ge.Scene
	player *humanPlayer

	nearDist       float64
	nearDistPixels float64
	radius         float64
	diameter       float64
	scaleRatio     float64

	dark bool

	bossSpot *ge.Sprite
	bossPath *ge.Line

	cameraRect *ge.Rect
	colonies   []radarColonySpot
	turrets    []radarTurretSpot

	minimapRect gmath.Rect
	pos         gmath.Vec
	direction   gmath.Rad

	world  *worldState
	colony *colonyCoreNode
}

func newRadarNode(world *worldState, p *humanPlayer, dark bool) *radarNode {
	return &radarNode{
		world:    world,
		nearDist: 1536,
		player:   p,
		dark:     dark,
	}
}

func (r *radarNode) RemoveColony(colony *colonyCoreNode) {
	index := xslices.IndexWhere(r.colonies, func(elem radarColonySpot) bool {
		return elem.colony == colony
	})
	if index == -1 {
		return
	}
	c := r.colonies[index]
	c.spot.Dispose()
	r.colonies = xslices.RemoveAt(r.colonies, index)
}

func (r *radarNode) RemoveTurret(turret *colonyAgentNode) {
	index := xslices.IndexWhere(r.turrets, func(elem radarTurretSpot) bool {
		return elem.turret == turret
	})
	if index == -1 {
		return
	}
	t := r.turrets[index]
	t.spot.Dispose()
	r.turrets = xslices.RemoveAt(r.turrets, index)
}

func (r *radarNode) AddTurret(turret *colonyAgentNode) {
	spot := r.scene.NewSprite(assets.ImageRadarBossFar)
	spot.Pos.Base = &r.pos
	r.updateDarkTurrets()
	r.player.state.camera.UI.AddGraphics(spot)

	r.turrets = append(r.turrets, radarTurretSpot{
		turret: turret,
		spot:   spot,
	})
}

func (r *radarNode) AddColony(colony *colonyCoreNode) {
	spot := r.scene.NewSprite(assets.ImageRadarBossNear)
	spot.Pos.Base = &r.pos
	r.updateDarkColonies()
	r.player.state.camera.UI.AddGraphics(spot)

	r.colonies = append(r.colonies, radarColonySpot{
		colony: colony,
		spot:   spot,
	})
}

func (r *radarNode) SetBase(colony *colonyCoreNode) {
	if r.dark {
		panic("setting a base for a dark radar")
	}

	r.colony = colony
	r.bossSpot.Visible = false
	r.bossPath.Visible = false
}

func (r *radarNode) IsDisposed() bool { return false }

func (r *radarNode) Init(scene *ge.Scene) {
	r.scene = scene

	img := assets.ImageRadar
	if r.dark {
		img = assets.ImageDarkRadar
	}
	r.sprite = scene.NewSprite(img)
	r.sprite.Pos.Offset = gmath.Vec{
		X: 8 + r.sprite.ImageWidth()/2,
		Y: 1080/2 - (8 + r.sprite.ImageHeight()/2),
	}
	r.player.state.camera.UI.AddGraphics(r.sprite)

	r.radius = 55.0
	r.diameter = r.radius * 2
	r.nearDistPixels = r.radius - 1
	r.scaleRatio = r.nearDistPixels / r.nearDist

	r.pos = (gmath.Vec{X: 65, Y: 74}).Add(gmath.Vec{
		X: 8,
		Y: 1080/2 - r.sprite.ImageHeight() - 8,
	})

	if !r.dark {
		r.wave = scene.NewSprite(assets.ImageRadarWave)
		r.wave.Pos.Base = &r.pos
		r.wave.Rotation = &r.direction
		r.player.state.camera.UI.AddGraphics(r.wave)

		r.bossPath = ge.NewLine(ge.Pos{}, ge.Pos{})
		var pathColor ge.ColorScale
		pathColor.SetColor(ge.RGB(0x91234e))
		r.bossPath.SetColorScale(pathColor)
		r.bossPath.Visible = false
		r.player.state.camera.UI.AddGraphics(r.bossPath)
	}

	r.bossSpot = ge.NewSprite(scene.Context())
	r.bossSpot.Pos.Base = &r.pos
	r.bossSpot.Visible = false
	r.bossSpot.Centered = false
	r.player.state.camera.UI.AddGraphics(r.bossSpot)

	if r.dark {
		r.scaleRatio = r.diameter / r.world.width

		r.bossSpot.SetImage(r.scene.LoadImage(assets.ImageRadarAlliedSpot))
		r.bossSpot.Visible = true
		r.bossSpot.Centered = true

		cam := r.player.state.camera.Rect
		r.cameraRect = ge.NewRect(scene.Context(),
			math.Round(cam.Width()*r.scaleRatio),
			math.Round(cam.Height()*r.scaleRatio))
		r.cameraRect.OutlineWidth = 1
		r.cameraRect.FillColorScale.SetRGBA(0, 0, 0, 0)
		r.cameraRect.OutlineColorScale.SetColor(dpadBarColorNormal)
		r.cameraRect.Pos.Base = &r.pos
		r.player.state.camera.UI.AddGraphics(r.cameraRect)

		r.updateDark()

		r.minimapRect = r.sprite.BoundsRect()
		r.minimapRect.Min = r.minimapRect.Min.Add(gmath.Vec{X: 11, Y: 20})
		r.minimapRect.Max = r.minimapRect.Max.Sub(gmath.Vec{X: 11, Y: 9})
	}
}

func (r *radarNode) ResolveClick(clickPos gmath.Vec) (gmath.Vec, bool) {
	if !r.dark {
		return gmath.Vec{}, false
	}

	if r.minimapRect.Contains(clickPos) {
		p := clickPos.Sub(r.minimapRect.Min)
		return p.Mulf(r.world.width / r.diameter), true
	}

	return gmath.Vec{}, false
}

func (r *radarNode) setBossVisibility(visible bool) {
	r.bossSpot.Visible = visible
	r.bossPath.Visible = visible
}

func (r *radarNode) Update(delta float64) {
	if r.dark {
		r.updateDark()
	} else {
		r.update(delta)
	}
}

func (r *radarNode) translatePosToOffset(pos gmath.Vec) gmath.Vec {
	local := gmath.Vec{
		X: pos.X * r.scaleRatio,
		Y: pos.Y * r.scaleRatio,
	}
	return local.Sub(gmath.Vec{X: r.radius, Y: r.radius})
}

func (r *radarNode) updateDarkColonies() {
	for _, c := range r.colonies {
		c.spot.Pos.Offset = r.translatePosToOffset(c.colony.pos)
	}
}

func (r *radarNode) updateDarkTurrets() {
	for _, t := range r.turrets {
		t.spot.Pos.Offset = r.translatePosToOffset(t.turret.pos)
	}
}

func (r *radarNode) UpdateCamera() {
	if !r.dark {
		return
	}
	cameraOffset := r.player.state.camera.CenterPos()
	r.cameraRect.Pos.Offset = r.translatePosToOffset(cameraOffset)
}

func (r *radarNode) updateDark() {
	if r.world.boss != nil {
		r.bossSpot.Pos.Offset = r.translatePosToOffset(r.world.boss.pos)
	} else {
		r.bossSpot.Visible = false
	}

	r.updateDarkColonies()
	r.updateDarkTurrets()
	for _, t := range r.turrets {
		t.spot.Pos.Offset = r.translatePosToOffset(t.turret.pos)
	}
}

func (r *radarNode) update(delta float64) {
	if r.world.nodeRunner.IsPaused() {
		return
	}

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
