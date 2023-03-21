package eui

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/roboden-game/assets"
	"golang.org/x/image/font"
)

var (
	normalTextColor    = ge.RGB(0x9dd793)
	caretColor         = ge.RGB(0xe7c34b)
	disabledCaretColor = ge.RGB(0x766326)
)

type Resources struct {
	Button         *ButtonResource
	TabButton      *ButtonResource
	ItemButton     *ToggleButtonResource
	BigItemButton  *ToggleButtonResource
	TextInput      *TextInputResource
	ButtonSelected *ButtonResource
	Panel          *PanelResource
}

type TextInputResource struct {
	Image      *widget.TextInputImage
	Padding    widget.Insets
	TextColors *widget.TextInputColor
	FontFace   font.Face
}

type PanelResource struct {
	Image   *image.NineSlice
	Padding widget.Insets
}

type ButtonResource struct {
	Image      *widget.ButtonImage
	Padding    widget.Insets
	TextColors *widget.ButtonTextColor
	FontFace   font.Face
}

type ToggleButtonResource struct {
	Image    *widget.ButtonImage
	AltImage *widget.ButtonImage
	Padding  widget.Insets
	Color    color.Color
	AltColor color.Color
	FontFace font.Face
}

type OptionButtonResource struct {
	Image      *widget.ButtonImage
	Padding    widget.Insets
	TextColors *widget.ButtonTextColor
	FontFace   font.Face
	Arrow      *widget.ButtonImageImage
}

type ListResources struct {
	Image        *widget.ScrollContainerImage
	Track        *widget.SliderTrackImage
	TrackPadding widget.Insets
	Handle       *widget.ButtonImage
	HandleSize   int
	FontFace     font.Face
	Entry        *widget.ListEntryColor
	EntryPadding widget.Insets
}

type SceneObject struct {
	ui *ebitenui.UI
}

func NewSceneObject(root *widget.Container) *SceneObject {
	return &SceneObject{
		ui: &ebitenui.UI{
			Container: root,
		},
	}
}

func (o *SceneObject) IsDisposed() bool { return false }

func (o *SceneObject) Init(scene *ge.Scene) {}

func (o *SceneObject) Update(delta float64) {
	o.ui.Update()
}

func (o *SceneObject) Draw(screen *ebiten.Image) {
	o.ui.Draw(screen)
}

func NewAnchorContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
		})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
}

func NewPageContentContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)))
}

func NewHorizontalContainer() *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
		})),
		widget.ContainerOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(10),
			widget.RowLayoutOpts.Padding(widget.Insets{Left: 32, Right: 32, Top: 32}),
		)))
}

func NewRowLayoutContainer(spacing int, rowscale []bool) *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
		widget.ContainerOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			// widget.WidgetOpts.MinSize(256, 32),
		),
		// widget.ContainerOpts.Layout(widget.NewRowLayout(
		// 	widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		// 	widget.RowLayoutOpts.Spacing(10),
		// )),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, rowscale),
			widget.GridLayoutOpts.Spacing(spacing, spacing),
		)),
	)
}

func NewSeparator(ld interface{}) widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(ld)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 2,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(ge.RGB(0x2a3944))),
	))

	return c
}

func NewCenteredLabel(res *Resources, text string, ff font.Face) *widget.Text {
	return widget.NewText(
		// widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
		// 	widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
		// 		// Position:           widget.RowLayoutPositionCenter,
		// 		HorizontalPosition: widget.AnchorLayoutPositionCenter,
		// 		VerticalPosition:   widget.AnchorLayoutPositionCenter,
		// 	}),
		// )),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text(text, ff, res.Button.TextColors.Idle),
	)
}

func NewLabel(res *Resources, text string, ff font.Face) *widget.Text {
	return widget.NewText(
		// widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(
		// 	widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
		// 		// Position:           widget.RowLayoutPositionCenter,
		// 		HorizontalPosition: widget.AnchorLayoutPositionCenter,
		// 		VerticalPosition:   widget.AnchorLayoutPositionCenter,
		// 	}),
		// )),
		widget.TextOpts.Text(text, ff, res.Button.TextColors.Idle),
	)
}

