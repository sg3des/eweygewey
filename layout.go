package fizzgui

import (
	"log"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

type ALIGN int

const (
	ALIGN_LEFT_TOP ALIGN = iota
	ALIGN_LEFT_MIDDLE
	ALIGN_LEFT_BOTTOM
	ALIGN_CENTER_TOP
	ALIGN_CENTER_MIDDLE
	ALIGN_CENTER_BOTTOM
	ALIGN_RIGHT_TOP
	ALIGN_RIGHT_MIDDLE
	ALIGN_RIGHT_BOTTOM
)

//Size contains universial reresentation for fixed and percent sizes
type Size struct {
	value    float32
	minValue float32
	percent  bool
}

//ParseSize parse string value to float and percent bool
func ParseSize(val string) (i float32, percent bool) {
	if val == "" || val == "auto" {
		return
	}

	if strings.HasSuffix(val, "%") {
		val = strings.TrimSuffix(val, "%")
		percent = true
	} else {
		val = strings.TrimSuffix(val, "px")
	}

	i64, err := strconv.ParseFloat(val, 32)
	if err != nil {
		log.Println(err)
	}

	i = float32(i64)
	if percent {
		i = i / 100
	}

	return
}

// NewSize return Size object contains value(ex:"100%","50px",etc) and potiner on maxValue for calculate percent value
func NewSize(val string, minValue float32) (s Size) {
	s = Size{
		minValue: minValue,
	}

	s.value, s.percent = ParseSize(val)

	return
}

//
//Offset is alternative of slice mgl32.Vec4
type Offset struct {
	L, T, R, B float32
}

//Vec4 - convert offset to mgl32.Vec4
func (o Offset) Vec4() mgl32.Vec4 {
	return mgl32.Vec4{o.L, o.T, o.R, o.B}
}

var DefaultMargin = Offset{4, 4, 4, 4}
var DefaultPadding = Offset{4, 4, 4, 4}

//Layout is reresentation of object size position and other options
type Layout struct {
	// Align ALIGN
	PositionFixed bool
	Square        bool

	parent *Layout
	x      Size
	y      Size
	w      Size
	h      Size

	//value fot position content into this layout
	X float32 //content X
	Y float32 //content Y
	W float32 //max content width
	H float32 //max content height

	Padding Offset
	Margin  Offset
}

func NewLayout(x, y, w, h string, parent *Layout) *Layout {
	r := &Layout{
		parent: parent,
		x:      NewSize(x, 0),
		y:      NewSize(y, 0),
		w:      NewSize(w, 0),
		h:      NewSize(h, 0),

		Padding: DefaultPadding,
		Margin:  DefaultMargin,
	}

	return r
}

func NewLayoutZero(parent *Layout) *Layout {
	return NewLayout("", "", "", "", parent)
}

func (l *Layout) SetWidth(val string) {
	l.w.value, l.w.percent = ParseSize(val)
}

func (l *Layout) SetHeight(val string) {
	l.h.value, l.h.percent = ParseSize(val)
}

func (l *Layout) SetX(val string) {
	l.x.value, l.x.percent = ParseSize(val)
}

func (l *Layout) SetY(val string) {
	l.y.value, l.y.percent = ParseSize(val)
}

//Update layout values, should be call each frame
func (l *Layout) Update() {
	r := l.parent.GetContentRect()
	// log.Println(r)

	//x
	l.X = r.TLX
	if l.x.percent {
		l.X += l.x.value * r.W
	} else {
		l.X += l.x.value
	}

	//y
	l.Y = r.TLY
	if l.y.percent {
		l.Y -= l.y.value * r.H
	} else {
		l.Y -= l.y.value
	}

	//height
	if l.h.percent {
		l.H = l.h.value * r.H
	} else {
		l.H = l.h.value
	}
	if l.H < l.h.minValue {
		l.H = l.h.minValue
	}

	//width

	if l.w.percent {
		l.W = l.w.value * r.W
	} else {
		l.W = l.w.value
	}
	if l.W < l.w.minValue {
		l.W = l.w.minValue
	}

	if l.Square {
		if l.W > l.H {
			l.H = l.H * (l.W / l.H)
		}
		if l.W < l.H {
			l.W = l.W * (l.H / l.W)
		}
	}

	// log.Println(l.Y, l.y.value)
}

//SetMinSize it`s width and height if incoming value more then exists
func (l *Layout) SetMinSize(w, h float32) {
	w += l.Margin.L + l.Margin.R + l.Padding.L + l.Padding.R
	h += l.Margin.T + l.Margin.B + l.Padding.T + l.Padding.B

	l.w.minValue = w
	l.h.minValue = h

	r := l.parent.GetContentRect()
	if l.w.minValue > r.W {
		l.w.minValue = r.W
	}
	if l.h.minValue > r.H {
		l.h.minValue = r.H
	}

}

func (l *Layout) SetCursor(cursor *Cursor) {
	if l.PositionFixed {
		return
	}

	l.X = cursor.X
	l.Y = cursor.Y

	r := l.parent.GetContentRect()
	if l.X+l.W > r.BRX {
		cursor.NextRow()
		l.X = cursor.X
		l.Y = cursor.Y
	}

	// l.Update()
}

func (l *Layout) SummOffsets() (o Offset) {
	o.L = l.Margin.L + l.Padding.L
	o.R = l.Margin.R + l.Padding.R
	o.T = l.Margin.T + l.Padding.T
	o.B = l.Margin.B + l.Padding.B
	return
}

func (l *Layout) GetTextPosLeft(h float32) (textPos mgl32.Vec2) {
	textPos[0] = l.X + l.Margin.L + l.Padding.L
	textPos[1] = l.Y - l.H/2 + h/2
	return
}

func (l *Layout) GetTextPosCenter(w, h float32) (textPos mgl32.Vec2) {
	if w >= l.W {
		return l.GetTextPosLeft(h)
	}

	textPos[0] = l.X + l.W/2 - w/2
	textPos[1] = l.Y - l.H/2 + h/2
	return
}

func (l *Layout) GetTextPosRight(w, h float32) (textPos mgl32.Vec2) {
	if w >= l.W {
		return l.GetTextPosLeft(h)
	}

	textPos[0] = l.X + l.W - w - l.Margin.R - l.Padding.R
	textPos[1] = l.Y - l.H/2 + h/2
	return
}

type Rect struct {
	TLX float32
	TLY float32
	BRX float32
	BRY float32

	W float32
	H float32
}

func (l *Layout) GetBackgroundRect() (r Rect) {
	r.TLX = l.X + l.Margin.L
	r.TLY = l.Y - l.Margin.T

	r.BRX = l.X + l.W - l.Margin.R
	r.BRY = l.Y - l.H + l.Margin.B

	r.W = l.W - l.Margin.L - l.Margin.R
	r.H = l.H - l.Margin.T - l.Margin.B

	if l.PositionFixed {
		r.TLX -= r.W / 2
		r.BRX -= r.W / 2
		r.TLY += r.H / 2
		r.BRY += r.H / 2
	}

	return
}

func (l *Layout) GetContentRect() (r Rect) {
	r.TLX = l.X + l.Margin.L + l.Padding.L
	r.TLY = l.Y - l.Margin.T - l.Padding.T

	r.BRX = l.X + l.W - l.Margin.R - l.Padding.R
	r.BRY = l.Y - l.H + l.Margin.B + l.Padding.B

	r.W = l.W - l.Margin.L - l.Margin.R - l.Padding.L - l.Padding.R
	r.H = l.H - l.Margin.T - l.Margin.B - l.Padding.T - l.Padding.B
	return
}

func (l *Layout) ContainsPoint(x, y float32) bool {
	r := l.GetBackgroundRect()
	if x > r.TLX && x < r.BRX && y < r.TLY && y > r.BRY {
		return true
	}

	return false
}