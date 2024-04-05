package assets

import (
	"github.com/quasilyte/roboden-game/assets/monofont"
)

var (
	Font1 = monofont.New1()
	Font2 = monofont.Scale(Font1, 2)
	Font3 = monofont.Scale(Font1, 3)

	Font1_3 = monofont.New1_3()
)
