package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func registerImageResources(ctx *ge.Context) {
	imageResources := map[resource.ImageID]resource.ImageInfo{
		ImageSmallShadow:    {Path: "image/small_shadow.png"},
		ImageMediumShadow:   {Path: "image/medium_shadow.png"},
		ImageBigShadow:      {Path: "image/big_shadow.png"},
		ImageUberBossShadow: {Path: "image/uber_boss_shadow.png"},

		ImageUpkeepBar: {Path: "image/upkeep_bar.png"},

		ImageChoiceWindow:         {Path: "image/ui/choice_window.png"},
		ImageChoiceRechargeWindow: {Path: "image/ui/choice_recharge_window.png"},
		ImageTutorialDialogue:     {Path: "image/ui/window_tutorial.png"},

		ImageRadar:         {Path: "image/ui/radar.png"},
		ImageRadarWave:     {Path: "image/ui/radar_wave.png"},
		ImageRadarBossFar:  {Path: "image/ui/radar_boss_far.png"},
		ImageRadarBossNear: {Path: "image/ui/radar_boss_near.png"},

		ImageLogoBg:     {Path: "image/ui/logo_bg.png"},
		ImageYellowLogo: {Path: "image/ui/yellow_logo.png"},
		ImageRedLogo:    {Path: "image/ui/red_logo.png"},
		ImageGreenLogo:  {Path: "image/ui/green_logo.png"},
		ImageBlueLogo:   {Path: "image/ui/blue_logo.png"},

		ImageSmallExplosion1:   {Path: "image/effects/small_explosion1.png", FrameWidth: 32},
		ImageVerticalExplosion: {Path: "image/effects/vertical_explosion.png", FrameWidth: 50},
		ImageBigExplosion:      {Path: "image/effects/big_explosion.png", FrameWidth: 64},

		ImageFactionDiode:       {Path: "image/faction_diode.png"},
		ImageColonyCoreSelector: {Path: "image/colony_core_selector.png"},
		ImageColonyCore:         {Path: "image/colony_core.png"},
		ImageColonyCoreFlying:   {Path: "image/colony_core_flying.png"},
		ImageColonyCoreHatch:    {Path: "image/colony_core_hatch.png"},
		ImageColonyCoreDiode:    {Path: "image/colony_core_diode.png", FrameWidth: 4},
		ImageColonyCoreShadow:   {Path: "image/colony_core_shadow.png"},

		ImageWorkerAgent:    {Path: "image/drones/worker_agent.png", FrameWidth: 9},
		ImageMilitiaAgent:   {Path: "image/drones/militia_agent.png", FrameWidth: 11},
		ImageMortarAgent:    {Path: "image/drones/mortar_agent.png", FrameWidth: 21},
		ImageCripplerAgent:  {Path: "image/drones/crippler_agent.png", FrameWidth: 15},
		ImageFlamerAgent:    {Path: "image/drones/flamer_agent.png", FrameWidth: 21},
		ImageRepairAgent:    {Path: "image/drones/repair_agent.png", FrameWidth: 17},
		ImageAntiAirAgent:   {Path: "image/drones/antiair_agent.png", FrameWidth: 17},
		ImageServoAgent:     {Path: "image/drones/servo_agent.png", FrameWidth: 15},
		ImageRechargerAgent: {Path: "image/drones/recharger_agent.png", FrameWidth: 17},
		ImageRefresherAgent: {Path: "image/drones/refresher_agent.png", FrameWidth: 23},
		ImageFighterAgent:   {Path: "image/drones/fighter_agent.png", FrameWidth: 15},
		ImagePrismAgent:     {Path: "image/drones/prism_agent.png", FrameWidth: 15},
		ImageDestroyerAgent: {Path: "image/drones/destroyer_agent.png", FrameWidth: 33},
		ImageRepellerAgent:  {Path: "image/drones/repeller_agent.png", FrameWidth: 15},
		ImageFreighterAgent: {Path: "image/drones/freighter_agent.png", FrameWidth: 17},
		ImageRedminerAgent:  {Path: "image/drones/redminer_agent.png", FrameWidth: 13},
		ImageGeneratorAgent: {Path: "image/drones/generator_agent.png", FrameWidth: 15},

		ImageColonyDamageMask: {Path: "image/colony_damage_mask.png"},

		ImageEssenceSourceDissolveMask:    {Path: "image/resources/essence_source_dissolve_mask.png"},
		ImageEssenceCrystalSource:         {Path: "image/resources/crystal_source.png", FrameWidth: 16},
		ImageEssenceGoldSource:            {Path: "image/resources/gold_source.png", FrameWidth: 28},
		ImageEssenceIronSource:            {Path: "image/resources/iron_source.png", FrameWidth: 32},
		ImageEssenceScrapSource:           {Path: "image/resources/scrap_source.png"},
		ImageEssenceSmallScrapSource:      {Path: "image/resources/small_scrap_source.png"},
		ImageEssenceScrapCreepSource:      {Path: "image/resources/scrap_source_creep.png"},
		ImageEssenceSmallScrapCreepSource: {Path: "image/resources/small_scrap_source_creep.png"},
		ImageEssenceBigScrapCreepSource:   {Path: "image/resources/big_scrap_source_creep.png"},
		ImageEssenceSource:                {Path: "image/resources/essence_source.png", FrameWidth: 32},
		ImageRedEssenceSource:             {Path: "image/resources/red_essence_source.png", FrameWidth: 32},

		ImageEliteCrawlerCreep: {Path: "image/creeps/elite_crawler_creep.png", FrameWidth: 23},
		ImageCrawlerCreep:      {Path: "image/creeps/crawler_creep.png", FrameWidth: 23},
		ImagePrimitiveCreep:    {Path: "image/creeps/tier1_creep.png"},
		ImageCreepTier2:        {Path: "image/creeps/tier2_creep.png"},
		ImageCreepTier3:        {Path: "image/creeps/tier3_creep.png", FrameWidth: 25},
		ImageTurretCreep:       {Path: "image/creeps/turret_creep.png"},
		ImageTankCreep:         {Path: "image/creeps/tank_creep.png"},
		ImageUberBoss:          {Path: "image/creeps/uber_boss.png", FrameWidth: 40},
		ImageUberBossOpen:      {Path: "image/creeps/uber_boss_open.png"},
		ImageCreepBase:         {Path: "image/creeps/creep_base.png", FrameWidth: 32},

		ImageBackgroundTiles: {Path: "image/landscape/tiles.png"},
		ImageMountains:       {Path: "image/landscape/mountains.png", FrameWidth: 32},
		ImageLandCrack:       {Path: "image/landscape/landcrack.png", FrameWidth: 32},
		ImageLandCrack2:      {Path: "image/landscape/landcrack2.png", FrameWidth: 32},
		ImageLandCrack3:      {Path: "image/landscape/landcrack3.png", FrameWidth: 32},
		ImageLandCrack4:      {Path: "image/landscape/landcrack4.png", FrameWidth: 32},

		ImageEliteCrawlerProjectile: {Path: "image/projectile/elite_crawler_projectile.png"},
		ImageTankProjectile:         {Path: "image/projectile/tank_projectile.png"},
		ImageAssaultProjectile:      {Path: "image/projectile/assault_projectile.png"},
		ImageCripplerProjectile:     {Path: "image/projectile/crippler_projectile.png"},
		ImageMilitiaProjectile:      {Path: "image/projectile/militia_projectile.png"},
		ImageRepellerProjectile:     {Path: "image/projectile/repeller_projectile.png"},
		ImageFighterProjectile:      {Path: "image/projectile/fighter_projectile.png"},
		ImageWandererProjectile:     {Path: "image/projectile/wanderer_projectile.png"},
		ImageFlamerProjectile:       {Path: "image/projectile/flamer_projectile.png"},
		ImageMortarProjectile:       {Path: "image/projectile/mortar_projectile.png"},
		ImageAntiAirMissile:         {Path: "image/projectile/aa_missile.png"},
		ImageMissile:                {Path: "image/projectile/missile.png"},

		ImageUIButtonIdle:             {Path: "image/ebitenui/button-idle.png"},
		ImageUIButtonHover:            {Path: "image/ebitenui/button-hover.png"},
		ImageUIButtonPressed:          {Path: "image/ebitenui/button-pressed.png"},
		ImageUIButtonDisabled:         {Path: "image/ebitenui/button-disabled.png"},
		ImageUIButtonSelectedIdle:     {Path: "image/ebitenui/button-selected-idle.png"},
		ImageUIButtonSelectedHover:    {Path: "image/ebitenui/button-selected-hover.png"},
		ImageUIButtonSelectedPressed:  {Path: "image/ebitenui/button-selected-pressed.png"},
		ImageUIButtonSelectedDisabled: {Path: "image/ebitenui/button-selected-disabled.png"},
		ImageUIArrowDownIdle:          {Path: "image/ebitenui/arrow-down-idle.png"},
		ImageUIArrowDownDisabled:      {Path: "image/ebitenui/arrow-down-disabled.png"},
		ImageUIOptionButtonIdle:       {Path: "image/ebitenui/combo-button-idle.png"},
		ImageUIOptionButtonHover:      {Path: "image/ebitenui/combo-button-hover.png"},
		ImageUIOptionButtonPressed:    {Path: "image/ebitenui/combo-button-pressed.png"},
		ImageUIOptionButtonDisabled:   {Path: "image/ebitenui/combo-button-disabled.png"},
		ImageUIListIdle:               {Path: "image/ebitenui/list-idle.png"},
		ImageUIListDisabled:           {Path: "image/ebitenui/list-disabled.png"},
		ImageUIListMask:               {Path: "image/ebitenui/list-mask.png"},
		ImageUIListTrackIdle:          {Path: "image/ebitenui/list-track-idle.png"},
		ImageUIListTrackDisabled:      {Path: "image/ebitenui/list-track-disabled.png"},
		ImageUISliderHandleIdle:       {Path: "image/ebitenui/slider-handle-idle.png"},
		ImageUISliderHandleHover:      {Path: "image/ebitenui/slider-handle-hover.png"},
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
	ImageUberBossOpen
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
	ImageAntiAirAgent
	ImagePrismAgent
	ImageRepairAgent
	ImageRechargerAgent
	ImageRefresherAgent
	ImageFighterAgent
	ImageDestroyerAgent
	ImageCripplerAgent
	ImageMortarAgent
	ImageRedminerAgent
	ImageRepellerAgent
	ImageServoAgent
	ImageFreighterAgent
	ImageEssenceCrystalSource
	ImageEssenceGoldSource
	ImageEssenceIronSource
	ImageEssenceScrapSource
	ImageEssenceSmallScrapSource
	ImageEssenceScrapCreepSource
	ImageEssenceSmallScrapCreepSource
	ImageEssenceBigScrapCreepSource
	ImageEssenceSource
	ImageRedEssenceSource
	ImageCrawlerCreep
	ImageEliteCrawlerCreep
	ImagePrimitiveCreep
	ImageCreepTier2
	ImageCreepTier3
	ImageTurretCreep
	ImageTankCreep

	ImageBackgroundTiles
	ImageMountains
	ImageLandCrack
	ImageLandCrack2
	ImageLandCrack3
	ImageLandCrack4

	ImageEliteCrawlerProjectile
	ImageTankProjectile
	ImageAssaultProjectile
	ImageCripplerProjectile
	ImageMilitiaProjectile
	ImageRepellerProjectile
	ImageFighterProjectile
	ImageWandererProjectile
	ImageFlamerProjectile
	ImageMortarProjectile
	ImageAntiAirMissile
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
