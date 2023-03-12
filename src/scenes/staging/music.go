package staging

import (
	"time"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type musicTrack struct {
	audio  resource.AudioID
	length time.Duration
}

var musicTrackList = []musicTrack{
	{audio: assets.AudioMusicTrack1, length: 5*time.Minute + 12*time.Second},
	{audio: assets.AudioMusicTrack2, length: 2*time.Minute + 44*time.Second},
	{audio: assets.AudioMusicTrack3, length: 2*time.Minute + 3*time.Second},
}

type musicPlayer struct {
	scene *ge.Scene

	musicTrackSlider gmath.Slider

	nextTrackTime time.Time

	timeCheckDelay float64
}

func newMusicPlayer(scene *ge.Scene) *musicPlayer {
	p := &musicPlayer{scene: scene}
	p.musicTrackSlider.SetBounds(0, len(musicTrackList)-1)
	return p
}

func (p *musicPlayer) Start() {
	p.playNextTrack(time.Now())
}

func (p *musicPlayer) playNextTrack(now time.Time) {
	track := musicTrackList[p.musicTrackSlider.Value()]
	p.musicTrackSlider.Inc()

	p.scene.Audio().PauseCurrentMusic()
	p.scene.Audio().PlayMusic(track.audio)
	p.nextTrackTime = now.Add(track.length)
}

func (p *musicPlayer) Update(delta float64) {
	// Don't do time.Now() every logical frame.
	// Check it once in a short while.
	p.timeCheckDelay += delta
	if p.timeCheckDelay < 0.35 {
		return
	}
	p.timeCheckDelay = delta

	p.timeCheckDelay = 0
	realTime := time.Now()
	if realTime.Before(p.nextTrackTime) {
		return // It's not the time yet
	}
	p.playNextTrack(realTime)
}
