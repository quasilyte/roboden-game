package staging

import (
	"github.com/quasilyte/ge"
)

type nodeRunner struct {
	paused bool

	speedMultiplier float64

	victoryCheckDelay float64

	scene *ge.Scene

	timePlayed float64
	ticks      int

	creepCoordinator *creepCoordinator

	objects      []ge.SceneObject
	addedObjects []ge.SceneObject
}

func newNodeRunner(speedMultiplier float64) *nodeRunner {
	return &nodeRunner{
		objects:         make([]ge.SceneObject, 0, 512),
		addedObjects:    make([]ge.SceneObject, 0, 32),
		speedMultiplier: speedMultiplier,
	}
}

func (r *nodeRunner) Init(scene *ge.Scene) {
	r.scene = scene
}

func (r *nodeRunner) SetPaused(paused bool) {
	r.paused = paused
}

func (r *nodeRunner) IsPaused() bool {
	return r.paused
}

func (r *nodeRunner) AddObject(o ge.SceneObject) {
	r.addedObjects = append(r.addedObjects, o)
	o.Init(r.scene)
}

func (r *nodeRunner) Update(delta float64) {
	if r.paused {
		return
	}

	computedDelta := delta * r.speedMultiplier
	r.timePlayed += computedDelta
	r.ticks++

	r.creepCoordinator.Update(computedDelta)

	liveObjects := r.objects[:0]
	for _, o := range r.objects {
		if o.IsDisposed() {
			continue
		}
		o.Update(computedDelta)
		liveObjects = append(liveObjects, o)
	}
	r.objects = liveObjects
	r.objects = append(r.objects, r.addedObjects...)
	r.addedObjects = r.addedObjects[:0]
}
