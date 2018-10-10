// Copyright 2015, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

/*
Based primarily on gltext found at https://github.com/go-gl/gltext
But also based on examples from the freetype-go project:

	https://github.com/golang/freetype

This implementation differs in the way the images are rendered and then
copied into an OpenGL texture. In addition to that, this module can
create a renderable 'string' node which is a bunch of polygons with uv's
mapped to the appropriate glyphs.
*/

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	ft "github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	imgfont "golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var FontGlyphs = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890., :[]{}\\|<>;\"'~`?/-+_=()*&^%$#@!"

// runeData stores information pulled from the freetype parsing of glyphs.
type runeData struct {
	imgX, imgY                    int     // offset into the image texture for the top left position of rune
	advanceWidth, leftSideBearing float32 // HMetric data from glyph
	advanceHeight, topSideBearing float32 // VMetric data from glyph
	uvMinX, uvMinY                float32
	uvMaxX, uvMaxY                float32
}

// Font contains data regarding a font and the texture that was created
// with the specified set of glyphs. It can then be used to create
// renderable string objects.
type Font struct {
	Texture     graphics.Texture
	TextureSize int
	Glyphs      string
	GlyphHeight float32
	GlyphWidth  float32
	locations   map[rune]runeData
	opts        truetype.Options
	face        imgfont.Face
}

// NewFont loads the font from a file and 'registers' it with the UI manager.
func NewFont(name string, fontFilepath string, scaleInt int, glyphs string) (*Font, error) {

	fontBytes, err := ioutil.ReadFile(fontFilepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load font from path: '%s' \n%v", fontFilepath, err)
	}

	f, err := newFont(fontBytes, scaleInt, glyphs)
	if err != nil {
		return nil, err
	}
	fonts[name] = f

	return f, nil
}

func LoadFont(name string, fontBytes []byte, scaleInt int, glyphs string) (*Font, error) {
	f, err := newFont(fontBytes, scaleInt, glyphs)
	if err != nil {
		return nil, err
	}
	fonts[name] = f
	return f, nil
}

// GetFont attempts to get the font by name from the Manager's collection
// It returns the font on success or nil on failure.
func GetFont(name string) *Font {
	f, ok := fonts[name]
	if !ok {
		log.Fatalf("font by name '%s' not loaded", name)
	}
	return f
}

