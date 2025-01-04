package dtos

import (
	"uno/models/game"
)

type SyncDTO struct {
	Player game.Player `json:"player"`
	Game   GameState   `json:"game"`
	Room   RoomState   `json:"room"`
}

type GameState struct {
	Topcard game.Card `json:"topcard"`
	Turn    string    `json:"turn"`
	Reverse bool      `json:"reverse"`
}

type RoomState struct {
	Players    []string `json:"players"`
	RoomId     int           `json:"id"`
	MaxPlayers int           `json:"max_players"`
}

func (dto SyncDTO) Serialize() []byte {
	return Serialize(
		dto, "sync")
}
