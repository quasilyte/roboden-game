package gamedata

import "github.com/quasilyte/roboden-game/serverapi"

type SeedKind int

const (
	SeedNormal   SeedKind = iota
	SeedInfernal          // 666
	SeedLeet              // 1337
)

func GetSeedKind(seed int64, config serverapi.ReplayLevelConfig) SeedKind {
	switch config.RawGameMode {
	case "classic":
		return classicSeedMap[seed]
	case "arena":
		kind := arenaSeedMap[seed]
		switch kind {
		case SeedInfernal:
			if config.Environment != int(EnvInferno) {
				return SeedNormal
			}
		}
		return kind
	default:
		return SeedNormal
	}
}

var arenaSeedMap = map[int64]SeedKind{
	// A tough seed that unlocks the achievement.
	//
	// Requirements:
	// * Arena mode
	// * Inferno environment
	//
	// Effects:
	// * x3 lava geysers
	// * More lava (+80%)
	// * x2 creeps in waves
	// * Wave-spawned creeps have 65% hp (35% decrease)
	666: SeedInfernal,
}

var classicSeedMap = map[int64]SeedKind{
	// A very tough seed that unlocks the achievement.
	//
	// Requirements:
	// * Classic mode
	//
	// Effects:
	// * +8 ion mortars on the map
	// * Every air base has a crawlers factory nearby
	// * Dreadnought always spawns stealth crawlers
	// * x2 red crystals on the map
	1337: SeedLeet,
}
