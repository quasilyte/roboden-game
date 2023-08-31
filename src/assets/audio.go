package assets

import (
	"runtime"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func RegisterMusicResource(ctx *ge.Context, config *Config, progress *float64) {
	audioResources := map[resource.AudioID]resource.AudioInfo{
		AudioMusicTrack1: {Path: "$music/deadly_windmills.ogg", Volume: -0.3, Group: SoundGroupMusic},
		AudioMusicTrack2: {Path: "$music/war_path.ogg", Volume: -0.3, Group: SoundGroupMusic},
		AudioMusicTrack3: {Path: "$music/crush.ogg", Volume: -0.3, Group: SoundGroupMusic},
	}

	if config.ExtraMusic {
		audioResources[AudioMusicTrack4] = resource.AudioInfo{
			Path:   "$music/track4.ogg",
			Volume: -0.3,
			Group:  SoundGroupMusic,
		}
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

func RegisterAudioResource(ctx *ge.Context, config *Config, progress *float64) {
	audioResources := map[resource.AudioID]resource.AudioInfo{
		AudioVictory:                    {Path: "$sfx/victory.wav", Volume: -0.05},
		AudioWaveStart:                  {Path: "$sfx/wave_start.wav", Volume: 0},
		AudioError:                      {Path: "$sfx/error.wav", Volume: -0.25},
		AudioClick:                      {Path: "$sfx/button_click.wav", Volume: -0.3},
		AudioPing:                       {Path: "$sfx/ping.wav", Volume: 0},
		AudioBaseSelect:                 {Path: "$sfx/base_select.wav", Volume: -0.55},
		AudioChoiceMade:                 {Path: "$sfx/choice_made.wav", Volume: -0.45},
		AudioChoiceReady:                {Path: "$sfx/choice_ready.wav", Volume: -0.55},
		AudioColonyLanded:               {Path: "$sfx/colony_landed.wav", Volume: -0.2},
		AudioHarvesterEffect:            {Path: "$sfx/harvester.wav", Volume: -0.2},
		AudioEssenceCollected:           {Path: "$sfx/essence_collected.wav", Volume: -0.55},
		AudioCourierResourceBeam:        {Path: "$sfx/courier_resource_beam.wav", Volume: -0.3},
		AudioAgentProduced:              {Path: "$sfx/agent_produced.wav", Volume: -0.4},
		AudioAgentRecycled:              {Path: "$sfx/agent_recycled.wav", Volume: -0.3},
		AudioAgentConsumed:              {Path: "$sfx/drone_consumed.wav", Volume: -0.25},
		AudioAgentDestroyed:             {Path: "$sfx/agent_destroyed.wav", Volume: -0.25},
		AudioSiegeRocket1:               {Path: "$sfx/siege_rocket1.wav", Volume: +0.1},
		AudioSiegeRocket2:               {Path: "$sfx/siege_rocket2.wav", Volume: +0.1},
		AudioFighterBeam:                {Path: "$sfx/fighter_beam.wav", Volume: -0.35},
		AudioTankColonyBlasterShot:      {Path: "$sfx/tank_colony_blaster.wav", Volume: -0.15},
		AudioRelictAgentShot:            {Path: "$sfx/relict_agent_shot.wav", Volume: -0.35},
		AudioDefenderShot:               {Path: "$sfx/defender_shot.wav", Volume: -0.5},
		AudioFirebugShot:                {Path: "$sfx/flamethrower.wav", Volume: -0.2},
		AudioGunpointShot:               {Path: "$sfx/gunpoint_shot.wav", Volume: -0.3},
		AudioBeamTowerShot:              {Path: "$sfx/beamtower_shot.wav", Volume: -0.4},
		AudioRepulseTowerAttack:         {Path: "$sfx/artifact_tower_attack.wav", Volume: -0.4},
		AudioTetherShot:                 {Path: "$sfx/tether_shot.wav", Volume: -0.05},
		AudioWandererBeam:               {Path: "$sfx/wanderer_beam.wav", Volume: -0.3},
		AudioScoutShot:                  {Path: "$sfx/scout_shot.wav", Volume: -0.3},
		AudioCripplerShot:               {Path: "$sfx/crippler_shot.wav", Volume: -0.3},
		AudioKamizakeAttack:             {Path: "$sfx/kamikaze_attack.wav", Volume: 0.2},
		AudioScavengerShot:              {Path: "$sfx/scavenger_shot.wav", Volume: -0.3},
		AudioCourierShot:                {Path: "$sfx/courier_shot.wav", Volume: 0},
		AudioDisintegratorShot:          {Path: "$sfx/disintegrator_shot.wav", Volume: -0.2},
		AudioTargeterShot:               {Path: "$sfx/targeter_shot.wav", Volume: -0.8},
		AudioMortarShot:                 {Path: "$sfx/mortar_shot.wav", Volume: -0.3},
		AudioCommanderShot:              {Path: "$sfx/commander_shot.wav", Volume: +0.2},
		AudioRoombaShot:                 {Path: "$sfx/roomba_shot.wav", Volume: 0.2},
		AudioAssaultShot:                {Path: "$sfx/assault_shot.wav", Volume: -0.5},
		AudioDominatorShot:              {Path: "$sfx/dominator_shot.wav", Volume: -0.25},
		AudioHowitzerShot:               {Path: "$sfx/howitzer_shot.wav", Volume: 0},
		AudioHowitzerLaserShot:          {Path: "$sfx/howitzer_laser_shot.wav", Volume: -0.25},
		AudioStormbringerShot:           {Path: "$sfx/stormbringer_shot.wav", Volume: -0.15},
		AudioStunBeam:                   {Path: "$sfx/stun_laser.wav", Volume: -0.3},
		AudioServantShot:                {Path: "$sfx/servant_shot.wav", Volume: -0.35},
		AudioServantWave:                {Path: "$sfx/servant_wave.wav", Volume: 0},
		AudioRechargerBeam:              {Path: "$sfx/recharger_beam.wav", Volume: -0.45},
		AudioRepairBeam:                 {Path: "$sfx/repair_beam.wav", Volume: -0.25},
		AudioRepellerBeam:               {Path: "$sfx/repeller_beam.wav", Volume: -0.4},
		AudioDestroyerBeam:              {Path: "$sfx/destroyer_beam.wav", Volume: -0.3},
		AudioStealth:                    {Path: "$sfx/stealth.wav", Volume: -0.25},
		AudioMarauderShot:               {Path: "$sfx/marauder_shot.wav", Volume: -0.5},
		AudioPrismShot:                  {Path: "$sfx/prism_shot.wav", Volume: -0.45},
		AudioSkirmisherShot:             {Path: "$sfx/skirmisher_shot.wav", Volume: -0.3},
		AudioScarabShot:                 {Path: "$sfx/scarab_shot.wav", Volume: -0.4},
		AudioRailgun:                    {Path: "$sfx/railgun.wav", Volume: -0.3},
		AudioAntiAirMissiles:            {Path: "$sfx/aa_missiles.wav", Volume: -0.5},
		AudioMissile:                    {Path: "$sfx/missile.wav", Volume: -0.3},
		AudioTankShot:                   {Path: "$sfx/tank_shot.wav", Volume: -0.3},
		AudioIonMortarShot:              {Path: "$sfx/ion_mortar_shot.wav", Volume: -0.3},
		AudioHeavyCrawlerShot:           {Path: "$sfx/heavy_crawler_shot.wav", Volume: -0.25},
		AudioEliteCrawlerShot:           {Path: "$sfx/elite_crawler_shot.wav", Volume: -0.35},
		AudioStealthCrawlerShot:         {Path: "$sfx/stealth_crawler_shot.wav", Volume: -0.3},
		AudioFortressAttack:             {Path: "$sfx/fortress_attack.wav", Volume: -0.3},
		AudioTemplarAttack:              {Path: "$sfx/templar_attack.wav", Volume: -0.4},
		AudioMagmaShot1:                 {Path: "$sfx/magma_shot1.wav", Volume: -0.45},
		AudioMagmaShot2:                 {Path: "$sfx/magma_shot2.wav", Volume: -0.45},
		AudioMagmaShot3:                 {Path: "$sfx/magma_shot3.wav", Volume: -0.45},
		AudioCloning1:                   {Path: "$sfx/cloning1.wav", Volume: -0.3},
		AudioCloning2:                   {Path: "$sfx/cloning2.wav", Volume: -0.3},
		AudioMerging1:                   {Path: "$sfx/merging1.wav", Volume: -0.65},
		AudioMerging2:                   {Path: "$sfx/merging2.wav", Volume: -0.65},
		AudioIonZap1:                    {Path: "$sfx/ion_zap1.wav", Volume: -0.4},
		AudioIonZap2:                    {Path: "$sfx/ion_zap2.wav", Volume: -0.4},
		AudioPurpleExplosion1:           {Path: "$sfx/purple_explosion1.wav", Volume: -0.4},
		AudioPurpleExplosion2:           {Path: "$sfx/purple_explosion2.wav", Volume: -0.4},
		AudioPurpleExplosion3:           {Path: "$sfx/purple_explosion3.wav", Volume: -0.4},
		AudioAbombLaunch:                {Path: "$sfx/abomb_launch.wav", Volume: -0.2},
		AudioAbombExplosion:             {Path: "$sfx/abomb_explosion.wav", Volume: -0.1},
		AudioExplosion1:                 {Path: "$sfx/explosion1.wav", Volume: -0.45},
		AudioExplosion2:                 {Path: "$sfx/explosion2.wav", Volume: -0.45},
		AudioExplosion3:                 {Path: "$sfx/explosion3.wav", Volume: -0.45},
		AudioExplosion4:                 {Path: "$sfx/explosion4.wav", Volume: -0.45},
		AudioExplosion5:                 {Path: "$sfx/explosion5.wav", Volume: -0.45},
		AudioIonBlast1:                  {Path: "$sfx/ion_blast1.wav", Volume: -0.45},
		AudioIonBlast2:                  {Path: "$sfx/ion_blast2.wav", Volume: -0.45},
		AudioIonBlast3:                  {Path: "$sfx/ion_blast3.wav", Volume: -0.45},
		AudioIonBlast4:                  {Path: "$sfx/ion_blast4.wav", Volume: -0.45},
		AudioLavaBurst1:                 {Path: "$sfx/lava_geyser1.wav", Volume: -0.45},
		AudioLavaBurst2:                 {Path: "$sfx/lava_geyser2.wav", Volume: -0.45},
		AudioLavaBurst3:                 {Path: "$sfx/lava_geyser3.wav", Volume: -0.45},
		AudioLavaBurst4:                 {Path: "$sfx/lava_geyser4.wav", Volume: -0.45},
		AudioMagmaExplosion1:            {Path: "$sfx/magma_explosion1.wav", Volume: -0.5},
		AudioMagmaExplosion2:            {Path: "$sfx/magma_explosion2.wav", Volume: -0.6},
		AudioTankColonyBlasterExplosion: {Path: "$sfx/tank_colony_blaster_explosion.wav", Volume: -0.1},
		AudioTeleportCharge:             {Path: "$sfx/teleport_charge.wav", Volume: -0.2},
		AudioTeleportDone:               {Path: "$sfx/teleport_done.wav", Volume: -0.2},
		AudioOrganicRestored:            {Path: "$sfx/organic_restored.wav", Volume: -0.2},
		AudioWispShocker:                {Path: "$sfx/wisp_shocker.wav", Volume: -0.2},
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

func NumAudioSamples(id resource.AudioID) int {
	switch id {
	case AudioPurpleExplosion1:
		return 3
	case AudioExplosion1:
		return 5
	case AudioIonBlast1:
		return 4
	case AudioLavaBurst1:
		return 4
	case AudioMagmaExplosion1:
		return 2
	case AudioSiegeRocket1:
		return 2
	case AudioMagmaShot1:
		return 3
	default:
		return 1
	}
}

const (
	AudioNone resource.AudioID = iota

	AudioVictory
	AudioWaveStart
	AudioError
	AudioClick
	AudioPing
	AudioBaseSelect
	AudioChoiceMade
	AudioChoiceReady
	AudioColonyLanded
	AudioHarvesterEffect
	AudioEssenceCollected
	AudioCourierResourceBeam
	AudioAgentProduced
	AudioAgentRecycled
	AudioAgentConsumed
	AudioAgentDestroyed
	AudioWandererBeam
	AudioStunBeam
	AudioSiegeRocket1
	AudioSiegeRocket2
	AudioServantShot
	AudioServantWave
	AudioRechargerBeam
	AudioRepairBeam
	AudioRoombaShot
	AudioMortarShot
	AudioCommanderShot
	AudioCripplerShot
	AudioKamizakeAttack
	AudioScavengerShot
	AudioCourierShot
	AudioDisintegratorShot
	AudioTargeterShot
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
	AudioRepulseTowerAttack
	AudioTetherShot
	AudioDefenderShot
	AudioFirebugShot
	AudioFighterBeam
	AudioTankColonyBlasterShot
	AudioRelictAgentShot
	AudioTankShot
	AudioHeavyCrawlerShot
	AudioEliteCrawlerShot
	AudioStealthCrawlerShot
	AudioFortressAttack
	AudioTemplarAttack
	AudioRepellerBeam
	AudioDestroyerBeam
	AudioPrismShot
	AudioSkirmisherShot
	AudioIonMortarShot
	AudioScarabShot
	AudioAntiAirMissiles
	AudioMissile
	AudioRailgun
	AudioMagmaShot1
	AudioMagmaShot2
	AudioMagmaShot3
	AudioCloning1
	AudioCloning2
	AudioMerging1
	AudioMerging2
	AudioIonZap1
	AudioIonZap2
	AudioMagmaExplosion1
	AudioMagmaExplosion2
	AudioTankColonyBlasterExplosion
	AudioPurpleExplosion1
	AudioPurpleExplosion2
	AudioPurpleExplosion3
	AudioAbombLaunch
	AudioAbombExplosion
	AudioExplosion1
	AudioExplosion2
	AudioExplosion3
	AudioExplosion4
	AudioExplosion5
	AudioIonBlast1
	AudioIonBlast2
	AudioIonBlast3
	AudioIonBlast4
	AudioLavaBurst1
	AudioLavaBurst2
	AudioLavaBurst3
	AudioLavaBurst4
	AudioTeleportCharge
	AudioTeleportDone
	AudioOrganicRestored
	AudioWispShocker

	AudioMusicTrack1
	AudioMusicTrack2
	AudioMusicTrack3
	AudioMusicTrack4
)
