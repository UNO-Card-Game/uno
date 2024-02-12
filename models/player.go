package models

import "fmt"

type Player struct {
	Name string
	*Deck
}

func NewPlayer(name string) *Player {
	player := &Player{
		Name: name,
		Deck: NewDeck(),
	}
	return player
}

func (player *Player) CardInHand() {
	fmt.Println(player.Name, "has")
	for i, c := range player.Deck.Cards {
		fmt.Println("Card", i, ":\t", c.LogCard())
	}
}

func (player *Player) AddCards(cards []Card) {
	for _, c := range cards {
		player.AddCard(c)
	}
}
