package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/viewport"
)

type messageManager struct {
	world *worldState

	cam *viewport.Camera

	messageTimer     float64
	messageTimeLimit float64
	message          *messageNode
	queue            []queuedMessageInfo
}

type queuedMessageInfo struct {
	forceWorldPos bool
	targetPos     ge.Pos
	trackedObject ge.SceneObject
	text          string
	timer         float64
	onReady       func()
}

func newMessageManager(world *worldState, cam *viewport.Camera) *messageManager {
	return &messageManager{
		world: world,
		cam:   cam,
	}
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

func (m *messageManager) MessageIsEmpty() bool {
	return len(m.queue) == 0 && m.message == nil
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
	worldPos := info.forceWorldPos || info.targetPos.Base != nil
	if worldPos {
		m.message = newWorldTutorialHintNode(m.cam, messagePos, info.targetPos, info.text)
	} else {
		m.message = newScreenTutorialHintNode(m.cam, messagePos, info.targetPos.Offset, info.text)
	}
	m.message.trackedObject = info.trackedObject
	m.world.rootScene.AddObject(m.message)
	if info.onReady != nil {
		info.onReady()
	}
}
