package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/pathing"
	"github.com/quasilyte/roboden-game/viewport"
)

func randIterate[T any](rand *gmath.Rand, slice []T, f func(x T) bool) T {
	var result T
	if len(slice) == 0 {
		return result
	}
	var slider gmath.Slider
	slider.SetBounds(0, len(slice)-1)
	slider.TrySetValue(rand.IntRange(0, len(slice)-1))
	inc := rand.Bool()
	for i := 0; i < len(slice); i++ {
		x := slice[slider.Value()]
		if inc {
			slider.Inc()
		} else {
			slider.Dec()
		}
		if f(x) {
			result = x
			break
		}
	}
	return result
}

func randomSectorPos(rng *gmath.Rand, sector gmath.Rect) gmath.Vec {
	return gmath.Vec{
		X: rng.FloatRange(sector.Min.X, sector.Max.X),
		Y: rng.FloatRange(sector.Min.Y, sector.Max.Y),
	}
}

func posMove(pos gmath.Vec, d pathing.Direction) gmath.Vec {
	switch d {
	case pathing.DirRight:
		return pos.Add(gmath.Vec{X: pathing.CellSize})
	case pathing.DirDown:
		return pos.Add(gmath.Vec{Y: pathing.CellSize})
	case pathing.DirLeft:
		return pos.Add(gmath.Vec{X: -pathing.CellSize})
	case pathing.DirUp:
		return pos.Add(gmath.Vec{Y: -pathing.CellSize})
	default:
		return pos
	}
}

func posIsFree(world *worldState, skipColony *colonyCoreNode, pos gmath.Vec, radius float64) bool {
	wallCheckRadius := radius + 24
	for _, wall := range world.walls {
		if wall.CollidesWith(pos, wallCheckRadius) {
			return false
		}
	}

	for _, source := range world.essenceSources {
		if source.pos.DistanceTo(pos) < (radius + source.stats.size) {
			return false
		}
	}
	for _, construction := range world.constructions {
		if construction.pos.DistanceTo(pos) < (radius + 40) {
			return false
		}
	}
	for _, colony := range world.colonies {
		for _, turret := range colony.turrets {
			if turret.pos.DistanceTo(pos) < (radius + 32) {
				return false
			}
		}
		// TODO: flying colonies are not a problem.
		if colony == skipColony {
			continue
		}
		if colony.pos.DistanceTo(pos) < (radius + 40) {
			return false
		}
	}
	for _, creep := range world.creeps {
		if creep.stats.shadowImage == assets.ImageNone && creep.pos.DistanceTo(pos) < (radius+creep.stats.size) {
			return false
		}
	}
	return true
}

func createAreaExplosion(scene *ge.Scene, camera *viewport.Camera, rect gmath.Rect, allowVertical bool) {
	// FIXME: Rect.Center() does not work properly in gmath.
	center := gmath.Vec{
		X: rect.Max.X - rect.Width()*0.5,
		Y: rect.Max.Y - rect.Height()*0.5,
	}
	size := rect.Width() * rect.Height()
	minExplosions := gmath.ClampMin(size/120.0, 1)
	numExplosions := scene.Rand().IntRange(int(minExplosions), int(minExplosions*1.3))
	above := !allowVertical
	for numExplosions > 0 {
		offset := gmath.Vec{
			X: scene.Rand().FloatRange(-rect.Width()*0.4, rect.Width()*0.4),
			Y: scene.Rand().FloatRange(-rect.Height()*0.4, rect.Height()*0.4),
		}
		if numExplosions >= 4 && scene.Rand().Chance(0.4) {
			numExplosions -= 4
			scene.AddObject(newEffectNode(camera, center.Add(offset), above, assets.ImageBigExplosion))
		} else {
			numExplosions--
			if allowVertical && scene.Rand().Chance(0.4) {
				effect := newEffectNode(camera, center.Add(offset), above, assets.ImageVerticalExplosion)
				scene.AddObject(effect)
				effect.anim.SetSecondsPerFrame(0.035)
			} else {
				createMuteExplosion(scene, camera, above, center.Add(offset))
			}
		}
	}
	playExplosionSound(scene, camera, center)
}

func createMuteExplosion(scene *ge.Scene, camera *viewport.Camera, above bool, pos gmath.Vec) {
	explosion := newEffectNode(camera, pos, above, assets.ImageSmallExplosion1)
	scene.AddObject(explosion)
}

func playIonExplosionSound(scene *ge.Scene, camera *viewport.Camera, pos gmath.Vec) {
	explosionSoundIndex := scene.Rand().IntRange(0, 1)
	explosionSound := resource.AudioID(int(assets.AudioIonZap1) + explosionSoundIndex)
	playSound(scene, camera, explosionSound, pos)
}

func playExplosionSound(scene *ge.Scene, camera *viewport.Camera, pos gmath.Vec) {
	explosionSoundIndex := scene.Rand().IntRange(0, 4)
	explosionSound := resource.AudioID(int(assets.AudioExplosion1) + explosionSoundIndex)
	playSound(scene, camera, explosionSound, pos)
}

func createExplosion(scene *ge.Scene, camera *viewport.Camera, above bool, pos gmath.Vec) {
	createMuteExplosion(scene, camera, above, pos)
	playExplosionSound(scene, camera, pos)
}

func spriteRect(pos gmath.Vec, sprite *ge.Sprite) gmath.Rect {
	offset := gmath.Vec{X: sprite.FrameWidth * 0.5, Y: sprite.FrameHeight * 0.5}
	return gmath.Rect{
		Min: pos.Sub(offset),
		Max: pos.Add(offset),
	}
}

func roundedPos(pos gmath.Vec) gmath.Vec {
	x := int(pos.X)
	y := int(pos.Y)
	if x%2 != 0 {
		x++
	}
	if y%2 != 0 {
		y++
	}
	return gmath.Vec{X: float64(x), Y: float64(y)}
}

func correctedPos(sector gmath.Rect, pos gmath.Vec, pad float64) gmath.Vec {
	if pos.X < (pad + sector.Min.X) {
		pos.X = pad + sector.Min.X
	} else if pos.X > (sector.Max.X - pad) {
		pos.X = sector.Max.X - pad
	}
	if pos.Y < (pad + sector.Min.Y) {
		pos.Y = pad + sector.Min.Y
	} else if pos.Y > (sector.Max.Y - pad) {
		pos.Y = sector.Max.Y - pad
	}
	return pos
}

func snipePos(projectileSpeed float64, fireFrom, targetPos, targetVelocity gmath.Vec) gmath.Vec {
	if targetVelocity.IsZero() || projectileSpeed == 0 {
		return targetPos
	}
	dist := targetPos.DistanceTo(fireFrom)
	predictedPos := targetPos.Add(targetVelocity.Mulf(dist / projectileSpeed))
	return predictedPos
}

func retreatPos(rand *gmath.Rand, dist float64, objectPos, threatPos gmath.Vec) gmath.Vec {
	direction := threatPos.AngleToPoint(objectPos) + gmath.Rad(rand.FloatRange(-0.2, 0.2))
	return objectPos.MoveInDirection(dist, direction)
}

func playSound(scene *ge.Scene, camera *viewport.Camera, id resource.AudioID, pos gmath.Vec) {
	if camera.ContainsPos(pos) {
		scene.Audio().PlaySound(id)
	}
}
