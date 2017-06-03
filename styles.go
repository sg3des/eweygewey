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

	BorderColor mgl32.Vec4
	BorderWidth float32

	Texture *Texture
}

func NewStyle(textColor, bgColor, borderColor mgl32.Vec4, borderWidth float32) Style {
	return Style{
		exist:           true,
		TextColor:       textColor,
		BackgroundColor: bgColor,
		BorderColor:     borderColor,
		BorderWidth:     borderWidth,
	}
}

func NewStyleTexture(tc *Texture, bgColor mgl32.Vec4) Style {
	return Style{
		exist:           true,
		TextColor:       TextColor,
		BackgroundColor: bgColor,
		Texture:         tc,
	}
}

//Default colors
var (
	BGColorContainer = mgl32.Vec4{0.15, 0.15, 0.15, 0.75}

	BGColor         = mgl32.Vec4{0.3, 0.3, 0.3, 1}
	BGColorHover    = mgl32.Vec4{0.4, 0.4, 0.4, 1}
	BGColorSelected = mgl32.Vec4{0.5, 0.5, 0.5, 1}

	BGColorBtn      = mgl32.Vec4{0.18, 0.18, 0.18, 1}
	BGColorBtnHover = mgl32.Vec4{0.28, 0.28, 0.28, 1}

	BGColorHighlight = mgl32.Vec4{0.17, 0.4, 0.63, 1}

	TextColor         = mgl32.Vec4{0.8, 0.8, 0.8, 1}
	TextColorSelected = mgl32.Vec4{0.9, 0.9, 0.9, 1}
	TextColorHiglight = mgl32.Vec4{0.17, 0.4, 0.63, 1}

	BorderColor         = mgl32.Vec4{0.15, 0.15, 0.15, 1}
	BorderColorHiglight = mgl32.Vec4{0.17, 0.4, 0.63, 1}

	BGColorImage      = mgl32.Vec4{0.9, 0.9, 0.9, 1}
	BGColorImageHover = mgl32.Vec4{1, 1, 1, 1}
)

//Default styles
var (
	DefaultContainerStyle Style
	DefaultTextStyle      Style

	DefaultBtnStyle       Style
	DefaultBtnStyleHover  Style
	DefaultBtnStyleActive Style

	DefaultInputStyle       Style
	DefaultInputStyleActive Style

	DefaultDaDItemStyle      Style
	DefaultDaDItemStyleHover Style
)

func initDefaultStyles() {
	n := mgl32.Vec4{}
	DefaultContainerStyle = NewStyle(n, BGColorContainer, n, 0)
	DefaultTextStyle = NewStyle(TextColor, n, n, 0)

	DefaultBtnStyle = NewStyle(TextColor, BGColorBtn, BorderColor, 2)
	DefaultBtnStyleHover = NewStyle(TextColorSelected, BGColorBtnHover, BorderColor, 2)
	DefaultBtnStyleActive = NewStyle(TextColorSelected, BGColorHighlight, BorderColor, 2)

	DefaultInputStyle = NewStyle(TextColor, BGColor, n, 0)
	DefaultInputStyleActive = NewStyle(TextColorSelected, BGColorHover, BorderColorHiglight, 2)

	DefaultDaDItemStyle = NewStyle(n, BGColorImage, n, 0)
	DefaultDaDItemStyleHover = NewStyle(n, BGColorImageHover, n, 0)
}
