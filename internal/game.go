package internal

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"uno/models"
	"uno/models/constants/color"
	"uno/models/constants/rank"

	"github.com/gorilla/websocket"
)

type Game struct {
	Players          []*models.Player
	GameDeck         *models.GameDeck
	DisposedGameDeck *models.GameDeck
	GameStarted      bool
	CurrentTurn      int
	GameDirection    bool
	ActivePlayer     *models.Player //pointer to active player
	mu               sync.Mutex
	GameTopCard      *models.Card // Top card of the game
	GameFirstMove    bool
	Network          Network
}

func NewGame() *Game {
	gameDeck := models.NewGameDeck() //Initialised Game Deck
	disposedGameDeck := &models.GameDeck{
		Deck: &models.Deck{
			Cards: make([]models.Card, 0), // Initialize the Cards slice
		},
	}
	topcard := &gameDeck.Cut(1)[0]
	var (
		game = &Game{
			Players:          make([]*models.Player, 0),
			GameDeck:         gameDeck,
			DisposedGameDeck: disposedGameDeck,
			GameStarted:      false,
			GameDirection:    false,
			GameTopCard:      topcard,
			Network:          *NewNetwork(),
		}
	)
	return game
}

func (g *Game) AddPlayer(player *models.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	player.AddCards(g.GameDeck.Cut(7))
	g.Players = append(g.Players, player)
}

func (g *Game) NextTurn() {
	// g.ActivePlayer.mu.Lock()
	// defer g.ActivePlayer.mu.Unlock()
	//check for Game winner
	if g.ActivePlayer.Deck.NumberOfCards() == 0 {
		g.declareWinner(g.ActivePlayer)
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
	//nextPlayer.Send("About TO DRAWWWWWW CARDSSSS")
	topCard := g.GameTopCard
	playableCard := nextPlayer.HasPlayableCard(topCard)
	if !playableCard {
		// Draw a card from the deck
		if g.GameDeck.NumberOfCards() == 0 {
			// If the game deck is empty, reshuffle the disposed deck
			g.ShuffleDiscardPileToDeck()
		}
		g.PerformDrawAction(g.getNextPlayer(), 1) //Draw 1 card
		playableCard = nextPlayer.HasPlayableCard(topCard)

		// If the player still doesn't have a playable card after drawing, skip their turn
		if !playableCard {
			g.Network.SendMessage(nextPlayer, "You don't have a playable card")
			g.skipNextTurn()
			g.NextTurn() // Call NextTurn again to move to the next player
			return
		}

	}
	g.CurrentTurn = nextTurn
	g.ActivePlayer = g.Players[g.CurrentTurn]
	g.Network.SendMessage(g.ActivePlayer, fmt.Sprintf("It's your turn, %s! ", g.ActivePlayer.Name))
	//g.DisposedGameDeck

	//g.ActivePlayer.Send()
}
func (g *Game) Start() {
	// g.mu.Lock()
	// defer g.mu.Unlock()

	// Start the first player's turn
	g.CurrentTurn = 0
	g.GameFirstMove = true
	g.ActivePlayer = g.Players[g.CurrentTurn] //g.Players is already a pointer
	g.GameStarted = true
	go g.Network.BroadcastMessages()
	for _, p := range g.Players {

		err := g.Network.SendMessage(p, fmt.Sprintf("It's %s's turn.Please play your turn.\n", g.ActivePlayer.Name))
		if err != nil {
			// Handle the error
			fmt.Errorf("Error sending message to player %s: %v\n", p.Name, err)
		}
	}

	g.SyncAllPlayers()
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
				g.Network.SendMessage(player, "Invalid color choice.Try again with correct color <blue,red,green,yellow>")
				return
			}

			// Update the game's top card
			// Add the card to the disposed deck
			disposedCard := player.Deck.RemoveCard(cardIdx)
			g.DisposedGameDeck.AddCard(disposedCard)
			g.GameTopCard = &card

			for _, p := range g.Players {
				g.Network.SendMessage(p, fmt.Sprintf("%s played %s and changed the color to %s", player.Name, card.LogCard(), newColor))
			}
			// Perform additional game logic for  DRAW4 card
			if card.Rank == rank.DRAW_4 {
				g.PerformDrawAction(g.getNextPlayer(), 4)
				g.skipNextTurn()

			}
			g.Network.SendMessage(player, "Your turn is over.")
			g.NextTurn()
		} else {
			g.Network.SendMessage(player, "Invalid move.Add New Color to WILD or DRAW_4 in format playcard <cardIndex> <color>. Try again !!")
			return
		}

	} else if g.IsValidMove(card, player) { //Check if valid move
		// Perform game logic for playing the card

		// Notify all players about the played card
		if card.Type() == "action-card" {
			g.dealwithActionCards(card)
		}
		for _, p := range g.Players {
			g.Network.SendMessage(p, fmt.Sprintf("%s played %s", player.Name, card.LogCard()))
		}

		// Add the card to the disposed deck
		disposedCard := player.Deck.RemoveCard(cardIdx)
		g.DisposedGameDeck.AddCard(disposedCard)
		g.GameTopCard = &disposedCard

		// Move to the next turn
		g.Network.SendMessage(player, "Your turn is over.")
		g.NextTurn()
	} else {
		// Notify the player that the move is invalid
		g.Network.SendMessage(player, "Invalid move.Wrong card or wrong player . Try again.")
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

func (g *Game) PerformDrawAction(player *models.Player, card_count int) {
	cardsDrawn := g.GameDeck.Cut(card_count)
	player.AddCards(cardsDrawn)
	for _, card := range cardsDrawn {
		g.Network.SendMessage(player, fmt.Sprintf("%s Drew %s", player.Name, card.LogCard()))
	}
	g.Network.SendMessage(player, fmt.Sprintf("%s Drew Drew %d cards  ", player.Name, card_count))

}
func (g *Game) ShuffleDiscardPileToDeck() {
	if len(g.DisposedGameDeck.Deck.Cards) > 0 {
		// Shuffle the discard pile
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(g.DisposedGameDeck.Deck.Cards), func(i, j int) {
			g.DisposedGameDeck.Deck.Cards[i], g.DisposedGameDeck.Deck.Cards[j] = g.DisposedGameDeck.Deck.Cards[j], g.DisposedGameDeck.Deck.Cards[i]
		})

		// Create a new deck from the shuffled discard pile
		g.GameDeck.Deck.Cards = g.DisposedGameDeck.Deck.Cards
		g.GameDeck.Deck.Counter = g.DisposedGameDeck.Deck.Counter

		// Clear the discard pile
		g.DisposedGameDeck.Deck.Cards = make([]models.Card, 0)
	}
}

