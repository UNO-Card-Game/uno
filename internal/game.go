package internal

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
	"uno/models/commands"
	"uno/models/constants/color"
	"uno/models/constants/rank"
	"uno/models/dtos"
	"uno/models/game"

	"github.com/gorilla/websocket"
)

type Game struct {
	Players          []*game.Player
	GameDeck         *game.GameDeck
	DisposedGameDeck *game.GameDeck
	Room             *Room
	GameStarted      bool
	CurrentTurn      int
	GameDirection    bool
	ActivePlayer     *game.Player //pointer to active player
	mu               sync.Mutex
	TopCard          game.Card // Top card of the game
	GameFirstMove    bool
	Network          Network
}

func NewGame() *Game {
	gameDeck := game.NewGameDeck() //Initialised Game Deck
	disposedGameDeck := &game.GameDeck{
		Deck: &game.Deck{
			Cards: make([]game.Card, 0), // Initialize the Cards slice
		},
	}
	topcard := gameDeck.Cut(1)[0]
	var (
		game = &Game{
			Players:          make([]*game.Player, 0),
			GameDeck:         gameDeck,
			DisposedGameDeck: disposedGameDeck,
			GameStarted:      false,
			GameDirection:    false,
			TopCard:          topcard,
			Network:          *NewNetwork(),
		}
	)
	return game
}

func (g *Game) AddPlayer(player *game.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	player.AddCards(g.GameDeck.Cut(7))
	g.Players = append(g.Players, player)
}

func (g *Game) NextTurn() {
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
	topCard := g.TopCard
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
			g.Network.SendInfoMessage(nextPlayer, "You don't have a playable card")
			g.skipNextTurn()
			g.NextTurn() // Call NextTurn again to move to the next player
			return
		}

	}
	g.CurrentTurn = nextTurn
	g.ActivePlayer = g.Players[g.CurrentTurn]
	g.Network.SendInfoMessage(g.ActivePlayer, "It is your turn.")

}
func (g *Game) Start() {
	// Start the first player's turn
	g.CurrentTurn = 0
	g.GameFirstMove = true
	g.ActivePlayer = g.Players[g.CurrentTurn]
	g.GameStarted = true
}

func (g *Game) PlayCard(player *game.Player, cardIdx int, newColorStr ...string) {
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
				g.Network.SendInfoMessage(player, "Invalid color choice.Try again with correct color <blue,red,green,yellow>")
				return
			}

			// Update the game's top card
			// Add the card to the disposed deck
			disposedCard := player.Deck.RemoveCard(cardIdx)
			g.DisposedGameDeck.AddCard(disposedCard)
			g.TopCard = card

			g.Network.BroadcastInfoMessage(fmt.Sprintf("%s played %s and changed the color to %s", player.Name, card.LogCard(), newColor))

			// Perform additional game logic for  DRAW4 card
			if card.Rank == rank.DRAW_4 {
				g.PerformDrawAction(g.getNextPlayer(), 4)
				g.skipNextTurn()

			}
			g.Network.SendInfoMessage(player, "Your turn is over.")
			g.NextTurn()
		} else {
			g.Network.SendInfoMessage(player, "Invalid move.Add New Color to WILD or DRAW_4 in format playcard <cardIndex> <color>. Try again !!")
			return
		}

	} else if g.IsValidMove(card, player) { //Check if valid move
		// Perform game logic for playing the card

		// Notify all players about the played card
		if card.Type() == "action-card" {
			g.dealwithActionCards(card)
		}
		g.Network.BroadcastInfoMessage(fmt.Sprintf("%s played %s", player.Name, card.LogCard()))

		// Add the card to the disposed deck
		disposedCard := player.Deck.RemoveCard(cardIdx)
		g.DisposedGameDeck.AddCard(disposedCard)
		g.TopCard = disposedCard

		// Move to the next turn
		g.Network.SendInfoMessage(player, "Your turn is over.")
	} else {
		// Notify the player that the move is invalid
		g.Network.SendInfoMessage(player, "Invalid move.Wrong card or wrong player . Try again.")
	}
}

func (g *Game) IsValidMove(playedCard game.Card, player *game.Player) bool {
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
	return playedCard.IsSameColor(g.TopCard) || playedCard.IsSameRank(g.TopCard)
}

func (g *Game) PerformDrawAction(player *game.Player, card_count int) {
	cardsDrawn := g.GameDeck.Cut(card_count)
	player.AddCards(cardsDrawn)
	for _, card := range cardsDrawn {
		g.Network.SendInfoMessage(player, fmt.Sprintf("%s Drew %s", player.Name, card.LogCard()))
	}
	g.Network.SendInfoMessage(player, fmt.Sprintf("%s Drew Drew %d cards  ", player.Name, card_count))

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
		g.DisposedGameDeck.Deck.Cards = make([]game.Card, 0)
	}
}

