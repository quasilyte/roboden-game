package assets

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"

	_ "image/png"
)

func registerAudioResource(ctx *ge.Context) {
	audioResources := map[resource.AudioID]resource.AudioInfo{
		AudioError:            {Path: "audio/error.wav", Volume: -0.25},
		AudioClick:            {Path: "audio/button_click.wav", Volume: -0.3},
		AudioBaseSelect:       {Path: "audio/base_select.wav", Volume: -0.4},
		AudioChoiceMade:       {Path: "audio/choice_made.wav", Volume: -0.45},
		AudioChoiceReady:      {Path: "audio/choice_ready.wav", Volume: -0.55},
		AudioColonyLanded:     {Path: "audio/colony_landed.wav", Volume: -0.2},
		AudioEssenceCollected: {Path: "audio/essence_collected.wav", Volume: -0.55},
		AudioAgentCloned:      {Path: "audio/agent_cloned.wav", Volume: -0.3},
		AudioAgentProduced:    {Path: "audio/agent_produced.wav", Volume: -0.25},
		AudioAgentRecycled:    {Path: "audio/agent_recycled.wav", Volume: -0.3},
		AudioAgentDestroyed:   {Path: "audio/agent_destroyed.wav", Volume: -0.25},
		AudioFighterBeam:      {Path: "audio/fighter_beam.wav", Volume: -0.35},
		AudioWandererBeam:     {Path: "audio/wanderer_beam.wav", Volume: -0.3},
		AudioMilitiaShot:      {Path: "audio/militia_shot.wav", Volume: -0.3},
		AudioCripplerShot:     {Path: "audio/crippler_shot.wav", Volume: -0.3},
		AudioMortarShot:       {Path: "audio/mortar_shot.wav", Volume: -0.3},
		AudioAssaultShot:      {Path: "audio/assault_shot.wav", Volume: -0.5},
		AudioFlamerShot:       {Path: "audio/flamer_shot.wav", Volume: -0.35},
		AudioStunBeam:         {Path: "audio/stun_laser.wav", Volume: -0.3},
		AudioRechargerBeam:    {Path: "audio/recharger_beam.wav", Volume: -0.4},
		AudioRepairBeam:       {Path: "audio/repair_beam.wav", Volume: -0.3},
		AudioRepellerBeam:     {Path: "audio/repeller_beam.wav", Volume: -0.3},
		AudioDestroyerBeam:    {Path: "audio/destroyer_beam.wav", Volume: -0.3},
		AudioRailgun:          {Path: "audio/railgun.wav", Volume: -0.3},
		AudioAntiAirMissiles:  {Path: "audio/aa_missiles.wav", Volume: -0.35},
		AudioMissile:          {Path: "audio/missile.wav", Volume: -0.3},
		AudioTankShot:         {Path: "audio/tank_shot.wav", Volume: -0.3},
		AudioEliteCrawlerShot: {Path: "audio/elite_crawler_shot.wav", Volume: -0.3},
		AudioCloning1:         {Path: "audio/cloning1.wav", Volume: -0.3},
		AudioCloning2:         {Path: "audio/cloning2.wav", Volume: -0.3},
		AudioMerging1:         {Path: "audio/merging1.wav", Volume: -0.4},
		AudioMerging2:         {Path: "audio/merging2.wav", Volume: -0.4},
		AudioExplosion1:       {Path: "audio/explosion1.wav", Volume: -0.4},
		AudioExplosion2:       {Path: "audio/explosion2.wav", Volume: -0.4},
		AudioExplosion3:       {Path: "audio/explosion3.wav", Volume: -0.4},
		AudioExplosion4:       {Path: "audio/explosion4.wav", Volume: -0.4},
		AudioExplosion5:       {Path: "audio/explosion5.wav", Volume: -0.4},

		AudioMusicTrack1: {Path: "audio/music/deadly_windmills.ogg", Volume: -0.3, Group: SoundGroupMusic},
		AudioMusicTrack2: {Path: "audio/music/war_path.ogg", Volume: -0.3, Group: SoundGroupMusic},
		AudioMusicTrack3: {Path: "audio/music/crush.ogg", Volume: -0.3, Group: SoundGroupMusic},
	}

	for id, res := range audioResources {
		ctx.Loader.AudioRegistry.Set(id, res)
		ctx.Loader.LoadAudio(id)
	}
}

const (
	AudioNone resource.AudioID = iota

	AudioError
	AudioClick
	AudioBaseSelect
	AudioChoiceMade
	AudioChoiceReady
	AudioColonyLanded
	AudioEssenceCollected
	AudioAgentCloned
	AudioAgentProduced
	AudioAgentRecycled
	AudioAgentDestroyed
	AudioWandererBeam
	AudioStunBeam
	AudioRechargerBeam
	AudioRepairBeam
	AudioMortarShot
	AudioCripplerShot
	AudioMilitiaShot
	AudioFlamerShot
	AudioAssaultShot
	AudioFighterBeam
	AudioTankShot
	AudioEliteCrawlerShot
	AudioRepellerBeam
	AudioDestroyerBeam
	AudioAntiAirMissiles
	AudioMissile
	AudioRailgun
	AudioCloning1
	AudioCloning2
	AudioMerging1
	AudioMerging2
	AudioExplosion1
	AudioExplosion2
	AudioExplosion3
	AudioExplosion4
	AudioExplosion5

	AudioMusicTrack1
	AudioMusicTrack2
	AudioMusicTrack3
)
