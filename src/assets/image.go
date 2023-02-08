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
		ImageColonyCoreSelector: {Path: "image/colony_core_selector.png"},
		ImageColonyCore:         {Path: "image/colony_core.png"},
		ImageColonyCoreFlying:   {Path: "image/colony_core_flying.png"},
		ImageColonyCoreHatch:    {Path: "image/colony_core_hatch.png"},
		ImageColonyCoreShadow:   {Path: "image/colony_core_shadow.png"},
		ImageWorkerAgent:        {Path: "image/worker_agent.png", FrameWidth: 9},
		ImageMilitiaAgent:       {Path: "image/militia_agent.png", FrameWidth: 11},
		ImageRepairAgent:        {Path: "image/repair_agent.png", FrameWidth: 17},
		ImageRechargerAgent:     {Path: "image/recharger_agent.png", FrameWidth: 17},
		ImageFighterAgent:       {Path: "image/fighter_agent.png", FrameWidth: 17},
		ImageRepellerAgent:      {Path: "image/repeller_agent.png", FrameWidth: 15},
		ImageFreighterAgent:     {Path: "image/freighter_agent.png", FrameWidth: 17},
		ImageGeneratorAgent:     {Path: "image/generator_agent.png", FrameWidth: 15},

		ImageEssenceCrystalSource:      {Path: "image/crystal_source.png"},
		ImageEssenceGoldSource:         {Path: "image/gold_source.png"},
		ImageEssenceIronSource:         {Path: "image/iron_source.png"},
		ImageEssenceWasteSource:        {Path: "image/waste_source.png"},
		ImageEssenceScrapSource:        {Path: "image/scrap_source.png"},
		ImageEssenceSmallScrapSource:   {Path: "image/small_scrap_source.png"},
		ImageEssenceSource:             {Path: "image/essence_source.png"},
		ImageEssenceSourceDissolveMask: {Path: "image/essence_source_dissolve_mask.png"},

		ImagePrimitiveCreep:  {Path: "image/primitive_creep.png"},
		ImagePrimitiveCreep2: {Path: "image/primitive_creep2.png"},

		ImageBackgroundTiles: {Path: "image/tiles.png"},

		ImageMilitiaProjectile:         {Path: "image/militia_projectile.png"},
		ImageRepellerProjectile:        {Path: "image/repeller_projectile.png"},
		ImageFighterProjectile:         {Path: "image/fighter_projectile.png"},
		ImageWandererProjectile:        {Path: "image/wanderer_projectile.png"},
		ImageWandererStunnerProjectile: {Path: "image/wanderer_stunner_projectile.png"},
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

	ImageChoiceWindow
	ImageChoiceRechargeWindow

	ImageSmallExplosion1

	ImageButtonX
	ImageButtonY
	ImageButtonA
	ImageButtonB
	ImageButtonRB

	ImageFactionDiode
	ImageUberBoss
	ImageUberBossShadow
	ImageColonyCoreSelector
	ImageColonyCore
	ImageColonyCoreFlying
	ImageColonyCoreHatch
	ImageColonyCoreShadow
	ImageWorkerAgent
	ImageGeneratorAgent
	ImageMilitiaAgent
	ImageRepairAgent
	ImageRechargerAgent
	ImageFighterAgent
	ImageRepellerAgent
	ImageFreighterAgent
	ImageEssenceCrystalSource
	ImageEssenceGoldSource
	ImageEssenceIronSource
	ImageEssenceWasteSource
	ImageEssenceScrapSource
	ImageEssenceSmallScrapSource
	ImageEssenceSource
	ImageEssenceSourceDissolveMask
	ImagePrimitiveCreep
	ImagePrimitiveCreep2

	ImageBackgroundTiles

	ImageMilitiaProjectile
	ImageRepellerProjectile
	ImageFighterProjectile
	ImageWandererProjectile
	ImageWandererStunnerProjectile
)