func NewBigItemButton(res *Resources, img *ebiten.Image, onclick func()) *ItemButton {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(98, 80),
		),
	)

	b := widget.NewButton(
		widget.ButtonOpts.Image(res.BigItemButton.Image),
		widget.ButtonOpts.TextPadding(res.BigItemButton.Padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			onclick()
		}),
	)
	container.AddChild(b)

	container.AddChild(widget.NewGraphic(
		widget.GraphicOpts.Image(img),
	))

	return &ItemButton{
		Widget: container,
		button: b,
		res:    res.BigItemButton,
	}
}

type ItemButton struct {
	Widget widget.PreferredSizeLocateableWidget
	button *widget.Button
	label  *widget.Text
	state  bool
	res    *ToggleButtonResource
}

func (b *ItemButton) IsToggled() bool {
	return b.state
}

func (b *ItemButton) SetDisabled(disabled bool) {
	b.button.GetWidget().Disabled = disabled
}

func (b *ItemButton) Toggle() {
	b.state = !b.state
	if b.state {
		b.button.Image = b.res.AltImage
		if b.label != nil {
			b.label.Color = b.res.AltColor
		}
	} else {
		b.button.Image = b.res.Image
		if b.label != nil {
			b.label.Color = b.res.Color
		}
	}
}

func NewItemButton(res *Resources, img *ebiten.Image, label string, onclick func()) *ItemButton {
	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(64, 32),
		),
	)

	b := widget.NewButton(
		widget.ButtonOpts.Image(res.ItemButton.Image),
		widget.ButtonOpts.TextPadding(res.ItemButton.Padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			onclick()
		}),
	)
	container.AddChild(b)

	if img != nil {
		container.AddChild(widget.NewGraphic(
			widget.GraphicOpts.Image(img),
		))
	}

	result := &ItemButton{
		Widget: container,
		button: b,
		res:    res.ItemButton,
	}

	if label != "" {
		paddingContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(
				widget.NewAnchorLayout(widget.AnchorLayoutOpts.Padding(widget.Insets{
					Top: 22,
				})),
			),
		)
		labelWidget := widget.NewText(
			widget.TextOpts.Text(label, res.ItemButton.FontFace, res.Button.TextColors.Idle),
			widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
		paddingContainer.AddChild(labelWidget)
		container.AddChild(paddingContainer)
		result.label = labelWidget
	}

	return result
}

func NewButton(res *Resources, scene *ge.Scene, text string, onclick func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.Button.Image),
		widget.ButtonOpts.Text(text, res.Button.FontFace, res.Button.TextColors),
		widget.ButtonOpts.TextPadding(res.Button.Padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			scene.Audio().PlaySound(assets.AudioClick)
			onclick()
		}),
	)
}

func NewButtonSelected(res *Resources, text string) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.ButtonSelected.Image),
		widget.ButtonOpts.Text(text, res.ButtonSelected.FontFace, res.ButtonSelected.TextColors),
		widget.ButtonOpts.TextPadding(res.ButtonSelected.Padding),
	)
}

func NewPanel(res *Resources, minWidth, minHeight int) *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(res.Panel.Image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(res.Panel.Padding),
		)),
		widget.ContainerOpts.WidgetOpts(
			// instruct the container's anchor layout to center the button both horizontally and vertically
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal:  true,
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
			widget.WidgetOpts.MinSize(minWidth, minHeight),
		),
	)
}

func NewTextInput(res *Resources, ff font.Face, f func(s string)) *widget.TextInput {
	options := []widget.TextInputOpt{
		// widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		// 	Stretch: true,
		// })),
		widget.TextInputOpts.Image(res.TextInput.Image),
		widget.TextInputOpts.Color(res.TextInput.TextColors),
		widget.TextInputOpts.Padding(res.TextInput.Padding),
		widget.TextInputOpts.Face(res.TextInput.FontFace),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(res.TextInput.FontFace, 2),
		),
		widget.TextInputOpts.AllowDuplicateSubmit(true),
		widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
			f(args.InputText)
		}),
	}
	t := widget.NewTextInput(options...)
	return t
}

