package main

import (
	"github.com/go-gl/gl"
	"math"
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
	if x < 0 || y < 0 || x >= scene.width || y >= scene.height {
		return WallStone
	}
	return scene.walls[x+y*scene.width]
}

func (scene *Scene) isNotWall(x, y int) int {
	if x < 0 || y < 0 || x >= scene.width || y >= scene.height {
		return 0
	}
	if scene.walls[x+y*scene.width] == WallNone {
		return 1
	}
	return 0
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

	tileX := float32(math.Floor(float64(p.position[0] + 0.5)))
	tileY := float32(math.Floor(float64(p.position[1] + 0.5)))

	right := p.position[0] > tileX
	left := p.position[0] < tileX
	top := p.position[1] > tileY
	bottom := p.position[1] < tileY

	if right && scene.getWall(int(tileX+1), int(tileY)) != WallNone {
		p.position[0] = tileX
	}
	if left && scene.getWall(int(tileX-1), int(tileY)) != WallNone {
		p.position[0] = tileX
	}
	if top && scene.getWall(int(tileX), int(tileY+1)) != WallNone {
		p.position[1] = tileY
	}
	if bottom && scene.getWall(int(tileX), int(tileY-1)) != WallNone {
		p.position[1] = tileY
	}

	if top && right && scene.getWall(int(tileX+1), int(tileY+1)) != WallNone {
		dx := p.position[0] - tileX
		dy := p.position[1] - tileY
		if dx > dy {
			p.position[1] = tileY
		} else {
			p.position[0] = tileX
		}
	}
	if top && left && scene.getWall(int(tileX-1), int(tileY+1)) != WallNone {
		dx := tileX - p.position[0]
		dy := p.position[1] - tileY
		if dx > dy {
			p.position[1] = tileY
		} else {
			p.position[0] = tileX
		}
	}

	if bottom && right && scene.getWall(int(tileX+1), int(tileY-1)) != WallNone {
		dx := p.position[0] - tileX
		dy := tileY - p.position[1]
		if dx > dy {
			p.position[1] = tileY
		} else {
			p.position[0] = tileX
		}
	}
	if bottom && left && scene.getWall(int(tileX-1), int(tileY-1)) != WallNone {
		dx := tileX - p.position[0]
		dy := tileY - p.position[1]
		if dx > dy {
			p.position[1] = tileY
		} else {
			p.position[0] = tileX
		}
	}

	_, _, _ = left, top, bottom
	ops.screenCenter = p.position
}

func (p *Player) draw(draw *Draw) {
	gl.PushMatrix()
	gl.Translatef(p.position[0], p.position[1], 0)
	gl.DrawArrays(gl.QUADS, 0, 6)
	gl.PopMatrix()
}
