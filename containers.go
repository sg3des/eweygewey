package fizzgui

// Container represents a collection of widgets in the user interface.
type Container struct {
	ID               string
	Hidden           bool
	AutoAdjustHeight bool

	//not yet ready
	ShowScrollBar  bool
	IsScrollable   bool
	ScrollBarWidth float32
	ScrollOffset   float32

	FontName string
	Style    Style
	Layout   *Layout

	Zorder uint8

	Widgets []*Widget
}

//NewContainer creates new container for widgets
//x,y,w,h is string size ex: "80%", "200px"...
func NewContainer(id string, x, y, w, h string) *Container {
	c := &Container{
		ID:             id,
		Layout:         NewLayout(x, y, w, h, wndLayout),
		ScrollBarWidth: 10,
		Style:          DefaultContainerStyle,
		FontName:       "Default",
	}

	containers = append(containers, c)
	c.Layout.Update()

	return c
}

func (c *Container) addWidget(wgt *Widget) {
	c.Widgets = append(c.Widgets, wgt)
}

//Close function remove this window from window slice
func (c *Container) Close() {
	DelContainer(c)
}

// construct should be call each frame
func (c *Container) construct() {
	c.Layout.Update()
	// Keys.GetKeys()

	// empty out the cmd list and start a new command
	// c.zcmds = make(map[uint8][]*cmdList)

	if c.IsScrollable && HoverContainer == c {
		c.ScrollOffset -= Mouse.ScrollDelta //c.Owner.GetScrollWheelDelta(true)
		if c.ScrollOffset < 0.0 {
			c.ScrollOffset = 0.0
		}
	}

	for i := 0; i < len(c.Widgets); i++ {
		if c.Widgets[i].destroy == true {
			c.Widgets[i] = nil
			c.Widgets = append(c.Widgets[:i], c.Widgets[i+1:]...)
			i--
		}
	}

	//cursor initialize with content point of left X and Top Y
	cursor := c.newCursor()
	for _, wgt := range c.Widgets {
		w, h := wgt.draw(cursor)
		cursor.add(w, h)
	}

	if !c.Hidden {
		c.draw(cursor.Y)
	}
}

//Cursor provide point to widgets position
type Cursor struct {
	Layout *Layout

	X         float32
	Y         float32
	rowHeight float32
}

func (c *Container) newCursor() (cursor *Cursor) {
	r := c.Layout.GetContentRect()

	cursor = &Cursor{
		Layout: c.Layout,
		X:      r.TLX,
		Y:      r.TLY,
	}

	return
}

func (cursor *Cursor) add(w, h float32) {
	if cursor.rowHeight < h {
		cursor.rowHeight = h
	}

	cursor.X += w
	r := cursor.Layout.GetContentRect()

	if cursor.X >= r.BRX {
		cursor.X = r.TLX
		cursor.Y -= cursor.rowHeight
		cursor.rowHeight = 0
	}
}

func (cursor *Cursor) NextRow() {
	r := cursor.Layout.GetContentRect()

	cursor.X = r.TLX
	cursor.Y -= cursor.rowHeight
	cursor.rowHeight = 0
}

func (c *Container) draw(bry float32) {
	cmd := GetFirstCmd(c.Zorder)

	r := c.Layout.GetBackgroundRect()
	if c.AutoAdjustHeight {
		r.BRY = bry - c.Layout.Padding.B
	}

	if tex := c.Style.Texture; tex != nil {
		cmd.DrawFilledRect(r, c.Style.BackgroundColor, tex.Tex, tex.Offset)
	} else {
		cmd.DrawFilledRect(r, c.Style.BackgroundColor, defaultTextureSampler, whitePixelUv)

		if c.Style.BorderWidth > 0 && c.Style.BorderColor[3] > 0 {
			renderBorder(cmd, r, c.Style)
		}
	}

}
