package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"

	"github.com/tbogdala/fizzle"
	"github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/fizzle/graphicsprovider/opengl"

	"github.com/sg3des/fizzgui"
)

var (
	window *glfw.Window
	gfx    graphicsprovider.GraphicsProvider
)

func init() {
	runtime.LockOSThread()
}

func main() {
	log.SetFlags(log.Lshortfile)

	window, gfx = initGraphics("fizzgui-example", 800, 600)

	if err := fizzgui.Init(window, gfx); err != nil {
		log.Fatalln("Failed initialize fizzgui, reason:", err)
	}

	// load a font
	_, err := fizzgui.NewFont("Default", "../assets/Roboto-Bold.ttf", 18, fizzgui.FontGlyphs)
	if err != nil {
		log.Fatalln("Failed to load the font file, reason:", err)
	}

	uiPack, err := fizzgui.AddTexturePackImg("../assets/texture.png")
	if err != nil {
		log.Fatalln(err)
	}

	//left container
	left := fizzgui.NewContainer("left-container", "2%", "2%", "48%", "96%")
	left.AutoAdjustHeight = true

	left.NewText("full width text").Layout.SetWidth("100%")

	l := left.NewText("left")
	l.Layout.SetWidth("33%")

	c := left.NewText("center")
	c.Layout.SetWidth("33%")
	c.TextAlign = fizzgui.TALIGN_CENTER

	r := left.NewText("right")
	r.Layout.SetWidth("33.9%")
	r.TextAlign = fizzgui.TALIGN_RIGHT

	left.NewRow()

	left.NewButton("text width", wgtCallback)
	left.NewRow()
	left.NewButton("button 50%", wgtCallback).Layout.SetWidth("50%")
	left.NewRow()
	left.NewButton("button full width", wgtCallback).Layout.SetWidth("100%")

	left.NewInput("input0", &inp0, wgtCallback)
	left.NewInput("input1", &inp1, wgtCallback)

	left.NewCheckbox(&ok, wgtCallback)
	left.NewText("checkbox")

	left.NewRow().Layout.SetHeight("20px")
	btn := left.NewButton("Button", wgtCallback)
	btn.Layout.SetWidth("300px")
	btn.Layout.SetHeight("51px")

	normalState := uiPack.NewChunk(550, 250, 852, 302)
	hoverState := uiPack.NewChunk(550, 306, 852, 358)
	btn.Texture = normalState
	btn.Style = fizzgui.NewStyleTexture(normalState)
	btn.StyleHover = fizzgui.NewStyleTexture(hoverState)

	//progressbar
	left.NewProgressBar(&progress, 0, 100, func(btn *fizzgui.Widget) {
		log.Println("PROGRESS:", progress)
		progress = 0
	})
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			progress += 0.1
		}
	}()

	//right container
	right := fizzgui.NewContainer("right-container", "50%", "2%", "48%", "96%")

	dad := right.NewDragAndDropGroup("id")
	dad.NewSlot("slot0", "10%", "10%", "12%", "12%", dadCallback)
	dad.NewSlot("slot0", "30%", "10%", "12%", "12%", dadCallback)

	red := "../assets/red.png"
	green := "../assets/green.png"
	dad.NewItem("item0", "10%", "30%", "10%", "10%", red, "red")
	dad.NewItem("item1", "30%", "30%", "10%", "10%", green, "green")

	//start render
	renderLoop()
}

var inp0 string
var inp1 string
var ok bool
var progress float32

func wgtCallback(wgt *fizzgui.Widget) {
	fmt.Println(wgt.Text, inp0, inp1, ok, progress)
}

func dadCallback(item *fizzgui.DADItem, slot *fizzgui.DADSlot, val interface{}) bool {
	fmt.Println(item.ID, slot.ID, val)
	return true
}

func renderLoop() {
	for !window.ShouldClose() {
		w, h := window.GetFramebufferSize()
		gfx.Viewport(0, 0, int32(w), int32(h))
		gfx.ClearColor(0.4, 0.4, 0.4, 1)
		gfx.Clear(graphicsprovider.COLOR_BUFFER_BIT | graphicsprovider.DEPTH_BUFFER_BIT)

		// draw the user interface
		fizzgui.Construct()

		// draw the screen and get any input
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) (*glfw.Window, graphicsprovider.GraphicsProvider) {

	err := glfw.Init()
	if err != nil {
		panic("Can't init glfw! " + err.Error())
	}

	// request a OpenGL 3.3 core context
	glfw.WindowHint(glfw.Samples, 0)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// do the actual window creation
	window, err := glfw.CreateWindow(w, h, title, nil, nil)
	if err != nil {
		panic("Failed to create the main window! " + err.Error())
	}

	window.MakeContextCurrent()

	glfw.SwapInterval(1) // if 0 disable v-sync

	// initialize OpenGL
	gfx, err := opengl.InitOpenGL()
	if err != nil {
		panic("Failed to initialize OpenGL! " + err.Error())
	}
	fizzle.SetGraphics(gfx)

	// set some additional OpenGL flags
	gfx.BlendEquation(graphicsprovider.FUNC_ADD)
	gfx.BlendFunc(graphicsprovider.SRC_ALPHA, graphicsprovider.ONE_MINUS_SRC_ALPHA)
	gfx.Enable(graphicsprovider.BLEND)
	gfx.Enable(graphicsprovider.TEXTURE_2D)
	gfx.Enable(graphicsprovider.CULL_FACE)

	window.SetKeyCallback(keyCallback)

	return window, gfx
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}
