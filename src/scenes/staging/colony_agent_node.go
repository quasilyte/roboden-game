package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/viewport"
)

const (
	agentFlightHeight float64 = 40.0
	agentPickupSpeed  float64 = 40.0
)

var turretDamageTextureList = []resource.ImageID{
	assets.ImageTurretDamageMask1,
	assets.ImageTurretDamageMask2,
	assets.ImageTurretDamageMask3,
	assets.ImageTurretDamageMask4,
}

type colonyAgentKind uint8

const (
	agentWorker colonyAgentKind = iota
	agentMilitia

	// Tier2
	agentFreighter
	agentRedminer
	agentCrippler
	agentFighter
	agentPrism
	agentServo
	agentRepeller
	agentRepair
	agentRecharger
	agentGenerator
	agentMortar
	agentAntiAir

	// Tier3
	agentRefresher
	agentFlamer
	agentDestroyer

	agentKindNum

	// Buildings (not real agents/drones)
	agentGunpoint
)

type colonyAgentMode uint8

const (
	agentModeStandby colonyAgentMode = iota
	agentModeAlignStandby
	agentModeCharging
	agentModeMineEssence
	agentModeRepairBase
	agentModeRepairTurret
	agentModeReturn
	agentModePatrol
	agentModeFollow
	agentModeMove
	agentModePanic
	agentModeAttack
	agentModeMakeClone
	agentModeWaitCloning
	agentModePickup
	agentModeResourceTakeoff
	agentModeTakeoff
	agentModeRecycleReturn
	agentModeRecycleLanding
	agentModeMerging
	agentModeMergeTransform
	agentModeBuildBuilding
	agentModeGuardForever
)

type agentTraitBits uint64

const (
	traitNeverStop agentTraitBits = 1 << iota
	traitCounterClocwiseOrbiting
	traitWorkaholic
	traitDoOrDie
	traitLowHPBerserk
	traitLowHPRetreat
	traitLowHPRecycle
	traitLowHPPanic
)

type colonyAgentNode struct {
	anim       *ge.Animation
	sprite     *ge.Sprite
	shadow     *ge.Sprite
	diode      *ge.Sprite
	colonyCore *colonyCoreNode

	flashComponent damageFlashComponent

	scene *ge.Scene

	stats *agentStats

	cloningBeam *cloningBeamNode

	pos       gmath.Vec
	spritePos gmath.Vec

	traits agentTraitBits

	mode     colonyAgentMode
	waypoint gmath.Vec
	target   any

	payload         int
	cloneGen        int
	rank            int
	faction         factionTag
	cargoValue      float64
	cargoEliteValue float64
	reloadRate      float64

	height float64

	attackDelay  float64
	supportDelay float64

	maxHealth  float64
	health     float64
	maxEnergy  float64
	energy     float64
	energyBill float64

	resting bool
	speed   float64

	dist          float64
	waypointsLeft int

	EventDestroyed gsignal.Event[*colonyAgentNode]
}

func newColonyAgentNode(core *colonyCoreNode, stats *agentStats, pos gmath.Vec) *colonyAgentNode {
	a := &colonyAgentNode{
		colonyCore: core,
		stats:      stats,
		pos:        pos,
		height:     agentFlightHeight,
		reloadRate: 1,
	}
	return a
}

func (a *colonyAgentNode) AsRecipeSubject() recipeSubject {
	return recipeSubject{kind: a.stats.kind, faction: a.faction}
}

func (a *colonyAgentNode) Clone() *colonyAgentNode {
	// TODO: a clone should have the same current energy/health levels?
	if a.rank > 0 {
		panic("attempted to clone an elite unit")
	}
	cloned := newColonyAgentNode(a.colonyCore, a.stats, a.pos)
	cloned.speed = a.speed
	cloned.maxHealth = a.maxHealth
	cloned.maxEnergy = a.maxEnergy
	cloned.reloadRate = a.reloadRate
	cloned.traits = a.traits
	cloned.cloneGen = a.cloneGen + 1
	cloned.faction = a.faction
	return cloned
}

