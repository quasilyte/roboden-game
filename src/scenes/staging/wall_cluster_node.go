package staging

import (
	"fmt"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
)

type mountainKind int

const (
	mountainSmall mountainKind = iota
	mountainMedium
	mountainBig
	mountainWide
	mountainTall
)

type wallAtras struct {
	layers []wallAtlasLayer
}

type wallChunk struct {
	pos  gmath.Vec
	kind mountainKind
}

type wallAtlasLayer struct {
	weight  float64
	texture resource.ImageID
}

var (
	landcrackAtlas = []wallAtlasLayer{
		{texture: assets.ImageLandCrack, weight: 0.35},
		{texture: assets.ImageLandCrack2, weight: 0.3},
		{texture: assets.ImageLandCrack3, weight: 0.25},
		{texture: assets.ImageLandCrack4, weight: 0.1},
	}
)

type wallClusterNode struct {
	world *worldState

	atlas wallAtras

	rect      gmath.Rect
	rectShape bool

	chunks []wallChunk

	points []gmath.Vec
}

type wallClusterConfig struct {
	// Settings for image-filling walls like mountains.
	chunks []wallChunk

	// Settings for oriented walls like landcracks.
	world  *worldState
	atlas  wallAtras
	points []gmath.Vec
}

func newWallClusterNode(config wallClusterConfig) *wallClusterNode {
	return &wallClusterNode{
		world:  config.world,
		atlas:  config.atlas,
		points: config.points,
		chunks: config.chunks,
	}
}

func (w *wallClusterNode) IsDisposed() bool { return false }

func (w *wallClusterNode) Init(scene *ge.Scene) {
}

func (w *wallClusterNode) initChunks(bg *ge.TiledBackground, scene *ge.Scene) {
	w.points = make([]gmath.Vec, len(w.chunks), len(w.chunks)+8)

	pointSet := make(map[gmath.Vec]struct{}, len(w.points)+8)
	for i, chunk := range w.chunks {
		pointSet[chunk.pos] = struct{}{}
		w.points[i] = chunk.pos
	}

	if !w.world.simulation {
		w.drawMountains(bg, scene)
	}

	pushNewPoint := func(pos gmath.Vec) {
		if _, ok := pointSet[pos]; ok {
			return
		}
		pointSet[pos] = struct{}{}
		w.points = append(w.points, pos)
	}

	for _, chunk := range w.chunks {
		pointSet[chunk.pos] = struct{}{}
		switch chunk.kind {
		case mountainBig:
			pushNewPoint(chunk.pos.Add(gmath.Vec{Y: pathing.CellSize}))
			pushNewPoint(chunk.pos.Add(gmath.Vec{Y: -pathing.CellSize}))
			pushNewPoint(chunk.pos.Add(gmath.Vec{X: pathing.CellSize}))
			pushNewPoint(chunk.pos.Add(gmath.Vec{X: -pathing.CellSize}))
		case mountainWide:
			pushNewPoint(chunk.pos.Add(gmath.Vec{X: pathing.CellSize}))
			pushNewPoint(chunk.pos.Add(gmath.Vec{X: -pathing.CellSize}))
		case mountainTall:
			pushNewPoint(chunk.pos.Add(gmath.Vec{Y: pathing.CellSize}))
			pushNewPoint(chunk.pos.Add(gmath.Vec{Y: -pathing.CellSize}))
		}
	}

	w.initGeometryRect()
}

func (w *wallClusterNode) drawMountains(bg *ge.TiledBackground, scene *ge.Scene) {
	type pendingImage struct {
		data    *ebiten.Image
		options ebiten.DrawImageOptions
		rect    gmath.Rect
	}
	images := make([]pendingImage, 0, len(w.chunks)+8)

	for i, chunk := range w.chunks {
		var texture resource.ImageID
		numSprites := 1
		switch chunk.kind {
		case mountainSmall:
			texture = assets.ImageMountainSmall
			roll := w.world.localRand.Float()
			if roll < 0.7 {
				numSprites = 2
			} else if roll < 0.8 {
				numSprites = 3
			}
		case mountainMedium:
			texture = assets.ImageMountainMedium
		case mountainBig:
			texture = assets.ImageMountainBig
		case mountainWide:
			texture = assets.ImageMountainWide
		case mountainTall:
			texture = assets.ImageMountainTall
			roll := w.world.localRand.Float()
			if roll < 0.4 {
				numSprites = 2
			}
		default:
			panic("unexpected chunk size")
		}

		switch gamedata.EnvironmentKind(w.world.config.Environment) {
		case gamedata.EnvMoon:
			// That's the default.
		case gamedata.EnvSwamp:
			texture += 5
		}

		for j := 0; j < numSprites; j++ {
			img := scene.LoadImage(texture)
			width, height := img.Data.Size()
			numFrames := width / int(img.DefaultFrameWidth)
			var drawOptions ebiten.DrawImageOptions
			frameOffset := w.world.localRand.IntRange(0, numFrames-1) * int(img.DefaultFrameWidth)
			subImage := createSubImage(img, frameOffset)
			if w.world.localRand.Bool() {
				drawOptions.GeoM.Scale(-1, 1)
				drawOptions.GeoM.Translate(img.DefaultFrameWidth, 0)
			}

			min := gmath.Vec{
				X: w.points[i].X - float64(img.DefaultFrameWidth/2),
				Y: w.points[i].Y - float64(height/2),
			}
			drawOptions.GeoM.Translate(min.X, min.Y)
			imageRect := gmath.Rect{
				Min: min,
				Max: min.Add(gmath.Vec{X: img.DefaultFrameWidth, Y: float64(height)}),
			}
			var offset gmath.Vec
			if numSprites == 1 {
				offset = w.world.localRand.Offset(-4, 4)
			} else {
				offset = w.world.localRand.Offset(-8, 8)
			}
			imageRect.Min = imageRect.Min.Add(offset)
			imageRect.Max = imageRect.Max.Add(offset)
			drawOptions.GeoM.Translate(offset.X, offset.Y)
			images = append(images, pendingImage{
				data:    subImage,
				options: drawOptions,
				rect:    imageRect,
			})
		}
	}
	sort.SliceStable(images, func(i, j int) bool {
		shape1 := images[i].rect
		shape2 := images[j].rect
		return shape1.Max.Y < shape2.Max.Y
	})
	for _, img := range images {
		bg.DrawImage(img.data, &img.options)
	}
}

