package dtos

import (
	"uno/models/game"
)

type SyncDTO struct {
	Player game.Player `json:"player_name"`
	Game   GameState   `json:"game"`
}

type GameState struct {
	Topcard game.Card `json:"topcard"`
	Turn    string    `json:"turn"`
	Reverse bool      `json:"reverse"`
}

func (dto SyncDTO) Serialize() []byte {
	return Serialize(
		dto, "sync")
}
