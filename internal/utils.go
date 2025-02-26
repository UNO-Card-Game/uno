package internal

import (
	"uno/models/game"
)

func convertDirectionToInteger(direction bool) int {
	if direction {
		return 1
	}
	return -1
}
func removeCardFromHand(hand []game.Card, card game.Card) []game.Card {
	for i, c := range hand {
		if c == card {
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand
}
