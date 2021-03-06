package main

import (
	"github.com/Laremere/sdl2"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func main() {
	runtime.LockOSThread()

	res, err := http.Get("http://vps.redig.us")
	if err != nil {
		log.Fatal(err)
	}
	ipAddrBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server IP:", string(ipAddrBytes))

	screenWidth := 1280
	screenHeight := 720

	err = sdl.SdlInit()
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Line Of Sight", 30, 30, screenWidth, screenHeight, sdl.WindowShown|sdl.WindowOpengl)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Close()

	context, err := window.CreateContext()
	if err != nil {
		log.Fatal(err)
	}
	defer context.Delete()

	draw, err := SetupOpengl(screenWidth, screenHeight)
	if err != nil {
		log.Fatal(err)
	}

	mapBytes, err := ioutil.ReadFile("map.txt")
	if err != nil {
		log.Fatal(err)
	}
	mapString := string(mapBytes)

	scene := newScene(50, 50)

	k := 0
	for j := 0; j < 50; j++ {
		for i := 0; i < 50; i++ {
			mapItem := int64(-1)
			for mapItem < 0 {
				token := string(mapString[k])
				mapItem, err = strconv.ParseInt(token, 10, 32)
				if err != nil {
					mapItem = -1
				}
				k++
			}
			scene.setWall(i, 49-j, Wall(mapItem))
			// if i == 0 || i == 49 ||
			// 	j == 0 || j == 49 ||
			// 	(j%4 == 0 && i%4 == 0) {
			// 	scene.setWall(i, j, WallStone)
			// } else {
			// 	scene.setWall(i, j, WallNone)
			// }
		}
	}

	draw.generateWalls(scene)

	player := NewPlayer()
	scene.entities = append(scene.entities, player)
	scene.entities = append(scene.entities, newServerConn(string(ipAddrBytes), player))

	var inputState InputState
	inputState.keydown = make(map[string]bool)
	var outputState OutputState
	outputState.screenBounds[0] = float32(screenWidth)
	outputState.screenBounds[1] = float32(screenHeight)
	for running := true; running; {
		EndTime := time.Now().Add(time.Second / 60)

		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}

			switch event := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMoveEvent:
			case *sdl.KeyupEvent:
				inputState.keydown[event.Key] = false
			case *sdl.KeydownEvent:
				inputState.keydown[event.Key] = true
			default:
				//log.Println("Unkown event:", reflect.ValueOf(event).Type().Name(), event)
			}
		}

		inputState.direction = [2]float32{0, 0}
		if inputState.keydown["A"] {
			inputState.direction[0] -= 1
		}
		if inputState.keydown["D"] {
			inputState.direction[0] += 1
		}
		if inputState.keydown["W"] {
			inputState.direction[1] += 1
		}
		if inputState.keydown["S"] {
			inputState.direction[1] -= 1
		}
		if inputState.direction[1]*inputState.direction[0] != 0 {
			inputState.direction[0] *= 0.70710678118
			inputState.direction[1] *= 0.70710678118
		}

		for _, entity := range scene.entities {
			entity.step(scene, &inputState, &outputState)
		}

		draw.draw(scene, &outputState)
		window.GlSwap()
		//time.Sleep(time.Second / 30)
		time.Sleep(EndTime.Sub(time.Now()))
	}
}
