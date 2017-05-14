// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

import (
	"log"
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/fizzle/renderer/forward"
)

var (
	window *glfw.Window
	gfx    graphics.GraphicsProvider

	mainShader graphics.Program

	comboBuffer []float32
	indexBuffer []uint32
	comboVBO    graphics.Buffer
	indexVBO    graphics.Buffer
	vao         uint32
	faceCount   uint32

	frameTime time.Time
	dt        float32

	whitePixelUv = mgl.Vec4{1, 1, 1, 1}
	imagePixelUv = mgl.Vec4{0, 0, 1, 1}

	fonts map[string]*Font

	wndLayout *Layout

	containers []*Container

	ActiveWidget   *Widget
	HoverWidget    *Widget
	HoverContainer *Container

	renderer *forward.ForwardRenderer
)

//Init gui
func Init(glfwWindow *glfw.Window, graphProv graphics.GraphicsProvider) error {
	window = glfwWindow
	gfx = graphProv

	fonts = make(map[string]*Font)
	frameTime = time.Now()

	vao = gfx.GenVertexArray()
	comboVBO = gfx.GenBuffer()
	indexVBO = gfx.GenBuffer()

	var err error
	mainShader, err = compileShader(ShaderV, ShaderF)
	if err != nil {
		return err
	}

	wndLayout = &Layout{}
	updateWindowLayout()
	initMouse(window)
	initKeyboard(window)

	initDefaultStyles()

	return nil
}

func updateWindowLayout() {
	w, h := window.GetSize()
	wndLayout.X = 0
	wndLayout.Y = float32(h) // for top left anchor
	wndLayout.W = float32(w)
	wndLayout.H = float32(h)
}

func DelContainer(ptr *Container) {
	for i, c := range containers {
		if c == ptr {
			containers[i] = nil
			containers = append(containers[:i], containers[i+1:]...)
			return
		}
	}

	log.Println("WARNING: container not found")
}

// NewFont loads the font from a file and 'registers' it with the UI manager.
func NewFont(name string, fontFilepath string, scaleInt int, glyphs string) (*Font, error) {
	f, err := newFont(fontFilepath, scaleInt, glyphs)

	// if we succeeded, store the font with the name specified
	if err == nil {
		fonts[name] = f
	}

	return f, err
}

// GetFont attempts to get the font by name from the Manager's collection
// It returns the font on success or nil on failure.
func GetFont(name string) *Font {
	return fonts[name]
}

// Construct loops through all of the Windows in the Manager and creates all of the widgets and their data.
// This function does not buffer the result to VBO or do the actual rendering -- call Draw() for that.
func Construct() {
	// fmt.Println("===============================")
	// reset the display data
	comboBuffer = comboBuffer[:0]
	indexBuffer = indexBuffer[:0]
	faceCount = 0
	zcmds = make(map[uint8][]*cmdList)

	// textureStack = textureStack[:0]
	t := time.Now()
	dt = float32(t.Sub(frameTime).Seconds())
	frameTime = t

	Mouse.Update()
	updateWindowLayout()

	HoverContainer = nil
	for _, c := range containers {
		if c.Layout.ContainsPoint(Mouse.X, Mouse.Y) {
			HoverContainer = c
			break
		}
	}

	HoverWidget = nil
C:
	for _, c := range containers {
		for _, wgt := range c.Widgets {
			if wgt.Layout.ContainsPoint(Mouse.X, Mouse.Y) {
				HoverWidget = wgt
				break C
			}
		}
	}

	for _, c := range containers {
		if c != nil {
			c.construct()
		}
	}

	render()
}

func render() {
	const floatSize = 4
	const uintSize = 4

	const minZDepth = -100
	const maxZDepth = 100

	// gfx.Disable(graphics.DEPTH_TEST)
	// gfx.Enable(graphics.SCISSOR_TEST)

	var startIndex uint32
	var z uint8
	for z = 0; z < 255; z++ {
		cmds, ok := zcmds[z]
		if !ok {
			continue
		}
		for _, cmd := range cmds {
			if cmd.faceCount == 0 {
				continue
			}
			comboBuffer = append(comboBuffer, cmd.comboBuffer...)

			// reindex the index buffer to reference the correct vertex data
			highestIndex := uint32(0)
			for _, i := range cmd.indexBuffer {
				if i > highestIndex {
					highestIndex = i
				}
				indexBuffer = append(indexBuffer, i+startIndex)
			}
			faceCount += cmd.faceCount
			startIndex += highestIndex + 1
		}
	}

	// make sure that we're going to draw something
	if startIndex == 0 {
		return
	}

	gfx.BindVertexArray(vao)
	view := mgl.Ortho(0.5, wndLayout.W+0.5, 0.5, wndLayout.H+0.5, minZDepth, maxZDepth)

	// buffer the data
	gfx.BindBuffer(graphics.ARRAY_BUFFER, comboVBO)
	gfx.BufferData(graphics.ARRAY_BUFFER, floatSize*len(comboBuffer), gfx.Ptr(&comboBuffer[0]), graphics.STREAM_DRAW)
	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, indexVBO)
	gfx.BufferData(graphics.ELEMENT_ARRAY_BUFFER, uintSize*len(indexBuffer), gfx.Ptr(&indexBuffer[0]), graphics.STREAM_DRAW)

	// texID = make(map[graphics.Texture]int32)

	bindShader(view)

	// var prevTex graphics.Texture
	var indexOffset int
	for z = 0; z < 255; z++ {
		cmds, ok := zcmds[z]
		if !ok {
			continue
		}

		for _, cmd := range cmds {
			if cmd.faceCount == 0 {
				continue
			}
			// gfx.Scissor(0, 0, int32(wndLayout.W), int32(wndLayout.H))

			// TEX := gfx.GetUniformLocation(mainShader, "TEX")
			gfx.BindTexture(graphics.TEXTURE_2D, cmd.texture)
			// if prevTex != cmd.texture {
			// 	prevTex = cmd.texture
			// 	bindShader(view, cmd.texture)
			// }

			gfx.Viewport(0, 0, int32(wndLayout.W), int32(wndLayout.H))
			gfx.DrawElements(graphics.TRIANGLES, int32(cmd.faceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(indexOffset*uintSize))
			indexOffset += int(cmd.faceCount) * 3
		}
	}

	gfx.BindVertexArray(0)

	// gfx.Disable(graphics.SCISSOR_TEST)
	// gfx.Enable(graphics.DEPTH_TEST)
}
