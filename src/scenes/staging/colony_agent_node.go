package staging

import (
	"fmt"
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"

	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/pathing"
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

var stunnableModes = [256]bool{
	agentModeStandby:         true,
	agentModeMineEssence:     true,
	agentModeCourierFlight:   true,
	agentModeScavenge:        true,
	agentModeReturn:          true,
	agentModeRepairBase:      true,
	agentModeRepairTurret:    true,
	agentModePatrol:          true,
	agentModeMove:            true,
	agentModeFollow:          true,
	agentModeFollowCommander: true,
	agentModePanic:           true,
	agentModeRecycleReturn:   true,
	agentModeBuildBuilding:   true,
}

type colonyAgentMode uint8

const (
	agentModeStandby colonyAgentMode = iota
	agentModeAlignStandby
	agentModeCharging
	agentModePosing
	agentModeForcedCharging
	agentModeMineEssence
	agentModeMineSulfurEssence
	agentModeCourierFlight
	agentModeScavenge
	agentModeRepairBase
	agentModeRepairTurret
	agentModeCaptureBuilding
	agentModeReturn
	agentModePatrol
	agentModeFollow
	agentModeMove
	agentModeCloakHide
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
	agentModeMergingRoomba
	agentModeBuildBuilding
	agentModeGuardForever
	agentModeSiegeGuard
	agentModeHarvester
	agentModeRoombaGuard
	agentModeRoombaAttack
	agentModeRoombaPatrol
	agentModeRoombaCombatWait
	agentModeRoombaWait
	agentModeKamikazeAttack
	agentModeConsumeDrone
	agentModeFollowCommander
	agentModeBomberAttack

	agentModeRelictDroneFactory
	agentModeRelictPatrol
	agentModeRelictTakeoff
	agentModeRelictRoomba
)

type agentTraitBits uint64

const (
	traitNeverStop agentTraitBits = 1 << iota
	traitCounterClocwiseOrbiting
	traitWorkaholic
	traitDoOrDie
	traitAdventurer
	traitLowHPBerserk
	traitLowHPRetreat
	traitLowHPRecycle
	traitLowHPPanic
)

type colonyAgentNode struct {
	anim       *ge.Animation
	sprite     *ge.Sprite
	diode      *ge.Sprite
	colonyCore *colonyCoreNode

	flashComponent damageFlashComponent

	scene *ge.Scene

	stats *gamedata.AgentStats

	cloningBeam *cloningBeamNode

	shadowComponent shadowComponent

	pos gmath.Vec

	traits agentTraitBits
	path   pathing.GridPath

	mode     colonyAgentMode
	waypoint gmath.Vec
	dir      gmath.Vec
	target   any

	payload         int
	cloneGen        int
	rank            int
	extraLevel      int // Devourer level; drone factory units num
	commanderID     int
	faction         gamedata.FactionTag
	cargoValue      float64
	cargoEliteValue float64
	reloadRate      float64
	healthRegen     float64

	attackDelay  float64
	supportDelay float64
	specialDelay float64
	cloaking     float64

	maxHealth       float64
	health          float64
	maxEnergy       float64
	energy          float64
	energyBill      float64
	energyRegenRate float64
	slow            float64
	lifetime        float64
	energyTarget    float64

	insideForest bool
	tether       bool
	resting      bool
	disposed     bool
	speed        float64

	dist          float64
	waypointsLeft int

	EventDestroyed gsignal.Event[*colonyAgentNode]
}

func newColonyAgentNode(core *colonyCoreNode, stats *gamedata.AgentStats, pos gmath.Vec) *colonyAgentNode {
	a := &colonyAgentNode{
		colonyCore:      core,
		stats:           stats,
		pos:             pos,
		reloadRate:      1,
		energyRegenRate: 1,
	}
	return a
}

func (a *colonyAgentNode) AsRecipeSubject() gamedata.RecipeSubject {
	return gamedata.RecipeSubject{Kind: a.stats.Kind, Faction: a.faction}
}

func (a *colonyAgentNode) Clone() *colonyAgentNode {
	// TODO: a clone should have the same current energy/health levels?
	if a.rank > 0 {
		panic("attempted to clone an elite unit")
	}
	cloned := newColonyAgentNode(a.colonyCore, a.stats, a.pos)
	cloned.speed = a.speed
	if a.stats.Kind == gamedata.AgentDevourer {
		cloned.extraLevel = a.extraLevel
	}
	cloned.maxHealth = a.maxHealth
	cloned.maxEnergy = a.maxEnergy
	cloned.reloadRate = a.reloadRate
	cloned.traits = a.traits
	cloned.cloneGen = a.cloneGen + 1
	cloned.faction = a.faction
	cloned.energyRegenRate = a.energyRegenRate
	cloned.healthRegen = a.healthRegen
	return cloned
}

func (a *colonyAgentNode) Init(scene *ge.Scene) {
	a.scene = scene

	if a.stats.Tier == 1 {
		a.lifetime = scene.Rand().FloatRange(1.5*60, 3*60)
		// If it's a neutral drone, don't hurry to recycle it.
		// It's probably a new base and it may need drones to live for longer.
		// If evolution priority is high, neutral drones will be recycled anyway.
		if a.faction == gamedata.NeutralFactionTag {
			a.lifetime *= 2
		}
	}

	if a.cloneGen == 0 {
		a.energyRegenRate = 1 + a.stats.EnergyRegenRateBonus
		a.healthRegen = a.stats.SelfRepair
		a.maxHealth = a.stats.MaxHealth * a.world().droneHealthMultiplier
		if !a.IsTurret() {
			a.maxHealth *= scene.Rand().FloatRange(0.9, 1.1)
		}
		switch a.stats.Tier {
		case 1:
			a.maxEnergy = scene.Rand().FloatRange(80, 100)
		case 2:
			a.maxEnergy = scene.Rand().FloatRange(120, 200)
		case 3:
			a.maxEnergy = scene.Rand().FloatRange(150, 200)
		}
		a.speed = a.stats.Speed * scene.Rand().FloatRange(0.8, 1.1)

		switch a.faction {
		case gamedata.RedFactionTag:
			a.maxHealth *= 1.4
		case gamedata.GreenFactionTag:
			a.speed *= 1.25
		case gamedata.BlueFactionTag:
			a.maxEnergy *= 1.8
			a.energyRegenRate += 0.2
		case gamedata.YellowFactionTag:
			a.energyRegenRate += 0.5
		}
	}

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
			adventurerBits       uint64 = chance12 << (3 * chance12bits)
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
		if traitBitChance12Roll&adventurerBits == adventurerBits {
			a.traits |= traitAdventurer
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
			if a.stats.Tier == 1 {
				a.traits |= traitLowHPRecycle
			}
		case roll < 0.25:
			// 5% for berserk.
			a.traits |= traitLowHPBerserk
		case roll < 0.30:
			// 5% for panic.
			a.traits |= traitLowHPPanic
		}
	}

	if a.cloneGen == 0 {
		a.applyRankBonuses()
	}

	a.health = a.maxHealth
	a.energy = a.maxEnergy

	if a.IsFlying() && a.world().graphicsSettings.ShadowsEnabled {
		shadowImage := assets.ImageSmallShadow
		switch a.stats.Size {
		case gamedata.SizeMedium:
			shadowImage = assets.ImageMediumShadow
		case gamedata.SizeLarge:
			shadowImage = assets.ImageBigShadow
		}
		a.shadowComponent.Init(a.world(), shadowImage)
		a.shadowComponent.offset = 2
		a.shadowComponent.SetVisibility(true)
		a.shadowComponent.UpdatePos(a.pos)
	}

	a.sprite = scene.NewSprite(a.stats.Image)
	a.sprite.Pos.Base = &a.pos
	if a.IsFlying() {
		a.world().stage.AddSpriteAbove(a.sprite)
	} else {
		a.world().stage.AddSprite(a.sprite)
		// Turret damage is an optional shader.
		if a.IsTurret() && a.world().graphicsSettings.AllShadersEnabled {
			a.sprite.Shader = scene.NewShader(assets.ShaderColonyDamage)
			a.sprite.Shader.SetFloatValue("HP", 1.0)
			a.sprite.Shader.Enabled = false
			if a.stats.IsNeutral {
				a.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageBuildingDamageMask)
			} else {
				damageTexture := gmath.RandElem(a.world().localRand, turretDamageTextureList)
				a.sprite.Shader.Texture1 = scene.LoadImage(damageTexture)
			}
		}
	}

	a.flashComponent.sprite = a.sprite

	if a.faction != gamedata.NeutralFactionTag {
		diodeImg := assets.ImageFactionDiode
		if a.world().gameSettings.LargeDiodes {
			diodeImg = assets.ImageFactionDiodeLarge
		}
		a.diode = scene.NewSprite(diodeImg)
		a.diode.Pos.Base = &a.pos
		a.diode.Pos.Offset.Y = a.stats.DiodeOffset
		var colorScale ge.ColorScale
		colorScale.SetColor(gamedata.FactionByTag(a.faction).Color)
		a.diode.SetColorScale(colorScale)

		if a.IsFlying() {
			a.world().stage.AddSpriteAbove(a.diode)
		} else {
			a.world().stage.AddSprite(a.diode)
		}
	}

	if a.world().config.ExecMode != gamedata.ExecuteSimulation {
		// If there are no animation frames inside the image, do
		// not create the animation object.
		if a.sprite.FrameWidth != a.sprite.ImageWidth() {
			a.anim = ge.NewRepeatedAnimation(a.sprite, -1)
			if a.stats.AnimSpeed != 0 {
				a.anim.SetSecondsPerFrame(a.stats.AnimSpeed)
			}
			a.anim.Tick(a.world().localRand.FloatRange(0, 0.7))
			a.anim.SetOffsetY(float64(a.rank) * a.sprite.FrameHeight)
		}
	}

	a.supportDelay = scene.Rand().FloatRange(0.8, 2)

	if a.world().droneLabels && isHumanPlayer(a.colonyCore.player) {
		l := newDebugDroneLabelNode(a.colonyCore.player.GetState(), a)
		a.world().nodeRunner.AddObject(l)
	}

	a.initExtra()

	a.SetHeight(agentFlightHeight)
}

func (a *colonyAgentNode) initExtra() {
	switch a.stats {
	case gamedata.SiegeAgentStats:
		turret := newSiegeTurretNode(a.world(), a.pos)
		turret.Init(a.scene)
		a.EventDestroyed.Connect(nil, func(*colonyAgentNode) {
			turret.Dispose()
		})
		a.target = turret
		a.specialDelay = a.world().rand.FloatRange(3, 6)

	case gamedata.DroneFactoryAgentStats:
		hatch := a.scene.NewSprite(assets.ImageRelictFactoryHatch)
		hatch.Pos.Base = &a.pos
		hatch.Pos.Offset.Y = -11
		hatch.Visible = false
		a.lifetime = 1
		a.world().stage.AddSprite(hatch)
		a.EventDestroyed.Connect(nil, func(*colonyAgentNode) {
			hatch.Dispose()
		})
		a.target = hatch
		a.specialDelay = a.scene.Rand().FloatRange(10, 15)
	}
}

func (a *colonyAgentNode) SetHeight(h float64) {
	a.shadowComponent.UpdateHeight(a.pos, h, agentFlightHeight)
}

func (a *colonyAgentNode) IsDisposed() bool { return a.disposed }

func (a *colonyAgentNode) IsTurret() bool {
	return a.stats.IsTurret
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
		a.energyRegenRate += 0.1
		a.reloadRate = 1.3 // +30% attack/special reload speed
		a.healthRegen += 0.25

	case 2:
		// A super elite unit.
		a.maxHealth *= 1.5
		a.speed *= 1.2
		a.maxEnergy *= 2.0
		a.energyRegenRate += 0.3
		a.reloadRate = 1.6 // +60% attack/special reload speed
		a.healthRegen += 0.5
	}
}

func (a *colonyAgentNode) updatePatrolRadius() {
	a.dist = a.colonyCore.PatrolRadius()
	if a.stats == gamedata.CommanderAgentStats {
		a.dist = gmath.ClampMin(a.dist-40, 50)
	}
}

