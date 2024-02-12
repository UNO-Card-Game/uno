package main

import (
	"fmt"
	"github.com/AvikantSrivastava/uno/models"
)

func main() {
	var gameDeck = models.NewGameDeck()
	var player1 = models.NewPlayer("Player 1")
	var player2 = models.NewPlayer("Player 2")

	fmt.Println("Total cards in the deck before:", gameDeck.NumberOfCards())

	var set1 = gameDeck.Cut(5)
	var set2 = gameDeck.Cut(9)

	player1.AddCards(set1)
	player2.AddCards(set2)

	player1.CardInHand()
	player2.CardInHand()

}