// reverseGameDirection reverses the game direction
func (g *Game) reverseGameDirection() {
	g.GameDirection = !g.GameDirection
}

// skipNextTurn skips the next player's turn
func (g *Game) skipNextTurn() {
	nextPlayer := g.getNextPlayer()
	g.Network.SendMessage(nextPlayer, "Your turn is SKIPPED.......... ")

	g.switchtoNextPlayer()
}

// declareWinner declares the winner of the game
func (g *Game) declareWinner(winner *models.Player) {
	for _, p := range g.Players {
		g.Network.SendMessage(p, fmt.Sprintf("%s HAS WON THE GAME!!!!", winner.Name))
	}
	for _, p := range g.Players {
		g.Network.SendMessage(p, fmt.Sprintf("GAME OVER  %s ,CLOSING CONNECTION ", winner.Name))
		g.Network.CloseConnection(p)
	}
	// Perform any necessary  end-game animation with bubbleTea
}
func (g *Game) checkforUNO(player *models.Player) {
	for _, p := range g.Players {
		g.Network.SendMessage(p, fmt.Sprintf("UNO !!!! by %s ", player.Name))
	}
	// Perform any necessary  end-game animation with bubbleTea
}

// getNextPlayer returns the next player based on the game direction
func (g *Game) getNextPlayer() *models.Player {
	integerDirection := convertDirectionToInteger(g.GameDirection)

	nextTurn := (g.CurrentTurn + integerDirection) % len(g.Players)
	if nextTurn < 0 {
		nextTurn += len(g.Players)

	}
	return g.Players[nextTurn]
}
func (g *Game) dealwithActionCards(card models.Card) {
	cardType := card.Rank
	switch cardType {
	case "skip":
		g.skipNextTurn()
		break
	case "draw_2":
		g.PerformDrawAction(g.getNextPlayer(), 2)
		break
	case "reverse":
		g.reverseGameDirection()
		break
	default:
		// Handle unexpected card types here, e.g., log an error
		fmt.Println("Unexpected card type:", cardType)
	}
}

func (g *Game) switchtoNextPlayer() {
	integerDirection := convertDirectionToInteger(g.GameDirection)
	g.CurrentTurn = (g.CurrentTurn + integerDirection) % len(g.Players)
	if g.CurrentTurn < 0 {
		g.CurrentTurn += len(g.Players)

	}
	g.ActivePlayer = g.Players[g.CurrentTurn] // Set Active player
}

func (g *Game) HandleMessage(msg string, player *models.Player) {
	g.SyncAllPlayers()
	parts := strings.Split(msg, " ")
	command := parts[0]

	conn := g.Network.clients[*player]

	// Find the player index in the Players slice
	if command == "sync" {
		var state = models.SerializeSyncDTO(models.SyncDTO{
			Player: *player,
			Game: models.GameState{
				Topcard: *g.GameTopCard,
				Turn:    g.ActivePlayer.Name,
				Reverse: g.GameDirection,
			},
		})
		conn.WriteMessage(websocket.TextMessage, state)
	} else if command == "playcard" {
		if len(parts) < 2 {
			// Handle invalid command format
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid command format.\n Usage: playcard <cardIndex> and Usage: playcard <cardIndex> <color> for DRAW 4 and WILD"))

		} else if len(parts) == 2 {
			cardidx, _ := strconv.Atoi(parts[1])
			g.PlayCard(player, cardidx)
			return
		} else if len(parts) == 3 {
			cardidx, _ := strconv.Atoi(parts[1])
			newColor := parts[2]

			g.PlayCard(playerPtr, cardidx, newColor)
			return
	   } else if command == "draw" {
		if g.ActivePlayer == player {
			g.PerformDrawAction(player, 1)
			g.NextTurn()
		}

	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("Invalid command format"))
	}
}

func (g *Game) SyncAllPlayers() {
	for _, player := range g.Players {
		state := models.SerializeSyncDTO(models.SyncDTO{
			Player: *player,
			Game: models.GameState{
				Topcard: *g.GameTopCard,
				Turn:    g.ActivePlayer.Name,
				Reverse: g.GameDirection,
			},
		})
		conn := g.Network.clients[*player]
		conn.WriteMessage(websocket.TextMessage, state)
	}
}
