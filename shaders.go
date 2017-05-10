// Copyright 2016, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package fizzgui

import (
	"fmt"

	mgl "github.com/go-gl/mathgl/mgl32"
	graphics "github.com/tbogdala/fizzle/graphicsprovider"
)

var (
	MainVertShader = `#version 330
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

	MainFragShader = `#version 330
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

	ImageVerShader = `#version 330
  uniform mat4 VIEW;
  
  in vec2 VERTEX_POSITION;
  in vec2 VERTEX_UV;
  in vec4 VERTEX_COLOR;

  out vec2 vs_uv;
  out vec4 vs_color;
  
  void main()
  {
    vs_color = VERTEX_COLOR;
    gl_Position = VIEW * vec4(VERTEX_POSITION, 0.0, 1.0);
    vs_uv = VERTEX_UV;
  }`

	ImageFragShader = `#version 330
  uniform sampler2D IMAGE;
  
  in vec2 vs_uv;
  in vec4 vs_color;
  
  out vec4 frag_color;
  void main()
  {
    frag_color = vs_color * texture(IMAGE, vs_uv).rgba;
  }`
)

// bindMainShader sets the program, VAO, uniforms and attributes required for the
// controls to be drawn from the command buffers
func bindMainShader(view mgl.Mat4) {
	const floatSize = 4
	const uintSize = 4
	const posOffset = 0
	const uvOffset = floatSize * 2
	const texIdxOffset = floatSize * 4
	const colorOffset = floatSize * 5
	const VBOStride = floatSize * (2 + 2 + 1 + 4) // vert / uv / texIndex / color

	gfx.UseProgram(mainShader)
	gfx.BindVertexArray(vao)

	// bind the uniforms and attributes
	shaderViewMatrix := gfx.GetUniformLocation(mainShader, "VIEW")
	gfx.UniformMatrix4fv(shaderViewMatrix, 1, false, view)

	for _, font := range fonts {
		shaderTex0 := gfx.GetUniformLocation(mainShader, "TEX[0]")
		if shaderTex0 >= 0 {
			if font != nil {
				gfx.ActiveTexture(graphics.TEXTURE0)
				gfx.BindTexture(graphics.TEXTURE_2D, font.Texture)
				gfx.Uniform1i(shaderTex0, 0)
			}
		}
		break
	}

	var texUniLoc int32
	for _, tex := range textures {
		uniStr := fmt.Sprintf("TEX[%d]", tex.ID)
		texUniLoc = gfx.GetUniformLocation(mainShader, uniStr)
		if texUniLoc >= 0 {
			gfx.ActiveTexture(graphics.TEXTURE0 + tex.Texture)
			gfx.BindTexture(graphics.TEXTURE_2D, tex.Texture)
			gfx.Uniform1i(texUniLoc, int32(tex.ID))
		}
	}
	if len(textures) > 0 {
		// stupid magic
		gfx.Uniform1i(texUniLoc+1, int32(len(textures)+1))
	}

	shaderPosition := gfx.GetAttribLocation(mainShader, "VERTEX_POSITION")
	gfx.BindBuffer(graphics.ARRAY_BUFFER, comboVBO)
	gfx.EnableVertexAttribArray(uint32(shaderPosition))
	gfx.VertexAttribPointer(uint32(shaderPosition), 2, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(posOffset))

	uvPosition := gfx.GetAttribLocation(mainShader, "VERTEX_UV")
	gfx.EnableVertexAttribArray(uint32(uvPosition))
	gfx.VertexAttribPointer(uint32(uvPosition), 2, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(uvOffset))

	colorPosition := gfx.GetAttribLocation(mainShader, "VERTEX_COLOR")
	gfx.EnableVertexAttribArray(uint32(colorPosition))
	gfx.VertexAttribPointer(uint32(colorPosition), 4, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(colorOffset))

	texIdxPosition := gfx.GetAttribLocation(mainShader, "VERTEX_TEXTURE_INDEX")
	gfx.EnableVertexAttribArray(uint32(texIdxPosition))
	gfx.VertexAttribPointer(uint32(texIdxPosition), 1, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(texIdxOffset))

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, indexVBO)
}

func bindImageShader(view mgl.Mat4, image graphics.Texture) {
	const floatSize = 4
	const uintSize = 4
	const posOffset = 0
	const uvOffset = floatSize * 2
	const colorOffset = floatSize * 5
	const VBOStride = floatSize * (2 + 2 + 1 + 4) // vert / uv / texIndex / color

	gfx.UseProgram(imageShader)
	gfx.BindVertexArray(vao)

	IMAGE := gfx.GetUniformLocation(imageShader, "IMAGE")
	gfx.ActiveTexture(image)
	gfx.BindTexture(graphics.TEXTURE_2D, image)
	gfx.Uniform1i(IMAGE, 2) // i`m don`t now how it`s work

	// bind the uniforms and attributes
	shaderViewMatrix := gfx.GetUniformLocation(imageShader, "VIEW")
	gfx.UniformMatrix4fv(shaderViewMatrix, 1, false, view)

	shaderPosition := gfx.GetAttribLocation(imageShader, "VERTEX_POSITION")
	gfx.BindBuffer(graphics.ARRAY_BUFFER, comboVBO)
	gfx.EnableVertexAttribArray(uint32(shaderPosition))
	gfx.VertexAttribPointer(uint32(shaderPosition), 2, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(posOffset))

	uvPosition := gfx.GetAttribLocation(imageShader, "VERTEX_UV")
	gfx.EnableVertexAttribArray(uint32(uvPosition))
	gfx.VertexAttribPointer(uint32(uvPosition), 2, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(uvOffset))

	colorPosition := gfx.GetAttribLocation(imageShader, "VERTEX_COLOR")
	gfx.EnableVertexAttribArray(uint32(colorPosition))
	gfx.VertexAttribPointer(uint32(colorPosition), 4, graphics.FLOAT, false, VBOStride, gfx.PtrOffset(colorOffset))

	gfx.BindBuffer(graphics.ELEMENT_ARRAY_BUFFER, indexVBO)
}