func (a *colonyAgentNode) Init(scene *ge.Scene) {
	a.scene = scene

	if a.cloneGen == 0 && !a.IsTurret() {
		// There are 64 random bits in total.
		// Every bit adds 1/64 chance (~1.5%).
		// Number of bits => chance table:
		//   1 => 50%
		//   2 => 25%
		//   3 => 12.5%
		//   4 => 6.25%
		//   5 => 3.125%
		//   6 => 1.5625%
		const (
			chance12                    = 0b111
			chance12bits                = 3
			counterClockwiseBits uint64 = chance12 << (0 * chance12bits)
			workaholicBits       uint64 = chance12 << (1 * chance12bits)
			doOrDieBits          uint64 = chance12 << (2 * chance12bits)
		)
		traitBitChance12Roll := scene.Rand().Uint64()
		if traitBitChance12Roll&counterClockwiseBits == counterClockwiseBits {
			a.traits |= traitCounterClocwiseOrbiting
		}
		if traitBitChance12Roll&workaholicBits == workaholicBits {
			a.traits |= traitWorkaholic
		}
		if traitBitChance12Roll&doOrDieBits == doOrDieBits {
			a.traits |= traitDoOrDie
		}

		if scene.Rand().Chance(0.4) {
			a.traits |= traitNeverStop
		}

		// These trait bits can't be combined.
		// Only one of them will take place.
		roll := scene.Rand().Float()
		switch {
		case roll < 0.10:
			// 10% for retreat.
			a.traits |= traitLowHPRetreat
		case roll < 0.20:
			// 10% for recycle.
			if a.stats.tier == 1 {
				a.traits |= traitLowHPRecycle
			}
		case roll < 0.25:
			// 5% for berserk.
			a.traits |= traitLowHPBerserk
		case roll < 0.30:
			// 5% for panic.
			a.traits |= traitLowHPPanic
		}

		switch a.faction {
		case redFactionTag:
			a.maxHealth *= 1.4
		case greenFactionTag:
			a.speed *= 1.2
		case blueFactionTag:
			a.maxEnergy *= 1.8
		}
	}

	if a.cloneGen == 0 {
		a.maxHealth = a.stats.maxHealth * scene.Rand().FloatRange(0.9, 1.1)
		a.maxEnergy = scene.Rand().FloatRange(120, 200)
		a.speed = a.stats.speed * scene.Rand().FloatRange(0.8, 1.1)
		a.applyRankBonuses()
	}

	a.health = a.maxHealth
	a.energy = a.maxEnergy

	a.sprite = scene.NewSprite(a.stats.image)
	a.sprite.Pos.Base = &a.spritePos
	if a.IsFlying() {
		a.camera().AddGraphicsAbove(a.sprite)
	} else {
		a.camera().AddGraphics(a.sprite)
		a.sprite.Shader = scene.NewShader(assets.ShaderColonyDamage)
		a.sprite.Shader.SetFloatValue("HP", 1.0)
		damageTexture := gmath.RandElem(scene.Rand(), turretDamageTextureList)
		a.sprite.Shader.Texture1 = scene.LoadImage(damageTexture)
	}

	a.flashComponent.sprite = a.sprite

	if a.faction != neutralFactionTag {
		a.diode = scene.NewSprite(assets.ImageFactionDiode)
		a.diode.Pos.Base = &a.spritePos
		a.diode.Pos.Offset.Y = a.stats.diodeOffset
		var colorScale ge.ColorScale
		colorScale.SetColor(factionByTag(a.faction).color)
		a.diode.SetColorScale(colorScale)
		a.camera().AddGraphicsAbove(a.diode)
	}

	if !a.IsTurret() {
		shadowImage := assets.ImageSmallShadow
		switch a.stats.size {
		case sizeMedium:
			shadowImage = assets.ImageMediumShadow
		case sizeLarge:
			shadowImage = assets.ImageBigShadow
		}
		a.shadow = scene.NewSprite(shadowImage)
		a.shadow.Pos.Base = &a.spritePos
		a.camera().AddGraphics(a.shadow)
	}

	a.anim = ge.NewRepeatedAnimation(a.sprite, -1)
	a.anim.Tick(scene.Rand().FloatRange(0, 0.7))
	a.anim.SetOffsetY(float64(a.rank) * a.sprite.FrameHeight)
}

func (a *colonyAgentNode) IsDisposed() bool { return a.sprite.IsDisposed() }

func (a *colonyAgentNode) IsTurret() bool {
	return a.stats.kind == agentGunpoint
}

func (a *colonyAgentNode) applyRankBonuses() {
	switch a.rank {
	case 0:
		// A normal unit. No bonuses.

	case 1:
		// An elite unit.
		a.maxHealth *= 1.15
		a.speed *= 1.15
		a.maxEnergy *= 1.4
		a.reloadRate = 1.3 // +30% attack/special reload speed

	case 2:
		// A super elite unit.
		a.maxHealth *= 1.5
		a.speed *= 1.2
		a.maxEnergy *= 2.0
		a.reloadRate = 1.6 // +60% attack/special reload speed
	}
}

