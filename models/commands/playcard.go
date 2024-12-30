package commands

import "uno/models/game"

type PlayCardComamnd struct {
	Card   game.Card   `json:"card"`
	Player game.Player `json:"player_id"`
}
