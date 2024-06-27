package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"uno/models/game"
)

type Network struct {
	//clients map[*websocket.Conn]*models.Player
	clients     map[game.Player]*websocket.Conn
	upgrader    websocket.Upgrader
	broadcast   chan string
	gameStarted bool
	mutex       sync.Mutex
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
		mutex:       sync.Mutex{},
	}
}

func (n *Network) BroadcastMessages() {

	for {
		select {
		case message := <-n.broadcast:
			for _, conn := range n.clients {
				// Use a mutex to synchronize writes to the websocket connection
				n.mutex.Lock()
				err := conn.WriteMessage(websocket.TextMessage, []byte(": "+message))
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						fmt.Printf("Error writing message to client : %v\n", err)
						err := conn.Close()
						if err != nil {
							fmt.Println(err)
							return
						}
					} else {
						fmt.Println("Error writing message to client:", err)
					}
				}
				n.mutex.Unlock()
			}
		}
	}

}

func (n Network) ListenToClient(player *game.Player, game *Game) {

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

func (n Network) BroadcastMessage(message string) {
	n.broadcast <- fmt.Sprintf(string(message))
}

func (n Network) SendMessage(p *game.Player, message string) error {
	conn := n.clients[*p]
	n.mutex.Lock()
	defer n.mutex.Unlock()
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return fmt.Errorf("error sending message to p %s: %v", p.Name, err)
	}
	return nil
}

func (n Network) CloseConnection(p *game.Player) {
	conn := n.clients[*p]
	conn.Close()
}
