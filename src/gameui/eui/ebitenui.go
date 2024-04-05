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
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
	"github.com/quasilyte/roboden-game/steamsdk"
	"github.com/quasilyte/roboden-game/userdevice"
	"golang.org/x/image/font"
)

var (
	NormalTextColor    = ge.RGB(0x9dd793)
	CaretColor         = ge.RGB(0xe7c34b)
	disabledCaretColor = ge.RGB(0x766326)
)

type Widget = widget.PreferredSizeLocateableWidget

type Resources struct {
	Button         *ButtonResource
	TabButton      *ButtonResource
	ItemButton     *ToggleButtonResource
	BigItemButton  *ToggleButtonResource
	TextInput      *TextInputResource
	ButtonSelected *ButtonResource
	Panel          *PanelResource
	DarkPanel      *PanelResource

	Font1 *font.Face
	Font2 *font.Face
	Font3 *font.Face

	mobile bool
}

type TextInputResource struct {
	Image      *widget.TextInputImage
	Padding    widget.Insets
	TextColors *widget.TextInputColor
}

type PanelResource struct {
	Image   *image.NineSlice
	Padding widget.Insets
}

type ButtonResource struct {
	Image      *widget.ButtonImage
	Padding    widget.Insets
	TextColors *widget.ButtonTextColor
}

type ToggleButtonResource struct {
	Image    *widget.ButtonImage
	AltImage *widget.ButtonImage
	Padding  widget.Insets
	Color    color.Color
	AltColor color.Color
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
	o := &SceneObject{
		ui: &ebitenui.UI{
			Container: root,
		},
	}
	o.ui.DisableDefaultFocus = true
	return o
}

func (o *SceneObject) Unfocus() {
	focuser := o.ui.GetFocusedWidget()
	if focuser != nil {
		focuser.Focus(false)
	}
}

func (o *SceneObject) GetFocused() Widget {
	focuser := o.ui.GetFocusedWidget()
	if w, ok := focuser.(Widget); ok {
		return w
	}
	return nil
}

func (o *SceneObject) AddWindow(w *widget.Window) {
	o.ui.AddWindow(w)
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

func DebugContainerColor() widget.ContainerOpt {
	return widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{255, 0, 0, 255}))
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

func NewTransparentSeparator() widget.PreferredSizeLocateableWidget {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    4,
				Bottom: 4,
			}))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(
			widget.RowLayoutData{Stretch: true},
		)))

	c.AddChild(widget.NewGraphic(
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch:   true,
			MaxHeight: 1,
		})),
		widget.GraphicOpts.ImageNineSlice(image.NewNineSliceColor(color.RGBA{})),
	))

	return c
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
		Button: b,
		res:    res.BigItemButton,
	}
}

type RecipeView struct {
	Container *widget.Container
	Icon1     *widget.Graphic
	Icon2     *widget.Graphic
	Separator *widget.Text
}

func NewRecipeView(res *Resources) *RecipeView {
	iconsContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top: 14,
			}),
		)),
	)

	icon1 := widget.NewGraphic()
	iconsContainer.AddChild(icon1)

	separator := widget.NewText(
		widget.TextOpts.Text("", *res.Font2, res.Button.TextColors.Idle),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)
	iconsContainer.AddChild(separator)

	icon2 := widget.NewGraphic()
	iconsContainer.AddChild(icon2)

	n := &RecipeView{
		Container: iconsContainer,
		Icon1:     icon1,
		Icon2:     icon2,
		Separator: separator,
	}
	return n
}

func (r *RecipeView) SetImages(a, b *ebiten.Image) {
	r.Icon1.Image = a
	r.Icon2.Image = b
	if a == nil && b == nil {
		r.Container.GetWidget().Visibility = widget.Visibility_Hide
		r.Separator.Label = ""
	} else {
		r.Container.GetWidget().Visibility = widget.Visibility_Show
		r.Separator.Label = "+"
	}
}

type ItemButton struct {
	Widget widget.PreferredSizeLocateableWidget
	Button *widget.Button
	label  *widget.Text
	state  bool
	res    *ToggleButtonResource
}

func (b *ItemButton) IsToggled() bool {
	return b.state
}

func (b *ItemButton) SetDisabled(disabled bool) {
	b.Button.GetWidget().Disabled = disabled
}

func (b *ItemButton) SetToggled(state bool) {
	if b.state == state {
		return
	}
	b.Toggle()
}

