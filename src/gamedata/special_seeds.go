package gamedata

type SeedKind int

const (
	SeedNormal SeedKind = iota
	SeedLeet            // 1337
)

func GetSeedKind(seed int64, mode string) SeedKind {
	switch mode {
	case "classic":
		return classicSeedMap[seed]
	default:
		return SeedNormal
	}
}

var classicSeedMap = map[int64]SeedKind{
	// A very tough seed that unlocks the achievement.
	// Effects:
	// * +8 ion mortars on the map
	// * Every air base has a crawlers factory nearby
	// * x2 red crystals on the map
	1337: SeedLeet,
}
