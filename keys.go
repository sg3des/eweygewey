package fizzgui

import (
	"github.com/go-gl/glfw/v3.1/glfw"
)

var Keys *keyboard

type keyboard struct {
	keys  []KeyEvent
	runes []rune

	listening bool

	prevKeyCallback      glfw.KeyCallback
	prevCharModsCallback glfw.CharModsCallback
}

func initKeyboard(window *glfw.Window) {
	Keys = new(keyboard)
	Keys.prevKeyCallback = window.SetKeyCallback(Keys.charKeyCallback)
	Keys.prevCharModsCallback = window.SetCharModsCallback(Keys.charModsCallback)
}

func (kbrd *keyboard) DisableListening() {
	kbrd.listening = false
}

func (kbrd *keyboard) GetKeys() (keys []KeyEvent) {
	kbrd.listening = true
	keys = kbrd.keys
	kbrd.keys = kbrd.keys[:0]
	return
}

//KeyEvent contains keyCode, rune and pressed key modifiers
type KeyEvent struct {
	KeyCode glfw.Key

	Shift bool
	Ctrl  bool
	Alt   bool
	Super bool
}

func (kbrd *keyboard) charKeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

	if kbrd.listening {
		if action != glfw.Press && action != glfw.Repeat {
			return
		}

		k := KeyEvent{
			Shift:   mods == glfw.ModShift,
			Ctrl:    mods == glfw.ModControl,
			Alt:     mods == glfw.ModAlt,
			Super:   mods == glfw.ModSuper,
			KeyCode: key,
		}

		kbrd.keys = append(kbrd.keys, k)
	}

	if kbrd.prevKeyCallback != nil {
		kbrd.prevKeyCallback(w, key, scancode, action, mods)
	}
}

func (kbrd *keyboard) GetRunes() (runes []rune) {
	kbrd.listening = true
	runes = kbrd.runes
	kbrd.runes = kbrd.runes[:0]
	return
}

func (kbrd *keyboard) charModsCallback(w *glfw.Window, char rune, mods glfw.ModifierKey) {

	if kbrd.listening && char > 0 {

		kbrd.runes = append(kbrd.runes, char)
	}

	if kbrd.prevCharModsCallback != nil {
		kbrd.prevCharModsCallback(w, char, mods)
	}
}
