// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

import (
	mgl "github.com/go-gl/mathgl/mgl32"
)

var (
	// VertShader330 is the GLSL vertex shader program for the user interface.
	VertShader330 = `#version 330
  uniform mat4 VIEW;
  in vec2 VERTEX_POSITION;
  in vec2 VERTEX_UV;
  in float VERTEX_TEXTURE_INDEX;
  in vec4 VERTEX_COLOR;
  out vec2 vs_uv;
  out vec4 vs_color;
  out float vs_tex_index;
  void main()
  {
    vs_uv = VERTEX_UV;
    vs_color = VERTEX_COLOR;
    vs_tex_index = VERTEX_TEXTURE_INDEX;
    gl_Position = VIEW * vec4(VERTEX_POSITION, 0.0, 1.0);
  }`

	FragShader330 = `#version 330
  uniform sampler2D TEX[4];
  in vec2 vs_uv;
  in vec4 vs_color;
  in float vs_tex_index;
  out vec4 frag_color;
  void main()
  {
    switch(int(vs_tex_index))
    {
      case 0: frag_color = vs_color * texture(TEX[0], vs_uv).rgba; break;
      case 1: frag_color = vs_color * texture(TEX[1], vs_uv).rgba; break;
      case 2: frag_color = vs_color * texture(TEX[2], vs_uv).rgba; break;
      case 3: frag_color = vs_color * texture(TEX[3], vs_uv).rgba; break;
    }

  }`

	// FragShader330 is the GLSL fragment shader program for the user interface.
	// NOTE: 4 samplers is a hardcoded value now, but there's no reason it has to be that specifically.
	// FragShader330 = `#version 330
	//  uniform sampler2D TEX[4];
	//  in vec2 vs_uv;
	//  in vec4 vs_color;
	//  in float vs_tex_index;
	//  out vec4 frag_color;
	//  void main()
	//  {
	//    switch(int(vs_tex_index))
	//    {
	//      case 0: frag_color = vs_color * texture(TEX[0], vs_uv).rgba; break;
	//      case 1: frag_color = vs_color * texture(TEX[1], vs_uv).rgba; break;
	//      case 2: frag_color = vs_color * texture(TEX[2], vs_uv).rgba; break;
	//      case 3: frag_color = vs_color * texture(TEX[3], vs_uv).rgba; break;
	//    }

	//  }`
)

// Color takes the color parameters as integers and returns them
// as a float vector.
func Color(r, g, b, a int) mgl.Vec4 {
	return mgl.Vec4{float32(r) / 255.0, float32(g) / 255.0, float32(b) / 255.0, float32(a) / 255.0}
}
