package network

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"sync"
	"uno/internal"
)

var (
	clients  = make(map[*websocket.Conn]string)
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Accepts requests from every source
		},
	}
	broadcast   = make(chan string)
	gameStarted = false
	mutex       sync.Mutex // Mutex to synchronize writes to websocket connections
	// wg          sync.WaitGroup
	// goroutines  []func()
)

func BroadcastMessages() {

	for {
		select {
		case message := <-broadcast:
			for client, _ := range clients {
				// Use a mutex to synchronize writes to the websocket connection
				mutex.Lock()
				err := client.WriteMessage(websocket.TextMessage, []byte(": "+message))
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						fmt.Printf("Error writing message to client : %v\n", err)
						client.Close()
						delete(clients, client)
					} else {
						fmt.Println("Error writing message to client:", err)
					}
				}
				mutex.Unlock()
			}
		}
	}

}

func HandleConnections(w http.ResponseWriter, r *http.Request, game *internal.Game) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		delete(clients, conn)
		conn.Close()
	}()

	// Ask the user to enter their name
	var clientName string
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
		nameFound := false
		for _, p := range game.Players {
			if p.Name == string(name) {
				nameFound = true
				break
			}
		}
		if nameFound {
			// Name found, assign with ClientName Map
			clientName = string(name)
			break
		} else {
			// Name not found, ask the user to input again
			if err := conn.WriteMessage(websocket.TextMessage, []byte("Name not found. Please try again. Use names from players slice")); err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	clients[conn] = clientName
	for _, player := range game.Players {
		if player.Name == clientName {
			player.SetConn(conn)
		}
	}
	fmt.Printf("New client '%s' joined\n", clientName)

	// Wait for all players to join before starting the game
	if len(clients) == len(game.Players) && !gameStarted {
		start_msg := []byte("All players have joined. Send 'start' to begin the game.")

		if err := conn.WriteMessage(websocket.TextMessage, start_msg); err != nil {
			fmt.Println(err)
			return
		}
	}
	//go broadcastMessages()

	for {
		// Read message from browser/terminal
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// Start the game if the first player sends the "start" message
		if strings.ToLower(string(msg)) == "start" && len(clients) == len(game.Players) && !gameStarted {
			gameStarted = true

			game.Start()
			internal.UNoOLogoPrint()
			broadcast <- fmt.Sprintf("%s started the game", clientName)
		}

		// Add your own logic here to filter out certain messages
		game.HandleMessage(string(msg), conn, clientName)

		// Before broadcasting them to other clients
		if internal.ShouldBroadcast(string(msg)) {
			// Print the message to the console
			//fmt.Printf("%s sent: %s\n", clientName, string(msg))
			broadcast <- fmt.Sprintf("%s: %s", clientName, string(msg))
		}
	}
}
