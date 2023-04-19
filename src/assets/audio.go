package assets

import (
	"runtime"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func RegisterMusicResource(ctx *ge.Context, progress *float64) {
	audioResources := map[resource.AudioID]resource.AudioInfo{
		AudioMusicTrack1: {Path: "$music/deadly_windmills.ogg", Volume: -0.3, Group: SoundGroupMusic},
		AudioMusicTrack2: {Path: "$music/war_path.ogg", Volume: -0.3, Group: SoundGroupMusic},
		AudioMusicTrack3: {Path: "$music/crush.ogg", Volume: -0.3, Group: SoundGroupMusic},
	}

	singleThread := runtime.GOMAXPROCS(-1) == 1
	progressPerItem := 1.0 / float64(len(audioResources))
	for id, res := range audioResources {
		ctx.Loader.AudioRegistry.Set(id, res)
		ctx.Loader.LoadAudio(id)
		if progress != nil {
			*progress += progressPerItem
		}
		if singleThread {
			runtime.Gosched()
		}
	}
}

func RegisterAudioResource(ctx *ge.Context, progress *float64) {
	audioResources := map[resource.AudioID]resource.AudioInfo{
		AudioVictory:             {Path: "$sfx/victory.wav", Volume: -0.05},
		AudioWaveStart:           {Path: "$sfx/wave_start.wav", Volume: 0},
		AudioError:               {Path: "$sfx/error.wav", Volume: -0.25},
		AudioClick:               {Path: "$sfx/button_click.wav", Volume: -0.3},
		AudioBaseSelect:          {Path: "$sfx/base_select.wav", Volume: -0.4},
		AudioChoiceMade:          {Path: "$sfx/choice_made.wav", Volume: -0.45},
		AudioChoiceReady:         {Path: "$sfx/choice_ready.wav", Volume: -0.55},
		AudioColonyLanded:        {Path: "$sfx/colony_landed.wav", Volume: -0.2},
		AudioEssenceCollected:    {Path: "$sfx/essence_collected.wav", Volume: -0.55},
		AudioCourierResourceBeam: {Path: "$sfx/courier_resource_beam.wav", Volume: -0.3},
		AudioAgentProduced:       {Path: "$sfx/agent_produced.wav", Volume: -0.25},
		AudioAgentRecycled:       {Path: "$sfx/agent_recycled.wav", Volume: -0.3},
		AudioAgentDestroyed:      {Path: "$sfx/agent_destroyed.wav", Volume: -0.25},
		AudioFighterBeam:         {Path: "$sfx/fighter_beam.wav", Volume: -0.35},
		AudioDefenderShot:        {Path: "$sfx/defender_shot.wav", Volume: -0.45},
		AudioGunpointShot:        {Path: "$sfx/gunpoint_shot.wav", Volume: -0.3},
		AudioBeamTowerShot:       {Path: "$sfx/beamtower_shot.wav", Volume: -0.4},
		AudioTetherShot:          {Path: "$sfx/tether_shot.wav", Volume: 0},
		AudioWandererBeam:        {Path: "$sfx/wanderer_beam.wav", Volume: -0.3},
		AudioScoutShot:           {Path: "$sfx/scout_shot.wav", Volume: -0.3},
		AudioCripplerShot:        {Path: "$sfx/crippler_shot.wav", Volume: -0.3},
		AudioKamizakeAttack:      {Path: "$sfx/kamikaze_attack.wav", Volume: 0.2},
		AudioScavengerShot:       {Path: "$sfx/scavenger_shot.wav", Volume: -0.3},
		AudioCourierShot:         {Path: "$sfx/courier_shot.wav", Volume: 0},
		AudioDisintegratorShot:   {Path: "$sfx/disintegrator_shot.wav", Volume: -0.2},
		AudioMortarShot:          {Path: "$sfx/mortar_shot.wav", Volume: -0.3},
		AudioAssaultShot:         {Path: "$sfx/assault_shot.wav", Volume: -0.5},
		AudioDominatorShot:       {Path: "$sfx/dominator_shot.wav", Volume: -0.25},
		AudioHowitzerShot:        {Path: "$sfx/howitzer_shot.wav", Volume: 0},
		AudioHowitzerLaserShot:   {Path: "$sfx/howitzer_laser_shot.wav", Volume: -0.25},
		AudioStormbringerShot:    {Path: "$sfx/stormbringer_shot.wav", Volume: -0.15},
		AudioStunBeam:            {Path: "$sfx/stun_laser.wav", Volume: -0.3},
		AudioServantShot:         {Path: "$sfx/servant_shot.wav", Volume: -0.35},
		AudioServantWave:         {Path: "$sfx/servant_wave.wav", Volume: -0.1},
		AudioRechargerBeam:       {Path: "$sfx/recharger_beam.wav", Volume: -0.4},
		AudioRepairBeam:          {Path: "$sfx/repair_beam.wav", Volume: -0.25},
		AudioRepellerBeam:        {Path: "$sfx/repeller_beam.wav", Volume: -0.35},
		AudioDestroyerBeam:       {Path: "$sfx/destroyer_beam.wav", Volume: -0.3},
		AudioStealth:             {Path: "$sfx/stealth.wav", Volume: -0.25},
		AudioMarauderShot:        {Path: "$sfx/marauder_shot.wav", Volume: -0.4},
		AudioPrismShot:           {Path: "$sfx/prism_shot.wav", Volume: -0.4},
		AudioSkirmisherShot:      {Path: "$sfx/skirmisher_shot.wav", Volume: -0.3},
		AudioRailgun:             {Path: "$sfx/railgun.wav", Volume: -0.3},
		AudioAntiAirMissiles:     {Path: "$sfx/aa_missiles.wav", Volume: -0.4},
		AudioMissile:             {Path: "$sfx/missile.wav", Volume: -0.3},
		AudioTankShot:            {Path: "$sfx/tank_shot.wav", Volume: -0.3},
		AudioHeavyCrawlerShot:    {Path: "$sfx/heavy_crawler_shot.wav", Volume: -0.25},
		AudioEliteCrawlerShot:    {Path: "$sfx/elite_crawler_shot.wav", Volume: -0.3},
		AudioStealthCrawlerShot:  {Path: "$sfx/stealth_crawler_shot.wav", Volume: -0.3},
		AudioCloning1:            {Path: "$sfx/cloning1.wav", Volume: -0.3},
		AudioCloning2:            {Path: "$sfx/cloning2.wav", Volume: -0.3},
		AudioMerging1:            {Path: "$sfx/merging1.wav", Volume: -0.45},
		AudioMerging2:            {Path: "$sfx/merging2.wav", Volume: -0.45},
		AudioIonZap1:             {Path: "$sfx/ion_zap1.wav", Volume: -0.4},
		AudioIonZap2:             {Path: "$sfx/ion_zap2.wav", Volume: -0.4},
		AudioPurpleExplosion1:    {Path: "$sfx/purple_explosion1.wav", Volume: -0.4},
		AudioPurpleExplosion2:    {Path: "$sfx/purple_explosion2.wav", Volume: -0.4},
		AudioPurpleExplosion3:    {Path: "$sfx/purple_explosion3.wav", Volume: -0.4},
		AudioExplosion1:          {Path: "$sfx/explosion1.wav", Volume: -0.4},
		AudioExplosion2:          {Path: "$sfx/explosion2.wav", Volume: -0.4},
		AudioExplosion3:          {Path: "$sfx/explosion3.wav", Volume: -0.4},
		AudioExplosion4:          {Path: "$sfx/explosion4.wav", Volume: -0.4},
		AudioExplosion5:          {Path: "$sfx/explosion5.wav", Volume: -0.4},
	}

	singleThread := runtime.GOMAXPROCS(-1) == 1
	progressPerItem := 1.0 / float64(len(audioResources))
	for id, res := range audioResources {
		ctx.Loader.AudioRegistry.Set(id, res)
		ctx.Loader.LoadAudio(id)
		if progress != nil {
			*progress += progressPerItem
		}
		if singleThread {
			runtime.Gosched()
		}
	}
}

const (
	AudioNone resource.AudioID = iota

	AudioVictory
	AudioWaveStart
	AudioError
	AudioClick
	AudioBaseSelect
	AudioChoiceMade
	AudioChoiceReady
	AudioColonyLanded
	AudioEssenceCollected
	AudioCourierResourceBeam
	AudioAgentProduced
	AudioAgentRecycled
	AudioAgentDestroyed
	AudioWandererBeam
	AudioStunBeam
	AudioServantShot
	AudioServantWave
	AudioRechargerBeam
	AudioRepairBeam
	AudioMortarShot
	AudioCripplerShot
	AudioKamizakeAttack
	AudioScavengerShot
	AudioCourierShot
	AudioDisintegratorShot
	AudioStealth
	AudioMarauderShot
	AudioScoutShot
	AudioStormbringerShot
	AudioAssaultShot
	AudioDominatorShot
	AudioHowitzerShot
	AudioHowitzerLaserShot
	AudioGunpointShot
	AudioBeamTowerShot
	AudioTetherShot
	AudioDefenderShot
	AudioFighterBeam
	AudioTankShot
	AudioHeavyCrawlerShot
	AudioEliteCrawlerShot
	AudioStealthCrawlerShot
	AudioRepellerBeam
	AudioDestroyerBeam
	AudioPrismShot
	AudioSkirmisherShot
	AudioAntiAirMissiles
	AudioMissile
	AudioRailgun
	AudioCloning1
	AudioCloning2
	AudioMerging1
	AudioMerging2
	AudioIonZap1
	AudioIonZap2
	AudioPurpleExplosion1
	AudioPurpleExplosion2
	AudioPurpleExplosion3
	AudioExplosion1
	AudioExplosion2
	AudioExplosion3
	AudioExplosion4
	AudioExplosion5

	AudioMusicTrack1
	AudioMusicTrack2
	AudioMusicTrack3
)
