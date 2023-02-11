package eui

import (
	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/colony-game/assets"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"golang.org/x/image/font"
)

type Resources struct {
	OptionButton   *OptionButtonResource
	Button         *ButtonResource
	ButtonSelected *ButtonResource
	List           *ListResources
}

type ButtonResource struct {
	Image      *widget.ButtonImage
	Padding    widget.Insets
	TextColors *widget.ButtonTextColor
	FontFace   font.Face
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

func NewOptionButton(res *Resources, entries []any) *widget.ListComboButton {
	return widget.NewListComboButton(
		widget.ListComboButtonOpts.SelectComboButtonOpts(
			widget.SelectComboButtonOpts.ComboButtonOpts(
				widget.ComboButtonOpts.ButtonOpts(
					widget.ButtonOpts.Image(res.OptionButton.Image),
					widget.ButtonOpts.TextPadding(res.OptionButton.Padding),
				),
			),
		),
		widget.ListComboButtonOpts.Text(res.OptionButton.FontFace, res.OptionButton.Arrow, res.OptionButton.TextColors),
		widget.ListComboButtonOpts.ListOpts(
			widget.ListOpts.Entries(entries),
			widget.ListOpts.ScrollContainerOpts(
				widget.ScrollContainerOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
						StretchHorizontal: true,
					}),
				),
				widget.ScrollContainerOpts.Image(res.List.Image),
			),
			widget.ListOpts.SliderOpts(
				widget.SliderOpts.Images(res.List.Track, res.List.Handle),
				widget.SliderOpts.MinHandleSize(res.List.HandleSize),
				widget.SliderOpts.TrackPadding(res.List.TrackPadding)),
			widget.ListOpts.EntryFontFace(res.List.FontFace),
			widget.ListOpts.EntryColor(res.List.Entry),
			widget.ListOpts.EntryTextPadding(res.List.EntryPadding),
		),
		widget.ListComboButtonOpts.EntryLabelFunc(func(e interface{}) string {
			return e.(string)
		}, func(e interface{}) string {
			return e.(string)
		}))
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

func NewRowLayoutContainer() *widget.Container {
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
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(10),
		)))
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

func NewLabel(res *Resources, text string, ff font.Face) *widget.Label {
	return widget.NewLabel(
		widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Position: widget.RowLayoutPositionCenter,
		}))),
		widget.LabelOpts.Text(text, ff, &widget.LabelColor{
			Idle:     res.Button.TextColors.Idle,
			Disabled: res.Button.TextColors.Disabled,
		}),
	)
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

func LoadResources(loader *resource.Loader) *Resources {
	result := &Resources{}

	{
		idle := loader.LoadImage(assets.ImageUIListIdle).Data
		disabled := loader.LoadImage(assets.ImageUIListDisabled).Data
		mask := loader.LoadImage(assets.ImageUIListMask).Data
		trackIdle := loader.LoadImage(assets.ImageUIListTrackIdle).Data
		trackDisabled := loader.LoadImage(assets.ImageUIListTrackDisabled).Data
		handleIdle := loader.LoadImage(assets.ImageUISliderHandleIdle).Data
		handleHover := loader.LoadImage(assets.ImageUISliderHandleHover).Data
		result.List = &ListResources{
			Image: &widget.ScrollContainerImage{
				Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
				Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
				Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
			},
			Track: &widget.SliderTrackImage{
				Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
				Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
				Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
			},
			TrackPadding: widget.Insets{
				Top:    5,
				Bottom: 24,
			},
			Handle: &widget.ButtonImage{
				Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
				Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
				Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
				Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
			},
			HandleSize: 5,
			FontFace:   loader.LoadFont(assets.FontSmall).Face,
			Entry: &widget.ListEntryColor{
				Unselected:                 ge.RGB(0xdff4ff),
				DisabledUnselected:         ge.RGB(0x5a7a91),
				Selected:                   ge.RGB(0xdff4ff),
				DisabledSelected:           ge.RGB(0x5a7a91),
				SelectedBackground:         ge.RGB(0x4b687a),
				DisabledSelectedBackground: ge.RGB(0x2a3944),
			},
			EntryPadding: widget.Insets{
				Left:   30,
				Right:  30,
				Top:    2,
				Bottom: 2,
			},
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
		buttonPadding := widget.Insets{
			Left:  30,
			Right: 30,
		}
		buttonColors := &widget.ButtonTextColor{
			Idle:     ge.RGB(0xdff4ff),
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
	}

	{
		arrow := &widget.ButtonImageImage{
			Idle:     loader.LoadImage(assets.ImageUIArrowDownIdle).Data,
			Disabled: loader.LoadImage(assets.ImageUIArrowDownDisabled).Data,
		}
		idle := nineSliceImage(loader.LoadImage(assets.ImageUIOptionButtonIdle).Data, 12, 0)
		hover := nineSliceImage(loader.LoadImage(assets.ImageUIOptionButtonHover).Data, 12, 0)
		pressed := nineSliceImage(loader.LoadImage(assets.ImageUIOptionButtonPressed).Data, 12, 0)
		disabled := nineSliceImage(loader.LoadImage(assets.ImageUIOptionButtonDisabled).Data, 12, 0)
		combinedImage := &widget.ButtonImage{
			Idle:     idle,
			Hover:    hover,
			Pressed:  pressed,
			Disabled: disabled,
		}
		buttonPadding := widget.Insets{
			Left:  30,
			Right: 30,
		}
		buttonColors := &widget.ButtonTextColor{
			Idle:     ge.RGB(0xdff4ff),
			Disabled: ge.RGB(0x5a7a91),
		}
		ff := loader.LoadFont(assets.FontSmall).Face
		result.OptionButton = &OptionButtonResource{
			Image:      combinedImage,
			Padding:    buttonPadding,
			TextColors: buttonColors,
			FontFace:   ff,
			Arrow:      arrow,
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
