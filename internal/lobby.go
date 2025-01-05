package internal

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"uno/models/dtos"
	"uno/models/game"

	"github.com/gorilla/websocket"
)

// Room represents a game room
type Room struct {
	id         int
	game       Game
	maxPlayers int
}

func NewRoom(maxPlayers int) *Room {
	roomId := generateID()

	r := &Room{
		id:         roomId,
		game:       *NewGame(),
		maxPlayers: maxPlayers,
	}
	r.game.Room = r
	rooms[roomId] = r
	return r
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
	room := NewRoom(maxPlayers)

	game := &room.game
	player := AddPlayerToRoom(&w, room.id, playerName)

	// Respond with the room id
	conn := UpgradeWebsocket(w, r, room)
	game.Network.AddClient(*player, conn)

	dto := dtos.ConnectionDTO{
		PlayerName: playerName,
		RoomID: room.id,
		MaxPlayers: maxPlayers,
		Players: room.game.getAllPlayers(),
	}
	conn.WriteMessage(websocket.TextMessage, dto.Serialize())

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
	conn := UpgradeWebsocket(w, r, room)
	game.Network.AddClient(*player, conn)

	dto := dtos.ConnectionDTO{
		PlayerName: playerName,
		RoomID: room.id,
		MaxPlayers: room.maxPlayers,
		Players: room.game.getAllPlayers(),
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

func UpgradeWebsocket(w http.ResponseWriter, r *http.Request, room *Room) *websocket.Conn {
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
