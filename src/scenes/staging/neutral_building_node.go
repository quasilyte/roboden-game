package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type neutralBuildingNode struct {
	world *worldState
	pos   gmath.Vec
	stats *gamedata.AgentStats

	sprite *ge.Sprite

	agent *colonyAgentNode
}

func newNeutralBuildingNode(world *worldState, stats *gamedata.AgentStats, pos gmath.Vec) *neutralBuildingNode {
	return &neutralBuildingNode{
		world: world,
		pos:   pos,
		stats: stats,
	}
}

func (b *neutralBuildingNode) Init(scene *ge.Scene) {
	b.sprite = scene.NewSprite(b.stats.Image)
	b.sprite.Pos.Base = &b.pos
	b.sprite.SetColorScaleRGBA(240, 240, 240, 255)
	b.sprite.Shader = scene.NewShader(assets.ShaderColonyDamage)
	b.sprite.Shader.Texture1 = scene.LoadImage(assets.ImageBuildingDamageMask)
	b.sprite.Shader.SetFloatValue("HP", 0.001)
	b.world.stage.AddSprite(b.sprite)

	b.world.MarkPos(b.pos)
}

func (b *neutralBuildingNode) AssignAgent(a *colonyAgentNode) {
	b.sprite.Visible = a == nil
	b.agent = a

	if a != nil {
		a.EventDestroyed.Connect(nil, func(*colonyAgentNode) {
			b.AssignAgent(nil)
		})
	}
}
