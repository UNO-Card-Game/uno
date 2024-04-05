package internal

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"uno/models"
	"uno/models/constants/color"
	"uno/models/constants/rank"

	"github.com/gorilla/websocket"
)

type Game struct {
	Players          []*models.Player
	GameDeck         *models.GameDeck
	DisposedGameDeck *models.GameDeck
	CurrentTurn      int
	GameDirection    bool
	ActivePlayer     *models.Player //pointer to active player
	mu               sync.Mutex
	GameTopCard      *models.Card //TopCard of Playing Game
	GameFirstMove    bool
}

func NewGame(playerNames []string) *Game {
	players := make([]*models.Player, len(playerNames)) //Clients
	for i, name := range playerNames {
		players[i] = models.NewPlayer(name)
	}

	gameDeck := models.NewGameDeck() //Initialised Game Deck
	for _, p := range players {
		p.AddCards(gameDeck.Cut(7)) //Players get the cards
	}
	disposedGameDeck := &models.GameDeck{
		Deck: &models.Deck{
			Cards: make([]models.Card, 0), // Initialize the Cards slice
		},
	}

	game := &Game{
		Players:          players,
		GameDeck:         gameDeck,
		DisposedGameDeck: disposedGameDeck,
	}
	return game
}

func (g *Game) NextTurn() {
	// g.ActivePlayer.mu.Lock()
	// defer g.ActivePlayer.mu.Unlock()

	if g.ActivePlayer != nil {

		g.ActivePlayer.Send("Your turn is over.")
	}
	//Check for UNO
	if g.ActivePlayer.Deck.NumberOfCards() == 1 {
		g.checkforUNO(g.ActivePlayer)
	}

	integerDirection := convertDirectionToInteger(g.GameDirection)

	nextTurn := (g.CurrentTurn + integerDirection) % len(g.Players)
	if nextTurn < 0 {
		nextTurn += len(g.Players)
	}

	nextPlayer := g.Players[nextTurn]
	nextPlayer.Send("About TO DRAWWWWWW CARDSSSS")
	topCard := g.GameTopCard
	playableCard := nextPlayer.HasPlayableCard(topCard)
	if !playableCard {
		// Draw a card from the deck
		if g.GameDeck.NumberOfCards() == 0 {
			// If the game deck is empty, reshuffle the disposed deck
			g.ShuffleDiscardPileToDeck()
		}
		g.PerformDrawAction(1) //Draw 1 card
		g.NextTurn()           // Call NextTurn again to move to the next player
		return
	}
	g.CurrentTurn = nextTurn
	g.ActivePlayer = g.Players[g.CurrentTurn]
	g.ActivePlayer.Send(fmt.Sprintf("It's your turn, %s! ", g.ActivePlayer.Name))
	//g.DisposedGameDeck

	//g.ActivePlayer.Send()
}
func (g *Game) Start() {
	// g.mu.Lock()
	// defer g.mu.Unlock()

	// Start the first player's turn
	g.CurrentTurn = 0
	g.GameDirection = true
	g.GameFirstMove = true
	g.ActivePlayer = g.Players[g.CurrentTurn] //g.Players is already a pointer
	for _, p := range g.Players {

		err := p.Send(fmt.Sprintf("It's %s's turn.Please play your turn.\n", g.ActivePlayer.Name))
		if err != nil {
			// Handle the error
			fmt.Errorf("Error sending message to player %s: %v\n", p.Name, err)
		}
	}

}
func (g *Game) PlayCard(player *models.Player, cardIdx int, newColorStr ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	card := player.Deck.Cards[cardIdx]

	// Handle WILD and DRAW4 cards
	if card.Type() == "action-card-no-color" {
		// Get the new color from the player
		if len(newColorStr) > 0 {
			newColor, err := color.ParseColor(newColorStr[0])
			card.Color = newColor // Set the card's color to the new color
			if err != nil {
				player.Send("Invalid color choice.Try again with correct color <blue,red,green,yellow>")
				return
			}

			// Update the game's top card
			// Add the card to the disposed deck
			disposedCard := player.Deck.RemoveCard(cardIdx)
			g.DisposedGameDeck.AddCard(disposedCard)
			g.GameTopCard = &card

			for _, p := range g.Players {
				p.Send(fmt.Sprintf("%s played %s and changed the color to %s", player.Name, card.LogCard(), newColor))
			}
			// Perform additional game logic for  DRAW4 card
			if card.Rank == rank.DRAW_4 {
				g.PerformDrawAction(4)
				g.skipNextTurn()

			}
			// Notify all players about the played card and the new color

			// Move to the next turn
			//ACTION CARD CANNOT BE LAST PLAYABLE CARD ,Hence not checking for winner
			g.NextTurn()
		} else {
			player.Send("Invalid move.Add New Color to WILD or DRAW_4 in format playcard <cardIndex> <color>. Try again !!")
			return
		}

	} else if g.IsValidMove(card, player) { //Check if valid move
		// Perform game logic for playing the card
		// Update game state, check for UNO, etc.

		// Notify all players about the played card
		if card.Type() == "action-card" {
			g.dealwithActionCards(card)
		}
		for _, p := range g.Players {
			p.Send(fmt.Sprintf("%s played %s", player.Name, card.LogCard()))
		}

		// Add the card to the disposed deck
		disposedCard := player.Deck.RemoveCard(cardIdx)
		g.DisposedGameDeck.AddCard(disposedCard)
		g.GameTopCard = &disposedCard

		// Move to the next turn
		g.NextTurn()
	} else {
		// Notify the player that the move is invalid
		player.Send("Invalid move.Wrong card or wrong player . Try again.")
	}
}