func (b *ItemButton) Toggle() {
	b.state = !b.state
	if b.state {
		b.Button.Image = b.res.AltImage
		if b.label != nil {
			b.label.Color = b.res.AltColor
		}
	} else {
		b.Button.Image = b.res.Image
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
		Button: b,
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

type ButtonConfig struct {
	Scene *ge.Scene

	Font font.Face

	Text string

	OnPressed func()
	OnHover   func()
}

func NewSmallButton(res *Resources, scene *ge.Scene, text string, onclick func()) *widget.Button {
	return NewButtonWithConfig(res, ButtonConfig{
		Font:      *res.Font1,
		Scene:     scene,
		Text:      text,
		OnPressed: onclick,
	})
}

func NewButton(res *Resources, scene *ge.Scene, text string, onclick func()) *widget.Button {
	return NewButtonWithConfig(res, ButtonConfig{
		Font:      *res.Font2,
		Scene:     scene,
		Text:      text,
		OnPressed: onclick,
	})
}

func NewButtonWithConfig(res *Resources, config ButtonConfig) *widget.Button {
	ff := config.Font
	if ff == nil {
		ff = *res.Font2
	}

	options := []widget.ButtonOpt{
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.Button.Image),
		widget.ButtonOpts.Text(config.Text, ff, res.Button.TextColors),
		widget.ButtonOpts.TextPadding(res.Button.Padding),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if config.Scene != nil {
				config.Scene.Audio().PlaySound(assets.AudioClick)
			}
			if config.OnPressed != nil {
				config.OnPressed()
			}
		}),
	}

	if config.OnHover != nil {
		options = append(options, widget.ButtonOpts.CursorEnteredHandler(func(args *widget.ButtonHoverEventArgs) {
			config.OnHover()
		}))
	}

	b := widget.NewButton(options...)
	return b
}

type SelectButtonConfig struct {
	Resources *Resources
	Input     *gameinput.Handler

	Value          *int
	BoolValue      *bool
	Label          string
	ValueNames     []string
	DisabledValues []int

	PlaySound bool

	LayoutData any

	OnPressed func()
	OnHover   func()
}

type SelectButton struct {
	Widget         *widget.Button
	input          *gameinput.Handler
	slider         gmath.Slider
	value          *int
	boolValue      *bool
	disabledValues []int
	key            string
	valueNames     []string
	scene          *ge.Scene
	onPressed      func()
	playSound      bool
}

func (b *SelectButton) Init(scene *ge.Scene) {
	b.scene = scene
}

func (b *SelectButton) IsDisposed() bool { return false }

func (b *SelectButton) Update(delta float64) {
	if !b.Widget.IsFocused() {
		return
	}

	if b.input.ActionIsJustPressed(controls.ActionMenuFocusLeft) {
		b.ChangeValue(false)
		b.input.MarkConsumed(controls.ActionMenuFocusLeft)
	}
	if b.input.ActionIsJustPressed(controls.ActionMenuFocusRight) {
		b.ChangeValue(true)
		b.input.MarkConsumed(controls.ActionMenuFocusRight)
	}
}

func (b *SelectButton) ChangeValue(increase bool) {
	for {
		if increase {
			b.slider.Inc()
		} else {
			b.slider.Dec()
		}

		if b.boolValue != nil {
			*b.boolValue = b.slider.Value() != 0
			break
		}

		// Non-bool values path.
		*b.value = b.slider.Value()
		if !xslices.Contains(b.disabledValues, *b.value) {
			break
		}
	}

	b.Widget.Text().Label = b.makeLabel()
	if b.playSound {
		b.scene.Audio().PlaySound(assets.AudioClick)
	}
	if b.onPressed != nil {
		b.onPressed()
	}
}

func (b *SelectButton) makeLabel() string {
	if b.key == "" {
		return b.valueNames[b.slider.Value()]
	}
	return b.key + ": " + b.valueNames[b.slider.Value()]
}

func NewSelectButton(config SelectButtonConfig) *SelectButton {
	if config.Input == nil {
		panic("nil input")
	}

	b := &SelectButton{
		value:          config.Value,
		boolValue:      config.BoolValue,
		disabledValues: config.DisabledValues,
		key:            config.Label,
		valueNames:     config.ValueNames,
		playSound:      config.PlaySound,
		onPressed:      config.OnPressed,
		input:          config.Input,
	}

	maxValue := len(config.ValueNames) - 1
	if config.BoolValue != nil {
		maxValue = 1
	}

	b.slider.SetBounds(0, maxValue)
	if config.BoolValue != nil {
		if *b.boolValue {
			b.slider.TrySetValue(1)
		} else {
			b.slider.TrySetValue(0)
		}
	} else {
		b.slider.TrySetValue(*b.value)
	}

	buttonOpts := []widget.ButtonOpt{}
	if config.LayoutData != nil {
		buttonOpts = append(buttonOpts, widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(config.LayoutData)))
	}
	button := newButtonSelected(config.Resources, b.makeLabel(), buttonOpts...)
	b.Widget = button

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
		b.ChangeValue(increase)
	})

	if config.OnHover != nil {
		button.CursorEnteredEvent.AddHandler(func(args interface{}) {
			config.OnHover()
		})
	}

	return b
}

