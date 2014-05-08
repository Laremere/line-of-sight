package main

import (
	"errors"
	"github.com/go-gl/gl"
)

type Draw struct {
	simpleQuad   gl.Buffer
	walls        gl.Buffer
	wallLength   int
	simpleShader gl.Program
	wallShader   gl.Program
}

func SetupOpengl(screenWidth, screenHeight int) (*Draw, error) {
	var draw Draw
	gl.Init()

	gl.Viewport(0, 0, screenWidth, screenHeight)
	gl.Enable(gl.STENCIL_TEST)

	buffers := make([]gl.Buffer, 2)
	gl.GenBuffers(buffers)

	draw.simpleQuad = buffers[0]
	draw.walls = buffers[1]

	draw.simpleQuad.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, 8*2*4*1, //bytes * per vertex * per quad * quads
		[]float32{
			-0.5, -0.5,
			0.5, -0.5,
			0.5, 0.5,
			-0.5, 0.5,
		}, gl.STATIC_DRAW)
	draw.simpleQuad.Unbind(gl.ARRAY_BUFFER)

	vs := compileShader(gl.VERTEX_SHADER, `
		#version 150 compatibility

		in vec2 position;

		out vec3 Color;

		void main()
		{
			Color = vec3(abs(position.x - position.y),0,0);
		    gl_Position = gl_ProjectionMatrix * gl_ModelViewMatrix * vec4(position, 0.0, 1.0);
		}
		`)

	fs := compileShader(gl.FRAGMENT_SHADER, `
		#version 150

		uniform vec3 triangleColor;
		in vec3 Color;

		out vec4 outColor;

		void main()
		{
		    outColor = vec4(triangleColor, 1.0);
		}
		`)

	shader := gl.CreateProgram()
	shader.AttachShader(vs)
	shader.AttachShader(fs)
	shader.BindFragDataLocation(0, "outColor")
	shader.Link()
	draw.simpleShader = shader

	vs = compileShader(gl.VERTEX_SHADER, `
		#version 150 compatibility
		in vec3 position;

		void main()
		{
			vec4 pos = gl_ProjectionMatrix * gl_ModelViewMatrix * vec4(position, 1.0);
			if (pos.z < 0){
				pos.xy = normalize(pos.xy) * 10;
			}
		    gl_Position = pos;
		}
		`)

	fs = compileShader(gl.FRAGMENT_SHADER, `
		#version 150
		out vec4 outColor;

		void main()
		{
		    outColor = vec4(1.0, 1.0, 1.0, 1.0);
		}
		`)

	shader = gl.CreateProgram()
	shader.AttachShader(vs)
	shader.AttachShader(fs)
	shader.BindFragDataLocation(0, "outColor")
	shader.Link()
	draw.wallShader = shader

	return &draw, nil
}

func compileShader(shaderType gl.GLenum, source string) gl.Shader {
	shader := gl.CreateShader(shaderType)
	shader.Source(source)
	shader.Compile()
	if shader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		panic(errors.New("Error compiling shader: " + shader.GetInfoLog()))
	}
	return shader
}

func (draw *Draw) generateWalls(scene *Scene) {
	vertexes := make([]float32, 0)
	for i := 0; i < scene.width-1; i++ {
		for j := 0; j < scene.height-1; j++ {
			if (scene.getWall(i, j) == WallNone) != (scene.getWall(i+1, j) == WallNone) {
				vertexes = append(vertexes, float32(i)+0.5)
				vertexes = append(vertexes, float32(j)-0.5)
				vertexes = append(vertexes, 0)
				vertexes = append(vertexes, float32(i)+0.5)
				vertexes = append(vertexes, float32(j)+0.5)
				vertexes = append(vertexes, 0)
				vertexes = append(vertexes, float32(i)+0.5)
				vertexes = append(vertexes, float32(j)+0.5)
				vertexes = append(vertexes, 1)
				vertexes = append(vertexes, float32(i)+0.5)
				vertexes = append(vertexes, float32(j)-0.5)
				vertexes = append(vertexes, 1)
			}
			if (scene.getWall(i, j) == WallNone) != (scene.getWall(i, j+1) == WallNone) {
				vertexes = append(vertexes, float32(i)-0.5)
				vertexes = append(vertexes, float32(j)+0.5)
				vertexes = append(vertexes, 0)
				vertexes = append(vertexes, float32(i)+0.5)
				vertexes = append(vertexes, float32(j)+0.5)
				vertexes = append(vertexes, 0)
				vertexes = append(vertexes, float32(i)+0.5)
				vertexes = append(vertexes, float32(j)+0.5)
				vertexes = append(vertexes, 1)
				vertexes = append(vertexes, float32(i)-0.5)
				vertexes = append(vertexes, float32(j)+0.5)
				vertexes = append(vertexes, 1)
			}
		}
	}

	draw.walls.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexes)*8, vertexes, gl.DYNAMIC_DRAW)
	draw.walls.Unbind(gl.ARRAY_BUFFER)
	draw.wallLength = len(vertexes)

}

func (draw *Draw) draw(scene *Scene, ops *OutputState) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(
		float64(-1*ops.screenBounds[0]/2/32),
		float64(ops.screenBounds[0]/2/32),
		float64(-1*ops.screenBounds[1]/2/32),
		float64(ops.screenBounds[1]/2/32),
		-10, 10)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Translatef(ops.screenCenter[0]*-1, ops.screenCenter[1]*-1, 0)

	draw.walls.Bind(gl.ARRAY_BUFFER)
	draw.wallShader.Use()

	posAttrib := draw.wallShader.GetAttribLocation("position")
	posAttrib.AttribPointer(3, gl.FLOAT, false, 0, nil)
	posAttrib.EnableArray()

	gl.DrawArrays(gl.QUADS, 0, draw.wallLength)

	draw.walls.Bind(gl.ARRAY_BUFFER)

	draw.simpleQuad.Bind(gl.ARRAY_BUFFER)
	draw.simpleShader.Use()

	posAttrib = draw.simpleShader.GetAttribLocation("position")
	posAttrib.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posAttrib.EnableArray()

	uniColor := draw.simpleShader.GetUniformLocation("triangleColor")
	uniColor.Uniform3f(0.0, 1.0, 0.0)

	for i := 0; i < scene.width; i++ {
		for j := 0; j < scene.width; j++ {
			if scene.getWall(i, j) == WallStone {

				gl.PushMatrix()
				gl.Translatef(float32(i), float32(j), 0)
				gl.DrawArrays(gl.QUADS, 0, 6)
				gl.PopMatrix()
			}
		}
	}
	uniColor = draw.simpleShader.GetUniformLocation("triangleColor")
	uniColor.Uniform3f(1.0, 0.0, 0.0)

	for _, entity := range scene.entities {
		entity.draw(draw)
	}

	draw.simpleQuad.Unbind(gl.ARRAY_BUFFER)

}
