package eui

import (
	"image"
	"strings"
	"unicode"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/gameinput"
)

type Keyboard struct {
	EventSubmit    gsignal.Event[gsignal.Void]
	EventBackspace gsignal.Event[gsignal.Void]
	EventKey       gsignal.Event[rune]
	EventClosed    gsignal.Event[gsignal.Void]

	Window *widget.Window

	time  float64
	input *gameinput.Handler

	rect   gmath.Rect
	closed bool

	upcase        bool
	letterButtons []*widget.Button
}

type KeyboardConfig struct {
	Resources *Resources

	Input *gameinput.Handler

	DigitsOnly bool

	Scene *ge.Scene
}

func (k *Keyboard) Close() {
	if k.closed {
		return
	}
	k.closed = true
	k.EventClosed.Emit(gsignal.Void{})
	k.Window.Close()
}

func (k *Keyboard) Init(scene *ge.Scene) {}

func (k *Keyboard) IsDisposed() bool {
	return k.closed
}

func (k *Keyboard) Update(delta float64) {
	k.time += delta
	if k.time < 0.5 {
		return
	}

	clickPos, ok := k.input.ClickPos(controls.ActionClick)
	if !ok {
		return
	}
	if !k.rect.Contains(clickPos) {
		k.Close()
	}
}

func NewTextKeyboard(config KeyboardConfig) *Keyboard {
	k := &Keyboard{
		input: config.Input,
	}

	a := NewAnchorContainer()

	width := 872
	height := 190

	pos := gmath.Vec{
		X: (config.Scene.Context().ScreenWidth - float64(width)) / 2,
		Y: (config.Scene.Context().ScreenHeight - 44) - float64(height),
	}

	panel := NewDarkPanel(config.Resources, width, height)

	topGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(6, 0))))

	rightGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(2, 2))))

	rightGrid.AddChild(NewButton(config.Resources, config.Scene, " ← ", func() {
		k.EventBackspace.Emit(gsignal.Void{})
	}))

	rightGrid.AddChild(NewButton(config.Resources, config.Scene, " ↵ ", func() {
		k.EventSubmit.Emit(gsignal.Void{})
	}))

	rightGrid.AddChild(NewButton(config.Resources, config.Scene, " ↑ ", func() {
		k.upcase = !k.upcase
		for _, b := range k.letterButtons {
			var newLabel string
			if k.upcase {
				// Lowercase -> uppercase.
				newLabel = strings.ToUpper(b.Text().Label)
			} else {
				// Uppercase -> lowercase.
				newLabel = strings.ToLower(b.Text().Label)
			}
			b.Text().Label = newLabel
		}
	}))

	leftGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(10),
			widget.GridLayoutOpts.Spacing(2, 2))))

	topGrid.AddChild(leftGrid)
	topGrid.AddChild(rightGrid)

	panel.AddChild(topGrid)
	a.AddChild(panel)

	layout := [][10]rune{
		{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'},
		{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p'},
		{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l', '.'},
		{'z', 'x', 'c', 'v', 'b', 'n', 'm', ' ', '_', '-'},
	}

	for _, row := range layout {
		for i := range row {
			ch := row[i]
			if ch == 0x0 {
				invisButton := NewButton(config.Resources, config.Scene, " ", func() {})
				invisButton.GetWidget().Visibility = widget.Visibility_Hide_Blocking
				invisButton.GetWidget().Disabled = true
				leftGrid.AddChild(invisButton)
				continue
			}
			disabled := config.DigitsOnly && !unicode.IsDigit(ch)
			label := string(ch)
			if ch == ' ' {
				label = "␣"
			}
			b := NewButton(config.Resources, config.Scene, label, func() {
				r := ch
				if unicode.IsLetter(ch) && k.upcase {
					r = unicode.ToUpper(ch)
				}
				k.EventKey.Emit(r)
			})
			leftGrid.AddChild(b)
			if unicode.IsLetter(ch) {
				k.letterButtons = append(k.letterButtons, b)
			}
			b.GetWidget().Disabled = disabled
		}
	}

	rect := gmath.Rect{
		Min: pos,
		Max: pos.Add(gmath.Vec{X: float64(width), Y: float64(height)}),
	}
	k.rect = rect

	w := widget.NewWindow(
		widget.WindowOpts.Location(image.Rect(
			int(rect.Min.X), int(rect.Min.Y),
			int(rect.Max.X), int(rect.Max.Y),
		)),
		widget.WindowOpts.Contents(a),
		widget.WindowOpts.CloseMode(widget.NONE),
		// widget.WindowOpts.Modal(),
	)

	k.Window = w

	return k
}