func (a *colonyAgentNode) AssignMode(mode colonyAgentMode, pos gmath.Vec, target any) bool {
	if a.IsTurret() {
		panic("assigning a mode to a turret")
	}

	switch mode {
	case agentModePatrol:
		a.mode = mode
		a.dist = a.colonyCore.PatrolRadius()
		a.waypoint = a.orbitingWaypoint()
		a.waypointsLeft = a.scene.Rand().IntRange(40, 70)
		return true

	case agentModeWaitCloning:
		a.mode = mode
		a.target = target
		a.waypoint = gmath.Vec{}
		return true

	case agentModeMakeClone:
		a.mode = mode
		a.target = target
		a.dist = a.scene.Rand().FloatRange(1.2, 2) // cloning time
		a.energyBill += 20
		targetPos := target.(*colonyAgentNode).pos
		a.waypoint = a.pos.DirectionTo(targetPos).Mulf(110).Add(targetPos).Add(a.scene.Rand().Offset(-20, 20))
		return true

	case agentModeMerging:
		a.mode = mode
		a.target = target
		a.dist = a.scene.Rand().FloatRange(5, 6) // merging time
		return true

	case agentModeAlignStandby:
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
			a.cloningBeam = nil
		}
		a.mode = mode
		a.waypoint = a.pos.Sub(gmath.Vec{Y: agentFlightHeight - a.height})
		return true

	case agentModeMove:
		a.mode = mode
		a.waypoint = pos
		return true

	case agentModePanic:
		a.mode = mode
		a.waypoint = a.pos
		a.waypointsLeft = a.scene.Rand().IntRange(4, 9)
		return true

	case agentModeStandby:
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
			a.cloningBeam = nil
		}
		a.mode = mode
		maxDist := a.colonyCore.realRadius
		if !a.stats.canPatrol {
			maxDist *= 0.65
		}
		a.dist = a.scene.Rand().FloatRange(40, maxDist)
		a.waypoint = a.orbitingWaypoint()
		a.waypointsLeft = a.scene.Rand().IntRange(30, 60)
		return true

	case agentModeFollow, agentModeAttack:
		isPatrol := a.mode == agentModePatrol
		a.mode = agentModeFollow // attack is a long-range follow
		a.target = target
		a.waypoint = a.followWaypoint(target.(*creepNode).pos)
		if isPatrol {
			a.waypointsLeft = a.scene.Rand().IntRange(5, 7)
		} else {
			a.waypointsLeft = a.scene.Rand().IntRange(7, 9)
		}
		if mode == agentModeAttack {
			a.waypointsLeft += 5
			if a.hasTrait(traitDoOrDie) {
				a.waypointsLeft += 5
			}
		}
		return true

	case agentModeCharging:
		a.mode = mode
		a.waypoint = gmath.Vec{}
		return true

	case agentModeMineEssence:
		if !a.stats.canGather {
			return false
		}
		source := target.(*essenceSourceNode)
		if source.stats == redOilSource && a.stats.kind != agentRedminer {
			return false
		}
		energyCost := source.pos.DistanceTo(a.pos) / 2
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.energyBill += energyCost
		a.mode = mode
		a.waypoint = roundedPos(source.pos.Sub(gmath.Vec{Y: agentFlightHeight}).Add(a.scene.Rand().Offset(-8, 8)))
		a.target = target
		return true

	case agentModeTakeoff:
		a.mode = mode
		a.waypoint = a.pos.Sub(gmath.Vec{Y: agentFlightHeight})
		return true

	case agentModePickup:
		a.mode = mode
		a.waypoint = a.pos.Add(gmath.Vec{Y: agentFlightHeight})
		return true

	case agentModeRecycleReturn:
		a.mode = mode
		a.waypoint = a.colonyCore.GetEntrancePos().Sub(gmath.Vec{Y: agentFlightHeight})
		return true

	case agentModeRecycleLanding:
		a.mode = mode
		a.waypoint = a.colonyCore.GetEntrancePos()
		return true

	case agentModeRepairTurret:
		energyCost := 40.0
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.target = target
		a.mode = mode
		a.energyBill += energyCost
		a.dist = a.scene.Rand().FloatRange(3, 4) // repair time
		a.waypoint = gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(target.(*colonyAgentNode).pos)
		return true

	case agentModeRepairBase:
		energyCost := 40.0
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.mode = mode
		a.energyBill += energyCost
		a.dist = a.scene.Rand().FloatRange(3, 4) // repair time
		a.waypoint = gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(a.colonyCore.pos)
		return true

	case agentModeBuildBuilding:
		construction := target.(*constructionNode)
		energyCost := construction.pos.DistanceTo(a.pos) * 0.6
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.mode = mode
		a.energyBill += energyCost
		a.dist = a.scene.Rand().FloatRange(5, 7) // build time
		a.target = target
		a.waypoint = gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(construction.pos)
		return true
	}

	return false
}

func (a *colonyAgentNode) orbitingWaypoint() gmath.Vec {
	var direction gmath.Vec
	if a.pos == a.colonyCore.pos {
		direction = gmath.RadToVec(a.scene.Rand().Rad())
	} else {
		direction = a.pos.DirectionTo(a.colonyCore.pos)
	}
	dir := gmath.Rad(0.4)
	if a.hasTrait(traitCounterClocwiseOrbiting) {
		dir = -0.4
	}
	return direction.Rotated(dir).Mulf(a.dist).Add(a.colonyCore.pos)
}

