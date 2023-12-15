package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type tetherNode struct {
	world *worldState

	source *colonyAgentNode
	target targetable

	line *ge.TextureLine

	shaderTime float64
	lifespan   float64
	mayDetach  bool
}

func newTetherNode(world *worldState, source *colonyAgentNode, target targetable) *tetherNode {
	return &tetherNode{
		world:    world,
		target:   target,
		source:   source,
		lifespan: 8,
	}
}

func (tether *tetherNode) Init(scene *ge.Scene) {
	begin := ge.Pos{Base: tether.source.GetPos(), Offset: gmath.Vec{Y: -8}}
	end := ge.Pos{Base: tether.target.GetPos()}
	tether.line = ge.NewTextureLine(scene.Context(), begin, end)
	tether.line.SetTexture(gamedata.TetherBeaconAgentStats.BeamTexture)
	if tether.world.graphicsSettings.AllShadersEnabled {
		tether.line.Shader = scene.NewShader(assets.ShaderSlideX)
		tether.line.Shader.SetFloatValue("Time", 0)
	}
	tether.world.stage.AddGraphics(tether.line)

	if drone, ok := tether.target.(*colonyAgentNode); ok {
		tether.mayDetach = drone.stats.Kind == gamedata.AgentKamikaze
	}
}

func (tether *tetherNode) IsDisposed() bool {
	return tether.line.IsDisposed()
}

func (tether *tetherNode) dispose() {
	switch target := tether.target.(type) {
	case *colonyCoreNode:
		target.tether--
	case *colonyAgentNode:
		target.tether = false
	}
	tether.lifespan = 0
	tether.line.Dispose()
}

func (tether *tetherNode) Update(delta float64) {
	if !tether.line.Shader.IsNil() {
		tether.shaderTime += delta * gamedata.TetherBeaconAgentStats.BeamSlideSpeed
		tether.line.Shader.SetFloatValue("Time", tether.shaderTime)
	}

	if tether.target != nil {
		if tether.target.IsDisposed() {
			tether.dispose()
			return
		}
		if tether.mayDetach {
			mode := tether.target.(*colonyAgentNode).mode
			if mode == agentModeKamikazeAttack || mode == agentModeSentinelPatrol {
				tether.dispose()
				return
			}
		}
		if colony, ok := tether.target.(*colonyCoreNode); ok {
			if colony.waypoint.IsZero() {
				tether.dispose()
				return
			}
		}
	}

	if tether.source != nil && tether.source.IsDisposed() {
		tether.dispose()
		return
	}

	tether.lifespan -= delta
	beamRange := gamedata.TetherBeaconAgentStats.SupportRange
	distSqr := tether.source.pos.DistanceSquaredTo(*tether.target.GetPos())
	if distSqr > (beamRange * beamRange) {
		tether.lifespan -= delta * 3
	}
	if distSqr > (beamRange*beamRange)*1.25 {
		tether.dispose()
		return
	}
	if tether.lifespan <= 0 {
		tether.dispose()
		return
	}
}
