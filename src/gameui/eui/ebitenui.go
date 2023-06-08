package eui

import (
	"image/color"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/userdevice"
	"golang.org/x/image/font"
)

var (
	NormalTextColor    = ge.RGB(0x9dd793)
	CaretColor         = ge.RGB(0xe7c34b)
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

	mobile bool
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

func (o *SceneObject) Init(scene *ge.Scene) {
	// o.ui.DisableDefaultFocus = true
}

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

func NewGridContainer(columns int, opts ...widget.GridLayoutOpt) *widget.Container {
	containerOpts := []widget.GridLayoutOpt{
		widget.GridLayoutOpts.Columns(columns),
	}
	containerOpts = append(containerOpts, opts...)
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(containerOpts...)))
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

func NewRowLayoutContainerWithMinWidth(minWidth, spacing int, rowscale []bool) *widget.Container {
	return widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
			StretchVertical:   true,
		})),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.MinSize(minWidth, 0)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, rowscale),
			widget.GridLayoutOpts.Spacing(spacing, spacing),
		)),
	)
}

func NewRowLayoutContainer(spacing int, rowscale []bool) *widget.Container {
	return NewRowLayoutContainerWithMinWidth(0, spacing, rowscale)
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

func NewCenteredLabel(text string, ff font.Face) *widget.Text {
	return widget.NewText(
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
		widget.TextOpts.Text(text, ff, NormalTextColor),
	)
}

func NewColoredLabel(text string, ff font.Face, clr color.RGBA, options ...widget.TextOpt) *widget.Text {
	opts := []widget.TextOpt{
		widget.TextOpts.Text(text, ff, clr),
	}
	if len(options) != 0 {
		opts = append(opts, options...)
	}
	return widget.NewText(opts...)
}

func NewLabel(text string, ff font.Face, options ...widget.TextOpt) *widget.Text {
	return NewColoredLabel(text, ff, NormalTextColor, options...)
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

func NewItemButton(res *Resources, img *ebiten.Image, ff font.Face, label string, labelOffset int, onclick func()) *ItemButton {
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
					Top: labelOffset,
				})),
			),
		)
		labelWidget := widget.NewText(
			widget.TextOpts.Text(label, ff, res.Button.TextColors.Idle),
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

func NewSmallButton(res *Resources, scene *ge.Scene, text string, onclick func()) *widget.Button {
	return newButton(res, assets.BitmapFont1, scene, text, onclick)
}

func NewButton(res *Resources, scene *ge.Scene, text string, onclick func()) *widget.Button {
	return newButton(res, res.Button.FontFace, scene, text, onclick)
}

func newButton(res *Resources, ff font.Face, scene *ge.Scene, text string, onclick func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.Button.Image),
		widget.ButtonOpts.Text(text, ff, res.Button.TextColors),
		widget.ButtonOpts.TextPadding(res.Button.Padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			scene.Audio().PlaySound(assets.AudioClick)
			onclick()
		}),
	)
}

type SelectButtonConfig struct {
	Resources *Resources
	Input     gameinput.Handler
	Scene     *ge.Scene // If press sound is needed

	Value          *int
	Label          string
	ValueNames     []string
	DisabledValues []int

	OnPressed func()
	OnHover   func()
}

