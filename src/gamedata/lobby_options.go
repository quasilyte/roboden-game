package gamedata

type LobbyOption struct {
	ScoreCost int
	Category  string
}

var LobbyOptionMap = map[string]LobbyOption{
	"super_creeps":       {ScoreCost: SuperCreepsOptionCost, Category: "difficulty"},
	"creep_fortress":     {ScoreCost: FortressOptionCost, Category: "difficulty"},
	"ion_mortars":        {ScoreCost: IonMortarOptionCost, Category: "difficulty"},
	"coordinator_creeps": {ScoreCost: CoordinatorCreepsOptionCost, Category: "difficulty"},
	"grenadier_creeps":   {ScoreCost: GrenadierCreepsOptionCost, Category: "difficulty"},
}
