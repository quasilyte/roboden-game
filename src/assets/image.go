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

		ImageUpkeepBar: {Path: "image/upkeep_bar.png"},

		ImageChoiceWindow:         {Path: "image/choice_window.png"},
		ImageChoiceRechargeWindow: {Path: "image/choice_recharge_window.png"},
		ImageTutorialDialogue:     {Path: "image/window_tutorial.png"},

		ImageRadar:         {Path: "image/radar.png"},
		ImageRadarWave:     {Path: "image/radar_wave.png"},
		ImageRadarBossFar:  {Path: "image/radar_boss_far.png"},
		ImageRadarBossNear: {Path: "image/radar_boss_near.png"},

		ImageSmallExplosion1:   {Path: "image/small_explosion1.png", FrameWidth: 32},
		ImageVerticalExplosion: {Path: "image/vertical_explosion.png", FrameWidth: 50},
		ImageBigExplosion:      {Path: "image/big_explosion.png", FrameWidth: 64},

		ImageLogoBg:     {Path: "image/logo_bg.png"},
		ImageYellowLogo: {Path: "image/yellow_logo.png"},
		ImageRedLogo:    {Path: "image/red_logo.png"},
		ImageGreenLogo:  {Path: "image/green_logo.png"},
		ImageBlueLogo:   {Path: "image/blue_logo.png"},

		ImageFactionDiode:       {Path: "image/faction_diode.png"},
		ImageUberBoss:           {Path: "image/uber_boss.png", FrameWidth: 40},
		ImageUberBossShadow:     {Path: "image/uber_boss_shadow.png"},
		ImageCreepBase:          {Path: "image/creep_base.png", FrameWidth: 32},
		ImageColonyCoreSelector: {Path: "image/colony_core_selector.png"},
		ImageColonyCore:         {Path: "image/colony_core.png"},
		ImageColonyCoreFlying:   {Path: "image/colony_core_flying.png"},
		ImageColonyCoreHatch:    {Path: "image/colony_core_hatch.png"},
		ImageColonyCoreDiode:    {Path: "image/colony_core_diode.png", FrameWidth: 4},
		ImageColonyCoreShadow:   {Path: "image/colony_core_shadow.png"},

		ImageWorkerAgent:    {Path: "image/drones/worker_agent.png", FrameWidth: 9},
		ImageMilitiaAgent:   {Path: "image/drones/militia_agent.png", FrameWidth: 11},
		ImageCripplerAgent:  {Path: "image/drones/crippler_agent.png", FrameWidth: 15},
		ImageFlamerAgent:    {Path: "image/drones/flamer_agent.png", FrameWidth: 21},
		ImageRepairAgent:    {Path: "image/drones/repair_agent.png", FrameWidth: 17},
		ImageServoAgent:     {Path: "image/drones/servo_agent.png", FrameWidth: 15},
		ImageRechargerAgent: {Path: "image/drones/recharger_agent.png", FrameWidth: 17},
		ImageRefresherAgent: {Path: "image/drones/refresher_agent.png", FrameWidth: 31},
		ImageFighterAgent:   {Path: "image/drones/fighter_agent.png", FrameWidth: 15},
		ImageDestroyerAgent: {Path: "image/drones/destroyer_agent.png", FrameWidth: 33},
		ImageRepellerAgent:  {Path: "image/drones/repeller_agent.png", FrameWidth: 15},
		ImageFreighterAgent: {Path: "image/drones/freighter_agent.png", FrameWidth: 17},
		ImageRedminerAgent:  {Path: "image/drones/redminer_agent.png", FrameWidth: 13},
		ImageGeneratorAgent: {Path: "image/drones/generator_agent.png", FrameWidth: 15},

		ImageEssenceSourceDissolveMask: {Path: "image/essence_source_dissolve_mask.png"},
		ImageColonyDamageMask:          {Path: "image/colony_damage_mask.png"},

		ImageEssenceCrystalSource:         {Path: "image/crystal_source.png", FrameWidth: 16},
		ImageEssenceGoldSource:            {Path: "image/gold_source.png", FrameWidth: 28},
		ImageEssenceIronSource:            {Path: "image/iron_source.png", FrameWidth: 32},
		ImageEssenceWasteSource:           {Path: "image/waste_source.png"},
		ImageEssenceScrapSource:           {Path: "image/scrap_source.png"},
		ImageEssenceSmallScrapSource:      {Path: "image/small_scrap_source.png"},
		ImageEssenceScrapCreepSource:      {Path: "image/scrap_source_creep.png"},
		ImageEssenceSmallScrapCreepSource: {Path: "image/small_scrap_source_creep.png"},
		ImageEssenceBigScrapCreepSource:   {Path: "image/big_scrap_source_creep.png"},
		ImageEssenceSource:                {Path: "image/essence_source.png", FrameWidth: 32},
		ImageRedEssenceSource:             {Path: "image/red_essence_source.png", FrameWidth: 32},

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

		ImageUIButtonIdle:             {Path: "image/ui/button-idle.png"},
		ImageUIButtonHover:            {Path: "image/ui/button-hover.png"},
		ImageUIButtonPressed:          {Path: "image/ui/button-pressed.png"},
		ImageUIButtonDisabled:         {Path: "image/ui/button-disabled.png"},
		ImageUIButtonSelectedIdle:     {Path: "image/ui/button-selected-idle.png"},
		ImageUIButtonSelectedHover:    {Path: "image/ui/button-selected-hover.png"},
		ImageUIButtonSelectedPressed:  {Path: "image/ui/button-selected-pressed.png"},
		ImageUIButtonSelectedDisabled: {Path: "image/ui/button-selected-disabled.png"},
		ImageUIArrowDownIdle:          {Path: "image/ui/arrow-down-idle.png"},
		ImageUIArrowDownDisabled:      {Path: "image/ui/arrow-down-disabled.png"},
		ImageUIOptionButtonIdle:       {Path: "image/ui/combo-button-idle.png"},
		ImageUIOptionButtonHover:      {Path: "image/ui/combo-button-hover.png"},
		ImageUIOptionButtonPressed:    {Path: "image/ui/combo-button-pressed.png"},
		ImageUIOptionButtonDisabled:   {Path: "image/ui/combo-button-disabled.png"},
		ImageUIListIdle:               {Path: "image/ui/list-idle.png"},
		ImageUIListDisabled:           {Path: "image/ui/list-disabled.png"},
		ImageUIListMask:               {Path: "image/ui/list-mask.png"},
		ImageUIListTrackIdle:          {Path: "image/ui/list-track-idle.png"},
		ImageUIListTrackDisabled:      {Path: "image/ui/list-track-disabled.png"},
		ImageUISliderHandleIdle:       {Path: "image/ui/slider-handle-idle.png"},
		ImageUISliderHandleHover:      {Path: "image/ui/slider-handle-hover.png"},
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
	ImageTutorialDialogue

	ImageRadar
	ImageRadarWave
	ImageRadarBossFar
	ImageRadarBossNear

	ImageSmallExplosion1
	ImageVerticalExplosion
	ImageBigExplosion

	ImageUpkeepBar

	ImageColonyDamageMask
	ImageEssenceSourceDissolveMask

	ImageLogoBg
	ImageYellowLogo
	ImageRedLogo
	ImageGreenLogo
	ImageBlueLogo

	ImageFactionDiode
	ImageUberBoss
	ImageUberBossShadow
	ImageCreepBase
	ImageColonyCoreSelector
	ImageColonyCore
	ImageColonyCoreFlying
	ImageColonyCoreHatch
	ImageColonyCoreDiode
	ImageColonyCoreShadow
	ImageWorkerAgent
	ImageGeneratorAgent
	ImageMilitiaAgent
	ImageFlamerAgent
	ImageRepairAgent
	ImageRechargerAgent
	ImageRefresherAgent
	ImageFighterAgent
	ImageDestroyerAgent
	ImageCripplerAgent
	ImageRedminerAgent
	ImageRepellerAgent
	ImageServoAgent
	ImageFreighterAgent
	ImageEssenceCrystalSource
	ImageEssenceGoldSource
	ImageEssenceIronSource
	ImageEssenceWasteSource
	ImageEssenceScrapSource
	ImageEssenceSmallScrapSource
	ImageEssenceScrapCreepSource
	ImageEssenceSmallScrapCreepSource
	ImageEssenceBigScrapCreepSource
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

	ImageUIButtonIdle
	ImageUIButtonHover
	ImageUIButtonPressed
	ImageUIButtonDisabled
	ImageUIButtonSelectedIdle
	ImageUIButtonSelectedHover
	ImageUIButtonSelectedPressed
	ImageUIButtonSelectedDisabled
	ImageUIArrowDownIdle
	ImageUIArrowDownDisabled
	ImageUIOptionButtonIdle
	ImageUIOptionButtonHover
	ImageUIOptionButtonPressed
	ImageUIOptionButtonDisabled
	ImageUIListIdle
	ImageUIListDisabled
	ImageUIListMask
	ImageUIListTrackIdle
	ImageUIListTrackDisabled
	ImageUISliderHandleIdle
	ImageUISliderHandleHover
)
