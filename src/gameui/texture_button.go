package gameui

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gsignal"
)

type TextureButton struct {
	pos     ge.Pos
	imageID resource.ImageID
	cursor  Cursor

	sprite *ge.Sprite

	EventClicked gsignal.Event[gsignal.Void]
}

func NewTextureButton(pos ge.Pos, img resource.ImageID, cursor Cursor) *TextureButton {
	return &TextureButton{
		pos:     pos,
		imageID: img,
		cursor:  cursor,
	}
}

func (b *TextureButton) Init(scene *ge.Scene) {
	b.sprite = scene.NewSprite(b.imageID)
	b.sprite.Centered = false
	b.sprite.Pos = b.pos
	scene.AddGraphicsAbove(b.sprite, 1)
}

func (b *TextureButton) IsDisposed() bool {
	return false
}

func (b *TextureButton) SetVisibility(visible bool) {
	b.sprite.Visible = visible
}

func (b *TextureButton) HandleInput(action input.Action) bool {
	pos, ok := b.cursor.ClickPos(action)
	if !ok || !b.sprite.BoundsRect().Contains(pos) {
		return false
	}
	b.EventClicked.Emit(gsignal.Void{})
	return true
}

func (b *TextureButton) Update(delta float64) {}