func (a *colonyAgentNode) Update(delta float64) {
	a.anim.Tick(delta)
	a.flashComponent.Update(delta)

	if a.shadow != nil {
		a.shadow.Pos.Offset.Y = math.Round(a.height + 4)
		newShadowAlpha := float32(1.0 - ((a.height / agentFlightHeight) * 0.5))
		a.shadow.SetAlpha(newShadowAlpha)
	}

	// FIXME: this should be fixed in the ge package.
	a.spritePos.X = math.Round(a.pos.X)
	a.spritePos.Y = math.Round(a.pos.Y)

	if a.energyBill != 0 {
		a.energy -= delta * 2
		a.energyBill = gmath.ClampMin(a.energyBill-delta*2, 0)
	}

	if a.resting {
		a.energy = gmath.ClampMax(a.energy+(delta*0.5), a.maxEnergy)
		if a.energy > a.maxEnergy*0.6 {
			a.resting = false
		}
	} else {
		if a.mode != agentModeStandby && a.mode != agentModeCharging && a.energy < a.maxEnergy*0.5 {
			a.resting = true
		}
	}

	a.processAttack(delta)
	a.processSupport(delta)

	switch a.mode {
	case agentModeStandby:
		a.updateStandby(delta)
	case agentModeAlignStandby:
		a.updateAlignStandby(delta)
	case agentModeCharging:
		a.updateCharging(delta)
	case agentModeMineEssence:
		a.updateMineEssence(delta)
	case agentModePickup:
		a.updatePickup(delta)
	case agentModeReturn:
		a.updateReturn(delta)
	case agentModePatrol:
		a.updatePatrol(delta)
	case agentModeMove:
		a.updateMove(delta)
	case agentModePanic:
		a.updatePanic(delta)
	case agentModeFollow:
		a.updateFollow(delta)
	case agentModeWaitCloning:
		a.updateWaitCloning(delta)
	case agentModeMakeClone:
		a.updateMakeClone(delta)
	case agentModeMerging:
		a.updateMerging(delta)
	case agentModeResourceTakeoff:
		a.updateResourceTakeoff(delta)
	case agentModeTakeoff:
		a.updateTakeoff(delta)
	case agentModeRecycleReturn:
		a.updateRecycleReturn(delta)
	case agentModeRecycleLanding:
		a.updateRecycleLanding(delta)
	case agentModeBuildBuilding:
		a.updateBuildBase(delta)
	case agentModeRepairBase:
		a.updateRepairBase(delta)
	case agentModeRepairTurret:
		a.updateRepairTurret(delta)
	case agentModeGuardForever:
		// Just chill.
	}
}

func (a *colonyAgentNode) Dispose() {
	a.sprite.Dispose()
	if a.shadow != nil {
		a.shadow.Dispose()
	}
	if a.diode != nil {
		a.diode.Dispose()
	}
	if a.cloningBeam != nil {
		a.cloningBeam.Dispose()
		a.cloningBeam = nil
	}
}

func (a *colonyAgentNode) Destroy() {
	a.EventDestroyed.Emit(a)
	a.Dispose()
}

func (a *colonyAgentNode) IsFlying() bool {
	return !a.IsTurret()
}

func (a *colonyAgentNode) ReceiveEnergyDamage(damage float64) {
	a.energy = gmath.ClampMin(a.energy-damage, 0)
}

func (a *colonyAgentNode) explode() {
	if a.IsTurret() {
		createAreaExplosion(a.scene, a.camera(), spriteRect(a.pos, a.sprite), true)
		scraps := a.colonyCore.world.NewEssenceSourceNode(scrapSource, a.pos.Add(gmath.Vec{Y: 2}))
		a.scene.AddObject(scraps)
		return
	}

	playSound(a.scene, a.camera(), assets.AudioAgentDestroyed, a.pos)
	a.colonyCore.actionPriorities.AddWeight(prioritySecurity, 0.05)
	a.colonyCore.actionPriorities.AddWeight(priorityGrowth, 0.01)

	roll := a.scene.Rand().Float()
	if roll < 0.3 {
		createExplosion(a.scene, a.camera(), true, a.pos)
	} else {
		var scraps *essenceSourceStats
		if roll > 0.6 {
			scraps = smallScrapSource
			if a.stats.size != sizeSmall {
				scraps = scrapSource
			}
		}
		fall := newDroneFallNode(a.colonyCore.world, scraps, a.stats.image, a.shadow.ImageID(), a.pos, a.height)
		fall.FrameOffsetY = float64(a.rank) * a.sprite.FrameHeight
		a.scene.AddObject(fall)
	}
}

func (a *colonyAgentNode) OnBuildingRepair(amount float64) {
	// Turrets are 2 times easier to repair than a base, hence x2 multiplier.
	amount *= 2
	if a.health >= a.maxHealth {
		return
	}
	a.health = gmath.ClampMax(a.health+amount, a.maxHealth)
	a.updateHealthShader()
}

func (a *colonyAgentNode) updateHealthShader() {
	percentage := a.health / a.maxHealth
	a.sprite.Shader.SetFloatValue("HP", percentage)
	a.sprite.Shader.Enabled = percentage < 0.95
}

func (a *colonyAgentNode) CanAttack(mask targetKind) bool {
	return a.stats.weapon.TargetFlags&mask != 0
}

