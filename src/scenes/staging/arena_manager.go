package staging

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/timeutil"
)

type arenaCreepInfo struct {
	stats    *creepStats
	cost     int
	minLevel int
}

type arenaManager struct {
	level      int
	waveBudget int

	info            *messageNode
	overviewText    string
	infoUpdateDelay float64
	levelStartDelay float64

	attackSides []int

	spawnAreas []gmath.Rect

	waveInfo arenaWaveInfo

	scene *ge.Scene
	world *worldState

	creepSelectionSlice      []*arenaCreepInfo
	groupCreepSelectionSlice []*arenaCreepInfo
	basicFlyingCreeps        []*arenaCreepInfo
	basicGroundCreeps        []*arenaCreepInfo

	crawlerCreepInfo        *arenaCreepInfo
	eliteCrawlerCreepInfo   *arenaCreepInfo
	stealthCrawlerCreepInfo *arenaCreepInfo
	wandererCreepInfo       *arenaCreepInfo
	stunnerCreepInfo        *arenaCreepInfo
	assaultCreepInfo        *arenaCreepInfo
	builderCreepInfo        *arenaCreepInfo
}

type arenaWaveGroup struct {
	units []*creepStats
	side  int
}

type arenaWaveInfo struct {
	groups []arenaWaveGroup

	builders        bool
	flyingAttackers bool
	groundAttackers bool

	attackSides [4]bool
}

func newArenaManager(world *worldState) *arenaManager {
	return &arenaManager{
		world: world,
		waveInfo: arenaWaveInfo{
			groups: make([]arenaWaveGroup, 0, 8),
		},
		attackSides:              []int{0, 1, 2, 3},
		creepSelectionSlice:      make([]*arenaCreepInfo, 0, 16),
		groupCreepSelectionSlice: make([]*arenaCreepInfo, 0, 16),
	}
}

func (m *arenaManager) IsDisposed() bool {
	return false
}

func (m *arenaManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.level = 1
	m.waveBudget = 20
	m.levelStartDelay = 90

	m.crawlerCreepInfo = &arenaCreepInfo{
		stats: crawlerCreepStats,
		cost:  4,
	}
	m.eliteCrawlerCreepInfo = &arenaCreepInfo{
		stats:    eliteCrawlerCreepStats,
		cost:     6,
		minLevel: 2,
	}
	m.stealthCrawlerCreepInfo = &arenaCreepInfo{
		stats:    stealthCrawlerCreepStats,
		cost:     7,
		minLevel: 3,
	}
	m.wandererCreepInfo = &arenaCreepInfo{
		stats: wandererCreepStats,
		cost:  6,
	}
	m.stunnerCreepInfo = &arenaCreepInfo{
		stats:    stunnerCreepStats,
		cost:     9,
		minLevel: 2,
	}
	m.assaultCreepInfo = &arenaCreepInfo{
		stats:    assaultCreepStats,
		cost:     15,
		minLevel: 5,
	}
	m.builderCreepInfo = &arenaCreepInfo{
		stats:    builderCreepStats,
		cost:     25,
		minLevel: 7,
	}

	m.basicFlyingCreeps = []*arenaCreepInfo{
		m.wandererCreepInfo,
		m.stunnerCreepInfo,
		m.assaultCreepInfo,
	}
	m.basicGroundCreeps = []*arenaCreepInfo{
		m.crawlerCreepInfo,
		m.eliteCrawlerCreepInfo,
		m.stealthCrawlerCreepInfo,
	}

	pad := 160.0
	offscreenPad := 160.0
	m.spawnAreas = []gmath.Rect{
		// right border (east)
		{Min: gmath.Vec{X: m.world.width, Y: pad}, Max: gmath.Vec{X: m.world.width + offscreenPad, Y: m.world.height - pad}},
		// bottom border (south)
		{Min: gmath.Vec{X: pad, Y: m.world.height}, Max: gmath.Vec{X: m.world.width - pad, Y: m.world.height + offscreenPad}},
		// left border (west)
		{Min: gmath.Vec{X: -offscreenPad, Y: pad}, Max: gmath.Vec{X: 0, Y: m.world.height - pad}},
		// top border (north)
		{Min: gmath.Vec{X: pad, Y: -offscreenPad}, Max: gmath.Vec{X: m.world.width - pad, Y: 0}},
	}

	m.infoUpdateDelay = 5
	m.prepareWaveInfo()
	m.overviewText = m.createWaveOverviewText()
	m.info = m.createWaveInfoMessageNode()
	scene.AddObject(m.info)
}

