package fizzgui

import (
	"image"
	_ "image/png"
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/graphicsprovider"
)

type TextureID int32

var textures []*TexturePack

type TexturePack struct {
	ID            TextureID
	Width, Height float32
	Texture       graphicsprovider.Texture
}

func AddTexturePackImg(imgPath string) (*TexturePack, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}

	info, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}

	tex, err := fizzle.LoadImageToTexture(imgPath)
	if err != nil {
		return nil, err
	}

	id := TextureID(len(textures) + 1)

	tp := &TexturePack{
		ID:      id,
		Texture: tex,
		Width:   float32(info.Width),
		Height:  float32(info.Height),
	}
	textures = append(textures, tp)
	// textures[id] = tp

	return tp, nil
}

// func AddTexturePack(tex graphicsprovider.Texture) *TexturePack {

// 	return tp
// }

type TextureChunk struct {
	pack   *TexturePack
	Offset mgl32.Vec4
}

func (tp *TexturePack) NewChunk(x0, y0, x1, y1 float32) *TextureChunk {
	x0 = x0 / tp.Width
	y0 = 1 - y0/tp.Height

	x1 = x1 / tp.Width
	y1 = 1 - y1/tp.Height

	tc := &TextureChunk{
		pack:   tp,
		Offset: mgl32.Vec4{x0, y1, x1, y0},
	}
	return tc
}

func LoadImage(imgPath string) (graphicsprovider.Texture, error) {
	return fizzle.LoadImageToTexture(imgPath)
}
