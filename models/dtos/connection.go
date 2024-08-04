package dtos

type ConnectionDTO struct {
	PlayerName string   `json:"player_name"`
	RoomID     int      `json:"room_id"`
	MaxPlayers int      `json:"max_players"`
	Players    []string `json:"players"`
}

func (dto ConnectionDTO) Serialize() []byte {
	return Serialize(
		dto, "connection")
}