// reverseGameDirection reverses the game direction
func (g *Game) reverseGameDirection() {
	g.GameDirection = !g.GameDirection
}

// skipNextTurn skips the next player's turn
func (g *Game) skipNextTurn() {
	nextPlayer := g.getNextPlayer()
	g.Network.SendInfoMessage(nextPlayer, "Your turn is SKIPPED")

	g.switchtoNextPlayer()
}

// declareWinner declares the winner of the game
func (g *Game) declareWinner(winner *game.Player) {
	for _, p := range g.Players {
		g.Network.SendInfoMessage(p, fmt.Sprintf("%s HAS WON THE GAME!!!!", winner.Name))
	}
	for _, p := range g.Players {
		g.Network.SendInfoMessage(p, fmt.Sprintf("GAME OVER  %s ,CLOSING CONNECTION ", winner.Name))
		g.Network.CloseConnection(p)
	}
	// Perform any necessary  end-game animation with bubbleTea
}
func (g *Game) checkforUNO(player *game.Player) {
	for _, p := range g.Players {
		g.Network.SendInfoMessage(p, fmt.Sprintf("UNO !!!! by %s ", player.Name))
	}
}

// getNextPlayer returns the next player based on the game direction
func (g *Game) getNextPlayer() *game.Player {
	integerDirection := convertDirectionToInteger(g.GameDirection)

	nextTurn := (g.CurrentTurn + integerDirection) % len(g.Players)
	if nextTurn < 0 {
		nextTurn += len(g.Players)

	}
	return g.Players[nextTurn]
}

func (g *Game) getAllPlayers() []string {
	var playerNames []string
	for _, player := range g.Players {
		playerNames = append(playerNames, player.Name)
	}
	return playerNames
}

func (g *Game) dealwithActionCards(card game.Card) {
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
}

func (g *Game) HandleCommand(data []byte, player *game.Player) {
	cmd, err := commands.DeserializeCommand(data)
	if err != nil {
		log.Fatalf("Failed to deserialize command: %v", err)
	}

	switch c := cmd.(type) {
	case *commands.SyncCommand:
		g.SyncPlayer(player)
	case *commands.PlayCardCommand:
		if g.ActivePlayer == player {
			g.PlayCard(player, c.CardIndex)
			g.NextTurn()
		}
		g.SyncAllPlayers()
	case *commands.DrawCardComamnd:
		if g.ActivePlayer == player {
			g.PerformDrawAction(player, 1)
			g.NextTurn()
		}
		g.SyncAllPlayers()
	default:
		log.Printf("Unknown command type: %T", c)

	}

}

func (g *Game) SyncPlayer(p *game.Player) {
	if p == nil {
		log.Printf("Cannot sync a nil player")
		return
	}

	if g.TopCard == nil {
		log.Printf("GameTopCard is nil; cannot sync player %s", p.Name)
		return
	}

	activePlayer := g.ActivePlayer
	if activePlayer == nil {
		log.Printf("ActivePlayer is nil; cannot sync")
		return
	}

	if activePlayer.Name == "" {
		log.Printf("ActivePlayer's Name is empty; cannot sync")
		return
	}

	dto := dtos.SyncDTO{
		Player: *p,
		Game: dtos.GameState{
			Topcard: g.TopCard,
			Turn:    activePlayer.Name,
			Reverse: g.GameDirection,
		},
		Room: dtos.RoomState{
			Players:    g.getAllPlayers(),
			RoomId:     g.Room.id,
			MaxPlayers: g.Room.maxPlayers,
		},
	}

	conn, ok := g.Network.clients[*p]
	if !ok {
		log.Printf("Player %s not found in network clients", p.Name)
		return
	}

	lock, ok := g.Network.locks[*p]
	if !ok {
		log.Printf("No mutex found for player %s", p.Name)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	err := conn.WriteMessage(websocket.TextMessage, dto.Serialize())
	if err != nil {
		log.Printf("Failed to sync player %s: %v", p.Name, err)
	}
}

func (g *Game) SyncAllPlayers() {
	var wg sync.WaitGroup

	for _, playerPtr := range g.Players {
		wg.Add(1)

		go func(player *game.Player) {
			defer wg.Done()
			g.SyncPlayer(player)
		}(playerPtr)
	}

	wg.Wait()
}
