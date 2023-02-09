package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func registerImageResources(ctx *ge.Context) {
	imageResources := map[resource.ImageID]resource.ImageInfo{
		ImageSmallShadow:  {Path: "image/small_shadow.png"},
		ImageMediumShadow: {Path: "image/medium_shadow.png"},
		ImageBigShadow:    {Path: "image/big_shadow.png"},

		ImageChoiceWindow:         {Path: "image/choice_window.png"},
		ImageChoiceRechargeWindow: {Path: "image/choice_recharge_window.png"},

		ImageSmallExplosion1: {Path: "image/small_explosion1.png", FrameWidth: 32},

		ImageButtonX:  {Path: "image/button_x.png"},
		ImageButtonY:  {Path: "image/button_y.png"},
		ImageButtonA:  {Path: "image/button_a.png"},
		ImageButtonB:  {Path: "image/button_b.png"},
		ImageButtonRB: {Path: "image/button_rb.png"},

		ImageFactionDiode:       {Path: "image/faction_diode.png"},
		ImageUberBoss:           {Path: "image/uber_boss.png"},
		ImageUberBossShadow:     {Path: "image/uber_boss_shadow.png"},
		ImageCreepBase:          {Path: "image/creep_base.png", FrameWidth: 32},
		ImageColonyCoreSelector: {Path: "image/colony_core_selector.png"},
		ImageColonyCore:         {Path: "image/colony_core.png"},
		ImageColonyCoreFlying:   {Path: "image/colony_core_flying.png"},
		ImageColonyCoreHatch:    {Path: "image/colony_core_hatch.png"},
		ImageColonyCoreShadow:   {Path: "image/colony_core_shadow.png"},
		ImageWorkerAgent:        {Path: "image/worker_agent.png", FrameWidth: 9},
		ImageMilitiaAgent:       {Path: "image/militia_agent.png", FrameWidth: 11},
		ImageCripplerAgent:      {Path: "image/crippler_agent.png", FrameWidth: 19},
		ImageFlamerAgent:        {Path: "image/flamer_agent.png", FrameWidth: 21},
		ImageRepairAgent:        {Path: "image/repair_agent.png", FrameWidth: 17},
		ImageRechargerAgent:     {Path: "image/recharger_agent.png", FrameWidth: 17},
		ImageFighterAgent:       {Path: "image/fighter_agent.png", FrameWidth: 17},
		ImageRepellerAgent:      {Path: "image/repeller_agent.png", FrameWidth: 15},
		ImageFreighterAgent:     {Path: "image/freighter_agent.png", FrameWidth: 17},
		ImageRedminerAgent:      {Path: "image/redminer_agent.png", FrameWidth: 13},
		ImageGeneratorAgent:     {Path: "image/generator_agent.png", FrameWidth: 15},

		ImageEssenceSourceDissolveMask: {Path: "image/essence_source_dissolve_mask.png"},
		ImageColonyDamageMask:          {Path: "image/colony_damage_mask.png"},

		ImageEssenceCrystalSource:    {Path: "image/crystal_source.png"},
		ImageEssenceGoldSource:       {Path: "image/gold_source.png", FrameWidth: 28},
		ImageEssenceIronSource:       {Path: "image/iron_source.png", FrameWidth: 32},
		ImageEssenceWasteSource:      {Path: "image/waste_source.png"},
		ImageEssenceScrapSource:      {Path: "image/scrap_source.png"},
		ImageEssenceSmallScrapSource: {Path: "image/small_scrap_source.png"},
		ImageEssenceBigScrapSource:   {Path: "image/big_scrap_source.png"},
		ImageEssenceSource:           {Path: "image/essence_source.png"},
		ImageRedEssenceSource:        {Path: "image/red_essence_source.png"},

		ImagePrimitiveCreep: {Path: "image/primitive_creep.png"},
		ImageCreepTier2:     {Path: "image/tier2_creep.png"},
		ImageCreepTier3:     {Path: "image/tier3_creep.png", FrameWidth: 25},
		ImageTurretCreep:    {Path: "image/turret_creep.png"},
		ImageLandCreep:      {Path: "image/land_creep.png"},

		ImageBackgroundTiles: {Path: "image/tiles.png"},

		ImageTankProjectile:     {Path: "image/tank_projectile.png"},
		ImageAssaultProjectile:  {Path: "image/assault_projectile.png"},
		ImageCripplerProjectile: {Path: "image/crippler_projectile.png"},
		ImageMilitiaProjectile:  {Path: "image/militia_projectile.png"},
		ImageRepellerProjectile: {Path: "image/repeller_projectile.png"},
		ImageFighterProjectile:  {Path: "image/fighter_projectile.png"},
		ImageWandererProjectile: {Path: "image/wanderer_projectile.png"},
		ImageFlamerProjectile:   {Path: "image/flamer_projectile.png"},
		ImageMissile:            {Path: "image/missile.png"},
	}

	for id, res := range imageResources {
		ctx.Loader.ImageRegistry.Set(id, res)
		ctx.Loader.LoadImage(id)
	}
}

const (
	ImageNone resource.ImageID = iota

	ImageSmallShadow
	ImageMediumShadow
	ImageBigShadow

	ImageChoiceWindow
	ImageChoiceRechargeWindow

	ImageSmallExplosion1

	ImageColonyDamageMask
	ImageEssenceSourceDissolveMask

	ImageButtonX
	ImageButtonY
	ImageButtonA
	ImageButtonB
	ImageButtonRB

	ImageFactionDiode
	ImageUberBoss
	ImageUberBossShadow
	ImageCreepBase
	ImageColonyCoreSelector
	ImageColonyCore
	ImageColonyCoreFlying
	ImageColonyCoreHatch
	ImageColonyCoreShadow
	ImageWorkerAgent
	ImageGeneratorAgent
	ImageMilitiaAgent
	ImageFlamerAgent
	ImageRepairAgent
	ImageRechargerAgent
	ImageFighterAgent
	ImageCripplerAgent
	ImageRedminerAgent
	ImageRepellerAgent
	ImageFreighterAgent
	ImageEssenceCrystalSource
	ImageEssenceGoldSource
	ImageEssenceIronSource
	ImageEssenceWasteSource
	ImageEssenceScrapSource
	ImageEssenceSmallScrapSource
	ImageEssenceBigScrapSource
	ImageEssenceSource
	ImageRedEssenceSource
	ImagePrimitiveCreep
	ImageCreepTier2
	ImageCreepTier3
	ImageTurretCreep
	ImageLandCreep

	ImageBackgroundTiles

	ImageTankProjectile
	ImageAssaultProjectile
	ImageCripplerProjectile
	ImageMilitiaProjectile
	ImageRepellerProjectile
	ImageFighterProjectile
	ImageWandererProjectile
	ImageFlamerProjectile
	ImageMissile
)
