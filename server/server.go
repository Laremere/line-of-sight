package main

import (
	"encoding/gob"
	"github.com/Laremere/line-of-sight/common"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	go masterLoop()

	go func() {
		playerId := 0
		for {
			playerIds <- playerId
			playerId++
		}
	}()

	name, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range addrs {
		if strings.Contains(addr, ".") {
			log.Println(addr)
			_, err = http.PostForm("http://vps.redig.us", url.Values{"ipAddr": {addr}})
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	tcp, err := net.Listen("tcp", ":2667")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := tcp.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Println("New connection: ", conn.LocalAddr())
	player := Player{id: <-playerIds, toSend: make(chan *common.ServerState)}
	playerNew <- &player
	player.gobIn = gob.NewDecoder(conn)
	player.gobout = gob.NewEncoder(conn)
	go func() {
		var state common.ClientState
		for {

			err := player.gobIn.Decode(&state)
			if err != nil {
				conn.Close()
				log.Println(err)
				playerDelete <- player.id
				return
			}
			playerUpdates <- playerUpdate{
				player.id, state.Position,
			}
		}
	}()

	go func() {
		for {
			update := <-player.toSend
			err := player.gobout.Encode(update)
			if err != nil {
				log.Println(err)
				conn.Close()
				playerDelete <- player.id
				break
			}
		}
	}()
}

func masterLoop() {
	ticker := time.NewTicker(time.Second / 60)
	players := make(map[int]*Player)
	for {
		select {
		case player := <-playerNew:
			players[player.id] = player
		case id := <-playerDelete:
			delete(players, id)
			log.Println("Client closed", id)
		case update := <-playerUpdates:
			players[update.id].position = update.position
		case <-ticker.C:
			serverState := common.ServerState{
				make([]common.Player, 0, len(players)),
				0.2,
			}

			for _, player := range players {
				serverState.Players = append(serverState.Players, common.Player{
					player.position, [3]float32{1.0, 0.0, 0.0},
				})
			}

			for _, player := range players {
				player.toSend <- &serverState
			}
		}
	}
}

var playerIds = make(chan int)

var playerUpdates = make(chan playerUpdate)
var playerNew = make(chan *Player)
var playerDelete = make(chan int)

type playerUpdate struct {
	id       int
	position [2]float32
}

type Player struct {
	id       int
	toSend   chan *common.ServerState
	position [2]float32
	state    PlayerState
	gobIn    *gob.Decoder
	gobout   *gob.Encoder
}

type PlayerState int

const (
	playerRun PlayerState = iota
	playerIt
	playerInvincible
)
