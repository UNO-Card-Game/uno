package commands

import "uno/models/game"

type PlayCardCommand struct {
	CardIndex int         `json:"card_index"`
	Player    game.Player `json:"player_id"`
}
