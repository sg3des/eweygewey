package fizzgui

import (
	"log"

	"github.com/go-gl/glfw/v3.1/glfw"
)

type TALIGN int

const (
	TALIGN_LEFT TALIGN = iota
	TALIGN_CENTER
	TALIGN_RIGHT
)

type STATE int

const (
	STATE_NORMAL STATE = iota
	STATE_ACTIVE
	STATE_CHECKED
)

type WidgetConstructor func(cmd *cmdList) (*RenderData, Style)
type Callback func(wgt *Widget)

type Widget struct {
	ID string

	Text      string
	TextAlign TALIGN
	font      *Font

	Texture *TextureChunk //Global widget texture

	Style       Style
	StyleHover  Style
	StyleActive Style

	State STATE

	Zorder uint8

	Layout    *Layout
	Container *Container

	OnActive   Callback
	OnKeyEnter Callback

	ConstructorData interface{}
	Constructor     WidgetConstructor
}

func (wgt *Widget) IsHover() bool {
	if HoverWidget == wgt {
		return true
	}
	return false
}

func (wgt *Widget) IsActive() bool {
	if ActiveWidget == wgt {
		if wgt.State < STATE_ACTIVE {
			wgt.State = STATE_ACTIVE
		}
		return true
	}

	return false
}

func (wgt *Widget) IsClick() (click bool, onWidget bool) {
	ma := Mouse.GetButtonAction(0)
	// ma := wgt.Window.Owner.GetMouseButtonAction(0)
	// mx, my := wgt.Window.Owner.GetMousePosition()
	if ma == MouseClick || ma == MouseDoubleClick {
		click = true
		if wgt.Layout.ContainsPoint(Mouse.X, Mouse.Y) {
			onWidget = true
		}
	}
	return
}

func (wgt *Widget) IsMouseDown() (down bool, onWidget bool) {
	ma := Mouse.GetButtonAction(0)

	if ma == MouseDown {
		down = true
		if wgt.Layout.ContainsPoint(Mouse.X, Mouse.Y) {
			onWidget = true
		}
	}

	return
}

func (wgt *Widget) draw(cursor *Cursor) (w, h float32) {
	l := wgt.Layout
	l.Update()

	var wt, ht float32
	if wgt.font != nil {
		if wgt.Text != "" {
			wt, ht, _ = wgt.font.GetRenderSize(wgt.Text)
		} else {
			wt, ht, _ = wgt.font.GetRenderSize("`j*}")
		}
		l.SetMinSize(wt, ht)
	}

	l.SetCursor(cursor)

	style := wgt.Style

	if wgt.StyleHover.exist && wgt.IsHover() {
		style = wgt.StyleHover
	}

	if wgt.StyleActive.exist && wgt.IsActive() {
		style = wgt.StyleActive
	}

	var cmd *cmdList
	if ActiveWidget == wgt {
		cmd = wgt.Container.GetLastCmd(wgt.Zorder + 1)
	} else {
		cmd = wgt.Container.GetLastCmd(wgt.Zorder)
	}

	var crd *RenderData
	if wgt.Constructor != nil {
		var cstyle Style
		crd, cstyle = wgt.Constructor(cmd)
		if cstyle.exist {
			style = cstyle
		}
	}

	r := l.GetBackgroundRect()

	tc := wgt.Texture
	if style.Texture != nil {
		tc = style.Texture
	}

	if tc != nil {
		wgt.renderTexture(cmd, r, style, tc)
	} else if style.BackgroundColor[3] > 0 {
		wgt.renderBackground(cmd, r, style)
	}

	if style.BorderWidth > 0 && style.BorderColor[3] > 0 {
		wgt.renderBorder(cmd, r, style)
	}

	//render text
	if wgt.font != nil && wgt.Text != "" {

		var maxWidth float32 = -1
		if l.w.value > 0 {
			r := l.GetContentRect()
			maxWidth = r.W
		}
		rt := wgt.renderText(style, wt, ht, maxWidth)
		cmd.AddFaces(rt.ComboBuffer, rt.IndexBuffer, rt.Faces)
	}

	if crd != nil && crd.Faces > 0 {
		cmd.AddFaces(crd.ComboBuffer, crd.IndexBuffer, crd.Faces)
	}

	// cursor = mgl.Vec2{cursor[0] + l.W, cursor[1] - l.H}

	if l.PositionFixed {
		return 0, 0
	}

	return l.W, l.H
}