func (a *colonyAgentNode) AssignMode(mode colonyAgentMode, pos gmath.Vec, target any) bool {
	if a.IsTurret() {
		panic("assigning a mode to a turret")
	}

	switch mode {
	case agentModeReturn:
		entranceNum := a.scene.Rand().IntRange(0, 2)
		a.setWaypoint(a.colonyCore.GetStoragePos().Add(gmath.Vec{Y: float64(entranceNum) * 8}))
		a.mode = mode
		return true

	case agentModePatrol:
		a.mode = mode
		a.updatePatrolRadius()
		a.setWaypoint(a.orbitingWaypoint(a.colonyCore.pos, a.dist))
		a.waypointsLeft = a.scene.Rand().IntRange(40, 70)
		return true

	case agentModeWaitCloning:
		a.mode = mode
		a.target = target
		a.clearWaypoint()
		return true

	case agentModeMakeClone:
		a.mode = mode
		a.target = target
		a.dist = a.scene.Rand().FloatRange(1.2, 2) // cloning time
		a.energyBill += 20
		targetPos := target.(*colonyAgentNode).pos
		a.setWaypoint(a.pos.DirectionTo(targetPos).Mulf(110).Add(targetPos).Add(a.scene.Rand().Offset(-20, 20)))
		return true

	case agentModeMerging, agentModeMergingRoomba:
		a.mode = mode
		a.target = target
		a.dist = a.scene.Rand().FloatRange(8, 10) // merging time
		a.setWaypoint(pos)
		if mode == agentModeMergingRoomba {
			a.dist *= 1.5
		}
		return true

	case agentModeAlignStandby:
		a.shadowComponent.SetVisibility(true)
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
			a.cloningBeam = nil
		}
		a.sprite.SetColorScale(ge.ColorScale{R: 1, G: 1, B: 1, A: 1})
		a.mode = mode
		a.setWaypoint(a.pos.Sub(gmath.Vec{Y: agentFlightHeight - a.shadowComponent.height}))
		return true

	case agentModeMove:
		a.mode = mode
		a.setWaypoint(pos)
		return true

	case agentModePanic:
		a.mode = mode
		a.setWaypoint(a.pos)
		a.waypointsLeft = a.scene.Rand().IntRange(4, 9)
		return true

	case agentModeStandby:
		if a.shadowComponent.height != agentFlightHeight {
			return a.AssignMode(agentModeAlignStandby, pos, target)
		}
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
			a.cloningBeam = nil
		}
		a.mode = mode
		maxDist := a.colonyCore.realRadius
		if !a.stats.CanPatrol {
			maxDist *= 0.65
		}
		a.dist = a.scene.Rand().FloatRange(40, maxDist)
		a.setWaypoint(a.orbitingWaypoint(a.colonyCore.pos, a.dist))
		a.waypointsLeft = 0
		return true

	case agentModeBomberAttack:
		a.mode = mode
		a.target = target
		a.waypointsLeft = 5
		a.setWaypoint(target.(*creepNode).pos.Sub(gmath.Vec{Y: agentFlightHeight}))
		return true

	case agentModeFollow, agentModeAttack:
		isPatrol := a.mode == agentModePatrol
		a.mode = agentModeFollow // attack is a long-range follow
		a.target = target
		a.setWaypoint(a.followWaypoint(target.(*creepNode).pos))
		if isPatrol {
			a.waypointsLeft = a.scene.Rand().IntRange(4, 6)
		} else {
			a.waypointsLeft = a.scene.Rand().IntRange(6, 8)
		}
		if mode == agentModeAttack {
			a.waypointsLeft += 5
			if a.hasTrait(traitDoOrDie) {
				a.waypointsLeft += 15
			}
		}
		return true

	case agentModeCloakHide:
		a.mode = mode
		a.clearWaypoint()
		return true

	case agentModeCharging, agentModeForcedCharging:
		a.mode = mode
		a.clearWaypoint()
		return true

	case agentModePosing:
		a.mode = mode
		a.dist = pos.X // idle time
		a.clearWaypoint()
		return true

	case agentModeCourierFlight:
		colony := target.(*colonyCoreNode)
		energyCost := gmath.ClampMax(colony.pos.DistanceTo(a.pos)*0.33, 100)
		if a.tether {
			energyCost *= 0.5
		}
		if a.stats.Kind == gamedata.AgentTrucker {
			// Truckers consume 20% less energy for flights.
			energyCost *= 0.8
		}
		a.energyBill += energyCost
		a.target = target
		a.mode = mode
		a.setWaypoint(a.pos)
		return true

	case agentModeScavenge:
		source := target.(*essenceSourceNode)
		energyCost := source.pos.DistanceTo(a.pos) * 0.33
		if a.tether {
			energyCost *= 0.5
		}
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		if a.stats.Kind == gamedata.AgentMarauder && a.specialDelay == 0 {
			a.doCloak(20)
			a.specialDelay = 10
		}
		a.energyBill += energyCost
		a.mode = agentModeMineEssence
		a.setWaypoint(roundedPos(source.pos.Sub(gmath.Vec{Y: agentFlightHeight}).Add(a.scene.Rand().Offset(-8, 8))))
		a.target = target
		return true

	case agentModeMineSulfurEssence:
		if !a.stats.CanGather {
			return false
		}
		switch a.stats.Kind {
		case gamedata.AgentCourier, gamedata.AgentTrucker:
			return false
		}
		source := target.(*essenceSourceNode)
		energyCost := (source.pos.DistanceTo(a.pos) * 0.3) + 20
		if a.tether {
			energyCost *= 0.5
		}
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.dist = a.scene.Rand().FloatRange(20, 25) // time to harvest
		a.energyBill += energyCost
		a.mode = mode
		a.setWaypoint(roundedPos(source.pos.Sub(gmath.Vec{Y: agentFlightHeight}).Add(a.scene.Rand().Offset(-10, 10))))
		a.target = target
		return true

	case agentModeMineEssence:
		if !a.stats.CanGather {
			return false
		}
		switch a.stats.Kind {
		case gamedata.AgentCourier, gamedata.AgentTrucker:
			// Couriers try to keep their energy for travelling between the bases.
			if a.energy < 120 || a.energyBill > 10 {
				return false
			}
		}
		source := target.(*essenceSourceNode)
		if source.stats == redOilSource && a.stats.Kind != gamedata.AgentRedminer {
			return false
		}
		energyCost := source.pos.DistanceTo(a.pos) * 0.5
		if a.tether {
			energyCost *= 0.5
		}
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.energyBill += energyCost
		a.mode = mode
		a.setWaypoint(roundedPos(source.pos.Sub(gmath.Vec{Y: agentFlightHeight}).Add(a.scene.Rand().Offset(-8, 8))))
		a.target = target
		return true

	case agentModeTakeoff:
		a.mode = mode
		a.setWaypoint(a.pos.Sub(gmath.Vec{Y: agentFlightHeight - a.shadowComponent.height}))
		a.shadowComponent.SetVisibility(false)
		return true

	case agentModePickup:
		a.mode = mode
		a.setWaypoint(a.pos.Add(gmath.Vec{Y: agentFlightHeight}))
		return true

	case agentModeRecycleReturn:
		a.mode = mode
		offset := gmath.Vec{Y: agentFlightHeight - a.colonyCore.stats.DefaultHeight}
		a.setWaypoint(a.colonyCore.GetEntrancePos().Sub(offset))
		return true

	case agentModeRecycleLanding:
		a.mode = mode
		a.setWaypoint(a.colonyCore.GetEntrancePos())
		a.shadowComponent.SetVisibility(false)
		return true

	case agentModeCaptureBuilding:
		const energyCost = 20.0
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.target = target
		a.mode = mode
		a.energyBill += energyCost
		a.dist = a.scene.Rand().FloatRange(5, 7) // working time
		a.setWaypoint(gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(target.(*neutralBuildingNode).pos))
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
		a.setWaypoint(gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(target.(*colonyAgentNode).pos))
		return true

	case agentModeRepairBase:
		energyCost := 40.0
		if energyCost > a.energy && !a.hasTrait(traitWorkaholic) {
			return false
		}
		a.mode = mode
		a.energyBill += energyCost
		a.dist = a.scene.Rand().FloatRange(3, 4) // repair time
		a.setWaypoint(gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(a.colonyCore.pos))
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
		a.setWaypoint(gmath.RadToVec(a.scene.Rand().Rad()).Mulf(64.0).Add(construction.pos))
		return true

	case agentModeKamikazeAttack:
		a.clearWaypoint()
		a.mode = mode
		a.target = target
		return true

	case agentModeConsumeDrone:
		a.setWaypoint(target.(*colonyAgentNode).pos.Add(a.scene.Rand().Offset(-4, 4)))
		a.mode = mode
		a.target = target
		return true

	case agentModeFollowCommander:
		a.mode = mode
		a.target = target
		a.setWaypoint(target.(*colonyAgentNode).pos.Add(a.scene.Rand().Offset(-16, 16)))
		a.waypointsLeft = a.scene.Rand().IntRange(60, 100)
		weaponRange := a.stats.Weapon.AttackRange
		switch {
		case weaponRange <= 150:
			a.dist = 70
		case weaponRange <= 200:
			a.dist = 60
		default:
			a.dist = 50
		}
		return true
	}

	return false
}

func (a *colonyAgentNode) orbitingWaypoint(center gmath.Vec, dist float64) gmath.Vec {
	return orbitingWaypoint(a.world(), a.pos, center, dist, !a.hasTrait(traitCounterClocwiseOrbiting))
}

func (a *colonyAgentNode) Update(delta float64) {
	if a.anim != nil {
		a.anim.Tick(delta)
	}
	a.flashComponent.Update(delta)

	if a.stats.Tier == 1 {
		a.lifetime -= delta
	}

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

	a.slow = gmath.ClampMin(a.slow-delta, 0)
	a.specialDelay = gmath.ClampMin(a.specialDelay-delta, 0)

	if a.cloaking > 0 {
		a.cloaking -= delta
		if a.cloaking <= 0 {
			a.doUncloak()
		}
	}

	a.processAttack(delta)
	a.processSupport(delta)

	switch a.mode {
	case agentModeStandby:
		a.updateStandby(delta)
	case agentModeAlignStandby:
		a.updateAlignStandby(delta)
	case agentModePosing:
		a.updatePosing(delta)
	case agentModeCharging:
		a.updateCharging(delta)
	case agentModeForcedCharging:
		a.updateForcedCharging(delta)
	case agentModeCloakHide:
		a.updateCloakHide(delta)
	case agentModeMineEssence:
		a.updateMineEssence(delta)
	case agentModeMineSulfurEssence:
		a.updateMineSulfurEssence(delta)
	case agentModePickup:
		a.updatePickup(delta)
	case agentModeReturn:
		a.updateReturn(delta)
	case agentModePatrol:
		a.updatePatrol(delta)
	case agentModeFollowCommander:
		a.updateFollowCommander(delta)
	case agentModeMove:
		a.updateMove(delta)
	case agentModePanic:
		a.updatePanic(delta)
	case agentModeCourierFlight:
		a.updateCourierFlight(delta)
	case agentModeFollow:
		a.updateFollow(delta)
	case agentModeWaitCloning:
		a.updateWaitCloning(delta)
	case agentModeMakeClone:
		a.updateMakeClone(delta)
	case agentModeMerging, agentModeMergingRoomba:
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
	case agentModeCaptureBuilding:
		a.updateCaptureBuilding(delta)
	case agentModeKamikazeAttack:
		a.updateKamikazeAttack(delta)
	case agentModeBomberAttack:
		a.updateBomberAttack(delta)
	case agentModeConsumeDrone:
		a.updateConsumeDrone(delta)
	case agentModeRoombaPatrol, agentModeRoombaGuard, agentModeRoombaAttack:
		a.updateRoombaPatrol(delta)
	case agentModeRoombaWait, agentModeRoombaCombatWait:
		a.updateRoombaWait(delta)
	case agentModeHarvester:
		a.updateHarvester(delta)
	case agentModeRelictDroneFactory:
		a.updateRelictDroneFactory(delta)
	case agentModeRelictTakeoff:
		a.updateRelictTakeoff(delta)
	case agentModeRelictPatrol:
		a.updateRelictPatrol(delta)
	case agentModeRelictRoomba:
		a.updateRelictRoomba(delta)
	case agentModeSiegeGuard:
		a.updateSiegeGuard(delta)
	case agentModeGuardForever:
		// Just chill.
	}
}