func NewSelectButton(config SelectButtonConfig) *widget.Button {
	maxValue := len(config.ValueNames) - 1
	value := config.Value
	key := config.Label
	valueNames := config.ValueNames

	var slider gmath.Slider
	slider.SetBounds(0, maxValue)
	slider.TrySetValue(*value)
	makeLabel := func() string {
		if key == "" {
			return valueNames[slider.Value()]
		}
		return key + ": " + valueNames[slider.Value()]
	}

	button := newButtonSelected(config.Resources, makeLabel())

	button.ClickedEvent.AddHandler(func(args interface{}) {
		increase := false
		{
			cursorPos := config.Input.AnyCursorPos()
			buttonRect := button.GetWidget().Rect
			buttonWidth := buttonRect.Dx()
			if cursorPos.X >= float64(buttonRect.Min.X)+float64(buttonWidth)*0.5 {
				increase = true
			}
		}

		for {
			if increase {
				slider.Inc()
			} else {
				slider.Dec()
			}
			*value = slider.Value()
			if !xslices.Contains(config.DisabledValues, *value) {
				break
			}
		}

		button.Text().Label = makeLabel()
		if config.Scene != nil {
			config.Scene.Audio().PlaySound(assets.AudioClick)
		}
		if config.OnPressed != nil {
			config.OnPressed()
		}
	})

	if config.OnHover != nil {
		button.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			config.OnHover()
		})
	}

	return button
}

type BoolSelectButtonConfig struct {
	Resources *Resources
	Scene     *ge.Scene // If press sound is needed

	Value      *bool
	Label      string
	ValueNames []string

	OnPressed func()
	OnHover   func()
}

func NewBoolSelectButton(config BoolSelectButtonConfig) widget.PreferredSizeLocateableWidget {
	var slider gmath.Slider
	slider.SetBounds(0, 1)
	value := config.Value
	key := config.Label
	valueNames := config.ValueNames
	if *value {
		slider.TrySetValue(1)
	}
	button := newButtonSelected(config.Resources, key+": "+valueNames[slider.Value()])

	button.ClickedEvent.AddHandler(func(args interface{}) {
		slider.Inc()
		*value = slider.Value() != 0
		button.Text().Label = key + ": " + valueNames[slider.Value()]
		if config.Scene != nil {
			config.Scene.Audio().PlaySound(assets.AudioClick)
		}
		if config.OnPressed != nil {
			config.OnPressed()
		}
	})

	if config.OnHover != nil {
		button.GetWidget().CursorEnterEvent.AddHandler(func(args interface{}) {
			config.OnHover()
		})
	}

	return button
}

func newButtonSelected(res *Resources, text string) *widget.Button {
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

func NewTextInput(res *Resources, ff font.Face, opts ...widget.TextInputOpt) *widget.TextInput {
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
		// widget.TextInputOpts.SubmitHandler(func(args *widget.TextInputChangedEventArgs) {
		// 	f(args.InputText)
		// }),
	}
	options = append(options, opts...)
	t := widget.NewTextInput(options...)
	return t
}

func LoadResources(device userdevice.Info, loader *resource.Loader) *Resources {
	result := &Resources{
		mobile: device.IsMobile,
	}

	newNineSlice := func(img *ebiten.Image, centerWidth, centerHeight int) *image.NineSlice {
		w, h := img.Size()
		return image.NewNineSlice(img,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight})
	}

	{
		idle := loader.LoadImage(assets.ImageUITextInputIdle).Data
		disabled := loader.LoadImage(assets.ImageUITextInputIdle).Data
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
			FontFace: assets.BitmapFont2,
			TextColors: &widget.TextInputColor{
				Idle:          NormalTextColor,
				Disabled:      NormalTextColor,
				Caret:         CaretColor,
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
			Color:    NormalTextColor,
			AltColor: ge.RGB(0x000000),
			FontFace: assets.BitmapFont2,
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
			Color:    NormalTextColor,
			AltColor: ge.RGB(0x000000),
			FontFace: assets.BitmapFont2,
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
			Idle:     NormalTextColor,
			Disabled: ge.RGB(0x5a7a91),
		}
		result.Button = &ButtonResource{
			Image: &widget.ButtonImage{
				Idle:     idle,
				Hover:    hover,
				Pressed:  pressed,
				Disabled: disabled,
			},
			Padding:    buttonPadding,
			TextColors: buttonColors,
			FontFace:   assets.BitmapFont2,
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
			FontFace:   assets.BitmapFont2,
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
			FontFace:   assets.BitmapFont2,
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
