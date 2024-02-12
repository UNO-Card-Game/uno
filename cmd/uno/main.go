package main

import (
	"fmt"
	"github.com/AvikantSrivastava/uno/internal"
)

func main() {
	players := []string{"avikant", "player2", "player3"}
	game := internal.NewGame(players)
	topCard := game.GameDeck.TopCard()
	fmt.Println(topCard.LogCard())
}