func (g *Game) IsValidMove(playedCard models.Card, player *models.Player) bool {
	// Check if it's the first move of the game and the player is active
	if player == g.ActivePlayer {
		if playedCard.Type() == "action-card-no-color" {
			return true
		}
		if g.GameFirstMove {
			g.GameFirstMove = false // First card will not check deck's top card
			return true
		}
	}

	// If it's not the active player's turn, disallow move
	if player != g.ActivePlayer {
		return false
	}

	// If the played card matches the color or rank of the top card, it's a valid move
	return playedCard.IsSameColor(*g.GameTopCard) || playedCard.IsSameRank(*g.GameTopCard)
}

func (g *Game) HandleMessage(msg string, conn *websocket.Conn, clientName string) {
	parts := strings.Split(msg, " ")
	command := parts[0]

	// Find the player index in the Players slice
	playerPtr := findPlayer(g.Players, clientName)
	if playerPtr == nil {
		// Handle the case where the player is not found
		conn.WriteMessage(websocket.TextMessage, []byte("Player not found."))
		return
	}

	switch command {
	case "playcard":

		if len(parts) < 2 {
			// Handle invalid command format

			conn.WriteMessage(websocket.TextMessage, []byte("Invalid command format.\n Usage: playcard <cardIndex> and Usage: playcard <cardIndex> <color> for DRAW 4 and WILD"))

		} else if len(parts) == 2 {
			cardidx, _ := strconv.Atoi(parts[1])
			g.PlayCard(playerPtr, cardidx)
			return
		} else if len(parts) == 3 {
			cardidx, _ := strconv.Atoi(parts[1])
			newColor := parts[2]

			g.PlayCard(playerPtr, cardidx, newColor)
			return
		}
		// Call the PlayCard function for the player

	case "showcards":
		// Call the ShowCards function for the player
		your_cards := playerPtr.CardInHand()
		conn.WriteMessage(websocket.TextMessage, []byte(your_cards))
	case "topcard":
		conn.WriteMessage(websocket.TextMessage, []byte(g.GameTopCard.LogCard()))
	default:
		conn.WriteMessage(websocket.TextMessage, []byte("Chat msg"))
	}
}
