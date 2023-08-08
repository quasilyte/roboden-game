package staging

import (
	"github.com/quasilyte/ge"
)

type nodeRunner struct {
	paused bool

	speedMultiplier float64

	world *worldState

	scene *ge.Scene

	timePlayed float64
	ticks      int

	creepCoordinator *creepCoordinator

	projectiles      []*projectileNode
	addedProjectiles []*projectileNode
	objects          []ge.SceneObject
	addedObjects     []ge.SceneObject
}

func newNodeRunner(speedMultiplier float64) *nodeRunner {
	return &nodeRunner{
		projectiles:      make([]*projectileNode, 0, 128),
		addedProjectiles: make([]*projectileNode, 0, 40),
		objects:          make([]ge.SceneObject, 0, 512),
		addedObjects:     make([]ge.SceneObject, 0, 32),
		speedMultiplier:  speedMultiplier,
	}
}

func (r *nodeRunner) Init(scene *ge.Scene) {
	r.scene = scene
}

func (r *nodeRunner) SetPaused(paused bool) {
	if paused {
		r.world.result.Paused = true
	}
	r.paused = paused
}

func (r *nodeRunner) IsPaused() bool {
	return r.paused
}

func (r *nodeRunner) AddProjectile(p *projectileNode) {
	r.addedProjectiles = append(r.addedProjectiles, p)
	p.Init(r.scene)
}

func (r *nodeRunner) AddObject(o ge.SceneObject) {
	r.addedObjects = append(r.addedObjects, o)
	o.Init(r.scene)
}

func (r *nodeRunner) ComputeDelta(delta float64) float64 {
	return delta * r.speedMultiplier
}

func (r *nodeRunner) runTick(computedDelta float64) {
	r.timePlayed += computedDelta
	r.ticks++

	r.creepCoordinator.Update(computedDelta)
	r.world.Update()

	liveProjectiles := r.projectiles[:0]
	for _, p := range r.projectiles {
		if p.IsDisposed() {
			p.world.freeProjectileNode(p)
			continue
		}
		p.Update(computedDelta)
		liveProjectiles = append(liveProjectiles, p)
	}
	r.projectiles = liveProjectiles
	r.projectiles = append(r.projectiles, r.addedProjectiles...)
	r.addedProjectiles = r.addedProjectiles[:0]

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

func (r *nodeRunner) Update(delta float64) {
	if r.paused {
		return
	}

	if r.speedMultiplier == 2 {
		// Run two ticks at x1 speed.
		r.runTick(delta)
		r.runTick(delta)
	} else {
		r.runTick(r.ComputeDelta(delta))
	}
}
