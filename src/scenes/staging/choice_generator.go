package staging

import (
	"strings"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type specialChoiceKind int

const (
	specialChoiceNone specialChoiceKind = iota
	specialIncreaseRadius
	specialDecreaseRadius
	specialBuildGunpoint
	specialBuildColony
	specialAttack
	specialChoiceMoveColony
)

type selectedChoice struct {
	Index    int
	Cooldown float64
	Faction  gamedata.FactionTag
	Option   choiceOption
	Pos      gmath.Vec
	Player   player
}

type choiceOption struct {
	text    string
	effects []choiceOptionEffect
	special specialChoiceKind
	icon    resource.ImageID
	cost    float64
}

type choiceOptionEffect struct {
	priority colonyPriority
	value    float64
}

type choiceSelection struct {
	cards   []choiceOption
	special choiceOption
}

var specialChoicesTable = [...]choiceOption{
	specialAttack: {
		text:    "attack",
		special: specialAttack,
		cost:    5,
		icon:    assets.ImageActionAttack,
	},

	specialBuildColony: {
		text:    "build_colony",
		special: specialBuildColony,
		cost:    25,
		icon:    assets.ImageActionBuildColony,
	},
	specialBuildGunpoint: {
		text:    "build_gunpoint",
		special: specialBuildGunpoint,
		cost:    10,
		icon:    assets.ImageActionBuildTurret,
	},

	specialIncreaseRadius: {
		text:    "increase_radius",
		special: specialIncreaseRadius,
		cost:    15,
		icon:    assets.ImageActionIncreaseRadius,
	},
	specialDecreaseRadius: {
		text:    "decrease_radius",
		special: specialDecreaseRadius,
		cost:    4,
		icon:    assets.ImageActionDecreaseRadius,
	},
}

var choiceOptionList = []choiceOption{
	{
		text: "resources",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.2},
		},
	},
	{
		text: "growth",
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.2},
		},
	},
	{
		text: "security",
		effects: []choiceOptionEffect{
			{priority: prioritySecurity, value: 0.2},
		},
	},
	{
		text: "evolution",
		effects: []choiceOptionEffect{
			{priority: priorityEvolution, value: 0.2},
		},
	},

	{
		text: "resources+growth",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: priorityGrowth, value: 0.15},
		},
	},
	{
		text: "resources+security",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: prioritySecurity, value: 0.15},
		},
	},
	{
		text: "resources+evolution",
		effects: []choiceOptionEffect{
			{priority: priorityResources, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
	{
		text: "growth+security",
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.15},
			{priority: prioritySecurity, value: 0.15},
		},
	},
	{
		text: "growth+evolution",
		effects: []choiceOptionEffect{
			{priority: priorityGrowth, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
	{
		text: "security+evolution",
		effects: []choiceOptionEffect{
			{priority: prioritySecurity, value: 0.15},
			{priority: priorityEvolution, value: 0.15},
		},
	},
}

type choiceState int

const (
	choiceCharging choiceState = iota
	choiceReady
)

type choiceGenerator struct {
	targetValue float64
	value       float64

	world *worldState

	state choiceState

	player player

	shuffledOptions []choiceOption

	beforeSpecialShuffle int
	specialOptionIndex   int
	buildTurret          bool
	increaseRadius       bool
	specialChoiceKinds   []specialChoiceKind
	specialChoices       []choiceOption

	EventChoiceReady    gsignal.Event[choiceSelection]
	EventChoiceSelected gsignal.Event[selectedChoice]
}

func newChoiceGenerator(world *worldState) *choiceGenerator {
	g := &choiceGenerator{
		world: world,
	}

	g.shuffledOptions = make([]choiceOption, len(choiceOptionList))
	copy(g.shuffledOptions, choiceOptionList)

	d := world.rootScene.Dict()

	translateText := func(s string) string {
		keys := strings.Split(s, "+")
		for _, k := range keys {
			s = strings.Replace(s, k, d.Get("game.choice", k), 1)
		}
		s = strings.ReplaceAll(s, "+", "\n+\n")
		s = strings.ReplaceAll(s, " ", "\n")
		return s
	}

	// Now translate the options.
	for i := range g.shuffledOptions {
		o := &g.shuffledOptions[i]
		o.text = translateText(o.text)
	}

	g.specialChoiceKinds = []specialChoiceKind{
		specialBuildColony,
	}

	if world.config.AttackActionAvailable {
		g.specialChoiceKinds = append(g.specialChoiceKinds, specialAttack)
	}
	if world.config.RadiusActionAvailable {
		g.specialChoiceKinds = append(g.specialChoiceKinds, specialDecreaseRadius)
	}

	// Now translate the special choices.
	g.specialChoices = make([]choiceOption, len(specialChoicesTable))
	copy(g.specialChoices, specialChoicesTable[:])
	for i := range g.specialChoices {
		o := &g.specialChoices[i]
		if o.text == "" {
			continue
		}
		o.text = translateText(o.text)
	}

	return g
}

func (g *choiceGenerator) Init(scene *ge.Scene) {}

func (g *choiceGenerator) IsDisposed() bool { return false }

func (g *choiceGenerator) IsReady() bool {
	return g.state == choiceReady
}

func (g *choiceGenerator) Update(delta float64) {
	if g.state != choiceCharging {
		return
	}

	g.value += delta
	if g.value >= g.targetValue {
		g.generateChoices()
		return
	}
}

func (g *choiceGenerator) TryExecute(cardIndex int, pos gmath.Vec) bool {
	if g.player.GetState().selectedColony.mode != colonyModeNormal {
		return false
	}
	if cardIndex != -1 {
		return g.activateChoice(cardIndex)
	}
	return g.activateMoveChoice(pos)
}

func (g *choiceGenerator) activateMoveChoice(pos gmath.Vec) bool {
	if g.state != choiceReady {
		return false
	}
	cooldown := 8.0
	g.startCharging(cooldown)
	g.EventChoiceSelected.Emit(selectedChoice{
		Index:    -1,
		Option:   choiceOption{special: specialChoiceMoveColony},
		Pos:      pos,
		Cooldown: cooldown,
		Player:   g.player,
	})
	return true
}

func (g *choiceGenerator) activateChoice(i int) bool {
	if g.state != choiceReady {
		return false
	}

	choice := selectedChoice{
		Faction: gamedata.FactionTag(i + 1),
		Index:   i,
		Player:  g.player,
	}
	cooldown := 10.0
	if i == 4 {
		// A special action is selected.
		cooldown = g.specialChoices[g.specialOptionIndex].cost
		choice.Option = g.specialChoices[g.specialOptionIndex]
	} else {
		choice.Option = g.shuffledOptions[i]
	}
	choice.Cooldown = cooldown

	g.startCharging(cooldown)

	g.EventChoiceSelected.Emit(choice)
	return true
}

func (g *choiceGenerator) startCharging(targetValue float64) {
	g.value = 0
	g.targetValue = targetValue
	g.state = choiceCharging
}

func (g *choiceGenerator) generateChoices() {
	g.state = choiceReady

	gmath.Shuffle(g.world.rand, g.shuffledOptions)

	if g.beforeSpecialShuffle == 0 {
		g.buildTurret = !g.buildTurret
		g.increaseRadius = !g.increaseRadius
		gmath.Shuffle(g.world.rand, g.specialChoiceKinds)
		g.beforeSpecialShuffle = len(g.specialChoiceKinds)
	}
	g.beforeSpecialShuffle--
	specialIndex := g.beforeSpecialShuffle

	specialOptionKind := g.specialChoiceKinds[specialIndex]
	switch specialOptionKind {
	case specialBuildColony:
		if g.buildTurret && g.world.config.BuildTurretActionAvailable {
			specialOptionKind = specialBuildGunpoint
		}
	case specialDecreaseRadius:
		if g.increaseRadius {
			specialOptionKind = specialIncreaseRadius
		}
	}
	g.specialOptionIndex = int(specialOptionKind)

	g.EventChoiceReady.Emit(choiceSelection{
		cards:   g.shuffledOptions[:4],
		special: g.specialChoices[g.specialOptionIndex],
	})
}
