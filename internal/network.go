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
		dto := dtos.InfoDTO{Message: "All players have joined. Game has started."}
		conn_info_dto := dtos.ConnectionDTO{
			player.Name,
			r.id,
			r.maxPlayers,
			r.game.getAllPlayers(),
		}
		game.Network.BroadcastMessage(dto.Serialize())
		game.Network.BroadcastMessage(conn_info_dto.Serialize())
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

		// Add your own logic here to filter out certain messages
		game.HandleMessage(string(msg), player)

		// Before broadcasting them to other clients
		if ShouldBroadcast(string(msg)) {
			// Print the message to the console
			//fmt.Printf("%s sent: %s\n", clientName, string(msg))
			n.broadcast <- fmt.Sprintf("%s: %s", player.Name, string(msg))
		}
	}
}

func (n Network) BroadcastMessage(message []byte) {
	n.broadcast <- string(message)
}

func (n *Network) SendMessage(p *game.Player, message string) error {
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

func (n Network) CloseConnection(p *game.Player) {
	conn := n.clients[*p]
	conn.Close()
}
