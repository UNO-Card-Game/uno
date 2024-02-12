package internal

import "github.com/AvikantSrivastava/uno/models"

type Game struct {
	Players  []*models.Player
	GameDeck *models.GameDeck
	Reverse  bool
}

func NewGame(playerNames []string) *Game {
	players := make([]*models.Player, len(playerNames))
	for i, name := range playerNames {
		players[i] = models.NewPlayer(name)
	}
	gameDeck := models.NewGameDeck()
	for _, p := range players {
		p.AddCards(gameDeck.Cut(7))
	}

	game := &Game{
		Players:  players,
		GameDeck: gameDeck,
	}
	return game
}
