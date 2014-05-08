package main

import (
	"github.com/go-gl/gl"
)

type Scene struct {
	width, height int
	walls         []Wall
	entities      []Entity
}

func newScene(width, height int) *Scene {
	return &Scene{
		width, height,
		make([]Wall, width*height),
		make([]Entity, 0),
	}
}

func (scene *Scene) getWall(x, y int) Wall {
	return scene.walls[x+y*scene.width]
}

func (scene *Scene) setWall(x, y int, wall Wall) {
	scene.walls[x+y*scene.width] = wall
}

type Wall int

const (
	WallNone Wall = iota
	WallStone
)

type Entity interface {
	step(*Scene, *InputState, *OutputState)
	draw(*Draw)
}

type InputState struct {
	direction [2]float32
	keydown   map[string]bool
}

type OutputState struct {
	screenCenter [2]float32
	screenBounds [2]float32
}

type Player struct {
	position [2]float32
}

func NewPlayer() *Player {
	return &Player{[2]float32{5, 5}}
}

func (p *Player) step(scene *Scene, ips *InputState, ops *OutputState) {
	var speed float32 = 0.2
	p.position[0] += ips.direction[0] * speed
	p.position[1] += ips.direction[1] * speed

	ops.screenCenter = p.position
}

func (p *Player) draw(draw *Draw) {
	gl.PushMatrix()
	gl.Translatef(p.position[0], p.position[1], 0)
	gl.DrawArrays(gl.QUADS, 0, 6)
	gl.PopMatrix()
}
