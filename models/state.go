package models

import "encoding/json"

type SyncDTO struct {
	Player Player
	Game   GameState
}

type GameState struct {
	Topcard Card
	Turn    string
	Reverse bool
}

func SerializeSyncDTO(syncDTO SyncDTO) []byte {
	data, err := json.Marshal(syncDTO)
	if err != nil {
		return nil
	}
	return data
}
