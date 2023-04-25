package sim

import (
	"log"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/session"
)

type Controller struct {
	state *session.State

	scene *ge.Scene

	server *httpServer
}

func NewController(state *session.State) *Controller {
	return &Controller{state: state}
}

func (c *Controller) Init(scene *ge.Scene) {
	c.scene = scene

	log.SetFlags(log.Ltime)

	c.state.Persistent.Settings.EffectsVolumeLevel = 0
	c.state.Persistent.Settings.MusicVolumeLevel = 0
	scene.Context().Audio.SetGroupVolume(0, 0)
	scene.Context().Audio.SetGroupVolume(1, 0)

	c.server = newHTTPServer(":7070")
	c.server.EventRequest.Connect(nil, c.handleSimulation)
	c.server.EventShutdown.Connect(nil, c.handleShutdown)
	go c.server.Start()
}

func (c *Controller) Update(delta float64) {}

func (c *Controller) handleShutdown(err error) {
	panic(err)
}

func (c *Controller) handleSimulation(req simulationRequest) {
	config := req.Config.Clone()

	sim := staging.NewController(c.state, config, nil)
	sim.SetReplayActions(req.Actions)
	runner, scene := ge.NewSimulatedScene(c.scene.Context(), sim)
	sim.Init(scene)

	var simResult staging.GameSimulationResult
	start := time.Now()
	for {
		stop := false
		for i := 0; i < 60*60; i++ {
			runner.Update(1.0 / 60.0)
			simResult, stop = sim.GetSimulationResult()
		}
		if stop {
			break
		}
		if time.Since(start) >= (20 * time.Second) {
			c.server.resp.Err = "simulation takes too long"
			return
		}
	}

	c.server.resp.Results.Score = simResult.Score
	c.server.resp.Results.Time = int(simResult.Time)
	c.server.resp.Results.Victory = simResult.Victory
	c.server.resp.Results.Ticks = simResult.Ticks
}
