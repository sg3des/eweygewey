package fizzgui

import (
	"log"

	"github.com/go-gl/glfw/v3.2/glfw"
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

type WidgetConstructor func() Style
type Callback func(wgt *Widget)

type Widget struct {
	ID string

	Hidden  bool
	destroy bool

	Text      string
	TextAlign TALIGN
	Font      *Font

	Texture *Texture //Global widget texture

	Style       Style
	StyleHover  Style
	StyleActive Style

	State STATE

	Zorder uint8 //initial value
	Z      uint8 //current working value

	Layout    *Layout
	Container *Container

	OnActive   Callback
	OnKeyEnter Callback

	ConstructorData interface{}
	Constructor     WidgetConstructor
}

func (wgt *Widget) Destroy() {
	wgt.Hidden = true
	wgt.destroy = true
}

func (wgt *Widget) SetStyles(normal, hover, active Style, tex *Texture) {
	wgt.Style = normal
	wgt.StyleHover = hover
	wgt.StyleActive = active
	wgt.Texture = tex
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
	wgt.Z = wgt.Zorder + wgt.Container.Zorder

	l := wgt.Layout
	l.Update()

	var wt, ht float32
	if wgt.Font != nil {
		if wgt.Text != "" {
			wt, ht, _ = wgt.Font.GetRenderSize(wgt.Text)
		} else {
			wt, ht, _ = wgt.Font.GetRenderSize("`j*}")
		}
		l.SetMinSize(l.AddOffsets(wt, ht))
	}

	l.SetCursor(cursor)

	style := wgt.Style

	if wgt.StyleHover.exist && wgt.IsHover() {
		style = wgt.StyleHover
	}

	if wgt.StyleActive.exist && wgt.IsActive() {
		style = wgt.StyleActive
	}

	if ActiveWidget == wgt {
		wgt.Z++
	}

	if wgt.Constructor != nil {
		if cstyle := wgt.Constructor(); cstyle.exist {
			style = cstyle
		}
	}

	if wgt.Hidden {
		return 0, 0
	}

	r := l.GetBackgroundRect()
	switch {
	case style.exist && style.Texture != nil:
		wgt.renderTexture(r, style, style.Texture)
	case style.exist && wgt.Texture != nil:
		wgt.renderTexture(r, style, wgt.Texture)
	case style.exist && style.BackgroundColor[3] > 0:
		wgt.renderBackground(r, style)
	}

	if wgt.Font != nil && wgt.Text != "" {
		wgt.renderText(l.GetContentRect(), style, wt, ht)
	}

	if l.PositionFixed {
		return 0, 0
	}

	return l.W, l.H
}

func (wgt *Widget) renderText(r Rect, style Style, w, h float32) {

	var maxWidth float32 = -1
	if wgt.Layout.w.value > 0 {
		maxWidth = r.W
	}

	var rt *RenderData
	switch wgt.TextAlign {
	case TALIGN_LEFT:
		rt = wgt.Font.CreateTextAdv(wgt.Layout.GetTextPosLeft(h), style.TextColor, maxWidth, -1, -1, wgt.Text)
	case TALIGN_CENTER:
		rt = wgt.Font.CreateTextAdv(wgt.Layout.GetTextPosCenter(w, h), style.TextColor, maxWidth, -1, -1, wgt.Text)
	case TALIGN_RIGHT:
		rt = wgt.Font.CreateTextAdv(wgt.Layout.GetTextPosRight(w, h), style.TextColor, maxWidth, -1, -1, wgt.Text)
	}

	cmd := GetLastCmd(wgt.Z)
	cmd.texture = wgt.Font.Texture
	cmd.AddFaces(rt.ComboBuffer, rt.IndexBuffer, rt.Faces)
}

func (wgt *Widget) renderTexture(r Rect, style Style, tc *Texture) {
	cmd := GetLastCmd(wgt.Z)
	cmd.DrawFilledRect(r, style.BackgroundColor, tc.Tex, tc.Offset)
}

func (wgt *Widget) renderBackground(r Rect, style Style) {
	cmd := GetLastCmd(wgt.Z)
	cmd.DrawFilledRect(r, style.BackgroundColor, defaultTextureSampler, whitePixelUv)

	if style.BorderWidth > 0 && style.BorderColor[3] > 0 {
		renderBorder(cmd, r, style)
	}
}

func renderBorder(cmd *cmdList, r Rect, style Style) {
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

	cmd.DrawFilledRect(left, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.DrawFilledRect(right, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.DrawFilledRect(top, style.BorderColor, defaultTextureSampler, whitePixelUv)
	cmd.DrawFilledRect(bottom, style.BorderColor, defaultTextureSampler, whitePixelUv)
}

func (c *Container) NewText(text string) *Widget {
	wgt := &Widget{
		Text:      text,
		Font:      GetFont(c.FontName),
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
		Font:        GetFont(c.FontName),
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

func (wgt *Widget) buttonConstructor() (style Style) {

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
		Font:        GetFont(c.FontName),
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

func (wgt *Widget) inputConstructor() (style Style) {

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

		lenText := wgt.Font.OffsetForIndex(*inp.value, inp.cursor)

		r := wgt.Layout.GetContentRect()
		r.TLX += lenText + 1
		r.BRX = r.TLX + 2

		cmd := GetLastCmd(wgt.Z + 1)
		cmd.DrawFilledRect(r, wgt.StyleActive.TextColor, defaultTextureSampler, whitePixelUv)
	}
	if inp.cursorTimer > 1 {
		inp.cursorTimer = 0
	}

	return wgt.StyleActive
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

func (wgt *Widget) checkboxConstructor() (style Style) {

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

		cmd := GetLastCmd(wgt.Zorder + 1)
		cmd.DrawFilledRect(r, wgt.StyleActive.TextColor, defaultTextureSampler, whitePixelUv)
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

func (wgt *Widget) progressbarConstructor() (style Style) {

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

	cmd := GetLastCmd(wgt.Zorder + 1)
	cmd.DrawFilledRect(r, wgt.StyleActive.TextColor, defaultTextureSampler, whitePixelUv)

	return
}

//
// DragAndDrop ===================================
//

//DADGroup compound slot and items
type DADGroup struct {
	*Container

	ID string

	Slots []*DADSlot
	Items []*DADItem

	activeItem *DADItem
	activeSlot *DADSlot
}

//NewDragAndDropGroup create new drag and drop group
func NewDragAndDropGroup(id string) *DADGroup {
	c := NewContainer(id, "", "", "100%", "100%")
	c.Style.BackgroundColor[3] = 0
	group := &DADGroup{c, id, nil, nil, nil, nil}

	return group
}

//DADItem it is movable item
type DADItem struct {
	*Widget

	Slot  *DADSlot
	Group *DADGroup
	Value interface{}
}

//NewItem - create new DADItem
func (group *DADGroup) NewItem(id, img string, value interface{}) *DADItem {
	c := group.Container

	wgt := &Widget{
		ID:         id,
		Hidden:     true,
		Zorder:     2,
		Container:  c,
		Style:      DefaultDaDItemStyle,
		StyleHover: DefaultDaDItemStyleHover,
		Layout:     NewLayout("0", "0", "100%", "100%", nil),
	}

	wgt.Layout.Margin = Offset{1, 1, 1, 1}
	wgt.Layout.Padding = Offset{0, 0, 0, 0}
	wgt.Layout.PositionFixed = true
	wgt.Layout.Square = true
	wgt.Layout.HAlign = HAlignCenter
	wgt.Layout.VAlign = VAlignMiddle

	var err error
	wgt.Texture, err = NewTextureImg(img)
	if err != nil {
		log.Println(err)
	}

	c.addWidget(wgt)

	item := &DADItem{wgt, nil, group, value}
	wgt.Constructor = item.constructor

	group.Items = append(group.Items, item)

	return item
}

func (item *DADItem) constructor() (style Style) {

	if item.Slot == nil {
		item.Hidden = true
		return
	} else {
		item.Hidden = false
	}

	item.Slot.placeItem(item)

	//if left mouse btn is pressed then drag item by mouse
	down, onWidget := item.IsMouseDown()
	if down {
		if onWidget && item.Group.activeItem == nil {
			item.Group.activeItem = item //set active item
		}

		if item.Group.activeItem == item {
			ActiveWidget = item.Widget
			item.Widget.Zorder++
			//move item with mouse
			item.Layout.X = Mouse.X
			item.Layout.Y = Mouse.Y
		}
	}

	//if left mouse btn is released then calculate place
	click, _ := item.IsClick()
	if click && item.Group.activeItem == item {

		if slot := item.Group.activeSlot; slot != nil {

			var canPut = true
			if slot.callback != nil {
				canPut = slot.callback(item, slot, item.Value)
			}

			if canPut { //place item to slot
				slot.PlaceItem(item)
			}
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
func (c *Container) NewSlot(group *DADGroup, id string, x, y, w, h string, f DADCallback) *DADSlot {

	wgt := &Widget{
		ID:          id,
		Container:   c,
		Style:       DefaultBtnStyle,
		StyleHover:  DefaultBtnStyleHover,
		StyleActive: DefaultBtnStyleActive,
		Layout:      NewLayout(x, y, w, h, c.Layout),
	}

	if x != "" && y != "" {
		wgt.Layout.PositionFixed = true
	}

	wgt.Layout.Padding = Offset{1, 1, 1, 1}
	if w == h {
		wgt.Layout.Square = true
	}

	c.addWidget(wgt)

	slot := &DADSlot{wgt, nil, group, f}
	group.Slots = append(group.Slots, slot)

	wgt.Constructor = slot.constructor

	return slot
}

func (slot *DADSlot) constructor() (style Style) {

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

func (slot *DADSlot) PlaceItem(item *DADItem) {
	var prevSlot = item.Slot
	var prevItem = slot.Item

	if prevSlot != nil && prevItem != nil {
		if prevSlot == slot && prevItem == item {
			return
		}
	}

	//clear previous positions
	if item.Slot != nil {
		item.Slot.Item = nil
	}
	if slot.Item != nil {
		slot.Item.Slot = nil
	}

	slot.placeItem(item)
	if prevSlot != nil && prevItem != nil {
		prevSlot.placeItem(prevItem)
	}
	item.Layout.Update()
}

func (slot *DADSlot) placeItem(item *DADItem) {
	slot.Item = item
	item.Slot = slot
	item.Layout.parent = slot.Layout
	item.Zorder = slot.Zorder + 1
}