func (w *wallClusterNode) initGeometryRect() {
	w.rect.Min = w.points[0]
	w.rect.Max = w.points[0]

	for _, p := range w.points {
		if p.X < w.rect.Min.X {
			w.rect.Min.X = p.X
		}
		if p.X > w.rect.Max.X {
			w.rect.Max.X = p.X
		}
		if p.Y < w.rect.Min.Y {
			w.rect.Min.Y = p.Y
		}
		if p.Y > w.rect.Max.Y {
			w.rect.Max.Y = p.Y
		}
	}

	width := int((w.rect.Max.X-w.rect.Min.X)/wallTileSize) + 1
	height := int((w.rect.Max.Y-w.rect.Min.Y)/wallTileSize) + 1
	areaSqr := width * height
	innerAreaWidth := gmath.ClampMin(width-2, 0)
	innerAreaHeight := gmath.ClampMin(height-2, 0)
	innerAreaSqr := innerAreaWidth * innerAreaHeight
	numPoints := len(w.points)
	w.rectShape = numPoints+innerAreaSqr == areaSqr
	if w.rectShape {
		for _, p := range w.points {
			if p.Y == w.rect.Min.Y && p.X <= w.rect.Max.X {
				continue // Top line
			}
			if p.Y == w.rect.Max.Y && p.X <= w.rect.Max.X {
				continue // Bottom line
			}
			if p.X == w.rect.Min.X && p.Y <= w.rect.Max.Y {
				continue // Left line
			}
			if p.X == w.rect.Max.X && p.Y <= w.rect.Max.Y {
				continue // Right line
			}
			w.rectShape = false
			break
		}
	}
}

func (w *wallClusterNode) initOriented(bg *ge.TiledBackground, scene *ge.Scene) {
	if len(w.points) > maxWallSegments {
		panic(fmt.Sprintf("too many segments in a wall cluster: %d", len(w.points)))
	}
	if len(w.points) == 0 {
		panic("empty wall cluster")
	}

	layerPicker := gmath.NewRandPicker[resource.ImageID](w.world.localRand)
	for _, l := range w.atlas.layers {
		layerPicker.AddOption(l.texture, l.weight)
	}

	w.initGeometryRect()
	origin := w.rect.Min

	getGridCoords := func(pos gmath.Vec) (int, int) {
		x := int(pos.X) / int(wallTileSize)
		y := int(pos.Y) / int(wallTileSize)
		return x, y
	}

	clusterMap := [maxWallSegments][maxWallSegments]uint8{}
	for i, p := range w.points {
		id := uint8(i + 1)
		x, y := getGridCoords(p.Sub(origin))
		if clusterMap[y][x] != 0 {
			panic("duplicated wall tile")
		}
		clusterMap[y][x] = (id + 1)
	}

	type checkOption struct {
		dx int8
		dy int8
	}
	checkList := [4]checkOption{
		{1, 0},
		{0, 1},
		{-1, 0},
		{0, -1},
	}
	maxX := len(clusterMap)
	maxY := len(clusterMap)
	for i := range w.points {
		p := w.points[i]
		connectionsMask := uint8(0)
		wallX, wallY := getGridCoords(p.Sub(origin))
		for bitIndex, option := range checkList {
			x := wallX + int(option.dx)
			y := wallY + int(option.dy)
			if x < 0 || x >= maxX || y < 0 || y >= maxY {
				continue // This grid cell is out of bounds
			}
			if clusterMap[y][x] == 0 {
				continue // No walls here
			}
			bitMask := 1 << bitIndex
			connectionsMask |= uint8(bitMask)
		}

		if !w.world.simulation {
			texture := layerPicker.Pick()
			var drawOptions ebiten.DrawImageOptions
			img := scene.LoadImage(texture)
			drawOptions.GeoM.Translate(w.points[i].X-(wallTileSize/2), w.points[i].Y-(wallTileSize/2))
			frameOffset := int(connectionsMask) * int(wallTileSize)
			subImage := createSubImage(img, frameOffset)
			bg.DrawImage(subImage, &drawOptions)
		}
	}
}

func (w *wallClusterNode) Update(delta float64) {
}

func (w *wallClusterNode) CollidesWith(pos gmath.Vec, r float64) bool {
	bounds := w.rect
	bounds.Min.X = gmath.ClampMin(bounds.Min.X-r, 0)
	bounds.Min.Y = gmath.ClampMin(bounds.Min.Y-r, 0)
	bounds.Max.X = gmath.ClampMax(bounds.Max.X+r, w.world.width)
	bounds.Max.Y = gmath.ClampMax(bounds.Max.Y+r, w.world.height)

	if w.rectShape {
		return bounds.Contains(pos)
	}

	// If the extended rect doesn't contain a pos, there is no need to check further.
	if !bounds.Contains(pos) {
		return false
	}

	r2 := r * r
	for _, p := range w.points {
		if p.DistanceSquaredTo(pos) < r2 {
			return true
		}
	}

	return false
}
