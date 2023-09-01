package staging

import (
	"runtime"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/viewport"
)

type cameraMode int

const (
	camManual cameraMode = iota
	camCinematic
	camCinematicOffline
)

type cameraManager struct {
	*viewport.Camera

	world *worldState
	mode  cameraMode
	input *gameinput.Handler

	wheelScrollStyle gameinput.WheelScrollStyle

	cameraPanStartPos        gmath.Vec
	cameraPanDragPos         gmath.Vec
	cameraPanSpeed           float64
	cameraDragSpeed          float64
	cameraPanBoundary        float64
	cameraToggleSpeed        float64
	cameraToggleProgress     float64
	cameraToggleTarget       gmath.Vec
	cameraToggleSnapDistSqr  float64
	cameraToggleSnapProgress float64

	cinematicSwitchDelay float64
}

func newCameraManager(world *worldState, cam *viewport.Camera) *cameraManager {
	m := &cameraManager{
		world:  world,
		Camera: cam,
	}

	switch m.world.gameSettings.ScrollingSpeed {
	case 0:
		m.cameraDragSpeed = 0.5
	case 1:
		m.cameraDragSpeed = 0.8
	case 2:
		// The default speed, x1 factor.
		// This is the most pleasant and convenient to use, but could
		// be too slow for a pro player.
		m.cameraDragSpeed = 1
	case 3:
		// Just a bit faster.
		m.cameraDragSpeed = 1.2
	case 4:
		m.cameraDragSpeed = 2
	}
	m.cameraPanSpeed = float64(m.world.gameSettings.ScrollingSpeed+1) * 4

	if m.world.gameSettings.EdgeScrollRange != 0 {
		m.cameraPanBoundary = 1
		if runtime.GOARCH == "wasm" {
			m.cameraPanBoundary = 8
		}
		m.cameraPanBoundary += 2 * float64(m.world.gameSettings.EdgeScrollRange-1)
	}

	m.wheelScrollStyle = gameinput.WheelScrollStyle(m.world.gameSettings.WheelScrollingMode)

	return m
}

func (m *cameraManager) InitManualMode(h *gameinput.Handler) {
	m.mode = camManual
	m.input = h
	m.cameraToggleSpeed = 1
	m.cameraToggleSnapDistSqr = 80 * 80
	m.cameraToggleSnapProgress = 0.9
	if !h.HasMouseInput() {
		m.cameraPanBoundary = 0
	}
}

func (m *cameraManager) InitCinematicMode() {
	m.mode = camCinematicOffline
	m.cameraToggleSpeed = 0.2
	m.cameraToggleSnapDistSqr = 6 * 6
	m.cameraToggleSnapProgress = 0.99
}

func (m *cameraManager) HandleInput() {
	if m.mode != camManual {
		return
	}

	if !m.world.deviceInfo.IsMobile() {
		// Camera panning only makes sense on non-mobile devices
		// where we have a keyboard/gamepad or a cursor.
		var cameraPan gmath.Vec
		if m.input.ActionIsPressed(controls.ActionPanRight) {
			cameraPan.X += m.cameraPanSpeed
		}
		if m.input.ActionIsPressed(controls.ActionPanDown) {
			cameraPan.Y += m.cameraPanSpeed
		}
		if m.input.ActionIsPressed(controls.ActionPanLeft) {
			cameraPan.X -= m.cameraPanSpeed
		}
		if m.input.ActionIsPressed(controls.ActionPanUp) {
			cameraPan.Y -= m.cameraPanSpeed
		}
		if cameraPan.IsZero() {
			if m.wheelScrollStyle == gameinput.WheelScrollDrag {
				if info, ok := m.input.JustPressedActionInfo(controls.ActionPanAlt); ok {
					m.cameraPanDragPos = m.Offset
					m.cameraPanStartPos = info.Pos
				} else if info, ok := m.input.PressedActionInfo(controls.ActionPanAlt); ok {
					m.cameraToggleTarget = gmath.Vec{}
					posDelta := m.cameraPanStartPos.Sub(info.Pos).Mulf(m.cameraDragSpeed)
					newPos := m.cameraPanDragPos.Add(posDelta)
					m.SetOffset(newPos)
				}
			} else {
				if info, ok := m.input.PressedActionInfo(controls.ActionPanAlt); ok {
					cameraCenter := m.Rect.Center()
					cameraPan = gmath.RadToVec(cameraCenter.AngleToPoint(info.Pos)).Mulf(m.cameraPanSpeed * 0.8)
				}
			}
		}
		if cameraPan.IsZero() && m.cameraPanBoundary != 0 {
			// Mouse cursor can pan the camera too.
			cursor := m.input.CursorPos().Sub(m.ScreenPos)
			if cursor.X >= m.Rect.Width()-m.cameraPanBoundary {
				cameraPan.X += m.cameraPanSpeed
			}
			if cursor.Y >= m.Rect.Height()-m.cameraPanBoundary {
				cameraPan.Y += m.cameraPanSpeed
			}
			if cursor.X < m.cameraPanBoundary {
				cameraPan.X -= m.cameraPanSpeed
			}
			if cursor.Y < m.cameraPanBoundary {
				cameraPan.Y -= m.cameraPanSpeed
			}
		}
		if !cameraPan.IsZero() {
			m.cameraToggleTarget = gmath.Vec{}
		}
		m.Pan(cameraPan)
	} else {
		// On mobile devices we expect a touch screen support.
		// Instead of panning, we use dragging here.
		if m.input.ActionIsJustPressed(controls.ActionPanDrag) {
			m.cameraPanDragPos = m.Offset
		}
		if info, ok := m.input.PressedActionInfo(controls.ActionPanDrag); ok {
			m.cameraToggleTarget = gmath.Vec{}
			posDelta := info.StartPos.Sub(info.Pos).Mulf(m.cameraDragSpeed)
			newPos := m.cameraPanDragPos.Add(posDelta)
			m.SetOffset(newPos)
		}
	}
}

