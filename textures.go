package fizzgui

import (
	"image"
	_ "image/png"
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/graphicsprovider"
)

type TexturePack struct {
	Width, Height float32
	Tex           graphicsprovider.Texture
}

func NewTexturePack(img string) (*TexturePack, error) {
	f, err := os.Open(img)
	if err != nil {
		return nil, err
	}

	info, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}

	tex, err := fizzle.LoadImageToTexture(img)
	if err != nil {
		return nil, err
	}

	tp := &TexturePack{
		Tex:    tex,
		Width:  float32(info.Width),
		Height: float32(info.Height),
	}

	return tp, nil
}

type Texture struct {
	Tex    graphicsprovider.Texture
	Offset mgl32.Vec4
}

func (tp *TexturePack) NewChunk(x0, y0, x1, y1 float32) *Texture {
	if x0 > x1 {
		x1, x0 = x0, x1
	}
	if y0 < y1 {
		y1, y0 = y0, y1
	}

	x0 = x0 / tp.Width
	y0 = 1 - y0/tp.Height

	x1 = x1 / tp.Width
	y1 = 1 - y1/tp.Height

	tc := &Texture{
		Tex:    tp.Tex,
		Offset: mgl32.Vec4{x0, y0, x1, y1},
	}
	return tc
}

func NewTextureImg(img string) (*Texture, error) {
	tex, err := fizzle.LoadImageToTexture(img)
	if err != nil {
		return nil, err
	}
	return &Texture{tex, imagePixelUv}, nil
}
