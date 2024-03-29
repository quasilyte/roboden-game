package menus

import (
	"time"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/controls"
	"github.com/quasilyte/roboden-game/descriptions"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/gameui"
	"github.com/quasilyte/roboden-game/gameui/eui"
	"github.com/quasilyte/roboden-game/session"
	"github.com/quasilyte/roboden-game/timeutil"
)

type SchemaMenuController struct {
	state *session.State

	mode gamedata.Mode

	slotSelectorPos gmath.Vec
	slotSelector    *ge.Rect

	saveButton       *widget.Button
	renameButton     *widget.Button
	loadButton       *widget.Button
	loadDronesButton *widget.Button

	selectedSlot int
	buttons      []*widget.Button
	schemas      [10]*gamedata.SavedSchema

	helpLabel *widget.Text

	scene *ge.Scene
}

func NewSchemaMenuController(state *session.State, mode gamedata.Mode) *SchemaMenuController {
	return &SchemaMenuController{
		state: state,
		mode:  mode,
	}
}

func (c *SchemaMenuController) Init(scene *ge.Scene) {
	c.scene = scene
	c.initUI()

	c.slotSelector = ge.NewRect(scene.Context(), 224, 45)
	c.slotSelector.Centered = false
	c.slotSelector.Pos.Base = &c.slotSelectorPos
	c.slotSelector.OutlineWidth = 1
	c.slotSelector.Visible = false
	c.slotSelector.FillColorScale.SetRGBA(0, 0, 0, 0)
	c.slotSelector.OutlineColorScale.SetColor(eui.CaretColor)
	scene.AddGraphics(c.slotSelector)

	scene.DelayedCall(0.05, func() {
		if c.slotSelector.Visible {
			return
		}
		c.selectSlot(0)
	})
}

func (c *SchemaMenuController) Update(delta float64) {
	c.state.MenuInput.Update()
	if c.state.MenuInput.ActionIsJustPressed(controls.ActionMenuBack) {
		c.back()
		return
	}
}

func (c *SchemaMenuController) initUI() {
	eui.AddBackground(c.state.BackgroundImage, c.scene)
	uiResources := c.state.Resources.UI

	root := eui.NewAnchorContainer()
	rowContainer := eui.NewRowLayoutContainer(10, nil)
	root.AddChild(rowContainer)

	d := c.scene.Dict()

	smallFont := assets.BitmapFont1

	helpLabel := eui.NewLabel("", smallFont)
	helpLabel.MaxWidth = 268
	c.helpLabel = helpLabel

	titleLabel := eui.NewCenteredLabel(d.Get("menu.main.schema"), assets.BitmapFont3)
	rowContainer.AddChild(titleLabel)

	numSlots := 10

	navTree := gameui.NewNavTree()
	navBlock := navTree.NewBlock()
	numColumns := 2
	numRows := numSlots / numColumns

	backButton := eui.NewButton(uiResources, c.scene, d.Get("menu.back"), func() {
		c.back()
	})
	backButtonElem := navBlock.NewElem(backButton)

	rootGrid := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(numColumns),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))
	leftGrid := eui.NewGridContainer(2, widget.GridLayoutOpts.Spacing(8, 4),
		widget.GridLayoutOpts.Stretch([]bool{true, false}, nil))

	var gridButtonElems []*gameui.NavElem

	for i := 0; i < numSlots; i++ {
		slotIndex := i
		key := c.state.SchemaDataKey(c.mode, i)
		var schema *gamedata.SavedSchema
		if c.state.CheckGameItem(key) {
			var s gamedata.SavedSchema
			if err := c.state.LoadGameItem(key, &s); err == nil {
				schema = &s
			}
		}
		c.schemas[slotIndex] = schema
		b := eui.NewSmallButton(uiResources, c.scene, "", func() {
			c.selectSlot(slotIndex)
		})
		c.buttons = append(c.buttons, b)
		b.CursorEnteredEvent.AddHandler(func(args interface{}) {
			c.updateHelpText(slotIndex)
		})
		b.GetWidget().MinWidth = 220
		leftGrid.AddChild(b)
		c.updateSlotLabel(slotIndex)
		gridButtonElems = append(gridButtonElems, navBlock.NewElem(b))
	}

	rightPanel := eui.NewTextPanel(uiResources, 320, 0)
	rightPanel.AddChild(helpLabel)

	rootGrid.AddChild(leftGrid)
	rootGrid.AddChild(rightPanel)

	rowContainer.AddChild(rootGrid)

	rowContainer.AddChild(eui.NewTransparentSeparator())

	buttonsGrid := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
			Stretch: true,
		})),
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, nil),
			widget.GridLayoutOpts.Spacing(4, 4))))

	c.saveButton = eui.NewButton(uiResources, c.scene, d.Get("menu.save_schema"), func() {
		schema := c.createSchema(c.selectedSlot)
		key := c.state.SchemaDataKey(c.mode, c.selectedSlot)
		c.state.SaveGameItem(key, schema)
		c.schemas[c.selectedSlot] = &schema
		c.selectSlot(c.selectedSlot) // Force a reload
		c.updateSlotLabel(c.selectedSlot)
		c.updateHelpText(c.selectedSlot)
	})
	c.saveButton.GetWidget().Disabled = true

	c.loadButton = eui.NewButton(uiResources, c.scene, d.Get("menu.load_schema"), func() {
		schema := c.schemas[c.selectedSlot]
		if schema == nil {
			return
		}
		config := c.state.GetConfigForMode(c.mode)
		config.ReplayLevelConfig = schema.Config
		c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, c.mode))
	})
	c.loadButton.GetWidget().Disabled = true

	c.renameButton = eui.NewButton(uiResources, c.scene, d.Get("menu.rename_schema"), func() {
		schema := c.schemas[c.selectedSlot]
		if schema == nil {
			return
		}
		c.scene.Context().ChangeScene(NewSchemaNameMenuController(c.state, c.mode, c.selectedSlot))
	})
	c.renameButton.GetWidget().Disabled = true

	c.loadDronesButton = eui.NewButton(uiResources, c.scene, d.Get("menu.load_schema_drones"), func() {
		schema := c.schemas[c.selectedSlot]
		if schema == nil {
			return
		}
		config := c.state.GetConfigForMode(c.mode)
		config.ReplayLevelConfig.CoreDesign = schema.Config.CoreDesign
		config.ReplayLevelConfig.TurretDesign = schema.Config.TurretDesign
		config.ReplayLevelConfig.Tier2Recipes = schema.Config.Tier2Recipes
		c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, c.mode))
	})
	c.loadDronesButton.GetWidget().Disabled = true

	buttonsGrid.AddChild(c.saveButton)
	buttonsGrid.AddChild(c.loadButton)
	buttonsGrid.AddChild(c.renameButton)
	buttonsGrid.AddChild(c.loadDronesButton)

	rowContainer.AddChild(buttonsGrid)

	rowContainer.AddChild(backButton)

	bindNavGrid(gridButtonElems, numColumns, numRows)
	saveButtonElem := navBlock.NewElem(c.saveButton)
	loadButtonElem := navBlock.NewElem(c.loadButton)
	renameButtonElem := navBlock.NewElem(c.renameButton)
	loadDronesButtonElem := navBlock.NewElem(c.loadDronesButton)
	controlButtonElems := []*gameui.NavElem{
		saveButtonElem,
		loadButtonElem,
		renameButtonElem,
		loadDronesButtonElem,
	}
	bindNavGrid(controlButtonElems, 2, 2)

	for _, b := range gridButtonElems {
		if b.Edges[gameui.NavDown] != nil {
			continue
		}
		b.Edges[gameui.NavDown] = saveButtonElem
	}
	saveButtonElem.Edges[gameui.NavUp] = gridButtonElems[len(gridButtonElems)-2]
	loadButtonElem.Edges[gameui.NavUp] = gridButtonElems[len(gridButtonElems)-1]
	renameButtonElem.Edges[gameui.NavDown] = backButtonElem
	loadDronesButtonElem.Edges[gameui.NavDown] = backButtonElem
	backButtonElem.Edges[gameui.NavUp] = renameButtonElem

	setupUI(c.scene, root, c.state.MenuInput, navTree)
}