func newButtonSelected(res *Resources, text string, opts ...widget.ButtonOpt) *widget.Button {
	options := []widget.ButtonOpt{
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ButtonOpts.Image(res.ButtonSelected.Image),
		widget.ButtonOpts.Text(text, *res.Font2, res.ButtonSelected.TextColors),
		widget.ButtonOpts.TextPadding(res.ButtonSelected.Padding),
	}
	options = append(options, opts...)
	return widget.NewButton(options...)
}

func NewTextPanel(res *Resources, minWidth, minHeight int) *widget.Container {
	return NewDarkPanel(res, minWidth, minHeight,
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    16,
				Bottom: 16,
				Left:   20,
				Right:  20,
			}),
		)))
}

func NewDarkPanel(res *Resources, minWidth, minHeight int, opts ...widget.ContainerOpt) *widget.Container {
	return newPanel(res, minWidth, minHeight, true, opts...)
}

func NewPanel(res *Resources, minWidth, minHeight int, opts ...widget.ContainerOpt) *widget.Container {
	return newPanel(res, minWidth, minHeight, false, opts...)
}

func newPanel(res *Resources, minWidth, minHeight int, dark bool, opts ...widget.ContainerOpt) *widget.Container {
	panelRes := res.Panel
	if dark {
		panelRes = res.DarkPanel
	}
	options := []widget.ContainerOpt{
		widget.ContainerOpts.BackgroundImage(panelRes.Image),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(4),
			widget.RowLayoutOpts.Padding(panelRes.Padding),
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
	}
	options = append(options, opts...)

	return widget.NewContainer(options...)
}

type TextInputConfig struct {
	SteamDeck bool
}

func NewTextInput(res *Resources, config TextInputConfig, opts ...widget.TextInputOpt) *widget.TextInput {
	options := []widget.TextInputOpt{
		// widget.TextInputOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
		// 	Stretch: true,
		// })),
		widget.TextInputOpts.Image(res.TextInput.Image),
		widget.TextInputOpts.Color(res.TextInput.TextColors),
		widget.TextInputOpts.Padding(res.TextInput.Padding),
		widget.TextInputOpts.Face(*res.Font1),
		widget.TextInputOpts.CaretOpts(
			widget.CaretOpts.Size(*res.Font1, 2),
		),
		widget.TextInputOpts.AllowDuplicateSubmit(true),
	}
	options = append(options, opts...)
	t := widget.NewTextInput(options...)

	if config.SteamDeck {
		t.GetWidget().FocusEvent.AddHandler(func(args any) {
			focusEvent := args.(*widget.WidgetFocusEventArgs)
			if !focusEvent.Focused {
				return
			}
			_ = steamsdk.ShowSteamDeckKeyboard(WidgetRect(t.GetWidget()))
		})
	}

	return t
}

func WidgetRect(w *widget.Widget) gmath.Rect {
	rect := w.Rect
	return gmath.Rect{
		Min: gmath.Vec{X: float64(rect.Min.X), Y: float64(rect.Min.Y)},
		Max: gmath.Vec{X: float64(rect.Max.X), Y: float64(rect.Max.Y)},
	}
}

func LoadResources(dst *Resources, device userdevice.Info, loader *resource.Loader) *Resources {
	dst.mobile = device.IsMobile()
	result := dst

	{
		idle := loader.LoadImage(assets.ImageUITextInputIdle).Data
		disabled := loader.LoadImage(assets.ImageUITextInputIdle).Data
		result.TextInput = &TextInputResource{
			Image: &widget.TextInputImage{
				Idle:     image.NewNineSlice(idle, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
				Disabled: image.NewNineSlice(disabled, [3]int{9, 14, 6}, [3]int{9, 14, 6}),
			},
			Padding: widget.Insets{
				Left:   14,
				Right:  14,
				Top:    14,
				Bottom: 10,
			},
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
			Image: nineSliceImage(idle, 10, 10),
			Padding: widget.Insets{
				Left:   16,
				Right:  16,
				Top:    10,
				Bottom: 10,
			},
		}
	}

	{
		idle := loader.LoadImage(assets.ImageUIPanelIdleDark).Data
		result.DarkPanel = &PanelResource{
			Image: nineSliceImage(idle, 10, 10),
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

func AddBackground(img *ebiten.Image, scene *ge.Scene) {
	if img == nil {
		return
	}
	s := ge.NewSprite(scene.Context())
	s.Centered = false
	s.SetColorScale(ge.ColorScale{R: 0.35, G: 0.35, B: 0.35, A: 1})
	s.SetImage(resource.Image{Data: img})
	scene.AddGraphics(s)
}
