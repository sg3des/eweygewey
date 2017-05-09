// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package main

import (
	"log"
	"runtime"
	"time"

	glfw "github.com/go-gl/glfw/v3.1/glfw"

	"github.com/sg3des/fizzgui"

	fizzle "github.com/tbogdala/fizzle"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
	"github.com/tbogdala/fizzle/graphicsprovider/opengl"
)

const (
	fontScale    = 18
	fontFilepath = "../assets/Roboto-Bold.ttf"

	imgPotions     = "../assets/potions.png"
	imgPotionRed   = "../assets/potion_red.png"
	imgPotionBlack = "../assets/potion_black.png"

	width  = 800
	height = 600
)

var (
	window *glfw.Window
	gfx    graphics.GraphicsProvider
)

// GLFW event handling must run on the main OS thread
func init() {
	runtime.LockOSThread()
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	window, gfx = initGraphics("fizzgui-example", width, height)

	if err := fizzgui.Init(window, gfx); err != nil {
		log.Fatalln("Failed initialize fizzgui, reason:", err)
	}

	// load a font
	_, err := fizzgui.NewFont("Default", fontFilepath, fontScale, fizzgui.FontGlyphs)
	if err != nil {
		log.Fatalln("Failed to load the font file, reason:", err)
	}

	uiPack, err := fizzgui.AddTexturePackImg("../assets/texture.png")
	if err != nil {
		log.Fatalln(err)
	}

	potionsPack, err := fizzgui.AddTexturePackImg("../assets/potions.png")
	if err != nil {
		log.Fatalln(err)
	}

	//left container
	left := fizzgui.NewContainer("left-container", "2%", "2%", "48%", "96%")
	left.AutoAdjustHeight = false

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

	left.NewInput("id", &inp, wgtCallback)

	left.NewCheckbox(&ok, wgtCallback)
	left.NewText("checkbox")

	left.NewRow()
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
	right.AutoAdjustHeight = false

	dad := right.NewDragAndDropGroup("id")
	dad.NewSlot("slot0", "10%", "10%", "12%", "12%", dadCallback)
	dad.NewSlot("slot0", "30%", "10%", "12%", "12%", dadCallback)

	redPotion := potionsPack.NewChunk(62, 122, 118, 178)
	greenPotion := potionsPack.NewChunk(182, 62, 238, 118)
	dad.NewItem("item0", "10%", "30%", "10%", "10%", redPotion, nil)
	dad.NewItem("item1", "30%", "30%", "10%", "10%", greenPotion, nil)

	//start render
	render()
}

var inp string
var ok bool
var progress float32

func wgtCallback(wgt *fizzgui.Widget) {
	log.Println(wgt.Text, inp, ok, progress)
}

func dadCallback(item *fizzgui.DADItem, slot *fizzgui.DADSlot, val interface{}) bool {
	log.Println(item.ID, slot.ID, val)
	return true
}

// initGraphics creates an OpenGL window and initializes the required graphics libraries.
// It will either succeed or panic.
func initGraphics(title string, w int, h int) (*glfw.Window, graphics.GraphicsProvider) {

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

	window.SetKeyCallback(keyCallback)

	return window, gfx
}

func render() {
	// set some additional OpenGL flags
	gfx.BlendEquation(graphics.FUNC_ADD)
	gfx.BlendFunc(graphics.SRC_ALPHA, graphics.ONE_MINUS_SRC_ALPHA)
	gfx.Enable(graphics.BLEND)
	gfx.Enable(graphics.TEXTURE_2D)
	gfx.Enable(graphics.CULL_FACE)

	for !window.ShouldClose() {
		gfx.Viewport(0, 0, width, height)
		gfx.ClearColor(0.4, 0.4, 0.4, 1)
		gfx.Clear(graphics.COLOR_BUFFER_BIT | graphics.DEPTH_BUFFER_BIT)

		// draw the user interface
		fizzgui.Construct()
		// draw the screen and get any input
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
