package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/viewport"
)

type messageManager struct {
	world *worldState

	cam *viewport.Camera

	mainMessage *messageNode

	messageTimer     float64
	messageTimeLimit float64
	message          *messageNode
	queue            []queuedMessageInfo

	EventMessageClicked     gsignal.Event[gmath.Vec]
	EventMainMessageClicked gsignal.Event[gsignal.Void]
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
	if m.mainMessage != nil && m.mainMessage.IsDisposed() {
		m.mainMessage = nil
	}

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

func (m *messageManager) SetMainMessage(info queuedMessageInfo) {
	messagePos := gmath.Vec{X: 16, Y: 70}
	worldPos := info.forceWorldPos || info.targetPos.Base != nil
	if worldPos {
		m.mainMessage = newWorldTutorialHintNode(m.cam, messagePos, info.targetPos, info.text)
	} else {
		m.mainMessage = newScreenTutorialHintNode(m.cam, messagePos, info.targetPos.Offset, info.text)
	}
	m.world.rootScene.AddObject(m.mainMessage)
}

func (m *messageManager) HandleInput(clickPos gmath.Vec) bool {
	if m.mainMessage != nil {
		if m.mainMessage.ContainsPos(clickPos) {
			m.EventMainMessageClicked.Emit(gsignal.Void{})
			return true
		}
	}

	if m.message != nil {
		if m.message.ContainsPos(clickPos) {
			m.EventMessageClicked.Emit(m.message.targetPos.Resolve())
			return true
		}
	}

	return false
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
