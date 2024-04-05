package models

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	Name string
	*Deck
	Conn *websocket.Conn
	mu   sync.Mutex
}

//var mutexsend sync.Mutex

func NewPlayer(name string) *Player {
	player := &Player{
		Name: name,
		Deck: NewDeck(),
	}
	return player
}

func (player *Player) CardInHand() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s has:\n", player.Name))
	for i, c := range player.Deck.Cards {
		sb.WriteString(fmt.Sprintf("Card %d:\t%s\n", i, c.LogCard()))
	}
	return sb.String()
}

func (player *Player) AddCards(cards []Card) {
	for _, c := range cards {
		player.AddCard(c)
	}
}
func (p *Player) SetConn(conn *websocket.Conn) {
	p.Conn = conn
	fmt.Sprintf("Success writing conn for %s", p.Name)
}
func (p *Player) Send(message string) error {
	// Check if the connection is nil
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Conn == nil {
		return fmt.Errorf("connection is nil for player: %s", p.Name)
	}

	// Send the message through the WebSocket connection
	err := p.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	//mutexsend.Unlock()
	if err != nil {
		return fmt.Errorf("error sending message to player %s: %v", p.Name, err)
	}
	return nil

}

func (p *Player) HasPlayableCard(topCard *Card) bool {
	if topCard == nil {
		// If topCard is a null pointer i.e first card is draw or wildcard, return true
		return true
	}

	for _, card := range p.Deck.Cards {
		switch {
		case card.Type() == "action-card" && card.Color == topCard.Color: // DRAW2 ,SKIP ,REVERSE don't have rank
			return true
		case card.Type() == "action-card-no-color":
			return true
		case card.Color == topCard.Color || card.Rank == topCard.Rank:
			return true
		}
	}

	return false
}
func (p *Player) CloseConnection() error {
	if p.Conn != nil {
		return p.Conn.Close()
	}
	return nil
}