func (m *arenaManager) incLevel() {

	m.level++
	if m.level%5 == 0 {
		m.levelStartDelay = 4.0 * 60
		m.waveBudget += 25
	} else {
		m.levelStartDelay = 2.5 * 60
		m.waveBudget += 10
	}
}

func (m *arenaManager) Update(delta float64) {
	m.levelStartDelay -= delta
	if m.levelStartDelay <= 0 {
		m.spawnCreeps()
		m.incLevel()
		m.prepareWaveInfo()
		m.overviewText = m.createWaveOverviewText()
		if m.info != nil {
			m.info.Dispose()
		}
		m.info = m.createWaveInfoMessageNode()
		m.scene.AddObject(m.info)
	}

	m.infoUpdateDelay -= delta
	if m.infoUpdateDelay <= 0 {
		m.infoUpdateDelay = 5 + m.infoUpdateDelay
		m.info.UpdateText(m.createWaveInfoText())
	}
}

func (m *arenaManager) createWaveInfoMessageNode() *messageNode {
	s := m.createWaveInfoText()
	message := newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
	message.xpadding = 20
	return message
}

func (m *arenaManager) createWaveOverviewText() string {
	d := m.scene.Dict()

	var buf strings.Builder
	buf.Grow(128)

	buf.WriteString(d.Get("game.wave_direction"))
	buf.WriteString(": ")
	if m.waveInfo.attackSides == [4]bool{true, true, true, true} {
		buf.WriteString(d.Get("game.side.all"))
	} else {
		sideParts := make([]string, 0, 2)
		for side, hasAttackers := range m.waveInfo.attackSides {
			if !hasAttackers {
				continue
			}
			switch side {
			case 0:
				sideParts = append(sideParts, d.Get("game.side.east"))
			case 1:
				sideParts = append(sideParts, d.Get("game.side.south"))
			case 2:
				sideParts = append(sideParts, d.Get("game.side.west"))
			case 3:
				sideParts = append(sideParts, d.Get("game.side.north"))
			}
		}
		buf.WriteString(strings.Join(sideParts, ", "))
	}

	unitKindParts := make([]string, 0, 4)
	if m.waveInfo.groundAttackers {
		unitKindParts = append(unitKindParts, d.Get("drone.target.ground"))
	}
	if m.waveInfo.flyingAttackers {
		unitKindParts = append(unitKindParts, d.Get("drone.target.flying"))
	}
	buf.WriteByte('\n')
	buf.WriteString(d.Get("game.wave_units"))
	buf.WriteString(": ")
	buf.WriteString(strings.Join(unitKindParts, ", "))
	if m.waveInfo.builders {
		buf.WriteByte('\n')
		buf.WriteString(d.Get("game.wave_special_units"))
		buf.WriteString(": ")
		buf.WriteString(d.Get("game.wave_builders"))
	}

	return buf.String()
}

func (m *arenaManager) createWaveInfoText() string {
	d := m.scene.Dict()

	var buf strings.Builder
	buf.Grow(256)
	buf.WriteString(d.Get("game.wave"))
	buf.WriteByte(' ')
	buf.WriteString(strconv.Itoa(m.level))
	buf.WriteString(" ")
	buf.WriteString(d.Get("game.wave_starts_in"))
	buf.WriteByte(' ')
	buf.WriteString(timeutil.FormatDuration(d, time.Duration(m.levelStartDelay*float64(time.Second))))
	if m.overviewText != "" {
		buf.WriteByte('\n')
		buf.WriteString(m.overviewText)
	}

	return buf.String()
}