func (a *colonyAgentNode) dispose() {
	a.disposed = true
	a.sprite.Dispose()
	a.shadowComponent.Dispose()
	if a.diode != nil {
		a.diode.Dispose()
	}
	if a.cloningBeam != nil {
		a.cloningBeam.Dispose()
		a.cloningBeam = nil
	}

	if a.stats.Kind == gamedata.AgentHarvester && a.target != nil {
		target := a.target.(*essenceSourceNode)
		target.beingHarvested = false
	}
}

func (a *colonyAgentNode) Destroy() {
	a.EventDestroyed.Emit(a)
	a.dispose()
}

func (a *colonyAgentNode) IsFlying() bool {
	return a.stats.IsFlying
}

func (a *colonyAgentNode) GetTargetInfo() targetInfo {
	return targetInfo{building: a.stats.IsBuilding, flying: a.IsFlying()}
}

func (a *colonyAgentNode) ReceiveEnergyDamage(damage float64) {
	a.energy = gmath.ClampMin(a.energy-damage, 0)
}

func (a *colonyAgentNode) doUncloak() {
	a.cloaking = 0
	a.sprite.SetAlpha(1)
}

func (a *colonyAgentNode) doCloak(d float64) {
	a.cloaking = d
	a.sprite.SetAlpha(0.2)
	if !a.world().simulation {
		a.world().nodeRunner.AddObject(newEffectNode(a.world(), a.pos, aboveEffectLayer, assets.ImageCloakWave))
	}
	playSound(a.world(), assets.AudioStealth, a.pos)
}

