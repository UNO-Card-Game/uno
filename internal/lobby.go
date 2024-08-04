package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"uno/models/dtos"
	"uno/models/game"
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
	game := &room.game
	player := AddPlayerToRoom(&w, roomId, playerName)

	// Respond with the room id
	dto := dtos.ConnectionDTO{
		playerName,
		room.id,
		maxPlayers,
		room.game.getAllPlayers(),
	}
	res := dto.Serialize()
	conn := UpgradeWebsocket(w, r, *room)
	game.Network.clients[*player] = conn
	conn.WriteMessage(websocket.TextMessage, res)
	game.Network.ListenToClient(player, room)

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

	dto := dtos.ConnectionDTO{
		playerName,
		room.id,
		room.maxPlayers,
		room.game.getAllPlayers(),
	}
	conn.WriteMessage(websocket.TextMessage, dto.Serialize())

	game.Network.ListenToClient(player, room)
}

func AddPlayerToRoom(w *http.ResponseWriter, roomId int, playerName string) *game.Player {
	r, ok := rooms[roomId]
	if !ok {
		http.Error(*w, "Room not found", http.StatusNotFound)
		return nil
	}
	g := &r.game
	if len(g.Players) >= r.maxPlayers {
		http.Error(*w, "Room is full", http.StatusForbidden)
		return nil
	}

	player := game.NewPlayer(playerName)
	g.AddPlayer(player)
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