func (m *cameraManager) findTwoColonies(pstate *playerState, dist float64) [2]*colonyCoreNode {
	for _, c1 := range pstate.colonies {
		for _, c2 := range pstate.colonies {
			if c1 == c2 {
				continue
			}
			if c1.pos.DistanceTo(c2.pos) <= dist {
				return [2]*colonyCoreNode{c1, c2}
			}
		}
	}
	return [2]*colonyCoreNode{}
}

func (m *cameraManager) Update(delta float64) {
	if !m.cameraToggleTarget.IsZero() {
		m.cameraToggleProgress = gmath.ClampMax(m.cameraToggleProgress+(delta*m.cameraToggleSpeed), 1)
		m.CenterOn(m.CenterPos().LinearInterpolate(m.cameraToggleTarget, m.cameraToggleProgress))
		if m.cameraToggleProgress >= m.cameraToggleSnapProgress || m.CenterPos().DistanceSquaredTo(m.cameraToggleTarget) < m.cameraToggleSnapDistSqr {
			m.CenterOn(m.cameraToggleTarget)
			m.cameraToggleTarget = gmath.Vec{}
		}
	}

	if m.mode == camCinematic {
		m.cinematicSwitchDelay = gmath.ClampMin(m.cinematicSwitchDelay-delta, 0)
		if m.cinematicSwitchDelay == 0 {
			pstate := gmath.RandElem(m.world.localRand, m.world.players).GetState()
			if m.world.boss != nil && m.world.localRand.Chance(0.6) {
				for _, c := range pstate.colonies {
					if c.pos.DistanceTo(m.world.boss.pos) < 500 {
						if m.world.localRand.Chance(0.85) {
							m.cinematicSwitchDelay = m.world.localRand.FloatRange(10, 14)
							m.ToggleCamera(midpoint(m.world.boss.pos, c.pos))
							return
						}
					}
				}
			}
			if m.world.boss != nil && !m.world.boss.IsFlying() && m.world.localRand.Chance(0.9) {
				m.cinematicSwitchDelay = m.world.localRand.FloatRange(8, 12)
				m.ToggleCamera(m.world.boss.pos.Add(m.world.localRand.Offset(-40, 40)))
				return
			}
			roll := m.world.localRand.Float()
			switch {
			case roll < 0.55: // 55%
				// Show one of the colonies.
				m.cinematicSwitchDelay = m.world.localRand.FloatRange(15, 25)
				if len(pstate.colonies) > 0 {
					pos := gmath.RandElem(m.world.localRand, pstate.colonies).pos
					if len(pstate.colonies) > 1 && m.world.localRand.Chance(0.4) {
						pair := m.findTwoColonies(pstate, 320)
						if pair[0] != nil {
							pos = midpoint(pair[0].pos, pair[1].pos)
						}
					} else {
						if m.world.localRand.Chance(0.3) {
							pos.X += m.world.localRand.FloatRange(-40, 40)
						}
					}
					if m.world.localRand.Bool() {
						pos = pos.Add(m.world.localRand.Offset(-32, 32))
					}
					m.ToggleCamera(pos)
				}
			case roll < 0.7: // 15%
				// Show enemy colony.
				m.cinematicSwitchDelay = m.world.localRand.FloatRange(7, 10)
				if m.world.boss != nil {
					m.ToggleCamera(m.world.boss.pos)
				}
			case roll < 0.9: // 20%
				// Show an interesting creep.
				creep := randIterate(m.world.localRand, m.world.creeps, func(creep *creepNode) bool {
					switch creep.stats.Kind {
					case gamedata.CreepHowitzer, gamedata.CreepBuilder, gamedata.CreepServant:
						return true
					case gamedata.CreepTurret:
						if creep.stats == gamedata.IonMortarCreepStats {
							return m.world.localRand.Chance(0.5)
						}
					case gamedata.CreepFortress:
						return m.world.localRand.Chance(0.6)
					case gamedata.CreepCrawlerBase, gamedata.CreepBase:
						return m.world.localRand.Chance(0.35)
					case gamedata.CreepCrawler:
						if m.world.localRand.Chance(0.8) {
							return !creep.waypoint.IsZero()
						}
					}
					return false
				})
				if creep != nil {
					m.cinematicSwitchDelay = m.world.localRand.FloatRange(8, 10)
					m.ToggleCamera(creep.pos)
				}
			case roll < 0.95: // 5%
				// Show some random creep.
				m.cinematicSwitchDelay = m.world.localRand.FloatRange(6, 8)
				if len(m.world.creeps) != 0 {
					m.ToggleCamera(gmath.RandElem(m.world.localRand, m.world.creeps).pos)
				}
			default: // 5%
				// Show a random pos on a map.
				m.cinematicSwitchDelay = m.world.localRand.FloatRange(6, 8)
				m.ToggleCamera(randomSectorPos(m.world.localRand, m.world.innerRect))
			}
		}
	}
}

func (m *cameraManager) ToggleCamera(pos gmath.Vec) {
	m.cameraToggleTarget = pos
	m.cameraToggleProgress = 0
}