func (a *colonyAgentNode) explode() {
	if !a.stats.IsFlying {
		createAreaExplosion(a.world(), spriteRect(a.pos, a.sprite), normalEffectLayer)
		if !a.stats.IsNeutral && a.IsTurret() || a.scene.Rand().Chance(0.3) {
			a.world().CreateScrapsAt(scrapSource, a.pos.Add(gmath.Vec{Y: 2}))
		}
		return
	}

	explosionRoll := a.scene.Rand().Float()
	explodesCompletely := explosionRoll < 0.3

	if !explodesCompletely {
		playSound(a.world(), assets.AudioAgentDestroyed, a.pos)
	}

	if a.colonyCore.GetSecurityPriority() < 0.4 {
		a.colonyCore.AddPriority(prioritySecurity, 0.04)
	}
	if a.scene.Rand().Chance(0.6) {
		a.colonyCore.AddPriority(priorityGrowth, 0.01)
	}

	if explodesCompletely {
		createExplosion(a.world(), aboveEffectLayer, a.pos)
	} else {
		var scraps *essenceSourceStats
		if !a.stats.IsNeutral {
			if explosionRoll > 0.6 {
				scraps = smallScrapSource
				if a.stats.Size != gamedata.SizeSmall {
					scraps = scrapSource
				}
			}
		}

		shadowImg := a.shadowComponent.GetImageID()
		fall := newDroneFallNode(a.world(), scraps, a.stats.Image, shadowImg, a.pos, a.shadowComponent.height)
		fall.FrameOffsetY = float64(a.rank) * a.sprite.FrameHeight
		a.world().nodeRunner.AddObject(fall)
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
	if a.sprite.Shader.IsNil() {
		return
	}
	percentage := a.health / a.maxHealth
	a.sprite.Shader.SetFloatValue("HP", percentage)
	a.sprite.Shader.Enabled = percentage < 0.95
}

func (a *colonyAgentNode) CanAttack(mask gamedata.TargetKind) bool {
	return a.stats.Weapon != nil && a.stats.Weapon.TargetFlags&mask != 0
}

func (a *colonyAgentNode) IsCloaked() bool {
	return a.cloaking > 0
}

func (a *colonyAgentNode) onLowHealthDamage(source targetable) {
	if a.stats.Kind == gamedata.AgentKamikaze && source.IsFlying() {
		switch a.mode {
		case agentModePatrol, agentModeStandby, agentModeMineEssence, agentModeReturn, agentModeFollow:
			if creep, ok := source.(*creepNode); ok {
				a.health = gmath.ClampMax(a.health+10, a.maxHealth)
				a.dist = 0.1 + a.scene.Rand().FloatRange(0.05, 0.25)
				a.AssignMode(agentModeKamikazeAttack, gmath.Vec{}, creep)
				return
			}
		}
	}

	if a.mode == agentModeRoombaCombatWait && a.scene.Rand().Chance(0.4) {
		a.mode = agentModeRoombaGuard
		a.sendTo(a.colonyCore.pos.Add(a.scene.Rand().Offset(-180, 180)), layerNormal)
		if a.specialDelay == 0 {
			a.energy = gmath.ClampMax(a.energy+50, a.maxEnergy)
			a.specialDelay = 20
			a.resting = false
		}
		return
	}

	// Don't do anything weird when colony is being relocated.
	if a.colonyCore.mode != colonyModeNormal {
		return
	}

	switch a.mode {
	case agentModeStandby, agentModeFollow, agentModePatrol:
		// OK, can interrupt.
	default:
		// Most modes can't be safely be interrupted like this.
		return
	}

	if a.stats.CanCloak && !a.IsCloaked() && a.specialDelay == 0 {
		a.AssignMode(agentModeCloakHide, gmath.Vec{}, nil)
		a.doCloak(a.scene.Rand().FloatRange(6, 10))
		a.specialDelay = a.scene.Rand().FloatRange(6, 10)
		return
	}

	switch {
	case a.hasTrait(traitLowHPBerserk):
		// Berserks go straight into the danger when low on health.
		a.AssignMode(agentModeMove, source.GetPos().Add(a.scene.Rand().Offset(-20, 20)), nil)
	case a.hasTrait(traitLowHPRecycle):
		// Recycle agents may go to recycle themselves on low health.
		if a.scene.Rand().Chance(0.8) {
			a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
		}
	case a.hasTrait(traitLowHPRetreat):
		// Agents with retreat trait will try to fly away from a threat on low health.
		pos := retreatPos(a.scene.Rand(), a.scene.Rand().FloatRange(80, 140), a.pos, *source.GetPos())
		a.AssignMode(agentModeMove, pos, nil)
	case a.hasTrait(traitLowHPPanic):
		// Agents with panic trait will stop what they're doing and fly like crazy.
		a.AssignMode(agentModePanic, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) OnDamage(damage gamedata.DamageValue, source targetable) {
	if a.disposed {
		return
	}

	if damage.Health > 0 {
		multiplier := 1.0 - a.damageReduction()
		healthDamage := damage.Health * multiplier
		a.health -= healthDamage

		if a.health < 0 {
			a.explode()
			a.Destroy()
			return
		}
	}

	if !a.IsTurret() {
		if a.colonyCore.GetSecurityPriority() < 0.3 && a.scene.Rand().Chance(1.0-a.colonyCore.GetSecurityPriority()) {
			a.colonyCore.AddPriority(prioritySecurity, 0.01)
		}
	}

	a.energy = gmath.ClampMin(a.energy-damage.Energy, 0)
	a.slow = gmath.ClampMax(a.slow+damage.Slow, 5)

	if damage.Health != 0 {
		if !damage.HasFlag(gamedata.DmgflagNoFlash) {
			a.flashComponent.SetFlash(a.world().localRand.FloatRange(0.07, 0.14))
		}
		if a.IsTurret() {
			a.updateHealthShader()
		}
		if a.stats.Kind == gamedata.AgentRoomba && !source.IsFlying() {
			switch a.mode {
			case agentModeRoombaPatrol:
				if source.GetPos().DistanceSquaredTo(a.pos) < a.stats.Weapon.AttackRangeSqr && a.scene.Rand().Chance(0.2) {
					a.mode = agentModeRoombaCombatWait
					a.dist = a.scene.Rand().FloatRange(6, 10)
					a.clearWaypoint()
					return
				}
			case agentModeRoombaWait:
				if source != a.target && source.GetPos().DistanceSquaredTo(a.pos) > a.stats.Weapon.AttackRangeSqr && a.scene.Rand().Chance(0.45) {
					a.mode = agentModeRoombaAttack
					a.target = source
					a.sendTo(midpoint(a.pos, *source.GetPos()), layerNormal)
					return
				}
			}
		}
	}

	if a.health <= (a.maxHealth*0.33) || a.health <= damage.Health {
		a.onLowHealthDamage(source)
	}

	if !a.stats.IsFlying {
		return
	}

	if damage.HasFlag(gamedata.DmgflagStun) {
		maxEnergy := 0.8
		energyDamage := 8.0
		if damage.HasFlag(gamedata.DmgflagStunImproved) {
			maxEnergy = 0.95
			energyDamage = 12.0
		}
		if a.energy < a.maxEnergy*maxEnergy {
			if stunnableModes[a.mode] && a.scene.Rand().Chance(0.9) {
				a.clearCargo()
				a.AssignMode(agentModeForcedCharging, gmath.Vec{}, nil)
				a.energyTarget = gmath.ClampMax(a.energy+a.scene.Rand().FloatRange(4, 8), a.maxEnergy-0.01)
				return
			} else {
				a.energy = gmath.ClampMin(a.energy-energyDamage, 0)
			}
		}
	}

	if damage.Morale != 0 && a.stats.Kind != gamedata.AgentCommander {
		// Note that agentModeFollowCommander is not listed here.
		// A commanded unit is immune to the panic effect.
		switch a.mode {
		case agentModeMineEssence:
			a.clearCargo()
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)

		case agentModePatrol, agentModeStandby, agentModeFollow:
			if !a.scene.Rand().Chance(damage.Morale) {
				break
			}
			effectRoll := a.scene.Rand().Float()
			if effectRoll < 0.4 {
				a.AssignMode(agentModePanic, gmath.Vec{}, nil)
			} else {
				pos := retreatPos(a.scene.Rand(), a.scene.Rand().FloatRange(80, 140), a.pos, *source.GetPos())
				a.AssignMode(agentModeMove, pos, nil)
			}
		}
	}
}

func (a *colonyAgentNode) GetPos() *gmath.Vec { return &a.pos }

func (a *colonyAgentNode) GetVelocity() gmath.Vec {
	if !a.hasWaypoint() {
		return gmath.Vec{}
	}
	return a.dir.Mulf(a.movementSpeed())
}

func (a *colonyAgentNode) processSupport(delta float64) {
	if !a.stats.HasSupport {
		return
	}

	a.supportDelay = gmath.ClampMin(a.supportDelay-(delta*a.reloadRate), 0)

	if a.supportDelay != 0 {
		return
	}

	setDelay := true
	switch a.stats.Kind {
	case gamedata.AgentRecharger:
		a.doRecharge()
	case gamedata.AgentRepair:
		a.doRepair()
	case gamedata.AgentScavenger, gamedata.AgentMarauder:
		a.doScavenge()
	case gamedata.AgentDisintegrator:
		// Reload depends on the target being there or not.
		setDelay = false
		a.doDisintegratorAttack()
	case gamedata.AgentDevourer:
		a.doConsumeDrone()
	case gamedata.AgentTetherBeacon:
		setDelay = false
		a.doTether()
	}
	if setDelay {
		a.supportDelay = a.stats.SupportReload * a.scene.Rand().FloatRange(0.7, 1.4)
	}
}

func (a *colonyAgentNode) doConsumeDrone() {
	if a.colonyCore.mode != colonyModeNormal {
		return
	}
	switch a.mode {
	case agentModeStandby, agentModePatrol, agentModeFollowCommander:
		// OK
	default:
		return
	}

	if a.colonyCore.agents.NumAvailableWorkers() < 5 || a.colonyCore.agents.NumAvailableFighters() < 5 {
		return
	}

	if a.extraLevel >= gamedata.DevourerMaxLevel {
		// A max-developed devourer will only consume for healing.
		if a.health >= (a.maxHealth * 0.6) {
			return
		}
		if a.colonyCore.resources < 30 {
			return
		}
	} else {
		// A developing devourer may consume even on full health, sometimes.
		if a.health >= a.maxHealth && a.scene.Rand().Chance(0.65) {
			return
		}
		if a.colonyCore.resources < 50 {
			return
		}
	}

	// Prefer the kind of drones that is less scarce.
	bestKind := gamedata.AgentWorker
	if a.colonyCore.agents.NumAvailableFighters() > a.colonyCore.agents.NumAvailableWorkers() {
		bestKind = gamedata.AgentScout
	}

	var bestTarget *colonyAgentNode
	bestScore := 0.0
	a.colonyCore.agents.Find(searchWorkers|searchFighters|searchRandomized|searchOnlyAvailable, func(x *colonyAgentNode) bool {
		if x.stats.Tier != 1 {
			return false
		}
		if x.stats.Kind == gamedata.AgentScout && x.health == x.maxHealth {
			return false
		}
		score := 2.0
		if a.stats.Kind == bestKind {
			score += 0.5
		}
		if a.faction == gamedata.NeutralFactionTag {
			score += 0.5
		}
		if a.rank != 0 {
			score -= float64(a.rank) * 0.5
		}
		multiplier := (2.0 - (x.health / x.maxHealth)) + (1.2 - (x.energy / x.maxEnergy))
		score *= multiplier
		if score > bestScore {
			bestTarget = x
			bestScore = score
		}
		return false
	})
	if bestTarget == nil {
		return
	}

	// Make it wait for the devourer to come closer.
	// If anything goes wrong, it will get back to normal after some time.
	// 8 seconds should be enough.
	bestTarget.AssignMode(agentModePosing, gmath.Vec{X: 8}, nil)

	a.AssignMode(agentModeConsumeDrone, gmath.Vec{}, bestTarget)
}

func (a *colonyAgentNode) tetherTarget(target *colonyAgentNode) {
	if target.mode == agentModeForcedCharging {
		target.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		target.energy = gmath.ClampMax(target.energy+5, target.maxEnergy)
	}
	target.tether = true
	if target.energyBill > 50 {
		target.energyBill = gmath.ClampMin(target.energyBill-20, 50)
	}
	a.world().nodeRunner.AddObject(newTetherNode(a.world(), a, target))
	playSound(a.world(), assets.AudioTetherShot, a.pos)
}

func (a *colonyAgentNode) doTether() {
	// If it's connected to the colony, it can't boost anyone else.
	if a.target != nil {
		if tether, ok := a.target.(*tetherNode); ok {
			a.supportDelay = a.scene.Rand().FloatRange(0.5, 2)
			if !tether.IsDisposed() {
				return
			}
			a.target = nil
		}
	}

	colonyTarget := randIterate(a.scene.Rand(), a.world().allColonies, func(colony *colonyCoreNode) bool {
		if colony.waypoint.IsZero() {
			return false
		}
		return colony.pos.DistanceSquaredTo(a.pos) <= a.stats.SupportRangeSqr
	})
	if colonyTarget != nil && a.target != colonyTarget {
		tether := newTetherNode(a.world(), a, colonyTarget)
		a.target = tether
		colonyTarget.tether++
		a.world().nodeRunner.AddObject(tether)
		playSound(a.world(), assets.AudioTetherShot, a.pos)
		a.supportDelay = a.stats.SupportReload * a.scene.Rand().FloatRange(0.95, 1.35)
		return
	}

	const maxNumberOfTargets = 4
	actionsLeft := maxNumberOfTargets
	actionsLeft -= a.walkTetherTargets(a.colonyCore, actionsLeft, func(x *colonyAgentNode) {
		a.tetherTarget(x)
	})
	if actionsLeft != 0 {
		randIterate(a.scene.Rand(), a.world().allColonies, func(colony *colonyCoreNode) bool {
			if actionsLeft <= 0 {
				return true
			}
			if colony == a.colonyCore || a.pos.DistanceSquaredTo(colony.pos) > (a.stats.SupportRangeSqr*1.2) {
				return false
			}
			actionsLeft -= a.walkTetherTargets(colony, actionsLeft, func(x *colonyAgentNode) {
				a.tetherTarget(x)
			})
			return actionsLeft <= 0
		})
	}

	if actionsLeft != maxNumberOfTargets {
		a.supportDelay = a.stats.SupportReload * a.scene.Rand().FloatRange(0.95, 1.35)
		return
	}
	a.supportDelay = a.scene.Rand().FloatRange(0.5, 2)
}

func (a *colonyAgentNode) doDisintegratorAttack() {
	switch a.mode {
	case agentModePatrol, agentModeStandby, agentModeFollow, agentModeMineEssence:
		// OK
	default:
		return
	}

	const attackEnergyCost = 40.0
	if a.energy < attackEnergyCost || a.shadowComponent.height != agentFlightHeight {
		return
	}
	targets := a.findAttackTargets()
	if len(targets) == 0 {
		a.supportDelay = a.scene.Rand().FloatRange(0.15, 1.2)
		return
	}
	a.energy -= attackEnergyCost
	a.supportDelay = a.stats.SupportReload * a.scene.Rand().FloatRange(0.8, 1.2)
	target := targets[0]
	toPos := snipePos(a.stats.Weapon.ProjectileSpeed, a.pos, *target.GetPos(), target.GetVelocity())
	p := a.world().newProjectileNode(projectileConfig{
		World:    a.world(),
		Weapon:   a.stats.Weapon,
		Attacker: a,
		ToPos:    toPos,
		Target:   target,
	})
	a.world().nodeRunner.AddProjectile(p)
	a.AssignMode(agentModeForcedCharging, gmath.Vec{}, nil)
	playSound(a.world(), a.stats.Weapon.AttackSound, a.pos)
	a.world().nodeRunner.AddObject(newEffectNode(a.world(), a.pos, aboveEffectLayer, assets.ImagePurpleIonZap))
	a.energyTarget = a.energy + a.scene.Rand().FloatRange(22, 35)
}

func (a *colonyAgentNode) doScavenge() {
	if a.colonyCore.mode != colonyModeNormal {
		return
	}
	switch a.mode {
	case agentModeStandby, agentModePatrol:
		// OK
	default:
		return
	}
	if a.energy < 20 || a.energyBill > 100 {
		return
	}
	if a.colonyCore.resources > a.colonyCore.maxVisualResources() {
		return
	}

	maxDistSqr := 256.0 * 256.0
	if a.stats.Kind == gamedata.AgentMarauder {
		maxDistSqr = 300.0 * 300.0
	}

	var bestSource *essenceSourceNode
	bestScore := 0.0
	for _, source := range a.world().essenceSources {
		if !source.stats.scrap {
			continue
		}
		distSqr := a.pos.DistanceSquaredTo(source.pos)
		if distSqr > maxDistSqr {
			continue
		}
		score := distSqr * a.scene.Rand().FloatRange(0.6, 1.6)
		if score != 0 && score > bestScore {
			bestScore = score
			bestSource = source
		}
	}
	if bestSource != nil {
		a.AssignMode(agentModeScavenge, gmath.Vec{}, bestSource)
	}
}

func (a *colonyAgentNode) doRecharge() {
	const rechargerEnergyRecorery float64 = 25.0
	target := a.colonyCore.agents.Find(searchWorkers|searchFighters|searchRandomized, func(x *colonyAgentNode) bool {
		return x != a &&
			x.mode != agentModeKamikazeAttack &&
			(x.energy+rechargerEnergyRecorery) < x.maxEnergy &&
			x.pos.DistanceSquaredTo(a.pos) < gamedata.RechargerAgentStats.SupportRangeSqr
	})
	if target != nil {
		if !a.world().simulation {
			a.createBeam(target, gamedata.RechargerAgentStats)
		}
		target.energy = gmath.ClampMax(target.energy+rechargerEnergyRecorery, target.maxEnergy)
		playSound(a.world(), assets.AudioRechargerBeam, a.pos)
	}
}

func (a *colonyAgentNode) createBeam(target targetable, beamStats *gamedata.AgentStats) {
	offset := gmath.Vec{Y: a.stats.FireOffset}
	targetPos := target.GetPos()
	if a.stats.BeamShift != 0 {
		shift := targetPos.DirectionTo(a.pos).Mulf(a.stats.BeamShift)
		offset = offset.Add(shift)
	}
	from := ge.Pos{Base: &a.pos, Offset: offset}
	to := ge.Pos{Base: targetPos, Offset: gmath.Vec{Y: -2}}
	if beamStats.BeamTexture == nil {
		beam := newBeamNode(a.world(), from, to, beamStats.BeamColor)
		beam.width = beamStats.BeamWidth
		a.world().nodeRunner.AddObject(beam)
	} else {
		beam := newTextureBeamNode(a.world(), from, to, beamStats.BeamTexture, beamStats.BeamSlideSpeed, beamStats.BeamOpaqueTime)
		a.world().nodeRunner.AddObject(beam)
	}
	if a.stats.BeamExplosion != assets.ImageNone {
		createEffect(a.world(), effectConfig{
			Pos:   target.GetPos().Add(a.world().localRand.Offset(-6, 6)),
			Layer: effectLayerFromBool(target.IsFlying()),
			Image: a.stats.BeamExplosion,
		})
	}
}

func (a *colonyAgentNode) doRepair() {
	target := a.colonyCore.agents.Find(searchWorkers|searchFighters|searchRandomized, func(x *colonyAgentNode) bool {
		return x != a &&
			x.mode != agentModeKamikazeAttack &&
			x.health < x.maxHealth &&
			x.pos.DistanceSquaredTo(a.pos) < gamedata.RepairAgentStats.SupportRangeSqr
	})
	if target != nil {
		if !a.world().simulation {
			a.createBeam(target, gamedata.RepairAgentStats)
		}
		target.health = gmath.ClampMax(target.health+3, target.maxHealth)
		playSound(a.world(), assets.AudioRepairBeam, a.pos)
	}
}

func (a *colonyAgentNode) walkTetherTargets(colony *colonyCoreNode, num int, f func(x *colonyAgentNode)) int {
	targets := a.world().tmpTargetSlice[:0]
	processed := 0
	colony.agents.Find(searchWorkers|searchRandomized, func(x *colonyAgentNode) bool {
		if processed >= num {
			return true
		}
		if x.tether {
			return false
		}
		if x.pos.DistanceSquaredTo(a.pos) > a.stats.SupportRangeSqr {
			return false
		}
		switch x.mode {
		case agentModeKamikazeAttack, agentModeBomberAttack, agentModeConsumeDrone, agentModeCharging, agentModePanic, agentModeWaitCloning, agentModeMakeClone, agentModeRecycleReturn, agentModeRecycleLanding, agentModeMerging, agentModePosing:
			// Modes that are never targeted.
			return false
		}

		switch x.mode {
		case agentModeMineEssence, agentModeReturn, agentModeCourierFlight, agentModeForcedCharging, agentModeMineSulfurEssence:
			// The best modes to be hastened.
			processed++
			f(x)
		default:
			if len(targets) < num {
				targets = append(targets, x)
			}
		}
		return processed >= num
	})
	if processed < num {
		for _, target := range targets {
			f(target.(*colonyAgentNode))
			processed++
			if processed >= num {
				break
			}
		}
	}
	return processed
}

func (a *colonyAgentNode) findAttackTargets() []targetable {
	w := a.world()

	maxTargets := a.stats.Weapon.MaxTargets
	targets := w.tmpTargetSlice[:0]
	w.WalkCreeps(a.pos, a.stats.Weapon.AttackRange, func(creep *creepNode) bool {
		if a.isValidTarget(creep) {
			targets = append(targets, creep)
		}
		return len(targets) >= maxTargets
	})

	return targets
}

func (a *colonyAgentNode) isValidTarget(creep *creepNode) bool {
	return isValidCreepTarget(a.pos, creep, a.stats.Weapon)
}

func (a *colonyAgentNode) attackWithProjectile(target targetable, burstSize int) {
	attackWithProjectile(a.world(), a.stats.Weapon, a, target, burstSize, a.mode == agentModeFollowCommander)
}

func (a *colonyAgentNode) attackTargets(targets []targetable, burstSize int) {
	for _, target := range targets {
		if a.stats.Weapon.ProjectileSpeed != 0 {
			a.attackWithProjectile(target, burstSize)
		} else {
			// TODO: this code is duplited with creep node.
			if !a.world().simulation {
				a.createBeam(target, a.stats)
			}
			target.OnDamage(multipliedDamage(target, a.stats.Weapon), a)
		}
	}
}

func (a *colonyAgentNode) processAttack(delta float64) {
	if a.stats.Weapon == nil || a.stats.Kind == gamedata.AgentDisintegrator {
		return
	}

	reloaded := delta * a.reloadRate
	if a.resting {
		reloaded *= 0.7
	}
	a.attackDelay = gmath.ClampMin(a.attackDelay-reloaded, 0)
	if a.attackDelay != 0 {
		return
	}
	if a.IsCloaked() || a.insideForest || a.mode == agentModeForcedCharging {
		return
	}

	targets := a.findAttackTargets()
	if len(targets) == 0 {
		a.attackDelay = 0.75 * a.scene.Rand().FloatRange(0.8, 1.4)
		return
	}

	reloadMultiplier := a.scene.Rand().FloatRange(0.8, 1.2)
	if a.stats == gamedata.BeamTowerAgentStats {
		reloadMultiplier += (a.specialDelay * 0.3)
		a.specialDelay += ((a.stats.Weapon.Reload) + 1.75) * reloadMultiplier
	}

	a.attackDelay = a.stats.Weapon.Reload * reloadMultiplier
	a.energy = gmath.ClampMin(a.energy-a.stats.Weapon.EnergyCost, 0)

	switch a.stats.Kind {
	case gamedata.AgentDestroyer:
		target := targets[0]
		offset := gmath.Vec{X: -7, Y: 2}
		offsetStep := gmath.Vec{X: 14}
		targetOffset := gmath.Vec{X: -4}
		targetOffsetStep := gmath.Vec{X: 8}
		for i := 0; i < 2; i++ {
			pos1 := ge.Pos{Base: &a.pos, Offset: offset}
			pos2 := ge.Pos{Base: target.GetPos(), Offset: targetOffset}
			beam := newBeamNode(a.world(), pos1, pos2, destroyerBeamColor)
			beam.width = 2
			a.world().nodeRunner.AddObject(beam)
			offset = offset.Add(offsetStep)
			targetOffset = targetOffset.Add(targetOffsetStep)
		}
		target.OnDamage(multipliedDamage(target, a.stats.Weapon), a)

	case gamedata.AgentPrism:
		target := targets[0]
		damage := a.stats.Weapon.Damage
		width := 1.0
		numReflections := 0
		pos := &a.pos
		// Prism damage:
		//   0 => 4
		//   1 => 6
		//   2 => 8
		//   3 => 10
		//   4 => 14 (max)
		a.colonyCore.agents.Find(searchFighters|searchRandomized, func(ally *colonyAgentNode) bool {
			if ally.stats.Kind != gamedata.AgentPrism || ally == a {
				return false
			}
			if ally.pos.DistanceSquaredTo(*pos) > (96 * 96) {
				return false
			}
			ally.attackDelay += float64(numReflections) * 0.1
			beam := newBeamNode(a.world(), ge.Pos{Base: pos}, ge.Pos{Base: &ally.pos}, prismBeamColors[numReflections])
			beam.width = width
			a.world().nodeRunner.AddObject(beam)
			numReflections++
			damage.Health += gamedata.PrismDamagePerReflection
			if numReflections < gamedata.PrismMaxReflections {
				width++
			}
			pos = &ally.pos
			return numReflections >= gamedata.PrismMaxReflections
		})
		if numReflections == gamedata.PrismMaxReflections {
			damage.Health += gamedata.PrismDamagePerMax
		}
		beam := newBeamNode(a.world(), ge.Pos{Base: pos}, ge.Pos{Base: target.GetPos()}, prismBeamColors[numReflections])
		beam.width = width
		a.world().nodeRunner.AddObject(beam)
		damage.Health *= damageMultiplier(target.GetTargetInfo(), a.stats.Weapon)
		target.OnDamage(damage, a)
		createEffect(a.world(), effectConfig{
			Pos:   target.GetPos().Add(a.world().localRand.Offset(-6, 6)),
			Image: a.stats.BeamExplosion,
			Layer: effectLayerFromBool(target.IsFlying()),
		})

	case gamedata.AgentDevourer:
		// Every consumed drone gives +1 to the power level.
		// Every power level gives +1 projectile (burst size).
		burstSize := a.stats.Weapon.BurstSize + a.extraLevel
		a.attackTargets(targets, burstSize)

	case gamedata.AgentRoomba:
		for _, target := range targets {
			creep, ok := target.(*creepNode)
			if ok {
				switch creep.stats.Kind {
				case gamedata.CreepWisp, gamedata.CreepWispLair:
					// Roombas don't attack wisps.
					continue
				}
			}
			a.attackWithProjectile(target, a.stats.Weapon.BurstSize)
		}

	default:
		a.attackTargets(targets, a.stats.Weapon.BurstSize)
	}

	playSound(a.world(), a.stats.Weapon.AttackSound, a.pos)
}

func (a *colonyAgentNode) damageReduction() float64 {
	extraReduction := 0.0
	if a.mode == agentModeBomberAttack {
		extraReduction = 0.25
	}
	return a.stats.DamageReduction + extraReduction
}

func (a *colonyAgentNode) movementSpeed() float64 {
	var baseSpeed float64
	switch a.mode {
	case agentModeFollowCommander:
		baseSpeed = a.speed + 25
	case agentModeKamikazeAttack:
		return 2 * a.speed
	case agentModeBomberAttack:
		baseSpeed = a.speed + 50
	case agentModeTakeoff, agentModeRecycleLanding:
		return 30
	case agentModePickup, agentModeResourceTakeoff, agentModeAlignStandby:
		baseSpeed = agentPickupSpeed
	default:
		baseSpeed = a.speed
	}
	multiplier := 1.0
	if a.resting {
		multiplier = 0.5
	}
	if a.slow > 0 {
		multiplier *= 0.55
	}
	if a.tether {
		multiplier *= 2
		baseSpeed += 5
	}
	return baseSpeed * multiplier
}

func (a *colonyAgentNode) moveTowardsWithSpeed(delta, speed float64) bool {
	// This method is slightly more optimized that from.MoveTowards(dest).
	travelled := speed * delta
	distSqr := a.pos.DistanceSquaredTo(a.waypoint)
	if distSqr < travelled*travelled || distSqr < gmath.Epsilon*gmath.Epsilon {
		a.changePos(a.waypoint)
		return true
	}
	a.changePos(a.pos.Add(a.dir.Mulf(travelled)))
	return false
}

func (a *colonyAgentNode) moveTowards(delta float64) bool {
	return a.moveTowardsWithSpeed(delta, a.movementSpeed())
}

func (a *colonyAgentNode) updateFollowCommander(delta float64) {
	if a.healthRegen != 0 {
		a.health = gmath.ClampMax(a.health+(delta*a.healthRegen), a.maxHealth)
	}

	commander := a.target.(*colonyAgentNode)
	if commander.IsDisposed() {
		a.target = nil
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}

	if a.moveTowards(delta) {
		a.waypointsLeft--
		if a.waypointsLeft == 0 {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		a.setWaypoint(a.orbitingWaypoint(commander.pos, a.dist))
	}
}

func (a *colonyAgentNode) updatePatrol(delta float64) {
	if a.healthRegen != 0 {
		a.health = gmath.ClampMax(a.health+(delta*a.healthRegen), a.maxHealth)
	}

	if a.moveTowards(delta) {
		a.waypointsLeft--
		if a.waypointsLeft == 0 {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		a.updatePatrolRadius()
		a.setWaypoint(a.orbitingWaypoint(a.colonyCore.pos, a.dist))
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
	height := a.shadowComponent.height + delta*30
	if a.moveTowards(delta) {
		height = agentFlightHeight
	}
	a.shadowComponent.UpdateHeight(a.pos, height, agentFlightHeight)
	if height == agentFlightHeight {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		a.shadowComponent.SetVisibility(true)
	}
}

func (a *colonyAgentNode) updateRecycleReturn(delta float64) {
	if a.moveTowards(delta) {
		a.colonyCore.openHatchTime = 1.5
		a.AssignMode(agentModeRecycleLanding, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateRepairBase(delta float64) {
	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			a.clearWaypoint()
			// TODO: use local rand and do not create this beam in simulation?
			buildPos := ge.Pos{
				Base:   &a.colonyCore.pos,
				Offset: gmath.Vec{X: a.scene.Rand().FloatRange(-18, 18)},
			}
			beam := newCloningBeamNode(a.world(), abeamCloning, &a.pos, buildPos)
			a.cloningBeam = beam
			a.world().nodeRunner.AddObject(beam)
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

func (a *colonyAgentNode) leaveForest() {
	if a.insideForest {
		a.insideForest = false
		a.setVisibility(true)
		createEffect(a.world(), effectConfig{
			Pos:            a.pos,
			Image:          assets.ImageDisappearSmokeSmall,
			AnimationSpeed: animationSpeedVerySlow,
		})
	}
}

func (a *colonyAgentNode) setVisibility(visible bool) {
	a.sprite.Visible = visible
	if a.diode != nil {
		a.diode.Visible = visible
	}
}

func (a *colonyAgentNode) handleForestTransition(nextWaypoint gmath.Vec) {
	if !a.world().hasForests {
		return
	}

	needEffect := false
	switch checkForestState(a.world(), a.insideForest, a.pos, nextWaypoint) {
	case forestStateEnter:
		needEffect = true
		a.insideForest = true
		a.setVisibility(false)
	case forestStateLeave:
		needEffect = true
		a.insideForest = false
		a.setVisibility(true)
	}

	if needEffect {
		createEffect(a.world(), effectConfig{
			Pos:            a.pos,
			Image:          assets.ImageDisappearSmokeSmall,
			AnimationSpeed: animationSpeedVerySlow,
		})
	}
}

func (a *colonyAgentNode) updateRelictTakeoff(delta float64) {
	height := a.shadowComponent.height + delta*30
	if a.moveTowards(delta) {
		height = agentFlightHeight
	}
	a.shadowComponent.UpdateHeight(a.pos, height, agentFlightHeight)
	if height == agentFlightHeight {
		a.mode = agentModeRelictPatrol
		a.shadowComponent.SetVisibility(true)
	}
}

func (a *colonyAgentNode) updateSiegeGuard(delta float64) {
	if a.specialDelay != 0 {
		return
	}

	turret := a.target.(*siegeTurretNode)
	if turret.ammo == 0 {
		a.specialDelay = a.world().rand.FloatRange(60, 70)
		turret.ammo = siegeTurretAmmo
		turret.target = nil
		return
	}
	if turret.target != nil && turret.target.IsDisposed() {
		turret.target = nil
		a.specialDelay = a.world().rand.FloatRange(5, 7)
		return
	}
	if turret.target == nil {
		targetCandidates := &a.world().tmpTargetSlice
		*targetCandidates = (*targetCandidates)[:0]
		a.world().WalkCreeps(a.pos, gamedata.SiegeAgentWeapon.AttackRange, func(creep *creepNode) bool {
			// Check for CanBeTargeted because a Howitzer may be crossing a forest.
			if !creep.stats.SiegeTargetable || !creep.CanBeTargeted() {
				return false
			}
			if creep.pos.DistanceSquaredTo(a.pos) > gamedata.SiegeAgentWeapon.AttackRangeSqr {
				return false
			}
			*targetCandidates = append(*targetCandidates, creep)
			return false
		})
		if len(*targetCandidates) != 0 {
			turret.target = gmath.RandElem(a.world().rand, *targetCandidates).(*creepNode)
		}
		if turret.target == nil {
			// Can't find a target.
			a.specialDelay = a.world().rand.FloatRange(5, 8)
			if turret.ammo < siegeTurretAmmo {
				turret.ammo++
				a.specialDelay *= 2
			}
			return
		}
	}

	if !turret.target.stats.Building {
		// If it can move, check whether it's too far or not.
		if turret.target.pos.DistanceSquaredTo(a.pos) > 1.1*gamedata.SiegeAgentWeapon.AttackRangeSqr {
			turret.target = nil
			a.specialDelay = a.world().rand.FloatRange(2, 4)
			return
		}
	}

	if turret.ammo%2 == 0 {
		a.specialDelay = 0.7
	} else {
		a.specialDelay = gamedata.SiegeAgentWeapon.Reload * a.world().rand.FloatRange(0.9, 1.1)
	}
	turret.ammo--

	fireOffset := turret.SetRotation(a.pos.AngleToPoint(turret.target.pos).Normalized())
	toPos := snipePos(gamedata.SiegeAgentWeapon.ProjectileSpeed, a.pos, turret.target.pos, turret.target.GetVelocity())
	p := a.world().newProjectileNode(projectileConfig{
		World:      a.world(),
		Weapon:     gamedata.SiegeAgentWeapon,
		Attacker:   a,
		ToPos:      toPos,
		Target:     turret.target,
		FireOffset: fireOffset,
	})
	a.world().nodeRunner.AddProjectile(p)
}

func (a *colonyAgentNode) updateRelictRoomba(delta float64) {
	if a.hasWaypoint() {
		if a.attackDelay < 1 && a.moveTowards(delta) {
			if a.path.HasNext() {
				// TODO: remove code duplication with crawlers.
				nextPos := nextPathWaypoint(a.world(), a.pos, &a.path, layerLandColony)
				a.setWaypoint(nextPos.Add(a.world().rand.Offset(-2, 2)))
				return
			}
			a.clearWaypoint()
			a.specialDelay = a.world().rand.FloatRange(2, 6)
		}
		return
	}

	if a.specialDelay > 0 {
		return
	}
	owner := a.target.(*colonyCoreNode)
	if a.pos.DistanceTo(owner.pos) <= 220 {
		a.specialDelay = a.world().rand.FloatRange(0.8, 4)
		return
	}

	dist := a.world().rand.FloatRange(40, 128)
	dst := owner.pos.Add(gmath.RadToVec(a.world().rand.Rad()).Mulf(dist))
	a.sendTo(a.pos.MoveTowards(dst, 350), layerLandColony)
}

func (a *colonyAgentNode) updateRelictPatrol(delta float64) {
	factory := a.target.(*colonyAgentNode)
	if factory.IsDisposed() {
		a.OnDamage(gamedata.DamageValue{Health: 1000}, a)
		return
	}

	a.energy = gmath.ClampMax(a.energy+delta*0.5*a.energyRegenRate, a.maxEnergy)
	a.health = gmath.ClampMax(a.health+(delta*a.healthRegen), a.maxHealth)

	if a.attackDelay > 1 {
		a.waypoint = gmath.Vec{}
		return
	}

	if a.waypoint.IsZero() {
		// a.dist is squared.
		if a.pos.DistanceSquaredTo(factory.waypoint) < a.dist {
			return
		}
		a.setWaypoint(factory.waypoint.Add(a.scene.Rand().Offset(-40, 40)))
	}
	if a.moveTowards(delta) {
		a.waypoint = gmath.Vec{}
	}
}

func (a *colonyAgentNode) updateRelictDroneFactory(delta float64) {
	// Hatch closing ticker.
	if a.lifetime > 0 {
		a.lifetime -= delta
		if a.lifetime <= 0 {
			hatch := a.target.(*ge.Sprite)
			a.lifetime = 0
			hatch.Visible = true
		}
	}
	if a.health < a.maxHealth*0.3 {
		// Keep the hatch open when heavily damaged.
		a.target.(*ge.Sprite).Visible = false
		a.lifetime = 1
		return
	}

	// Produce new units.
	const maxUnits = 4
	if a.specialDelay == 0 {
		if a.extraLevel >= maxUnits {
			a.specialDelay = a.scene.Rand().FloatRange(4, 8)
		} else {
			a.specialDelay = a.scene.Rand().FloatRange(10, 15)
			a.extraLevel++
			spawnPos := a.pos.Sub(gmath.Vec{Y: 12})
			unit := newColonyAgentNode(a.colonyCore, gamedata.RelictAgentStats, spawnPos)
			unit.mode = agentModeRelictTakeoff
			unit.target = a
			unit.dist = a.scene.Rand().FloatRange(20, 80)
			unit.dist *= unit.dist
			world := a.world()
			world.nodeRunner.AddObject(unit)
			world.mercs = append(world.mercs, unit)
			unit.setWaypoint(unit.pos.Sub(gmath.Vec{Y: agentFlightHeight}))
			unit.SetHeight(0)
			unit.shadowComponent.SetVisibility(false)
			unit.EventDestroyed.Connect(nil, func(*colonyAgentNode) {
				a.extraLevel--
				a.specialDelay += 5
				world.mercs = xslices.Remove(world.mercs, unit)
			})
			hatch := a.target.(*ge.Sprite)
			hatch.Visible = false
			a.lifetime = 2
		}
	}

	// Choose a new waypoint for troops.
	a.supportDelay = gmath.ClampMin(a.supportDelay-(delta*a.reloadRate), 0)
	if a.supportDelay == 0 {
		a.supportDelay = a.scene.Rand().FloatRange(6, 12)
		rect := gmath.Rect{
			Min: a.pos.Sub(gmath.Vec{X: 840, Y: 840}),
			Max: a.pos.Add(gmath.Vec{X: 840, Y: 840}),
		}
		a.waypoint = correctedPos(a.world().rect, randomSectorPos(a.scene.Rand(), rect), 64)
	}
}

func (a *colonyAgentNode) updateHarvester(delta float64) {
	var target *essenceSourceNode
	if a.target != nil {
		target = a.target.(*essenceSourceNode)
		if target.IsDisposed() {
			a.target = nil
			a.clearWaypoint()
			a.specialDelay = a.world().rand.FloatRange(4, 8)
			return
		}
	}

	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			if a.path.HasNext() {
				// TODO: remove code duplication with crawlers and roombas.
				nextPos := nextPathWaypoint(a.world(), a.pos, &a.path, layerNormal)
				a.handleForestTransition(nextPos)
				a.setWaypoint(nextPos.Add(a.world().rand.Offset(-4, 4)))
				return
			}
			a.path = pathing.GridPath{}
			a.leaveForest()
			dist := a.waypoint.DistanceTo(target.pos)
			if dist < 10 {
				a.clearWaypoint()
			} else {
				a.setWaypoint(target.pos.Add(a.world().rand.Offset(-4, 4)))
			}
		}
		return
	}

	if a.specialDelay > 0 {
		return
	}

	if target != nil {
		if a.colonyCore.resources >= a.colonyCore.maxVisualResources() {
			a.specialDelay = 4
			return
		}

		a.specialDelay = a.world().rand.FloatRange(4.5, 6.5)
		harvested := target.Harvest(2)
		value := float64(harvested) * target.stats.value
		a.colonyCore.AddGatheredResources(value)
		if a.health < a.maxHealth {
			a.OnBuildingRepair(1)
		}

		if !a.world().simulation {
			smokeRoll := a.world().localRand.Float()
			var sprite *ge.Sprite
			switch {
			case smokeRoll < 0.3:
				sprite = a.scene.NewSprite(assets.ImageSmokeDown)
				sprite.Pos.Offset = a.pos.Add(gmath.Vec{X: 1, Y: 16})
			case smokeRoll < 0.6:
				sprite = a.scene.NewSprite(assets.ImageSmokeSide)
				sprite.Pos.Offset = a.pos.Add(gmath.Vec{X: 20, Y: 11})
			case smokeRoll < 0.9:
				sprite = a.scene.NewSprite(assets.ImageSmokeSide)
				sprite.FlipHorizontal = true
				sprite.Pos.Offset = a.pos.Add(gmath.Vec{X: -16, Y: 11})
			default:
				// No smoke.
			}
			if sprite != nil {
				e := newEffectNodeFromSprite(a.world(), normalEffectLayer, sprite)
				e.noFlip = true
				e.anim.SetAnimationSpan(0.3)
				a.world().nodeRunner.AddObject(e)
				playSound(a.world(), assets.AudioHarvesterEffect, a.pos)
			}
		}

		return
	}

	var closestTarget *essenceSourceNode
	var closestSpot gmath.Vec
	closestDistSqr := math.MaxFloat64
	randIterate(a.world().rand, a.world().essenceSources, func(e *essenceSourceNode) bool {
		if !e.stats.harvesterTarget || e.beingHarvested {
			return false
		}
		coord := a.world().pathgrid.PosToCoord(e.pos)
		// Find a nearby closest cell.
		freeOffset := randIterate(a.world().rand, resourceNearOffsets, func(offset pathing.GridCoord) bool {
			probe := coord.Add(offset)
			return a.world().CellIsFree(probe, layerNormal)
		})
		if freeOffset.IsZero() {
			return false
		}
		distSqr := e.pos.DistanceSquaredTo(a.pos)
		if distSqr < closestDistSqr {
			closestDistSqr = distSqr
			closestTarget = e
			closestSpot = a.world().pathgrid.CoordToPos(coord.Add(freeOffset))
		}
		return false
	})
	if closestTarget == nil {
		a.specialDelay = a.world().rand.FloatRange(10, 30)
		return
	}

	a.specialDelay = 5
	a.target = closestTarget
	a.sendTo(closestSpot, layerNormal)
	closestTarget.beingHarvested = true
}

func (a *colonyAgentNode) updateRoombaWait(delta float64) {
	if a.healthRegen != 0 {
		a.health = gmath.ClampMax(a.health+(delta*a.healthRegen), a.maxHealth)
	}

	a.dist -= delta
	a.health = gmath.ClampMax(a.health+(delta*0.3), a.maxHealth)
	a.energy = gmath.ClampMax(a.energy+(delta*2.5), a.maxEnergy)
	if a.dist <= 0 {
		a.mode = agentModeRoombaPatrol
	}
}

func (a *colonyAgentNode) sendTo(pos gmath.Vec, l pathing.GridLayer) {
	p := a.world().BuildPath(a.pos, pos, l)
	a.path = p.Steps
	a.setWaypoint(a.world().pathgrid.AlignPos(a.pos))
}

func (a *colonyAgentNode) clearWaypoint() {
	a.waypoint = gmath.Vec{}
}

func (a *colonyAgentNode) hasWaypoint() bool {
	return !a.waypoint.IsZero()
}

func (a *colonyAgentNode) setWaypoint(pos gmath.Vec) {
	a.waypoint = pos
	a.dir = pos.DirectionTo(a.pos)
}

func (a *colonyAgentNode) updateRoombaPatrol(delta float64) {
	if a.energy < 0 {
		// Discharged. It needs some time to recover.
		a.mode = agentModeRoombaWait
		a.dist = a.scene.Rand().FloatRange(5, 25)
		a.clearWaypoint()
		a.energy = a.scene.Rand().FloatRange(10, 20)
		return
	}

	// Moving towards destination (or a target).
	if a.hasWaypoint() {
		a.energy -= 2.5 * delta
		if a.moveTowards(delta) {
			if a.target != nil {
				target := a.target.(*creepNode)
				if target.IsDisposed() {
					a.target = nil
				} else if a.pos.DistanceSquaredTo(target.pos) <= (a.stats.Weapon.AttackRangeSqr * a.supportDelay) {
					a.mode = agentModeRoombaCombatWait
					a.dist = a.scene.Rand().FloatRange(7, 16)
					a.clearWaypoint()
					a.target = nil
					return
				}
			}
			if a.path.HasNext() {
				if a.mode == agentModeRoombaPatrol && a.health < a.maxHealth*0.8 && a.scene.Rand().Chance(0.1) {
					a.health = gmath.ClampMax(a.health+2, a.maxHealth)
					a.mode = agentModeRoombaWait
					a.dist = a.scene.Rand().FloatRange(4, 10)
					a.clearWaypoint()
					if !a.world().simulation && !a.insideForest {
						createEffect(a.world(), effectConfig{
							Pos:            a.pos.Sub(gmath.Vec{Y: 10}),
							Image:          assets.ImageRoombaSmoke,
							AnimationSpeed: animationSpeedSlowest,
						})
					}
					return
				}
				// TODO: remove code duplication with crawlers.
				nextPos := nextPathWaypoint(a.world(), a.pos, &a.path, layerNormal)
				a.handleForestTransition(nextPos)
				a.setWaypoint(nextPos.Add(a.world().rand.Offset(-4, 4)))
				return
			}
			a.clearWaypoint()
		}
		return
	}

	if a.target != nil {
		target := a.target.(*creepNode)
		if target.IsDisposed() {
			a.target = nil
		} else {
			a.sendTo(target.pos.Add(a.scene.Rand().Offset(-80, 80)), layerNormal)
			return
		}
	}

	if a.mode == agentModeRoombaAttack {
		a.mode = agentModeRoombaCombatWait
		a.dist = a.scene.Rand().FloatRange(12, 20)
		return
	}
	if a.mode == agentModeRoombaGuard && a.scene.Rand().Chance(0.9) {
		a.mode = agentModeRoombaWait
		a.dist = a.scene.Rand().FloatRange(10, 25)
		return
	}

	if len(a.world().turrets) != 0 && a.scene.Rand().Chance(0.15) {
		turret := gmath.RandElem(a.scene.Rand(), a.world().turrets)
		a.sendTo(turret.pos.Add(a.scene.Rand().Offset(-80, 80)), layerNormal)
		a.mode = agentModeRoombaGuard
		return
	}

	// Try to find a new target.
	newTarget := randIterate(a.scene.Rand(), a.world().creeps, func(creep *creepNode) bool {
		switch creep.stats.Kind {
		case gamedata.CreepBase, gamedata.CreepCrawlerBase, gamedata.CreepTurret:
			return true
		case gamedata.CreepHowitzer:
			return creep.CanBeTargeted()
		default:
			return false
		}
	})
	if newTarget != nil {
		a.supportDelay = a.scene.Rand().FloatRange(0.4, 0.95)
		a.target = newTarget
		targetPos := newTarget.pos.Add(a.scene.Rand().Offset(-80, 80))
		if a.world().HasTreesAt(targetPos, 0) {
			targetPos = newTarget.pos.Add(a.scene.Rand().Offset(-120, 120))
		}
		a.sendTo(targetPos, layerNormal)
	} else {
		if a.scene.Rand().Chance(0.4) {
			targetPos := correctedPos(a.world().rect, randomSectorPos(a.scene.Rand(), a.world().rect), 480)
			a.sendTo(targetPos, layerNormal)
		} else {
			a.mode = agentModeRoombaWait
			a.dist = a.scene.Rand().FloatRange(2, 5)
		}
	}
}

func (a *colonyAgentNode) updateConsumeDrone(delta float64) {
	target := a.target.(*colonyAgentNode)
	if target.IsDisposed() || target.mode != agentModePosing {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}

	if a.moveTowards(delta) {
		// Give a partial drone cost refund.
		playSound(a.world(), assets.AudioAgentConsumed, a.pos)
		a.colonyCore.resources += target.stats.Cost * 0.5
		a.colonyCore.eliteResources += float64(target.rank)
		if a.extraLevel < gamedata.DevourerMaxLevel {
			a.extraLevel++
			a.maxHealth += 5
		}
		a.health = gmath.ClampMax(a.health+target.maxHealth*2, a.maxHealth)
		target.Destroy()
		a.world().nodeRunner.AddObject(newEffectNode(a.world(), a.pos, aboveEffectLayer, assets.ImageDroneConsumed))
	}
}

func (a *colonyAgentNode) updateBomberAttack(delta float64) {
	creep := a.target.(*creepNode)

	if creep.IsDisposed() {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}

	if !a.hasWaypoint() {
		a.setWaypoint(creep.pos.Sub(gmath.Vec{Y: agentFlightHeight}))
	}

	if a.moveTowards(delta) {
		bombingPos := a.waypoint.Add(gmath.Vec{Y: agentFlightHeight})
		if bombingPos.DistanceSquaredTo(creep.pos) > (36 * 36) {
			if a.waypointsLeft > 0 {
				a.waypointsLeft--
				nextPos := snipePos(a.movementSpeed(), a.pos, creep.pos, creep.GetVelocity())
				a.setWaypoint(nextPos.Sub(gmath.Vec{Y: agentPickupSpeed}))
			} else {
				a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			}
		} else {
			bomb := newBombNode(a, a.world())
			a.world().nodeRunner.AddObject(bomb)
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		}
	}
}

func (a *colonyAgentNode) updateKamikazeAttack(delta float64) {
	creep := a.target.(*creepNode)

	if creep.IsDisposed() {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}

	if a.dist > 0 {
		a.dist -= delta
		if a.dist <= 0 {
			a.dist = -1
			playSound(a.world(), assets.AudioKamizakeAttack, a.pos)
		}
	}

	if !a.hasWaypoint() {
		a.setWaypoint(a.getCloserWaypoint(creep.pos, 4, 16))
	}

	const explosionRangeSqr float64 = 40 * 40
	const explosionDamage float64 = 35.0
	const damageFlags = gamedata.DmgflagNoFlash | gamedata.DmgflagUnblockable
	if a.moveTowards(delta) {
		if a.pos.DistanceSquaredTo(creep.pos) > (explosionRangeSqr + 8) {
			a.clearWaypoint()
			return
		}
		createEffect(a.world(), effectConfig{
			Pos:      a.pos,
			Layer:    aboveEffectLayer,
			Image:    assets.ImageBigVerticalExplosion1,
			Rotation: a.pos.AngleToPoint(creep.pos) - math.Pi/2,
		})
		playSound(a.world(), assets.AudioExplosion1, a.pos)
		creep.OnDamage(gamedata.DamageValue{Health: explosionDamage, Flags: damageFlags}, a)
		for _, otherCreep := range a.world().creeps {
			if !otherCreep.IsFlying() || otherCreep == creep {
				continue
			}
			distSqr := otherCreep.pos.DistanceSquaredTo(a.pos)
			if distSqr > explosionRangeSqr {
				continue
			}
			otherCreep.OnDamage(gamedata.DamageValue{Health: explosionDamage * 0.5, Flags: damageFlags}, a)
		}
		a.Destroy()
		return
	}
}

func (a *colonyAgentNode) updateCaptureBuilding(delta float64) {
	if a.hasWaypoint() {
		target := a.target.(*neutralBuildingNode)
		if a.moveTowards(delta) {
			a.clearWaypoint()
			// TODO: use local rand and do not create this beam in simulation?
			buildPos := ge.Pos{
				Base:   &target.pos,
				Offset: gmath.Vec{X: a.scene.Rand().FloatRange(-10, 10)},
			}
			beam := newCloningBeamNode(a.world(), abeamCloning, &a.pos, buildPos)
			a.cloningBeam = beam
			a.world().nodeRunner.AddObject(beam)
			return
		}
		return
	}
	a.dist -= delta
	if a.dist <= 0 {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		target := a.target.(*neutralBuildingNode)
		if target.agent == nil {
			constructed := newColonyAgentNode(a.colonyCore, target.stats, target.pos)
			a.colonyCore.AcceptTurret(constructed)
			a.world().nodeRunner.AddObject(constructed)
			constructed.health = constructed.maxHealth * 0.01
			constructed.updateHealthShader()
			target.AssignAgent(constructed)
		}
		return
	}
}

func (a *colonyAgentNode) updateRepairTurret(delta float64) {
	target := a.target.(*colonyAgentNode)
	if target.IsDisposed() {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			// TODO: use local rand and do not create this beam in simulation?
			a.clearWaypoint()
			buildPos := ge.Pos{
				Base:   &target.pos,
				Offset: gmath.Vec{X: a.scene.Rand().FloatRange(-10, 10)},
			}
			beam := newCloningBeamNode(a.world(), abeamCloning, &a.pos, buildPos)
			a.cloningBeam = beam
			a.world().nodeRunner.AddObject(beam)
			return
		}
		return
	}
	a.dist -= delta
	if a.dist <= 0 {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		amountRepaired := a.scene.Rand().FloatRange(5, 8)
		if a.faction == gamedata.GreenFactionTag {
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
	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			target.attention += 2
			a.clearWaypoint()
			buildPos := target.GetConstructPos()
			beam := newCloningBeamNode(a.world(), abeamCloning, &a.pos, buildPos)
			a.cloningBeam = beam
			a.world().nodeRunner.AddObject(beam)
			return
		}
		return
	}
	amountConstructed := delta
	if a.faction == gamedata.GreenFactionTag {
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
	prevHeight := a.shadowComponent.height
	a.shadowComponent.UpdateHeight(a.pos, a.shadowComponent.height-delta*30, agentFlightHeight)
	darkenHeight := a.colonyCore.stats.DefaultHeight + 3
	if prevHeight >= darkenHeight && a.shadowComponent.height < darkenHeight {
		a.sprite.SetColorScaleRGBA(200, 200, 200, 255)
	}

	if a.moveTowards(delta) {
		a.colonyCore.resources += a.stats.Cost * 0.9
		if a.rank != 0 {
			a.colonyCore.eliteResources += float64(a.rank)
		}
		playSound(a.world(), assets.AudioAgentRecycled, a.pos)
		a.Destroy()
	}
}

func (a *colonyAgentNode) updateMerging(delta float64) {
	target := a.target.(*colonyAgentNode)
	if target.IsDisposed() || target.mode != a.mode {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		target.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		return
	}
	if !a.hasWaypoint() {
		dist := target.pos.DistanceTo(a.pos)
		if dist > 64 {
			a.setWaypoint(a.pos.MoveTowards(target.pos, dist-20).Add(a.scene.Rand().Offset(-8, 8)))
			return
		}
	}
	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			a.clearWaypoint()
		}
		return
	}
	if a.cloningBeam == nil {
		beam := newCloningBeamNode(a.world(), abeamMerging, &a.pos, ge.Pos{Base: &target.pos})
		a.cloningBeam = beam
		a.world().nodeRunner.AddObject(beam)
	}
	a.dist -= delta
	if a.pos.DistanceSquaredTo(target.pos) > (10 * 10) {
		a.changePos(a.pos.MoveTowards(target.pos, delta*12))
	} else {
		// Merging is x3 faster when units are next to each other.
		a.dist -= delta * 2
		if a.mode == agentModeMergingRoomba {
			if a.shadowComponent.height > 2 {
				descent := gmath.ClampMax(20*delta, a.shadowComponent.height-6)
				a.pos.Y += descent
				a.shadowComponent.UpdateHeight(a.pos, a.shadowComponent.height-descent, agentFlightHeight)
			}
		}
	}
	if a.dist <= 0 {
		mergingFailed := (a.mode == agentModeMergingRoomba && !posIsFreeWithFlags(a.world(), nil, a.pos, 2, collisionSkipSmallCrawlers|collisionSkipTeleporters)) ||
			(a.pos.DistanceSquaredTo(target.pos) > (30 * 30))
		if mergingFailed {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			target.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		a.cloningBeam.Dispose()
		a.cloningBeam = nil
		newStats := mergeAgents(a.world(), a, target)
		if newStats == nil {
			panic(fmt.Sprintf("empty merge result for %s %s + %s %s", a.faction, a.stats.Kind, target.faction, target.stats.Kind))
		}
		var newAgent *colonyAgentNode
		if a.mode == agentModeMergingRoomba {
			newAgent = newColonyAgentNode(a.colonyCore, newStats, target.pos)
			a.colonyCore.AcceptRoomba(newAgent)
		} else {
			newAgent = a.colonyCore.NewColonyAgentNode(newStats, target.pos)
		}
		var newFaction gamedata.FactionTag
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
		if newStats.Tier == 2 {
			newFaction = a.colonyCore.pickAgentFaction()
		} else {
			newFaction = a.faction
			if newFaction == gamedata.NeutralFactionTag || (target.faction != gamedata.NeutralFactionTag && a.faction != target.faction && a.scene.Rand().Bool()) {
				newFaction = target.faction
			}
		}
		newAgent.faction = newFaction
		a.world().nodeRunner.AddObject(newAgent)
		if newAgent.stats == gamedata.RoombaAgentStats {
			newAgent.mode = agentModeRoombaWait
			newAgent.dist = 1
		} else {
			newAgent.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		}
		target.Destroy()
		a.Destroy()
		createEffect(a.world(), effectConfig{
			Pos:     newAgent.pos,
			Layer:   effectLayerFromBool(newAgent.stats != gamedata.RoombaAgentStats),
			Image:   assets.ImageMergingComplete,
			Rotates: true,
		})
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
	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			a.clearWaypoint()
			beam := newCloningBeamNode(a.world(), abeamCloning, &a.pos, ge.Pos{Base: &target.pos})
			a.cloningBeam = beam
			a.world().nodeRunner.AddObject(beam)
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
		a.world().nodeRunner.AddObject(clone)
		a.world().result.DronesProduced++
		clone.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		createEffect(a.world(), effectConfig{
			Pos:     clone.pos,
			Layer:   aboveEffectLayer,
			Image:   assets.ImageCloningComplete,
			Rotates: true,
		})
		return
	}
}

func (a *colonyAgentNode) getCloserWaypoint(targetPos gmath.Vec, spread, preferredDist float64) gmath.Vec {
	currentDist := a.pos.DistanceTo(targetPos)
	if currentDist <= preferredDist {
		return targetPos.Add(a.scene.Rand().Offset(-spread, spread))
	}
	const maxMoveDist float64 = 96.0
	dist := gmath.ClampMax(maxMoveDist*a.scene.Rand().FloatRange(0.8, 1.2), gmath.ClampMin(currentDist-preferredDist, maxMoveDist*0.25))
	result := targetPos.DirectionTo(a.pos).Mulf(dist).Add(a.pos).Add(a.scene.Rand().Offset(-28, 28))
	return result
}

func (a *colonyAgentNode) followWaypoint(targetPos gmath.Vec) gmath.Vec {
	rng := a.stats.Weapon.AttackRange * 0.8
	preferredDist := gmath.ClampMin(rng, 80)
	return a.getCloserWaypoint(targetPos, rng, preferredDist)
}

func (a *colonyAgentNode) updateMove(delta float64) {
	if a.moveTowards(delta) {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updatePanic(delta float64) {
	if a.moveTowards(delta) {
		a.waypointsLeft--
		a.clearWaypoint()
	}

	if !a.hasWaypoint() {
		if a.waypointsLeft <= 0 {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		waypoint := a.pos.Add(a.scene.Rand().Offset(-32, 32))
		a.setWaypoint(correctedPos(a.world().rect, waypoint, 64))
	}
}

func (a *colonyAgentNode) updateCourierFlight(delta float64) {
	if a.moveTowards(delta) {
		target := a.target.(*colonyCoreNode)
		if target.IsDisposed() || target.mode != colonyModeNormal {
			if a.payload != 0 {
				// Has some payload, should return it back.
				a.AssignMode(agentModeReturn, gmath.Vec{}, nil)
			} else {
				a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			}
			return
		}
		if target.pos.DistanceSquaredTo(a.pos) < (70 * 70) {
			if a.payload != 0 {
				target.resources += a.cargoValue
				a.clearCargo()
			}
			beam := newTextureBeamNode(a.world(), ge.Pos{Base: &a.pos}, ge.Pos{Base: &target.pos}, a.stats.BeamTexture, a.stats.BeamSlideSpeed, a.stats.BeamOpaqueTime)
			a.world().nodeRunner.AddObject(beam)
			playSound(a.world(), assets.AudioCourierResourceBeam, a.pos)
			dist := target.pos.DistanceTo(a.colonyCore.pos)
			// Now go back and bring some resources.
			if dist > 115 {
				a.payload = a.maxPayload()
				a.cargoValue = float64(a.payload) * (math.Trunc((dist-115)/15) * 0.2)
				a.cargoValue = gmath.ClampMax(a.cargoValue, 12)
				a.AssignMode(agentModeReturn, gmath.Vec{}, nil)
			} else {
				a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			}
			// Get a minor repair+recharge.
			a.health = gmath.ClampMax(a.health+5, a.maxHealth)
			a.energy = gmath.ClampMax(a.energy+15, a.maxEnergy)
			return
		}
		// TODO: use a followPos here?
		a.setWaypoint(a.pos.DirectionTo(target.pos).Mulf(60).Add(target.pos).Add(a.scene.Rand().Offset(-20, 20)))
	}
}

func (a *colonyAgentNode) updateFollow(delta float64) {
	if a.moveTowards(delta) {
		target := a.target.(*creepNode)
		if a.waypointsLeft == 0 || target.IsDisposed() || !target.CanBeTargeted() {
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
			return
		}
		a.waypointsLeft--
		a.setWaypoint(a.followWaypoint(target.pos))
	}
}

func (a *colonyAgentNode) updateAlignStandby(delta float64) {
	speed := a.movementSpeed()
	height := a.shadowComponent.height + delta*speed
	if a.moveTowardsWithSpeed(delta, speed) {
		height = agentFlightHeight
	}
	a.shadowComponent.UpdateHeight(a.pos, height, agentFlightHeight)
	if height == agentFlightHeight {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateStandby(delta float64) {
	if a.healthRegen != 0 {
		a.health = gmath.ClampMax(a.health+(delta*a.healthRegen), a.maxHealth)
	}

	a.energy = gmath.ClampMax(a.energy+delta*0.5*a.energyRegenRate, a.maxEnergy)
	if a.moveTowards(delta) {
		if a.stats.Tier == 1 && a.lifetime < 0 && a.colonyCore.mode == colonyModeNormal {
			a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
			return
		}
		a.setWaypoint(a.orbitingWaypoint(a.colonyCore.pos, a.dist))
		if a.hasTrait(traitAdventurer) {
			a.waypointsLeft++
			if a.waypointsLeft > 10 {
				a.waypointsLeft = 0
				traitRoll := a.scene.Rand().Float()
				if a.hasTrait(traitAdventurer) && traitRoll <= 0.4 {
					pos := a.pos.Add(a.scene.Rand().Offset(-100, 100))
					a.AssignMode(agentModeMove, pos, nil)
					// Add some energy to compensate for this unproductive behavior.
					a.energy = gmath.ClampMax(a.energy+5, a.maxEnergy)
					a.energyBill = gmath.ClampMin(a.energyBill-5, 0)
					return
				}
			}
		}
		if a.colonyCore.mode == colonyModeNormal && !a.hasTrait(traitNeverStop) && a.energy < 40 && a.scene.Rand().Chance(0.2) {
			a.energyBill *= 0.5
			a.AssignMode(agentModeCharging, gmath.Vec{}, nil)
			return
		}
	}
}

func (a *colonyAgentNode) updateCloakHide(delta float64) {
	a.energy = gmath.ClampMax(a.energy+delta*a.energyRegenRate, a.maxEnergy)
	if a.cloaking <= 0 {
		a.health = gmath.ClampMax(a.health+2, a.maxHealth)
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateCharging(delta float64) {
	a.energy = gmath.ClampMax(a.energy+delta*4*a.energyRegenRate, a.maxEnergy)
	if a.energy >= a.maxEnergy*0.55 {
		a.energyBill = 0
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updatePosing(delta float64) {
	a.dist -= delta
	if a.dist <= 0 {
		a.dist = 0
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateForcedCharging(delta float64) {
	a.energy = gmath.ClampMax(a.energy+delta*2.0*a.energyRegenRate, a.maxEnergy)
	if a.energy >= a.energyTarget {
		a.energyTarget = 0
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateMineSulfurEssence(delta float64) {
	if a.hasWaypoint() {
		if a.moveTowards(delta) {
			a.waypoint = gmath.Vec{}
			if !a.world().simulation {
				source := a.target.(*essenceSourceNode)
				offset := a.world().localRand.Offset(-6, 6)
				beam := newCloningBeamNode(a.world(), abeamSulfurMining, &a.pos, ge.Pos{Base: &source.pos, Offset: offset})
				a.cloningBeam = beam
				a.world().nodeRunner.AddObject(beam)
			}
		}
		return
	}

	a.dist -= delta
	if a.tether {
		// The gather speed is doubled.
		a.dist -= delta
	}
	if a.dist <= 0 {
		source := a.target.(*essenceSourceNode)
		harvested := source.Harvest(a.maxPayload())
		a.payload = harvested
		a.cargoValue = float64(harvested) * source.stats.value
		if a.cloningBeam != nil {
			a.cloningBeam.Dispose()
			a.cloningBeam = nil
		}
		a.AssignMode(agentModeReturn, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) updateMineEssence(delta float64) {
	if a.moveTowards(delta) {
		source := a.target.(*essenceSourceNode)
		if source.IsDisposed() {
			if a.IsCloaked() {
				a.doUncloak()
			}
			a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
		} else {
			a.AssignMode(agentModePickup, gmath.Vec{}, nil)
		}
	}
}

func (a *colonyAgentNode) updatePickup(delta float64) {
	speed := a.movementSpeed()
	height := a.shadowComponent.height - delta*speed
	if a.moveTowardsWithSpeed(delta, speed) {
		height = 0
		a.mode = agentModeResourceTakeoff
		a.setWaypoint(a.pos.Sub(gmath.Vec{Y: agentFlightHeight}))
		source := a.target.(*essenceSourceNode)
		harvested := source.Harvest(a.maxPayload())
		a.payload = harvested
		a.cargoValue = float64(harvested) * source.stats.value
		a.cargoEliteValue = float64(harvested) * source.stats.eliteValue
	}
	a.shadowComponent.UpdateHeight(a.pos, height, agentFlightHeight)
}

func (a *colonyAgentNode) updateResourceTakeoff(delta float64) {
	speed := a.movementSpeed()
	height := a.shadowComponent.height + delta*speed
	if a.moveTowardsWithSpeed(delta, speed) {
		height = agentFlightHeight
	}
	a.shadowComponent.UpdateHeight(a.pos, height, agentFlightHeight)
	if height == agentFlightHeight {
		a.AssignMode(agentModeReturn, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) clearCargo() {
	a.payload = 0
	a.cargoValue = 0
	a.cargoEliteValue = 0
}

func (a *colonyAgentNode) updateReturn(delta float64) {
	if a.moveTowards(delta) {
		if a.IsCloaked() {
			a.doUncloak()
		}
		if a.payload != 0 {
			a.colonyCore.AddGatheredResources(a.cargoValue)
			a.colonyCore.eliteResources += a.cargoEliteValue
			a.world().result.EliteResourcesGathered = a.cargoEliteValue
			a.clearCargo()
			playSound(a.world(), assets.AudioEssenceCollected, a.pos)
		}
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
}

func (a *colonyAgentNode) world() *worldState {
	return a.colonyCore.world
}

func (a *colonyAgentNode) hasTrait(t agentTraitBits) bool {
	return a.traits&t != 0
}

func (a *colonyAgentNode) maxPayload() int {
	n := a.stats.MaxPayload
	if a.faction == gamedata.YellowFactionTag {
		n += 2
	}
	return n
}

func (a *colonyAgentNode) changePos(pos gmath.Vec) {
	a.pos = pos
	a.shadowComponent.UpdatePos(pos)
}
