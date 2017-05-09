# FizzGUI

FizzGUI is and OpenGL GUI for [Fizzle][fizzle] engine, сonstructed from [EweyGewey][EweyGewey], but heavily reworked. 


UNDER CONSTRUCTION
==================

At present, it is very much in an alpha stage with new development adding in
features, widgets and possibly API breaks. Any API break should increment the
minor version number and any patch release tags should remain compatible even
in development 0.x versions.

Screenshots
-----------

Here's some of what's available right now in the [example][example]:

![screenshot][screenshot]


Requirements
------------

* [Mathgl][mgl32] - for 3d math
* [Freetype][freetype] - for dynamic font texture generation
* [Fizzle][fizzle] - provides an OpenGL 3/es2/es3 abstraction
* [GLFW][glfw] (v3.1) - currently GLFW is the only 'host' support for input


Differences
-----------

* Windows replaces with Containers, containers can not move and not have a title, scrollbars(current) not available too.
* Containers may create various widgets, widgets placed one by one, if in row not enough space, widget move to new row(in html it`s looks like a *float*).
* Widget can be have fixed position
* Smart layout system for positioning containers and widgets 
* Some widgets may have callbacks(signals) calls on events(ex: press button)


Current Features
----------------

* Containers
    * Text
    * Input text
    * Button
    * Checkbox
    * Progressbar
    * Images
    * Drag and Drop system


TODO
----

The following need to be addressed in order to start releases:

* more widgets:
    * text wrapping
    * multi-line text editors
    * combobox
* editbox cursor doesn't start where mouse was clicked
* and more other


LICENSE
=======

Original package [EweyGewey][EweyGewey] is released under the BSD license. See the [LICENSE][license-link] file for more details.


[EweyGewey]: https://github.com/tbogdala/eweygewey
[golang]: https://golang.org/
[fizzle]: https://github.com/tbogdala/fizzle
[glfw]: https://github.com/go-gl/glfw
[mgl]: https://github.com/go-gl/mathgl
[freetype]: https://github.com/golang/freetype


[screenshot]: examples/screenshots/example.png
[example]: examples/new/example.go

[license-link]: https://raw.githubusercontent.com/tbogdala/eweygewey/master/LICENSE
