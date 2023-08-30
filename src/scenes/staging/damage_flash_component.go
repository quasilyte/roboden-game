package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type damageFlashComponent struct {
	sprite *ge.Sprite
	flash  float64
}

func (c *damageFlashComponent) ChangeSprite(s *ge.Sprite) {
	c.resetColors()
	c.sprite = s
}

func (c *damageFlashComponent) resetColors() {
	c.flash = 0
	c.sprite.SetColorScale(ge.ColorScale{
		R: 1,
		G: 1,
		B: 1,
		A: c.sprite.GetAlpha(),
	})
}

func (c *damageFlashComponent) SetFlash(t float64) {
	c.flash = t
	c.sprite.SetColorScale(ge.ColorScale{
		R: 1 + 0.4,
		G: 1 + 0.4,
		B: 1 + 0.4,
		A: c.sprite.GetAlpha(),
	})
}

func (c *damageFlashComponent) Update(delta float64) {
	if c.flash == 0 {
		return
	}

	c.flash = gmath.ClampMin(c.flash-delta, 0)
	if c.flash == 0 {
		c.resetColors()
	}
}
