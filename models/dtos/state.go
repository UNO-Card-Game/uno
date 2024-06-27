package dtos

import (
	"uno/models/game"
)

type SyncDTO struct {
	Player game.Player
	Game   GameState
}

type GameState struct {
	Topcard game.Card
	Turn    string
	Reverse bool
}

func (dto SyncDTO) Serialize() []byte {
	return Serialize(
		dto, "sync")
}
