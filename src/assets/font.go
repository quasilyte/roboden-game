package assets

import (
	"github.com/quasilyte/bitsweetfont"
)

var (
	Font1 = bitsweetfont.New1()
	Font2 = bitsweetfont.Scale(Font1, 2)
	Font3 = bitsweetfont.Scale(Font1, 3)

	Font1_3 = bitsweetfont.New1_3()
)