func (wgt *Widget) renderText(style Style, w, h, maxWidth float32) (rt *RenderData) {
	//calculate size from text

	switch wgt.TextAlign {
	case TALIGN_LEFT:
		rt = wgt.font.CreateTextAdv(wgt.Layout.GetTextPosLeft(h), style.TextColor, maxWidth, -1, -1, wgt.Text)
	case TALIGN_CENTER:
		rt = wgt.font.CreateTextAdv(wgt.Layout.GetTextPosCenter(w, h), style.TextColor, maxWidth, -1, -1, wgt.Text)
	case TALIGN_RIGHT:
		rt = wgt.font.CreateTextAdv(wgt.Layout.GetTextPosRight(w, h), style.TextColor, maxWidth, -1, -1, wgt.Text)
	}

	return
}

func (wgt *Widget) renderTexture(cmd *cmdList, r Rect, style Style, tc *TextureChunk) {
	// cmd.textureID = tc.pack.ID
	combos, indexes, fc := cmd.DrawRectFilledDC(r, style.BackgroundColor, tc.pack.Texture, tc.Offset)
	cmd.AddFaces(combos, indexes, fc)
}

func (wgt *Widget) renderBackground(cmd *cmdList, r Rect, style Style) {
	combos, indexes, fc := cmd.DrawRectFilledDC(r, style.BackgroundColor, defaultTextureSampler, whitePixelUv)
	cmd.AddFaces(combos, indexes, fc)
}

