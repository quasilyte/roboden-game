package staging

import (
	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type lavaPuddleNode struct {
	rect      gmath.Rect
	centerPos gmath.Vec
	sprite    *ge.Sprite

	fireDelay float64
	attacker  magmaDummyAttacker

	shaderTime      float64
	shaderTimeSpeed float64

	numResourceSpawns int
	maxResourceSpawns int

	world *worldState
}

type magmaDummyAttacker struct {
	pos gmath.Vec
}

func (a *magmaDummyAttacker) IsDisposed() bool { return false }

func (a *magmaDummyAttacker) GetPos() *gmath.Vec { return &a.pos }

func (a *magmaDummyAttacker) GetVelocity() gmath.Vec { return gmath.Vec{} }

func (a *magmaDummyAttacker) IsFlying() bool { return false }

func (a *magmaDummyAttacker) OnDamage(gamedata.DamageValue, targetable) {}

func (a *magmaDummyAttacker) GetTargetInfo() targetInfo {
	return targetInfo{}
}

func newLavaPuddleNode(world *worldState, rect gmath.Rect) *lavaPuddleNode {
	return &lavaPuddleNode{
		rect:      rect,
		world:     world,
		centerPos: rect.Min.Add(gmath.Vec{X: rect.Width() * 0.5, Y: rect.Height() * 0.5}),
	}
}

func (lava *lavaPuddleNode) Init(scene *ge.Scene) {
	lava.sprite = ge.NewSprite(scene.Context())
	lava.sprite.Centered = false
	lava.sprite.Pos.Base = &lava.rect.Min

	lava.fireDelay = lava.world.rand.FloatRange(30, 200)

	lava.maxResourceSpawns = lava.world.rand.IntRange(0, 3)

	if lava.world.graphicsSettings.AllShadersEnabled {
		lava.sprite.Shader = scene.NewShader(assets.ShaderLavaPuddle)
		lava.shaderTimeSpeed = 2.5 * lava.world.localRand.FloatRange(0.95, 1.1)
		lava.shaderTime = lava.world.localRand.FloatRange(545.7, 964.70)
		lava.sprite.Shader.SetFloatValue("Time", lava.shaderTime)
		lava.sprite.Shader.SetIntValue("Seed", lava.world.localRand.IntRange(0, 999))
	}

	texture := ebiten.NewImage(int(lava.rect.Width()), int(lava.rect.Height()))
	lava.sprite.SetImage(resource.Image{Data: texture})

	layerPicker := gmath.NewRandPicker[resource.ImageID](lava.world.localRand)
	for _, l := range lavaAtlas {
		layerPicker.AddOption(l.texture, l.weight)
	}

	for y := 0.0; y < lava.rect.Height(); y += 32.0 {
		for x := 0.0; x < lava.rect.Width(); x += 32.0 {
			tileImages := scene.LoadImage(layerPicker.Pick())
			lava.drawTile(texture, tileImages, x, y)
		}
	}

	lava.world.stage.AddSpriteBelow(lava.sprite)
}

func (lava *lavaPuddleNode) drawTile(dst *ebiten.Image, texture resource.Image, x, y float64) {
	var tileIndex int
	var flipHorizontal bool
	var flipVertical bool
	if x == 0 {
		switch {
		case y == 0:
			if lava.world.localRand.Bool() {
				tileIndex = 0
			} else {
				tileIndex = 2
				flipHorizontal = true
			}
		case y == lava.rect.Height()-32:
			if lava.world.localRand.Bool() {
				tileIndex = 6
			} else {
				tileIndex = 8
				flipHorizontal = true
			}
		default:
			if lava.world.localRand.Bool() {
				tileIndex = 3
			} else {
				tileIndex = 5
				flipHorizontal = true
			}
		}
	} else if x == lava.rect.Width()-32 {
		switch {
		case y == 0:
			if lava.world.localRand.Bool() {
				tileIndex = 2
			} else {
				tileIndex = 0
				flipHorizontal = true
			}
		case y == lava.rect.Height()-32:
			if lava.world.localRand.Bool() {
				tileIndex = 8
			} else {
				tileIndex = 6
				flipHorizontal = true
			}
		default:
			if lava.world.localRand.Bool() {
				tileIndex = 5
			} else {
				tileIndex = 3
				flipHorizontal = true
			}
		}
	} else {
		switch {
		case y == 0:
			tileIndex = 1
			flipHorizontal = lava.world.localRand.Bool()
		case y == lava.rect.Height()-32:
			tileIndex = 7
			flipHorizontal = lava.world.localRand.Bool()
		default:
			tileIndex = 4
			flipHorizontal = lava.world.localRand.Bool()
			flipVertical = lava.world.localRand.Bool()
		}
	}

	tile := createSubImage(texture, tileIndex*32)
	var drawOptions ebiten.DrawImageOptions
	if flipHorizontal {
		drawOptions.GeoM.Scale(-1, 1)
		drawOptions.GeoM.Translate(32, 0)
	}
	if flipVertical {
		drawOptions.GeoM.Scale(1, -1)
		drawOptions.GeoM.Translate(0, 32)
	}
	drawOptions.GeoM.Translate(x, y)
	dst.DrawImage(tile, &drawOptions)
}

func (lava *lavaPuddleNode) IsDisposed() bool { return false }

func (lava *lavaPuddleNode) Update(delta float64) {
	if !lava.sprite.Shader.IsNil() {
		lava.shaderTime += delta * lava.shaderTimeSpeed
		if lava.shaderTime > 999999999 {
			lava.shaderTime = lava.world.localRand.FloatRange(0, 9)
		}
		lava.sprite.Shader.SetFloatValue("Time", lava.shaderTime)
	}

	lava.fireDelay = gmath.ClampMin(lava.fireDelay-delta, 0)
	if lava.fireDelay != 0 {
		return
	}

	lava.fireDelay = lava.world.rand.FloatRange(10, 30)
	spawnPad := gmath.Vec{X: 20, Y: 20}
	spawnRect := gmath.Rect{
		Min: lava.rect.Min.Add(spawnPad),
		Max: lava.rect.Max.Sub(spawnPad),
	}
	spawnPos := randomSectorPos(lava.world.rand, spawnRect)

	weapon := gamedata.MagmaHazardWeapon
	var target targetable
	var attackPos gmath.Vec
	if lava.world.rand.Chance(0.3) {
		lava.world.FindTargetableAgents(spawnPos, true, weapon.AttackRange, func(a *colonyAgentNode) bool {
			target = a
			attackPos = snipePos(weapon.ProjectileSpeed, spawnPos, a.pos, a.GetVelocity())
			return true
		})
		if target == nil {
			lava.world.WalkCreeps(spawnPos, weapon.AttackRange, func(creep *creepNode) bool {
				if creep.stats.Kind != gamedata.CreepCrawler {
					return false
				}
				if creep.pos.DistanceSquaredTo(spawnPos) > weapon.AttackRangeSqr {
					return false
				}
				target = creep
				attackPos = snipePos(weapon.ProjectileSpeed, spawnPos, creep.pos, creep.GetVelocity())
				return true
			})
		}
	}
	if target == nil {
		dist := lava.world.rand.FloatRange(80, weapon.AttackRange)
		attackPos = gmath.RadToVec(lava.world.rand.Rad()).Mulf(dist).Add(spawnPos)
	}

	lava.attacker.pos = spawnPos
	p := lava.world.newProjectileNode(projectileConfig{
		World:    lava.world,
		Weapon:   weapon,
		Attacker: &lava.attacker,
		ToPos:    attackPos,
		Target:   target,
	})
	p.trailCounter = 0.1
	lava.world.nodeRunner.AddProjectile(p)
	createEffect(p.world, effectConfig{Pos: spawnPos, Image: assets.ImageFireBurst})

	if target == nil && lava.numResourceSpawns < lava.maxResourceSpawns && lava.world.rand.Chance(0.3) {
		p.EventDetonated.Connect(nil, func(pos gmath.Vec) {
			if !posIsFree(lava.world, nil, attackPos, 20) {
				return
			}
			res := lava.world.NewEssenceSourceNode(magmaRockSource, pos)
			lava.world.nodeRunner.AddObject(res)
			lava.numResourceSpawns++
			res.EventDestroyed.Connect(nil, func(*essenceSourceNode) {
				lava.numResourceSpawns--
			})
		})
	}
}

func (lava *lavaPuddleNode) CollidesWith(pos gmath.Vec, r float64) bool {
	offset := gmath.Vec{X: r * 0.5, Y: r * 0.5}
	objectRect := gmath.Rect{
		Min: pos.Sub(offset),
		Max: pos.Add(offset),
	}
	return lava.rect.Overlaps(objectRect)
}
