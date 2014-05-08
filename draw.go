package main

import (
	"errors"
	"github.com/go-gl/gl"
)

type Draw struct {
	screenWidth,
	screenHeight int
	simpleQuad       gl.Buffer
	walls            gl.Buffer
	wallLength       int
	simpleShader     gl.Program
	losBlockerShader gl.Program
	LOSfb            gl.Framebuffer
	LOStex           gl.Texture
	backgroundShader gl.Program
	backgroundQuad   gl.Buffer
	wallShader       gl.Program
}

func SetupOpengl(screenWidth, screenHeight int) (*Draw, error) {
	var draw Draw
	draw.screenHeight = screenHeight
	draw.screenWidth = screenWidth
	gl.Init()

	gl.Viewport(0, 0, screenWidth, screenHeight)
	draw.createLosBuffer()

	buffers := make([]gl.Buffer, 3)
	gl.GenBuffers(buffers)

	draw.simpleQuad = buffers[0]
	draw.walls = buffers[1]
	draw.backgroundQuad = buffers[2]

	draw.simpleQuad.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, 8*2*4*1, //bytes * per vertex * per quad * quads
		[]float32{
			-0.5, -0.5,
			0.5, -0.5,
			0.5, 0.5,
			-0.5, 0.5,
		}, gl.STATIC_DRAW)
	draw.simpleQuad.Unbind(gl.ARRAY_BUFFER)

	draw.backgroundQuad.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, 8*2*4*1, //bytes * per vertex * per quad * quads
		[]float32{
			-50, -50,
			100, -50,
			100, 100,
			-50, 100,
		}, gl.STATIC_DRAW)
	draw.backgroundQuad.Unbind(gl.ARRAY_BUFFER)

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
		out float outColor;

		void main()
		{
		    outColor = 1.0;
		}
		`)

	shader = gl.CreateProgram()
	shader.AttachShader(vs)
	shader.AttachShader(fs)
	shader.BindFragDataLocation(0, "outColor")
	shader.Link()
	draw.losBlockerShader = shader

	vs = compileShader(gl.VERTEX_SHADER, `
		#version 150 compatibility
		in vec2 position;
		out vec2 screenPos;
		out vec2 worldPos;
		void main()
		{
			worldPos = position;
		    screenPos = (gl_ProjectionMatrix * gl_ModelViewMatrix * vec4(position, 0.0, 1.0)).xy;
		    gl_Position = vec4(screenPos, 0.0, 1.0);
		}
		`)

	fs = compileShader(gl.FRAGMENT_SHADER, `
		#version 150

		in vec2 screenPos;
		in vec2 worldPos;
		out vec4 outColor;
		uniform sampler2D los;

		//From stack overflow
		float rand(vec2 co){
 		   return fract(sin(dot(co.xy ,vec2(12.9898,78.233))) * 43758.5453);
		}

		void main()
		{
			float shadow = texture(los,(screenPos + vec2(1,1))/ 2).r;
			if (shadow > 0.5){
				float grayScale = rand(floor(worldPos * vec2(5,10)));
				grayScale = round(grayScale) / 40 + 0.1;
				outColor = vec4(grayScale, grayScale, grayScale, 1.0);
			} else {
			    outColor = vec4(0.7,0.7,0.7,1.0);
			}
		}
		`)

	shader = gl.CreateProgram()
	shader.AttachShader(vs)
	shader.AttachShader(fs)
	shader.BindFragDataLocation(0, "outColor")
	shader.Link()
	draw.backgroundShader = shader

	vs = compileShader(gl.VERTEX_SHADER, `
		#version 150 compatibility
		in vec2 position;
		out vec2 worldPos;
		void main()
		{
			worldPos = position;
		    gl_Position = (gl_ProjectionMatrix * gl_ModelViewMatrix * vec4(position, 0.0, 1.0));
		}
		`)

	fs = compileShader(gl.FRAGMENT_SHADER, `
		#version 150

		in vec2 worldPos;
		out vec4 outColor;
		uniform int neighbors;
		// 210
		// 4 3
		// 765

		void main()
		{
			float grayscale = 0.1 + clamp(sin((worldPos.x - worldPos.y) * 12.5663706144),0,1)/5;
			if (worldPos.x < -0.3 && (neighbors & (1 << 4)) > 0){
				grayscale = 0.3;
			}
			if (worldPos.x > 0.3 && (neighbors & (1 << 3)) > 0){
				grayscale = 0.3;
			}
			if (worldPos.y < -0.3 && (neighbors & (1 << 6)) > 0){
				grayscale = 0.3;
			}
			if (worldPos.y > 0.3 && (neighbors & (1 << 1)) > 0){
				grayscale = 0.3;
			}
			if (worldPos.x > 0.3 && worldPos.y > 0.3 && (neighbors & 1) > 0){
				grayscale = 0.3;
			}
			if (worldPos.x < -0.3 && worldPos.y > 0.3 && (neighbors & 1 << 2) > 0){
				grayscale = 0.3;
			}
			if (worldPos.x > 0.3 && worldPos.y < -0.3 && (neighbors & 1 << 5) > 0){
				grayscale = 0.3;
			}
			if (worldPos.x < -0.3 && worldPos.y < -0.3 && (neighbors & 1 << 7) > 0){
				grayscale = 0.3;
			}

		    outColor = vec4(grayscale,grayscale,grayscale,1.0);
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

func (draw *Draw) createLosBuffer() {
	draw.LOSfb = gl.GenFramebuffer()
	draw.LOSfb.Bind()

	draw.LOStex = gl.GenTexture()
	draw.LOStex.Bind(gl.TEXTURE_2D)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R8,
		draw.screenWidth, draw.screenHeight,
		0, gl.RED, gl.BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0,
		gl.TEXTURE_2D, draw.LOStex, 0)

	draw.LOStex.Unbind(gl.TEXTURE_2D)
	draw.LOSfb.Unbind()
}