func (m *arenaManager) spawnCreeps() {
	for _, g := range m.waveInfo.groups {
		sector := m.spawnAreas[g.side]
		spawnPos := randomSectorPos(m.world.rand, sector)
		targetPos := correctedPos(m.world.rect, randomSectorPos(m.world.rand, sector), 520)
		for _, creepStats := range g.units {
			creepPos := spawnPos
			spawnDelay := 0.0
			if creepStats.shadowImage == assets.ImageNone {
				attemptPos := creepPos.Add(m.world.rand.Offset(-60, 60))
				// Ground unit.
				// Spawn them really close to the map edge.
				deployed := false
				for i := 0; i < 4; i++ {
					if attemptPos.X <= 0 {
						spawnDelay = (-attemptPos.X) / creepStats.speed
						attemptPos.X = 1
					} else if attemptPos.X >= m.world.width {
						spawnDelay = (attemptPos.X - m.world.width) / creepStats.speed
						attemptPos.X = m.world.width - 1
					}
					if attemptPos.Y <= 0 {
						spawnDelay = (-attemptPos.Y) / creepStats.speed
						attemptPos.Y = 1
					} else if attemptPos.Y >= m.world.height {
						spawnDelay = (attemptPos.Y - m.world.height) / creepStats.speed
						attemptPos.Y = m.world.height - 1
					}
					coord := m.world.pathgrid.PosToCoord(attemptPos)
					if m.world.pathgrid.CellIsFree(coord) {
						deployed = true
						creepPos = attemptPos
						break
					}
				}
				if !deployed {
					continue
				}
			} else {
				creepPos = creepPos.Add(m.world.rand.Offset(-60, 60))
			}
			creepTargetPos := targetPos.Add(m.world.rand.Offset(-60, 60))
			if spawnDelay > 0 {
				spawner := newCreepSpawnerNode(m.world, spawnDelay, creepPos, creepTargetPos, creepStats)
				m.scene.AddObject(spawner)
			} else {
				creep := m.world.NewCreepNode(creepPos, creepStats)
				m.scene.AddObject(creep)
				creep.SendTo(creepTargetPos)
			}
		}
	}
}

func (m *arenaManager) prepareWaveInfo() {
	budget := m.waveBudget

	// First decide which kind of attack we're doing.
	attackDirectionRoll := m.world.rand.Float()
	numAttackSides := 0
	budgetMultiplier := 1.0
	switch {
	case attackDirectionRoll < 0.5:
		numAttackSides = 1
	case attackDirectionRoll < 0.8:
		numAttackSides = 2
		budgetMultiplier = 0.75
	default:
		numAttackSides = 4
		budgetMultiplier = 0.4
	}

	groups := m.waveInfo.groups[:0]
	m.waveInfo = arenaWaveInfo{}

	gmath.Shuffle(m.world.rand, m.attackSides)
	sides := m.attackSides[:numAttackSides]
	for _, side := range sides {
		m.waveInfo.attackSides[side] = true
		sideBudget := int(math.Round(float64(budget) * budgetMultiplier))
		creepSelection := m.creepSelectionSlice[:0]
		selectionRoll := m.world.rand.Float()
		allowFlying := true
		switch {
		case selectionRoll <= 0.5:
			// Flying-only creeps.
			creepSelection = append(creepSelection, m.basicFlyingCreeps...)
			m.waveInfo.flyingAttackers = true
		case selectionRoll <= 0.8:
			// Ground-only creeps.
			creepSelection = append(creepSelection, m.basicGroundCreeps...)
			m.waveInfo.groundAttackers = true
			allowFlying = false
		default:
			creepSelection = append(creepSelection, m.basicFlyingCreeps...)
			creepSelection = append(creepSelection, m.basicGroundCreeps...)
			m.waveInfo.flyingAttackers = true
			m.waveInfo.groundAttackers = true
		}
		if allowFlying && m.world.rand.Chance(0.45) {
			creepSelection = append(creepSelection, m.builderCreepInfo)
		}

		const maxGroupBudget = 90
		for sideBudget > 0 {
			groupCreepSelection := m.groupCreepSelectionSlice[:0]
			groupCreepSelection = append(groupCreepSelection, creepSelection...)
			g := arenaWaveGroup{side: side}
			localBudget := sideBudget
			if localBudget > maxGroupBudget {
				localBudget = maxGroupBudget
			}
			sideBudget -= localBudget
			for {
				creep, budgetRemaining, ok := m.pickUnit(localBudget, groupCreepSelection)
				if !ok {
					break
				}
				if creep.kind == creepBuilder {
					m.waveInfo.builders = true
					groupCreepSelection = xslices.Remove(groupCreepSelection, m.builderCreepInfo)
				}
				localBudget = budgetRemaining
				g.units = append(g.units, creep)
			}
			groups = append(groups, g)
		}
	}

	m.waveInfo.groups = groups
}

func (m *arenaManager) pickUnit(budget int, selection []*arenaCreepInfo) (*creepStats, int, bool) {
	if budget < selection[0].cost {
		return nil, budget, false
	}
	creepInfo := randIterate(m.world.rand, selection, func(x *arenaCreepInfo) bool {
		return x.cost <= budget && x.minLevel <= m.level
	})
	if creepInfo.cost != 0 && creepInfo.cost <= budget {
		return creepInfo.stats, budget - creepInfo.cost, true
	}
	return nil, budget, false
}