func LoadResources(loader *resource.Loader) *Resources {
	result := &Resources{}

	newNineSlice := func(img *ebiten.Image, centerWidth, centerHeight int) *image.NineSlice {
		w, h := img.Size()
		return image.NewNineSlice(img,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight})
	}

	{
		idle := loader.LoadImage(assets.ImageUITextInputIdle).Data
		disabled := loader.LoadImage(assets.ImageUITextInputIdle).Data
		ff := loader.LoadFont(assets.FontSmall).Face
		result.TextInput = &TextInputResource{
			Image: &widget.TextInputImage{
				Idle:     image.NewNineSlice(idle, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
				Disabled: image.NewNineSlice(disabled, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
			},
			Padding: widget.Insets{
				Left:   8,
				Right:  8,
				Top:    4,
				Bottom: 4,
			},
			FontFace: ff,
			TextColors: &widget.TextInputColor{
				Idle:          normalTextColor,
				Disabled:      normalTextColor,
				Caret:         caretColor,
				DisabledCaret: disabledCaretColor,
			},
		}
	}

	{
		idle := loader.LoadImage(assets.ImageUIPanelIdle).Data
		result.Panel = &PanelResource{
			Image: newNineSlice(idle, 10, 10),
			Padding: widget.Insets{
				Left:   16,
				Right:  16,
				Top:    10,
				Bottom: 10,
			},
		}
	}

	{
		idle := nineSliceImage(loader.LoadImage(assets.ImageUIBigItemButtonIdle).Data, 12, 0)
		hover := nineSliceImage(loader.LoadImage(assets.ImageUIBigItemButtonHover).Data, 12, 0)
		pressed := nineSliceImage(loader.LoadImage(assets.ImageUIBigItemButtonPressed).Data, 12, 0)
		disabled := nineSliceImage(loader.LoadImage(assets.ImageUIBigItemButtonDisabled).Data, 12, 0)
		altIdle := nineSliceImage(loader.LoadImage(assets.ImageUIAltBigItemButtonIdle).Data, 12, 0)
		altHover := nineSliceImage(loader.LoadImage(assets.ImageUIAltBigItemButtonHover).Data, 12, 0)
		altPressed := nineSliceImage(loader.LoadImage(assets.ImageUIAltBigItemButtonPressed).Data, 12, 0)
		altDisabled := nineSliceImage(loader.LoadImage(assets.ImageUIAltBigItemButtonDisabled).Data, 12, 0)
		buttonPadding := widget.Insets{
			Left:  30,
			Right: 30,
		}
		ff := loader.LoadFont(assets.FontSmall).Face
		result.BigItemButton = &ToggleButtonResource{
			Image: &widget.ButtonImage{
				Idle:     idle,
				Hover:    hover,
				Pressed:  pressed,
				Disabled: disabled,
			},
			AltImage: &widget.ButtonImage{
				Idle:     altIdle,
				Hover:    altHover,
				Pressed:  altPressed,
				Disabled: altDisabled,
			},
			Padding:  buttonPadding,
			Color:    normalTextColor,
			AltColor: ge.RGB(0x000000),
			FontFace: ff,
		}
	}

	{
		idle := nineSliceImage(loader.LoadImage(assets.ImageUIItemButtonIdle).Data, 12, 0)
		hover := nineSliceImage(loader.LoadImage(assets.ImageUIItemButtonHover).Data, 12, 0)
		pressed := nineSliceImage(loader.LoadImage(assets.ImageUIItemButtonPressed).Data, 12, 0)
		disabled := nineSliceImage(loader.LoadImage(assets.ImageUIItemButtonDisabled).Data, 12, 0)
		altIdle := nineSliceImage(loader.LoadImage(assets.ImageUIAltItemButtonIdle).Data, 12, 0)
		altHover := nineSliceImage(loader.LoadImage(assets.ImageUIAltItemButtonHover).Data, 12, 0)
		altPressed := nineSliceImage(loader.LoadImage(assets.ImageUIAltItemButtonPressed).Data, 12, 0)
		altDisabled := nineSliceImage(loader.LoadImage(assets.ImageUIAltItemButtonDisabled).Data, 12, 0)
		buttonPadding := widget.Insets{
			Left:  30,
			Right: 30,
		}
		ff := loader.LoadFont(assets.FontNormal).Face
		result.ItemButton = &ToggleButtonResource{
			Image: &widget.ButtonImage{
				Idle:     idle,
				Hover:    hover,
				Pressed:  pressed,
				Disabled: disabled,
			},
			AltImage: &widget.ButtonImage{
				Idle:     altIdle,
				Hover:    altHover,
				Pressed:  altPressed,
				Disabled: altDisabled,
			},
			Padding:  buttonPadding,
			Color:    normalTextColor,
			AltColor: ge.RGB(0x000000),
			FontFace: ff,
		}
	}

	{
		idle := nineSliceImage(loader.LoadImage(assets.ImageUIButtonIdle).Data, 12, 0)
		hover := nineSliceImage(loader.LoadImage(assets.ImageUIButtonHover).Data, 12, 0)
		pressed := nineSliceImage(loader.LoadImage(assets.ImageUIButtonPressed).Data, 12, 0)
		disabled := nineSliceImage(loader.LoadImage(assets.ImageUIButtonDisabled).Data, 12, 0)
		selectedIdle := nineSliceImage(loader.LoadImage(assets.ImageUIButtonSelectedIdle).Data, 12, 0)
		selectedHover := nineSliceImage(loader.LoadImage(assets.ImageUIButtonSelectedHover).Data, 12, 0)
		selectedPressed := nineSliceImage(loader.LoadImage(assets.ImageUIButtonSelectedPressed).Data, 12, 0)
		selectedDisabled := nineSliceImage(loader.LoadImage(assets.ImageUIButtonSelectedDisabled).Data, 12, 0)
		tabIdle := nineSliceImage(loader.LoadImage(assets.ImageUITabButtonIdle).Data, 12, 0)
		tabHover := nineSliceImage(loader.LoadImage(assets.ImageUITabButtonHover).Data, 12, 0)
		tabPressed := nineSliceImage(loader.LoadImage(assets.ImageUITabButtonPressed).Data, 12, 0)
		tabDisabled := nineSliceImage(loader.LoadImage(assets.ImageUITabButtonDisabled).Data, 12, 0)
		buttonPadding := widget.Insets{
			Left:  30,
			Right: 30,
		}
		buttonColors := &widget.ButtonTextColor{
			Idle:     normalTextColor,
			Disabled: ge.RGB(0x5a7a91),
		}
		ff := loader.LoadFont(assets.FontSmall).Face
		result.Button = &ButtonResource{
			Image: &widget.ButtonImage{
				Idle:     idle,
				Hover:    hover,
				Pressed:  pressed,
				Disabled: disabled,
			},
			Padding:    buttonPadding,
			TextColors: buttonColors,
			FontFace:   ff,
		}
		result.ButtonSelected = &ButtonResource{
			Image: &widget.ButtonImage{
				Idle:     selectedIdle,
				Hover:    selectedHover,
				Pressed:  selectedPressed,
				Disabled: selectedDisabled,
			},
			Padding:    buttonPadding,
			TextColors: buttonColors,
			FontFace:   ff,
		}
		result.TabButton = &ButtonResource{
			Image: &widget.ButtonImage{
				Idle:     tabIdle,
				Hover:    tabHover,
				Pressed:  tabPressed,
				Disabled: tabDisabled,
			},
			Padding:    buttonPadding,
			TextColors: buttonColors,
			FontFace:   ff,
		}
	}

	return result
}

func nineSliceImage(i *ebiten.Image, centerWidth, centerHeight int) *image.NineSlice {
	w, h := i.Size()
	return image.NewNineSlice(i,
		[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
		[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight})
}
