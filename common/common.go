package common

type Player struct {
	Position [2]float32
	Color    [3]float32
}

type ServerState struct {
	Players []Player
	Speed   float32
}

type ClientState struct {
	Position [2]float32
}
