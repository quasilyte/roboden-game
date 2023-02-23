package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type damageFlashComponent struct {
	sprite *ge.Sprite
	flash  float64
}

func (c *damageFlashComponent) Update(delta float64) {
	if c.flash == 0 {
		return
	}

	c.flash = gmath.ClampMin(c.flash-delta, 0)
	if c.flash == 0 {
		c.sprite.SetColorScale(defaultColorScale)
	} else {
		x := float32(c.flash * 2)
		c.sprite.SetColorScale(ge.ColorScale{
			R: 1 + x,
			G: 1 + x,
			B: 1 + x,
			A: 1,
		})
	}
}
