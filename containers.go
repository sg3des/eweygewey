// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

// BuildCallback is a type for the function that builds the widgets for the window.
// type WindowContrsuctor func(window *Window)

// Container represents a collection of widgets in the user interface.
type Container struct {
	// ID is the widget id string for the window for claiming focus.
	ID string

	// ShowScrollBar indicates if the scroll bar should be attached to the side
	// of the window
	ShowScrollBar bool

	// IsScrollable indicates if the window should scroll the contents based
	// on mouse scroll wheel input.
	IsScrollable bool

	// AutoAdjustHeight indicates if the window's height should be automatically
	// adjusted to accommodate all of the widgets.
	AutoAdjustHeight bool

	// cmds is the slice of cmdLists used to to render the window
	zcmds map[uint8][]*cmdList

	FontName string

	Style *Style

	ScrollBarWidth float32
	ScrollOffset   float32

	Layout *Layout

	Widgets []*Widget

	// wgtCursor    mgl.Vec2
	// wgtRowHeight float32
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
	c.zcmds = make(map[uint8][]*cmdList)

	if c.IsScrollable && HoverContainer == c {
		c.ScrollOffset -= Mouse.ScrollDelta //c.Owner.GetScrollWheelDelta(true)
		if c.ScrollOffset < 0.0 {
			c.ScrollOffset = 0.0
		}
	}

	//cursor initialize with content point of left X and Top Y
	cursor := c.newCursor()
	for _, wgt := range c.Widgets {
		w, h := wgt.draw(cursor)
		cursor.add(w, h)
	}

	// if c.AutoAdjustHeight {
	// 	c.Layout.h.minValue = cursor.Y
	// }

	c.draw(cursor.Y)
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

// draw builds the background for the window
func (c *Container) draw(bry float32) {
	var combos []float32
	var indexes []uint32
	var fc uint32

	// get the first cmdList and insert the frame data into it
	cmd := c.GetFirstCmd(0)

	r := c.Layout.GetBackgroundRect()
	if c.AutoAdjustHeight {
		r.BRY = bry - c.Layout.Padding.B
	}

	if c.Style.Texture != nil {
		// build the background of the window
		combos, indexes, fc = cmd.DrawRectFilledDC(r, c.Style.BackgroundColor, c.Style.Texture.pack.Texture, c.Style.Texture.Offset)
		cmd.PrefixFaces(combos, indexes, fc)
	} else {
		// build the background of the window
		combos, indexes, fc = cmd.DrawRectFilledDC(r, c.Style.BackgroundColor, defaultTextureSampler, whitePixelUv)
		cmd.PrefixFaces(combos, indexes, fc)
	}

}

func NewCmdList(layout *Layout) *cmdList {
	cmdList := newCmdList()
	cmdList.clipRect = layout.GetBackgroundRect()
	return cmdList
}

//GetFirstCmd create new cmd and insert in to first element
func (c *Container) GetFirstCmd(z uint8) *cmdList {
	if _, ok := c.zcmds[z]; !ok {
		c.zcmds[z] = []*cmdList{}
	}

	prepandCmd := NewCmdList(c.Layout)
	c.zcmds[z] = append([]*cmdList{prepandCmd}, c.zcmds[z]...)

	return c.zcmds[z][0]
}

//GetLastCmd will return the last non-custom cmdList
func (c *Container) GetLastCmd(z uint8) *cmdList {
	if _, ok := c.zcmds[z]; !ok {
		c.zcmds[z] = []*cmdList{}
	}

	appendCmd := NewCmdList(c.Layout)
	c.zcmds[z] = append(c.zcmds[z], appendCmd)

	return appendCmd
}

// addNewCmd creates a new cmdList and adds it to the window's slice of cmlLists.
// func (c *Container) addNewCmd() *cmdList {
// 	log.Println("new cmd")
// 	if len(c.cmds) == 0 {
// 		return c.getFirstCmd()
// 	}
// 	newCmd := c.makeCmdList()
// 	c.cmds = append(c.cmds, newCmd)
// 	return newCmd
// }
