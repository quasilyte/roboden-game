package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type messageManager struct {
	world *worldState

	messageTimer     float64
	messageTimeLimit float64
	message          *messageNode
	queue            []queuedMessageInfo
}

type queuedMessageInfo struct {
	targetPos     ge.Pos
	trackedObject ge.SceneObject
	text          string
	timer         float64
}

func newMessageManager(world *worldState) *messageManager {
	return &messageManager{world: world}
}

func (m *messageManager) Update(delta float64) {
	if m.message == nil && len(m.queue) == 0 {
		return
	}

	if m.message != nil {
		m.messageTimer += delta
		if m.messageTimer >= m.messageTimeLimit {
			m.message.Dispose()
			m.message = nil
		}
	}

	if m.message == nil && len(m.queue) != 0 {
		m.nextMessage()
	}
}

func (m *messageManager) AddMessage(info queuedMessageInfo) {
	m.queue = append(m.queue, info)
}

func (m *messageManager) nextMessage() {
	info := m.queue[0]
	copy(m.queue[:len(m.queue)-1], m.queue[1:])
	m.queue = m.queue[:len(m.queue)-1]

	m.messageTimer = 0
	m.messageTimeLimit = info.timer
	messagePos := gmath.Vec{X: 16, Y: 202}
	if info.targetPos.Base != nil {
		m.message = newWorldTutorialHintNode(m.world.camera, messagePos, info.targetPos, info.text)
	} else {
		m.message = newScreenTutorialHintNode(m.world.camera, messagePos, info.targetPos.Offset, info.text)
	}
	m.message.trackedObject = info.trackedObject
	m.world.rootScene.AddObject(m.message)
}
