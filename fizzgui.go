// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

var (
	window *glfw.Window
	gfx    graphics.GraphicsProvider

	mainShader  graphics.Program
	imageShader graphics.Program

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
)

//Init gui
func Init(glfwWindow *glfw.Window, graphProv graphics.GraphicsProvider) error {
	window = glfwWindow
	gfx = graphProv

	fonts = make(map[string]*Font)
	frameTime = time.Now()

	vao = gfx.GenVertexArray()

	var err error
	mainShader, err = compileShader(MainVertShader, MainFragShader)
	if err != nil {
		return err
	}

	imageShader, err = compileShader(ImageVerShader, ImageFragShader)
	if err != nil {
		return err
	}

	comboVBO = gfx.GenBuffer()
	indexVBO = gfx.GenBuffer()

	wndLayout = &Layout{}
	updateWindowLayout()
	initMouse(window)
	initKeyboard(window)

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

// GetContainer returns a container based on the id string passed in
func GetContainer(id string) *Container {
	for _, c := range containers {
		if c.ID == id {
			return c
		}
	}

	return nil
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
	// reset the display data
	comboBuffer = comboBuffer[:0]
	indexBuffer = indexBuffer[:0]
	faceCount = 0

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

// render buffers the UI vertex data into the rendering pipeline and does
// the actual render call.
func render() {
	const floatSize = 4
	const uintSize = 4
	const posOffset = 0
	const uvOffset = floatSize * 2
	const texIdxOffset = floatSize * 4
	const colorOffset = floatSize * 5
	const VBOStride = floatSize * (2 + 2 + 1 + 4) // vert / uv / texIndex / color

	// FIXME: move the zdepth definitions elsewhere
	const minZDepth = -100.0
	const maxZDepth = 100.0

	gfx.Disable(graphics.DEPTH_TEST)
	gfx.Enable(graphics.SCISSOR_TEST)

	// for now, loop through all of the windows and copy all of the data into the manager's buffer
	// FIXME: this could be buffered straight from the cmdList
	var startIndex uint32
	for _, c := range containers {
		var z uint8
		for ; z < 255; z++ {
			cmds, ok := c.zcmds[z]
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

	// this should be set to true when the uniforms and attributes, etc... need to be rebound
	// needRebinding := true

	// loop through the windows and each window's draw cmd list
	indexOffset := uint32(0)
	for _, c := range containers {
		var z uint8
		for ; z < 255; z++ {
			cmds, ok := c.zcmds[z]
			if !ok {
				continue
			}

			for _, cmd := range cmds {
				if cmd.faceCount == 0 {
					continue
				}

				// gfx.Scissor(int32(cmd.clipRect.TLX), int32(cmd.clipRect.BRY), int32(cmd.clipRect.W), int32(cmd.clipRect.H))

				if cmd.image > 0 {
					bindImageShader(view, cmd.image)
				} else {
					bindMainShader(view)
				}

				gfx.Viewport(0, 0, int32(wndLayout.W), int32(wndLayout.H))

				gfx.DrawElements(graphics.TRIANGLES, int32(cmd.faceCount*3), graphics.UNSIGNED_INT, gfx.PtrOffset(int(indexOffset)*uintSize))
				indexOffset += cmd.faceCount * 3

			}
		}
	}

	// log.Println(needRebinding)

	gfx.BindVertexArray(0)
	gfx.Disable(graphics.SCISSOR_TEST)
	gfx.Enable(graphics.DEPTH_TEST)
}

func compileShader(vertShader, fragShader string) (graphics.Program, error) {
	// create the program
	prog := gfx.CreateProgram()

	// create the vertex shader
	var status int32
	vs := gfx.CreateShader(graphics.VERTEX_SHADER)
	gfx.ShaderSource(vs, vertShader)
	gfx.CompileShader(vs)
	gfx.GetShaderiv(vs, graphics.COMPILE_STATUS, &status)
	if status == graphics.FALSE {
		log := gfx.GetShaderInfoLog(vs)
		return 0, fmt.Errorf("Failed to compile the vertex shader:\n%s", log)
	}
	defer gfx.DeleteShader(vs)

	// create the fragment shader
	fs := gfx.CreateShader(graphics.FRAGMENT_SHADER)
	gfx.ShaderSource(fs, fragShader)
	gfx.CompileShader(fs)
	gfx.GetShaderiv(fs, graphics.COMPILE_STATUS, &status)
	if status == graphics.FALSE {
		log := gfx.GetShaderInfoLog(fs)
		return 0, fmt.Errorf("Failed to compile the fragment shader:\n%s", log)
	}
	defer gfx.DeleteShader(fs)

	// attach the shaders to the program and link
	gfx.AttachShader(prog, vs)
	gfx.AttachShader(prog, fs)
	gfx.LinkProgram(prog)
	gfx.GetProgramiv(prog, graphics.LINK_STATUS, &status)
	if status == graphics.FALSE {
		log := gfx.GetProgramInfoLog(prog)
		return 0, fmt.Errorf("Failed to link the program!\n%s", log)
	}

	return prog, nil
}