func (a *colonyAgentNode) OnDamage(damage damageValue, source gmath.Vec) {
	a.health -= damage.health

	if a.health < 0 {
		a.explode()
		a.Destroy()
	}

	if a.health <= 8 {
		switch {
		case a.hasTrait(traitLowHPBerserk):
			// Berserks go straight into the danger when low on health.
			if a.mode != agentModeMove {
				a.AssignMode(agentModeMove, source.Add(a.scene.Rand().Offset(-20, 20)), nil)
			}
		case a.hasTrait(traitLowHPRecycle):
			// Recycle agents may go to recycle themselves on low health.
			if a.colonyCore.mode == colonyModeNormal && a.mode != agentModeRecycleReturn && a.mode != agentModeRecycleLanding {
				if a.scene.Rand().Chance(0.8) {
					a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
				}
			}
		case a.hasTrait(traitLowHPRetreat):
			// Agents with retreat trait will try to fly away from a threat on low health.
			if a.mode != agentModeMove {
				pos := retreatPos(a.scene.Rand(), a.scene.Rand().FloatRange(80, 140), a.pos, source)
				a.AssignMode(agentModeMove, pos, nil)
			}
		case a.hasTrait(traitLowHPPanic):
			// Agents with panic trait will stop what they're doing and fly like crazy.
			if a.mode != agentModePanic {
				a.AssignMode(agentModePanic, gmath.Vec{}, nil)
			}
		}

	}

	if damage.health != 0 {
		a.flashComponent.flash = 0.2
		if a.IsTurret() {
			a.updateHealthShader()
		}
	}

	a.energy = gmath.ClampMin(a.energy-damage.energy, 0)

	if !a.IsTurret() {
		if a.colonyCore.GetSecurityPriority() < 0.65 && a.scene.Rand().Chance(1.0-a.colonyCore.GetSecurityPriority()) {
			a.colonyCore.actionPriorities.AddWeight(prioritySecurity, 0.01)
		}
	}
}

func (a *colonyAgentNode) GetPos() *gmath.Vec { return &a.pos }

func (a *colonyAgentNode) GetVelocity() gmath.Vec {
	if a.waypoint.IsZero() {
		return gmath.Vec{}
	}
	return a.pos.VecTowards(a.waypoint, a.movementSpeed())
}

func (a *colonyAgentNode) processSupport(delta float64) {
	switch a.stats.kind {
	case agentRepair, agentRecharger, agentRefresher:
		// OK
	default:
		return
	}

	a.supportDelay = gmath.ClampMin(a.supportDelay-(delta*a.reloadRate), 0)

	if a.supportDelay != 0 {
		if a.stats.kind == agentRefresher {
			a.attackDelay = gmath.ClampMin(a.attackDelay-(delta*a.reloadRate), 0)
			if a.attackDelay != 0 {
				return
			}
			a.attackDelay = repairAgentStats.supportReload * a.scene.Rand().FloatRange(0.7, 1.4)
			a.doRepair()
		}
		return
	}

	a.supportDelay = a.stats.supportReload * a.scene.Rand().FloatRange(0.7, 1.4)

	switch a.stats.kind {
	case agentRecharger, agentRefresher:
		a.doRecharge()
	case agentRepair:
		a.doRepair()
	}
}

func (a *colonyAgentNode) doRecharge() {
	const rechargerEnergyRecorery float64 = 25.0
	target := a.colonyCore.agents.Find(searchWorkers|searchFighters|searchRandomized, func(x *colonyAgentNode) bool {
		return x != a &&
			(x.energy+rechargerEnergyRecorery) < x.maxEnergy &&
			x.pos.DistanceTo(a.pos) < rechargeAgentStats.supportRange
	})
	if target != nil {
		beam := newBeamNode(a.camera(), ge.Pos{Base: &a.pos}, ge.Pos{Base: &target.pos}, rechargerBeamColor)
		beam.width = 2
		target.energy = gmath.ClampMax(target.energy+rechargerEnergyRecorery, target.maxEnergy)
		a.scene.AddObject(beam)
		playSound(a.scene, a.camera(), assets.AudioRechargerBeam, a.pos)
	}
}

func (a *colonyAgentNode) doRepair() {
	target := a.colonyCore.agents.Find(searchWorkers|searchFighters|searchRandomized, func(x *colonyAgentNode) bool {
		return x != a &&
			x.health < x.maxHealth &&
			x.pos.DistanceTo(a.pos) < repairAgentStats.supportRange
	})
	if target != nil {
		beam := newBeamNode(a.camera(), ge.Pos{Base: &a.pos}, ge.Pos{Base: &target.pos}, repairBeamColor)
		beam.width = 2
		target.health = gmath.ClampMax(target.health+3, target.maxHealth)
		a.scene.AddObject(beam)
		playSound(a.scene, a.camera(), assets.AudioRepairBeam, a.pos)
	}
}

