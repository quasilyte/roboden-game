package staging

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/pathing"
)

type forestClusterNode struct {
	world *worldState

	outerRect gmath.Rect
	innerRect gmath.Rect
	rects     []gmath.Rect

	config forestClusterConfig
}

type forestClusterConfig struct {
	pos    gmath.Vec
	width  int
	height int
}

func newForestClusterNode(world *worldState, config forestClusterConfig) *forestClusterNode {
	f := &forestClusterNode{
		world:  world,
		config: config,
	}

	innerOffset := gmath.Vec{X: pathing.CellSize * 2, Y: pathing.CellSize * 2}
	originPos := f.config.pos
	f.outerRect = gmath.Rect{
		Min: originPos,
		Max: originPos.Add(gmath.Vec{
			X: float64(f.config.width) * pathing.CellSize,
			Y: float64(f.config.height) * pathing.CellSize,
		}),
	}
	f.innerRect = gmath.Rect{
		Min: f.outerRect.Min.Add(innerOffset),
		Max: f.outerRect.Max.Sub(innerOffset),
	}

	return f
}

func (f *forestClusterNode) init(scene *ge.Scene, snowy bool) []pendingImage {
	// TODO: make() with a capacity hint.
	var images []pendingImage

	textureID := assets.ImageTrees
	if snowy {
		textureID = assets.ImageSnowTrees
	}

	texture := scene.LoadImage(textureID)
	numFrames := texture.Data.Bounds().Dx() / int(texture.DefaultFrameWidth)

	for y := 0; y < f.config.height; y++ {
		for x := 0; x < f.config.width; x++ {
			pos := f.config.pos.Add(gmath.Vec{
				X: float64(x) * pathing.CellSize,
				Y: float64(y) * pathing.CellSize,
			})

			if f.world.HasTreesAt(pos, 0) {
				continue
			}

			isInner := (x >= 2 && x <= (f.config.width-1)-2) &&
				(y >= 2 && y <= (f.config.height-1)-2)
			skipChance := 0.0
			if x == 0 || y == 0 || y == (f.config.height-1) || x == (f.config.width-1) {
				skipChance = 0.7
			} else if !isInner {
				skipChance = 0.35
			}
			if f.world.rand.Chance(skipChance) {
				continue
			}

			if !isInner {
				f.rects = append(f.rects, gmath.Rect{
					Min: pos,
					Max: pos.Add(gmath.Vec{X: pathing.CellSize, Y: pathing.CellSize}),
				})
			}

			if !f.world.simulation {
				var numSprites int
				if isInner {
					numSprites = f.world.localRand.IntRange(3, 5)
				} else {
					numSprites = f.world.localRand.IntRange(2, 3)
				}
				startAngle := f.world.localRand.Rad()
				angle := startAngle
				arcFinished := false
				for i := 0; i < numSprites; i++ {
					var drawOptions ebiten.DrawImageOptions
					frameOffset := f.world.localRand.IntRange(0, numFrames-1) * int(texture.DefaultFrameWidth)
					subImage := createSubImage(texture, frameOffset)
					if f.world.localRand.Bool() {
						drawOptions.GeoM.Scale(-1, 1)
						drawOptions.GeoM.Translate(texture.DefaultFrameWidth, 0)
					}
					var offset gmath.Vec
					if i == 0 {
						offset = f.world.localRand.Offset(-4, 4)
					} else {
						dist := f.world.localRand.FloatRange(8, 15)
						var dir gmath.Vec
						if !arcFinished {
							dir = gmath.RadToVec(angle)
							angleDelta := gmath.Rad(f.world.localRand.FloatRange(0.1, 0.7))
							angle += angleDelta
						} else {
							dir = gmath.RadToVec(f.world.localRand.Rad())
						}
						offset = dir.Mulf(dist)
						if !arcFinished && f.world.localRand.Chance(0.4) {
							arcFinished = true
						}
					}
					drawPos := pos.Add(offset)
					drawOptions.GeoM.Translate(drawPos.X, drawPos.Y)
					images = append(images, pendingImage{
						data:      subImage,
						options:   drawOptions,
						drawOrder: drawPos.Y + texture.DefaultFrameHeight,
					})
				}
			}
		}
	}

	return images
}

func (f *forestClusterNode) walkRects(visit func(rect gmath.Rect)) {
	visit(f.innerRect)
	for _, r := range f.rects {
		visit(r)
	}
}

func (f *forestClusterNode) CollidesWith(pos gmath.Vec, r float64) bool {
	if r == 0 {
		return f.ContainsPos(pos)
	}

	offset := gmath.Vec{X: r * 0.5, Y: r * 0.5}
	objectRect := gmath.Rect{
		Min: pos.Sub(offset),
		Max: pos.Add(offset),
	}

	if !f.outerRect.Overlaps(objectRect) {
		return false
	}
	if f.innerRect.Overlaps(objectRect) {
		return true
	}

	for _, r := range f.rects {
		if r.Overlaps(objectRect) {
			return true
		}
	}

	return false
}

func (f *forestClusterNode) ContainsPos(pos gmath.Vec) bool {
	// Two fast paths: one for the miss, one for the hit.
	if !f.outerRect.Contains(pos) {
		return false
	}
	if f.innerRect.Contains(pos) {
		return true
	}

	for _, r := range f.rects {
		if r.Contains(pos) {
			return true
		}
	}

	return false
}