// LoadFont uses the Go freetype library to parse it and render the specified glyphs to a texture that is then buffered into OpenGL.
func newFont(fontBytes []byte, scaleInt int, glyphs string) (f *Font, e error) {
	f = new(Font)
	scale := fixed.I(scaleInt)

	// allocate the location map
	f.locations = make(map[rune]runeData)

	// parse the truetype font data
	ttfData, err := ft.ParseFont(fontBytes)
	if err != nil {
		return f, fmt.Errorf("Failed to prase the truetype font data.\n%v", err)
	}

	f.opts.Size = float64(scaleInt)
	f.face = truetype.NewFace(ttfData, &f.opts)

	// this may have negative components, but get the bounds for the font
	glyphBounds := ttfData.Bounds(scale)

	// width and height are getting +2 here since the glyph will be buffered by a
	// pixel in the texture
	glyphDimensions := glyphBounds.Max.Sub(glyphBounds.Min)
	glyphWidth := fixedInt26ToFloat(glyphDimensions.X) + 3
	glyphHeight := fixedInt26ToFloat(glyphDimensions.Y)
	glyphHeight *= 1.1
	glyphCeilWidth := int(math.Ceil(float64(glyphWidth))) + 4
	glyphCeilHeight := int(math.Ceil(float64(glyphHeight)))

	// create the buffer image used to draw the glyphs
	glyphRect := image.Rect(1, 1, glyphCeilWidth, glyphCeilHeight)
	glyphImg := image.NewRGBA(glyphRect)

	// calculate the area needed for the font texture
	var fontTexSize = 512
	minAreaNeeded := (glyphCeilWidth) * (glyphCeilHeight) * len(glyphs)
	for (fontTexSize * fontTexSize) < minAreaNeeded {
		fontTexSize = fontTexSize * 2
		if fontTexSize > 2048 {
			return f, fmt.Errorf("Font texture was going to exceed 2048x2048 (%d) and that's currently not supported.", fontTexSize)
		}
	}

	// create the font image
	fontImgRect := image.Rect(0, 0, fontTexSize, fontTexSize)
	fontImg := image.NewRGBA(fontImgRect)

	// the number of glyphs
	fontRowSize := fontTexSize / glyphCeilWidth

	// create the freetype context
	c := ft.NewContext()
	c.SetDPI(70)
	c.SetFont(ttfData)
	c.SetFontSize(float64(scaleInt))
	c.SetClip(glyphImg.Bounds())
	c.SetDst(glyphImg)
	c.SetSrc(image.White)
	// c.SetHinting(imgfont.HintingFull)

	var fx, fy int
	for _, ch := range glyphs {

		index := ttfData.Index(ch)
		metricH := ttfData.HMetric(scale, index)
		metricV := ttfData.VMetric(scale, index)

		fxGW := fx * glyphCeilWidth
		fyGH := fy * glyphCeilHeight

		f.locations[ch] = runeData{
			fxGW, fyGH,
			fixedInt26ToFloat(metricH.AdvanceWidth), fixedInt26ToFloat(metricH.LeftSideBearing),
			fixedInt26ToFloat(metricV.AdvanceHeight), fixedInt26ToFloat(metricV.TopSideBearing),
			float32(fxGW) / float32(fontTexSize), (float32(fyGH) + glyphHeight) / float32(fontTexSize),
			(float32(fxGW) + glyphWidth) / float32(fontTexSize), float32(fyGH) / float32(fontTexSize),
		}

		pt := ft.Pt(1, 1+int(c.PointToFixed(float64(scaleInt))>>6))
		_, err := c.DrawString(string(ch), pt)
		if err != nil {
			return f, fmt.Errorf("Freetype returned an error while drawing a glyph: %v.", err)
		}

		// copy the glyph image into the font image
		for subY := 0; subY < glyphCeilHeight; subY++ {
			for subX := 0; subX < glyphCeilWidth; subX++ {
				glyphRGBA := glyphImg.RGBAAt(subX, subY)
				fontImg.SetRGBA((fxGW)+subX, (fyGH)+subY, glyphRGBA)
			}
		}

		// erase the glyph image buffer
		draw.Draw(glyphImg, glyphImg.Bounds(), image.Transparent, image.ZP, draw.Src)

		// adjust the pointers into the font image
		fx++
		if fx > fontRowSize {
			fx = 0
			fy++
		}
	}

	// set the white point
	fontImg.SetRGBA(fontTexSize-1, fontTexSize-1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// buffer the font image into an OpenGL texture
	f.Glyphs = glyphs
	f.TextureSize = fontTexSize
	f.GlyphWidth = glyphWidth
	f.GlyphHeight = glyphHeight
	f.Texture = loadRGBAToTexture(fontImg.Pix, int32(fontImg.Rect.Max.X))

	return
}

// Destroy releases the OpenGL texture for the font.
func (f *Font) Destroy() {
	gfx.DeleteTexture(f.Texture)
}

// GetCurrentScale returns the scale value for the font based on the current
// Manager's resolution vs the resolution the UI was designed for.
func (f *Font) GetCurrentScale() float32 {
	return 1
	// 	w, h := window.GetSize()
	// 	if w > h {
	// 		return float32(w) / float32(h)
	// 	}
	// 	// _, uiHeight := f.Owner.GetResolution()
	// 	// designHeight := f.Owner.GetDesignHeight()
	// 	return float32(h) / float32(w)
}

// GetRenderSize returns the width and height necessary in pixels for the
// font to display a string. The third return value is the advance height the string.
func (f *Font) GetRenderSize(s string) (float32, float32, float32) {
	var w, h float32

	// see how much to scale the size based on current resolution vs desgin resolution
	fontScale := f.GetCurrentScale()

	for _, ch := range s {
		bounds, _, _ := f.face.GlyphBounds(ch)
		glyphDimensions := bounds.Max.Sub(bounds.Min)

		adv, _ := f.face.GlyphAdvance(ch)
		w += fixedInt26ToFloat(adv)

		glyphDYf := fixedInt26ToFloat(glyphDimensions.Y)
		if h < glyphDYf {
			h = glyphDYf
		}
	}

	metrics := f.face.Metrics()
	// advH := fixedInt26ToFloat(metrics.Ascent+metrics.Descent/2) * fontScale
	// log.Println(fixedInt26ToFloat(metrics.Ascent)*fontScale, fixedInt26ToFloat(metrics.Descent)*fontScale, fixedInt26ToFloat(metrics.Height)*fontScale)

	w = w * fontScale
	mhMax := metrics.Ascent
	mhMin := metrics.Ascent
	if metrics.Height > metrics.Ascent {
		mhMax = metrics.Height
	}

	if metrics.Height < metrics.Ascent {
		mhMin = metrics.Height
	}

	h = fixedInt26ToFloat(mhMax-metrics.Descent/2) * fontScale
	advH := fixedInt26ToFloat(mhMin+metrics.Descent/2) * fontScale

	// log.Println(metrics, h, advH)
	// log.Println(f.GlyphHeight)

	// h = f.GlyphHeight
	// advH = f.GlyphHeight
	// advH = f.GlyphHeight * 0.5
	// return w * fontScale, h * fontScale, advH * fontScale
	return w, h, advH
}

// OffsetFloor returns the maximum width offset that will fit between characters that
// is still smaller than the offset passed in.
func (f *Font) OffsetFloor(msg string, offset float32) float32 {
	var w float32

	// see how much to scale the size based on current resolution vs desgin resolution
	fontScale := f.GetCurrentScale()

	for _, ch := range msg {
		adv, ok := f.face.GlyphAdvance(ch)
		if !ok {
			fmt.Printf("ERROR on glyphadvance for %c!\n", ch)
		}
		advf := fixedInt26ToFloat(adv)

		// break if we go over the distance
		if w+advf > offset {
			break
		}
		w += advf
	}

	return w * fontScale
}

// OffsetForIndex returns the width offset that will fit just before the `stopIndex`
// number character in the msg.
func (f *Font) OffsetForIndex(msg string, stopIndex int) float32 {
	return f.OffsetForIndexAdv(msg, 0, stopIndex)
}

// OffsetForIndexAdv returns the width offset that will fit just before the `stopIndex`
// number character in the msg, starting at charStartIndex.
func (f *Font) OffsetForIndexAdv(msg string, charStartIndex int, stopIndex int) float32 {
	var w float32

	// see how much to scale the size based on current resolution vs desgin resolution
	fontScale := f.GetCurrentScale()
	var i int
	for _, ch := range msg[charStartIndex:] {
		// calculate up to the stopIndex but do not include it
		if i+charStartIndex >= stopIndex {
			break
		}
		adv, _ := f.face.GlyphAdvance(ch)
		w += fixedInt26ToFloat(adv)
		i++
	}

	return w * fontScale
}

// fixedInt26ToFloat converts a fixed int 26:6 precision to a float32.
func fixedInt26ToFloat(fixedInt fixed.Int26_6) float32 {
	var result float32
	i := int32(fixedInt)
	result += float32(i >> 6)
	result += float32(i&0x003F) / float32(64.0)
	return result
}

// RenderData is a structure containing the raw OpenGL VBO data needed
// to render a text string for a given texture.
type RenderData struct {
	ComboBuffer         []float32 // the combo VBO data (vert/uv/color)
	IndexBuffer         []uint32  // the element index VBO data
	Faces               uint32    // the number of faces in the text string
	Width               float32   // the width in pixels of the text string
	Height              float32   // the height in pixels of the text string
	AdvanceHeight       float32   // the amount of pixels to move the pen in the verticle direction
	CursorOverflowRight bool      // whether or not the cursor was too far to the right for string width
}

// CreateText makes a new renderable object from the supplied string
// using the data in the font. The data is returned as a RenderData object.
func (f *Font) CreateText(pos mgl.Vec2, color mgl.Vec4, msg string) *RenderData {
	return f.CreateTextAdv(pos, color, -1, -1, -1, msg)
}

// CreateText makes a new renderable object from the supplied string
// using the data in the font. The string returned will be the maximum amount of the msg that fits
// the specified maxWidth (if greater than 0.0) starting at the charOffset specified.
// The data is returned as a RenderData object.
func (f *Font) CreateTextAdv(pos mgl.Vec2, color mgl.Vec4, maxWidth float32, charOffset int, cursorPosition int, s string) *RenderData {
	// this is the texture ID of the font to use in the shader; by default
	// the library always binds the font to the first texture sampler.
	const floatTexturePosition = 0.0

	// // sanity checks
	// originalLen := len(msg)
	// trimmedMsg := msg
	// if charOffset > 0 && charOffset < originalLen {
	// 	// trim the string based on incoming character offset
	// 	trimmedMsg = trimmedMsg[charOffset:]
	// }

	// get the length of our message
	l := len(s)

	// create the arrays to hold the data to buffer to OpenGL
	comboBuffer := make([]float32, 0, l*(2+2+4)*4) // pos, uv, color4
	indexBuffer := make([]uint32, 0, l*6)          // two faces * three indexes

	// do a preliminary test to see how much room the message will take up
	dimX, dimY, advH := f.GetRenderSize(s)

	fontScale := f.GetCurrentScale()

	// loop through the message
	var totalChars int
	var scaledSize float32
	var cursorOverflowRight bool

	var penX = pos[0] - 1
	var penY = pos[1] - f.GlyphHeight*fontScale + advH/4 //float32(advH)
	var chi int
	for _, ch := range s {

		// get the rune data
		chData := f.locations[ch]

		/*
			bounds, _, _ := f.face.GlyphBounds(ch)
			glyphD := bounds.Max.Sub(bounds.Min)
			glyphAdvW, _ := f.face.GlyphAdvance(ch)
			metrics := f.face.Metrics()
			glyphAdvH := float32(metrics.Ascent.Round())

			glyphH := float32(glyphD.Y.Round())
			glyphW := float32(glyphD.X.Round())
			advHeight := glyphAdvH
			advWidth := float32(glyphAdvW.Round())
		*/

		glyphH := f.GlyphHeight
		glyphW := f.GlyphWidth
		// advHeight := chData.advanceHeight
		advWidth := chData.advanceWidth

		// possibly stop here if we're going to overflow the max width
		if maxWidth > 0.0 && scaledSize+(advWidth*fontScale) > maxWidth {
			// we overflowed the size of the string, now check to see if
			// the cursor position is covered within this string or if that hasn't
			// been reached yet.
			if cursorPosition >= 0 && cursorPosition-charOffset > chi {
				// cursorOverflowRight = true
			}

			// adjust the dimX here since we shortened the string
			dimX = scaledSize
			break
		}
		scaledSize += advWidth * fontScale

		// setup the coordinates for ther vetexes
		x0 := penX
		// y0 := penY - (glyphH-advHeight)*fontScale
		y0 := penY
		x1 := x0 + glyphW*fontScale
		y1 := y0 + glyphH*fontScale
		// y1 := y0 + dimY
		s0 := chData.uvMinX
		t0 := chData.uvMinY
		s1 := chData.uvMaxX
		t1 := chData.uvMaxY

		// set the vertex data
		comboBuffer = append(comboBuffer, x1)
		comboBuffer = append(comboBuffer, y0)
		comboBuffer = append(comboBuffer, s1)
		comboBuffer = append(comboBuffer, t0)
		comboBuffer = append(comboBuffer, floatTexturePosition)
		comboBuffer = append(comboBuffer, color[:]...)

		comboBuffer = append(comboBuffer, x1)
		comboBuffer = append(comboBuffer, y1)
		comboBuffer = append(comboBuffer, s1)
		comboBuffer = append(comboBuffer, t1)
		comboBuffer = append(comboBuffer, floatTexturePosition)
		comboBuffer = append(comboBuffer, color[:]...)

		comboBuffer = append(comboBuffer, x0)
		comboBuffer = append(comboBuffer, y1)
		comboBuffer = append(comboBuffer, s0)
		comboBuffer = append(comboBuffer, t1)
		comboBuffer = append(comboBuffer, floatTexturePosition)
		comboBuffer = append(comboBuffer, color[:]...)

		comboBuffer = append(comboBuffer, x0)
		comboBuffer = append(comboBuffer, y0)
		comboBuffer = append(comboBuffer, s0)
		comboBuffer = append(comboBuffer, t0)
		comboBuffer = append(comboBuffer, floatTexturePosition)
		comboBuffer = append(comboBuffer, color[:]...)

		startIndex := uint32(chi) * 4
		indexBuffer = append(indexBuffer, startIndex)
		indexBuffer = append(indexBuffer, startIndex+1)
		indexBuffer = append(indexBuffer, startIndex+2)

		indexBuffer = append(indexBuffer, startIndex+2)
		indexBuffer = append(indexBuffer, startIndex+3)
		indexBuffer = append(indexBuffer, startIndex)

		// advance the pen
		penX += advWidth * fontScale
		totalChars++
		chi++
	}

	return &RenderData{
		ComboBuffer:         comboBuffer,
		IndexBuffer:         indexBuffer,
		Faces:               uint32(totalChars * 2),
		Width:               float32(dimX),
		Height:              float32(dimY),
		AdvanceHeight:       float32(advH),
		CursorOverflowRight: cursorOverflowRight,
	}
}

// loadRGBAToTexture takes a byte slice and throws it into an OpenGL texture.
func loadRGBAToTexture(rgba []byte, imageSize int32) graphics.Texture {
	return loadRGBAToTextureExt(rgba, imageSize, graphics.NEAREST, graphics.NEAREST, graphics.CLAMP_TO_EDGE, graphics.CLAMP_TO_EDGE)
}

// loadRGBAToTextureExt takes a byte slice and throws it into an OpenGL texture.
func loadRGBAToTextureExt(rgba []byte, imageSize, magFilter, minFilter, wrapS, wrapT int32) graphics.Texture {
	tex := gfx.GenTexture()

	gfx.ActiveTexture(graphics.TEXTURE0)
	gfx.BindTexture(graphics.TEXTURE_2D, tex)
	// gfx.GenerateMipmap(graphics.TEXTURE_2D)

	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MAG_FILTER, magFilter)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_MIN_FILTER, minFilter)

	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_S, wrapS)
	gfx.TexParameteri(graphics.TEXTURE_2D, graphics.TEXTURE_WRAP_T, wrapT)

	gfx.TexImage2D(graphics.TEXTURE_2D, 0, graphics.RGBA, imageSize, imageSize, 0, graphics.RGBA, graphics.UNSIGNED_BYTE, gfx.Ptr(rgba), len(rgba))
	return tex
}