func (c *SchemaMenuController) updateHelpText(i int) {
	d := c.scene.Dict()
	schema := c.schemas[i]
	if schema != nil {
		c.helpLabel.Label = descriptions.SchemaText(d, i, schema)
	} else {
		c.helpLabel.Label = d.Get("menu.empty_schema_slot")
	}
}

func (c *SchemaMenuController) updateSlotLabel(i int) {
	label := c.scene.Dict().Get("menu.replay.empty")
	schema := c.schemas[i]
	if schema != nil {
		if schema.Name != "" {
			label = schema.Name
		} else {
			label = timeutil.FormatDateISO8601(schema.Date, true)
		}
	}
	c.buttons[i].Text().Label = label
}

func (c *SchemaMenuController) createSchema(i int) gamedata.SavedSchema {
	name := ""
	currentSchema := c.schemas[i]
	if currentSchema != nil {
		name = currentSchema.Name
	}
	config := c.state.GetConfigForMode(c.mode)
	return gamedata.SavedSchema{
		Name:   name,
		Date:   time.Now(),
		Config: config.ReplayLevelConfig,
	}
}

func (c *SchemaMenuController) selectSlot(i int) {
	c.selectedSlot = i

	c.saveButton.GetWidget().Disabled = false

	c.slotSelector.Visible = true
	b := c.buttons[i]
	rect := b.GetWidget().Rect
	c.slotSelectorPos.X = float64(rect.Min.X) - 2
	c.slotSelectorPos.Y = float64(rect.Min.Y) - 2

	if c.schemas[i] == nil {
		c.loadButton.GetWidget().Disabled = true
		c.loadDronesButton.GetWidget().Disabled = true
		c.renameButton.GetWidget().Disabled = true
	} else {
		c.loadButton.GetWidget().Disabled = false
		c.loadDronesButton.GetWidget().Disabled = false
		c.renameButton.GetWidget().Disabled = false
	}

	// row := i / 2
	// col := i % 2
	// c.slotSelectorPos.X = float64(col)*220 + 96
	// c.slotSelectorPos.Y = float64(row) * 32
}

func (c *SchemaMenuController) back() {
	c.scene.Context().ChangeScene(NewLobbyMenuController(c.state, c.mode))
}