func (a *colonyAgentNode) processAttack(delta float64) {
	if !a.stats.canPatrol {
		return
	}

	a.attackDelay = gmath.ClampMin(a.attackDelay-(delta*a.reloadRate), 0)
	if a.attackDelay != 0 {
		return
	}

	a.attackDelay = a.stats.weapon.Reload * a.scene.Rand().FloatRange(0.8, 1.2)

	targets := a.colonyCore.world.tmpTargetSlice[:0]
	for _, c := range a.colonyCore.world.creeps {
		if len(targets) >= a.stats.weapon.MaxTargets {
			break
		}
		if !a.CanAttack(c.TargetKind()) {
			continue
		}
		if c.pos.DistanceTo(a.pos) >= a.stats.weapon.AttackRange {
			continue
		}
		targets = append(targets, c)
	}
	if len(targets) == 0 {
		return
	}

	switch a.stats.kind {
	case agentDestroyer:
		target := targets[0]
		offset := gmath.Vec{X: -7, Y: 2}
		offsetStep := gmath.Vec{X: 14}
		targetOffset := gmath.Vec{X: -4}
		targetOffsetStep := gmath.Vec{X: 8}
		for i := 0; i < 2; i++ {
			pos1 := ge.Pos{Base: &a.pos, Offset: offset}
			pos2 := ge.Pos{Base: target.GetPos(), Offset: targetOffset}
			beam := newBeamNode(a.camera(), pos1, pos2, destroyerBeamColor)
			beam.width = 2
			a.scene.AddObject(beam)
			offset = offset.Add(offsetStep)
			targetOffset = targetOffset.Add(targetOffsetStep)
		}
		target.OnDamage(a.stats.weapon.Damage, a.pos)

	case agentPrism:
		target := targets[0]
		damage := a.stats.weapon.Damage
		width := 1.0
		numReflections := 0
		pos := &a.pos
		a.colonyCore.agents.Find(searchFighters|searchRandomized, func(ally *colonyAgentNode) bool {
			if ally.stats.kind != agentPrism || ally == a {
				return false
			}
			if ally.pos.DistanceSquaredTo(*pos) > (196 * 196) {
				return false
			}
			ally.attackDelay += float64(numReflections) * 0.3
			beam := newBeamNode(a.camera(), ge.Pos{Base: pos}, ge.Pos{Base: &ally.pos}, prismBeamColors[numReflections])
			beam.width = width
			a.scene.AddObject(beam)
			numReflections++
			damage.health++
			width++
			pos = &ally.pos
			return numReflections >= 3
		})
		beam := newBeamNode(a.camera(), ge.Pos{Base: pos}, ge.Pos{Base: target.GetPos()}, prismBeamColors[numReflections])
		beam.width = width
		a.scene.AddObject(beam)
		target.OnDamage(damage, a.pos)

	default:
		for _, target := range targets {
			toPos := snipePos(a.stats.weapon.ProjectileSpeed, a.pos, *target.GetPos(), target.GetVelocity())
			for i := 0; i < a.stats.weapon.BurstSize; i++ {
				fireDelay := float64(i) * a.stats.weapon.BurstDelay
				p := newProjectileNode(projectileConfig{
					Camera:    a.colonyCore.world.camera,
					Weapon:    a.stats.weapon,
					FromPos:   &a.pos,
					ToPos:     toPos,
					Target:    target,
					FireDelay: fireDelay,
				})
				a.scene.AddObject(p)
			}
		}
	}

	playSound(a.scene, a.camera(), a.stats.weapon.AttackSound, a.pos)
}

func (a *colonyAgentNode) movementSpeed() float64 {
	switch a.mode {
	case agentModeTakeoff, agentModeRecycleLanding:
		return 30
	case agentModePickup, agentModeResourceTakeoff, agentModeAlignStandby:
		return agentPickupSpeed
	}
	if a.resting {
		return a.speed * 0.5
	}
	return a.speed
}

func (a *colonyAgentNode) moveTowards(delta float64, pos gmath.Vec) bool {
	travelled := a.movementSpeed() * delta
	if a.pos.DistanceTo(pos) <= travelled {
		a.pos = pos
		return true
	}
	a.pos = a.pos.MoveTowards(pos, travelled)
	return false
}

func (a *colonyAgentNode) updatePatrol(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		a.dist = a.colonyCore.PatrolRadius()
		a.waypoint = a.orbitingWaypoint()
		a.waypointsLeft--
		if a.waypointsLeft == 0 {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
	}
}

