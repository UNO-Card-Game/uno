package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"uno/models"
)

type Network struct {
	//clients map[*websocket.Conn]*models.Player
	clients     map[models.Player]*websocket.Conn
	upgrader    websocket.Upgrader
	broadcast   chan string
	gameStarted bool
	mutex       sync.Mutex
}

func NewNetwork() *Network {
	return &Network{
		clients: make(map[models.Player]*websocket.Conn),
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

func (n Network) EstablishConnections(w http.ResponseWriter, r *http.Request, game *Game) {
	conn, err := n.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//defer func() {
	//	p := n.clients[conn]
	//	delete(n.clients, conn)
	//	conn.Close()
	//}()

	// Ask the user to enter their name
	var player *models.Player
	player = nil
	for {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Please enter your name: ")); err != nil {
			fmt.Println(err)
			return
		}
		_, name, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		// Check if the name exists in the players slice
		for _, p := range game.Players {
			if p.Name == string(name) {
				player = p
				break
			}
		}

		if player != nil {
			// Name found, assign with ClientName Map
			n.clients[*player] = conn
			fmt.Printf("New client '%s' joined\n", player.Name)
			break
		} else {
			// Name not found, ask the user to input again
			if err := conn.WriteMessage(websocket.TextMessage, []byte("Name not found. Please try again. Use names from players slice")); err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	// Wait for all players to join before starting the game
	if len(n.clients) == len(game.Players) && !n.gameStarted {
		start_msg := []byte("All players have joined. Send 'start' to begin the game.")
		game.Start()
		n.broadcast <- fmt.Sprintf("%s started the game", player.Name)
		if err := conn.WriteMessage(websocket.TextMessage, start_msg); err != nil {
			fmt.Println(err)
			return
		}
	}

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

func (n Network) ListenToClients() {

}

func (n Network) SendMessage(p *models.Player, message string) error {
	conn := n.clients[*p]
	fmt.Printf("conn:", conn)
	n.mutex.Lock()
	defer n.mutex.Unlock()
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		return fmt.Errorf("error sending message to p %s: %v", p.Name, err)
	}
	return nil
}

func (n Network) CloseConnection(p *models.Player) {
	conn := n.clients[*p]
	conn.Close()
}
