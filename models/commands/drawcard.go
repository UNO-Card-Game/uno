package commands

import "uno/models/game"

type DrawCardComamnd struct {
	NumberofCards int         `json:"number_of_cards"`
	Player        game.Player `json:"player_id"`
}
