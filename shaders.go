// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

var ShaderV = `#version 330
uniform mat4 VIEW;

in vec2 VERTEX_POSITION;
in vec2 VERTEX_UV;
in vec4 VERTEX_COLOR;

out vec2 uv;
out vec4 color;

void main() {
	uv = VERTEX_UV;
	color = VERTEX_COLOR;
	gl_Position = VIEW * vec4(VERTEX_POSITION, 0, 1);
}`

var ShaderF = `#version 330
uniform sampler2D TEX;
in vec2 uv;
in vec4 color;
out vec4 frag_color;

void main() {
	vec4 sum = vec4(0.0);

	sum += texture(TEX, vec2(uv.x - 4.0, uv.y - 4.0)) * 0.0162162162;
  sum += texture(TEX, vec2(uv.x - 3.0, uv.y - 3.0)) * 0.0540540541;
  sum += texture(TEX, vec2(uv.x - 2.0, uv.y - 2.0)) * 0.1216216216;
  sum += texture(TEX, vec2(uv.x - 1.0, uv.y - 1.0)) * 0.1945945946;

  sum += texture(TEX, vec2(uv.x, uv.y)) * 0.2270270270;

  sum += texture(TEX, vec2(uv.x + 1.0, uv.y + 1.0)) * 0.1945945946;
  sum += texture(TEX, vec2(uv.x + 2.0, uv.y + 2.0)) * 0.1216216216;
  sum += texture(TEX, vec2(uv.x + 3.0, uv.y + 3.0)) * 0.0540540541;
  sum += texture(TEX, vec2(uv.x + 4.0, uv.y + 4.0)) * 0.0162162162;

  frag_color = color * vec4(sum.rgb*1.5, texture(TEX, uv).a);
	//frag_color = color * texture(TEX, uv);
}`

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

func bindShader(view mgl.Mat4) {
	const posOffset = 0
	const uvOffset = 8
	const colorOffset = 20
	const VBOStride = 36

	gfx.UseProgram(mainShader)
	gfx.BindVertexArray(vao)

	TEX := gfx.GetUniformLocation(mainShader, "TEX")
	gfx.ActiveTexture(graphics.TEXTURE0)
	// gfx.BindTexture(graphics.TEXTURE_2D, tex)
	gfx.Uniform1i(TEX, 0)

	// bind the uniforms and attributes
	VIEW := gfx.GetUniformLocation(mainShader, "VIEW")
	gfx.UniformMatrix4fv(VIEW, 1, false, view)

	VERTEX_POSITION := gfx.GetAttribLocation(mainShader, "VERTEX_POSITION")
	gfx.BindBuffer(graphics.ARRAY_BUFFER, comboVBO)
	gfx.EnableVertexAttribArray(uint32(VERTEX_POSITION))
	gfx.VertexAttribPointer(uint32(VERTEX_POSITION), 2, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(posOffset))

	VERTEX_UV := gfx.GetAttribLocation(mainShader, "VERTEX_UV")
	gfx.EnableVertexAttribArray(uint32(VERTEX_UV))
	gfx.VertexAttribPointer(uint32(VERTEX_UV), 2, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(uvOffset))

	VERTEX_COLOR := gfx.GetAttribLocation(mainShader, "VERTEX_COLOR")
	gfx.EnableVertexAttribArray(uint32(VERTEX_COLOR))
	gfx.VertexAttribPointer(uint32(VERTEX_COLOR), 4, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(colorOffset))

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, indexVBO)
}
