package internal

import (
	"uno/models"

	"golang.org/x/exp/slices"
)

func ShouldBroadcast(msg string) bool {
	privateCommands := []string{"showcards,playcard"}

	for _, command := range privateCommands {
		if command == msg {
			return false
		}
	}

	return true
}
func findPlayer(players []*models.Player, name string) *models.Player {
	idx := slices.IndexFunc(players, func(p *models.Player) bool {
		return p.Name == name
	})
	if idx == -1 {
		return nil
	}
	return players[idx]
}
func convertDirectionToInteger(direction bool) int {
	if direction {
		return 1
	}
	return -1
}
func removeCardFromHand(hand []models.Card, card models.Card) []models.Card {
	for i, c := range hand {
		if c == card {
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand
}
