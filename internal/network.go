package internal

import (
	"fmt"
	"net/http"
	"sync"
	"uno/models/dtos"
	"uno/models/game"

	"github.com/gorilla/websocket"
)

type Network struct {
	//clients map[*websocket.Conn]*models.Player
	clients     map[game.Player]*websocket.Conn
	upgrader    websocket.Upgrader
	broadcast   chan string
	syncChannel chan string
	gameStarted bool
	locks       map[game.Player]*sync.Mutex
}

func NewNetwork() *Network {
	return &Network{
		clients: make(map[game.Player]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Accepts requests from every source
			},
		},
		broadcast:   make(chan string),
		gameStarted: false,
		locks:       make(map[game.Player]*sync.Mutex),
	}
}

func (n *Network) BroadcastMessages() {
	for {
		select {
		case message := <-n.broadcast:
			for player, conn := range n.clients {
				go func(player game.Player, conn *websocket.Conn) {
					// Use the per-player mutex to ensure thread-safe writes
					lock := n.locks[player]
					lock.Lock()
					defer lock.Unlock()

					err := conn.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
							fmt.Printf("Unexpected close error for player %s: %v\n", player.Name, err)
						} else {
							fmt.Printf("Error writing message to player %s: %v\n", player.Name, err)
						}

						// Close the connection and remove the player from the map
						conn.Close()
						delete(n.clients, player)
						delete(n.locks, player)
					}
				}(player, conn)
			}
		}
	}
}

func (n Network) ListenToClient(player *game.Player, r *Room) {
	game := r.game
	if len(game.Network.clients) == r.maxPlayers && game.GameStarted == false {
		game.Start()
		conn_info_dto := dtos.ConnectionDTO{
			player.Name,
			r.id,
			r.maxPlayers,
			r.game.getAllPlayers(),
		}
		game.Network.SendMessage(player, conn_info_dto.Serialize())
		game.Network.BroadcastInfoMessage("All players have joined. Game has started.")
	} else {
		dto := dtos.InfoDTO{Message: "Waiting for players to join the game."}
		game.Network.BroadcastMessage(dto.Serialize())
	}
	conn := n.clients[*player]
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// game.HandleMessage(string(msg), player)
		game.HandleCommand(msg, player)

	}
}

func (n Network) BroadcastMessage(message []byte) {
	n.broadcast <- string(message)
}

// TODO: Decomission this function
func (n *Network) SendMessageOld(p *game.Player, message string) error {
	conn, exists := n.clients[*p]
	if !exists {
		return fmt.Errorf("player %s not found in network clients", p.Name)
	}

	lock, exists := n.locks[*p]
	if !exists {
		return fmt.Errorf("no mutex found for player %s", p.Name)
	}

	// Lock the mutex for the player's connection
	lock.Lock()
	defer lock.Unlock()

	// Perform the write operation
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return fmt.Errorf("error sending message to player %s: %v", p.Name, err)
	}
	return nil
}

func (n *Network) SendMessage(p *game.Player, message []byte) error {
	conn, exists := n.clients[*p]
	if !exists {
		return fmt.Errorf("player %s not found in network clients", p.Name)
	}

	lock, exists := n.locks[*p]
	if !exists {
		return fmt.Errorf("no mutex found for player %s", p.Name)
	}

	// Lock the mutex for the player's connection
	lock.Lock()
	defer lock.Unlock()

	// Perform the write operation
	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return fmt.Errorf("error sending message to player %s: %v", p.Name, err)
	}
	return nil
}

func (n *Network) SendInfoMessage(p *game.Player, message string) {
	dto := dtos.InfoDTO{Message: message}
	n.SendMessage(p, dto.Serialize())
}

func (n *Network) BroadcastInfoMessage(message string) {
	dto := dtos.InfoDTO{Message: message}
	n.broadcast <- string(dto.Serialize())
}

func (n *Network) BroadcastConnectionInfo() {

}

func (n Network) CloseConnection(p *game.Player) {
	conn := n.clients[*p]
	conn.Close()
}
