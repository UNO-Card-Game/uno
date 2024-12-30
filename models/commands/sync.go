package commands

import "uno/models/game"

type SyncCommand struct {
	Player game.Player `json:"player_id"`
}
