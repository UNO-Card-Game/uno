package internal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"uno/models"
)

// Room represents a game room
type Room struct {
	id         int  `json:"id"`
	game       Game `json:"game"`
	maxPlayers int  `json:"max_players"`
}

const MAX_ROOMS = 100

var rooms map[int]*Room

func init() {
	rooms = make(map[int]*Room)
}

// CreateRoomHandler handles requests to create a new room
func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if len(rooms) >= MAX_ROOMS {
		http.Error(w, "Maximum number of rooms reached", http.StatusForbidden)
		return
	}

	// Get query parameters
	playerName := r.URL.Query().Get("player_name")
	if playerName == "" {
		http.Error(w, "Missing player_name parameter", http.StatusBadRequest)
		return
	}

	maxPlayersStr := r.URL.Query().Get("max_players")
	if maxPlayersStr == "" {
		http.Error(w, "Missing max_players parameter", http.StatusBadRequest)
		return
	}

	maxPlayers, err := strconv.Atoi(maxPlayersStr)
	if err != nil {
		http.Error(w, "Invalid max_players parameter", http.StatusBadRequest)
		return
	}

	// Generate a unique id for the room if not provided
	roomId := generateID()

	// Create a new room and add it to the rooms map
	room := &Room{
		id:         roomId,
		game:       *NewGame(),
		maxPlayers: maxPlayers,
	}
	rooms[roomId] = room

	// Respond with the room id
	response := map[string]int{"room_id": roomId}
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}
	game := &room.game
	player := AddPlayerToRoom(&w, roomId, playerName)

	conn := UpgradeWebsocket(w, r, *room)
	game.Network.clients[*player] = conn
	conn.WriteMessage(websocket.TextMessage, res)
	game.Network.ListenToClient(player, game)

}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	playerName := r.URL.Query().Get("player_name")
	roomIdStr := r.URL.Query().Get("room_id")

	// Check if player_name and room_id are provided
	if playerName == "" || roomIdStr == "" {
		http.Error(w, "player_name and room_id are required", http.StatusBadRequest)
		return
	}

	// Validate room id
	roomId, err := strconv.Atoi(roomIdStr)
	if err != nil {
		http.Error(w, "room_id must be a valid integer", http.StatusBadRequest)
		return
	}
	room := rooms[roomId]
	game := &room.game
	player := AddPlayerToRoom(&w, roomId, playerName)
	conn := UpgradeWebsocket(w, r, *room)
	game.Network.clients[*player] = conn

	if len(game.Network.clients) == room.maxPlayers && game.GameStarted == false {
		game.Start()
		game.Network.BroadcastMessage("All players have joined. Game has started.")
	} else {
		game.Network.SendMessage(player, "Waiting for players to join the game.")
	}
	game.Network.ListenToClient(player, game)
}

func AddPlayerToRoom(w *http.ResponseWriter, roomId int, playerName string) *models.Player {
	room, ok := rooms[roomId]
	if !ok {
		http.Error(*w, "Room not found", http.StatusNotFound)
		return nil
	}
	game := &room.game
	if len(game.Players) >= room.maxPlayers {
		http.Error(*w, "Room is full", http.StatusForbidden)
		return nil
	}

	player := models.NewPlayer(playerName)
	game.AddPlayer(player)
	return player
}

func UpgradeWebsocket(w http.ResponseWriter, r *http.Request, room Room) *websocket.Conn {
	conn, err := room.game.Network.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return nil
	}
	return conn
}

// Generate a unique room id
func generateID() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(MAX_ROOMS)
}
