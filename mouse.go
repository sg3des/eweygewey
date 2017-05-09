// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

import (
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// constants used for polling the state of a mouse button
const (
	MouseDown = iota
	MouseUp
	MouseClick
	MouseDoubleClick
)

var Mouse *mouse

type mouse struct {
	window *glfw.Window

	frameTime time.Time
	dt        float32

	X float32
	Y float32

	prevScrollCallback glfw.ScrollCallback
	ScrollSpeed        float32
	ScrollDelta        float32

	doubleClickThreshold float64

	buttonsTracker map[int]mouseButtonData
}

func initMouse(window *glfw.Window) {
	Mouse = &mouse{
		window:               window,
		ScrollSpeed:          10,
		doubleClickThreshold: 0.5,
		buttonsTracker:       make(map[int]mouseButtonData),
	}

	Mouse.prevScrollCallback = window.SetScrollCallback(Mouse.scollCallback)
}

func (m *mouse) scollCallback(w *glfw.Window, xoff float64, yoff float64) {
	m.ScrollDelta += float32(yoff) * m.ScrollSpeed

	if m.prevScrollCallback != nil {
		m.prevScrollCallback(w, xoff, yoff)
	}
}

//Update should be call each frame
func (m *mouse) Update() {

	t := time.Now()
	m.dt = float32(t.Sub(m.frameTime).Seconds())
	m.frameTime = t

	_, wy := m.window.GetSize()
	mx, my := m.window.GetCursorPos()
	m.X = float32(mx)
	m.Y = float32(wy) - float32(my)
}

//GetPosition of mouse
func (m *mouse) GetPosition() (x, y float32) {
	return m.X, m.Y
}

// mouseButtonData is used to track button presses between frames.
type mouseButtonData struct {
	// lastPress is the time the last UP->DOWN transition took place
	lastPress time.Time

	// lastPressLocation is the position of the mouse when the last UP->DOWN
	// transition took place
	lastPressLocation mgl32.Vec2

	// lastAction was the last detected action for the button
	lastAction int

	// doubleClickDetected should be set to true if the last UP->DOWN->UP
	// sequence was fast enough to be a double click.
	doubleClickDetected bool

	// lastCheckedAt should be set to the time the functions last checked
	// for action. This way, input can be polled only once per frame.
	lastCheckedAt time.Time
}

//GetButtonAction
func (m *mouse) GetButtonAction(button int) int {
	var action int
	var mbData mouseButtonData
	var tracked bool

	// get the mouse button data and return the stale result if we're
	// in the same frame.
	mbData, tracked = m.buttonsTracker[button]
	if tracked == true && mbData.lastCheckedAt == m.frameTime {
		return mbData.lastAction
	}

	// poll the button action
	glfwAction := window.GetMouseButton(glfw.MouseButton(int(glfw.MouseButton1) + button))
	if glfwAction == glfw.Release {
		action = MouseUp
	} else if glfwAction == glfw.Press {
		action = MouseDown
	} else if glfwAction == glfw.Repeat {
		action = MouseDown
	}

	// see if we're tracking this button yet
	if tracked == false {
		// create a new mouse button tracker data object
		if action == MouseDown {
			// mx, my := uiman.GetMousePosition()
			mbData.lastPressLocation = mgl32.Vec2{m.X, m.Y}
			mbData.lastPress = m.frameTime
		} else {
			mbData.lastPress = time.Unix(0, 0)
		}
	} else {
		if action == MouseDown {
			// check to see if there was a transition from UP to DOWN
			if mbData.lastAction == MouseUp {
				// check to see the time between the last UP->DOWN transition
				// and this one. If it's less than the double click threshold
				// then change the doubleClickDetected member so that the
				// next DOWN->UP will return a double click instead.
				if m.frameTime.Sub(mbData.lastPress).Seconds() < m.doubleClickThreshold {
					mbData.doubleClickDetected = true
				}

				// count this as a press and log the time
				// mx, my := uiman.GetMousePosition()
				mbData.lastPressLocation = mgl32.Vec2{m.X, m.Y}
				mbData.lastPress = m.frameTime
			}
		} else {
			// check to see if there was a transition from DOWN to UP
			if mbData.lastAction == MouseDown {
				if mbData.doubleClickDetected {
					// return the double click
					action = MouseDoubleClick

					// reset the tracker
					mbData.doubleClickDetected = false
				} else {
					// return the single click
					action = MouseClick
				}
			}
		}
	}

	// put the updated data back into the map and return the action
	mbData.lastAction = action
	mbData.lastCheckedAt = m.frameTime

	m.buttonsTracker[button] = mbData
	return action
}

// // SetInputHandlers sets the input callbacks for the GUI Manager to work with
// // GLFW. This function takes advantage of closures to track input across
// // multiple calls.
// func setInputHandlers(window *glfw.Window) {

// 	// uiman.GetMouseDownPosition = func(button int) (float32, float32) {
// 	// 	// test to see if we polled the delta this frame
// 	// 	if needsMousePosCheck {
// 	// 		// if not, then update the location data
// 	// 		uiman.GetMousePosition()
// 	// 	}

// 	// 	// is the mouse button down?
// 	// 	if uiman.GetMouseButtonAction(button) != MouseUp {
// 	// 		var tracked bool
// 	// 		var mbData mouseButtonData

// 	// 		// get the mouse button data and return the stale result if we're
// 	// 		// in the same frame.
// 	// 		mbData, tracked = mouseButtonTracker[button]
// 	// 		if tracked == true {
// 	// 			return mbData.lastPressLocation[0], mbData.lastPressLocation[1]
// 	// 		}
// 	// 	}

// 	// 	// mouse not down or not tracked.
// 	// 	return -1.0, -1.0
// 	// }

// 	// create our own handler for the scroll wheel which then passes the
// 	// correct data to our own scroll wheel handler function

// 	// stores all of the key press events
// 	keyBuffer := []KeyPressEvent{}

// 	// make a translation table from GLFW->EweyGewey key codes
// 	keyTranslation := make(map[glfw.Key]int)
// 	keyTranslation[glfw.KeyWorld1] = EweyKeyWorld1
// 	keyTranslation[glfw.KeyWorld2] = EweyKeyWorld2
// 	keyTranslation[glfw.KeyEscape] = EweyKeyEscape
// 	keyTranslation[glfw.KeyEnter] = EweyKeyEnter
// 	keyTranslation[glfw.KeyTab] = EweyKeyTab
// 	keyTranslation[glfw.KeyBackspace] = EweyKeyBackspace
// 	keyTranslation[glfw.KeyInsert] = EweyKeyInsert
// 	keyTranslation[glfw.KeyDelete] = EweyKeyDelete
// 	keyTranslation[glfw.KeyRight] = EweyKeyRight
// 	keyTranslation[glfw.KeyLeft] = EweyKeyLeft
// 	keyTranslation[glfw.KeyDown] = EweyKeyDown
// 	keyTranslation[glfw.KeyUp] = EweyKeyUp
// 	keyTranslation[glfw.KeyPageUp] = EweyKeyPageUp
// 	keyTranslation[glfw.KeyPageDown] = EweyKeyPageDown
// 	keyTranslation[glfw.KeyHome] = EweyKeyHome
// 	keyTranslation[glfw.KeyEnd] = EweyKeyEnd
// 	keyTranslation[glfw.KeyCapsLock] = EweyKeyCapsLock
// 	keyTranslation[glfw.KeyNumLock] = EweyKeyNumLock
// 	keyTranslation[glfw.KeyPrintScreen] = EweyKeyPrintScreen
// 	keyTranslation[glfw.KeyPause] = EweyKeyPause
// 	keyTranslation[glfw.KeyF1] = EweyKeyF1
// 	keyTranslation[glfw.KeyF2] = EweyKeyF2
// 	keyTranslation[glfw.KeyF3] = EweyKeyF3
// 	keyTranslation[glfw.KeyF4] = EweyKeyF4
// 	keyTranslation[glfw.KeyF5] = EweyKeyF5
// 	keyTranslation[glfw.KeyF6] = EweyKeyF6
// 	keyTranslation[glfw.KeyF7] = EweyKeyF7
// 	keyTranslation[glfw.KeyF8] = EweyKeyF8
// 	keyTranslation[glfw.KeyF9] = EweyKeyF9
// 	keyTranslation[glfw.KeyF10] = EweyKeyF10
// 	keyTranslation[glfw.KeyF11] = EweyKeyF11
// 	keyTranslation[glfw.KeyF12] = EweyKeyF12
// 	keyTranslation[glfw.KeyF13] = EweyKeyF13
// 	keyTranslation[glfw.KeyF14] = EweyKeyF14
// 	keyTranslation[glfw.KeyF15] = EweyKeyF15
// 	keyTranslation[glfw.KeyF16] = EweyKeyF16
// 	keyTranslation[glfw.KeyF17] = EweyKeyF17
// 	keyTranslation[glfw.KeyF18] = EweyKeyF18
// 	keyTranslation[glfw.KeyF19] = EweyKeyF19
// 	keyTranslation[glfw.KeyF20] = EweyKeyF20
// 	keyTranslation[glfw.KeyF21] = EweyKeyF21
// 	keyTranslation[glfw.KeyF22] = EweyKeyF22
// 	keyTranslation[glfw.KeyF23] = EweyKeyF23
// 	keyTranslation[glfw.KeyF24] = EweyKeyF24
// 	keyTranslation[glfw.KeyF25] = EweyKeyF25
// 	keyTranslation[glfw.KeyLeftShift] = EweyKeyLeftShift
// 	keyTranslation[glfw.KeyLeftAlt] = EweyKeyLeftAlt
// 	keyTranslation[glfw.KeyLeftControl] = EweyKeyLeftControl
// 	keyTranslation[glfw.KeyLeftSuper] = EweyKeyLeftSuper
// 	keyTranslation[glfw.KeyRightShift] = EweyKeyRightShift
// 	keyTranslation[glfw.KeyRightAlt] = EweyKeyRightAlt
// 	keyTranslation[glfw.KeyRightControl] = EweyKeyRightControl
// 	keyTranslation[glfw.KeyRightSuper] = EweyKeyRightSuper

// 	//keyTranslation[glfw.Key] = EweyKey

// 	// create our own handler for key input so that it can buffer the keys
// 	// and then consume them in an edit box or whatever widget has focus.
// 	var prevKeyCallback glfw.KeyCallback
// 	var prevCharModsCallback glfw.CharModsCallback
// 	prevKeyCallback = window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
// 		if action != glfw.Press && action != glfw.Repeat {
// 			return
// 		}

// 		// we have a new event, so init the structure
// 		var kpe KeyPressEvent

// 		// try to look it up in the translation table; if it exists, then we log
// 		// the event; if it doesn't exist, then we assume it will be caught by
// 		// the CharMods callback.
// 		code, okay := keyTranslation[key]
// 		if okay == false {
// 			// there are some exceptions to this that will get implemented here.
// 			// when ctrl is held down, it doesn't appear that runes get sent
// 			// through the CharModsCallback function, so we must handle the
// 			// ones we want here.
// 			if (key == glfw.KeyV) && (mods&glfw.ModControl == glfw.ModControl) {
// 				kpe.Rune = 'V'
// 				kpe.IsRune = true
// 				kpe.CtrlDown = true
// 			} else {
// 				return
// 			}
// 		} else {
// 			kpe.KeyCode = code

// 			// set the modifier flags
// 			if mods&glfw.ModShift == glfw.ModShift {
// 				kpe.ShiftDown = true
// 			}
// 			if mods&glfw.ModAlt == glfw.ModAlt {
// 				kpe.AltDown = true
// 			}
// 			if mods&glfw.ModControl == glfw.ModControl {
// 				kpe.CtrlDown = true
// 			}
// 			if mods&glfw.ModSuper == glfw.ModSuper {
// 				kpe.SuperDown = true
// 			}
// 		}

// 		// add it to the keys that have been buffered
// 		keyBuffer = append(keyBuffer, kpe)

// 		// if there was a pre-existing callback, we'll chain it here
// 		if prevKeyCallback != nil {
// 			prevKeyCallback(w, key, scancode, action, mods)
// 		}
// 	})

// 	window.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {
// 		var kpe KeyPressEvent
// 		//fmt.Printf("SetCharModsCallback Rune: %v | mods:%v | ctrl: %v\n", char, mods, mods&glfw.ModControl)

// 		// set the character
// 		kpe.Rune = char
// 		kpe.IsRune = true

// 		// set the modifier flags
// 		if mods&glfw.ModShift == glfw.ModShift {
// 			kpe.ShiftDown = true
// 		}
// 		if mods&glfw.ModAlt == glfw.ModAlt {
// 			kpe.AltDown = true
// 		}
// 		if mods&glfw.ModControl == glfw.ModControl {
// 			kpe.CtrlDown = true
// 		}
// 		if mods&glfw.ModSuper == glfw.ModSuper {
// 			kpe.SuperDown = true
// 		}

// 		// add it to the keys that have been buffered
// 		keyBuffer = append(keyBuffer, kpe)

// 		// if there was a pre-existing callback, we'll chain it here
// 		if prevCharModsCallback != nil {
// 			prevCharModsCallback(w, char, mods)
// 		}
// 	})

// 	uiman.GetKeyEvents = func() []KeyPressEvent {
// 		returnVal := keyBuffer
// 		keyBuffer = keyBuffer[:0]
// 		return returnVal
// 	}

// 	uiman.ClearKeyEvents = func() {
// 		keyBuffer = keyBuffer[:0]
// 	}

// 	uiman.GetClipboardString = func() (string, error) {
// 		return window.GetClipboardString()
// 	}

// 	uiman.SetClipboardString = func(clippy string) {
// 		window.SetClipboardString(clippy)
// 	}
// }
