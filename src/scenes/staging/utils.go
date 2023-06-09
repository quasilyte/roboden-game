package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
)

func pointToLineDistance(point, a, b gmath.Vec) float64 {
	s1 := -b.Y + a.Y
	s2 := b.X - a.X
	return math.Abs((point.X-a.X)*s1+(point.Y-a.Y)*s2) / math.Sqrt(s1*s1+s2*s2)
}

func midpoint(a, b gmath.Vec) gmath.Vec {
	return a.Add(b).Mulf(0.5)
}

// ? ? ? ?
// ? o o ?
// ? o x ?
// ? ? ? ?
var colonyNearCellOffsets = []pathing.GridCoord{
	{X: -2, Y: -2},
	{X: -1, Y: -2},
	{X: 0, Y: -2},
	{X: 1, Y: -2},
	{X: 1, Y: -1},
	{X: 1, Y: 0},
	{X: 1, Y: 1},
	{X: 0, Y: 1},
	{X: -1, Y: 1},
	{X: -2, Y: 1},
	{X: -2, Y: 0},
	{X: -2, Y: -1},
}

// ? ? ? ? ?
// ? o o . ?
// ? o x . ?
// ? . . . ?
// ? ? ? ? ?
var colonyNear2x2CellOffsets = []pathing.GridCoord{
	{X: -2, Y: -2},
	{X: -1, Y: -2},
	{X: 0, Y: -2},
	{X: 1, Y: -2},
	{X: 2, Y: -2},
	{X: 2, Y: -1},
	{X: 2, Y: 0},
	{X: 2, Y: 1},
	{X: 2, Y: 2},
	{X: 1, Y: 2},
	{X: 0, Y: 2},
	{X: -1, Y: 2},
	{X: -2, Y: 2},
	{X: -2, Y: 1},
	{X: -2, Y: 0},
	{X: -2, Y: -1},
}

func moveTowardsWithSpeed(from, to gmath.Vec, delta, speed float64) (gmath.Vec, bool) {
	travelled := speed * delta
	result := from.MoveTowards(to, travelled)
	return result, result == to
}

func randIterate[T any](rand *gmath.Rand, slice []T, f func(x T) bool) T {
	var result T
	if len(slice) == 0 {
		return result
	}
	if len(slice) == 1 {
		// Don't use rand() if there is only 1 element.
		x := slice[0]
		if f(x) {
			result = x
		}
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

type collisionFlags int

const (
	collisionSkipSmallCrawlers collisionFlags = 1 << iota
	collisionSkipTeleporters
)

func posIsFree(world *worldState, skipColony *colonyCoreNode, pos gmath.Vec, radius float64) bool {
	return posIsFreeWithFlags(world, skipColony, pos, radius, 0)
}

func posIsFreeWithFlags(world *worldState, skipColony *colonyCoreNode, pos gmath.Vec, radius float64, flags collisionFlags) bool {
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
	for _, colony := range world.allColonies {
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

	skipSmall := flags&collisionSkipSmallCrawlers != 0
	for _, creep := range world.creeps {
		if skipSmall && creep.stats.Kind == gamedata.CreepCrawler {
			continue
		}
		if creep.stats.ShadowImage == assets.ImageNone && creep.pos.DistanceTo(pos) < (radius+creep.stats.Size) {
			return false
		}
	}

	if flags&collisionSkipTeleporters == 0 {
		for _, tp := range world.teleporters {
			if tp.pos.DistanceTo(pos) < (radius + 54) {
				return false
			}
		}
	}

	return true
}

func createAreaExplosion(world *worldState, rect gmath.Rect, allowVertical bool) {
	// FIXME: Rect.Center() does not work properly in gmath.
	center := gmath.Vec{
		X: rect.Max.X - rect.Width()*0.5,
		Y: rect.Max.Y - rect.Height()*0.5,
	}
	size := rect.Width() * rect.Height()
	minExplosions := gmath.ClampMin(size/120.0, 1)
	numExplosions := world.rand.IntRange(int(minExplosions), int(minExplosions*1.3))
	above := !allowVertical
	for numExplosions > 0 {
		offset := gmath.Vec{
			X: world.rand.FloatRange(-rect.Width()*0.4, rect.Width()*0.4),
			Y: world.rand.FloatRange(-rect.Height()*0.4, rect.Height()*0.4),
		}
		if numExplosions >= 4 && world.rand.Chance(0.4) {
			numExplosions -= 4
			world.nodeRunner.AddObject(newEffectNode(world.stage, center.Add(offset), above, assets.ImageBigExplosion))
		} else {
			numExplosions--
			if allowVertical && world.rand.Chance(0.4) {
				effect := newEffectNode(world.stage, center.Add(offset), above, assets.ImageVerticalExplosion)
				world.nodeRunner.AddObject(effect)
				effect.anim.SetSecondsPerFrame(0.035)
			} else {
				createMuteExplosion(world, above, center.Add(offset))
			}
		}
	}
	playExplosionSound(world, center)
}

func createMuteExplosion(world *worldState, above bool, pos gmath.Vec) {
	explosion := newEffectNode(world.stage, pos, above, assets.ImageSmallExplosion1)
	world.nodeRunner.AddObject(explosion)
}

func playIonExplosionSound(world *worldState, pos gmath.Vec) {
	explosionSoundIndex := world.localRand.IntRange(0, 1)
	explosionSound := resource.AudioID(int(assets.AudioIonZap1) + explosionSoundIndex)
	playSound(world, explosionSound, pos)
}

func playExplosionSound(world *worldState, pos gmath.Vec) {
	explosionSoundIndex := world.localRand.IntRange(0, 4)
	explosionSound := resource.AudioID(int(assets.AudioExplosion1) + explosionSoundIndex)
	playSound(world, explosionSound, pos)
}

func createBigVerticalExplosion(world *worldState, pos gmath.Vec) {
	world.nodeRunner.AddObject(newEffectNode(world.stage, pos, false, assets.ImageBigVerticalExplosion))
	playExplosionSound(world, pos)
}

func createExplosion(world *worldState, above bool, pos gmath.Vec) {
	createMuteExplosion(world, above, pos)
	playExplosionSound(world, pos)
}

func spriteRect(pos gmath.Vec, sprite *ge.Sprite) gmath.Rect {
	offset := gmath.Vec{X: sprite.FrameWidth * 0.5, Y: sprite.FrameHeight * 0.5}
	return gmath.Rect{
		Min: pos.Sub(offset),
		Max: pos.Add(offset),
	}
}

func roundedPos(pos gmath.Vec) gmath.Vec {
	return gmath.Vec{
		X: math.Round(pos.X),
		Y: math.Round(pos.Y),
	}
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

func playSound(world *worldState, id resource.AudioID, pos gmath.Vec) {
	for _, cam := range world.cameras {
		if cam.ContainsPos(pos) {
			world.rootScene.Audio().PlaySound(id)
		}
	}
}
