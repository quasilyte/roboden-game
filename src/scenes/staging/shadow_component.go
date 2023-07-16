package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type shadowComponent struct {
	sprite *ge.Sprite
	pos    gmath.Vec
	height float64
	offset float64
}

func (shadow *shadowComponent) SetVisibility(visible bool) {
	if shadow.sprite != nil {
		shadow.sprite.Visible = visible
	}
}

func (shadow *shadowComponent) Dispose() {
	if shadow.sprite != nil {
		shadow.sprite.Dispose()
	}
}

func (shadow *shadowComponent) GetImageID() resource.ImageID {
	if shadow.sprite == nil {
		return assets.ImageNone
	}
	return shadow.sprite.ImageID()
}

func (shadow *shadowComponent) Init(world *worldState, imageID resource.ImageID) {
	shadow.sprite = world.rootScene.NewSprite(imageID)
	shadow.sprite.Pos.Base = &shadow.pos
	shadow.sprite.Visible = false
	world.stage.AddSprite(shadow.sprite)
}

func (shadow *shadowComponent) UpdatePos(objectPos gmath.Vec) {
	if shadow.sprite != nil {
		shadow.pos = objectPos.Add(gmath.Vec{Y: shadow.height + shadow.offset})
	}
}

func (shadow *shadowComponent) UpdateHeight(objectPos gmath.Vec, newHeight, maxHeight float64) {
	shadow.height = newHeight

	if shadow.sprite != nil {
		shadow.pos.Y = objectPos.Y + newHeight + shadow.offset
		newShadowAlpha := float32(1.0 - ((newHeight / maxHeight) * 0.5))
		shadow.sprite.SetAlpha(newShadowAlpha)
	}
}