func (wgt *Widget) renderBorder(cmd *cmdList, r Rect, style Style) {
	borderRect := r

	borderRect.TLX -= style.BorderWidth
	borderRect.TLY += style.BorderWidth
	borderRect.BRX += style.BorderWidth
	borderRect.BRY -= style.BorderWidth

	left := borderRect
	right := borderRect
	top := borderRect
	bottom := borderRect

	left.BRX = r.TLX
	right.TLX = r.BRX
	top.BRY = r.TLY
	bottom.TLY = r.BRY

	combos, indexes, fc := cmd.DrawRectFilledDC(left, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.AddFaces(combos, indexes, fc)

	combos, indexes, fc = cmd.DrawRectFilledDC(right, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.AddFaces(combos, indexes, fc)

	combos, indexes, fc = cmd.DrawRectFilledDC(top, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.AddFaces(combos, indexes, fc)

	combos, indexes, fc = cmd.DrawRectFilledDC(bottom, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.AddFaces(combos, indexes, fc)
}

func (c *Container) NewText(text string) *Widget {
	wgt := &Widget{
		Text:      text,
		font:      GetFont(c.FontName),
		Style:     DefaultTextStyle,
		Container: c,
		Layout:    NewLayoutZero(c.Layout),
	}

	c.addWidget(wgt)
	return wgt
}

//simple row
func (c *Container) NewRow() *Widget {
	wgt := &Widget{
		Style:     DefaultTextStyle,
		Container: c,
		Layout:    NewLayout("", "", "100%", "", c.Layout),
	}

	c.addWidget(wgt)
	return wgt
}

func (c *Container) NewButton(text string, f Callback) *Widget {
	wgt := &Widget{
		Text:        text,
		TextAlign:   TALIGN_CENTER,
		font:        GetFont(c.FontName),
		Style:       DefaultBtnStyle,
		StyleHover:  DefaultBtnStyleHover,
		StyleActive: DefaultBtnStyleActive,
		Container:   c,
		Layout:      NewLayoutZero(c.Layout),
		OnActive:    f,
	}
	wgt.Constructor = wgt.buttonConstructor

	c.addWidget(wgt)
	return wgt
}

func (wgt *Widget) buttonConstructor(cmd *cmdList) (crd *RenderData, style Style) {

	if wgt.OnActive != nil {
		click, onWidget := wgt.IsClick()
		if click && onWidget {
			wgt.OnActive(wgt)
			style = wgt.StyleActive
		}
	}

	return
}

// NewInput creates an editbox control that changes the value string.
func (c *Container) NewInput(id string, text *string, f Callback) *Widget {
	wgt := &Widget{
		ID:          id,
		Text:        *text,
		font:        GetFont(c.FontName),
		Style:       DefaultInputStyle,
		StyleActive: DefaultInputStyleActive,
		Container:   c,
		Layout:      NewLayout("", "", "100%", "28px", c.Layout),
		OnKeyEnter:  f,
	}
	wgt.ConstructorData = &input{value: text, runes: []rune(*text)}
	wgt.Constructor = wgt.inputConstructor

	c.addWidget(wgt)
	return wgt
}

type input struct {
	value       *string
	runes       []rune
	cursor      int
	cursorTimer float32
}

func (wgt *Widget) inputConstructor(cmd *cmdList) (crd *RenderData, style Style) {

	click, onWidget := wgt.IsClick()
	if click {
		if onWidget {
			ActiveWidget = wgt
			return
		} else if ActiveWidget == wgt {
			ActiveWidget = nil
			return
		}
	}

	if ActiveWidget == nil {
		Keys.DisableListening()
		return
	}

	if ActiveWidget != wgt {
		return
	}

	inp := wgt.ConstructorData.(*input)

	// grab the key events
	for _, k := range Keys.GetKeys() {

		switch k.KeyCode {
		case glfw.KeyRight:
			if inp.cursor < len(inp.runes) {
				inp.cursor++
			}
		case glfw.KeyLeft:
			if inp.cursor > 0 {
				inp.cursor--
			}
		case glfw.KeyBackspace:
			if inp.cursor > 0 {
				inp.runes = append(inp.runes[:inp.cursor-1], inp.runes[inp.cursor:]...)
				inp.cursor--
			}
		case glfw.KeyDelete:
			if inp.cursor < len(inp.runes) {
				inp.runes = append(inp.runes[:inp.cursor], inp.runes[inp.cursor+1:]...)
			}
		case glfw.KeyEnter, glfw.KeyKPEnter:
			if wgt.OnKeyEnter != nil {
				wgt.OnKeyEnter(wgt)
			}
			ActiveWidget = nil
		case glfw.KeyEscape:
			ActiveWidget = nil
		case glfw.KeyEnd:
			inp.cursor = len(inp.runes)
		case glfw.KeyHome:
			inp.cursor = 0
		case glfw.KeyV:
			if k.Ctrl {
				str, _ := window.GetClipboardString()
				inp.runes = append(inp.runes[:inp.cursor],
					append([]rune(str), inp.runes[inp.cursor:]...)...)
			}
		}
	}

	runes := Keys.GetRunes()
	inp.runes = append(inp.runes[:inp.cursor],
		append(runes, inp.runes[inp.cursor:]...)...)
	inp.cursor += len(runes)

	*inp.value = string(inp.runes)
	wgt.Text = *inp.value

	inp.cursorTimer += dt
	if inp.cursorTimer < 0.6 {
		//render text cursor vertical line

		lenText := wgt.font.OffsetForIndex(*inp.value, inp.cursor)

		r := wgt.Layout.GetContentRect()
		r.TLX += lenText + 1
		r.BRX = r.TLX + 2

		crd = new(RenderData)
		crd.ComboBuffer, crd.IndexBuffer, crd.Faces = cmd.DrawRectFilledDC(r, wgt.StyleActive.TextColor, defaultTextureSampler, whitePixelUv)
	}
	if inp.cursorTimer > 1 {
		inp.cursorTimer = 0
	}

	return crd, wgt.StyleActive
}

func (c *Container) NewCheckbox(value *bool, f Callback) *Widget {
	wgt := &Widget{
		Style:       DefaultBtnStyle,
		StyleHover:  DefaultBtnStyleHover,
		StyleActive: DefaultBtnStyleActive,
		Container:   c,
		Layout:      NewLayout("", "", "28px", "28px", c.Layout),
		OnActive:    f,
	}
	wgt.ConstructorData = &checkbox{value: value}
	wgt.Constructor = wgt.checkboxConstructor

	c.addWidget(wgt)

	return wgt
}

type checkbox struct {
	value *bool
}

func (wgt *Widget) checkboxConstructor(cmd *cmdList) (crd *RenderData, style Style) {

	chk := wgt.ConstructorData.(*checkbox)

	click, onWidget := wgt.IsClick()
	if click && onWidget {
		*chk.value = !*chk.value
		log.Println(*chk.value)

		if *chk.value {
			wgt.State = STATE_CHECKED
		} else {
			wgt.State = STATE_NORMAL
		}

		if wgt.OnActive != nil {
			wgt.OnActive(wgt)
		}
	}

	if wgt.State == STATE_CHECKED && wgt.StyleActive.exist {
		r := wgt.Layout.GetContentRect()

		crd = new(RenderData)
		crd.ComboBuffer, crd.IndexBuffer, crd.Faces = cmd.DrawRectFilledDC(r, wgt.StyleActive.TextColor, defaultTextureSampler, whitePixelUv)
	}

	return
}

func (c *Container) NewProgressBar(value *float32, min, max float32, f Callback) *Widget {
	wgt := &Widget{
		Style:       DefaultBtnStyle,
		StyleActive: DefaultBtnStyleActive,
		Container:   c,
		Layout:      NewLayout("", "", "100%", "28px", c.Layout),
		OnActive:    f,
	}

	wgt.ConstructorData = &progressbar{value, min, max}
	wgt.Constructor = wgt.progressbarConstructor

	c.addWidget(wgt)

	return wgt
}

type progressbar struct {
	value    *float32
	min, max float32
}

func (wgt *Widget) progressbarConstructor(cmd *cmdList) (crd *RenderData, style Style) {

	data := wgt.ConstructorData.(*progressbar)
	if *data.value <= data.min {
		return
	}

	r := wgt.Layout.GetContentRect()

	percent := *data.value / data.max
	if percent >= 1 {
		percent = 1
		if wgt.OnActive != nil {
			wgt.OnActive(wgt)
		}
	}

	r.BRX = r.TLX + r.W*percent

	crd = new(RenderData)
	crd.ComboBuffer, crd.IndexBuffer, crd.Faces = cmd.DrawRectFilledDC(r, wgt.StyleActive.TextColor, defaultTextureSampler, whitePixelUv)

	return
}

//
// DragAndDrop ===================================
//

//DADGroup compound slot and items
type DADGroup struct {
	ID        string
	container *Container

	activeItem *DADItem
	activeSlot *DADSlot
}

//NewDragAndDropGroup create new drag and drop group
func (c *Container) NewDragAndDropGroup(id string) *DADGroup {
	return &DADGroup{ID: id, container: c}
}

//DADItem it is movable item
type DADItem struct {
	*Widget

	Slot  *DADSlot
	Group *DADGroup
	Value interface{}
}

//NewItem - create new DADItem
func (group *DADGroup) NewItem(id, x, y, w, h string, tc *TextureChunk, value interface{}) (*DADItem, error) {
	c := group.container

	wgt := &Widget{
		ID:         id,
		Container:  c,
		Style:      DefaultDaDItemStyle,
		StyleHover: DefaultDaDItemStyleHover,
		Layout:     NewLayout(x, y, w, h, c.Layout),
	}

	wgt.Layout.PositionFixed = true
	if w == h {
		wgt.Layout.Square = true
	}

	wgt.Texture = tc

	c.addWidget(wgt)

	item := &DADItem{wgt, nil, group, value}
	wgt.Constructor = item.constructor

	return item, nil
}

func (item *DADItem) constructor(cmd *cmdList) (crd *RenderData, style Style) {

	if item.Slot != nil {
		item.Layout.X = item.Slot.Layout.X
		item.Layout.Y = item.Slot.Layout.Y
	}

	down, onWidget := item.IsMouseDown()
	if down {
		if onWidget && item.Group.activeItem == nil {
			item.Group.activeItem = item //set active item
		}

		if item.Group.activeItem == item {
			ActiveWidget = item.Widget
			//move item with mouse
			item.Layout.X = Mouse.X
			item.Layout.Y = Mouse.Y
		}
	}

	click, _ := item.IsClick()
	if click && item.Group.activeItem == item {
		slot := item.Group.activeSlot
		if slot != nil {

			var place = true
			if slot.callback != nil {
				place = slot.callback(item, slot, item.Value)
			}

			if place { //place item to slot
				item.Layout.X = slot.Layout.X
				item.Layout.Y = slot.Layout.Y

				item.Slot = slot
				if slot.Item != nil {
					slot.Item.Slot = nil
				}
				slot.Item = item
			}

		} else {
			//clear slot and reset item position to default
			if item.Slot != nil {
				item.Slot.Item = nil
			}
			item.Slot = nil
			item.Layout.Update()
		}

		item.Group.activeItem = nil
		item.Group.activeSlot = nil

	}

	return
}

//DADCallback call when item place to slot, in argument: item, slot and item value, boolean returned value allows place item to this slot or not
type DADCallback func(item *DADItem, slot *DADSlot, value interface{}) bool

//DADSlot is slot for place drag and drop item
type DADSlot struct {
	*Widget

	Item     *DADItem
	Group    *DADGroup
	callback DADCallback
}

//NewSlot - create new drag and drop slot and call callback on it
func (group *DADGroup) NewSlot(id string, x, y, w, h string, f DADCallback) *DADSlot {
	c := group.container

	wgt := &Widget{
		ID:          id,
		Container:   c,
		Style:       DefaultBtnStyle,
		StyleHover:  DefaultBtnStyleHover,
		StyleActive: DefaultBtnStyleActive,
		Layout:      NewLayout(x, y, w, h, c.Layout),
	}

	wgt.Layout.PositionFixed = true
	if w == h {
		wgt.Layout.Square = true
	}

	c.addWidget(wgt)

	slot := &DADSlot{wgt, nil, group, f}
	wgt.Constructor = slot.constructor

	return slot
}

func (slot *DADSlot) constructor(cmd *cmdList) (crd *RenderData, style Style) {

	if item := slot.Group.activeItem; item != nil {
		style = slot.StyleHover

		if slot.Layout.ContainsPoint(Mouse.X, Mouse.Y) {
			style = slot.StyleActive
			slot.Group.activeSlot = slot
		} else if slot.Group.activeSlot == slot {
			slot.Group.activeSlot = nil
		}
	}

	return
}
