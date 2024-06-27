package dtos

import "uno/models/game"

type ConnectionDTO struct {
	PlayerName string
	RoomID     int
	MaxPlayers int
	Players    []*game.Player
}

func (dto ConnectionDTO) Serialize() []byte {
	return Serialize(
		dto, "connection")
}
