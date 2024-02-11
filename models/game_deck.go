package models

import (
	color "github.com/AvikantSrivastava/uno/models/constants/color"
	rank "github.com/AvikantSrivastava/uno/models/constants/rank"
)

type GameDeck struct {
	*Deck
}

func NewGameDeck() *GameDeck {
	gd := &GameDeck{
		Deck: NewDeck(),
	}
	gd.initColoredCards()
	gd.initNonColoredCards()
	gd.Shuffle()
	return gd
}

func (gd *GameDeck) initColoredCards() {
	for _, c := range color.ALLColors {
		for _, r := range rank.NumberCards {
			gd.AddCard(Card{Color: c, Rank: r})
		}
		for _, r := range rank.ActionCards {
			gd.AddCard(Card{Color: c, Rank: r})
		}
	}
}

func (gd *GameDeck) initNonColoredCards() {
	for _, r := range rank.ActionCardsNoColor {
		gd.AddCard(Card{Color: "", Rank: r})
	}
}
