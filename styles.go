package fizzgui

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Color takes the color parameters as integers and returns them
// as a float vector.
func Color(r, g, b, a int) mgl32.Vec4 {
	return mgl32.Vec4{float32(r) / 255.0, float32(g) / 255.0, float32(b) / 255.0, float32(a) / 255.0}
}

type Style struct {
	exist           bool
	TextColor       mgl32.Vec4
	BackgroundColor mgl32.Vec4

	BorderWidth float32
	BorderColor mgl32.Vec4

	Texture *TextureChunk
}

func NewStyle(textColor, bgColor mgl32.Vec4) Style {
	return Style{
		exist:           true,
		TextColor:       textColor,
		BackgroundColor: bgColor,
	}
}

func NewStyleTexture(tc *TextureChunk) Style {
	return Style{
		exist:           true,
		TextColor:       TextColor,
		BackgroundColor: mgl32.Vec4{1, 1, 1, 1},
		Texture:         tc,
	}
}

//CONTAINER STYLE
var DefaultContainerStyle = &Style{
	BackgroundColor: ContainerBGColor,
}

//TEXT STYLE
var DefaultTextStyle = Style{
	exist:     true,
	TextColor: TextColor,
}

//BUTTON STYLE
var DefaultBtnStyle = Style{
	exist:           true,
	TextColor:       TextColor,
	BackgroundColor: BGColorDark,
	BorderColor:     BorderColor,
	BorderWidth:     2,
}

var DefaultBtnStyleHover = Style{
	exist:           true,
	TextColor:       TextColorSelected,
	BackgroundColor: BGColorDarkHover,
	BorderColor:     BorderColor,
	BorderWidth:     2,
}

var DefaultBtnStyleActive = Style{
	exist:           true,
	TextColor:       TextColorSelected,
	BackgroundColor: BGColorHighlight,
	BorderColor:     BorderColor,
	BorderWidth:     2,
}

//INPUT STYLE
var DefaultInputStyle = Style{
	exist:           true,
	TextColor:       TextColor,
	BackgroundColor: BGColor,
}

var DefaultInputStyleActive = Style{
	exist:           true,
	TextColor:       TextColorSelected,
	BackgroundColor: BGColorHover,
	BorderWidth:     2,
	BorderColor:     BorderColorHiglight,
}

//DRAG AND DROP STYLE
var DefaultDaDItemStyle = Style{
	exist:           true,
	BackgroundColor: mgl32.Vec4{1, 1, 1, 1},
}
var DefaultDaDItemStyleHover = Style{
	exist:           true,
	BackgroundColor: mgl32.Vec4{0.8, 0.8, 1, 1},
}

var ContainerBGColor = mgl32.Vec4{0.15, 0.15, 0.15, 0.75}

var BGColor = mgl32.Vec4{0.3, 0.3, 0.3, 1}
var BGColorHover = mgl32.Vec4{0.4, 0.4, 0.4, 1}
var BGColorSelected = mgl32.Vec4{0.5, 0.5, 0.5, 1}

var BGColorDark = mgl32.Vec4{0.18, 0.18, 0.18, 1}
var BGColorDarkHover = mgl32.Vec4{0.28, 0.28, 0.28, 1}

var BGColorHighlight = mgl32.Vec4{0.17, 0.4, 0.63, 1}

var TextColor = mgl32.Vec4{0.8, 0.8, 0.8, 1}
var TextColorSelected = mgl32.Vec4{0.9, 0.9, 0.9, 1}
var TextColorHiglight = mgl32.Vec4{0.17, 0.4, 0.63, 1}

var BorderColor = mgl32.Vec4{0.15, 0.15, 0.15, 1}
var BorderColorHiglight = mgl32.Vec4{0.17, 0.4, 0.63, 1}