func (draw *Draw) generateWalls(scene *Scene) {
	vertexes := make([]float32, 0)
	for i := 0; i < scene.width; i++ {
		for j := 0; j < scene.height; j++ {
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
	draw.wallLength = len(vertexes) / 3

}

func (draw *Draw) draw(scene *Scene, ops *OutputState) {
	draw.LOSfb.Bind()
	gl.Clear(gl.COLOR_BUFFER_BIT)
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
	draw.losBlockerShader.Use()

	posAttrib := draw.losBlockerShader.GetAttribLocation("position")
	posAttrib.AttribPointer(3, gl.FLOAT, false, 0, nil)
	posAttrib.EnableArray()

	gl.DrawArrays(gl.QUADS, 0, draw.wallLength)

	draw.walls.Unbind(gl.ARRAY_BUFFER)
	draw.LOSfb.Unbind()
	/////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////
	gl.Clear(gl.COLOR_BUFFER_BIT)
	/////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////
	gl.ActiveTexture(gl.TEXTURE0)
	draw.LOStex.Bind(gl.TEXTURE_2D)
	draw.backgroundQuad.Bind(gl.ARRAY_BUFFER)
	draw.backgroundShader.Use()

	posAttrib = draw.simpleShader.GetAttribLocation("position")
	posAttrib.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posAttrib.EnableArray()

	gl.DrawArrays(gl.QUADS, 0, 4)

	draw.backgroundQuad.Unbind(gl.ARRAY_BUFFER)
	draw.LOStex.Unbind(gl.TEXTURE_2D)
	/////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////
	draw.simpleQuad.Bind(gl.ARRAY_BUFFER)
	draw.wallShader.Use()

	posAttrib = draw.wallShader.GetAttribLocation("position")
	posAttrib.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posAttrib.EnableArray()

	neighborsAttrib := draw.wallShader.GetUniformLocation("neighbors")

	for i := 0; i < scene.width; i++ {
		for j := 0; j < scene.width; j++ {
			if scene.getWall(i, j) == WallStone {
				var neighbors int = scene.isNotWall(i-1, j-1)<<7 |
					scene.isNotWall(i, j-1)<<6 |
					scene.isNotWall(i+1, j-1)<<5 |
					scene.isNotWall(i-1, j)<<4 |
					scene.isNotWall(i+1, j)<<3 |
					scene.isNotWall(i-1, j+1)<<2 |
					scene.isNotWall(i, j+1)<<1 |
					scene.isNotWall(i+1, j+1)
					// 210
					// 4 3
					// 765
				neighborsAttrib.Uniform1i(neighbors)
				_ = neighbors
				gl.PushMatrix()
				gl.Translatef(float32(i), float32(j), 0)
				gl.DrawArrays(gl.QUADS, 0, 4)
				gl.PopMatrix()
			}
		}
	}
	/////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////
	draw.simpleShader.Use()

	posAttrib = draw.simpleShader.GetAttribLocation("position")
	posAttrib.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posAttrib.EnableArray()

	uniColor := draw.simpleShader.GetUniformLocation("triangleColor")
	uniColor.Uniform3f(0.0, 1.0, 0.0)

	for _, entity := range scene.entities {
		entity.draw(draw)
	}

	draw.simpleQuad.Unbind(gl.ARRAY_BUFFER)

}