func (a *colonyAgentNode) updateWaitCloning(delta float64) {
	cloner := a.target.(*colonyAgentNode)
	if cloner.IsDisposed() {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
}

func (a *colonyAgentNode) updateTakeoff(delta float64) {
	a.height += delta * 30
	if a.moveTowards(delta, a.waypoint) {
		a.height = agentFlightHeight
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateRecycleReturn(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		a.colonyCore.openHatchTime = 1.5
		a.AssignMode(agentModeRecycleLanding, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateRepairBase(delta float64) {
	if !a.waypoint.IsZero() {
		if a.moveTowards(delta, a.waypoint) {
			a.waypoint = gmath.Vec{}
			buildPos := ge.Pos{
				Base:   &a.colonyCore.pos,
				Offset: gmath.Vec{X: a.scene.Rand().FloatRange(-18, 18)},
			}
			beam := newCloningBeamNode(a.colonyCore.world.camera, false, &a.pos, buildPos)
			a.cloningBeam = beam
			a.scene.AddObject(beam)
			return
		}
		return
	}
	a.dist -= delta
	if a.dist <= 0 {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		a.colonyCore.OnHeal(a.scene.Rand().FloatRange(3, 5))
		return
	}
}

func (a *colonyAgentNode) updateRepairTurret(delta float64) {
	target := a.target.(*colonyAgentNode)
	if !a.waypoint.IsZero() {
		if a.moveTowards(delta, a.waypoint) {
			a.waypoint = gmath.Vec{}
			buildPos := ge.Pos{
				Base:   &target.pos,
				Offset: gmath.Vec{X: a.scene.Rand().FloatRange(-10, 10)},
			}
			beam := newCloningBeamNode(a.colonyCore.world.camera, false, &a.pos, buildPos)
			a.cloningBeam = beam
			a.scene.AddObject(beam)
			return
		}
		return
	}
	a.dist -= delta
	if a.dist <= 0 {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		amountRepaired := a.scene.Rand().FloatRange(3, 5)
		if a.faction == greenFactionTag {
			amountRepaired *= 1.5
		}
		target.OnBuildingRepair(amountRepaired)
		return
	}
}

func (a *colonyAgentNode) updateBuildBase(delta float64) {
	target := a.target.(*constructionNode)
	if target.IsDisposed() {
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
		}
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
	if !a.waypoint.IsZero() {
		if a.moveTowards(delta, a.waypoint) {
			target.attention += 2.5
			a.waypoint = gmath.Vec{}
			buildPos := target.GetConstructPos()
			beam := newCloningBeamNode(a.colonyCore.world.camera, false, &a.pos, buildPos)
			a.cloningBeam = beam
			a.scene.AddObject(beam)
			return
		}
		return
	}
	amountConstructed := delta
	if a.faction == greenFactionTag {
		amountConstructed *= 1.5
	}
	if target.Construct(amountConstructed, a.colonyCore) {
		return
	}
	a.dist -= delta
	if a.dist <= 0 || a.energy < 20 {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
}

func (a *colonyAgentNode) updateRecycleLanding(delta float64) {
	height := a.height
	a.height -= delta * 30
	if height >= 3 && a.height < 3 {
		a.sprite.SetColorScaleRGBA(200, 200, 200, 255)
	}
	if a.moveTowards(delta, a.waypoint) {
		a.colonyCore.resources += a.stats.cost * 0.9
		if a.rank != 0 {
			a.colonyCore.eliteResources += float64(a.rank)
		}
		playSound(a.scene, a.camera(), assets.AudioAgentRecycled, a.pos)
		a.Destroy()
	}
}

func (a *colonyAgentNode) updateMerging(delta float64) {
	target := a.target.(*colonyAgentNode)
	if target.IsDisposed() {
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
		}
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
	if a.waypoint.IsZero() {
		dist := target.pos.DistanceTo(a.pos)
		if dist > 64 {
			a.waypoint = a.pos.MoveTowards(target.pos, dist-20).Add(a.scene.Rand().Offset(-8, 8))
			return
		}
	}
	if !a.waypoint.IsZero() {
		if a.moveTowards(delta, a.waypoint) {
			a.waypoint = gmath.Vec{}
		}
		return
	}
	if a.cloningBeam == nil {
		beam := newCloningBeamNode(a.colonyCore.world.camera, true, &a.pos, ge.Pos{Base: &target.pos})
		a.cloningBeam = beam
		a.scene.AddObject(beam)
	}
	a.dist -= delta
	if a.pos.DistanceTo(target.pos) > 10 {
		a.pos = a.pos.MoveTowards(target.pos, delta*10)
	} else {
		// Merging is x2 faster when units are next to each other.
		a.dist -= delta
	}
	if a.dist <= 0 {
		a.cloningBeam.Dispose()
		a.cloningBeam = nil
		newStats := mergeAgents(a, target)
		newAgent := a.colonyCore.NewColonyAgentNode(newStats, target.pos)
		var newFaction factionTag
		rankScore := a.rank + target.rank
		switch rankScore {
		case 0:
			// Two normal units => normal unit.
		case 1:
			// Only one elite unit => a chance to get an elite.
			if a.scene.Rand().Chance(0.75) {
				newAgent.rank = 1
			}
		case 2:
			// Only one super elite unit or two elite units => a super elite or normal elite.
			if a.scene.Rand().Chance(0.75) {
				newAgent.rank = 2
			} else {
				newAgent.rank = 1
			}
		default:
			// Anything better is capped at rank 2.
			newAgent.rank = 2
		}
		if newStats.tier == 2 {
			newFaction = a.colonyCore.pickAgentFaction()
		} else {
			newFaction = a.faction
			if newFaction == neutralFactionTag || (target.faction != neutralFactionTag && a.faction != target.faction && a.scene.Rand().Bool()) {
				newFaction = target.faction
			}
		}
		newAgent.faction = newFaction
		a.scene.AddObject(newAgent)
		newAgent.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		target.Destroy()
		a.Destroy()
		return
	}
}

func (a *colonyAgentNode) updateMakeClone(delta float64) {
	target := a.target.(*colonyAgentNode)
	if target.IsDisposed() {
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
		}
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
	if !a.waypoint.IsZero() {
		if a.moveTowards(delta, a.waypoint) {
			a.waypoint = gmath.Vec{}
			beam := newCloningBeamNode(a.colonyCore.world.camera, false, &a.pos, ge.Pos{Base: &target.pos})
			a.cloningBeam = beam
			a.scene.AddObject(beam)
			return
		}
		return
	}
	a.dist -= delta
	if a.dist <= 0 {
		a.cloningBeam.Dispose()
		a.cloningBeam = nil
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		target.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		clone := a.colonyCore.CloneAgentNode(target)
		a.scene.AddObject(clone)
		clone.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
}

func (a *colonyAgentNode) followWaypoint(targetPos gmath.Vec) gmath.Vec {
	preferredDist := gmath.ClampMin(a.stats.weapon.AttackRange*0.6, 80)
	return a.pos.DirectionTo(targetPos).Mulf(preferredDist).Add(targetPos).Add(a.scene.Rand().Offset(-52, 52))
}

func (a *colonyAgentNode) updateMove(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updatePanic(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		a.waypointsLeft--
		a.waypoint = gmath.Vec{}
	}

	if a.waypoint.IsZero() {
		if a.waypointsLeft <= 0 {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		waypoint := a.pos.Add(a.scene.Rand().Offset(-32, 32))
		a.waypoint = correctedPos(a.colonyCore.world.rect, waypoint, 64)
	}
}

func (a *colonyAgentNode) updateFollow(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		target := a.target.(*creepNode)
		if a.waypointsLeft == 0 || target.IsDisposed() {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		a.waypointsLeft--
		a.waypoint = a.followWaypoint(target.pos)
	}
}

func (a *colonyAgentNode) updateAlignStandby(delta float64) {
	a.height += delta * agentPickupSpeed
	if a.moveTowards(delta, a.waypoint) {
		a.height = agentFlightHeight
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateStandby(delta float64) {
	a.energy += delta * 0.5
	if a.moveTowards(delta, a.waypoint) {
		if a.waypointsLeft == 0 {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		a.waypointsLeft--
		a.waypoint = a.orbitingWaypoint()
		if !a.hasTrait(traitNeverStop) && a.energy < 40 && a.scene.Rand().Chance(0.2) {
			a.AssignMode(agentModeCharging, gmath.Vec{}, nil)
			return
		}
	}
}

func (a *colonyAgentNode) updateCharging(delta float64) {
	a.energy += delta * 3
	if a.energy >= 50 {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateMineEssence(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		source := a.target.(*essenceSourceNode)
		if source.IsDisposed() {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		} else {
			a.AssignMode(agentModePickup, gmath.Vec{}, nil)
		}
	}
}

func (a *colonyAgentNode) updatePickup(delta float64) {
	a.height -= delta * agentPickupSpeed
	if a.moveTowards(delta, a.waypoint) {
		a.height = 0
		a.mode = agentModeResourceTakeoff
		a.waypoint = a.pos.Sub(gmath.Vec{Y: agentFlightHeight})
		source := a.target.(*essenceSourceNode)
		harvested := source.Harvest(a.maxPayload())
		a.payload = harvested
		a.cargoValue = float64(harvested) * source.stats.value
		a.cargoEliteValue = float64(harvested) * source.stats.eliteValue
	}
}

func (a *colonyAgentNode) updateResourceTakeoff(delta float64) {
	a.height += delta * agentPickupSpeed
	if a.moveTowards(delta, a.waypoint) {
		a.height = agentFlightHeight
		entranceNum := a.scene.Rand().IntRange(0, 2)
		a.waypoint = a.colonyCore.GetStoragePos().Add(gmath.Vec{Y: float64(entranceNum) * 8})
		a.mode = agentModeReturn
	}
}

func (a *colonyAgentNode) updateReturn(delta float64) {
	if a.moveTowards(delta, a.waypoint) {
		if a.payload != 0 {
			a.colonyCore.resources += a.cargoValue
			a.colonyCore.world.result.ResourcesGathered += a.cargoValue
			a.colonyCore.eliteResources += a.cargoEliteValue
			a.colonyCore.world.result.EliteResourcesGathered = a.cargoEliteValue
			a.payload = 0
			a.cargoValue = 0
			a.cargoEliteValue = 0
			playSound(a.scene, a.camera(), assets.AudioEssenceCollected, a.pos)
		}
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) camera() *viewport.Camera {
	return a.colonyCore.world.camera
}

func (a *colonyAgentNode) hasTrait(t agentTraitBits) bool {
	return a.traits&t != 0
}

func (a *colonyAgentNode) maxPayload() int {
	n := a.stats.maxPayload
	if a.faction == yellowFactionTag {
		n++
	}
	return n
}
